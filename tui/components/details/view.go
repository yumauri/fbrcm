package details

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) View() string {
	return m.ViewWithBorder(m.active)
}

func (m Model) ViewWithBorder(borderActive bool) string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	body := strings.Split(m.viewport.View(), "\n")
	return renderPanel(body, m.width, m.height, m.active, borderActive, m.Invalid(), m.scrollbar())
}

func renderPanel(body []string, width, height int, active, borderActive, invalid bool, scrollbar scrollbarState) string {
	if width <= 1 || height <= 1 {
		return ""
	}

	borderStyle := styles.BorderStyle(borderActive)
	if invalid && borderActive {
		borderStyle = lipgloss.NewStyle().Foreground(styles.PaletteError)
	}

	panelWidth := max(width-1, 1)
	innerWidth := max(panelWidth-4, 0)
	contentHeight := max(height-2, 0)

	titleRendered, titleWidth := styles.PanelHeaderTitle(panelTitleKey(), panelTitleLabel, active, max(panelWidth-2, 0))
	topPrefixWidth := min(2, panelWidth)
	topPrefix := borderStyle.Render("╭" + strings.Repeat("─", max(topPrefixWidth-1, 0)))
	topFillWidth := max(panelWidth-topPrefixWidth-titleWidth, 0)
	topFill := borderStyle.Render(strings.Repeat("─", topFillWidth))
	lines := []string{" " + topPrefix + titleRendered + topFill}
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

type scrollbarState struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

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
