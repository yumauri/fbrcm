package app

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	parameterscmp "github.com/yumauri/fbrcm/tui/components/parameters"
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

func TestHistoryDiffRequestOpensGenericDiffModal(t *testing.T) {
	m := New(nil)
	m.width, m.height = 80, 24
	next, _, handled := m.updateAppMessage(parameterscmp.HistoryDiffRequestedMsg{
		Project: core.Project{ProjectID: "demo", Name: "Demo"},
		Input: dictdiff.Input{
			EntityName: "Parameter: WEB / flag",
			Left: dictdiff.NamedDictionary{
				Name:       "Earlier version: v1",
				Properties: dictdiff.Dictionary{"value · default": dictdiff.Boolean(true)},
			},
			Right: dictdiff.NamedDictionary{
				Name:       "Later version: v2",
				Properties: dictdiff.Dictionary{"value · default": dictdiff.Boolean(false)},
			},
		},
	})
	if !handled || !next.diffView.IsOpen() {
		t.Fatalf("history diff request = handled:%v open:%v", handled, next.diffView.IsOpen())
	}
	view := next.diffView.View()
	for _, want := range []string{
		"Parameter:", "WEB / flag", "Earlier version:", "v1",
		"Later version:", "v2", "value · default",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("history diff modal misses %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "alice@example.com") {
		t.Fatalf("ordinary History diff unexpectedly shows blame attribution:\n%s", view)
	}
}

func TestParameterBlameLoadedOpensAttributedDiffModal(t *testing.T) {
	m := New(nil)
	m.width, m.height = 80, 24
	version := firebase.RemoteConfigVersion{
		VersionNumber: "142",
		UpdateTime:    "2026-09-03T12:34:58Z",
		UpdateUser:    firebase.RemoteConfigUser{Name: "Alice Smith", Email: "alice@example.com"},
	}
	next, _, handled := m.updateAppMessage(parameterBlameLoadedMsg{
		project: core.Project{ProjectID: "demo", Name: "Demo"},
		history: core.ParameterHistory{Changes: []core.ParameterHistoryChange{{
			PreviousVersion: "141",
			Version:         version,
			Change: rcdiff.ParameterChange{
				Key: "flag", Group: "WEB", Kind: rcdiff.ChangeChanged,
				Current: &firebase.RemoteConfigParam{ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "false"}},
				Final:   &firebase.RemoteConfigParam{ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "true"}},
			},
		}}},
	})
	if !handled || !next.diffView.IsOpen() {
		t.Fatalf("blame result = handled:%v open:%v", handled, next.diffView.IsOpen())
	}
	view := next.diffView.View()
	for _, want := range []string{
		"Parameter:", "WEB / flag", "Earlier version:", "v141", "Later version:", "v142",
		rcdisplay.FormatRemoteConfigVersionAttribution(version), "false", "true",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("blame diff modal misses %q:\n%s", want, view)
		}
	}
}

func TestParameterBlameWithoutDirectChangeShowsExplanation(t *testing.T) {
	m := New(nil)
	m.dialog = m.dialog.SetBounds(0, 0, 80, 24)
	next, cmd, handled := m.updateAppMessage(parameterBlameLoadedMsg{
		project: core.Project{ProjectID: "demo", Name: "Demo"},
		history: core.ParameterHistory{Boundary: &core.ParameterHistoryBoundary{
			Version: "1", Present: true,
		}},
	})
	if !handled || cmd != nil || !next.dialog.IsOpen() {
		t.Fatalf("empty blame = handled:%v cmd:%v dialog:%v", handled, cmd != nil, next.dialog.IsOpen())
	}
	if view := next.dialog.View(); !strings.Contains(view, "oldest retained version") || !strings.Contains(view, "v1") {
		t.Fatalf("empty blame explanation missing boundary:\n%s", view)
	}
}
