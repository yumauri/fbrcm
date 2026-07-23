package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
	promotecmp "github.com/yumauri/fbrcm/tui/components/promote"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestPromoteOpensFromCursorAndRestoresWorkspaceLayout(t *testing.T) {
	m := viewTestModel(100, 30, panels.Projects)
	projects := []core.Project{{Name: "Development", ProjectID: "dev"}, {Name: "Production", ProjectID: "prod"}}
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: projects})
	m.setActive(panels.Projects)
	originalLogsHeight := m.logsHeight
	originalProjectsView := m.projects.ViewWithBorder(true, true)

	next, cmd, handled := m.updateGlobalPanelActionKey("v")
	if !handled || cmd != nil || !next.promote.IsOpen() || next.promote.Source().ProjectID != "dev" {
		t.Fatalf("open promote = handled:%v cmd:%v open:%v source:%q", handled, cmd != nil, next.promote.IsOpen(), next.promote.Source().ProjectID)
	}
	if next.active != panels.Projects || next.projectsMode != projectsPanelModeExpanded || next.logsMode != logsPanelModeExpanded || !next.promote.TargetPickerOpen() {
		t.Fatalf("target picker layout = active:%v projects:%v logs:%v target:%v", next.active, next.projectsMode, next.logsMode, next.promote.TargetPickerOpen())
	}
	if projectsView := next.projects.ViewWithBorder(true, true); projectsView != originalProjectsView {
		t.Fatal("opening the promote picker changed Projects panel rendering")
	}
	view := next.View().Content
	normalized := testutil.NormalizeViewSnapshot(view)
	if !strings.Contains(normalized, "Promote to…") || !strings.Contains(normalized, "Filter: Type to filter projects") || !strings.Contains(normalized, "▸ Production (prod)") {
		t.Fatalf("connected target picker overlay missing from app view:\n%s", normalized)
	}

	next, cmd, handled = next.updatePromoteMessage(promotecmp.TargetSelectedMsg{Source: projects[0], Target: projects[1], Mode: core.ProjectPromotionEffective})
	if !handled || cmd == nil || next.active != panels.Promote || next.projectsMode != projectsPanelModeExpanded || next.logsMode != logsPanelModeExpanded || !next.promote.WorkspaceOpen() {
		t.Fatalf("promotion layout = handled:%v cmd:%v active:%v projects:%v logs:%v workspace:%v", handled, cmd != nil, next.active, next.projectsMode, next.logsMode, next.promote.WorkspaceOpen())
	}

	next = next.closePromote()
	if next.promote.IsOpen() || next.active != panels.Projects || next.projectsMode != projectsPanelModeExpanded || next.logsMode != logsPanelModeExpanded || next.logsHeight != originalLogsHeight {
		t.Fatalf("restored layout = open:%v active:%v projects:%v logs:%v height:%d", next.promote.IsOpen(), next.active, next.projectsMode, next.logsMode, next.logsHeight)
	}
}

func TestPromoteRequiresTwoProjects(t *testing.T) {
	m := viewTestModel(100, 30, panels.Projects)
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{{Name: "Only", ProjectID: "only"}}})
	m.setActive(panels.Projects)
	next, _, handled := m.updateGlobalPanelActionKey("v")
	if !handled || next.promote.IsOpen() || !next.dialog.IsOpen() {
		t.Fatalf("single-project promote = handled:%v open:%v dialog:%v", handled, next.promote.IsOpen(), next.dialog.IsOpen())
	}
}

func TestPromoteTargetPickerCanRenderAboveScreen(t *testing.T) {
	m := viewTestModel(100, 30, panels.Projects)
	projects := []core.Project{
		{Name: "Source", ProjectID: "source"},
		{Name: "First", ProjectID: "first"},
		{Name: "Second", ProjectID: "second"},
		{Name: "Third", ProjectID: "third"},
		{Name: "Fourth", ProjectID: "fourth"},
	}
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: projects})
	m.setActive(panels.Projects)
	baselineHeight := len(strings.Split(m.View().Content, "\n"))

	next, _, handled := m.updateGlobalPanelActionKey("v")
	if !handled {
		t.Fatal("promote shortcut was not handled")
	}
	next.promote, _ = next.promote.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnd}))
	if _, y := next.promote.TargetPosition(); y >= 0 {
		t.Fatalf("target picker Y = %d, want a negative position", y)
	}
	view := testutil.NormalizeViewSnapshot(next.View().Content)
	if height := len(strings.Split(view, "\n")); height != baselineHeight {
		t.Fatalf("view height = %d after off-screen picker, want %d", height, baselineHeight)
	}
	aligned := false
	for line := range strings.SplitSeq(view, "\n") {
		if strings.Contains(line, "Source") && strings.Contains(line, "▸ Fourth (fourth)") {
			aligned = true
			break
		}
	}
	if !aligned {
		t.Fatalf("source and selected target are not aligned:\n%s", view)
	}
}

func TestPromoteWorkspaceUsesRegularPanelNavigation(t *testing.T) {
	m := promoteWorkspaceTestModel(t)

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: '1', Text: "1"}))
	if m.active != panels.Projects {
		t.Fatalf("active panel after 1 = %v, want Projects", m.active)
	}
	if !m.promote.WorkspaceOpen() {
		t.Fatal("focusing Projects closed Promote")
	}

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if m.active != panels.Promote {
		t.Fatalf("active panel after Tab = %v, want Promote", m.active)
	}

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: '0', Text: "0"}))
	if m.active != panels.Logs {
		t.Fatalf("active panel after 0 = %v, want Logs", m.active)
	}

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if m.active != panels.Promote {
		t.Fatalf("active panel after Logs Tab = %v, want Promote", m.active)
	}

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: '1', Text: "1"}))
	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: '9', Text: "9"}))
	if m.active != panels.Promote {
		t.Fatalf("active panel after 9 = %v, want Promote", m.active)
	}

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: '2', Text: "2"}))
	if m.active != panels.Promote {
		t.Fatalf("active panel after hidden workspace shortcut = %v, want Promote", m.active)
	}
}

func TestPromoteWorkspaceProjectSelectionReturnsToParameters(t *testing.T) {
	tests := []struct {
		name string
		key  tea.Key
	}{
		{name: "enter", key: tea.Key{Code: tea.KeyEnter}},
		{name: "space", key: tea.Key{Code: tea.KeySpace}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := promoteWorkspaceTestModel(t)
			m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: '1', Text: "1"}))

			m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tt.key))
			if m.promote.IsOpen() {
				t.Fatal("project selection left the Promote workspace open")
			}
			if m.active != panels.Parameters || m.parametersTab != panels.Parameters {
				t.Fatalf("active workspace after project selection = %v/%v, want Parameters", m.active, m.parametersTab)
			}
		})
	}
}

func TestStartingAnotherPromotionPreservesCurrentWorkspaceUntilTargetConfirmation(t *testing.T) {
	m := promoteWorkspaceTestModel(t)
	originalSource := m.promote.Source()
	originalTarget := m.promote.Target()
	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: '1', Text: "1"}))
	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: 'v', Text: "v"}))
	if !m.promote.TargetPickerOpen() || !m.promote.WorkspaceOpen() {
		t.Fatalf("replacement picker = picker:%v workspace:%v, want both open", m.promote.TargetPickerOpen(), m.promote.WorkspaceOpen())
	}
	if m.promote.Source().ProjectID != originalSource.ProjectID || m.promote.Target().ProjectID != originalTarget.ProjectID {
		t.Fatalf("replacement picker changed current pair to %v -> %v", m.promote.Source(), m.promote.Target())
	}

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if m.promote.TargetPickerOpen() || !m.promote.WorkspaceOpen() {
		t.Fatalf("canceled replacement = picker:%v workspace:%v, want picker closed over existing workspace", m.promote.TargetPickerOpen(), m.promote.WorkspaceOpen())
	}
	if m.promote.Source().ProjectID != originalSource.ProjectID || m.promote.Target().ProjectID != originalTarget.ProjectID {
		t.Fatalf("canceling replacement changed current pair to %v -> %v", m.promote.Source(), m.promote.Target())
	}

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: 'v', Text: "v"}))
	next, targetCmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = next.(Model)
	if targetCmd == nil {
		t.Fatal("confirming replacement target returned no command")
	}
	next, prepareCmd := m.Update(targetCmd())
	m = next.(Model)
	if prepareCmd == nil {
		t.Fatal("confirmed replacement did not begin loading")
	}
	if m.promote.TargetPickerOpen() || !m.promote.WorkspaceOpen() || m.active != panels.Promote {
		t.Fatalf("confirmed replacement = picker:%v workspace:%v active:%v", m.promote.TargetPickerOpen(), m.promote.WorkspaceOpen(), m.active)
	}
	if m.promote.Source().ProjectID != "prod" || m.promote.Target().ProjectID != "dev" {
		t.Fatalf("confirmed replacement pair = %s -> %s, want prod -> dev", m.promote.Source().ProjectID, m.promote.Target().ProjectID)
	}
}

func TestPromoteWorkspaceSupportsMaximizeAndQuit(t *testing.T) {
	m := promoteWorkspaceTestModel(t)
	if m.projectsMode != projectsPanelModeExpanded || m.logsMode != logsPanelModeExpanded {
		t.Fatal("promotion changed the existing panel layout")
	}

	m = updatePromoteWorkspace(t, m, tea.KeyPressMsg(tea.Key{Code: 'z', Text: "z"}))
	if m.projectsMode != projectsPanelModeCollapsed || m.logsMode != logsPanelModeCollapsed {
		t.Fatalf("layout after z = projects:%v logs:%v, want collapsed", m.projectsMode, m.logsMode)
	}

	_, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	if cmd == nil {
		t.Fatal("q did not request application quit")
	}
}

func TestPromoteEnterOpensGenericDiffModal(t *testing.T) {
	m := reviewedPromoteWorkspaceTestModel(t)
	next, command := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	m = next.(Model)
	if command == nil {
		t.Fatal("Enter on a promotion change returned no diff request")
	}
	next, _ = m.Update(command())
	m = next.(Model)
	if !m.diffView.IsOpen() {
		t.Fatal("promotion diff modal did not open")
	}
	view := testutil.NormalizeViewSnapshot(m.View().Content)
	for _, text := range []string{"Diff", "flag", "Development (dev)", "Production (prod)", "value · default"} {
		if !strings.Contains(view, text) {
			t.Fatalf("promotion diff modal misses %q:\n%s", text, view)
		}
	}
	if strings.Contains(view, "Diff · flag") {
		t.Fatalf("promotion diff modal still combines its title and entity name:\n%s", view)
	}

	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	m = next.(Model)
	if m.diffView.IsOpen() || !m.promote.WorkspaceOpen() {
		t.Fatalf("closing diff = diff:%v promote:%v", m.diffView.IsOpen(), m.promote.WorkspaceOpen())
	}
}

func TestPromoteWorkspaceOwnsRightPanelHitArea(t *testing.T) {
	m := promoteWorkspaceTestModel(t)
	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	panel, ok := m.panelAt(layout.leftWidth, 1)
	if !ok || panel != panels.Promote {
		t.Fatalf("right panel hit = (%v, %v), want Promote", panel, ok)
	}
}

func reviewedPromoteWorkspaceTestModel(t *testing.T) Model {
	t.Helper()
	m := promoteWorkspaceTestModel(t)
	source := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"flag": {ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "true"}},
	}}
	target := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"flag": {ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "false"}},
	}}
	sourceRaw, err := firebase.MarshalRemoteConfig(source)
	if err != nil {
		t.Fatal(err)
	}
	targetRaw, err := firebase.MarshalRemoteConfig(target)
	if err != nil {
		t.Fatal(err)
	}
	plan := &core.ProjectPromotionPlan{
		Source: core.ProjectPromotionSnapshot{
			Project:      core.Project{Name: "Development", ProjectID: "dev"},
			Raw:          sourceRaw,
			PublishedRaw: sourceRaw,
		},
		Target: core.ProjectPromotionSnapshot{
			Project:      core.Project{Name: "Production", ProjectID: "prod"},
			Raw:          targetRaw,
			PublishedRaw: targetRaw,
		},
		Plan: rcpromote.BuildPlan(source, target, rcpromote.Options{Prune: true}),
	}
	m.promote = m.promote.SetPlan(plan, true)
	m.setActive(panels.Promote)
	return m
}

func promoteWorkspaceTestModel(t *testing.T) Model {
	t.Helper()
	m := viewTestModel(100, 30, panels.Projects)
	projects := []core.Project{{Name: "Development", ProjectID: "dev"}, {Name: "Production", ProjectID: "prod"}}
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: projects})
	m.setActive(panels.Projects)

	next, _, handled := m.updateGlobalPanelActionKey("v")
	if !handled {
		t.Fatal("promote shortcut was not handled")
	}
	next, _, handled = next.updatePromoteMessage(promotecmp.TargetSelectedMsg{
		Source: projects[0], Target: projects[1], Mode: core.ProjectPromotionEffective,
	})
	if !handled || !next.promote.WorkspaceOpen() {
		t.Fatal("promotion workspace did not open")
	}
	return next
}

func updatePromoteWorkspace(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	next, _ := m.Update(msg)
	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("updated model type = %T, want app.Model", next)
	}
	return updated
}
