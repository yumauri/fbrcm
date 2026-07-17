package details

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m *Model) refreshViewport() {
	width := max(m.width-5, 1)
	m.nameInput.SetWidth(max(width-2, 1))
	m.priorityInput.SetWidth(max(width-2, 1))
	m.resizeDescriptionInput()
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(max(m.height-2, 1))
	m.viewport.SetContentLines(m.renderContentLines())
	m.ensureSelectedBlockVisible()
}

func (m Model) renderContentLines() []string {
	width := max(m.width-5, 1)
	if m.conditionData != nil {
		return m.renderConditionContentLines(width)
	}
	if m.groupData != nil {
		return m.renderGroupContentLines(width)
	}
	if m.data == nil {
		return padLines([]string{
			"Press Enter on parameter or condition",
			"to open details panel.",
		}, width)
	}

	lines := make([]string, 0, 32)
	lines = appendStyledField(lines, width, "Project", rcdisplay.FormatProject(m.data.Project.Name, m.data.Project.ProjectID), projectValueStyle)
	lines = appendEditableField(lines, width, "Group", m.renderGroupField(), m.fieldChanged(fieldGroup), false)
	lines = appendEditableField(lines, width, "Name", m.renderNameField(), m.fieldChanged(fieldName), m.invalidName())
	lines = appendEditableField(lines, width, "Type", m.renderTypeField(), m.fieldChanged(fieldType), false)
	lines = appendEditableField(lines, width, "Description", m.renderDescriptionField(), m.fieldChanged(fieldDescription), false)

	valuesTitle := fieldTitle("Values", m.valueChanged(), m.invalidValues())
	lines = append(lines, valuesTitle)
	for i, value := range m.data.Parameter.Values {
		prefix := "  "
		if m.activeField == fieldNone && i == m.selectedValue {
			prefix = "▸ "
		}

		conditionStyle := m.conditionStyle(value.Color)
		if value.Label == "default" {
			conditionStyle = conditionDefaultStyle
		}

		label := prefix + rcdisplay.FormatConditionLabel(value.Label)
		if m.activeField == fieldNone && i == m.selectedValue {
			label = selectedValueStyle.Render(ansi.Truncate(label, width, ""))
		} else {
			label = conditionStyle.Render(ansi.Truncate(label, width, ""))
		}
		lines = append(lines, label)

		valueLines := m.renderValueLines(value, max(width-4, 1))
		for _, line := range valueLines {
			lines = append(lines, ansi.Truncate("    "+line, width, ""))
		}
		lines = append(lines, "")
	}

	if len(m.data.Parameter.Values) == 0 {
		lines = append(lines, "No values.")
	}
	if len(m.AvailableConditions()) > 0 {
		label := "  " + addConditionalValueLabel
		if m.AddConditionalValueSelected() {
			label = selectedValueStyle.Render("▸ " + addConditionalValueLabel)
		} else {
			label = parameterKeyStyle.Render(label)
		}
		lines = append(lines, label)
	}

	return padLines(lines, width)
}

func (m Model) renderGroupContentLines(width int) []string {
	lines := make([]string, 0, 12)
	lines = appendStyledField(lines, width, "Project", rcdisplay.FormatProject(m.groupData.Project.Name, m.groupData.Project.ProjectID), projectValueStyle)
	lines = appendEditableField(lines, width, "Name", m.renderNameField(), m.groupFieldChanged(fieldName), m.invalidGroupName())
	lines = appendEditableField(lines, width, "Description", m.renderDescriptionField(), m.groupFieldChanged(fieldDescription), false)
	lines = appendStyledField(lines, width, "Parameters", fmt.Sprintf("%d", len(m.groupData.Group.Parameters)), styles.PanelText)
	return padLines(lines, width)
}

func (m Model) renderConditionContentLines(width int) []string {
	data := m.conditionData
	condition := data.Condition
	lines := make([]string, 0, 24+len(condition.Usages)*4)
	lines = appendStyledField(lines, width, "Project", rcdisplay.FormatProject(data.Project.Name, data.Project.ProjectID), projectValueStyle)
	priority := m.renderConditionPriorityField()
	lines = appendEditableField(lines, width, "Priority", priority, m.conditionFieldChanged(fieldConditionPriority), m.invalidConditionPriority())
	lines = appendEditableField(lines, width, "Name", m.renderConditionNameField(), m.conditionFieldChanged(fieldName), m.invalidConditionName())
	lines = appendEditableField(lines, width, "Color", m.renderConditionColorField(), m.conditionFieldChanged(fieldConditionColor), false)
	lines = appendEditableField(lines, width, "Expression", styles.PanelText.Render(m.conditionExpression), m.conditionExpression != condition.Expression, false)
	lines = appendEditableField(lines, width, "Description", m.renderDescriptionField(), m.conditionFieldChanged(fieldDescription), false)
	usedBy := "Used by " + rcdisplay.FormatCount(len(condition.Usages), "parameter", "parameters")
	lines = append(lines, fieldTitle(usedBy, len(m.conditionValueEdits()) > 0, m.invalidConditionValues()), "")
	if len(condition.Usages) == 0 {
		lines = append(lines, styles.PanelMuted.Italic(true).Render("No parameters use this condition."))
		return padLines(lines, width)
	}
	for index, usage := range condition.Usages {
		parameter := parameterKeyStyle.Render(usage.ParameterKey)
		if m.UsageSelected() && index == m.selectedUsage {
			parameter = selectedValueStyle.Render(usage.ParameterKey)
		}
		path := groupValueStyle.Render(usage.GroupLabel) + labelStyle.Render(" / ") + parameter
		lines = append(lines, ansi.Truncate(path, width, ""))
		lines = append(lines, m.renderConditionUsageValueLines(usage, width)...)
		lines = append(lines, "")
	}
	return padLines(lines, width)
}

func (m Model) renderConditionUsageValueLines(usage core.ConditionUsage, width int) []string {
	value := core.ParametersValue{
		Value:     usage.Value,
		RawValue:  usage.RawValue,
		ValueType: usage.ValueType,
		Empty:     usage.Plain && usage.RawValue == "",
		Plain:     usage.Plain,
	}
	const indent = "  "
	valueLines := m.renderValueLines(value, max(width-lipgloss.Width(indent), 1))
	lines := make([]string, 0, len(valueLines))
	for _, line := range valueLines {
		lines = append(lines, ansi.Truncate(indent+line, width, ""))
	}
	return lines
}

func appendStyledField(lines []string, width int, label, value string, style lipgloss.Style) []string {
	lines = append(lines, labelStyle.Render(label))
	for _, line := range wrappedLines(value, width) {
		lines = append(lines, style.Render(ansi.Truncate(line, width, "")))
	}
	lines = append(lines, "")
	return lines
}

func appendEditableField(lines []string, width int, label, value string, dirty, invalid bool) []string {
	labelText := fieldTitle(label, dirty, invalid)
	lines = append(lines, labelText)
	for line := range strings.SplitSeq(value, "\n") {
		lines = append(lines, ansi.Truncate(line, width, ""))
	}
	lines = append(lines, "")
	return lines
}

func fieldTitle(label string, dirty, invalid bool) string {
	switch {
	case invalid && dirty:
		return fieldInvalidDirtyStyle.Render(label)
	case invalid:
		return fieldInvalidStyle.Render(label)
	case dirty:
		return fieldDirtyStyle.Render(label)
	default:
		return labelStyle.Render(label)
	}
}

func wrappedLines(value string, width int) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{"-"}
	}
	rendered := lipgloss.NewStyle().Width(width).Render(value)
	return strings.Split(rendered, "\n")
}

func wrapLine(value string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	if value == "" {
		return []string{""}
	}
	wrapped := ansi.Hardwrap(value, width, true)
	parts := strings.Split(wrapped, "\n")
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}

func padLines(lines []string, width int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, line+strings.Repeat(" ", max(width-lipgloss.Width(line), 0)))
	}
	return out
}

func cloneParameterEntry(param core.ParametersEntry) core.ParametersEntry {
	param.Values = append([]core.ParametersValue(nil), param.Values...)
	return param
}

func cloneViewData(data *messages.ParameterViewData) *messages.ParameterViewData {
	if data == nil {
		return nil
	}
	next := *data
	next.Groups = append([]messages.ParameterGroupOption(nil), data.Groups...)
	next.ParameterKeys = append([]string(nil), data.ParameterKeys...)
	next.Conditions = append([]core.ParametersCondition(nil), data.Conditions...)
	next.Parameter = cloneParameterEntry(data.Parameter)
	return &next
}

func cloneConditionViewData(data *messages.ConditionViewData) *messages.ConditionViewData {
	if data == nil {
		return nil
	}
	next := *data
	next.Condition = cloneConditionEntry(data.Condition)
	next.ConditionNames = append([]string(nil), data.ConditionNames...)
	return &next
}

func cloneGroupViewData(data *messages.GroupViewData) *messages.GroupViewData {
	if data == nil {
		return nil
	}
	next := *data
	next.GroupNames = append([]string(nil), data.GroupNames...)
	next.Group.Parameters = append([]core.ParametersEntry(nil), data.Group.Parameters...)
	return &next
}

func cloneConditionEntry(condition core.ConditionEntry) core.ConditionEntry {
	condition.Usages = append([]core.ConditionUsage(nil), condition.Usages...)
	return condition
}

func parameterType(param core.ParametersEntry) string {
	for _, value := range param.Values {
		if strings.TrimSpace(value.ValueType) != "" {
			return value.ValueType
		}
	}
	return "unspecified"
}
