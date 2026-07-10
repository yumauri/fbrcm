package jsoninput

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderBox() string {
	border := borderStyle(m.Valid())
	body := strings.Split(m.renderArea(), "\n")
	innerWidth := max(m.screenW-6, 4)
	contentHeight := jsonContentHeight(m.screenH)
	scrollbar := expandedScrollbarState(m.visualLineCount(), m.area.ScrollYOffset(), contentHeight)

	lines := []string{border.Render("╭" + strings.Repeat("─", innerWidth) + "╮")}
	for i := range contentHeight {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		rightEdge := border.Render("│")
		if scrollbar.visible && i >= scrollbar.thumbStart && i <= scrollbar.thumbEnd {
			rightEdge = styles.ScrollbarThumb.Render("█")
		}
		if line == "" {
			line = strings.Repeat(" ", innerWidth)
		}
		lines = append(lines, border.Render("│")+line+rightEdge)
	}
	lines = append(lines, border.Render("│")+renderHelpFooter(jsonHelpText(innerWidth), innerWidth)+border.Render("│"))
	lines = append(lines, border.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(lines, "\n")
}

func (m Model) visualLineCount() int {
	lines := strings.Split(m.area.Value(), "\n")
	if len(lines) == 0 {
		return 1
	}
	gutter := lineNumberGutter(len(lines))
	contentWidth := max(max(m.screenW-6, 4)-gutter, 1)
	count := 0
	for _, line := range lines {
		count += len(wrapPlainLine(line, contentWidth))
	}
	return max(count, 1)
}

func jsonContentHeight(screenH int) int {
	return max(screenH-7, 3)
}

func jsonHelpText(width int) string {
	m := help.New()
	m.ShortSeparator = " • "
	m.Styles.ShortKey = styles.FilterText
	m.Styles.ShortDesc = styles.PanelMuted
	m.Styles.ShortSeparator = styles.PanelMuted
	m.Styles.Ellipsis = styles.PanelMuted
	m.SetWidth(width)
	return m.ShortHelpView([]key.Binding{
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionSave, "save"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionFormat, "format"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionCancel, "cancel"),
		tuiconfig.Binding(tuiconfig.BlockJSONInput, tuiconfig.ActionCopyValue, "copy"),
	})
}

func renderHelpFooter(text string, width int) string {
	if width <= 0 {
		return ""
	}
	return text + strings.Repeat(" ", max(width-lipgloss.Width(text), 0))
}

type expandedScrollbar struct {
	visible    bool
	thumbStart int
	thumbEnd   int
}

func expandedScrollbarState(total, offset, visible int) expandedScrollbar {
	if visible <= 0 {
		return expandedScrollbar{}
	}
	if total <= visible {
		return expandedScrollbar{}
	}
	thumbHeight := max(1, (visible*visible)/total)
	maxThumbStart := visible - thumbHeight
	maxOffset := max(total-visible, 1)
	thumbStart := (min(offset, maxOffset) * maxThumbStart) / maxOffset
	return expandedScrollbar{
		visible:    true,
		thumbStart: thumbStart,
		thumbEnd:   thumbStart + thumbHeight - 1,
	}
}
