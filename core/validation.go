package core

import rcvalue "github.com/yumauri/fbrcm/core/rc/value"

// IsJSONNumber reports whether value is a non-empty JSON number literal.
func IsJSONNumber(value string) bool {
	return rcvalue.IsJSONNumber(value)
}
