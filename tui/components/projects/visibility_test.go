package projects

import (
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/tui/messages"
)

func TestProjectVisibilityToggleAndMouseSelection(t *testing.T) {
	t.Setenv(env.ConfigDir, t.TempDir())
	t.Setenv(env.CacheDir, t.TempDir())
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	policy := config.DefaultProjectVisibility()
	policy.Blacklist.ProjectIDs = []string{"alpha"}
	if err := config.SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	m := New(nil).SetBounds(4, 3, 70, 18).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{
		{Name: "Alpha", ProjectID: "alpha", Templates: []rctarget.Kind{rctarget.Client, rctarget.Server}},
		{Name: "Beta", ProjectID: "beta"},
	}})
	if len(m.allProjects) != 1 || m.allProjects[0].ProjectID != "beta" {
		t.Fatalf("visible targets = %#v", m.allProjects)
	}
	m, _ = m.Update(keyPress('H'))
	if len(m.allProjects) != 3 || !strings.Contains(strings.Join(m.lines, "\n"), "blacklisted") {
		t.Fatalf("show hidden = %#v, lines %q", m.allProjects, m.lines)
	}
	// Hit regions follow the rendered, offset panel after template expansion.
	click := tea.MouseClickMsg{X: m.x + 3, Y: m.y + 1 + m.projectStarts[1] - m.viewport.YOffset(), Button: tea.MouseLeft}
	m, _ = m.Update(click)
	if m.cursor != 1 {
		t.Fatalf("click cursor = %d, want server template", m.cursor)
	}
	m, cmd := m.Update(click)
	selection := findMsg[messages.ProjectsSelectionChangedMsg](runBatch(t, cmd))
	if len(selection.Projects) != 1 || selection.Projects[0].ProjectID != "server@alpha" {
		t.Fatalf("double click = %#v", selection)
	}
	m, cmd = m.Update(keyPress('H'))
	selection = findMsg[messages.ProjectsSelectionChangedMsg](runBatch(t, cmd))
	if len(selection.Projects) != 0 || len(m.selected) != 0 || len(m.allProjects) != 1 {
		t.Fatalf("hiding must drop selected hidden projects: %#v, %#v", selection, m.selected)
	}
	stored, err := config.LoadProjectVisibility()
	if err != nil || !reflect.DeepEqual(stored, policy) {
		t.Fatalf("session toggle changed persistent policy: %#v, %v", stored, err)
	}
}

func TestEnabledEmptyWhitelistShowsRecoverableEmptyWorkspace(t *testing.T) {
	t.Setenv(env.ConfigDir, t.TempDir())
	t.Setenv(env.CacheDir, t.TempDir())
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	policy := config.DefaultProjectVisibility()
	policy.Whitelist.Enabled = true
	if err := config.SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	m := New(nil).SetBounds(0, 0, 50, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{{Name: "Alpha", ProjectID: "alpha", Disabled: true}}})
	if len(m.projects) != 0 || !strings.Contains(strings.Join(m.lines, "\n"), "Hidden by rules") {
		t.Fatalf("empty workspace = %q", m.lines)
	}
	m, _ = m.Update(keyPress('H'))
	if len(m.projects) != 1 || m.CurrentProjectEnabled() || !strings.Contains(strings.Join(m.lines, "\n"), "not whitelisted") {
		t.Fatalf("disabled project must remain disabled after showing: %#v", m.projects)
	}
}
