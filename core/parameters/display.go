package parameters

import (
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

// FormatRemoteConfigDisplayValue formats a Remote Config value for tree summaries
// and CLI table output.
func FormatRemoteConfigDisplayValue(value firebase.RemoteConfigValue, valueType string) string {
	return rcdisplay.FormatSummary(value, valueType)
}

// SummarizeRemoteConfigDisplayValue retains managed-value structure for styled
// human-readable renderers.
func SummarizeRemoteConfigDisplayValue(value firebase.RemoteConfigValue, valueType string) rcdisplay.ValueSummary {
	return rcdisplay.SummarizeValue(value, valueType)
}
