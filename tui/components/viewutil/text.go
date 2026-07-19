package viewutil

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

func PadRight(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

// IndentLines adds the same left inset to every rendered line in value.
func IndentLines(value string, width int) string {
	padding := strings.Repeat(" ", max(width, 0))
	return padding + strings.ReplaceAll(value, "\n", "\n"+padding)
}

// SelectionText renders selected list text without styling its surrounding inset.
func SelectionText(value string, selected bool) string {
	if selected {
		return styles.TitleStyle(true).Render(value)
	}
	return value
}

// SelectorLine renders the shared two-cell list inset used by profile selectors.
func SelectorLine(value string, selected bool) string {
	return "  " + SelectionText(value, selected)
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
