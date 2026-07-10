package display

import (
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
)

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
