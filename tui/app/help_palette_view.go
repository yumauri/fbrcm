package app

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	helpPaletteBorderStyle   = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
	helpPaletteGroupStyle    = styles.FilterText.Bold(true)
	helpPaletteDisabledStyle = styles.PanelMuted.Italic(true)
)

func (m Model) helpPaletteView() string {
	if !m.helpPalette.IsOpen() || m.width <= 0 || m.height <= 0 {
		return ""
	}
	actions := m.helpPalette.filtered(m.helpPaletteActions())
	listHeight := helpPaletteListHeight(m.height)
	boxWidth := min(max(m.width-8, 56), 100)
	contentWidth := max(boxWidth-2-viewutil.PopupPaddingLeft-viewutil.PopupPaddingRight, 1)

	input := m.helpPalette.input
	count := rcdisplay.FormatCount(len(actions), "action", "actions")
	input.SetWidth(max(contentWidth-lipgloss.Width(count)-1, 1))
	search := input.View()
	search += strings.Repeat(" ", max(contentWidth-lipgloss.Width(search)-lipgloss.Width(count)-1, 0))
	search += styles.PanelMuted.Render(count + " ")

	lines := []string{helpPaletteTopBorder(contentWidth)}
	for range viewutil.PopupPaddingTop {
		lines = append(lines, helpPaletteLine("", contentWidth))
	}
	lines = append(lines, helpPaletteLine(search, contentWidth), helpPaletteSeparator(contentWidth))
	lines = append(lines, m.helpPaletteActionLines(actions, listHeight, contentWidth)...)
	lines = append(lines, helpPaletteSeparator(contentWidth))
	status := "Select an action to see what it does"
	if len(actions) > 0 && m.helpPalette.cursor >= 0 && m.helpPalette.cursor < len(actions) {
		selected := actions[m.helpPalette.cursor]
		status = selected.description
		if !selected.enabled {
			status += " Unavailable: " + selected.reason
		}
	}
	lines = append(lines, helpPaletteLine(styles.PanelMuted.Render(status), contentWidth))
	footer := m.helpPaletteFooter(contentWidth)
	lines = append(lines, helpPaletteLine(footer, contentWidth), helpPaletteBottomBorder(contentWidth))
	return strings.Join(lines, "\n")
}

func (m Model) helpPaletteFooter(width int) string {
	help := newHelpModel()
	help.SetWidth(width)
	return help.ShortHelpView([]key.Binding{
		multiBinding("navigate",
			ref(tuiconfig.BlockHelp, tuiconfig.ActionUp),
			ref(tuiconfig.BlockHelp, tuiconfig.ActionDown),
		),
		tuiconfig.Binding(tuiconfig.BlockHelp, tuiconfig.ActionSubmit, "run"),
		multiBinding("close",
			ref(tuiconfig.BlockHelp, tuiconfig.ActionCancel),
			ref(tuiconfig.BlockGlobal, tuiconfig.ActionHelp),
		),
	})
}

func (m Model) helpPaletteActionLines(actions []helpPaletteAction, height, width int) []string {
	if len(actions) == 0 {
		lines := []string{helpPaletteLine(styles.PanelMuted.Italic(true).Render("No matching actions"), width)}
		for len(lines) < height {
			lines = append(lines, helpPaletteLine("", width))
		}
		return lines
	}
	start := min(m.helpPalette.scroll, len(actions)-1)
	end := min(start+height, len(actions))
	lines := make([]string, 0, height)
	previousGroup := ""
	if start > 0 {
		previousGroup = actions[start-1].group
	}
	for index := start; index < end; index++ {
		item := actions[index]
		group := ""
		if index == start || item.group != previousGroup {
			group = item.group
		}
		previousGroup = item.group
		lines = append(lines, helpPaletteLine(renderHelpPaletteAction(item, group, index == m.helpPalette.cursor, width), width))
	}
	for len(lines) < height {
		lines = append(lines, helpPaletteLine("", width))
	}
	return lines
}

func renderHelpPaletteAction(item helpPaletteAction, group string, selected bool, width int) string {
	groupWidth := min(22, max(width/4, 14))
	keyWidth := min(18, max(width/5, 10))
	textWidth := max(width-groupWidth-keyWidth-3, 1)
	groupText := viewutil.PadRight(viewutil.TruncatePlain(group, groupWidth), groupWidth)
	keys := viewutil.TruncatePlain(strings.Join(item.keys, "/"), keyWidth)
	keys = strings.Repeat(" ", max(keyWidth-lipgloss.Width(keys), 0)) + keys
	description := item.title
	if !item.enabled {
		description += " — " + item.reason
	}
	description = viewutil.PadRight(viewutil.TruncatePlain(description, textWidth), textWidth)

	left := groupText + "  " + description + " "
	if selected {
		selectionStyle := styles.TreeItemSelectionStyle()
		keyStyle := styles.FilterText.Background(selectionStyle.GetBackground())
		line := selectionStyle.Render(left) + keyStyle.Render(keys)
		return line + selectionStyle.Render(strings.Repeat(" ", max(width-lipgloss.Width(line), 0)))
	}
	if !item.enabled {
		return helpPaletteDisabledStyle.Render(left) + styles.FilterText.Render(keys)
	}
	return helpPaletteGroupStyle.Render(groupText) +
		styles.PanelText.Render(left[len(groupText):]) +
		styles.FilterText.Render(keys)
}

func helpPaletteTopBorder(width int) string {
	frameInner := viewutil.PopupInnerWidth(width)
	hint := tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionHelp)
	if len([]rune(hint)) > 1 {
		hint += " "
	}
	title, titleWidth := styles.PanelHeaderTab(hint, "Actions", true, true, max(frameInner-1, 0))
	fill := max(frameInner-titleWidth-1, 0)
	return helpPaletteBorderStyle.Render("╭─") + title + helpPaletteBorderStyle.Render(strings.Repeat("─", fill)+"╮")
}

func helpPaletteSeparator(width int) string {
	return helpPaletteBorderStyle.Render("├" + strings.Repeat("─", viewutil.PopupInnerWidth(width)) + "┤")
}

func helpPaletteBottomBorder(width int) string {
	return helpPaletteBorderStyle.Render("╰" + strings.Repeat("─", viewutil.PopupInnerWidth(width)) + "╯")
}

func helpPaletteLine(content string, width int) string {
	return helpPaletteBorderStyle.Render("│") + viewutil.PopupContentLine(content, width) + helpPaletteBorderStyle.Render("│")
}
