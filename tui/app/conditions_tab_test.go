package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestConditionsTabRemainsSelectedAcrossProjectsFocus(t *testing.T) {
	m := New(nil)
	m.setActive(panels.Conditions)
	m.setActive(panels.Projects)

	if got := m.nextTabPanel(); got != panels.Conditions {
		t.Fatalf("next tab from Projects = %v, want Conditions", got)
	}
	next, _, _ := m.updateAppMessage(messages.SetActivePanelMsg{Panel: panels.Parameters})
	if next.active != panels.Conditions {
		t.Fatalf("project selection active = %v; want Conditions", next.active)
	}
}

func TestGlobalFocusKeysWorkFromConditionDetails(t *testing.T) {
	m := New(nil)
	m.setActive(panels.Conditions)
	m.applyConditionSelection(messages.ConditionSelectionChangedMsg{
		Data: &messages.ConditionViewData{
			Project:   core.Project{ProjectID: "demo"},
			Condition: core.ConditionEntry{Name: "beta"},
		},
		Activate: true,
	})

	next, _, handled := m.updateAppMessage(tea.KeyPressMsg(tea.Key{Code: '2'}))
	if !handled || next.active != panels.Parameters {
		t.Fatalf("2 from condition Details: active=%v handled=%v; want Parameters, true", next.active, handled)
	}

	next.setActive(panels.Details)
	next, _, handled = next.updateAppMessage(tea.KeyPressMsg(tea.Key{Code: '3'}))
	if !handled || next.active != panels.Conditions {
		t.Fatalf("3 from condition Details: active=%v handled=%v; want Conditions, true", next.active, handled)
	}
}

func TestGlobalFocusKeyWorksFromParameterDetails(t *testing.T) {
	m := New(nil)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project:   core.Project{ProjectID: "demo"},
		Parameter: core.ParametersEntry{Key: "feature"},
	})
	m.detailsVisible = true
	m.setActive(panels.Details)

	next, _, handled := m.updateAppMessage(tea.KeyPressMsg(tea.Key{Code: '3'}))
	if !handled || next.active != panels.Conditions {
		t.Fatalf("3 from parameter Details: active=%v handled=%v; want Conditions, true", next.active, handled)
	}
}

func TestGlobalFocusKeyDoesNotStealDetailsFieldInput(t *testing.T) {
	m := New(nil)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project:   core.Project{ProjectID: "demo"},
		Parameter: core.ParametersEntry{Key: "feature"},
	})
	m.detailsVisible = true
	m.setActive(panels.Details)
	m.details, _ = m.details.ActivateName()

	next, _, handled := m.updateAppMessage(tea.KeyPressMsg(tea.Key{Code: '2'}))
	if !handled || next.active != panels.Details {
		t.Fatalf("2 while editing Details: active=%v handled=%v; want Details, true", next.active, handled)
	}
}

func TestConditionSelectionOpensReadOnlyDetailsAndReturnsToConditions(t *testing.T) {
	m := New(nil)
	m.setActive(panels.Conditions)
	data := &messages.ConditionViewData{
		Project: core.Project{ProjectID: "demo", Name: "Demo"},
		Condition: core.ConditionEntry{
			Priority:   1,
			Name:       "beta",
			Expression: "app.version > '2'",
		},
	}

	m.applyConditionSelection(messages.ConditionSelectionChangedMsg{Data: data, Activate: true})
	if !m.detailsVisible || m.active != panels.Details || !m.details.IsCondition() {
		t.Fatalf("condition details state = visible:%v active:%v condition:%v", m.detailsVisible, m.active, m.details.IsCondition())
	}
	m.closeDetailsPanel()
	if m.detailsVisible || m.active != panels.Conditions {
		t.Fatalf("closed condition details state = visible:%v active:%v", m.detailsVisible, m.active)
	}
}

func TestConditionMouseClickDoesNotReachHiddenParametersPanel(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := viewTestModel(100, 30, panels.Conditions)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project: project,
		Tree: &core.ParametersTree{Groups: []core.ParametersGroup{{
			Key:   "__default__",
			Label: "(root)",
			Parameters: []core.ParametersEntry{{
				Key: "hidden_parameter",
				Values: []core.ParametersValue{{
					Label: "default", Value: "hidden", RawValue: "hidden", ValueType: "STRING", Plain: true,
				}},
			}},
		}}},
	})
	m.conditions, _ = m.conditions.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.conditions, _ = m.conditions.Update(messages.ConditionsLoadedMsg{
		Project: project,
		Tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{
			{Priority: 1, Name: "first", Expression: "true"},
			{Priority: 2, Name: "second", Expression: "true"},
		}},
	})
	m.details = m.details.SetConditionData(&messages.ConditionViewData{
		Project: project, Condition: core.ConditionEntry{Name: "first"},
	})
	m.detailsVisible = true
	m.setActive(panels.Conditions)
	m.applyLayout()

	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	next, cmd, handled := m.updatePanelMouseMessage(tea.MouseClickMsg{
		X: layout.leftWidth + 2, Y: 3, Button: tea.MouseLeft,
	})
	if !handled || cmd == nil {
		t.Fatalf("condition click handled=%v cmd=%v; want true and selection command", handled, cmd)
	}
	selection, ok := cmd().(messages.ConditionSelectionChangedMsg)
	if !ok || selection.Data == nil || selection.Data.Condition.Name != "second" {
		t.Fatalf("condition click emitted %#v; want second condition selection", selection)
	}
	next.applyConditionSelection(selection)
	if !next.details.IsCondition() || next.details.ConditionData().Condition.Name != "second" {
		t.Fatalf("Details after condition click = %#v; want second condition", next.details.ConditionData())
	}
	if _, ok := next.parameters.CurrentParameterViewData(); ok {
		t.Fatal("hidden Parameters panel selected a parameter from the condition click coordinates")
	}
}

func TestConditionsUsesWorkspaceMaximizeBinding(t *testing.T) {
	m := New(nil)
	m.setActive(panels.Conditions)

	next, _, handled := m.updateGlobalKeyMessage("z")
	if !handled || next.projectsMode != projectsPanelModeCollapsed || next.logsMode != logsPanelModeCollapsed {
		t.Fatalf("z from Conditions: handled=%v projects=%v logs=%v; want maximized workspace", handled, next.projectsMode, next.logsMode)
	}
	next, _, handled = next.updateGlobalKeyMessage("z")
	if !handled || next.projectsMode != projectsPanelModeExpanded || next.logsMode != logsPanelModeExpanded {
		t.Fatalf("second z from Conditions: handled=%v projects=%v logs=%v; want restored workspace", handled, next.projectsMode, next.logsMode)
	}
}

func TestConditionsUsesParameterReloadBindings(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	newModel := func() Model {
		m := New(nil)
		m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
		m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
			Project: project,
			Tree:    &core.ParametersTree{Version: "1"},
			Source:  "cache",
		})
		m.conditions, _ = m.conditions.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
		m.conditions, _ = m.conditions.Update(messages.ConditionsLoadedMsg{
			Project: project,
			Tree:    &core.ConditionsTree{Version: "1"},
			Source:  "cache",
		})
		m.setActive(panels.Conditions)
		return m
	}

	for _, key := range []string{"u", "U"} {
		m := newModel()
		next, cmd, handled := m.updateGlobalKeyMessage(key)
		if !handled || cmd == nil {
			t.Fatalf("%s from Conditions: handled=%v cmd=%v; want refresh command", key, handled, cmd)
		}
		if got := next.conditions.ViewWithBorder(true, true); strings.Contains(got, "staled") {
			t.Fatalf("%s did not mark Conditions project as reloading:\n%s", key, got)
		}
	}
}
