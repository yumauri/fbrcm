package authpicker

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	pickerBorder = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
	pickerTitle  = lipgloss.NewStyle().Bold(true).Foreground(styles.PaletteSlateBright)
)

func (m Model) View() string {
	if !m.open || m.width <= 0 || m.height <= 0 {
		return ""
	}
	contentWidth := m.contentWidth()
	frameWidth := viewutil.PopupInnerWidth(contentWidth) + 2
	titleText := viewutil.TruncatePlain(" "+m.title+" ", max(frameWidth-6, 0))
	topFill := max(frameWidth-4-lipgloss.Width(titleText), 0)
	lines := []string{" " + pickerBorder.Render("╭─") + pickerTitle.Render(titleText) + pickerBorder.Render(strings.Repeat("─", topFill)+"─╮") + " "}
	for range viewutil.PopupPaddingTop {
		lines = append(lines, pickerFrameLine("", contentWidth))
	}
	for _, bodyLine := range m.body {
		line := ansi.Truncate(bodyLine, contentWidth, "…")
		lines = append(lines, pickerFrameLine(styles.PanelText.Render(line), contentWidth))
	}
	if len(m.body) > 0 {
		lines = append(lines, pickerFrameLine("", contentWidth))
	}

	rows := m.visibleRows()
	end := min(m.scroll+rows, len(m.options))
	for index := m.scroll; index < end; index++ {
		option := m.options[index]
		line := ansi.Truncate(optionLine(option), contentWidth, "…")
		style := styles.PanelText
		if index == m.cursor {
			style = styles.TitleStyle(true)
		}
		lines = append(lines, pickerFrameLine(style.Render(line), contentWidth))
	}
	if len(m.options) == 0 {
		lines = append(lines, pickerFrameLine(styles.PanelMuted.Render("No authentication configured"), contentWidth))
	}
	lines = append(lines, pickerFrameLine("", contentWidth))
	buttons := m.buttons.View()
	for buttonLine := range strings.SplitSeq(buttons, "\n") {
		aligned := strings.Repeat(" ", max(contentWidth-lipgloss.Width(buttonLine), 0)) + buttonLine
		lines = append(lines, pickerFrameLine(aligned, contentWidth))
	}
	lines = append(lines, " "+pickerBorder.Render("╰"+strings.Repeat("─", viewutil.PopupInnerWidth(contentWidth))+"╯")+" ")
	return strings.Join(lines, "\n")
}

func pickerFrameLine(line string, width int) string {
	return " " + pickerBorder.Render("│") + viewutil.PopupContentLine(line, width) + pickerBorder.Render("│") + " "
}
