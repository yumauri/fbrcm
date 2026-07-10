package projects

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func renderPanel(body string, width, height int, active bool, scrollbar scrollbarState, secondary secondaryTitleState, footer []string) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	borderStyle := styles.BorderStyle(active)
	titleStyle := styles.TitleStyle(active)
	innerWidth := max(width-1, 0)
	contentHeight := max(height-2-len(footer), 0)
	topPrefixWidth := min(2, width)
	titleText := viewutil.TruncatePlain(" "+panelTitle+" ", max(width-topPrefixWidth-1, 0))
	titleWidth := lipgloss.Width(titleText)
	topPrefix := borderStyle.Render(strings.Repeat("─", topPrefixWidth))
	mainRendered := titleStyle.Render(titleText)
	rightMarginWidth := 1
	secondaryText := secondary.text
	secondaryRendered := ""
	secondaryWidth := 0
	if secondaryText != "" {
		secondaryText = " " + viewutil.TruncatePlain(secondaryText, max(width-topPrefixWidth-titleWidth-rightMarginWidth-3-1, 0)) + " "
		secondaryRendered = secondary.style.Render(secondaryText)
		secondaryWidth = lipgloss.Width(secondaryText)
	}

	gapWidth := max(width-topPrefixWidth-titleWidth-secondaryWidth-rightMarginWidth-1, 0)
	if secondaryWidth > 0 {
		gapWidth = max(gapWidth, 2)
	}
	topGap := borderStyle.Render(strings.Repeat("─", gapWidth))
	topRightFill := borderStyle.Render(strings.Repeat("─", rightMarginWidth))
	top := topPrefix + mainRendered + topGap + secondaryRendered + topRightFill + borderStyle.Render("╮")

	lines := []string{top}
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
