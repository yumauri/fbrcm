package viewutil

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func PadRight(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

// TruncatePlain truncates value to at most width runes, counting runes rather
// than display width. It returns an empty string when width is not positive.
func TruncatePlain(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	return string(runes[:width])
}
