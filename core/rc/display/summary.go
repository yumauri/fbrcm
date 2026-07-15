package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
)

const localDateTimeLayout = "2006-01-02 15:04:05"

// FormatCount formats a count with the grammatically appropriate noun.
func FormatCount(count int, singular, plural string) string {
	noun := plural
	if count == 1 {
		noun = singular
	}
	return fmt.Sprintf("%d %s", count, noun)
}

// FormatLocalDateTime formats a timestamp in the local timezone for terminal displays.
func FormatLocalDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format(localDateTimeLayout)
}

// FormatSummary formats a Remote Config value for tree summaries and table output.
func FormatSummary(value firebase.RemoteConfigValue, valueType string) string {
	switch {
	case value.UseInAppDefault:
		return "<in-app default>"
	case len(value.PersonalizationValue) > 0:
		return "<personalization>"
	case len(value.RolloutValue) > 0:
		return "<rollout>"
	case value.Value == "":
		return "(empty " + EmptyValueType(valueType) + ")"
	default:
		return strings.ReplaceAll(value.Value, "\n", "\\n")
	}
}

// EmptyValueType normalizes a parameter value type for empty-value labels.
func EmptyValueType(valueType string) string {
	valueType = strings.TrimSpace(strings.ToLower(valueType))
	if valueType == "" {
		return "string"
	}
	return valueType
}

// FormatRawValue formats a stored raw value string for detail panels.
func FormatRawValue(value, valueType string) string {
	if value == "" {
		return "(empty " + EmptyValueType(valueType) + ")"
	}
	return strings.ReplaceAll(value, "\n", "\\n")
}
