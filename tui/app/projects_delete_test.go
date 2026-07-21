package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestProjectsDeleteKeyOpensLocalOnlyConfirmationForDisabledProject(t *testing.T) {
	m := New(newRenameTestService(t))
	m.width, m.height = 100, 30
	m.projects, _ = m.projects.Update(messages.ProjectsLoadedMsg{Projects: []core.Project{{
		Name: "Disabled", ProjectID: "disabled", AuthID: "main", Disabled: true,
	}}})
	m.setActive(panels.Projects)
	m.applyLayout()

	next, cmd, handled := m.updateGlobalPanelActionKey("x")
	if !handled || cmd != nil || !next.dialog.IsOpen() {
		t.Fatalf("delete handled=%v cmd=%v dialog=%v", handled, cmd != nil, next.dialog.IsOpen())
	}
	view := ansi.Strip(next.dialog.View())
	for _, want := range []string{"Delete Project?", "Disabled (disabled)", "Firebase is not changed.", "Delete", "Cancel"} {
		if !strings.Contains(view, want) {
			t.Fatalf("delete dialog missing %q:\n%s", want, view)
		}
	}
}
