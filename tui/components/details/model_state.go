package details

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) SetData(data *messages.ParameterViewData) Model {
	m.data = cloneViewData(data)
	m.groupData = nil
	m.conditionData = nil
	m.activeField = fieldNone
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.nameInput = newTextInput()
	m.descInput = newDescriptionInput()
	m.groupInput = newGroupInput()
	m.priorityInput = newTextInput()
	m.conditionColor = ""
	m.conditionExpression = ""
	m.groupKey = ""
	m.groupLabel = rootgroup.Label
	m.typeValue = "STRING"
	m.selectedValue = -1
	m.selectedUsage = -1
	m.selectedAddValue = false
	m.valuesInvalid = false
	if m.data != nil {
		m.originalParam = cloneParameterEntry(m.data.Parameter)
		m.selectedValue = m.data.SelectedValueIdx
		m.nameInput.SetValue(m.data.Parameter.Key)
		m.descInput.SetValue(m.data.Parameter.Description)
		m.groupKey = m.data.GroupKey
		m.groupLabel = m.data.GroupLabel
		for _, group := range m.data.Groups {
			if group.Key == m.data.GroupKey {
				m.groupLabel = group.Label
				break
			}
		}
		currentType := strings.ToUpper(parameterType(m.data.Parameter))
		for _, option := range typeOptions {
			if option == currentType {
				m.typeValue = option
				break
			}
		}
	}
	if m.data == nil {
		m.originalParam = core.ParametersEntry{}
	}
	m.originalCondition = core.ConditionEntry{}
	m.refreshViewport()
	if m.data != nil && m.selectedValue < 0 && m.activeField == fieldNone {
		m.viewport.GotoTop()
	}
	return m
}

func (m Model) SetConditionData(data *messages.ConditionViewData) Model {
	m.data = nil
	m.groupData = nil
	m.conditionData = cloneConditionViewData(data)
	m.activeField = fieldNone
	m.dropdownOpen = false
	m.selectedValue = -1
	m.selectedUsage = -1
	m.selectedAddValue = false
	m.valuesInvalid = false
	m.originalParam = core.ParametersEntry{}
	m.originalCondition = core.ConditionEntry{}
	m.nameInput = newTextInput()
	m.priorityInput = newTextInput()
	m.groupInput = newGroupInput()
	m.descInput = newDescriptionInput()
	m.conditionColor = ""
	m.conditionExpression = ""
	if m.conditionData != nil {
		m.originalCondition = cloneConditionEntry(m.conditionData.Condition)
		m.nameInput.SetValue(m.conditionData.Condition.Name)
		m.priorityInput.SetValue(strconv.Itoa(m.conditionData.Condition.Priority))
		m.conditionColor = m.conditionData.Condition.TagColor
		m.conditionExpression = m.conditionData.Condition.Expression
	}
	m.refreshViewport()
	m.viewport.GotoTop()
	return m
}

func (m Model) SetGroupData(data *messages.GroupViewData) Model {
	m.data = nil
	m.conditionData = nil
	m.groupData = cloneGroupViewData(data)
	m.activeField = fieldNone
	m.dropdownOpen = false
	m.selectedValue = -1
	m.selectedUsage = -1
	m.selectedAddValue = false
	m.valuesInvalid = false
	m.originalParam = core.ParametersEntry{}
	m.originalCondition = core.ConditionEntry{}
	m.nameInput = newTextInput()
	m.descInput = newDescriptionInput()
	m.groupInput = newGroupInput()
	m.priorityInput = newTextInput()
	if m.groupData != nil {
		m.nameInput.SetValue(m.groupData.Group.Key)
		m.descInput.SetValue(m.groupData.Group.Description)
	}
	m.refreshViewport()
	m.viewport.GotoTop()
	return m
}

func (m Model) SetValuesInvalid(invalid bool) Model {
	if m.valuesInvalid == invalid {
		return m
	}
	m.valuesInvalid = invalid
	m.refreshViewport()
	return m
}

func (m Model) SetSelectedValue(nextRaw string) Model {
	if m.UsageSelected() {
		usage := &m.conditionData.Condition.Usages[m.selectedUsage]
		usage.RawValue = nextRaw
		usage.Value = rcdisplay.FormatRawValue(nextRaw, usage.ValueType)
		m.refreshViewport()
		return m
	}
	if !m.ValueSelected() {
		return m
	}
	value := &m.data.Parameter.Values[m.selectedValue]
	value.RawValue = nextRaw
	value.Value = rcdisplay.FormatRawValue(nextRaw, m.selectedType())
	value.ValueType = m.selectedType()
	value.Empty = nextRaw == ""
	m.refreshViewport()
	return m
}

func (m Model) Dirty() bool {
	return (m.data != nil && m.hasChanges()) || (m.groupData != nil && m.groupHasChanges()) || (m.conditionData != nil && m.conditionHasChanges())
}

func (m Model) Invalid() bool {
	if m.conditionData != nil {
		return m.invalidConditionName() || m.invalidConditionPriority() || m.invalidConditionValues()
	}
	if m.groupData != nil {
		return m.invalidGroupName()
	}
	return m.invalidName() || m.invalidValues()
}

func (m Model) GroupEdit() (core.GroupDetailsEdit, bool) {
	if m.groupData == nil || !m.groupHasChanges() {
		return core.GroupDetailsEdit{}, false
	}
	return core.GroupDetailsEdit{
		Name: m.groupData.Group.Key, NextName: strings.TrimSpace(m.nameInput.Value()), NextDescription: m.descInput.Value(),
	}, true
}

func (m Model) InvalidReasons() []string {
	if m.conditionData != nil {
		return m.conditionInvalidReasons()
	}
	if m.groupData != nil {
		if !m.invalidGroupName() {
			return nil
		}
		if strings.TrimSpace(m.nameInput.Value()) == "" {
			return []string{"Group name is empty."}
		}
		return []string{"Group name already exists in this project."}
	}
	if m.data == nil {
		return nil
	}
	reasons := make([]string, 0, 2)
	if m.invalidName() {
		nextKey := strings.TrimSpace(m.nameInput.Value())
		if nextKey == "" {
			reasons = append(reasons, "Parameter name is empty.")
		} else {
			reasons = append(reasons, "Parameter name already exists in this project.")
		}
	}
	if m.invalidValues() {
		reasons = append(reasons, "One or more values are invalid for selected type "+m.selectedType()+".")
	}
	return reasons
}

// ConditionEdit returns the pending condition Details form mutation.
func (m Model) ConditionEdit() (core.ConditionDetailsEdit, bool) {
	if m.conditionData == nil || !m.conditionHasChanges() {
		return core.ConditionDetailsEdit{}, false
	}
	priority, _ := strconv.Atoi(strings.TrimSpace(m.priorityInput.Value()))
	return core.ConditionDetailsEdit{
		Name:           m.conditionData.Condition.Name,
		NextName:       strings.TrimSpace(m.nameInput.Value()),
		NextExpression: m.conditionExpression,
		NextTagColor:   m.conditionColor,
		NextPriority:   priority,
		ValueEdits:     m.conditionValueEdits(),
	}, true
}

func (m Model) Edit() (core.ParameterDetailsEdit, bool) {
	if m.data == nil || !m.hasChanges() {
		return core.ParameterDetailsEdit{}, false
	}
	return core.ParameterDetailsEdit{
		Create:          m.data.Parameter.Key == "",
		GroupKey:        m.data.GroupKey,
		ParamKey:        m.data.Parameter.Key,
		NextGroupKey:    m.selectedGroupKey(),
		NextParamKey:    strings.TrimSpace(m.nameInput.Value()),
		NextValueType:   m.selectedType(),
		NextDescription: m.descInput.Value(),
		ValueEdits:      m.valueEdits(),
	}, true
}

func (m Model) MarkSaved() Model {
	if m.groupData != nil {
		edit, ok := m.GroupEdit()
		if !ok {
			return m
		}
		m.groupData.Group.Key = edit.NextName
		m.groupData.Group.Label = edit.NextName
		m.groupData.Group.Description = edit.NextDescription
		m.activeField = fieldNone
		m.refreshViewport()
		return m
	}
	if m.conditionData != nil {
		edit, ok := m.ConditionEdit()
		if !ok {
			return m
		}
		m.conditionData.Condition.Name = edit.NextName
		m.conditionData.Condition.Expression = edit.NextExpression
		m.conditionData.Condition.TagColor = edit.NextTagColor
		m.conditionData.Condition.Priority = edit.NextPriority
		m.originalCondition = cloneConditionEntry(m.conditionData.Condition)
		m.activeField = fieldNone
		m.refreshViewport()
		return m
	}
	if m.data == nil {
		return m
	}
	edit, ok := m.Edit()
	if !ok {
		return m
	}
	m.data.GroupKey = edit.NextGroupKey
	m.data.GroupLabel = m.groupLabel
	m.data.Parameter.Key = edit.NextParamKey
	m.data.Parameter.Description = edit.NextDescription
	for i := range m.data.Parameter.Values {
		m.data.Parameter.Values[i].ValueType = edit.NextValueType
	}
	m.originalParam = cloneParameterEntry(m.data.Parameter)
	m.originalCondition = core.ConditionEntry{}
	m.activeField = fieldNone
	m.refreshViewport()
	return m
}

func (m Model) DeactivateField() Model {
	m.activeField = fieldNone
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.nameInput.Blur()
	m.descInput.Blur()
	m.groupInput.Blur()
	m.priorityInput.Blur()
	m.refreshViewport()
	return m
}

func (m Model) ActivateName() (Model, tea.Cmd) {
	m.activateField(fieldName)
	m.refreshViewport()
	return m, m.nameInput.Focus()
}

// ActivateConditionPriority activates the inline numeric priority field.
func (m Model) ActivateConditionPriority() (Model, tea.Cmd) {
	if m.conditionData == nil {
		return m, nil
	}
	m.activateField(fieldConditionPriority)
	m.refreshViewport()
	return m, m.priorityInput.Focus()
}

// ActivateConditionColor activates the inline condition color picker.
func (m Model) ActivateConditionColor() Model {
	if m.conditionData == nil {
		return m
	}
	m.activateField(fieldConditionColor)
	m.openDropdown()
	m.refreshViewport()
	return m
}

// SetConditionExpression stages a raw expression edit in the Details form.
func (m Model) SetConditionExpression(expression string) Model {
	if m.conditionData == nil {
		return m
	}
	m.conditionExpression = expression
	m.refreshViewport()
	return m
}

// ActivateGroup activates group editor.
func (m Model) ActivateGroup() (Model, tea.Cmd) {
	m.activateField(fieldGroup)
	m.openDropdown()
	m.refreshViewport()
	return m, nil
}
