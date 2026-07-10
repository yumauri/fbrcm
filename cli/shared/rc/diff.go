package rc

import (
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

// ParamSlotPreview identifies a parameter slot for import conflict previews.
type ParamSlotPreview = rcdiff.ParamSlotPreview

// RenderRemoteConfigDiff renders a human-readable diff between two Remote Config snapshots.
func RenderRemoteConfigDiff(currentCfg, finalCfg *firebase.RemoteConfig) (string, bool) {
	return rcdiff.RenderRemoteConfigDiff(currentCfg, finalCfg)
}

// RenderConflictPreview renders a single import conflict line for interactive merge.
func RenderConflictPreview(label string, currentValue, importValue any) string {
	return rcdiff.RenderConflictPreview(label, currentValue, importValue)
}

// RenderConflictChoiceValue summarizes a conflict choice for prompt labels.
func RenderConflictChoiceValue(value any) string {
	return rcdiff.RenderConflictChoiceValue(value)
}

// RenderRemovedParameterDetail renders a red-colored preview of a parameter that is about to be deleted.
func RenderRemovedParameterDetail(key, group string, param firebase.RemoteConfigParam) string {
	return rcdiff.RenderRemovedParameterDetail(key, group, param)
}
