package jsoninput

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderBox() string {
	border := borderStyle(m.Valid())
	body := strings.Split(m.renderArea(), "\n")
	contentWidth := jsonPopupContentWidth(m.screenW)
	innerWidth := viewutil.PopupInnerWidth(contentWidth)
	contentHeight := jsonContentHeight(m.screenH)
	scrollbar := viewutil.ScrollbarState(m.visualLineCount(), m.area.ScrollYOffset(), contentHeight)

	lines := []string{border.Render("╭" + strings.Repeat("─", innerWidth) + "╮")}
	for range viewutil.PopupPaddingTop {
		lines = append(lines, border.Render("│")+viewutil.PopupContentLine("", contentWidth)+border.Render("│"))
	}
	for i := range contentHeight {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		rightEdge := border.Render("│")
		if scrollbar.Visible && i >= scrollbar.ThumbStart && i <= scrollbar.ThumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		lines = append(lines, border.Render("│")+viewutil.PopupContentLine(line, contentWidth)+rightEdge)
	}
	lines = append(lines, border.Render("│")+viewutil.PopupContentLine(renderHelpFooter(jsonHelpText(contentWidth), contentWidth), contentWidth)+border.Render("│"))
	lines = append(lines, border.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(lines, "\n")
}

func (m Model) visualLineCount() int {
	lines := strings.Split(m.area.Value(), "\n")
	if len(lines) == 0 {
		return 1
	}
	gutter := lineNumberGutter(len(lines))
	contentWidth := max(jsonPopupContentWidth(m.screenW)-gutter, 1)
	count := 0
	for _, line := range lines {
		count += len(wrapPlainLine(line, contentWidth))
	}
	return max(count, 1)
}

func jsonPopupContentWidth(screenW int) int {
	return max(max(screenW-6, 4)-viewutil.PopupPaddingLeft-viewutil.PopupPaddingRight, 1)
}

func jsonContentHeight(screenH int) int {
	return max(screenH-7-viewutil.PopupPaddingTop, 3)
}

func jsonHelpText(width int) string {
	return viewutil.ShortHelpView(width,
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionSave, "save"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionFormat, "format"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionCancel, "cancel"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionCopyValue, "copy"),
	)
}

func renderHelpFooter(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}
