package core

import (
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/parameters"
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
