package details

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/parameters"
)

func (m Model) CurrentBoolValueAnchor() (parameters.BoolValueAnchor, bool) {
	value, ok := m.currentSelectedPlainValue("boolean")
	if !ok {
		return parameters.BoolValueAnchor{}, false
	}
	x, y := m.valueEditorPosition()
	return parameters.BoolValueAnchor{
		Project:      m.data.Project,
		GroupKey:     m.data.GroupKey,
		ParamKey:     m.data.Parameter.Key,
		ValueLabel:   value.Label,
		Value:        strings.EqualFold(strings.TrimSpace(value.RawValue), "true"),
		CurrentValue: value.RawValue,
		X:            x + 2,
		Y:            y,
	}, true
}

func (m Model) CurrentNumberValueAnchor() (parameters.NumberValueAnchor, bool) {
	value, ok := m.currentSelectedPlainValue("number")
	if !ok {
		return parameters.NumberValueAnchor{}, false
	}
	currentValue := strings.TrimSpace(value.RawValue)
	x, y := m.valueEditorPosition()
	return parameters.NumberValueAnchor{
		Project:      m.data.Project,
		GroupKey:     m.data.GroupKey,
		ParamKey:     m.data.Parameter.Key,
		ValueLabel:   value.Label,
		CurrentValue: currentValue,
		X:            x + 2,
		Y:            y - 1,
		Width:        max(lipgloss.Width(currentValue), 3),
		MaxWidth:     max(m.width-5, 3),
	}, true
}

func (m Model) CurrentStringValueAnchor(_ int) (parameters.StringValueAnchor, bool) {
	value, ok := m.currentSelectedPlainValue("string")
	if !ok {
		return parameters.StringValueAnchor{}, false
	}
	currentValue := value.RawValue
	x, y := m.valueEditorPosition()
	editorX := x + 2
	width := max(m.width-(editorX-m.x)-2, 15)
	return parameters.StringValueAnchor{
		Project:      m.data.Project,
		GroupKey:     m.data.GroupKey,
		ParamKey:     m.data.Parameter.Key,
		ValueLabel:   value.Label,
		CurrentValue: currentValue,
		X:            editorX,
		Y:            y - 1,
		Width:        width,
		MaxWidth:     width + 2,
		FullWidth:    false,
		Expanded:     strings.Contains(currentValue, "\n"),
	}, true
}

func (m Model) CurrentJSONValueAnchor() (parameters.JSONValueAnchor, bool) {
	value, ok := m.currentSelectedPlainValue("json")
	if !ok {
		return parameters.JSONValueAnchor{}, false
	}
	return parameters.JSONValueAnchor{
		Project:      m.data.Project,
		GroupKey:     m.data.GroupKey,
		ParamKey:     m.data.Parameter.Key,
		ValueLabel:   value.Label,
		CurrentValue: value.RawValue,
	}, true
}

func (m Model) currentSelectedPlainValue(valueType string) (core.ParametersValue, bool) {
	if !m.ValueSelected() {
		return core.ParametersValue{}, false
	}
	value := m.data.Parameter.Values[m.selectedValue]
	if !value.Plain {
		return core.ParametersValue{}, false
	}
	selectedType := strings.TrimSpace(strings.ToLower(m.selectedType()))
	if selectedType == "" {
		selectedType = "string"
	}
	if selectedType != valueType {
		return core.ParametersValue{}, false
	}
	return value, true
}

func (m Model) valueEditorPosition() (int, int) {
	line := m.valueConditionLine(m.selectedValue) + 1
	return m.x + 3, m.y + 1 + line - m.viewport.YOffset()
}
