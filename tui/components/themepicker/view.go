package themepicker

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

const firstOptionRow = 2

var pickerBorder = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)

func (m Model) View() string {
	if !m.open || m.width <= 0 || m.height <= 0 {
		return ""
	}
	contentWidth := m.contentWidth()
	lines := []string{pickerTopBorder(contentWidth)}
	lines = append(lines, pickerFrameLine("", contentWidth))

	rows := m.visibleRows()
	end := min(m.scroll+rows, len(m.options))
	for index := m.scroll; index < end; index++ {
		line := m.optionLine(m.options[index], contentWidth, index == m.cursor)
		lines = append(lines, pickerFrameLine(line, contentWidth))
	}
	if len(m.options) == 0 {
		lines = append(lines, pickerFrameLine(styles.PanelMuted.Render("No themes available"), contentWidth))
	}

	lines = append(lines, pickerFrameLine("", contentWidth))
	if m.saving {
		lines = append(lines, pickerFrameLine(styles.PanelMuted.Render("Saving…"), contentWidth))
	} else if m.errorText != "" {
		lines = append(lines, pickerFrameLine(styles.PanelMuted.Render(ansi.Truncate(m.errorText, contentWidth, "…")), contentWidth))
	}
	lines = append(lines, pickerFrameLine(pickerHelpLine(contentWidth), contentWidth))
	lines = append(lines, " "+pickerBorder.Render("╰"+strings.Repeat("─", pickerInnerWidth(contentWidth))+"╯")+" ")
	return strings.Join(lines, "\n")
}

func pickerTopBorder(contentWidth int) string {
	frameInner := pickerInnerWidth(contentWidth)
	hint := tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionThemes)
	if len([]rune(hint)) > 1 {
		hint += " "
	}
	title, titleWidth := styles.PanelHeaderTab(hint, "Themes", true, true, max(frameInner-1, 0))
	fill := max(frameInner-titleWidth-1, 0)
	return " " + pickerBorder.Render("╭─") + title + pickerBorder.Render(strings.Repeat("─", fill)+"╮") + " "
}

func pickerHelpLine(width int) string {
	return viewutil.ShortHelpView(width,
		viewutil.HelpBinding("↑/↓", "preview"),
		viewutil.HelpBinding("enter", "select"),
		viewutil.HelpBinding("esc", "close"),
	)
}

func (m Model) optionLine(option Option, width int, selected bool) string {
	selectionStyle := styles.TitleStyle(true)
	nameWidth := width - 2 - paletteWidth(option)
	if nameWidth < 1 {
		name := ansi.Truncate(option.Name, width, "…")
		if selected {
			return selectionStyle.Render(name)
		}
		return name
	}
	name := ansi.Truncate(option.Name, nameWidth, "…")
	gap := strings.Repeat(" ", max(width-lipgloss.Width(name)-paletteWidth(option), 2))
	name += gap
	if selected {
		name = selectionStyle.Render(name)
	}
	return name + palettePreview(option, selected)
}

func palettePreview(option Option, selected bool) string {
	selectionStyle := styles.TitleStyle(true)
	if option.Err != nil || option.Palette == nil {
		unavailableStyle := styles.PanelMuted
		if selected {
			unavailableStyle = unavailableStyle.Background(selectionStyle.GetBackground())
		}
		return unavailableStyle.Render("unavailable")
	}
	if styles.NoColorEnabled() {
		return ""
	}
	var preview strings.Builder
	for _, token := range corestyles.PreviewTokens() {
		swatchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(option.Palette[token]))
		if selected {
			swatchStyle = swatchStyle.Background(selectionStyle.GetBackground())
		}
		preview.WriteString(swatchStyle.Render(strings.Repeat(paletteSwatchGlyph, paletteSwatchWidth)))
	}
	return preview.String()
}

func pickerFrameLine(line string, width int) string {
	return " " + pickerBorder.Render("│") + viewutil.PopupContentLine(line, width) + " " + pickerBorder.Render("│") + " "
}

func pickerInnerWidth(contentWidth int) int {
	return viewutil.PopupInnerWidth(contentWidth) + 1
}
