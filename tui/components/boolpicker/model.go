package boolpicker

import (
	"strings"

	"charm.land/lipgloss/v2"

	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/styles"
)

type Model struct {
	x        int
	y        int
	values   []string
	selected int
	open     bool
}

func New() Model {
	return Model{}
}

// Open opens open for Model and returns the resulting state or error.
func (m Model) Open(x, y int, current bool) Model {
	m.x = x
	m.y = y
	if current {
		m.values = []string{"true", "false"}
	} else {
		m.values = []string{"false", "true"}
	}
	m.selected = 0
	m.open = true
	return m
}

// Close closes close for Model and returns the resulting state or error.
func (m Model) Close() Model {
	m.open = false
	m.values = nil
	m.selected = 0
	return m
}

// IsOpen reports open for Model and returns the resulting state or error.
func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) Position() (int, int) {
	return m.x, m.y - m.selected - 1
}

// Move moves move for Model and returns the resulting state or error.
func (m *Model) Move(delta int) {
	if len(m.values) == 0 {
		return
	}
	m.selected = (m.selected + delta + len(m.values)) % len(m.values)
}

func (m Model) Current() (bool, bool) {
	if !m.open || m.selected < 0 || m.selected >= len(m.values) {
		return false, false
	}
	return strings.EqualFold(m.values[m.selected], "true"), true
}

func (m Model) CurrentString() (string, bool) {
	if !m.open || m.selected < 0 || m.selected >= len(m.values) {
		return "", false
	}
	return m.values[m.selected], true
}

func (m Model) Changed() bool {
	return m.open && m.selected > 0
}

func (m Model) View() string {
	if !m.open || len(m.values) == 0 {
		return ""
	}
	width := max(lipgloss.Width(m.values[0]), lipgloss.Width(m.values[1]))
	lines := []string{
		borderStyle.Render("╭" + strings.Repeat("─", width+2) + "╮"),
		m.renderRow(0, width),
		m.renderRow(1, width),
		borderStyle.Render("╰" + strings.Repeat("─", width+2) + "╯"),
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderRow(index, width int) string {
	left := borderStyle.Render("│ ")
	if index == m.selected {
		left = borderStyle.Render("▸ ")
	}
	return left + valueStyle(m.values[index]).Render(padRight(m.values[index], width)) + borderStyle.Render(" │")
}

var borderStyle = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)

func valueStyle(value string) lipgloss.Style {
	style := corestyles.ValueTextStyle(value, "boolean")
	if styles.NoColorEnabled() {
		if strings.EqualFold(value, "true") || strings.EqualFold(value, "false") {
			return lipgloss.NewStyle().Bold(true)
		}
		return lipgloss.NewStyle()
	}
	return style
}

func padRight(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}
