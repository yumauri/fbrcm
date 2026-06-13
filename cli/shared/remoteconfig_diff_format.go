package shared

import (
	"bytes"
	"encoding/json"
	"strings"

	"charm.land/lipgloss/v2"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/firebase"
)

// FormatRemoteValue formats a Firebase Remote Config value for previews and diffs.
func FormatRemoteValue(value firebase.RemoteConfigValue) string {
	switch {
	case len(value.PersonalizationValue) > 0:
		return string(NormalizeExportJSON(bytes.TrimSpace(value.PersonalizationValue)))
	case len(value.RolloutValue) > 0:
		return string(NormalizeExportJSON(bytes.TrimSpace(value.RolloutValue)))
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

func formatGroupValue(group string) string {
	if group == "" {
		return "(root)"
	}
	return "[" + group + "]"
}

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(empty)"
	}
	return value
}

func colorAdded(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorAdded).Render(value)
}

func colorRemoved(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorRemoved).Render(value)
}

func colorChanged(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorChanged).Render(value)
}
