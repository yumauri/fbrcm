package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/tui/components/projectio"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestProjectImportAndExportActionsOpenForOneTarget(t *testing.T) {
	svc := newRenameTestService(t)
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m := New(svc)
	m.width, m.height = 100, 30
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{project}})
	m.setActive(panels.Projects)
	m.applyLayout()

	next, _, handled := m.updateGlobalPanelActionKey("i")
	if !handled || !next.projectIO.IsOpen() || next.projectIO.Mode() != projectio.ModeImport {
		t.Fatalf("import action handled=%v open=%v mode=%v", handled, next.projectIO.IsOpen(), next.projectIO.Mode())
	}
	next.projectIO = next.projectIO.Close()
	next, _, handled = next.updateGlobalPanelActionKey("e")
	if !handled || !next.projectIO.IsOpen() || next.projectIO.Mode() != projectio.ModeExport {
		t.Fatalf("export action handled=%v open=%v mode=%v", handled, next.projectIO.IsOpen(), next.projectIO.Mode())
	}
}

func TestProjectIOActionsUseCursorProjectRegardlessOfMarkedProjects(t *testing.T) {
	m := New(newRenameTestService(t))
	m.width, m.height = 100, 30
	projects := []core.Project{{Name: "One", ProjectID: "one"}, {Name: "Two", ProjectID: "two"}}
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: projects})
	m.projects = m.projects.SetActive(true)
	m.projects, _ = m.projects.Update(keyPressForApp(' '))
	m.projects, _ = m.projects.Update(keyPressForApp('j'))
	m.projects, _ = m.projects.Update(keyPressForApp(' '))
	m.setActive(panels.Projects)
	m.applyLayout()

	for _, action := range []string{"i", "e"} {
		next, _, handled := m.updateGlobalPanelActionKey(action)
		if !handled || !next.projectIO.IsOpen() || next.dialog.IsOpen() {
			t.Fatalf("action %q handled=%v io=%v dialog=%v", action, handled, next.projectIO.IsOpen(), next.dialog.IsOpen())
		}
		if view := ansi.Strip(next.projectIO.View()); !strings.Contains(view, "Project: Two (two)") {
			t.Fatalf("action %q did not target cursor project:\n%s", action, view)
		}
	}
}

func TestExportProjectCmdWritesNormalizedDraft(t *testing.T) {
	svc := newRenameTestService(t)
	raw := []byte(`{"version":{"versionNumber":"1"},"parameters":{"flag":{"defaultValue":{"value":"\u003con\u003e"}}}}`)
	cache := &config.ParametersCache{ETag: "etag-1", CachedAt: time.Now().UTC(), RemoteConfig: raw}
	if err := config.SaveParametersCache("demo", cache); err != nil {
		t.Fatalf("SaveParametersCache = %v", err)
	}
	if err := svc.SaveDraft("demo", raw); err != nil {
		t.Fatalf("SaveDraft = %v", err)
	}
	path := filepath.Join(t.TempDir(), "export.json")
	m := New(svc)
	msg, ok := m.exportProjectCmd(projectExportSession{project: core.Project{ProjectID: "demo"}, path: path, draft: true})().(projectExportCompletedMsg)
	if !ok || msg.err != nil {
		t.Fatalf("export result = %#v", msg)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile = %v", err)
	}
	if !strings.Contains(string(body), `"value": "<on>"`) {
		t.Fatalf("export body = %s", body)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("export mode = %v err=%v", info.Mode().Perm(), err)
	}
}

func TestProjectImportValidationDialogCloseButtonAcceptsMouseClick(t *testing.T) {
	m := viewTestModel(120, 30, panels.Logs)
	m.openErrorDialog(
		"Import Failed",
		core.Project{Name: "Demo", ProjectID: "demo"},
		"firebase error: validate remote config api returned 400 Bad Request: {\n  \"error\": {\n    \"code\": 400\n  }\n}",
	)
	x, y := dialogLabelPoint(t, m, "Close")

	nextModel, cmd := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	next, ok := nextModel.(Model)
	if !ok {
		t.Fatalf("updated model type = %T", nextModel)
	}
	if next.dialog.IsOpen() || cmd == nil {
		t.Fatalf("Close click left dialog open=%v cmd nil=%v", next.dialog.IsOpen(), cmd == nil)
	}
}

func TestProjectIOPopupEnablesAndRoutesMouseClicks(t *testing.T) {
	m := viewTestModel(100, 30, panels.Projects)
	var cmd tea.Cmd
	m.projectIO, cmd = m.projectIO.OpenExport(core.Project{Name: "Demo", ProjectID: "demo"}, true)
	if cmd != nil {
		t.Fatalf("OpenExport command = %v, want nil on source step", cmd)
	}
	if got := m.View().MouseMode; got != tea.MouseModeAllMotion {
		t.Fatalf("project I/O mouse mode = %v, want all motion", got)
	}
	x, y := projectIOLabelPoint(t, m, "Continue")
	nextModel, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	next, ok := nextModel.(Model)
	if !ok {
		t.Fatalf("updated model type = %T", nextModel)
	}
	if view := ansi.Strip(next.projectIO.View()); !strings.Contains(view, "Destination:") {
		t.Fatalf("Continue click did not open export destination:\n%s", view)
	}
}

func TestSuccessfulImportSelectsImportedProjectAndShowsParameters(t *testing.T) {
	alpha := core.Project{Name: "Alpha", ProjectID: "alpha"}
	beta := core.Project{Name: "Beta", ProjectID: "beta"}
	m := New(newRenameTestService(t))
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{alpha, beta}})
	m.projects = m.projects.SetActive(true)
	m.projects, _ = m.projects.Update(keyPressForApp(' '))
	m.projects, _ = m.projects.Update(keyPressForApp('j'))
	m.projects, _ = m.projects.Update(keyPressForApp(' '))
	m.setActive(panels.Projects)
	tree := &core.ParametersTree{Version: "2"}

	next, cmd, handled := m.updateProjectImportCompleted(projectImportCompletedMsg{
		plan:   &core.ProjectImportPlan{Project: beta},
		result: &core.ProjectImportResult{Tree: tree, Published: true},
	})
	if !handled || cmd == nil || !next.dialog.IsOpen() {
		t.Fatalf("completion handled=%v cmd nil=%v dialog open=%v", handled, cmd == nil, next.dialog.IsOpen())
	}
	if next.active != panels.Parameters {
		t.Fatalf("active panel = %v, want Parameters", next.active)
	}
	if got := next.projects.ActionTargets(); len(got) != 1 || got[0].ProjectID != beta.ProjectID {
		t.Fatalf("selected projects = %+v, want only beta", got)
	}
}

func dialogLabelPoint(t *testing.T, m Model, label string) (int, int) {
	t.Helper()
	x, y := m.dialog.Position()
	for row, line := range strings.Split(ansi.Strip(m.dialog.View()), "\n") {
		if before, _, found := strings.Cut(line, label); found {
			return x + lipgloss.Width(before), y + row
		}
	}
	t.Fatalf("dialog does not render %q", label)
	return 0, 0
}

func projectIOLabelPoint(t *testing.T, m Model, label string) (int, int) {
	t.Helper()
	x, y := m.projectIO.Position()
	foundX, foundY, found := 0, 0, false
	for row, line := range strings.Split(ansi.Strip(m.projectIO.View()), "\n") {
		if before, _, match := strings.Cut(line, label); match {
			foundX, foundY, found = x+lipgloss.Width(before), y+row, true
		}
	}
	if found {
		return foundX, foundY
	}
	t.Fatalf("project I/O popup does not render %q", label)
	return 0, 0
}
