package core

import (
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/parameters"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

type (
	ParametersTree      = parameters.Tree
	ParametersGroup     = parameters.Group
	ParametersEntry     = parameters.Entry
	ParametersValue     = parameters.Value
	ParametersCondition = parameters.Condition
)

// FormatRemoteConfigDisplayValue formats a Remote Config value for tree summaries
// and CLI table output.
func FormatRemoteConfigDisplayValue(value firebase.RemoteConfigValue, valueType string) string {
	return parameters.FormatRemoteConfigDisplayValue(value, valueType)
}

// SummarizeRemoteConfigDisplayValue retains managed-value structure for styled
// human-readable renderers.
func SummarizeRemoteConfigDisplayValue(value firebase.RemoteConfigValue, valueType string) rcdisplay.ValueSummary {
	return parameters.SummarizeRemoteConfigDisplayValue(value, valueType)
}
