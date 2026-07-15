package conditions

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

func renderConditionsTable(entries []core.ConditionEntry) string {
	return renderConditionsTableAtWidth(entries, shared.TerminalWidth())
}

func renderConditionsTableAtWidth(entries []core.ConditionEntry, terminalWidth int) string {
	headers := []string{"Priority", "Name", "Used By", "Expression"}
	rows := make([][]string, 0, len(entries))
	widths := headerWidths(headers)
	for _, entry := range entries {
		widths[0] = max(widths[0], lipgloss.Width(fmt.Sprintf("%d", entry.Priority)))
		widths[1] = max(widths[1], lipgloss.Width(entry.Name))
		widths[2] = max(widths[2], lipgloss.Width(fmt.Sprintf("%d", len(entry.Usages))))
		widths[3] = max(widths[3], lipgloss.Width(entry.Expression))
	}

	if terminalWidth > 0 {
		availableExpressionWidth := terminalWidth - tableWidth(widths[:3], len(headers))
		widths[3] = min(widths[3], max(lipgloss.Width(headers[3]), availableExpressionWidth))
	}
	for _, entry := range entries {
		row := []string{
			fmt.Sprintf("%d", entry.Priority),
			entry.Name,
			fmt.Sprintf("%d", len(entry.Usages)),
			ansi.Truncate(entry.Expression, widths[3], "…"),
		}
		rows = append(rows, row)
	}
	return styledTable(headers, rows, widths, map[int]bool{0: true, 2: true}, func(row, col int, style lipgloss.Style) lipgloss.Style {
		if row >= 0 && row < len(entries) && col == 1 && !clistyles.NoColorEnabled() {
			return style.Foreground(clistyles.ConditionLipglossColor(entries[row].TagColor))
		}
		return style
	})
}

func renderConditionDetails(entry core.ConditionEntry) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Priority: %d\n", entry.Priority)
	fmt.Fprintf(&b, "Name: %s\n", renderConditionDetailValue(entry.Name, entry.TagColor, false))
	fmt.Fprintf(&b, "Color: %s\n", renderConditionDetailValue(emptyDash(entry.TagColor), entry.TagColor, true))
	fmt.Fprintf(&b, "Expression: %s\n", entry.Expression)
	if entry.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", entry.Description)
	}
	fmt.Fprintf(&b, "Used by: %s\n", rcdisplay.FormatCount(len(entry.Usages), "parameter", "parameters"))
	if len(entry.Usages) == 0 {
		b.WriteString("\nNo parameters use this condition.")
		return b.String()
	}
	b.WriteString("\n")
	b.WriteString(renderUsagesTable(entry.Usages))
	return b.String()
}

func renderConditionDetailValue(value, tagColor string, circle bool) string {
	if circle && value != "—" {
		value = "● " + value
	}
	if clistyles.NoColorEnabled() {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ConditionLipglossColor(tagColor)).Render(value)
}

func renderUsagesTable(usages []core.ConditionUsage) string {
	headers := []string{"Group", "Parameter", "Type", "Value"}
	rows := make([][]string, 0, len(usages))
	widths := headerWidths(headers)
	for _, usage := range usages {
		row := []string{usage.GroupLabel, usage.ParameterKey, displayValueType(usage.ValueType), usage.Value}
		updateWidths(widths, row)
		rows = append(rows, row)
	}
	return styledTable(headers, rows, widths, nil, func(row, col int, style lipgloss.Style) lipgloss.Style {
		if clistyles.NoColorEnabled() || row < 0 || row >= len(usages) {
			return style
		}
		switch col {
		case 1:
			return style.Foreground(clistyles.PaletteBlueBright)
		case 3:
			return style.UnsetForeground().Inherit(clistyles.RemoteConfigValueStyle(usages[row].Value, usages[row].ValueType))
		}
		return style
	})
}

func displayValueType(valueType string) string {
	valueType = strings.ToUpper(strings.TrimSpace(valueType))
	if valueType == "" {
		return "STRING"
	}
	return valueType
}

func styledTable(headers []string, rows [][]string, widths []int, rightAligned map[int]bool, customize func(row, col int, style lipgloss.Style) lipgloss.Style) string {
	noColor := clistyles.NoColorEnabled()
	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if rightAligned[col] {
			style = style.AlignHorizontal(lipgloss.Right)
		}
		if noColor {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if row >= 0 && row%2 == 1 {
			style = style.Background(clistyles.ColorRowStripe)
		}
		style = style.Foreground(clistyles.PaletteSlateDim)
		return customize(row, col, style)
	}

	width := tableWidth(widths, len(headers))
	tbl := table.New().Headers(headers...).Rows(rows...).Width(width).Border(lipgloss.NormalBorder()).BorderHeader(true).BorderRow(false).StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func tableWidth(widths []int, columnCount int) int {
	width := 3*columnCount + 1
	for _, cellWidth := range widths {
		width += cellWidth
	}
	return width
}

func headerWidths(headers []string) []int {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = lipgloss.Width(header)
	}
	return widths
}

func updateWidths(widths []int, row []string) {
	for i, cell := range row {
		widths[i] = max(widths[i], lipgloss.Width(cell))
	}
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}
