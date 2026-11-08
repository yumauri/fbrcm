package projects

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func renderPanel(body string, width, height int, active, borderActive bool, scrollbar scrollbarState, secondary secondaryTitleState, footer []string) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(borderActive)
	innerWidth := max(width-1, 0)
	contentHeight := max(height-2-len(footer), 0)
	lines := []string{viewutil.SidebarHeader(width, panelTitleKey(), panelTitleLabel, active, borderStyle, secondary.text, secondary.style)}
	bodyLines := strings.Split(body, "\n")
	for i := range contentHeight {
		line := ""
		if i < len(bodyLines) {
			line = bodyLines[i]
		}
		padding := max(innerWidth-lipgloss.Width(line), 0)
		fill := strings.Repeat(" ", padding)
		rightEdge := borderStyle.Render("│")
		if scrollbar.visible && i >= scrollbar.thumbStart && i <= scrollbar.thumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		lines = append(lines, line+fill+rightEdge)
	}

	lines = append(lines, footer...)

	bottomFillWidth := max(width-1, 0)
	bottom := borderStyle.Render(strings.Repeat("─", bottomFillWidth))
	if width > 0 {
		bottom += borderStyle.Render("╯")
	}
	lines = append(lines, bottom)

	return strings.Join(lines, "\n")
}
