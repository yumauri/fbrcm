package app

import (
	"encoding/json"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestNewConditionUsesNameThenRawExpressionAndStagesIntoDraft(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	svc := newRenameTestService(t)
	raw := conditionTestRemoteConfigRaw()
	saveRenameParametersCache(t, project.ProjectID, raw)
	if err := svc.SaveDraft(project.ProjectID, raw); err != nil {
		t.Fatal(err)
	}

	m := New(svc)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{Project: project, Tree: &core.ParametersTree{}, Source: "draft", HasDraft: true})
	m.conditions, _ = m.conditions.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.conditions, _ = m.conditions.Update(messages.ConditionsLoadedMsg{Project: project, Tree: &core.ConditionsTree{}, Source: "draft"})
	m.setActive(panels.Conditions)

	if cmd := m.openNewConditionInput(); cmd != nil {
		_ = cmd()
	}
	setRenameInputValue(t, &m, "beta_users")
	var handled bool
	m, _, handled = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !handled {
		t.Fatal("Enter was not handled by the condition name input")
	}
	if !m.stringInput.IsOpen() || !m.stringInput.IsExpanded() || m.conditionEdit == nil || !m.conditionEdit.creating {
		t.Fatalf("raw expression editor state = open:%v expanded:%v session:%#v", m.stringInput.IsOpen(), m.stringInput.IsExpanded(), m.conditionEdit)
	}

	m.stringInput = m.stringInput.Close()
	m.stringInput, _ = m.stringInput.Open(0, 0, 20, 100, 100, 30, "percent <= 10", true, true)
	m, cmd, handled := m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter, Mod: tea.ModCtrl}))
	if !handled {
		t.Fatal("Ctrl+Enter was not handled by the condition expression editor")
	}
	if m.stringInput.IsOpen() || m.conditionEdit != nil {
		t.Fatalf("saved expression editor state = open:%v session:%#v", m.stringInput.IsOpen(), m.conditionEdit)
	}
	msg := runRenameCmd(t, cmd)
	if msg.Err != nil || !msg.HasDraft || msg.SelectConditionName != "beta_users" {
		t.Fatalf("condition mutation message = %#v", msg)
	}

	draftRaw, ok, err := svc.LoadDraft(project.ProjectID)
	if err != nil || !ok {
		t.Fatalf("LoadDraft = %v, %v", ok, err)
	}
	cfg, err := firebase.ParseRemoteConfig(draftRaw)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Conditions) != 1 || cfg.Conditions[0].Name != "beta_users" || cfg.Conditions[0].Expression != "percent <= 10" {
		t.Fatalf("draft conditions = %#v", cfg.Conditions)
	}
}

func conditionTestRemoteConfigRaw() json.RawMessage {
	raw, err := json.Marshal(firebase.RemoteConfig{Version: firebase.RemoteConfigVersion{VersionNumber: "1"}})
	if err != nil {
		panic(err)
	}
	return raw
}
