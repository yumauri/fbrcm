package viewutil

import (
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
	if summary.Kind != rcdisplay.ValueSummaryRollout || summary.Rollout == nil {
		return styles.PanelMuted.Render(summary.Text), true
	}
	return styles.PanelMuted.Render("◐ ") +
		styles.SecondaryTitleCount.Render(summary.Rollout.Percentage) +
		styles.PanelMuted.Render(" → ") +
		corestyles.ValueTextStyle(summary.Rollout.Value, valueType).Render(summary.Rollout.Value) +
		styles.PanelMuted.Render(" / ◑ (no change)"), true
}
