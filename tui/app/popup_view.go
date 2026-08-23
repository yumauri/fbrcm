package app

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

var popupBorderStyle = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)

func popupTopBorder(key, title string, width int) string {
	frameInner := viewutil.PopupInnerWidth(width)
	titleView, titleWidth := styles.PanelHeaderTab(key, title, true, true, max(frameInner-1, 0))
	fill := max(frameInner-titleWidth-1, 0)
	return popupBorderStyle.Render("╭─") + titleView + popupBorderStyle.Render(strings.Repeat("─", fill)+"╮")
}

func popupSeparator(width int) string {
	return popupBorderStyle.Render("├" + strings.Repeat("─", viewutil.PopupInnerWidth(width)) + "┤")
}

func popupBottomBorder(width int) string {
	return popupBorderStyle.Render("╰" + strings.Repeat("─", viewutil.PopupInnerWidth(width)) + "╯")
}

func popupLine(content string, width int) string {
	return popupBorderStyle.Render("│") + viewutil.PopupContentLine(content, width) + popupBorderStyle.Render("│")
}
