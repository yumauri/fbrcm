package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestParameterConditionalValueEnterOpensConditionDetails(t *testing.T) {
	m := detailsCrossNavigationTestModel(t)
	m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	if !handled || !m.details.IsCondition() || m.details.ConditionData().Condition.Name != "staff" {
		t.Fatalf("Enter = handled:%v condition:%v data:%#v", handled, m.details.IsCondition(), m.details.ConditionData())
	}
}

func TestConditionUsageEnterOpensSelectedParameterDetails(t *testing.T) {
	m := openCrossNavigationConditionDetails(t)
	m.details, _ = m.details.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if !m.details.UsageSelected() {
		t.Fatal("condition usage was not selected")
	}
	m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	data := m.details.Data()
	if !handled || data == nil || data.Parameter.Key != "enabled" || !m.details.ValueSelected() {
		t.Fatalf("Enter = handled:%v data:%#v valueSelected:%v", handled, data, m.details.ValueSelected())
	}
	anchor, ok := m.details.CurrentConditionalValueAnchor()
	if !ok || anchor.ValueLabel != "staff" {
		t.Fatalf("selected parameter value = %#v, ok=%v; want staff", anchor, ok)
	}
}

func TestConditionUsageRightAndEditOpenTypedValueEditor(t *testing.T) {
	keys := []tea.Key{{Code: tea.KeyRight}, {Code: 'e', Text: "e"}}
	for _, key := range keys {
		m := openCrossNavigationConditionDetails(t)
		m.details, _ = m.details.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
		m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(key))
		if !handled || !m.boolPicker.IsOpen() || m.stringInput.IsOpen() {
			t.Fatalf("%v = handled:%v bool:%v expression:%v", key.Code, handled, m.boolPicker.IsOpen(), m.stringInput.IsOpen())
		}
	}
}

func TestConditionUsageEditorStagesValueInConditionForm(t *testing.T) {
	m := openCrossNavigationConditionDetails(t)
	m.details, _ = m.details.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	m, _, _ = m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))
	m, _, _ = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	m, _, handled := m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))

	edit, changed := m.details.ConditionEdit()
	if !handled || !changed || len(edit.ValueEdits) != 1 || edit.ValueEdits[0].NextValue != "false" {
		t.Fatalf("submit = handled:%v changed:%v edit:%+v", handled, changed, edit)
	}
}

func openCrossNavigationConditionDetails(t *testing.T) Model {
	t.Helper()
	m := detailsCrossNavigationTestModel(t)
	m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !handled || !m.details.IsCondition() {
		t.Fatal("failed to open condition Details")
	}
	return m
}

func detailsCrossNavigationTestModel(t *testing.T) Model {
	t.Helper()
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	parameterTree := &core.ParametersTree{
		Conditions: []core.ParametersCondition{{Name: "staff", Color: "GREEN"}},
		Groups: []core.ParametersGroup{{
			Key: "__default__", Label: "(root)",
			Parameters: []core.ParametersEntry{{
				Key: "enabled",
				Values: []core.ParametersValue{
					{Label: "staff", Value: "true", RawValue: "true", ValueType: "BOOLEAN", Color: "GREEN", Plain: true},
					{Label: "default", Value: "false", RawValue: "false", ValueType: "BOOLEAN", Plain: true},
				},
			}},
		}},
	}
	conditionTree := &core.ConditionsTree{Conditions: []core.ConditionEntry{{
		Priority: 1, Name: "staff", Expression: "true", TagColor: "GREEN",
		Usages: []core.ConditionUsage{{
			GroupKey: "__default__", GroupLabel: "(root)", ParameterKey: "enabled",
			Value: "true", RawValue: "true", ValueType: "BOOLEAN", Plain: true,
		}},
	}}}

	m := viewTestModel(100, 32, panels.Details)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{Project: project, Tree: parameterTree, Source: "cache"})
	m.conditions, _ = m.conditions.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.conditions, _ = m.conditions.Update(messages.ConditionsLoadedMsg{Project: project, Tree: conditionTree, Source: "cache"})
	data, ok := m.parameters.ParameterViewData(project.ProjectID, "__default__", "enabled", "staff")
	if !ok {
		t.Fatal("parameter Details data missing")
	}
	m.details = m.details.SetData(data)
	m.detailsVisible = true
	m.setActive(panels.Details)
	return m
}
