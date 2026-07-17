package details

import (
	"strings"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

const addConditionalValueLabel = "+ Add conditional value"

// AvailableConditions returns project conditions that do not yet have a value
// on the parameter, preserving Remote Config evaluation order.
func (m Model) AvailableConditions() []core.ParametersCondition {
	if m.data == nil {
		return nil
	}
	used := make(map[string]struct{}, len(m.data.Parameter.Values))
	for _, value := range m.data.Parameter.Values {
		if value.Label != "default" {
			used[value.Label] = struct{}{}
		}
	}
	available := make([]core.ParametersCondition, 0, len(m.data.Conditions))
	for _, condition := range m.data.Conditions {
		if _, exists := used[condition.Name]; exists {
			continue
		}
		available = append(available, condition)
	}
	return available
}

// AddConditionalValue inserts a staged plain value for a project condition and
// selects it so the existing typed value editor can be opened immediately.
func (m Model) AddConditionalValue(name string) (Model, bool) {
	if m.data == nil {
		return m, false
	}
	var selected core.ParametersCondition
	found := false
	for _, condition := range m.AvailableConditions() {
		if condition.Name == name {
			selected = condition
			found = true
			break
		}
	}
	if !found {
		return m, false
	}

	raw := initialConditionalValue(m.selectedType())
	value := core.ParametersValue{
		Label:     selected.Name,
		Value:     rcdisplay.FormatRawValue(raw, m.selectedType()),
		RawValue:  raw,
		ValueType: m.selectedType(),
		Color:     selected.Color,
		Empty:     raw == "",
		Plain:     true,
	}
	m.data.Parameter.Values = append(m.data.Parameter.Values, value)
	m.reorderValuesByConditionPriority()
	for index := range m.data.Parameter.Values {
		if m.data.Parameter.Values[index].Label == selected.Name {
			m.selectedValue = index
			break
		}
	}
	m.activeField = fieldNone
	m.selectedAddValue = false
	m.refreshViewport()
	return m, true
}

// RemoveAddedConditionalValue removes a value that was introduced after the
// Details form was opened. Existing values are never removed by this helper.
func (m Model) RemoveAddedConditionalValue(name string) Model {
	if m.data == nil || m.originalHasValue(name) {
		return m
	}
	values := m.data.Parameter.Values
	for index := range values {
		if values[index].Label != name {
			continue
		}
		m.data.Parameter.Values = append(values[:index:index], values[index+1:]...)
		m.selectedValue = -1
		m.selectedAddValue = false
		m.refreshViewport()
		return m
	}
	return m
}

func (m Model) ConditionalValuePickerPosition() (int, int, bool) {
	if m.data == nil || len(m.AvailableConditions()) == 0 {
		return 0, 0, false
	}
	line := m.addConditionalValueLine()
	return m.x + 3, m.y + line - m.viewport.YOffset(), true
}

func (m Model) addConditionalValueAt(_, y int) bool {
	if len(m.AvailableConditions()) == 0 {
		return false
	}
	return y == m.y+1+m.addConditionalValueLine()-m.viewport.YOffset()
}

func (m Model) addConditionalValueLine() int {
	if m.data == nil {
		return 0
	}
	line := m.valuesTitleLine() + 1
	width := max(m.width-5, 1)
	for _, value := range m.data.Parameter.Values {
		line += m.valueVisualHeight(value, width)
	}
	if len(m.data.Parameter.Values) == 0 {
		line++
	}
	return line
}

func (m Model) originalHasValue(name string) bool {
	for _, value := range m.originalParam.Values {
		if value.Label == name {
			return true
		}
	}
	return false
}

func (m *Model) reorderValuesByConditionPriority() {
	if m.data == nil {
		return
	}
	byLabel := make(map[string]core.ParametersValue, len(m.data.Parameter.Values))
	for _, value := range m.data.Parameter.Values {
		byLabel[value.Label] = value
	}
	ordered := make([]core.ParametersValue, 0, len(m.data.Parameter.Values))
	for _, condition := range m.data.Conditions {
		if value, exists := byLabel[condition.Name]; exists {
			ordered = append(ordered, value)
			delete(byLabel, condition.Name)
		}
	}
	for _, value := range m.data.Parameter.Values {
		if value.Label == "default" {
			continue
		}
		if remaining, exists := byLabel[value.Label]; exists {
			ordered = append(ordered, remaining)
			delete(byLabel, value.Label)
		}
	}
	if value, exists := byLabel["default"]; exists {
		ordered = append(ordered, value)
	}
	m.data.Parameter.Values = ordered
}

func initialConditionalValue(valueType string) string {
	switch strings.ToUpper(strings.TrimSpace(valueType)) {
	case "BOOLEAN":
		return "false"
	case "NUMBER":
		return "0"
	case "JSON":
		return "{}"
	default:
		return ""
	}
}
