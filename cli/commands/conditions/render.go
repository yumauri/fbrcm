package conditions

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
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
	widths := shared.HeaderWidths(headers)
	for _, entry := range entries {
		widths[0] = max(widths[0], lipgloss.Width(fmt.Sprintf("%d", entry.Priority)))
		widths[1] = max(widths[1], lipgloss.Width(entry.Name))
		widths[2] = max(widths[2], lipgloss.Width(fmt.Sprintf("%d", len(entry.Usages))))
		widths[3] = max(widths[3], lipgloss.Width(entry.Expression))
	}

	if terminalWidth > 0 {
		availableExpressionWidth := terminalWidth - shared.TableWidth(widths[:3]) - 3
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
	return shared.StyledTable(headers, rows, widths, map[int]bool{0: true, 2: true}, func(row, col int, style lipgloss.Style) lipgloss.Style {
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
