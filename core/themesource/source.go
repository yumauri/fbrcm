// Package themesource reads theme files from hosted repositories without Git.
package themesource

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Limits bound work performed by one repository import.
const (
	MaximumFileBytes          = 1 << 20
	MaximumBatchBytes         = 16 << 20
	MaximumFiles              = 256
	maximumMetadataBytes      = 8 << 20
	maximumTotalMetadataBytes = 64 << 20
	maximumRequests           = 1024
	maximumPages              = 128
	maximumSegments           = 64
)

// Error carries source failure semantics independently of CLI presentation.
// Messages never include tokens, response bodies, or transport error text.
type Error struct {
	SourceURL    string
	Kind         string
	Provider     string
	Operation    string
	HTTPStatus   int
	RetryAfterMS int64
	Candidates   []string
	Err          error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "repository theme source: " + e.Kind
}
func (e *Error) Unwrap() error            { return e.Err }
func failure(kind, message string) *Error { return &Error{Kind: kind, Err: fmt.Errorf("%s", message)} }

// Provenance records exactly which repository snapshot was imported.
type Provenance struct {
	Provider     string `json:"provider" contract:"enum=github|gitlab"`
	Repository   string `json:"repository"`
	RequestedRef string `json:"requested_ref"`
	Commit       string `json:"commit" contract:"pattern=^(?:[0-9a-f]{40}|[0-9a-f]{64})$"`
	Path         string `json:"path"`
}

// Source is a parsed provider browser URL, before resolving its ref/path boundary.
type Source struct {
	URL        string
	Provider   string
	Repository string
	Directory  bool
	parts      []string
	ref        string
}

// Parse returns nil for ordinary file paths and raw HTTP sources. ref is only
// a disambiguator: it must identify a revision prefix already in the URL.
func Parse(raw, ref string) (*Source, error) {
	if utf8.RuneCountInString(ref) > 1024 || strings.ContainsFunc(ref, func(r rune) bool { return r <= 32 || r == 127 }) {
		return nil, failure("invalid", "--ref must be a nonempty revision prefix without whitespace or control characters, at most 1024 characters")
	}
	// Local paths may contain percent signs that are not URL escapes. Leave
	// their inspection to the existing file/directory importer.
	lower := strings.ToLower(raw)
	if ref == "" && !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, failure("invalid", "invalid theme source URL")
	}
	host := strings.ToLower(u.Hostname())
	if host != "github.com" && host != "gitlab.com" {
		if ref != "" {
			return nil, failure("invalid", "--ref requires a GitHub.com or GitLab.com browser URL")
		}
		return nil, nil
	}
	if u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Opaque != "" {
		return nil, failure("invalid", "repository theme sources require HTTPS without credentials or a custom port")
	}
	// Decode one segment at a time so encoded slashes cannot change the project boundary.
	encoded := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	segments := make([]string, len(encoded))
	for i, part := range encoded {
		decoded, decodeErr := url.PathUnescape(part)
		if decodeErr != nil || decoded == "" || decoded == "." || decoded == ".." || strings.ContainsAny(decoded, "\\") || strings.ContainsFunc(decoded, unicode.IsControl) {
			return nil, failure("invalid", "repository URL contains an invalid path segment")
		}
		segments[i] = decoded
	}
	s := &Source{ref: ref}
	var marker int
	if host == "github.com" {
		s.Provider = "github"
		marker = 2
	} else {
		s.Provider = "gitlab"
		marker = -1
		for i, part := range segments {
			if part == "-" {
				marker = i + 1
				break
			}
		}
		if marker >= 3 && marker < len(segments) && segments[marker] == "raw" && ref == "" {
			return nil, nil
		}
	}
	if marker < 2 || len(segments) <= marker+1 || (segments[marker] != "blob" && segments[marker] != "tree") {
		return nil, failure("invalid", "use a GitHub or GitLab browser URL containing blob/<ref>/<path> or tree/<ref>/<path>")
	}
	repoParts := segments[:marker]
	if s.Provider == "gitlab" {
		repoParts = segments[:marker-1]
	}
	for _, part := range repoParts {
		if strings.Contains(part, "/") {
			return nil, failure("invalid", "repository namespace segments cannot contain encoded slashes")
		}
	}
	s.Repository = strings.Join(repoParts, "/")
	s.Directory = segments[marker] == "tree"
	s.parts = strings.Split(strings.Join(segments[marker+1:], "/"), "/")
	if len(s.parts) > maximumSegments {
		return nil, failure("limit", "repository URL exceeds the 64-segment limit")
	}
	for _, part := range s.parts {
		if part == "" || part == "." || part == ".." {
			return nil, failure("invalid", "repository URL contains an invalid revision or path")
		}
	}
	if ref != "" {
		prefix := strings.Split(ref, "/")
		if len(prefix) > len(s.parts) || strings.Join(s.parts[:len(prefix)], "/") != ref {
			return nil, failure("invalid", "--ref must match the revision prefix already present in the browser URL")
		}
	}
	// Presentation query parameters and fragments are never sent to the API or output.
	u.RawQuery, u.Fragment, u.RawFragment, u.ForceQuery = "", "", "", false
	u.Host = host
	s.URL = u.String()
	return s, nil
}
