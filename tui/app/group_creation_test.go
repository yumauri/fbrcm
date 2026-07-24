package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestShiftAOpensEmptyGroupDetailsWithNameFocused(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m := viewTestModel(100, 32, panels.Parameters)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project: project,
		Tree: &core.ParametersTree{Groups: []core.ParametersGroup{{
			Key: "existing", Label: "existing",
		}}},
		Source: "cache",
	})

	var paletteAction *helpPaletteAction
	for _, action := range m.helpPaletteActions() {
		if action.block == tuiconfig.BlockParameters && action.action == tuiconfig.ActionNewGroup {
			item := action
			paletteAction = &item
			break
		}
	}
	if paletteAction == nil || !paletteAction.enabled || len(paletteAction.keys) != 1 || paletteAction.keys[0] != "A" {
		t.Fatalf("new group palette action = %#v", paletteAction)
	}

	m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'A', Text: "A"}))

	data := m.details.GroupData()
	if !handled || m.active != panels.Details || !m.detailsVisible || data == nil {
		t.Fatalf("Shift+A = handled:%v active:%v visible:%v data:%#v", handled, m.active, m.detailsVisible, data)
	}
	if data.Group.Key != "" || len(data.Group.Parameters) != 0 {
		t.Fatalf("new group = %#v, want empty group", data.Group)
	}
	if len(data.GroupNames) != 1 || data.GroupNames[0] != "existing" {
		t.Fatalf("group names = %v, want [existing]", data.GroupNames)
	}
	if !m.details.FieldActive() || !m.details.TextInputActive() {
		t.Fatal("new group name field is not focused")
	}

	next, _ := m.Update(struct{}{})
	m = next.(Model)
	if !m.detailsVisible || !m.details.IsNewGroup() {
		t.Fatal("background update closed new group Details")
	}

	next, _ = m.Update(messages.ParameterSelectionChangedMsg{
		GroupData: &messages.GroupViewData{
			Project: project,
			Group:   core.ParametersGroup{Key: "existing", Label: "existing"},
		},
	})
	m = next.(Model)
	if !m.detailsVisible || !m.details.IsNewGroup() {
		t.Fatal("background parameter selection replaced new group Details")
	}

	next, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: 'n', Text: "n"}))
	m = next.(Model)
	edit, changed := m.details.GroupEdit()
	if !m.detailsVisible || !m.details.IsNewGroup() || !changed || !edit.Create || edit.NextName != "n" {
		t.Fatalf("typed name = visible:%v new:%v changed:%v edit:%#v", m.detailsVisible, m.details.IsNewGroup(), changed, edit)
	}
}
