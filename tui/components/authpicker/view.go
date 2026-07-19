package authpicker

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

var pickerBorder = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)

func (m Model) View() string {
	if !m.open || m.width <= 0 || m.height <= 0 {
		return ""
	}
	boxWidth, _ := m.boxSize()
	inner := max(boxWidth-4, 1)
	title, titleWidth := styles.PanelHeaderTitle("", m.title, true, max(inner-1, 0))
	lines := []string{pickerBorder.Render("╭─") + title + pickerBorder.Render(strings.Repeat("─", max(inner-titleWidth-1, 0))+"╮")}
	for _, bodyLine := range m.body {
		line := ansi.Truncate(bodyLine, inner-2, "…")
		line = "  " + viewutil.PadRight(line, max(inner-2, 0))
		lines = append(lines, pickerBorder.Render("│")+styles.PanelText.Render(line)+pickerBorder.Render("│"))
	}
	if len(m.body) > 0 {
		lines = append(lines, pickerBorder.Render("│")+styles.PanelMuted.Render(strings.Repeat(" ", inner))+pickerBorder.Render("│"))
	}

	rows := m.visibleRows()
	end := min(m.scroll+rows, len(m.options))
	for index := m.scroll; index < end; index++ {
		option := m.options[index]
		line := option.Label
		if option.Detail != "" {
			line += "  ·  " + option.Detail
		}
		line = ansi.Truncate(line, inner-2, "…")
		line = "  " + viewutil.PadRight(line, max(inner-2, 0))
		style := styles.PanelText
		if index == m.cursor {
			style = styles.TitleStyle(true)
		}
		lines = append(lines, pickerBorder.Render("│")+style.Render(line)+pickerBorder.Render("│"))
	}
	if len(m.options) == 0 {
		lines = append(lines, pickerBorder.Render("│")+styles.PanelMuted.Render(viewutil.PadRight("  No authentication configured", inner))+pickerBorder.Render("│"))
	}
	help := styles.FilterText.Render("enter") + styles.PanelMuted.Render(" bind  •  ") + styles.FilterText.Render("esc") + styles.PanelMuted.Render(" cancel")
	help = ansi.Truncate(help, inner, "")
	lines = append(lines,
		pickerBorder.Render("│")+styles.PanelMuted.Render(strings.Repeat(" ", inner))+pickerBorder.Render("│"),
		pickerBorder.Render("│")+help+strings.Repeat(" ", max(inner-lipgloss.Width(help), 0))+pickerBorder.Render("│"),
		pickerBorder.Render("╰"+strings.Repeat("─", inner)+"╯"),
	)
	return strings.Join(lines, "\n")
}
