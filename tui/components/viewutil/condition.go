package viewutil

import (
	"strings"
)

// ConditionColorValue formats a Remote Config condition color consistently
// wherever it is presented as a field value.
func ConditionColorValue(color string) string {
	if strings.TrimSpace(color) == "" {
		return "No color"
	}
	return "● " + strings.ReplaceAll(color, "_", " ")
}
