package draft

import "github.com/yumauri/fbrcm/core/rootgroup"

// NormalizeGroupKey converts tree/TUI group keys to Firebase wire keys.
func NormalizeGroupKey(groupKey string) string {
	if groupKey == rootgroup.TreeKey {
		return rootgroup.WireKey
	}
	return groupKey
}
