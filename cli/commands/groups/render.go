package groups

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
)

func renderGroupsTable(groups []projectGroup, includeProject bool) string {
	return renderGroupsTableAtWidth(groups, includeProject, shared.TerminalWidth())
}

func renderGroupsTableAtWidth(groups []projectGroup, includeProject bool, terminalWidth int) string {
	headers := []string{"Name", "Parameters", "Description"}
	if includeProject {
		headers = append([]string{"Project"}, headers...)
	}
	widths := shared.HeaderWidths(headers)
	for _, item := range groups {
		row := groupTableRow(item, includeProject, item.Group.Description)
		shared.UpdateTableWidths(widths, row)
	}
	if terminalWidth > 0 {
		descriptionIndex := len(widths) - 1
		shrinkTableColumn(widths, descriptionIndex, lipgloss.Width(headers[descriptionIndex]), terminalWidth)
		if includeProject {
			shrinkTableColumn(widths, 0, lipgloss.Width(headers[0]), terminalWidth)
		}
		nameIndex := 0
		if includeProject {
			nameIndex = 1
		}
		shrinkTableColumn(widths, nameIndex, lipgloss.Width(headers[nameIndex]), terminalWidth)
	}
	rows := make([][]string, 0, len(groups))
	for _, item := range groups {
		descriptionIndex := len(widths) - 1
		row := groupTableRow(item, includeProject, ansi.Truncate(item.Group.Description, widths[descriptionIndex], "…"))
		if includeProject {
			row[0] = ansi.Truncate(row[0], widths[0], "…")
			row[1] = ansi.Truncate(row[1], widths[1], "…")
		} else {
			row[0] = ansi.Truncate(row[0], widths[0], "…")
		}
		rows = append(rows, row)
	}
	parameterIndex := 1
	nameIndex := 0
	if includeProject {
		parameterIndex = 2
		nameIndex = 1
	}
	return shared.StyledTable(headers, rows, widths, map[int]bool{parameterIndex: true}, func(row, col int, style lipgloss.Style) lipgloss.Style {
		if row >= 0 && col == nameIndex && !clistyles.NoColorEnabled() {
			return style.Foreground(clistyles.PaletteYellow)
		}
		return style
	})
}

func shrinkTableColumn(widths []int, index, minimum, terminalWidth int) {
	overflow := shared.TableWidth(widths) - terminalWidth
	if overflow <= 0 {
		return
	}
	widths[index] = max(minimum, widths[index]-overflow)
}

func groupTableRow(item projectGroup, includeProject bool, description string) []string {
	row := []string{item.Group.Key, fmt.Sprintf("%d", len(item.Group.Parameters)), description}
	if includeProject {
		row = append([]string{item.Project.ProjectID}, row...)
	}
	return row
}
