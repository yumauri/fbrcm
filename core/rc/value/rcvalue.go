package value

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
)

var jsonNumberPattern = regexp.MustCompile(`^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$`)

// IsJSONNumber reports whether value is a non-empty JSON number literal.
func IsJSONNumber(value string) bool {
	return jsonNumberPattern.MatchString(strings.TrimSpace(value))
}

// ValidRawValueForType reports whether value is valid for valueType using TUI
// details semantics (strict boolean literals, JSON number parse for NUMBER).
func ValidRawValueForType(value, valueType string) bool {
	value = strings.TrimSpace(value)
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "STRING", "":
		return true
	case "BOOLEAN":
		return value == "true" || value == "false"
	case "NUMBER":
		var number float64
		if err := json.Unmarshal([]byte(value), &number); err != nil {
			return false
		}
		return !math.IsInf(number, 0) && !math.IsNaN(number)
	case "JSON":
		return json.Valid([]byte(value))
	default:
		return false
	}
}

// ValidateRawValueForType validates value for valueType using draft mutation
// semantics (case-insensitive boolean, regex JSON number literal for NUMBER).
func ValidateRawValueForType(value, valueType string) error {
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "", "STRING":
		return nil
	case "BOOLEAN":
		switch strings.TrimSpace(strings.ToLower(value)) {
		case "true", "false":
			return nil
		default:
			return fmt.Errorf("invalid boolean")
		}
	case "NUMBER":
		if !IsJSONNumber(value) {
			return fmt.Errorf("invalid number")
		}
		return nil
	case "JSON":
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("invalid json")
		}
		return nil
	default:
		return fmt.Errorf("invalid value type %q", valueType)
	}
}
