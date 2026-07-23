package stringinput

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
)

func (m *Model) resize() {
	if m.expanded {
		innerWidth := stringPopupContentWidth(m.screenW)
		innerHeight := stringContentHeight(m.screenH)
		gutter := lineNumberGutter(m.area.LineCount())
		m.area.SetWidth(max(innerWidth-gutter, 1))
		m.area.SetHeight(innerHeight)
		return
	}
	innerWidth := max(m.minWidth, lipgloss.Width(m.text.Value())+1)
	if m.fullWidth {
		innerWidth = max(m.screenW-4, 1)
	} else {
		innerWidth = min(innerWidth, max(m.maxWidth-4, 1))
	}
	pos := m.text.Position()
	m.text.SetWidth(innerWidth)
	m.text.SetCursor(pos)
}

func (m Model) visualLineCount() int {
	lines := strings.Split(m.area.Value(), "\n")
	if len(lines) == 0 {
		return 1
	}
	gutter := lineNumberGutter(len(lines))
	contentWidth := max(stringPopupContentWidth(m.screenW)-gutter, 1)
	count := 0
	for _, line := range lines {
		count += len(wrapLine(line, contentWidth))
	}
	return max(count, 1)
}

func stringPopupContentWidth(screenW int) int {
	return max(max(screenW-6, 4)-viewutil.PopupPaddingLeft-viewutil.PopupPaddingRight, 1)
}

func stringContentHeight(screenH int) int {
	return max(screenH-7-viewutil.PopupPaddingTop, 3)
}
