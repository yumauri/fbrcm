package details

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/tui/messages"
)

func (m Model) SetData(data *messages.ParameterViewData) Model {
	m.data = cloneViewData(data)
	m.activeField = fieldNone
	m.dropdownOpen = false
	m.dropdownIndex = 0
	m.nameInput = newTextInput()
	m.descInput = newDescriptionInput()
	m.groupInput = newGroupInput()
	m.groupKey = ""
	m.groupLabel = rootgroup.Label
	m.typeValue = "STRING"
	m.selectedValue = -1
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
	m.refreshViewport()
	if m.data != nil && m.selectedValue < 0 && m.activeField == fieldNone {
		m.viewport.GotoTop()
	}
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
	return m.data != nil && m.hasChanges()
}

func (m Model) Invalid() bool {
	return m.invalidName() || m.invalidValues()
}

func (m Model) InvalidReasons() []string {
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
	m.refreshViewport()
	return m
}

func (m Model) ActivateName() (Model, tea.Cmd) {
	m.activateField(fieldName)
	m.refreshViewport()
	return m, m.nameInput.Focus()
}

// ActivateGroup activates group editor.
func (m Model) ActivateGroup() (Model, tea.Cmd) {
	m.activateField(fieldGroup)
	m.openDropdown()
	m.refreshViewport()
	return m, nil
}
