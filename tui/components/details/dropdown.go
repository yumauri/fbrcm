package details

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/styles"
)

type dropdownRow struct {
	Key   string
	Label string
	Input bool
	Color string
}

func (m Model) dropdownRows() []dropdownRow {
	switch m.activeField {
	case fieldGroup:
		if m.data == nil {
			return nil
		}
		out := make([]dropdownRow, 0, len(m.data.Groups)+1)
		root := dropdownRow{}
		hasRoot := false
		for _, group := range m.data.Groups {
			if core.NormalizeRemoteConfigGroupKey(group.Key) == core.NormalizeRemoteConfigGroupKey(m.groupKey) {
				continue
			}
			if core.NormalizeRemoteConfigGroupKey(group.Key) == "" {
				root = dropdownRow{Key: group.Key, Label: group.Label}
				hasRoot = true
				continue
			}
			out = append(out, dropdownRow{Key: group.Key, Label: group.Label})
		}
		out = append(out, dropdownRow{Input: true, Label: m.groupInput.Placeholder})
		if hasRoot {
			out = append(out, root)
		}
		return out
	case fieldType:
		out := make([]dropdownRow, 0, len(typeOptions)-1)
		for _, option := range typeOptions {
			if option == m.typeValue {
				continue
			}
			out = append(out, dropdownRow{Key: option, Label: option})
		}
		return out
	case fieldConditionColor:
		out := make([]dropdownRow, 0, len(core.ConditionDisplayColors)+1)
		out = append(out, dropdownRow{Label: "No color"})
		for _, color := range core.ConditionDisplayColors {
			out = append(out, dropdownRow{Key: color, Label: conditionColorValue(color), Color: color})
		}
		return out
	default:
		return nil
	}
}

func (m Model) fieldValueLine(field fieldID) int {
	if m.conditionData != nil {
		width := max(m.width-5, 1)
		line := 1 + len(wrappedLines(rcdisplay.FormatProject(m.conditionData.Project.Name, m.conditionData.Project.ProjectID), width)) + 1
		for _, candidate := range []fieldID{fieldConditionPriority, fieldName, fieldConditionColor} {
			if field == candidate {
				return line + 1
			}
			line += 3
		}
		return 0
	}
	if m.data == nil {
		return 0
	}
	width := max(m.width-5, 1)
	line := 0
	line += 1 + len(wrappedLines(rcdisplay.FormatProject(m.data.Project.Name, m.data.Project.ProjectID), width)) + 1
	if field == fieldGroup {
		return line + 1
	}
	line += 3
	if field == fieldName {
		return line + 1
	}
	line += 3
	if field == fieldType {
		return line + 1
	}
	line += 3
	if field == fieldDescription {
		return line + 1
	}
	return 0
}

func (m Model) valuesTitleLine() int {
	return m.fieldValueLine(fieldDescription) + m.descriptionVisualHeight() + 1
}

func (m Model) valueConditionLine(index int) int {
	if m.data == nil {
		return 0
	}
	width := max(m.width-5, 1)
	line := m.valuesTitleLine() + 1
	for i, value := range m.data.Parameter.Values {
		if i == index {
			return line
		}
		line += m.valueVisualHeight(value, width)
	}
	return line
}

// valueEndLine returns last rendered line for condition label plus value body.
func (m Model) valueEndLine(index int) int {
	if m.data == nil || index < 0 || index >= len(m.data.Parameter.Values) {
		return 0
	}
	width := max(m.width-5, 1)
	start := m.valueConditionLine(index)
	valueLines := m.renderValueLines(m.data.Parameter.Values[index], max(width-4, 1))
	return start + len(valueLines)
}

// valueVisualHeight returns condition label, rendered value, and trailing spacer height.
func (m Model) valueVisualHeight(value core.ParametersValue, width int) int {
	return 1 + len(m.renderValueLines(value, max(width-4, 1))) + 1
}

func (m Model) conditionUsageParameterLine(index int) int {
	if m.conditionData == nil {
		return 0
	}
	width := max(m.width-5, 1)
	condition := m.conditionData.Condition
	line := 1 + len(wrappedLines(rcdisplay.FormatProject(m.conditionData.Project.Name, m.conditionData.Project.ProjectID), width)) + 1
	line += 3 * 3 // priority, name, color
	line += 1 + len(strings.Split(m.conditionExpression, "\n")) + 1
	if condition.Description != "" {
		line += 1 + len(wrappedLines(condition.Description, width)) + 1
	}
	line += 2 // Used by heading and spacer.
	for usageIndex, usage := range condition.Usages {
		if usageIndex == index {
			return line
		}
		line += m.conditionUsageVisualHeight(usage, width)
	}
	return line
}

func (m Model) conditionUsageVisualHeight(usage core.ConditionUsage, width int) int {
	return 1 + len(m.renderConditionUsageValueLines(usage, width)) + 1
}

func (m Model) dropdownCurrentLabel() string {
	switch m.activeField {
	case fieldGroup:
		return m.groupLabel
	case fieldType:
		return m.typeValue
	case fieldConditionColor:
		return conditionColorValue(m.conditionColor)
	default:
		return ""
	}
}

func (m Model) dropdownCurrentStyle() lipgloss.Style {
	switch m.activeField {
	case fieldGroup:
		return groupValueStyle
	case fieldConditionColor:
		return m.conditionStyle(m.conditionColor)
	default:
		return styles.PanelText
	}
}

func (m *Model) openDropdown() {
	rows := m.dropdownRows()
	if len(rows) == 0 {
		return
	}
	m.dropdownOpen = true
	m.dropdownIndex = 0
	if m.activeField == fieldConditionColor {
		for index, row := range rows {
			if row.Key == m.conditionColor {
				m.dropdownIndex = index
				break
			}
		}
	}
	if rows[m.dropdownIndex].Input {
		_ = m.groupInput.Focus()
	} else {
		m.groupInput.Blur()
	}
}

func (m *Model) closeDropdown() {
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.groupInput = newGroupInput()
}

func (m *Model) moveDropdown(delta int) {
	rows := m.dropdownRows()
	if len(rows) == 0 {
		return
	}
	m.dropdownIndex = (m.dropdownIndex + delta + len(rows)) % len(rows)
	if rows[m.dropdownIndex].Input {
		_ = m.groupInput.Focus()
	} else {
		m.groupInput.Blur()
	}
}

func (m *Model) commitDropdown() {
	rows := m.dropdownRows()
	if len(rows) == 0 || m.dropdownIndex < 0 || m.dropdownIndex >= len(rows) {
		return
	}
	row := rows[m.dropdownIndex]
	if row.Input {
		value := strings.TrimSpace(m.groupInput.Value())
		if value == "" {
			return
		}
		m.groupKey = value
		m.groupLabel = value
	} else {
		switch m.activeField {
		case fieldGroup:
			m.groupKey = row.Key
			m.groupLabel = row.Label
		case fieldType:
			m.typeValue = row.Key
		case fieldConditionColor:
			m.conditionColor = row.Key
		}
	}
	m.closeDropdown()
}

func (m Model) dropdownInputSelected() bool {
	rows := m.dropdownRows()
	return m.dropdownIndex >= 0 && m.dropdownIndex < len(rows) && rows[m.dropdownIndex].Input
}

var dropdownBorderStyle = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)

func dropdownOptionStyle(selected bool) lipgloss.Style {
	if !selected {
		return styles.PanelText
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true)
	}
	return styles.PanelText.Bold(true).Foreground(styles.PaletteGold)
}
