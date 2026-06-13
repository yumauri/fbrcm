package moveparam

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

type Option struct {
	Key   string
	Label string
}

type Model struct {
	x        int
	y        int
	label    string
	options  []Option
	selected int
	input    textinput.Model
	open     bool
	search   string
	lastType time.Time
}

const (
	typeaheadTimeout  = 900 * time.Millisecond
	minGroupNameWidth = 9
)

func New() Model {
	return Model{input: newInput()}
}

// Open opens open for Model and returns the resulting state or error.
func (m Model) Open(x, y int, label string, options []Option) Model {
	m.x = x
	m.y = y
	m.label = label
	m.options = append([]Option(nil), options...)
	m.selected = 0
	m.input = newInput()
	m.open = true
	m.search = ""
	m.lastType = time.Time{}
	if m.selectedInput() {
		_ = m.input.Focus()
	}
	return m
}

// Close closes close for Model and returns the resulting state or error.
func (m Model) Close() Model {
	m.open = false
	m.label = ""
	m.options = nil
	m.selected = 0
	m.input = newInput()
	m.search = ""
	m.lastType = time.Time{}
	return m
}

// IsOpen reports open for Model and returns the resulting state or error.
func (m Model) IsOpen() bool {
	return m.open
}

func (m Model) Position() (int, int) {
	return m.x, m.y
}

func (m Model) ListPosition() (int, int) {
	connectorWidth, _ := m.layout()
	return m.x + connectorWidth + 2, m.y - m.selected
}

func (m Model) Current() (Option, bool) {
	if !m.open {
		return Option{}, false
	}
	if m.selectedInput() {
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			return Option{}, false
		}
		return Option{Key: value, Label: value}, true
	}
	optionIndex, ok := m.optionIndexForRow(m.selected)
	if !ok {
		return Option{}, false
	}
	return m.options[optionIndex], true
}

func (m Model) InputSelected() bool {
	return m.selectedInput()
}

// Move moves move for Model and returns the resulting state or error.
func (m *Model) Move(delta int) tea.Cmd {
	if m.rowsCount() == 0 {
		return nil
	}
	wasInput := m.selectedInput()
	m.selected = (m.selected + delta + m.rowsCount()) % m.rowsCount()
	m.search = ""
	m.lastType = time.Time{}
	if m.selectedInput() {
		if !wasInput {
			return m.input.Focus()
		}
		return nil
	}
	if wasInput {
		m.input.Blur()
		return nil
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.open || !m.selectedInput() {
		return nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *Model) Typeahead(key string, now time.Time) bool {
	if len(m.options) == 0 || utf8.RuneCountInString(key) != 1 || m.selectedInput() {
		return false
	}
	r, _ := utf8.DecodeRuneInString(key)
	if !unicode.IsPrint(r) {
		return false
	}

	needle := strings.ToLower(key)
	if !m.lastType.IsZero() && now.Sub(m.lastType) <= typeaheadTimeout {
		needle = m.search + needle
	}
	for row := 0; row < m.rowsCount(); row++ {
		optionIndex, ok := m.optionIndexForRow(row)
		if !ok {
			continue
		}
		if strings.HasPrefix(strings.ToLower(m.options[optionIndex].Label), needle) {
			m.selected = row
			m.search = needle
			m.lastType = now
			return true
		}
	}
	for row := 0; row < m.rowsCount(); row++ {
		optionIndex, ok := m.optionIndexForRow(row)
		if !ok {
			continue
		}
		if strings.HasPrefix(strings.ToLower(m.options[optionIndex].Label), strings.ToLower(key)) {
			m.selected = row
			m.search = strings.ToLower(key)
			m.lastType = now
			return true
		}
	}
	m.search = strings.ToLower(key)
	m.lastType = now
	return false
}

func (m Model) HeaderView() string {
	if !m.open || m.rowsCount() == 0 {
		return ""
	}

	connectorWidth, _ := m.layout()
	lines := []string{
		borderStyle.Render("╭" + strings.Repeat("─", connectorWidth+1)),
		borderStyle.Render("│ ") + parameterStyle.Render(m.label) + borderStyle.Render(" "),
		borderStyle.Render("╰" + strings.Repeat("─", connectorWidth+1)),
	}
	return strings.Join(lines, "\n")
}

func (m Model) ListView() string {
	if !m.open || m.rowsCount() == 0 {
		return ""
	}

	_, optionWidth := m.layout()
	lines := make([]string, 0, m.rowsCount()+2)
	topLeft := "╭"
	if m.selected == 0 {
		topLeft = "─"
	}
	bottomLeft := "╰"
	if m.selected == m.rowsCount()-1 {
		bottomLeft = "─"
	}
	lines = append(lines, borderStyle.Render(topLeft+strings.Repeat("─", optionWidth+2)+"╮"))
	input := m.input
	for i := 0; i < m.rowsCount(); i++ {
		left := borderStyle.Render("│ ")
		switch i {
		case m.selected:
			left = borderStyle.Render("▸ ")
		case m.selected - 1:
			left = borderStyle.Render("╯ ")
		case m.selected + 1:
			left = borderStyle.Render("╮ ")
		}
		content := ""
		if m.rowIsInput(i) {
			if i == m.selected {
				input.SetWidth(max(optionWidth-1, 1))
				content = padRenderedRight(input.View(), optionWidth)
			} else if value := strings.TrimSpace(m.input.Value()); value != "" {
				content = optionStyle.Render(padRight(value, optionWidth))
			} else {
				content = styles.PanelMuted.Render(padRight(input.Placeholder, optionWidth))
			}
		} else if optionIndex, ok := m.optionIndexForRow(i); ok {
			content = padRight(m.options[optionIndex].Label, optionWidth)
		}
		if i == m.selected {
			lines = append(lines, left+optionLineStyle(true).Render(content)+borderStyle.Render(" │"))
			continue
		}
		lines = append(lines, left+optionLineStyle(false).Render(content)+borderStyle.Render(" │"))
	}
	lines = append(lines, borderStyle.Render(bottomLeft+strings.Repeat("─", optionWidth+2)+"╯"))
	return strings.Join(lines, "\n")
}

var (
	borderStyle    = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
	parameterStyle = styles.PanelBody.Foreground(styles.PaletteBlueBright)
	optionStyle    = styles.PanelText
)

func optionLineStyle(selected bool) lipgloss.Style {
	if !selected {
		return optionStyle
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return optionStyle.Bold(true).Foreground(styles.PaletteGold)
}

func padRight(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

func padRenderedRight(value string, width int) string {
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

func (m Model) layout() (indent, optionWidth int) {
	optionWidth = minGroupNameWidth
	for _, option := range m.options {
		optionWidth = max(optionWidth, lipgloss.Width(option.Label))
	}
	optionWidth = max(optionWidth, lipgloss.Width(m.input.Placeholder))
	optionWidth = max(optionWidth, lipgloss.Width(m.input.Value())+1)
	return lipgloss.Width(m.label) + 1, optionWidth
}

func (m Model) rowsCount() int {
	if len(m.options) == 0 {
		return 0
	}
	return len(m.options) + 1
}

func (m Model) rootOptionIndex() int {
	if len(m.options) == 0 {
		return -1
	}
	last := len(m.options) - 1
	if m.options[last].Key == "" {
		return last
	}
	return -1
}

func (m Model) inputRowIndex() int {
	rootIndex := m.rootOptionIndex()
	if rootIndex >= 0 {
		return rootIndex
	}
	return len(m.options)
}

func (m Model) rowIsInput(row int) bool {
	return row == m.inputRowIndex()
}

// selectedInput selects selected input for Model and returns the resulting state or error.
func (m Model) selectedInput() bool {
	return m.rowIsInput(m.selected)
}

func (m Model) optionIndexForRow(row int) (int, bool) {
	if row < 0 || row >= m.rowsCount() || m.rowIsInput(row) {
		return 0, false
	}
	inputRow := m.inputRowIndex()
	if row > inputRow {
		return row - 1, true
	}
	return row, true
}

// moveInputStyles moves move input styles and returns the resulting value or error.
func moveInputStyles() textinput.Styles {
	inputStyles := textinput.DefaultDarkStyles()
	valueStyle := styles.PanelText
	placeholderStyle := styles.PanelMuted
	inputStyles.Focused.Text = valueStyle
	inputStyles.Focused.Prompt = valueStyle
	inputStyles.Focused.Placeholder = placeholderStyle
	inputStyles.Focused.Suggestion = valueStyle
	inputStyles.Blurred.Text = valueStyle
	inputStyles.Blurred.Prompt = valueStyle
	inputStyles.Blurred.Placeholder = placeholderStyle
	inputStyles.Blurred.Suggestion = valueStyle
	inputStyles.Cursor.Color = styles.PaletteYellow
	return inputStyles
}

func newInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "New group"
	input.SetStyles(moveInputStyles())
	input.Blur()
	return input
}
