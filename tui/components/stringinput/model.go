package stringinput

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/inputstyles"
)

type Model struct {
	x         int
	y         int
	minWidth  int
	maxWidth  int
	screenW   int
	screenH   int
	fullWidth bool
	expanded  bool
	text      textinput.Model
	area      textarea.Model
	open      bool
}

func New() Model {
	return Model{text: inputstyles.NewTextInput(), area: inputstyles.NewTextarea()}
}

func (m Model) Open(x, y, minWidth, maxWidth, screenW, screenH int, value string, fullWidth, expanded bool) (Model, tea.Cmd) {
	m.x = x
	m.y = y
	m.minWidth = max(minWidth, 15)
	m.maxWidth = max(maxWidth, 1)
	m.screenW = screenW
	m.screenH = screenH
	m.fullWidth = fullWidth
	m.expanded = expanded
	m.open = true
	m.text = inputstyles.NewTextInput()
	m.text.SetValue(value)
	m.text.CursorEnd()
	m.area = inputstyles.NewTextarea()
	m.area.SetValue(value)
	m.resize()
	m.resetAreaCursor()
	if m.expanded {
		return m, m.area.Focus()
	}
	return m, m.text.Focus()
}

func (m Model) Close() Model {
	m.open = false
	m.text.Blur()
	m.area.Blur()
	m.text.SetValue("")
	m.area.SetValue("")
	return m
}

func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) IsExpanded() bool {
	return m.expanded
}

func (m Model) Position() (int, int) {
	if m.expanded {
		return 2, 2
	}
	if m.fullWidth {
		return 0, m.y
	}
	return m.x, m.y
}

func (m Model) Value() string {
	if m.expanded {
		return m.area.Value()
	}
	return m.text.Value()
}

func (m Model) CanCollapse() bool {
	return !strings.Contains(m.Value(), "\n")
}

func (m Model) ToggleExpanded() (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	if m.expanded {
		if !m.CanCollapse() {
			return m, nil
		}
		value := m.area.Value()
		cursorCol := m.area.Column()
		m.expanded = false
		m.text = inputstyles.NewTextInput()
		m.text.SetValue(value)
		m.text.SetCursor(min(cursorCol, len([]rune(value))))
		m.resize()
		return m, m.text.Focus()
	}
	value := m.text.Value()
	cursorCol := m.text.Position()
	m.expanded = true
	m.area = inputstyles.NewTextarea()
	m.area.SetValue(value)
	m.resize()
	m.setAreaCursor(0, cursorCol)
	return m, m.area.Focus()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	if m.expanded {
		var cmd tea.Cmd
		m.area, cmd = m.area.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.text, cmd = m.text.Update(msg)
	m.resize()
	return m, cmd
}

func (m Model) View() string {
	if !m.open {
		return ""
	}
	if m.expanded {
		return m.renderExpandedBox()
	}
	return singleBorderStyle.Render(" " + m.text.View())
}

func (m *Model) resetAreaCursor() {
	for m.area.Line() > 0 {
		m.area.CursorUp()
	}
	m.area.CursorStart()
}

func (m *Model) setAreaCursor(line, col int) {
	m.resetAreaCursor()
	for m.area.Line() < line {
		m.area.CursorDown()
	}
	m.area.SetCursorColumn(col)
}
