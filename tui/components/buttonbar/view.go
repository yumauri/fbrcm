package buttonbar

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	buttonStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.PaletteSlateDark).Foreground(styles.PaletteSlateBright).Padding(0, 1)
	dangerFocusStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.PaletteError).Foreground(styles.PaletteError).Bold(true).Padding(0, 1)
	accentFocusStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(styles.PaletteBlueBright).Foreground(styles.PaletteBlueBright).Bold(true).Padding(0, 1)
)

func (m Model) View() string {
	items := m.rendered()
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		if i > 0 {
			parts = append(parts, " ")
		}
		parts = append(parts, item)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m Model) rendered() []string {
	out := make([]string, 0, len(m.buttons))
	for i, button := range m.buttons {
		label := button.Label
		style := buttonStyle
		if m.focused && i == m.selected {
			if styles.NoColorEnabled() {
				label = lipgloss.NewStyle().Bold(true).Reverse(true).Render(button.Label)
			} else if button.Variant == VariantDanger {
				style = dangerFocusStyle
			} else {
				style = accentFocusStyle
			}
		}
		out = append(out, style.Render(label))
	}
	return out
}

func printableWidth(value string) int { return lipgloss.Width(value) }
func renderedHeight(value string) int { return len(strings.Split(value, "\n")) }
