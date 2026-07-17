package details

import (
	"strings"

	"github.com/yumauri/fbrcm/core"
	rcvalue "github.com/yumauri/fbrcm/core/rc/value"
)

func (m Model) selectedGroupKey() string {
	return m.groupKey
}

func (m Model) groupFieldChanged(field fieldID) bool {
	if m.groupData == nil {
		return false
	}
	switch field {
	case fieldName:
		return strings.TrimSpace(m.nameInput.Value()) != m.groupData.Group.Key
	case fieldDescription:
		return m.descInput.Value() != m.groupData.Group.Description
	default:
		return false
	}
}

func (m Model) groupHasChanges() bool {
	return m.groupFieldChanged(fieldName) || m.groupFieldChanged(fieldDescription)
}

func (m Model) invalidGroupName() bool {
	if m.groupData == nil {
		return false
	}
	next := strings.TrimSpace(m.nameInput.Value())
	if next == "" {
		return true
	}
	for _, name := range m.groupData.GroupNames {
		if name != m.groupData.Group.Key && name == next {
			return true
		}
	}
	return false
}

func (m Model) selectedType() string {
	return m.typeValue
}

func (m Model) fieldChanged(field fieldID) bool {
	if m.data == nil {
		return false
	}
	switch field {
	case fieldGroup:
		return core.NormalizeRemoteConfigGroupKey(m.selectedGroupKey()) != core.NormalizeRemoteConfigGroupKey(m.data.GroupKey)
	case fieldName:
		return strings.TrimSpace(m.nameInput.Value()) != m.data.Parameter.Key
	case fieldType:
		return m.selectedType() != strings.ToUpper(parameterType(m.data.Parameter))
	case fieldDescription:
		return m.descInput.Value() != m.data.Parameter.Description
	default:
		return false
	}
}

func (m Model) valueChanged() bool {
	return len(m.valueEdits()) > 0
}

func (m Model) valueEdits() []core.ParameterValueEdit {
	if m.data == nil {
		return nil
	}
	original := make(map[string]string, len(m.originalParam.Values))
	for _, value := range m.originalParam.Values {
		original[value.Label] = value.RawValue
	}
	edits := make([]core.ParameterValueEdit, 0)
	for _, value := range m.data.Parameter.Values {
		originalValue, exists := original[value.Label]
		if exists && originalValue == value.RawValue {
			continue
		}
		edits = append(edits, core.ParameterValueEdit{
			Label:     value.Label,
			NextValue: value.RawValue,
		})
	}
	return edits
}

func (m Model) invalidName() bool {
	if m.data == nil {
		return false
	}
	nextKey := strings.TrimSpace(m.nameInput.Value())
	if nextKey == "" {
		return true
	}
	for _, key := range m.data.ParameterKeys {
		if key == m.data.Parameter.Key {
			continue
		}
		if key == nextKey {
			return true
		}
	}
	return false
}

func (m Model) invalidValues() bool {
	if m.valuesInvalid {
		return true
	}
	if m.data == nil {
		return false
	}
	valueType := m.selectedType()
	for _, value := range m.data.Parameter.Values {
		if !value.Plain {
			continue
		}
		if !rcvalue.ValidRawValueForType(value.RawValue, valueType) {
			return true
		}
	}
	return false
}

func (m Model) hasChanges() bool {
	return m.fieldChanged(fieldGroup) || m.fieldChanged(fieldName) || m.fieldChanged(fieldType) || m.fieldChanged(fieldDescription) || m.valueChanged()
}
