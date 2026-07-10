package parameters

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

func (m Model) valueNodeValueX(node visibleNode, param *core.ParametersEntry) int {
	layout := m.parameterRenderLayout()
	label := rcdisplay.FormatConditionLabel(param.Values[node.valueIdx].Label)
	conditionWidth := parameterConditionWidth(param)
	if layout.mode == parameterRenderModeNarrow {
		fillerWidth := max(conditionWidth-lipgloss.Width(label)+1, 1)
		return lipgloss.Width(compactBranchGlyph(layout.paramStart, m.valueConnector(node, param))) + 1 + lipgloss.Width(label) + 1 + fillerWidth + 1
	}
	leafOffset := 1
	if len(param.Values) == 1 {
		leafOffset = 2
	}
	leafOffset++
	leafValueStart := layout.valueStart + leafOffset
	labelStart := max(leafValueStart-conditionWidth-4, layout.paramStart+2)
	fillerWidth := max(leafValueStart-labelStart-lipgloss.Width(label)-3, 1)
	return lipgloss.Width(branchGlyph(layout.paramStart, labelStart, m.valueConnector(node, param))) + 1 + lipgloss.Width(label) + 1 + fillerWidth + 1
}

func (m Model) CurrentConditionalValueAnchor() (ConditionalValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return ConditionalValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return ConditionalValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return ConditionalValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	if value.Label == "" || value.Label == "default" {
		return ConditionalValueAnchor{}, false
	}
	return ConditionalValueAnchor{
		Project:    project.project,
		GroupKey:   node.groupKey,
		ParamKey:   node.paramKey,
		ValueLabel: value.Label,
	}, true
}

func (m Model) CurrentBoolValueAnchor() (BoolValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return BoolValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return BoolValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return BoolValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	if strings.TrimSpace(strings.ToLower(value.ValueType)) != "boolean" {
		return BoolValueAnchor{}, false
	}
	switch strings.TrimSpace(strings.ToLower(value.Value)) {
	case "true", "false":
	default:
		return BoolValueAnchor{}, false
	}
	screenLine := m.screenLineForOffset(m.cursor, m.offset)
	if screenLine < 0 {
		return BoolValueAnchor{}, false
	}
	valueX := m.valueNodeValueX(node, param)
	return BoolValueAnchor{
		Project:      project.project,
		GroupKey:     node.groupKey,
		ParamKey:     node.paramKey,
		ValueLabel:   value.Label,
		Value:        strings.EqualFold(value.Value, "true"),
		CurrentValue: value.RawValue,
		X:            m.x + valueX - 1,
		Y:            m.y + screenLine + 1,
	}, true
}

func (m Model) CurrentNumberValueAnchor() (NumberValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return NumberValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return NumberValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return NumberValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	if strings.TrimSpace(strings.ToLower(value.ValueType)) != "number" {
		return NumberValueAnchor{}, false
	}
	currentValue := strings.TrimSpace(value.Value)
	screenLine := m.screenLineForOffset(m.cursor, m.offset)
	if screenLine < 0 {
		return NumberValueAnchor{}, false
	}
	valueX := m.valueNodeValueX(node, param)
	return NumberValueAnchor{
		Project:      project.project,
		GroupKey:     node.groupKey,
		ParamKey:     node.paramKey,
		ValueLabel:   value.Label,
		CurrentValue: currentValue,
		X:            m.x + valueX - 1,
		Y:            m.y + screenLine,
		Width:        max(lipgloss.Width(currentValue), 3),
		MaxWidth:     max(m.viewportWidth()-valueX-1, 3),
	}, true
}

func (m Model) CurrentStringValueAnchor() (StringValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return StringValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return StringValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return StringValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	valueType := strings.TrimSpace(strings.ToLower(value.ValueType))
	if valueType != "string" && valueType != "" {
		return StringValueAnchor{}, false
	}
	if !value.Plain {
		return StringValueAnchor{}, false
	}
	screenLine := m.screenLineForOffset(m.cursor, m.offset)
	if screenLine < 0 {
		return StringValueAnchor{}, false
	}
	valueX := m.valueNodeValueX(node, param)
	currentValue := value.RawValue
	minWidth := max(lipgloss.Width(currentValue), 15)
	maxWidth := max(m.width-(valueX-1), 1)
	fullWidth := max(maxWidth-4, 1) < minWidth
	return StringValueAnchor{
		Project:      project.project,
		GroupKey:     node.groupKey,
		ParamKey:     node.paramKey,
		ValueLabel:   value.Label,
		CurrentValue: currentValue,
		X:            m.x + valueX - 1,
		Y:            m.y + screenLine,
		Width:        minWidth,
		MaxWidth:     maxWidth,
		FullWidth:    fullWidth,
		Expanded:     strings.Contains(currentValue, "\n"),
	}, true
}

func (m Model) CurrentJSONValueAnchor() (JSONValueAnchor, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return JSONValueAnchor{}, false
	}
	node := m.visible[m.cursor]
	if node.kind != nodeValue || node.transient {
		return JSONValueAnchor{}, false
	}
	project := m.projectByID(node.projectID)
	param := m.parameterByKey(node.projectID, node.groupKey, node.paramKey)
	if project == nil || param == nil || node.valueIdx < 0 || node.valueIdx >= len(param.Values) {
		return JSONValueAnchor{}, false
	}
	value := param.Values[node.valueIdx]
	if strings.TrimSpace(strings.ToLower(value.ValueType)) != "json" {
		return JSONValueAnchor{}, false
	}
	if !value.Plain {
		return JSONValueAnchor{}, false
	}
	return JSONValueAnchor{
		Project:      project.project,
		GroupKey:     node.groupKey,
		ParamKey:     node.paramKey,
		ValueLabel:   value.Label,
		CurrentValue: value.RawValue,
	}, true
}
