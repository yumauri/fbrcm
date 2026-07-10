package moveparam

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
	"strings"
)

var (
	borderStyle    = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
	parameterStyle = styles.PanelBody.Foreground(styles.PaletteBlueBright)
	optionStyle    = styles.PanelText
)

func (m Model) HeaderView() string {
	if !m.open || m.rowsCount() == 0 {
		return ""
	}
	connectorWidth, _ := m.layout()
	return strings.Join([]string{borderStyle.Render("╭" + strings.Repeat("─", connectorWidth+1)), borderStyle.Render("│ ") + parameterStyle.Render(m.label) + borderStyle.Render(" "), borderStyle.Render("╰" + strings.Repeat("─", connectorWidth+1))}, "\n")
}

func (m Model) ListView() string {
	if !m.open || m.rowsCount() == 0 {
		return ""
	}
	_, optionWidth := m.layout()
	lines := make([]string, 0, m.rowsCount()+2)
	topLeft, bottomLeft := "╭", "╰"
	if m.selected == 0 {
		topLeft = "─"
	}
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
				content = optionStyle.Render(viewutil.PadRight(value, optionWidth))
			} else {
				content = styles.PanelMuted.Render(viewutil.PadRight(input.Placeholder, optionWidth))
			}
		} else if optionIndex, ok := m.optionIndexForRow(i); ok {
			content = viewutil.PadRight(m.options[optionIndex].Label, optionWidth)
		}
		lines = append(lines, left+optionLineStyle(i == m.selected).Render(content)+borderStyle.Render(" │"))
	}
	lines = append(lines, borderStyle.Render(bottomLeft+strings.Repeat("─", optionWidth+2)+"╯"))
	return strings.Join(lines, "\n")
}

func optionLineStyle(selected bool) lipgloss.Style {
	if !selected {
		return optionStyle
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return optionStyle.Bold(true).Foreground(styles.PaletteGold)
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

func moveInputStyles() textinput.Styles {
	s := textinput.DefaultDarkStyles()
	value, placeholder := styles.PanelText, styles.PanelMuted
	s.Focused.Text = value
	s.Focused.Prompt = value
	s.Focused.Placeholder = placeholder
	s.Focused.Suggestion = value
	s.Blurred.Text = value
	s.Blurred.Prompt = value
	s.Blurred.Placeholder = placeholder
	s.Blurred.Suggestion = value
	s.Cursor.Color = styles.PaletteYellow
	return s
}
func newInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "New group"
	input.SetStyles(moveInputStyles())
	input.Blur()
	return input
}
