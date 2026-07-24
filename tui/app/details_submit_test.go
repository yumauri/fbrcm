package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestEnterSavesParameterAndGroupDetails(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *Model)
		input   string
	}{
		{name: "new group", prepare: prepareNewGroupDetails, input: "created_group"},
		{name: "new parameter", prepare: prepareNewParameterDetails, input: "created_parameter"},
		{name: "existing group", prepare: prepareExistingGroupDetails, input: "_edited"},
		{name: "existing parameter", prepare: prepareExistingParameterDetails, input: "_edited"},
	}

	for _, test := range tests {
		for _, hasDraft := range []bool{false, true} {
			mode := "publish confirmation"
			if hasDraft {
				mode = "existing draft"
			}
			t.Run(test.name+"/"+mode, func(t *testing.T) {
				m := newRenameTestModel(t, hasDraft)
				test.prepare(t, &m)
				m.details, _ = m.details.Update(tea.PasteMsg{Content: test.input})
				if !m.details.Dirty() || !m.details.FieldActive() {
					t.Fatalf("prepared Details = dirty:%v field:%v", m.details.Dirty(), m.details.FieldActive())
				}

				var cmd tea.Cmd
				var handled bool
				m, cmd, handled = m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
				if !handled || cmd != nil || m.dialog.IsOpen() || m.details.FieldActive() {
					t.Fatalf("first Enter = handled:%v cmd:%v dialog:%v field:%v; want field close only",
						handled, cmd != nil, m.dialog.IsOpen(), m.details.FieldActive())
				}

				m, cmd, handled = m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
				if !handled {
					t.Fatal("save Enter was not handled")
				}
				if !hasDraft {
					if cmd != nil || !m.dialog.IsOpen() {
						t.Fatalf("save Enter = cmd:%v dialog:%v; want confirmation dialog", cmd != nil, m.dialog.IsOpen())
					}
					return
				}
				if cmd == nil || m.dialog.IsOpen() {
					t.Fatalf("save Enter = cmd:%v dialog:%v; want immediate draft command", cmd != nil, m.dialog.IsOpen())
				}
				msg := runRenameCmd(t, cmd)
				if msg.Err != nil || msg.Source != "draft" || !msg.HasDraft || !msg.DetailsSaved {
					t.Fatalf("draft save = source:%q draft:%v saved:%v err:%v", msg.Source, msg.HasDraft, msg.DetailsSaved, msg.Err)
				}
			})
		}
	}
}

func prepareNewGroupDetails(t *testing.T, m *Model) {
	t.Helper()
	if cmd := m.openNewGroupDetails(); cmd == nil {
		t.Fatal("new group name field did not return a focus command")
	}
}

func prepareNewParameterDetails(t *testing.T, m *Model) {
	t.Helper()
	if !m.parameters.FocusParameter("demo", "__default__", "flag") {
		t.Fatal("failed to focus source parameter")
	}
	if cmd := m.openNewParameterDetails(); cmd == nil {
		t.Fatal("new parameter name field did not return a focus command")
	}
}

func prepareExistingGroupDetails(t *testing.T, m *Model) {
	t.Helper()
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m.details = m.details.SetGroupData(&messages.GroupViewData{
		Project: project,
		Group: core.ParametersGroup{
			Key: "group", Label: "group",
			Parameters: []core.ParametersEntry{{Key: "group_flag"}},
		},
		GroupNames: []string{"group"},
	})
	activateDetailsName(m)
}

func prepareExistingParameterDetails(t *testing.T, m *Model) {
	t.Helper()
	data, ok := m.parameters.ParameterViewData("demo", "__default__", "flag", "")
	if !ok {
		t.Fatal("existing parameter Details data is unavailable")
	}
	m.details = m.details.SetData(data)
	activateDetailsName(m)
}

func activateDetailsName(m *Model) {
	m.detailsVisible = true
	m.setActive(panels.Details)
	m.details, _ = m.details.ActivateName()
}
