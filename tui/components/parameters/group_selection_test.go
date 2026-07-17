package parameters

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

func TestGroupHeaderSelectionOpensMetadataDetails(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m := New(nil)
	m, _ = m.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m, _ = m.Update(messages.ParametersLoadedMsg{
		Project: project,
		Tree:    &core.ParametersTree{Groups: []core.ParametersGroup{{Key: "empty", Label: "empty", Description: "Metadata only"}}},
	})
	if !m.FocusGroup("demo", "empty") {
		t.Fatal("failed to focus group header")
	}
	msg := m.selectionChangedCmd(true)()
	selection, ok := msg.(messages.ParameterSelectionChangedMsg)
	if !ok || selection.GroupData == nil {
		t.Fatalf("selection = %#v, want group metadata", msg)
	}
	if selection.GroupData.Group.Description != "Metadata only" || !selection.Activate {
		t.Fatalf("group selection = %#v", selection.GroupData)
	}
}
