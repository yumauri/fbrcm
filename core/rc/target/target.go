// Package target parses the client/server template target syntax used by the CLI.
package target

import (
	"fmt"
	"strings"
)

type Kind string

const (
	Client Kind = "client"
	Server Kind = "server"
)

const ServerNamespace = "firebase-server"

type Target struct {
	Kind      Kind
	ProjectID string
}

// Parse treats an unqualified value as a client-template target. The explicit
// client@ prefix is accepted but omitted from the canonical target string.
func Parse(value string) (Target, error) {
	target, _, err := ParseSelector(value)
	return target, err
}

// ParseSelector parses a template target and reports whether the template kind
// was explicitly selected with a client@ or server@ prefix.
func ParseSelector(value string) (Target, bool, error) {
	value = strings.TrimSpace(value)
	kind := Client
	projectID := value
	explicit := false
	switch {
	case hasPrefixFold(value, string(Client)+"@"):
		explicit = true
		projectID = strings.TrimSpace(value[len(Client)+1:])
	case hasPrefixFold(value, string(Server)+"@"):
		explicit = true
		kind = Server
		projectID = strings.TrimSpace(value[len(Server)+1:])
	}
	if projectID == "" {
		return Target{}, explicit, fmt.Errorf("%s template target requires a project", kind)
	}
	return Target{Kind: kind, ProjectID: projectID}, explicit, nil
}

func (t Target) String() string {
	if t.Kind == Server {
		return string(Server) + "@" + t.ProjectID
	}
	return t.ProjectID
}

func (t Target) WithProjectID(projectID string) Target {
	t.ProjectID = projectID
	return t
}

func (t Target) Namespace() string {
	if t.Kind == Server {
		return ServerNamespace
	}
	return ""
}

// ExactFilter returns a target-aware exact project filter suitable for the
// CLI's --project flag.
func ExactFilter(value string) (string, error) {
	target, err := Parse(value)
	if err != nil {
		return "", err
	}
	target.ProjectID = "=" + target.ProjectID
	return target.String(), nil
}

func hasPrefixFold(value, prefix string) bool {
	return len(value) >= len(prefix) && strings.EqualFold(value[:len(prefix)], prefix)
}
