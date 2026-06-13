package numberinput

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/styles"
)

type Model struct {
	x        int
	y        int
	maxWidth int
	minWidth int
	input    textinput.Model
	open     bool
}

func New() Model {
	return Model{input: newInput()}
}

// Open opens open for Model and returns the resulting state or error.
func (m Model) Open(x, y, minWidth, maxWidth int, value string) (Model, tea.Cmd) {
	m.x = x
	m.y = y
	m.minWidth = max(minWidth+1, 3)
	m.maxWidth = max(maxWidth, m.minWidth+2)
	m.open = true
	m.input = newInput()
	m.input.SetValue(value)
	m.input.CursorEnd()
	m.setInputWidth()
	return m, m.input.Focus()
}

// Close closes close for Model and returns the resulting state or error.
func (m Model) Close() Model {
	m.open = false
	m.input.Blur()
	m.input.SetValue("")
	return m
}

// IsOpen reports open for Model and returns the resulting state or error.
func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) Position() (int, int) {
	return m.x, m.y
}

func (m Model) Value() string {
	return strings.TrimSpace(m.input.Value())
}

func (m Model) Valid() bool {
	value := m.Value()
	if value == "" {
		return false
	}
	return core.IsJSONNumber(value)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.setInputWidth()
	return m, cmd
}

func (m Model) View() string {
	if !m.open {
		return ""
	}
	return inputBorderStyle(m.Valid()).Render(" " + m.input.View())
}

// setInputWidth sets set input width for Model and returns the resulting state or error.
func (m *Model) setInputWidth() {
	innerWidth := max(m.minWidth, lipgloss.Width(m.input.Value())+1)
	innerWidth = min(innerWidth, max(m.maxWidth-2, 1))
	m.input.SetWidth(innerWidth)
}

func inputBorderStyle(valid bool) lipgloss.Style {
	color := styles.PaletteBlueBright
	if !valid {
		color = styles.PaletteError
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color)
}

func textinputStyles() textinput.Styles {
	inputStyles := textinput.DefaultDarkStyles()
	valueStyle := styles.FilterText
	inputStyles.Focused.Text = valueStyle
	inputStyles.Focused.Prompt = valueStyle
	inputStyles.Focused.Placeholder = valueStyle
	inputStyles.Focused.Suggestion = valueStyle
	inputStyles.Blurred.Text = valueStyle
	inputStyles.Blurred.Prompt = valueStyle
	inputStyles.Blurred.Placeholder = valueStyle
	inputStyles.Blurred.Suggestion = valueStyle
	inputStyles.Cursor.Color = styles.PaletteYellow
	return inputStyles
}

func newInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetStyles(textinputStyles())
	input.Blur()
	return input
}
