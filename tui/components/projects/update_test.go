package projects

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/styles"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestProjectsViewSnapshotLoaded(t *testing.T) {
	m := loadedProjectsModel()

	got := testutil.NormalizeViewSnapshot(m.View(true))
	if got != loadedProjectsSnapshot {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, loadedProjectsSnapshot)
	}
}

func TestProjectsKeyMovementAndSelection(t *testing.T) {
	m := loadedProjectsModel()

	m, _ = m.Update(keyPress(tea.KeyDown))
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}

	m, cmd := m.Update(keyPress(tea.KeyEnter))
	if cmd == nil {
		t.Fatalf("selection command is nil")
	}
	msgs := runBatch(t, cmd)
	selection := findMsg[messages.ProjectsSelectionChangedMsg](msgs)
	if len(selection.Projects) != 1 || selection.Projects[0].ProjectID != "beta" {
		t.Fatalf("selected projects = %+v, want beta", selection.Projects)
	}
	active := findMsg[messages.SetActivePanelMsg](msgs)
	if active.Panel != panels.Parameters {
		t.Fatalf("active panel = %v, want Parameters", active.Panel)
	}
	if !active.ResetParametersTab {
		t.Fatal("Enter selection did not request the Parameters tab")
	}
}

func TestProjectsUpdateIgnoredWhileLoading(t *testing.T) {
	m := loadedProjectsModel()
	m.loading = true

	next, cmd := m.Update(keyPress('u'))

	if cmd != nil {
		t.Fatalf("update returned command while loading")
	}
	if !next.loading {
		t.Fatalf("loading changed to false")
	}
}

func TestProjectsMarkChangesSelectionWithoutChangingPanel(t *testing.T) {
	m := loadedProjectsModel()
	_, cmd := m.Update(keyPress(tea.KeySpace))
	if cmd == nil {
		t.Fatal("mark selection command is nil")
	}
	msgs := runBatch(t, cmd)
	selection := findMsg[messages.ProjectsSelectionChangedMsg](msgs)
	if len(selection.Projects) != 1 || selection.Projects[0].ProjectID != "alpha" {
		t.Fatalf("marked projects = %+v, want alpha", selection.Projects)
	}
	for _, msg := range msgs {
		if _, ok := msg.(messages.SetActivePanelMsg); ok {
			t.Fatal("Space multiselect unexpectedly requested a panel change")
		}
	}
}

func TestDisabledProjectIgnoresSelectionMarkAndOpen(t *testing.T) {
	m := disabledProjectsModel()
	for _, key := range []tea.KeyPressMsg{keyPress(tea.KeyEnter), keyPress(tea.KeySpace), keyPress('o')} {
		next, cmd := m.Update(key)
		if cmd != nil {
			t.Fatalf("disabled action %q returned command", key.String())
		}
		if len(next.selected) != 0 {
			t.Fatalf("disabled action %q selected projects = %#v", key.String(), next.selected)
		}
	}
	if cmd := m.SelectOnly("alpha"); cmd != nil || len(m.selected) != 0 {
		t.Fatalf("SelectOnly disabled project = cmd:%v selected:%#v", cmd != nil, m.selected)
	}
}

func TestDisabledProjectUsesInactiveTabColorAndCompactSeparator(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := disabledProjectsModel()
	if got, want := m.projectStateStyle(0).GetForeground(), styles.PanelTitleInactiveTab.GetForeground(); !reflect.DeepEqual(got, want) {
		t.Fatalf("disabled foreground = %#v, want inactive tab %#v", got, want)
	}
	m.cursor = 1
	m.syncViewport()
	rendered := m.View(true)
	for _, rune := range []string{"A", "a"} {
		if !strings.Contains(rendered, styles.PanelTitleInactiveTab.Render(rune)) {
			t.Fatalf("disabled project rune %q does not use inactive tab style", rune)
		}
	}
	view := ansi.Strip(rendered)
	if !strings.Contains(view, "alpha · disabled") || strings.Contains(view, "alpha  · disabled") {
		t.Fatalf("disabled project separator is not compact:\n%s", view)
	}
}

func TestProjectBecomingDisabledIsUnselected(t *testing.T) {
	m := loadedProjectsModel()
	m.selected["alpha"] = struct{}{}
	cmd := m.ApplyProjectUpdates([]core.Project{{Name: "Alpha Project", ProjectID: "alpha", Disabled: true}})
	if cmd == nil || len(m.selected) != 0 {
		t.Fatalf("disabled update = cmd:%v selected:%#v, want cleared selection notification", cmd != nil, m.selected)
	}
	selection, ok := cmd().(messages.ProjectsSelectionChangedMsg)
	if !ok || len(selection.Projects) != 0 {
		t.Fatalf("disabled selection update = %#v, want empty project selection", selection)
	}
}

func TestProjectsFilterApplySelectsCurrentAndReleasesKeyboard(t *testing.T) {
	m := loadedProjectsModel()
	m, cmd := m.Update(keyText("/"))
	if cmd == nil {
		t.Fatalf("filter activation command is nil")
	}
	if !m.filter.Focused() {
		t.Fatalf("filter is not focused")
	}

	m, _ = m.Update(tea.PasteMsg{Content: "beta"})
	if len(m.projects) != 1 || m.projects[0].ProjectID != "beta" {
		t.Fatalf("filtered projects = %+v, want beta", m.projects)
	}

	m, cmd = m.Update(keyPress(tea.KeyEnter))
	msgs := runBatch(t, cmd)
	selection := findMsg[messages.ProjectsSelectionChangedMsg](msgs)
	if len(selection.Projects) != 1 || selection.Projects[0].ProjectID != "beta" {
		t.Fatalf("selected projects = %+v, want beta", selection.Projects)
	}
	capture := findMsg[messages.KeyboardCaptureMsg](msgs)
	if capture.Enabled {
		t.Fatalf("keyboard capture enabled, want false")
	}
	active := findMsg[messages.SetActivePanelMsg](msgs)
	if active.Panel != panels.Parameters {
		t.Fatalf("active panel = %v, want Parameters", active.Panel)
	}
	if !active.ResetParametersTab {
		t.Fatal("filtered Enter selection did not request the Parameters tab")
	}
}

func TestProjectsMouseWheelMovesCursor(t *testing.T) {
	m := loadedProjectsModel()

	m, _ = m.Update(tea.MouseWheelMsg{X: 2, Y: 2, Button: tea.MouseWheelDown})

	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
}

func TestActionTargetsUsesMarkedProjectsOrCurrentProject(t *testing.T) {
	m := loadedProjectsModel()
	m.cursor = 1
	if got := m.ActionTargets(); len(got) != 1 || got[0].ProjectID != "beta" {
		t.Fatalf("unmarked targets = %+v, want current beta", got)
	}
	m.selected["alpha"] = struct{}{}
	m.selected["gamma"] = struct{}{}
	got := m.ActionTargets()
	if len(got) != 2 || got[0].ProjectID != "alpha" || got[1].ProjectID != "gamma" {
		t.Fatalf("marked targets = %+v, want alpha and gamma", got)
	}
}

func TestTemplateRowsArePrimaryFirstAndCursorFollowsReorder(t *testing.T) {
	m := New(nil).SetBounds(0, 0, 32, 12).SetActive(true)
	project := core.Project{
		Name:            "Alpha Project",
		ProjectID:       "alpha",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Client,
	}
	m, _ = m.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{project}, Source: "cache"})
	if got := []string{m.allProjects[0].ProjectID, m.allProjects[1].ProjectID}; !reflect.DeepEqual(got, []string{"alpha", "server@alpha"}) {
		t.Fatalf("initial target order = %v", got)
	}
	m.cursor = 1
	project.PrimaryTemplate = rctarget.Server
	m, _ = m.Update(messages.ProjectTemplatePreferencesUpdatedMsg{
		Project: project,
	})
	if got := []string{m.allProjects[0].ProjectID, m.allProjects[1].ProjectID}; !reflect.DeepEqual(got, []string{"server@alpha", "alpha"}) {
		t.Fatalf("primary target order = %v", got)
	}
	current, ok := m.CurrentProject()
	if !ok || current.ProjectID != "server@alpha" || m.cursor != 0 {
		t.Fatalf("current after reorder = %#v, cursor %d", current, m.cursor)
	}
}

func TestToggleTemplatesKeyPersistsAndExpandsProject(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	project := core.Project{Name: "Alpha Project", ProjectID: "alpha", AuthID: "main"}
	if err := config.SaveProjects([]config.Project{project}, time.Now()); err != nil {
		t.Fatal(err)
	}
	projects, err := config.LoadProjects()
	if err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	m := New(svc).SetBounds(0, 0, 32, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{Projects: projects, Source: "cache"})
	m, cmd := m.Update(keyPress('t'))
	if cmd == nil {
		t.Fatal("toggle templates key returned no command")
	}
	msg, ok := cmd().(messages.ProjectTemplatePreferencesUpdatedMsg)
	if !ok || msg.Err != nil {
		t.Fatalf("toggle templates result = %#v", msg)
	}
	m, _ = m.Update(msg)
	if len(m.allProjects) != 2 || m.allProjects[0].ProjectID != "alpha" || m.allProjects[1].ProjectID != "server@alpha" {
		t.Fatalf("expanded targets = %#v", m.allProjects)
	}
	persisted, err := config.LoadProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted[0].Templates) != 2 || persisted[0].PrimaryTemplate != rctarget.Client {
		t.Fatalf("persisted preferences = %#v", persisted[0])
	}
}

func TestCollapsingTemplatesKeepsFocusedTargetAndDropsHiddenSelection(t *testing.T) {
	m := New(nil).SetBounds(0, 0, 32, 12).SetActive(true)
	project := core.Project{
		Name:            "Alpha Project",
		ProjectID:       "alpha",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Client,
	}
	m, _ = m.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{project}, Source: "cache"})
	m.selected["alpha"] = struct{}{}
	m.selected["server@alpha"] = struct{}{}
	m.cursor = 1
	project.Templates = []rctarget.Kind{rctarget.Server}
	project.PrimaryTemplate = rctarget.Server
	m, cmd := m.Update(messages.ProjectTemplatePreferencesUpdatedMsg{
		Project: project,
	})
	if len(m.allProjects) != 1 || m.allProjects[0].ProjectID != "server@alpha" || m.cursor != 0 {
		t.Fatalf("collapsed targets = %#v, cursor %d", m.allProjects, m.cursor)
	}
	if _, ok := m.selected["alpha"]; ok {
		t.Fatal("hidden client target remained selected")
	}
	if cmd == nil {
		t.Fatal("collapse did not notify downstream panels after removing a selected target")
	}
}

func TestPhysicalActionTargetsDeduplicateTemplateRows(t *testing.T) {
	m := New(nil).SetBounds(0, 0, 32, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{{
		Name:            "Alpha Project",
		ProjectID:       "alpha",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Client,
	}}})
	m.selected["alpha"] = struct{}{}
	m.selected["server@alpha"] = struct{}{}
	got := m.PhysicalActionTargets()
	if len(got) != 1 || got[0].ProjectID != "alpha" {
		t.Fatalf("physical action targets = %#v", got)
	}
}

func TestAuthBindingRequiresTwoCommonIdentitiesAndEnabledTargets(t *testing.T) {
	m := loadedProjectsModel()
	m.projects[0].DiscoveredBy = []string{"main"}
	m.allProjects[0].DiscoveredBy = []string{"main"}
	if m.AuthBindingAvailable() {
		t.Fatal("single discovered identity enabled auth binding")
	}
	m.projects[0].DiscoveredBy = []string{"main", "work"}
	m.allProjects[0].DiscoveredBy = []string{"main", "work"}
	if !m.AuthBindingAvailable() {
		t.Fatal("two discovered identities did not enable auth binding")
	}
	m.projects[0].Disabled = true
	m.allProjects[0].Disabled = true
	if m.AuthBindingAvailable() {
		t.Fatal("disabled project enabled auth binding")
	}
}

func TestCurrentProjectIgnoresMarkedProjects(t *testing.T) {
	m := loadedProjectsModel()
	m.selected["alpha"] = struct{}{}
	m.selected["gamma"] = struct{}{}
	m.cursor = 1

	project, ok := m.CurrentProject()
	if !ok || project.ProjectID != "beta" {
		t.Fatalf("current project = %+v ok=%v, want beta", project, ok)
	}
}

func TestCurrentProjectScreenRowTracksCursor(t *testing.T) {
	m := loadedProjectsModel()
	row, ok := m.CurrentProjectScreenRow()
	if !ok || row != 1 {
		t.Fatalf("first project row = %d, %v; want 1, true", row, ok)
	}
	m, _ = m.Update(keyPress(tea.KeyDown))
	row, ok = m.CurrentProjectScreenRow()
	if !ok || row != 4 {
		t.Fatalf("second project row = %d, %v; want 4, true", row, ok)
	}
	if _, ok = m.SetCollapsed(true).CurrentProjectScreenRow(); ok {
		t.Fatal("collapsed Projects panel returned a current project screen row")
	}
}

func TestSelectOnlyReplacesSelectionAndMovesCursor(t *testing.T) {
	m := loadedProjectsModel()
	m.selected["alpha"] = struct{}{}
	m.selected["gamma"] = struct{}{}

	cmd := m.SelectOnly("beta")
	if cmd == nil {
		t.Fatal("selection command is nil")
	}
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want beta at 1", m.cursor)
	}
	if got := m.ActionTargets(); len(got) != 1 || got[0].ProjectID != "beta" {
		t.Fatalf("action targets = %+v, want only beta", got)
	}
	selection, ok := cmd().(messages.ProjectsSelectionChangedMsg)
	if !ok || len(selection.Projects) != 1 || selection.Projects[0].ProjectID != "beta" {
		t.Fatalf("selection message = %#v, want only beta", selection)
	}
}

func TestApplyProjectUpdatesRefreshesProjectAndSelectedPayload(t *testing.T) {
	m := loadedProjectsModel()
	m.selected["alpha"] = struct{}{}
	cmd := m.ApplyProjectUpdates([]core.Project{{Name: "Alpha Project", ProjectID: "alpha", AuthID: "work"}})
	if got := m.allProjects[0].AuthID; got != "work" {
		t.Fatalf("allProjects auth = %q, want work", got)
	}
	if got := m.projects[0].AuthID; got != "work" {
		t.Fatalf("visible projects auth = %q, want work", got)
	}
	if cmd == nil {
		t.Fatal("selected project update did not notify downstream panels")
	}
	selection := findMsg[messages.ProjectsSelectionChangedMsg](runBatch(t, cmd))
	if len(selection.Projects) != 1 || selection.Projects[0].AuthID != "work" {
		t.Fatalf("selection update = %+v, want rebound alpha", selection.Projects)
	}
}

func TestRemoveProjectsRemovesVisibleAndSelectedProjects(t *testing.T) {
	m := loadedProjectsModel()
	m.selected["alpha"] = struct{}{}
	m.selected["beta"] = struct{}{}
	cmd := m.RemoveProjects([]core.Project{{ProjectID: "alpha"}, {ProjectID: "gamma"}})
	if len(m.allProjects) != 1 || m.allProjects[0].ProjectID != "beta" {
		t.Fatalf("allProjects = %+v, want beta", m.allProjects)
	}
	if len(m.projects) != 1 || m.projects[0].ProjectID != "beta" {
		t.Fatalf("visible projects = %+v, want beta", m.projects)
	}
	if cmd == nil {
		t.Fatal("selected project removal did not notify downstream panels")
	}
	selection := findMsg[messages.ProjectsSelectionChangedMsg](runBatch(t, cmd))
	if len(selection.Projects) != 1 || selection.Projects[0].ProjectID != "beta" {
		t.Fatalf("selection = %+v, want beta", selection.Projects)
	}
}

func loadedProjectsModel() Model {
	m := New(nil).SetBounds(0, 0, 32, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{
		Projects: []core.Project{
			{Name: "Alpha Project", ProjectID: "alpha"},
			{Name: "Beta Project", ProjectID: "beta"},
			{Name: "Gamma Project", ProjectID: "gamma"},
		},
		Source: "cache",
	})
	return m
}

func disabledProjectsModel() Model {
	m := New(nil).SetBounds(0, 0, 32, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{
		Projects: []core.Project{
			{Name: "Alpha Project", ProjectID: "alpha", Disabled: true},
			{Name: "Beta Project", ProjectID: "beta"},
		},
		Source: "cache",
	})
	return m
}

func keyPress(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func keyText(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text})
}

func runBatch(t *testing.T, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		out := make([]tea.Msg, 0, len(batch))
		for _, item := range batch {
			if item == nil {
				continue
			}
			out = append(out, item())
		}
		return out
	}
	return []tea.Msg{msg}
}

func findMsg[T any](msgs []tea.Msg) T {
	for _, msg := range msgs {
		if typed, ok := msg.(T); ok {
			return typed
		}
	}
	var zero T
	return zero
}

const loadedProjectsSnapshot = `── ¹Projects ────────────── 3 ─╮
 Alpha Project                 │
  alpha                        │
                               │
 Beta Project                  │
  beta                         │
                               │
 Gamma Project                 │
  gamma                        │
                               │
                               │
───────────────────────────────╯`
