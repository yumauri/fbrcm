package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validatePathSegment(name, kind string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("%s cannot be empty", kind)
	}
	if trimmed != name {
		return fmt.Errorf("%s cannot have leading or trailing whitespace", kind)
	}
	if name == "." || name == ".." {
		return fmt.Errorf("%s %q is reserved", kind, name)
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("%s cannot contain path separators", kind)
	}
	if filepath.Clean(name) != name {
		return fmt.Errorf("%s must be a single path segment", kind)
	}
	return nil
}
