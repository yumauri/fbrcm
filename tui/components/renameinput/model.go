package renameinput

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/inputstyles"
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

func (m Model) Open(x, y, minWidth, maxWidth int, value string) (Model, tea.Cmd) {
	m.x = x
	m.y = y
	m.minWidth = max(minWidth+1, 1)
	m.maxWidth = max(maxWidth, m.minWidth+2)
	m.open = true
	m.input = newInput()
	m.input.SetValue(value)
	m.input.CursorEnd()
	m.setInputWidth()
	return m, m.input.Focus()
}

func (m Model) Close() Model {
	m.open = false
	m.input.Blur()
	m.input.SetValue("")
	return m
}

func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) Position() (int, int) {
	return m.x, m.y
}

func (m Model) Value() string {
	return strings.TrimSpace(m.input.Value())
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
	return inputBorderStyle.Render(" " + m.input.View())
}

func (m *Model) setInputWidth() {
	innerWidth := max(m.minWidth, lipgloss.Width(m.input.Value())+1)
	innerWidth = min(innerWidth, max(m.maxWidth-2, 1))
	m.input.SetWidth(innerWidth)
}

var inputBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(styles.PaletteBlueBright)

func newInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetStyles(inputstyles.TextInput())
	input.Blur()
	return input
}
