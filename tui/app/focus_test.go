package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestTabFollowsAutomaticPanelActivation(t *testing.T) {
	m := viewTestModel(100, 30, panels.Projects)
	m = pressTab(t, m, panels.Projects)
	m = updateFocus(t, m, messages.SetActivePanelMsg{Panel: panels.Parameters}, panels.Parameters)
	m = pressTab(t, m, panels.Projects)
	m = pressTab(t, m, panels.Parameters)

	m = updateFocus(t, m, messages.ParameterSelectionChangedMsg{
		Data: &messages.ParameterViewData{
			Project:    core.Project{ProjectID: "demo"},
			GroupKey:   rootgroup.TreeKey,
			GroupLabel: rootgroup.Label,
			Parameter: core.ParametersEntry{Key: "feature", Values: []core.ParametersValue{{
				Label: "default", RawValue: "off", Value: "off", ValueType: "STRING", Plain: true,
			}}},
			SelectedValueIdx: -1,
		},
		Activate: true,
	}, panels.Details)
	m = pressTab(t, m, panels.Parameters)
	m = pressTab(t, m, panels.Details)

	// Clicking Projects while Details is open pairs Projects with Details.
	m = updateFocus(t, m, tea.MouseClickMsg{X: 1, Y: 1, Button: tea.MouseLeft}, panels.Projects)
	m = pressTab(t, m, panels.Details)
	m = pressTab(t, m, panels.Projects)
	m = pressTab(t, m, panels.Details)

	m = updateFocus(t, m, tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}), panels.Parameters)
	pressTab(t, m, panels.Projects)
}

func TestTabStaysInWorkspaceMenu(t *testing.T) {
	m := viewTestModel(80, 24, panels.Conditions)
	m = m.openWorkspaceMenu()
	if !m.workspaceMenu {
		t.Fatal("workspace menu did not open")
	}

	m = pressTab(t, m, panels.Conditions)
	if !m.workspaceMenu {
		t.Fatal("Tab closed the workspace menu")
	}
	m = updateFocus(t, m, tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}), panels.Conditions)
	pressTab(t, m, panels.Projects)
}

func TestTabStaysInDetailsDropdown(t *testing.T) {
	m := viewTestModel(100, 30, panels.Conditions)
	m = updateFocus(t, m, messages.ConditionSelectionChangedMsg{
		Data: &messages.ConditionViewData{
			Project:   core.Project{ProjectID: "demo"},
			Condition: core.ConditionEntry{Name: "beta"},
		},
		Activate: true,
	}, panels.Details)
	m = updateFocus(t, m, keyPress('c'), panels.Details)
	if !m.details.DropdownOpen() {
		t.Fatal("Details color dropdown did not open")
	}

	m = pressTab(t, m, panels.Details)
	if !m.details.DropdownOpen() {
		t.Fatal("Tab closed the Details color dropdown")
	}
	m = updateFocus(t, m, tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}), panels.Details)
	pressTab(t, m, panels.Conditions)
}

func pressTab(t *testing.T, m Model, want panels.ID) Model {
	t.Helper()
	return updateFocus(t, m, tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}), want)
}

func updateFocus(t *testing.T, m Model, msg tea.Msg, want panels.ID) Model {
	t.Helper()
	next, _ := m.Update(msg)
	updated := next.(Model)
	if updated.active != want {
		t.Fatalf("active panel after %v = %v, want %v", msg, updated.active, want)
	}
	return updated
}
