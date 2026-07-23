package details

import (
	"strconv"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rcvalue "github.com/yumauri/fbrcm/core/rc/value"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/styles"
)

func (m Model) renderConditionPriorityField() string {
	value := strings.TrimSpace(m.priorityInput.Value())
	if m.activeField == fieldConditionPriority {
		return styles.FilterText.Render(m.priorityInput.View())
	}
	return parameterKeyStyle.Render(value) + styles.PanelMuted.Render(" (earlier conditions take precedence)")
}

func (m Model) renderConditionNameField() string {
	if m.activeField == fieldName {
		return styles.FilterText.Render(m.nameInput.View())
	}
	return m.conditionStyle(m.conditionColor).Render(strings.TrimSpace(m.nameInput.Value()))
}

func (m Model) renderConditionColorField() string {
	value := conditionColorValue(m.conditionColor)
	if m.activeField == fieldConditionColor {
		return selectedDropdownFieldStyle().Render(value)
	}
	return m.conditionStyle(m.conditionColor).Render(value)
}

func conditionColorValue(color string) string {
	return viewutil.ConditionColorValue(color)
}

func (m Model) conditionFieldChanged(field fieldID) bool {
	if m.conditionData == nil {
		return false
	}
	switch field {
	case fieldConditionPriority:
		priority, err := strconv.Atoi(strings.TrimSpace(m.priorityInput.Value()))
		return err != nil || priority != m.conditionData.Condition.Priority
	case fieldName:
		return strings.TrimSpace(m.nameInput.Value()) != m.conditionData.Condition.Name
	case fieldConditionColor:
		return m.conditionColor != m.conditionData.Condition.TagColor
	default:
		return false
	}
}

func (m Model) conditionHasChanges() bool {
	return m.conditionFieldChanged(fieldConditionPriority) ||
		m.conditionFieldChanged(fieldName) ||
		m.conditionFieldChanged(fieldConditionColor) ||
		(m.conditionData != nil && m.conditionExpression != m.conditionData.Condition.Expression) ||
		len(m.conditionValueEdits()) > 0
}

func (m Model) conditionValueEdits() []core.ConditionUsageValueEdit {
	if m.conditionData == nil {
		return nil
	}
	original := make(map[string]string, len(m.originalCondition.Usages))
	for _, usage := range m.originalCondition.Usages {
		original[usage.GroupKey+"\x00"+usage.ParameterKey] = usage.RawValue
	}
	edits := make([]core.ConditionUsageValueEdit, 0)
	for _, usage := range m.conditionData.Condition.Usages {
		if original[usage.GroupKey+"\x00"+usage.ParameterKey] == usage.RawValue {
			continue
		}
		edits = append(edits, core.ConditionUsageValueEdit{
			GroupKey: usage.GroupKey, ParameterKey: usage.ParameterKey, NextValue: usage.RawValue,
		})
	}
	return edits
}

func (m Model) invalidConditionValues() bool {
	if m.valuesInvalid {
		return true
	}
	if m.conditionData == nil {
		return false
	}
	for _, usage := range m.conditionData.Condition.Usages {
		if usage.Plain && !rcvalue.ValidRawValueForType(usage.RawValue, usage.ValueType) {
			return true
		}
	}
	return false
}

func (m Model) invalidConditionName() bool {
	if m.conditionData == nil {
		return false
	}
	name := strings.TrimSpace(m.nameInput.Value())
	if _, err := core.NormalizeConditionName(name); err != nil {
		return true
	}
	for _, existing := range m.conditionData.ConditionNames {
		if existing != m.conditionData.Condition.Name && existing == name {
			return true
		}
	}
	return false
}

func (m Model) invalidConditionPriority() bool {
	if m.conditionData == nil {
		return false
	}
	priority, err := strconv.Atoi(strings.TrimSpace(m.priorityInput.Value()))
	return err != nil || priority < 1 || (len(m.conditionData.ConditionNames) > 0 && priority > len(m.conditionData.ConditionNames))
}

func (m Model) conditionInvalidReasons() []string {
	reasons := make([]string, 0, 2)
	if m.invalidConditionPriority() {
		maxPriority := len(m.conditionData.ConditionNames)
		if maxPriority > 0 {
			reasons = append(reasons, "Condition priority must be between 1 and "+strconv.Itoa(maxPriority)+".")
		} else {
			reasons = append(reasons, "Condition priority must be a positive number.")
		}
	}
	if m.invalidConditionName() {
		name := strings.TrimSpace(m.nameInput.Value())
		if name == "" {
			reasons = append(reasons, "Condition name is empty.")
		} else {
			reasons = append(reasons, "Condition name is invalid or already exists in this project.")
		}
	}
	if m.invalidConditionValues() {
		reasons = append(reasons, "One or more conditional values are invalid for their parameter type.")
	}
	return reasons
}

func (m Model) updatePriorityInput(msg tea.Msg) (Model, tea.Cmd) {
	next := m.priorityInput
	var cmd tea.Cmd
	next, cmd = next.Update(msg)
	for _, r := range next.Value() {
		if !unicode.IsDigit(r) {
			return m, nil
		}
	}
	m.priorityInput = next
	return m, cmd
}
