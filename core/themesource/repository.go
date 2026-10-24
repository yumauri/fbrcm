package themesource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var commitPattern = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)

// File identifies a regular file within the pinned snapshot.
type File struct {
	Name, Path string
	Size       int64
}
type entry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	Mode string `json:"mode"`
	SHA  string `json:"sha"`
	Size int64  `json:"size"`
}

func (e entry) regular() bool {
	return (e.Type == "blob" || e.Type == "file") && (e.Mode == "100644" || e.Mode == "100755")
}

// Repository is a resolved source. Files and reads always use its Commit.
type Repository struct {
	Provenance Provenance
	Directory  bool
	client     *Client
	rootTree   string
}

// Resolve discovers the revision boundary and validates the selected entry kind.
func (c *Client) Resolve(ctx context.Context) (*Repository, error) {
	source := c.source
	var matches []*Repository
	var missing error
	for split := 1; split <= len(source.parts); split++ {
		ref := strings.Join(source.parts[:split], "/")
		selectedPath := strings.Join(source.parts[split:], "/")
		if source.ref != "" && source.ref != ref || !source.Directory && selectedPath == "" {
			continue
		}
		repo, err := c.resolveCommit(ctx, ref, selectedPath)
		if err != nil {
			if isMissing(err) {
				missing = err
				continue
			}
			return nil, err
		}
		if selectedPath == "" {
			if source.Directory {
				matches = append(matches, repo)
			}
			continue
		}
		selected, err := repo.lookup(ctx, selectedPath)
		if err != nil {
			if isMissing(err) {
				missing = err
				continue
			}
			return nil, err
		}
		if source.Directory && selected.Type == "tree" || !source.Directory && selected.regular() {
			matches = append(matches, repo)
		}
	}
	if len(matches) == 0 {
		if missing != nil {
			return nil, missing
		}
		return nil, failure("not_found", "repository revision or regular theme path was not found or is inaccessible; check the URL, entry kind, and token access")
	}
	if len(matches) > 1 {
		candidates := make([]string, len(matches))
		for i, match := range matches {
			candidates[i] = match.Provenance.RequestedRef
		}
		return nil, &Error{Kind: "ambiguous", SourceURL: source.URL, Candidates: candidates, Err: fmt.Errorf("repository URL has multiple revision/path interpretations; specify --ref (%s)", strings.Join(candidates, ", "))}
	}
	return matches[0], nil
}

func (c *Client) resolveCommit(ctx context.Context, ref, selectedPath string) (*Repository, error) {
	repo := &Repository{client: c, Directory: c.source.Directory, Provenance: Provenance{Provider: c.source.Provider, Repository: c.source.Repository, RequestedRef: ref, Path: selectedPath}}
	if c.source.Provider == "github" {
		var result struct {
			SHA    string `json:"sha"`
			Commit struct {
				Tree struct {
					SHA string `json:"sha"`
				} `json:"tree"`
			} `json:"commit"`
		}
		if _, err := c.json(ctx, c.base()+"/commits/"+url.PathEscape(ref), &result); err != nil {
			return nil, err
		}
		repo.Provenance.Commit, repo.rootTree = result.SHA, result.Commit.Tree.SHA
		if !commitPattern.MatchString(repo.rootTree) {
			return nil, failure("request_failed", "repository API returned an invalid root tree")
		}
	} else {
		var result struct {
			ID string `json:"id"`
		}
		if _, err := c.json(ctx, c.base()+"/repository/commits/"+url.PathEscape(ref), &result); err != nil {
			return nil, err
		}
		repo.Provenance.Commit = result.ID
	}
	if !commitPattern.MatchString(repo.Provenance.Commit) {
		return nil, failure("request_failed", "repository API returned an invalid commit")
	}
	return repo, nil
}

func isMissing(err error) bool {
	var source *Error
	return errors.As(err, &source) && source.Kind == "not_found"
}

func (r *Repository) lookup(ctx context.Context, selectedPath string) (entry, error) {
	parent, name := path.Split(selectedPath)
	entries, err := r.list(ctx, strings.TrimSuffix(parent, "/"))
	if err != nil {
		return entry{}, err
	}
	for _, item := range entries {
		if item.Name == name {
			return item, nil
		}
	}
	return entry{}, failure("not_found", "repository path was not found or is inaccessible")
}

func (r *Repository) list(ctx context.Context, directory string) ([]entry, error) {
	if r.Provenance.Provider == "gitlab" {
		return r.gitlabList(ctx, directory)
	}
	tree := r.rootTree
	for part := range strings.SplitSeq(directory, "/") {
		if part == "" {
			continue
		}
		entries, err := r.githubTree(ctx, tree)
		if err != nil {
			return nil, err
		}
		next := ""
		for _, item := range entries {
			if item.Name == part && item.Type == "tree" {
				next = item.SHA
				break
			}
		}
		if next == "" {
			return nil, failure("not_found", "repository directory was not found")
		}
		tree = next
	}
	return r.githubTree(ctx, tree)
}

func (r *Repository) githubTree(ctx context.Context, tree string) ([]entry, error) {
	if !commitPattern.MatchString(tree) {
		return nil, failure("request_failed", "repository API returned an invalid tree")
	}
	endpoint := r.client.base() + "/git/trees/" + tree
	if cached, ok := r.client.cache[endpoint]; ok {
		return cached, nil
	}
	var result struct {
		Tree      []entry `json:"tree"`
		Truncated bool    `json:"truncated"`
	}
	if _, err := r.client.json(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	if result.Truncated {
		return nil, failure("limit", "repository directory listing is truncated; select a smaller directory")
	}
	for i := range result.Tree {
		result.Tree[i].Name = result.Tree[i].Path
		if !validName(result.Tree[i].Name) {
			return nil, failure("request_failed", "repository API returned an invalid entry name")
		}
	}
	r.client.cache[endpoint] = result.Tree
	return result.Tree, nil
}

func (r *Repository) gitlabList(ctx context.Context, directory string) ([]entry, error) {
	query := url.Values{"ref": {r.Provenance.Commit}, "path": {directory}, "per_page": {"100"}, "page": {"1"}}
	cacheKey := r.client.base() + "/repository/tree?" + query.Encode()
	if cached, ok := r.client.cache[cacheKey]; ok {
		return cached, nil
	}
	entries := make([]entry, 0)
	for page := 1; page <= maximumPages; page++ {
		query.Set("page", strconv.Itoa(page))
		var items []entry
		headers, err := r.client.json(ctx, r.client.base()+"/repository/tree?"+query.Encode(), &items)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if !validName(item.Name) || item.Path != path.Join(directory, item.Name) {
				return nil, failure("request_failed", "repository API returned an invalid entry path")
			}
		}
		entries = append(entries, items...)
		next := headers.Get("X-Next-Page")
		if next == "" && len(items) < 100 {
			r.client.cache[cacheKey] = entries
			return entries, nil
		}
		if next != "" && next != strconv.Itoa(page+1) {
			return nil, failure("request_failed", "repository API returned invalid pagination")
		}
	}
	return nil, failure("limit", "repository directory exceeds the 128-page listing limit")
}

func validName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\") && !strings.ContainsFunc(name, unicode.IsControl)
}

// Files returns regular files in the selected directory, or the selected file.
func (r *Repository) Files(ctx context.Context) ([]File, error) {
	if !r.Directory {
		item, err := r.lookup(ctx, r.Provenance.Path)
		if err != nil {
			return nil, err
		}
		return []File{{Name: item.Name, Path: r.Provenance.Path, Size: item.Size}}, nil
	}
	entries, err := r.list(ctx, r.Provenance.Path)
	if err != nil {
		return nil, err
	}
	files := make([]File, 0)
	for _, item := range entries {
		if item.regular() && path.Ext(item.Name) == ".toml" {
			files = append(files, File{Name: item.Name, Path: path.Join(r.Provenance.Path, item.Name), Size: item.Size})
		}
	}
	if len(files) > MaximumFiles {
		return nil, failure("limit", "repository directory exceeds the 256-theme limit")
	}
	return files, nil
}

// Read downloads a regular file from this snapshot within the import budgets.
func (r *Repository) Read(ctx context.Context, file File) ([]byte, error) {
	if file.Size > MaximumFileBytes {
		return nil, failure("limit", "repository theme exceeds the 1 MiB file limit")
	}
	endpoint := r.client.base() + "/contents/" + escapedPath(file.Path) + "?" + url.Values{"ref": {r.Provenance.Commit}}.Encode()
	if r.Provenance.Provider == "gitlab" {
		endpoint = r.client.base() + "/repository/files/" + url.PathEscape(file.Path) + "/raw?" + url.Values{"ref": {r.Provenance.Commit}}.Encode()
	}
	data, _, err := r.client.get(ctx, endpoint, true)
	return data, err
}
