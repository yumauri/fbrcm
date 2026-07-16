package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestDetailsAddConditionalValuePickerStagesTypedValue(t *testing.T) {
	m := conditionalValueDetailsTestModel()

	m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'a', Text: "a"}))
	if !handled || !m.moveParam.IsOpen() || m.conditionalAdd == nil {
		t.Fatalf("add key = handled:%v picker:%v session:%#v", handled, m.moveParam.IsOpen(), m.conditionalAdd)
	}
	option, ok := m.moveParam.Current()
	if !ok || option.Key != "staff" {
		t.Fatalf("selected option = %+v, ok=%v; want staff", option, ok)
	}

	m, _, handled = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !handled || !m.boolPicker.IsOpen() || m.conditionalAdd == nil || m.conditionalAdd.condition != "staff" {
		t.Fatalf("condition submit = handled:%v bool:%v session:%#v", handled, m.boolPicker.IsOpen(), m.conditionalAdd)
	}

	m, _, handled = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	edit, changed := m.details.Edit()
	if !handled || m.boolPicker.IsOpen() || m.conditionalAdd != nil || !changed || len(edit.ValueEdits) != 1 {
		t.Fatalf("value submit = handled:%v open:%v session:%#v changed:%v edit:%+v", handled, m.boolPicker.IsOpen(), m.conditionalAdd, changed, edit)
	}
	if edit.ValueEdits[0].Label != "staff" || edit.ValueEdits[0].NextValue != "false" {
		t.Fatalf("conditional edit = %+v, want staff=false", edit.ValueEdits[0])
	}
}

func TestDetailsAddConditionalValueCancelRemovesTransientValue(t *testing.T) {
	m := conditionalValueDetailsTestModel()
	m, _, _ = m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'a', Text: "a"}))
	m, _, _ = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m, _, handled := m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))

	if !handled || m.boolPicker.IsOpen() || m.conditionalAdd != nil || m.details.Dirty() {
		t.Fatalf("cancel = handled:%v open:%v session:%#v dirty:%v", handled, m.boolPicker.IsOpen(), m.conditionalAdd, m.details.Dirty())
	}
}

func conditionalValueDetailsTestModel() Model {
	m := viewTestModel(100, 32, panels.Details)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project:    core.Project{Name: "Demo", ProjectID: "demo"},
		GroupLabel: "(root)",
		Groups:     []messages.ParameterGroupOption{{Label: "(root)"}},
		Conditions: []core.ParametersCondition{{Name: "staff", Color: "GREEN"}},
		Parameter: core.ParametersEntry{
			Key: "flag",
			Values: []core.ParametersValue{{
				Label: "default", Value: "false", RawValue: "false", ValueType: "BOOLEAN", Plain: true,
			}},
		},
		SelectedValueIdx: -1,
	})
	m.detailsVisible = true
	m.setActive(panels.Details)
	return m
}
