package details

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m *Model) refreshViewport() {
	width := max(m.width-5, 1)
	m.nameInput.SetWidth(max(width-2, 1))
	m.resizeDescriptionInput()
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(max(m.height-2, 1))
	m.viewport.SetContentLines(m.renderContentLines())
	m.ensureSelectedBlockVisible()
}

func (m Model) renderContentLines() []string {
	width := max(m.width-5, 1)
	if m.data == nil {
		return padLines([]string{
			"Press Enter on parameter",
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

	return padLines(lines, width)
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
	next.Parameter = cloneParameterEntry(data.Parameter)
	return &next
}

func parameterType(param core.ParametersEntry) string {
	for _, value := range param.Values {
		if strings.TrimSpace(value.ValueType) != "" {
			return value.ValueType
		}
	}
	return "unspecified"
}
