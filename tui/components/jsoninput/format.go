package jsoninput

import (
	"bytes"
	"encoding/json"
)

func prettyJSON(value string) string {
	if !json.Valid([]byte(value)) {
		return value
	}
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(value), "", "  "); err != nil {
		return value
	}
	return out.String()
}

// PrettyJSON formats valid JSON with indentation; invalid input is returned unchanged.
func PrettyJSON(value string) string {
	return prettyJSON(value)
}

// IsValidJSON reports whether value is valid JSON.
func IsValidJSON(value string) bool {
	return json.Valid([]byte(value))
}

// CompactJSON returns minified JSON when value is valid.
func CompactJSON(value string) (string, bool) {
	if !json.Valid([]byte(value)) {
		return "", false
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(value)); err != nil {
		return "", false
	}
	return buf.String(), true
}
