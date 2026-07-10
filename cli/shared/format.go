package shared

import rcdisplay "github.com/yumauri/fbrcm/core/rc/display"

// FormatParameterHeader formats a parameter key with its group when present.
func FormatParameterHeader(key, group string) string {
	return rcdisplay.FormatParameterHeader(key, group)
}
