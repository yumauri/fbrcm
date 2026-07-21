package authpicker

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/buttonbar"
)

type Option struct {
	Key    string
	Label  string
	Detail string
}

type Model struct {
	x, y          int
	width, height int
	title         string
	body          []string
	options       []Option
	cursor        int
	scroll        int
	open          bool
	buttons       buttonbar.Model
}

func New() Model {
	return Model{buttons: newButtonBar()}
}

func (m Model) SetBounds(x, y, width, height int) Model {
	m.x, m.y, m.width, m.height = x, y, width, height
	m.ensureVisible()
	return m
}

func (m Model) Open(title string, body []string, options []Option, selected int) Model {
	m.title = title
	m.body = append([]string(nil), body...)
	m.options = append([]Option(nil), options...)
	m.cursor = min(max(selected, 0), max(len(options)-1, 0))
	m.scroll = 0
	m.open = true
	m.buttons = newButtonBar()
	m.ensureVisible()
	return m
}

func (m Model) Close() Model {
	m.title = ""
	m.body = nil
	m.options = nil
	m.cursor = 0
	m.scroll = 0
	m.open = false
	m.buttons = newButtonBar()
	return m
}

func (m Model) IsOpen() bool { return m.open }

func (m Model) Current() (Option, bool) {
	if !m.open || m.cursor < 0 || m.cursor >= len(m.options) {
		return Option{}, false
	}
	return m.options[m.cursor], true
}

func (m *Model) Move(delta int) {
	if len(m.options) == 0 {
		return
	}
	m.cursor = (m.cursor + delta + len(m.options)) % len(m.options)
	m.ensureVisible()
}

func (m *Model) MoveButton(delta int) {
	m.buttons.Move(delta)
}

func (m Model) SelectedButton() int {
	return m.buttons.Selected()
}

func (m *Model) SelectButtonAt(x, y int) bool {
	index, ok := m.buttonIndexAt(x, y)
	if !ok {
		return false
	}
	m.buttons = m.buttons.SetSelected(index)
	return true
}

func (m Model) Position() (int, int) {
	w, h := m.boxSize()
	return max(m.x+(m.width-w)/2, m.x), max(m.y+(m.height-h)/2, m.y)
}

func (m Model) visibleRows() int {
	return min(len(m.options), max(m.height-len(m.body)-8, 1))
}

func (m *Model) ensureVisible() {
	rows := m.visibleRows()
	if m.cursor < m.scroll {
		m.scroll = m.cursor
	}
	if m.cursor >= m.scroll+rows {
		m.scroll = m.cursor - rows + 1
	}
	m.scroll = max(0, min(m.scroll, max(len(m.options)-rows, 0)))
}

func (m Model) boxSize() (int, int) {
	contentWidth := 0
	for _, line := range m.body {
		contentWidth = max(contentWidth, lipgloss.Width(line))
	}
	for _, option := range m.options {
		contentWidth = max(contentWidth, lipgloss.Width(optionLine(option)))
	}
	contentWidth = max(contentWidth, lipgloss.Width(m.buttons.View()))
	contentWidth = max(contentWidth, lipgloss.Width(m.title)+2)
	contentWidth = min(contentWidth, max(m.width-7, 1))
	bodyHeight := len(m.body) + max(m.visibleRows(), 1)
	extra := 0
	if len(m.body) > 0 {
		extra = 1
	}
	bodyHeight += extra
	return contentWidth + 7, bodyHeight + lipgloss.Height(m.buttons.View()) + 4
}

func (m Model) contentWidth() int {
	width, _ := m.boxSize()
	return max(width-7, 1)
}

func (m Model) bodyHeight() int {
	height := len(m.body) + max(m.visibleRows(), 1)
	if len(m.body) > 0 {
		height++
	}
	return height
}

func (m Model) buttonIndexAt(x, y int) (int, bool) {
	boxX, boxY := m.Position()
	buttons := m.buttons.View()
	buttonLines := strings.Split(buttons, "\n")
	buttonX := boxX + 4 + max(m.contentWidth()-lipgloss.Width(buttonLines[0]), 0)
	buttonY := boxY + m.bodyHeight() + 3
	if y < buttonY || y >= buttonY+len(buttonLines) {
		return -1, false
	}
	return m.buttons.IndexAt(x-buttonX, y-buttonY)
}

func newButtonBar() buttonbar.Model {
	return buttonbar.New([]buttonbar.Button{
		{Label: "Bind", Variant: buttonbar.VariantAccent},
		{Label: "Cancel", Variant: buttonbar.VariantNeutral},
	}).SetFocused(true)
}

func optionLine(option Option) string {
	if option.Detail == "" {
		return option.Label
	}
	return option.Label + " · " + option.Detail
}
