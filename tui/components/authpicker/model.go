package authpicker

import "charm.land/lipgloss/v2"

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
}

func New() Model { return Model{} }

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
	contentWidth := lipgloss.Width(m.title)
	for _, line := range m.body {
		contentWidth = max(contentWidth, lipgloss.Width(line))
	}
	for _, option := range m.options {
		line := option.Label
		if option.Detail != "" {
			line += "  ·  " + option.Detail
		}
		contentWidth = max(contentWidth, lipgloss.Width(line))
	}
	contentWidth = min(max(contentWidth, 32), max(m.width-8, 1))
	extra := 0
	if len(m.body) > 0 {
		extra = len(m.body) + 1
	}
	return contentWidth + 4, m.visibleRows() + 5 + extra
}
