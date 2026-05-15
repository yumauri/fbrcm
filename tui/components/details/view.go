package details

import (
	"strings"

	"charm.land/lipgloss/v2"

	"fbrcm/tui/styles"
)

// View handles view for Model and returns the resulting state or error.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	body := strings.Split(m.viewport.View(), "\n")
	return renderPanel(body, m.width, m.height, m.active, m.bridgeActive, m.Invalid(), m.scrollbar())
}

// renderPanel renders render panel and returns the resulting value or error.
func renderPanel(body []string, width, height int, active, bridgeActive, invalid bool, scrollbar scrollbarState) string {
	if width <= 1 || height <= 1 {
		return ""
	}

	borderStyle := styles.BorderStyle(active)
	if invalid {
		borderStyle = lipgloss.NewStyle().Foreground(styles.PaletteError)
	}
	titleStyle := styles.TitleStyle(active)

	panelWidth := max(width-1, 1)
	innerWidth := max(panelWidth-4, 0)
	contentHeight := max(height-2, 0)

	titleText := truncatePlain(" "+panelTitle+" ", max(panelWidth-2, 0))
	titleWidth := lipgloss.Width(titleText)
	topPrefixWidth := min(2, panelWidth)
	topPrefix := borderStyle.Render("╭" + strings.Repeat("─", max(topPrefixWidth-1, 0)))
	topFillWidth := max(panelWidth-topPrefixWidth-titleWidth, 0)
	topFill := borderStyle.Render(strings.Repeat("─", topFillWidth))
	lines := []string{" " + topPrefix + titleStyle.Render(titleText) + topFill}
	for i := range contentHeight {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		padding := max(innerWidth-lipgloss.Width(line), 0)
		scrollCell := " "
		if scrollbar.visible {
			if i >= scrollbar.thumbStart && i <= scrollbar.thumbEnd {
				scrollCell = styles.ScrollbarThumb.Render("█")
			}
		}
		lines = append(lines, " "+borderStyle.Render("│")+" "+line+strings.Repeat(" ", padding)+" "+scrollCell)
	}

	lines = append(lines, " "+borderStyle.Render("╰"+strings.Repeat("─", max(panelWidth-1, 0))))
	return strings.Join(lines, "\n")
}

// truncatePlain handles truncate plain and returns the resulting value or error.
func truncatePlain(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}

	return string(runes[:width])
}

// scrollbarState holds scrollbar state state used by the details package.
type scrollbarState struct {
	// visible stores visible for scrollbarState.
	visible bool
	// thumbStart stores thumb start for scrollbarState.
	thumbStart int
	// thumbEnd stores thumb end for scrollbarState.
	thumbEnd int
}

// scrollbar handles scrollbar for Model and returns the resulting state or error.
func (m Model) scrollbar() scrollbarState {
	contentHeight := max(m.height-2, 1)
	totalLines := m.viewport.TotalLineCount()
	if contentHeight <= 0 || totalLines <= contentHeight {
		return scrollbarState{}
	}

	thumbHeight := max(2, (contentHeight*contentHeight)/totalLines)
	thumbHeight = min(thumbHeight, contentHeight)

	maxOffset := max(totalLines-contentHeight, 1)
	maxThumbStart := max(contentHeight-thumbHeight, 0)
	thumbStart := (m.viewport.YOffset() * maxThumbStart) / maxOffset

	return scrollbarState{
		visible:    true,
		thumbStart: thumbStart,
		thumbEnd:   min(thumbStart+thumbHeight-1, contentHeight-1),
	}
}
