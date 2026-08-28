package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidateQuotaProjectID validates one caller-selected Google Cloud quota
// project identifier without checking whether it exists or is accessible.
func ValidateQuotaProjectID(value string) error {
	if value == "" {
		return fmt.Errorf("quota project ID must not be empty")
	}
	if value != strings.TrimSpace(value) {
		return fmt.Errorf("quota project ID %q must not have surrounding whitespace", value)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("quota project ID must be valid UTF-8")
	}
	if strings.IndexFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return fmt.Errorf("quota project ID %q must not contain whitespace or control characters", value)
	}
	if err := ValidatePhysicalProjectID(value); err != nil {
		return fmt.Errorf("quota project ID %q must reference a physical project: %w", value, err)
	}
	return nil
}

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
