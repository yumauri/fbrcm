package conditions

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
	"github.com/yumauri/fbrcm/ops/shared"
)

func renderConditionsTableAtWidth(entries []core.ConditionEntry, terminalWidth int) string {
	items := make([]conditionListItem, len(entries))
	for i, entry := range entries {
		items[i] = conditionListItem{ConditionEntry: entry}
	}
	return renderConditionListTable(items, false, terminalWidth)
}

func renderConditionListTable(entries []conditionListItem, showProject bool, terminalWidth int) string {
	headers := []string{"Priority", "Name", "Used By", "Expression"}
	offset := 0
	if showProject {
		headers = append([]string{"Project"}, headers...)
		offset = 1
	}
	rows := make([][]string, 0, len(entries))
	widths := shared.HeaderWidths(headers)
	for _, entry := range entries {
		row := []string{fmt.Sprintf("%d", entry.Priority), entry.Name, fmt.Sprintf("%d", len(entry.Usages)), entry.Expression}
		if showProject {
			row = append([]string{entry.Project}, row...)
		}
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
	}
	flexible := []int{3 + offset, 1 + offset}
	if showProject {
		flexible = append(flexible, 0)
	}
	shared.FitTableColumns(widths, terminalWidth, flexible...)
	shared.TruncateTableHeaders(headers, widths)
	for i := range rows {
		rows[i][3+offset] = clistyles.RenderFirebaseAppPlatforms(rows[i][3+offset], conditionTableCellStyle(i, clistyles.PanelMuted))
	}
	shared.TruncateTableColumns(rows, widths, flexible...)
	return shared.StyledTable(headers, rows, widths, map[int]bool{offset: true, 2 + offset: true}, func(row, col int, style lipgloss.Style) lipgloss.Style {
		if row >= 0 && row < len(entries) && col == 1+offset && !clistyles.NoColorEnabled() {
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
	fmt.Fprintf(&b, "Expression: %s\n", clistyles.RenderFirebaseAppPlatforms(entry.Expression, lipgloss.NewStyle()))
	fmt.Fprintf(&b, "Used by: %s\n", rcdisplay.FormatCount(len(entry.Usages), "parameter", "parameters"))
	b.WriteString("\n")
	b.WriteString(renderUsagesTable(entry.Usages))
	return b.String()
}

func conditionTableCellStyle(row int, style lipgloss.Style) lipgloss.Style {
	if !clistyles.NoColorEnabled() && row%2 == 1 {
		return style.Background(clistyles.ColorRowStripe)
	}
	return style
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
	widths := shared.HeaderWidths(headers)
	for _, usage := range usages {
		row := []string{usage.GroupLabel, usage.ParameterKey, displayValueType(usage.ValueType), usage.Value}
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
	}
	return shared.StyledTable(headers, rows, widths, nil, func(row, col int, style lipgloss.Style) lipgloss.Style {
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

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}
