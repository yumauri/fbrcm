package viewutil

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// WrapRenderedLine hard-wraps one ANSI-styled line and applies a fixed left
// inset to continuation rows.
func WrapRenderedLine(value string, width, continuationIndent int) []string {
	if width <= 0 {
		return []string{""}
	}
	if lipgloss.Width(value) <= width {
		return []string{value}
	}

	continuationIndent = min(max(continuationIndent, 0), max(width-1, 0))
	indent := strings.Repeat(" ", continuationIndent)
	lines := make([]string, 0, 2)
	remaining := value
	for lipgloss.Width(remaining) > width {
		part := ansi.Truncate(remaining, width, "")
		partWidth := lipgloss.Width(part)
		if partWidth <= 0 {
			break
		}
		lines = append(lines, part)
		remaining = ansi.Cut(remaining, partWidth, lipgloss.Width(remaining))
		remaining = indent + remaining
	}
	if remaining != "" || len(lines) == 0 {
		lines = append(lines, remaining)
	}
	return lines
}
