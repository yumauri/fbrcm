package app

import (
	"testing"

	"github.com/yumauri/fbrcm/tui/panels"
)

func TestResetWorkspaceForProfileClearsProfileScopedState(t *testing.T) {
	m := viewTestModel(90, 24, panels.Parameters)
	m.detailsVisible = true
	m.parametersTab = panels.History
	m.projectsMode = projectsPanelModeCollapsed

	cmd := m.resetWorkspaceForProfile()

	if cmd == nil {
		t.Fatal("reset init command is nil")
	}
	if m.active != panels.Projects || m.parametersTab != panels.Parameters {
		t.Fatalf("active=%v parametersTab=%v, want Projects/Parameters", m.active, m.parametersTab)
	}
	if m.detailsVisible || m.projectsMode != projectsPanelModeExpanded {
		t.Fatalf("detailsVisible=%v projectsMode=%v, want reset", m.detailsVisible, m.projectsMode)
	}
	if m.width != 90 || m.height != 24 {
		t.Fatalf("size=%dx%d, want preserved 90x24", m.width, m.height)
	}
}
