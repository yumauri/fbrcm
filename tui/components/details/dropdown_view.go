package details

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) DropdownCurrentView() string {
	if !m.DropdownOpen() {
		return ""
	}
	value := m.dropdownCurrentLabel()
	width := max(lipgloss.Width(value), 1)
	return strings.Join([]string{
		dropdownBorderStyle.Render("╭" + strings.Repeat("─", width+2) + "╮"),
		dropdownBorderStyle.Render("│ ") + m.dropdownCurrentStyle().Render(viewutil.PadRight(value, width)) + dropdownBorderStyle.Render(" │"),
		dropdownBorderStyle.Render("╰" + strings.Repeat("─", width+2) + "╯"),
	}, "\n")
}

func (m Model) DropdownListView() string {
	if !m.DropdownOpen() {
		return ""
	}
	rows := m.dropdownRows()
	if len(rows) == 0 {
		return ""
	}
	width := 1
	for _, row := range rows {
		width = max(width, lipgloss.Width(row.Label))
	}
	if m.activeField == fieldGroup {
		width = max(width, lipgloss.Width(strings.TrimSpace(m.groupInput.Value()))+1)
		width = max(width, lipgloss.Width(m.groupInput.Placeholder))
	}
	lines := make([]string, 0, len(rows)+2)
	topLeft := "╭"
	if m.dropdownIndex == 0 {
		topLeft = "─"
	}
	bottomLeft := "╰"
	if m.dropdownIndex == len(rows)-1 {
		bottomLeft = "─"
	}
	lines = append(lines, dropdownBorderStyle.Render(topLeft+strings.Repeat("─", width+2)+"╮"))
	input := m.groupInput
	for i, row := range rows {
		left := dropdownBorderStyle.Render("│ ")
		switch i {
		case m.dropdownIndex:
			left = dropdownBorderStyle.Render("▸ ")
		case m.dropdownIndex - 1:
			left = dropdownBorderStyle.Render("╯ ")
		case m.dropdownIndex + 1:
			left = dropdownBorderStyle.Render("╮ ")
		}
		content := ""
		if row.Input {
			if i == m.dropdownIndex {
				input.SetWidth(max(width-1, 1))
				content = viewutil.PadRight(input.View(), width)
			} else if value := strings.TrimSpace(m.groupInput.Value()); value != "" {
				content = dropdownOptionStyle(false).Render(viewutil.PadRight(value, width))
			} else {
				content = styles.PanelMuted.Render(viewutil.PadRight(input.Placeholder, width))
			}
		} else if m.activeField == fieldConditionColor {
			style := m.conditionStyle(row.Color)
			if i == m.dropdownIndex {
				style = style.Bold(true)
			}
			content = style.Render(viewutil.PadRight(row.Label, width))
		} else {
			content = dropdownOptionStyle(i == m.dropdownIndex).Render(viewutil.PadRight(row.Label, width))
		}
		lines = append(lines, left+content+dropdownBorderStyle.Render(" │"))
	}
	lines = append(lines, dropdownBorderStyle.Render(bottomLeft+strings.Repeat("─", width+2)+"╯"))
	return strings.Join(lines, "\n")
}
