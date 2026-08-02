package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
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

func TestManagedConditionalValueEnterOpensConditionDetailsWithoutEnablingMutation(t *testing.T) {
	m := detailsCrossNavigationTestModel(t)
	data := m.details.Data()
	data.Parameter.Values[0].Plain = false
	data.Parameter.Values[0].Display = rcdisplay.ValueSummary{
		Kind: rcdisplay.ValueSummaryExperiment,
		Text: "⚗ (a/b test)",
	}
	m.details = m.details.SetData(data)

	m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !handled || !m.details.IsCondition() || m.details.ConditionData().Condition.Name != "staff" {
		t.Fatalf("Enter = handled:%v condition:%v data:%#v", handled, m.details.IsCondition(), m.details.ConditionData())
	}

	m = detailsCrossNavigationTestModel(t)
	data = m.details.Data()
	data.Parameter.Values[0].Plain = false
	data.Parameter.Values[0].Display = rcdisplay.ValueSummary{
		Kind: rcdisplay.ValueSummaryExperiment,
		Text: "⚗ (a/b test)",
	}
	m.details = m.details.SetData(data)
	if cmd := m.requestDeleteDetails(); cmd != nil || m.dialog.IsOpen() {
		t.Fatalf("managed conditional delete = cmd:%v dialog:%v, want disabled", cmd != nil, m.dialog.IsOpen())
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

func TestParameterValueInAppDefaultToggleDisablesTypedEditor(t *testing.T) {
	keys := []tea.Key{{Code: tea.KeyRight}, {Code: 'e', Text: "e"}}
	for _, key := range keys {
		m := detailsCrossNavigationTestModel(t)
		m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'd', Text: "d"}))
		value, ok := m.details.SelectedParameterValue()
		if !handled || !ok || !value.UseInAppDefault {
			t.Fatalf("d = handled:%v selected:%v value:%+v", handled, ok, value)
		}

		m, cmd, handled := m.updateKeyMessage(tea.KeyPressMsg(key))
		if !handled || cmd != nil || m.boolPicker.IsOpen() || m.stringInput.IsOpen() {
			t.Fatalf("%v = handled:%v cmd:%v bool:%v string:%v; want disabled", key.Code, handled, cmd != nil, m.boolPicker.IsOpen(), m.stringInput.IsOpen())
		}
	}
}

func TestParameterValueInAppDefaultToggleReturnsToRemoteValue(t *testing.T) {
	m := detailsCrossNavigationTestModel(t)
	m, _, _ = m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'd', Text: "d"}))
	m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'd', Text: "d"}))

	value, ok := m.details.SelectedParameterValue()
	if !handled || !ok || !value.Plain || value.UseInAppDefault || value.RawValue != "false" {
		t.Fatalf("second d = handled:%v selected:%v value:%+v, want plain false", handled, ok, value)
	}
}

func TestManagedNumberValueCannotOpenEditorOrToggleSource(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	tree := &core.ParametersTree{Groups: []core.ParametersGroup{{
		Key: "__default__", Label: "(root)",
		Parameters: []core.ParametersEntry{{
			Key: "funding_minimum_amount",
			Values: []core.ParametersValue{{
				Label: "default", Value: "◐ 10% → 20 | (no change)", ValueType: "NUMBER",
				Display: rcdisplay.ValueSummary{
					Kind:    rcdisplay.ValueSummaryRollout,
					Text:    "◐ 10% → 20 | (no change)",
					Rollout: &rcdisplay.RolloutSummary{Percentage: "10%", Value: "20"},
				},
			}},
		}},
	}}}
	m := viewTestModel(100, 32, panels.Parameters)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{Project: project, Tree: tree, Source: "cache"})
	if !m.parameters.FocusValue(project.ProjectID, "__default__", "funding_minimum_amount", 0) {
		t.Fatal("failed to focus rollout value")
	}
	m.setActive(panels.Parameters)

	m, cmd, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'e', Text: "e"}))
	if !handled || cmd != nil || m.numberInput.IsOpen() || m.stringInput.IsOpen() {
		t.Fatalf("edit = handled:%v cmd:%v number:%v string:%v, want disabled", handled, cmd != nil, m.numberInput.IsOpen(), m.stringInput.IsOpen())
	}

	data, ok := m.parameters.ParameterViewData(project.ProjectID, "__default__", "funding_minimum_amount", "default")
	if !ok {
		t.Fatal("managed parameter Details data missing")
	}
	m.details = m.details.SetData(data)
	m.detailsVisible = true
	m.setActive(panels.Details)
	for _, key := range []tea.Key{{Code: 'e', Text: "e"}, {Code: 'd', Text: "d"}} {
		m, cmd, handled = m.updateKeyMessage(tea.KeyPressMsg(key))
		if !handled || cmd != nil || m.numberInput.IsOpen() {
			t.Fatalf("%q = handled:%v cmd:%v number:%v, want disabled", key.Text, handled, cmd != nil, m.numberInput.IsOpen())
		}
	}
	value, selected := m.details.SelectedParameterValue()
	if !selected || !value.ReadOnly() || value.UseInAppDefault {
		t.Fatalf("managed value changed after disabled actions: selected:%v value:%+v", selected, value)
	}
}

func TestInAppDefaultActionPaletteAvailability(t *testing.T) {
	m := detailsCrossNavigationTestModel(t)
	m, _, _ = m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'd', Text: "d"}))

	var toggle, edit helpPaletteAction
	for _, item := range m.helpPaletteActions() {
		if item.block != tuiconfig.BlockDetails {
			continue
		}
		switch item.action {
		case tuiconfig.ActionToggleInAppDefault:
			toggle = item
		case tuiconfig.ActionEditValue:
			edit = item
		}
	}
	if !toggle.enabled || !strings.Contains(toggle.title, "in-app default") {
		t.Fatalf("toggle palette action = %+v, want enabled in-app-default action", toggle)
	}
	if edit.enabled || !strings.Contains(edit.reason, "does not use a remote value") {
		t.Fatalf("edit palette action = %+v, want unavailable for in-app default", edit)
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
