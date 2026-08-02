package viewutil

import (
	"strings"

	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/styles"
)

// RenderManagedValueSummary renders Firebase-managed and unknown Remote Config
// values with the shared TUI palette. The boolean reports whether the summary
// represents one of those special values.
func RenderManagedValueSummary(summary rcdisplay.ValueSummary, valueType string) (string, bool) {
	if summary.Kind == rcdisplay.ValueSummaryPlain {
		return "", false
	}
	if styles.NoColorEnabled() {
		return summary.Text, true
	}
	switch summary.Kind {
	case rcdisplay.ValueSummaryExperiment:
		if summary.Experiment == nil {
			return styles.PanelMuted.Render(summary.Text), true
		}
		var rendered strings.Builder
		rendered.WriteString(styles.PanelMuted.Render("⚗ "))
		if summary.Experiment.Percentage != "" {
			rendered.WriteString(styles.SecondaryTitleCount.Render(summary.Experiment.Percentage))
			rendered.WriteString(styles.PanelMuted.Render(" : "))
		}
		for index, value := range summary.Experiment.Values {
			if index > 0 {
				rendered.WriteString(styles.PanelMuted.Render(" | "))
			}
			rendered.WriteString(renderManagedValue(value, valueType))
		}
		return rendered.String(), true
	case rcdisplay.ValueSummaryRollout:
		if summary.Rollout == nil {
			return styles.PanelMuted.Render(summary.Text), true
		}
		return styles.PanelMuted.Render("◐ ") +
			styles.SecondaryTitleCount.Render(summary.Rollout.Percentage) +
			styles.PanelMuted.Render(" → ") +
			renderManagedValue(summary.Rollout.Value, valueType) +
			styles.PanelMuted.Render(" | ") +
			renderManagedValue("(no change)", valueType), true
	default:
		return styles.PanelMuted.Render(summary.Text), true
	}
}

func renderManagedValue(value, valueType string) string {
	if rcdisplay.IsManagedValuePlaceholder(value) {
		return styles.PanelMuted.Render(value)
	}
	return corestyles.ValueTextStyle(value, valueType).Render(value)
}
