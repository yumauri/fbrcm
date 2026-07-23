package moveparam

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/inputstyles"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
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
		lineStyle := optionLineStyle(i == m.selected)
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
			option := m.options[optionIndex]
			content = viewutil.PadRight(option.Label, optionWidth)
			lineStyle = styledOptionLineStyle(option, i == m.selected)
		}
		lines = append(lines, left+lineStyle.Render(content)+borderStyle.Render(" │"))
	}
	lines = append(lines, borderStyle.Render(bottomLeft+strings.Repeat("─", optionWidth+2)+"╯"))
	return strings.Join(lines, "\n")
}

func styledOptionLineStyle(option Option, selected bool) lipgloss.Style {
	style := optionStyle
	if option.Foreground != nil {
		style = style.Foreground(option.Foreground)
	}
	if selected && option.KeepForegroundOnSelect {
		return style.Bold(true)
	}
	if selected {
		return optionLineStyle(true)
	}
	return style
}

func optionLineStyle(selected bool) lipgloss.Style {
	return styles.SelectionListOptionStyle(selected)
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

func newInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "New group"
	input.SetStyles(inputstyles.InlineListTextInput())
	input.Blur()
	return input
}
