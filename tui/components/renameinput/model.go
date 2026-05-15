package renameinput

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"fbrcm/tui/styles"
)

// Model holds model state used by the renameinput package.
type Model struct {
	// x stores x for Model.
	x int
	// y stores y for Model.
	y int
	// maxWidth stores max width for Model.
	maxWidth int
	// minWidth stores min width for Model.
	minWidth int
	// input stores input for Model.
	input textinput.Model
	// open stores open for Model.
	open bool
}

// New constructs new and returns the resulting value or error.
func New() Model {
	return Model{input: newInput()}
}

// Open opens open for Model and returns the resulting state or error.
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

// Position handles position for Model and returns the resulting state or error.
func (m Model) Position() (int, int) {
	return m.x, m.y
}

// Value handles value for Model and returns the resulting state or error.
func (m Model) Value() string {
	return strings.TrimSpace(m.input.Value())
}

// Update updates update for Model and returns the resulting state or error.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.open {
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.setInputWidth()
	return m, cmd
}

// View handles view for Model and returns the resulting state or error.
func (m Model) View() string {
	if !m.open {
		return ""
	}
	return inputBorderStyle.Render(" " + m.input.View())
}

// setInputWidth sets set input width for Model and returns the resulting state or error.
func (m *Model) setInputWidth() {
	innerWidth := max(m.minWidth, lipgloss.Width(m.input.Value())+1)
	innerWidth = min(innerWidth, max(m.maxWidth-2, 1))
	m.input.SetWidth(innerWidth)
}

var inputBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(styles.PaletteBlueBright)

// textinputStyles handles textinput styles and returns the resulting value or error.
func textinputStyles() textinput.Styles {
	inputStyles := textinput.DefaultDarkStyles()
	filterStyle := styles.FilterText
	inputStyles.Focused.Text = filterStyle
	inputStyles.Focused.Prompt = filterStyle
	inputStyles.Focused.Placeholder = filterStyle
	inputStyles.Focused.Suggestion = filterStyle
	inputStyles.Blurred.Text = filterStyle
	inputStyles.Blurred.Prompt = filterStyle
	inputStyles.Blurred.Placeholder = filterStyle
	inputStyles.Blurred.Suggestion = filterStyle
	inputStyles.Cursor.Color = styles.PaletteYellow
	return inputStyles
}

// newInput constructs new input and returns the resulting value or error.
func newInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetStyles(textinputStyles())
	input.Blur()
	return input
}
