package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/tui/components/setup"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestStartupActivatesOnlyVisibleProject(t *testing.T) {
	alpha := core.Project{Name: "Alpha", ProjectID: "alpha"}
	beta := core.Project{Name: "Beta", ProjectID: "beta"}
	for _, test := range []struct {
		name      string
		projects  []core.Project
		whitelist bool
		blacklist bool
		sizeLater bool
		wantID    string
	}{
		{name: "only project", projects: []core.Project{alpha}, wantID: "alpha"},
		{name: "startup before terminal size", projects: []core.Project{alpha}, sizeLater: true, wantID: "alpha"},
		{name: "only whitelisted project", projects: []core.Project{alpha, beta}, whitelist: true, wantID: "alpha"},
		{name: "only non-blacklisted project", projects: []core.Project{alpha, beta}, blacklist: true, wantID: "alpha"},
		{name: "server target", projects: []core.Project{{Name: "Alpha", ProjectID: "alpha", Templates: []rctarget.Kind{rctarget.Server}}}, wantID: "server@alpha"},
		{name: "several projects", projects: []core.Project{alpha, beta}},
		{name: "empty workspace"},
		{name: "all hidden", projects: []core.Project{beta}, whitelist: true},
		{name: "disabled project", projects: []core.Project{{Name: "Alpha", ProjectID: "alpha", Disabled: true}}},
		{name: "two template rows", projects: []core.Project{{Name: "Alpha", ProjectID: "alpha", Templates: []rctarget.Kind{rctarget.Client, rctarget.Server}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(env.ConfigDir, t.TempDir())
			t.Setenv(env.CacheDir, t.TempDir())
			if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
				t.Fatal(err)
			}
			policy := config.DefaultProjectVisibility()
			policy.Whitelist.Enabled = test.whitelist
			policy.Whitelist.ProjectIDs = []string{"alpha"}
			if test.blacklist {
				policy.Blacklist.ProjectIDs = []string{"beta"}
			}
			if err := config.SaveProjectVisibility(policy); err != nil {
				t.Fatal(err)
			}
			m := viewTestModel(90, 24, panels.Projects)
			if test.sizeLater {
				m = New(nil)
			}
			next, cmd := m.Update(setup.WorkspaceReadyMsg{Projects: test.projects, Source: "cache"})
			m = next.(Model)
			if test.sizeLater {
				next, _ = m.Update(tea.WindowSizeMsg{Width: 90, Height: 24})
				m = next.(Model)
			}
			if !m.workspaceInitialized {
				t.Fatal("workspace was not initialized")
			}
			if test.wantID == "" {
				if m.active != panels.Projects || m.projectsMode != projectsPanelModeExpanded || cmd != nil {
					t.Fatalf("unexpected activation: active=%v mode=%v cmd=%v", m.active, m.projectsMode, cmd != nil)
				}
				return
			}
			if m.active != panels.Parameters || m.parametersTab != panels.Parameters || m.projectsMode != projectsPanelModeCollapsed {
				t.Fatalf("startup active=%v tab=%v mode=%v", m.active, m.parametersTab, m.projectsMode)
			}
			selection, ok := cmd().(messages.ProjectsSelectionChangedMsg)
			if !ok || len(selection.Projects) != 1 || selection.Projects[0].ProjectID != test.wantID {
				t.Fatalf("startup selection = %#v", selection)
			}
			next, loadCmd := m.Update(selection)
			m = next.(Model)
			if !m.parameters.HasProject(test.wantID) || loadCmd == nil {
				t.Fatal("automatic selection did not start loading the Parameters workspace")
			}
			next, _ = m.Update(tea.MouseClickMsg{X: collapsedProjectsWidth + 2, Y: 2, Button: tea.MouseLeft})
			m = next.(Model)
			if m.active != panels.Parameters {
				t.Fatal("mouse regions did not follow the collapsed startup layout")
			}
			// Subsequent discovery/account updates must preserve user focus/layout.
			m.setActive(panels.Logs)
			m.setProjectsMode(projectsPanelModeExpanded)
			next, _ = m.Update(setup.WorkspaceReadyMsg{Projects: test.projects, Source: "firebase"})
			m = next.(Model)
			if m.active != panels.Logs || m.projectsMode != projectsPanelModeExpanded {
				t.Fatal("later workspace reload stole focus or collapsed Projects")
			}
		})
	}
}

func TestProfileSwitchInitializesSingleProjectWorkspaceAgain(t *testing.T) {
	t.Setenv(env.ConfigDir, t.TempDir())
	t.Setenv(env.CacheDir, t.TempDir())
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	m := viewTestModel(90, 24, panels.Logs)
	m.workspaceInitialized = true
	next, _ := m.Update(setup.WorkspaceReadyMsg{Projects: []core.Project{{ProjectID: "alpha"}}, Reset: true})
	m = next.(Model)
	if m.active != panels.Parameters || m.projectsMode != projectsPanelModeCollapsed || !m.workspaceInitialized {
		t.Fatalf("profile workspace active=%v mode=%v initialized=%t", m.active, m.projectsMode, m.workspaceInitialized)
	}
}
