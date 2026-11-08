package viewutil

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

// SidebarHeader draws the open-left frame used by sidebar panels.
func SidebarHeader(width int, key, title string, active bool, border lipgloss.Style, badge string, badgeStyle lipgloss.Style) string {
	return panelHeader(width, "──", border, badge, badgeStyle, func(limit int) (string, int) {
		return styles.PanelHeaderTitle(key, title, active, limit)
	})
}

// PopupHeader draws a popup title with the same optional right-aligned badge.
func PopupHeader(width int, key, title string, active bool, border lipgloss.Style, badge string, badgeStyle lipgloss.Style) string {
	return panelHeader(width, "╭─", border, badge, badgeStyle, func(limit int) (string, int) {
		return styles.PanelHeaderTab(key, title, true, active, limit)
	})
}

func panelHeader(width int, prefix string, border lipgloss.Style, badge string, badgeStyle lipgloss.Style, renderTitle func(int) (string, int)) string {
	if width <= 0 {
		return ""
	}
	prefixWidth := min(2, width-1)
	available := width - prefixWidth - 1
	badgeWidth := 0
	if badge != "" && available >= 7 {
		badge = " " + TruncatePlain(badge, available-6) + " "
		badgeWidth = lipgloss.Width(badge)
	}
	titleLimit := available
	if badgeWidth > 0 {
		titleLimit -= badgeWidth + 3
	}
	renderedTitle, titleWidth := renderTitle(titleLimit)
	fill := available - titleWidth - badgeWidth
	tail := border.Render(strings.Repeat("─", fill) + "╮")
	if badgeWidth > 0 {
		tail = border.Render(strings.Repeat("─", fill-1)) + badgeStyle.Render(badge) + border.Render("─╮")
	}
	return border.Render(TruncatePlain(prefix, prefixWidth)) + renderedTitle + tail
}
