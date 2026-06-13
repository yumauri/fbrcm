package shared

import "fmt"

// FormatParameterHeader formats a parameter key with its group when present.
func FormatParameterHeader(key, group string) string {
	if group == "" {
		return key
	}
	return fmt.Sprintf("%s [%s]", key, group)
}
