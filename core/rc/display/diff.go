package display

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
)

// FormatDiff formats a Remote Config value for previews and diffs.
func FormatDiff(value firebase.RemoteConfigValue) string {
	switch {
	case len(value.PersonalizationValue) > 0:
		return string(normalizeJSON(bytes.TrimSpace(value.PersonalizationValue)))
	case len(value.RolloutValue) > 0:
		return string(normalizeJSON(bytes.TrimSpace(value.RolloutValue)))
	case value.UseInAppDefault:
		return "useInAppDefault"
	default:
		return FormatPlainValue(value.Value)
	}
}

// FormatPlainValue formats a plain string value for previews and diffs.
func FormatPlainValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(empty)"
	}
	quoted, err := json.Marshal(value)
	if err != nil {
		return value
	}
	if IsSimpleToken(value) {
		return value
	}
	return string(quoted)
}

// IsSimpleToken reports whether value can be shown without JSON quoting.
func IsSimpleToken(value string) bool {
	for _, r := range value {
		if r == ' ' || r == '\t' || r == '\n' || r == '"' {
			return false
		}
	}
	return true
}

func normalizeJSON(body []byte) []byte {
	return firebase.NormalizeJSONEscapes(body)
}
