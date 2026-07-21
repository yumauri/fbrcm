package app

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestProjectAuthPickerRebindsExactCurrentProject(t *testing.T) {
	svc := newRenameTestService(t)
	for _, id := range []string{"main", "work"} {
		if _, err := svc.AddGCloudAuth(id, id); err != nil {
			t.Fatalf("AddGCloudAuth(%q) = %v", id, err)
		}
	}
	projects := []config.Project{
		{Name: "Demo", ProjectID: "demo", AuthID: "main", DiscoveredBy: []string{"main", "work"}},
		{Name: "demo", ProjectID: "other", AuthID: "main", DiscoveredBy: []string{"main"}},
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	m := New(svc)
	m.width, m.height = 90, 24
	m.logsSized = true
	m.logsHeight = defaultLogsPanelHeight
	m.setActive(panels.Projects)
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: projects, Source: "cache"})
	m.applyLayout()
	if cmd := m.openProjectAuthPicker(); cmd != nil {
		t.Fatal("configured identities unexpectedly returned a setup command")
	}
	if !m.authPicker.IsOpen() || m.authBind == nil {
		t.Fatal("project auth picker did not open")
	}
	pickerView := ansi.Strip(m.authPicker.View())
	firstLine, _, _ := strings.Cut(pickerView, "\n")
	if !strings.Contains(firstLine, "Bind authentication") || strings.Contains(firstLine, "demo") {
		t.Fatalf("picker title includes project identity: %q", firstLine)
	}
	if !strings.Contains(pickerView, "Project: Demo (demo)") {
		t.Fatalf("picker body does not identify project:\n%s", pickerView)
	}
	m.authPicker.Move(1)
	cmd := m.submitAuthBinding()
	if cmd == nil {
		t.Fatal("auth binding command is nil")
	}
	msg, ok := cmd().(projectAuthBoundMsg)
	if !ok || msg.err != nil {
		t.Fatalf("binding result = %#v", msg)
	}
	m, _, handled := m.updateProjectAuthBound(msg)
	if !handled {
		t.Fatal("binding result was not handled")
	}

	demo, err := svc.ProjectByID("demo")
	if err != nil {
		t.Fatalf("ProjectByID demo = %v", err)
	}
	if demo.AuthID != "work" || strings.Join(demo.DiscoveredBy, ",") != "main,work" {
		t.Fatalf("rebound demo = %+v, want work binding and unchanged provenance", demo)
	}
	other, err := svc.ProjectByID("other")
	if err != nil {
		t.Fatalf("ProjectByID other = %v", err)
	}
	if other.AuthID != "main" {
		t.Fatalf("same-name project binding = %q, want main", other.AuthID)
	}
	if view := ansi.Strip(m.projects.View(true)); strings.Contains(view, "auth:") || strings.Contains(view, "work") {
		t.Fatalf("Projects panel exposes authentication identity:\n%s", view)
	}
}

func TestProjectAuthBindingDisabledWithOneIdentity(t *testing.T) {
	svc := newRenameTestService(t)
	if _, err := svc.AddGCloudAuth("main", "main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	projects := []core.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	m := New(svc)
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: projects})
	if m.authCount != 1 {
		t.Fatalf("authCount = %d, want 1", m.authCount)
	}
	if cmd := m.openProjectAuthPicker(); cmd != nil || m.authPicker.IsOpen() {
		t.Fatalf("one-auth binding = cmd:%v open:%v, want disabled", cmd != nil, m.authPicker.IsOpen())
	}
	if enabled, reason := m.contextualHelpActionAvailability(tuiconfig.BlockProjects, tuiconfig.ActionBindAuth); enabled || !strings.Contains(reason, "at least two") {
		t.Fatalf("binding availability = %v, %q", enabled, reason)
	}
}

func TestProjectAuthBindingDisabledWithOneDiscoveredIdentity(t *testing.T) {
	svc := newRenameTestService(t)
	for _, id := range []string{"main", "work"} {
		if _, err := svc.AddGCloudAuth(id, id); err != nil {
			t.Fatalf("AddGCloudAuth(%q) = %v", id, err)
		}
	}
	projects := []core.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main", DiscoveredBy: []string{"main"}}}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	m := New(svc)
	m.width, m.height = 90, 24
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: projects})
	if cmd := m.openProjectAuthPicker(); cmd != nil {
		t.Fatal("auth picker unexpectedly returned command")
	}
	if m.authPicker.IsOpen() {
		t.Fatal("auth picker opened with only one eligible identity")
	}
	if enabled, reason := m.contextualHelpActionAvailability(tuiconfig.BlockProjects, tuiconfig.ActionBindAuth); enabled || !strings.Contains(reason, "discover every selected project") {
		t.Fatalf("binding availability = %v, %q", enabled, reason)
	}
}

func TestProjectAuthBindingIgnoredForDisabledProject(t *testing.T) {
	m := New(nil)
	m.authCount = 2
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{{
		ProjectID: "disabled", Disabled: true, DiscoveredBy: []string{"main", "work"},
	}}})
	if cmd := m.openProjectAuthPicker(); cmd != nil || m.authPicker.IsOpen() {
		t.Fatalf("disabled project binding = cmd:%v open:%v, want ignored", cmd != nil, m.authPicker.IsOpen())
	}
}

func TestProjectAuthPickerCancelButtonClosesWithoutBinding(t *testing.T) {
	m := New(nil)
	m.authBind = &authBindingSession{targets: []core.Project{{ProjectID: "demo"}}}
	m.authPicker = m.authPicker.SetBounds(0, 0, 80, 24).Open("Bind authentication", []string{"Project: demo"}, nil, 0)
	m.authPicker.MoveButton(1)
	next, cmd, handled := m.updateAuthPicker(keyPressForApp(tea.KeyEnter))
	if !handled || cmd != nil || next.authPicker.IsOpen() || next.authBind != nil {
		t.Fatalf("cancel button handled=%v cmd=%v open=%v session=%v", handled, cmd != nil, next.authPicker.IsOpen(), next.authBind != nil)
	}
}

func TestProjectAuthPickerUsesMarkedProjectsAsBatchTargets(t *testing.T) {
	m := New(nil)
	projects := []core.Project{{ProjectID: "one"}, {ProjectID: "two"}, {ProjectID: "three"}}
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: projects})
	m.projects = m.projects.SetActive(true)
	m.projects, _ = m.projects.Update(keyPressForApp(' '))
	m.projects, _ = m.projects.Update(keyPressForApp('j'))
	m.projects, _ = m.projects.Update(keyPressForApp(' '))
	targets := m.projects.ActionTargets()
	if len(targets) != 2 || targets[0].ProjectID != "one" || targets[1].ProjectID != "two" {
		t.Fatalf("batch targets = %+v, want one and two", targets)
	}
}

func keyPressForApp(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}
