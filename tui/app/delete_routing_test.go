package app

import (
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func conditionalParametersTree() *core.ParametersTree {
	return &core.ParametersTree{
		Version:  "1",
		CachedAt: time.Now(),
		Groups: []core.ParametersGroup{
			{
				Key:   "__default__",
				Label: "(root)",
				Parameters: []core.ParametersEntry{
					{
						Key:     "feature_login",
						Summary: "3 values",
						Values: []core.ParametersValue{
							{Label: "android", Value: "a", RawValue: "a", ValueType: "STRING", Plain: true},
							{Label: "ios", Value: "b", RawValue: "b", ValueType: "STRING", Plain: true},
							{Label: "default", Value: "c", RawValue: "c", ValueType: "STRING", Plain: true},
						},
					},
				},
			},
		},
	}
}

func newDeleteRoutingTestModel(t *testing.T) Model {
	t.Helper()
	m := New(nil)
	m.setActive(panels.Parameters)
	m.width = 80
	m.height = 24
	m.dialog = m.dialog.SetBounds(0, 0, m.width, m.height)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{
		Projects: []core.Project{{Name: "Demo", ProjectID: "demo"}},
	})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Tree:    conditionalParametersTree(),
		Source:  "cache",
	})
	m.parameters = m.parameters.SetBounds(0, 0, 80, 20).SetActive(true)
	return m
}

func dialogTitle(t *testing.T, m Model) string {
	t.Helper()
	t.Setenv("NO_COLOR", "1")
	view := m.dialog.View()
	if view == "" {
		return ""
	}
	if strings.Contains(view, "Delete Parameter?") {
		return "Delete Parameter?"
	}
	if strings.Contains(view, "Delete Conditional Value?") {
		return "Delete Conditional Value?"
	}
	if strings.Contains(view, "Delete Conditional Value Failed") {
		return "Delete Conditional Value Failed"
	}
	return view
}

// TestDeleteKeyRoutesFirstConditionalToConditionalDelete guards against routing
// delete on the first conditional value (valueIdx 0) to whole-parameter delete.
func TestDeleteKeyRoutesFirstConditionalToConditionalDelete(t *testing.T) {
	m := newDeleteRoutingTestModel(t)
	if !m.parameters.FocusValue("demo", "__default__", "feature_login", 0) {
		t.Fatal("FocusValue failed for first conditional")
	}

	next, _, handled := m.updateParametersDeleteKey()
	if !handled {
		t.Fatal("updateParametersDeleteKey did not handle delete on first conditional")
	}
	if !next.dialog.IsOpen() {
		t.Fatal("delete dialog did not open")
	}
	title := dialogTitle(t, next)
	if strings.Contains(title, "Delete Parameter?") {
		t.Fatalf("routed to parameter delete, want conditional delete path; dialog:\n%s", next.dialog.View())
	}
	if !strings.Contains(title, "Conditional Value") {
		t.Fatalf("dialog title = %q, want conditional value delete path", title)
	}
}

func TestDeleteKeyRoutesDefaultValueToParameterDelete(t *testing.T) {
	m := newDeleteRoutingTestModel(t)
	if !m.parameters.FocusValue("demo", "__default__", "feature_login", 2) {
		t.Fatal("FocusValue failed for default value")
	}

	next, _, handled := m.updateParametersDeleteKey()
	if !handled {
		t.Fatal("updateParametersDeleteKey did not handle delete on default value")
	}
	if !next.dialog.IsOpen() {
		t.Fatal("delete dialog did not open")
	}
	title := dialogTitle(t, next)
	if !strings.Contains(title, "Delete Parameter?") {
		t.Fatalf("dialog title = %q, want parameter delete path", title)
	}
}

func TestDeleteKeyRoutesDetailsFirstConditionalToConditionalDelete(t *testing.T) {
	m := newDeleteRoutingTestModel(t)
	m.setActive(panels.Details)
	m.detailsVisible = true
	m.details = m.details.SetBounds(0, 0, 60, 20).SetActive(true)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project:          core.Project{Name: "Demo", ProjectID: "demo"},
		GroupKey:         "",
		GroupLabel:       "(root)",
		Parameter:        conditionalParametersTree().Groups[0].Parameters[0],
		SelectedValueIdx: 0,
	})

	next, _, handled := m.updateDeleteKey()
	if !handled {
		t.Fatal("updateDeleteKey did not handle delete from details panel")
	}
	if !next.dialog.IsOpen() {
		t.Fatal("delete dialog did not open")
	}
	title := dialogTitle(t, next)
	if strings.Contains(title, "Delete Parameter?") {
		t.Fatalf("routed to parameter delete, want conditional delete path; dialog:\n%s", next.dialog.View())
	}
	if !strings.Contains(title, "Conditional Value") {
		t.Fatalf("dialog title = %q, want conditional value delete path", title)
	}
}
