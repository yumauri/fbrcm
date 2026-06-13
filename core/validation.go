package core

import (
	"regexp"
	"strings"
)

var jsonNumberPattern = regexp.MustCompile(`^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$`)

// IsJSONNumber reports whether value is a non-empty JSON number literal.
func IsJSONNumber(value string) bool {
	return jsonNumberPattern.MatchString(strings.TrimSpace(value))
}
