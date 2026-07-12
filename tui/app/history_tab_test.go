package app

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestHistoryRemainsSelectedAcrossProjectsFocus(t *testing.T) {
	m := New(nil)
	m.setActive(panels.History)
	m.setActive(panels.Projects)

	if got := m.nextTabPanel(); got != panels.History {
		t.Fatalf("next tab from Projects = %v, want History", got)
	}

	next, _, handled := m.updateAppMessage(messages.SetActivePanelMsg{Panel: panels.Parameters})
	if !handled || next.active != panels.History {
		t.Fatalf("project selection active = %v, handled=%v; want History", next.active, handled)
	}
}

func TestExplicitParametersActivationReplacesHistoryTab(t *testing.T) {
	m := New(nil)
	m.setActive(panels.History)
	m.setActive(panels.Parameters)
	m.setActive(panels.Projects)

	if got := m.nextTabPanel(); got != panels.Parameters {
		t.Fatalf("next tab from Projects = %v, want Parameters", got)
	}
}

func TestSingleProjectSelectionReplacesHistoryWithParameters(t *testing.T) {
	m := New(nil)
	m.setActive(panels.History)
	m.setActive(panels.Projects)

	next, _, _ := m.updateAppMessage(messages.SetActivePanelMsg{
		Panel:              panels.Parameters,
		ResetParametersTab: true,
	})
	if next.active != panels.Parameters || next.parametersTab != panels.Parameters {
		t.Fatalf("single-project selection active=%v tab=%v; want Parameters", next.active, next.parametersTab)
	}
}

func TestMultiProjectSelectionPreservesHistoryTab(t *testing.T) {
	m := New(nil)
	m.setActive(panels.History)
	m.setActive(panels.Projects)

	next, _ := m.updateChildPanels(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{
		{ProjectID: "first", Name: "First"},
		{ProjectID: "second", Name: "Second"},
	}})
	if next.active != panels.Projects || next.parametersTab != panels.History {
		t.Fatalf("multiselect active=%v tab=%v; want Projects focus with History retained", next.active, next.parametersTab)
	}
}
