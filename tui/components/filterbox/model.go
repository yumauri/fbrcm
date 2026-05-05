package filterbox

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"fbrcm/core/filter"
	"fbrcm/tui/styles"
)

type Model struct {
	mode  filter.Mode
	input textinput.Model
}

func New() Model {
	input := textinput.New()
	input.Prompt = ""
	input.SetStyles(textinputStyles())
	input.Blur()

	return Model{
		mode:  filter.ModeFuzzy,
		input: input,
	}
}

func (m Model) Mode() filter.Mode {
	return m.mode
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m Model) Focused() bool {
	return m.input.Focused()
}

func (m Model) Visible() bool {
	return m.Focused() || m.Value() != ""
}

func (m Model) Height() int {
	if !m.Visible() {
		return 0
	}
	return 2
}

func (m *Model) Activate(mode filter.Mode) tea.Cmd {
	m.mode = mode
	m.input.CursorEnd()
	return m.input.Focus()
}

func (m *Model) Blur() {
	m.input.Blur()
}

func (m *Model) ClearAndBlur() {
	m.input.SetValue("")
	m.input.Blur()
}

func (m *Model) SetWidth(width int) {
	m.input.SetWidth(max(width-3, 1))
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.PasteMsg:
		m.input.SetValue(msg.Content)
		m.input.CursorEnd()
		return m, nil
	case tea.ClipboardMsg:
		m.input.SetValue(msg.Content)
		m.input.CursorEnd()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View(width int, active bool, count int) []string {
	if !m.Visible() || width <= 0 {
		return nil
	}

	borderStyle := styles.BorderStyle(active)
	textStyle := styles.FilterText
	innerWidth := max(width-1, 0)
	m.SetWidth(innerWidth)

	countText := textStyle.Render(" " + fmt.Sprintf("%d", count) + " ")
	countWidth := lipgloss.Width(countText)
	leftWidth := max(width-countWidth-2, 0)
	rightWidth := max(width-leftWidth-countWidth-1, 0)
	separator := borderStyle.Render(strings.Repeat("─", leftWidth)) +
		countText +
		borderStyle.Render(strings.Repeat("─", rightWidth)) +
		borderStyle.Render("┤")

	content := textStyle.Render(m.mode.Label()+" ") + m.input.View()
	content = truncateStyled(content, innerWidth)
	content += strings.Repeat(" ", max(innerWidth-lipgloss.Width(content), 0))
	content += borderStyle.Render("│")

	return []string{separator, content}
}

func ModeForKey(key string) (filter.Mode, bool) {
	return filter.ModeFromLabel(key)
}

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

func truncateStyled(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}

	var builder strings.Builder
	current := 0
	for _, r := range value {
		next := current + lipgloss.Width(string(r))
		if next > width {
			break
		}
		builder.WriteRune(r)
		current = next
	}
	return builder.String()
}
