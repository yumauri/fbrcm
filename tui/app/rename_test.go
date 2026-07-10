package app

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestSubmitRenameInputUnchangedClosesInput(t *testing.T) {
	m := newRenameTestModel(t, false)
	openRenameForParameter(t, &m)

	cmd := m.submitRenameInput()

	if cmd != nil {
		t.Fatalf("submitRenameInput returned command for unchanged rename")
	}
	if m.renameInput.IsOpen() {
		t.Fatalf("rename input is open, want closed")
	}
	if m.dialog.IsOpen() {
		t.Fatalf("dialog is open for unchanged rename")
	}
}

func TestSubmitRenameInputEmptyNameKeepsInputAndShowsError(t *testing.T) {
	m := newRenameTestModel(t, false)
	openRenameForParameter(t, &m)
	setRenameInputValue(t, &m, "   ")

	cmd := m.submitRenameInput()

	if cmd != nil {
		t.Fatalf("submitRenameInput returned command for invalid rename")
	}
	if !m.renameInput.IsOpen() {
		t.Fatalf("rename input closed after invalid rename")
	}
	if !m.dialog.IsOpen() {
		t.Fatalf("error dialog is not open")
	}
}

func TestSubmitParameterRenameWithoutDraftOpensDialog(t *testing.T) {
	m := newRenameTestModel(t, false)
	openRenameForParameter(t, &m)
	setRenameInputValue(t, &m, "renamed")

	cmd := m.submitRenameInput()

	if cmd != nil {
		t.Fatalf("submitRenameInput returned command for non-draft rename")
	}
	if m.renameInput.IsOpen() {
		t.Fatalf("rename input is open, want closed")
	}
	if !m.dialog.IsOpen() {
		t.Fatalf("confirmation dialog is not open")
	}
}

func TestSubmitParameterRenameWithDraftReturnsDraftCommand(t *testing.T) {
	m := newRenameTestModel(t, true)
	openRenameForParameter(t, &m)
	setRenameInputValue(t, &m, "renamed")

	cmd := m.submitRenameInput()

	if cmd == nil {
		t.Fatalf("submitRenameInput returned nil command for draft rename")
	}
	if m.renameInput.IsOpen() {
		t.Fatalf("rename input is open, want closed")
	}
	msg := runRenameCmd(t, cmd)
	if msg.Source != "draft" {
		t.Fatalf("Source = %q, want draft", msg.Source)
	}
	if !msg.HasDraft {
		t.Fatalf("HasDraft = false, want true")
	}
	if msg.Err != nil {
		t.Fatalf("Err = %v, want nil", msg.Err)
	}
}

func TestSubmitGroupRenameWithDraftReturnsDraftCommand(t *testing.T) {
	m := newRenameTestModel(t, true)
	openRenameForGroup(t, &m, "group")
	setRenameInputValue(t, &m, "next_group")

	msg := runRenameCmd(t, m.submitRenameInput())

	if msg.Source != "draft" {
		t.Fatalf("Source = %q, want draft", msg.Source)
	}
	if !msg.HasDraft {
		t.Fatalf("HasDraft = false, want true")
	}
	if msg.Err != nil {
		t.Fatalf("Err = %v, want nil", msg.Err)
	}
}

func TestSubmitDuplicateRenameWithDraftSelectsDuplicate(t *testing.T) {
	m := newRenameTestModel(t, true)
	m.parameters.FocusParameter("demo", "__default__", "flag")
	if cmd := m.openDuplicateInput(); cmd != nil {
		_ = cmd()
	}
	setRenameInputValue(t, &m, "flag_copy")

	msg := runRenameCmd(t, m.submitRenameInput())

	if msg.Source != "draft" {
		t.Fatalf("Source = %q, want draft", msg.Source)
	}
	if msg.SelectGroupKey != "__default__" || msg.SelectParamKey != "flag_copy" {
		t.Fatalf("selection = (%q, %q), want (__default__, flag_copy)", msg.SelectGroupKey, msg.SelectParamKey)
	}
	if msg.Err != nil {
		t.Fatalf("Err = %v, want nil", msg.Err)
	}
}

func TestCancelDuplicateRenameClearsTransientDuplicate(t *testing.T) {
	m := newRenameTestModel(t, false)
	m.parameters.FocusParameter("demo", "__default__", "flag")
	if cmd := m.openDuplicateInput(); cmd != nil {
		_ = cmd()
	}
	if _, _, _, _, ok := m.parameters.CurrentTransientDuplicate(); !ok {
		t.Fatalf("transient duplicate missing before cancel")
	}

	m.cancelRenameInput()

	if m.duplicate != nil {
		t.Fatalf("duplicate session was not cleared")
	}
	if m.renameInput.IsOpen() {
		t.Fatalf("rename input is open, want closed")
	}
	if _, _, _, _, ok := m.parameters.CurrentTransientDuplicate(); ok {
		t.Fatalf("transient duplicate still selected after cancel")
	}
}

func newRenameTestModel(t *testing.T, hasDraft bool) Model {
	t.Helper()
	svc := newRenameTestService(t)
	saveRenameParametersCache(t, "demo", renameRemoteConfigRaw("1"))
	if hasDraft {
		if err := svc.SaveDraft("demo", renameRemoteConfigRaw("2")); err != nil {
			t.Fatalf("SaveDraft returned error: %v", err)
		}
	}

	m := New(svc)
	m.setActive(panels.Parameters)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{
		Projects: []core.Project{{Name: "Demo", ProjectID: "demo"}},
	})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project:    core.Project{Name: "Demo", ProjectID: "demo"},
		Tree:       renameParametersTree(),
		Source:     "cache",
		HasDraft:   hasDraft,
		StaleDraft: false,
	})
	m.parameters = m.parameters.SetBounds(0, 0, 80, 20)
	return m
}

func newRenameTestService(t *testing.T) *core.Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	return svc
}

func saveRenameParametersCache(t *testing.T, projectID string, raw json.RawMessage) {
	t.Helper()
	cache := &config.ParametersCache{
		ETag:         "etag-1",
		CachedAt:     time.Now().UTC(),
		RemoteConfig: raw,
	}
	if err := config.SaveParametersCache(projectID, cache); err != nil {
		t.Fatalf("SaveParametersCache returned error: %v", err)
	}
}

func openRenameForParameter(t *testing.T, m *Model) {
	t.Helper()
	m.parameters.FocusParameter("demo", "__default__", "flag")
	cmd := m.openRenameInput()
	if cmd != nil {
		_ = cmd()
	}
	if !m.renameInput.IsOpen() {
		t.Fatalf("rename input did not open")
	}
}

func openRenameForGroup(t *testing.T, m *Model, groupKey string) {
	t.Helper()
	if !m.parameters.FocusParameter("demo", groupKey, "group_flag") {
		t.Fatalf("group parameter did not focus")
	}
	var cmd tea.Cmd
	m.parameters, cmd = m.parameters.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if cmd != nil {
		_ = cmd()
	}
	if _, currentGroupKey, _, ok := m.parameters.CurrentGroupRef(); !ok || currentGroupKey != groupKey {
		t.Fatalf("current group = %q, ok = %v; want %q, true", currentGroupKey, ok, groupKey)
	}
	cmd = m.openRenameInput()
	if cmd != nil {
		_ = cmd()
	}
	if !m.renameInput.IsOpen() {
		t.Fatalf("rename input did not open")
	}
}

func setRenameInputValue(t *testing.T, m *Model, value string) {
	t.Helper()
	x, y := m.renameInput.Position()
	m.renameInput = m.renameInput.Close()
	next, cmd := m.renameInput.Open(x, y, 1, 80, value)
	m.renameInput = next
	if cmd != nil {
		_ = cmd()
	}
}

func runRenameCmd(t *testing.T, cmd tea.Cmd) messages.ParametersLoadedMsg {
	t.Helper()
	if cmd == nil {
		t.Fatalf("command is nil")
	}
	msg, ok := cmd().(messages.ParametersLoadedMsg)
	if !ok {
		t.Fatalf("message type = %T, want ParametersLoadedMsg", msg)
	}
	return msg
}

func renameParametersTree() *core.ParametersTree {
	return &core.ParametersTree{
		Version: "1",
		ETag:    "etag-1",
		Groups: []core.ParametersGroup{
			{
				Key:   "__default__",
				Label: "Ungrouped",
				Parameters: []core.ParametersEntry{
					{
						Key: "flag",
						Values: []core.ParametersValue{
							{Label: "default", Value: "old", RawValue: "old", ValueType: "STRING"},
						},
					},
				},
			},
			{
				Key:   "group",
				Label: "group",
				Parameters: []core.ParametersEntry{
					{
						Key: "group_flag",
						Values: []core.ParametersValue{
							{Label: "default", Value: "group-old", RawValue: "group-old", ValueType: "STRING"},
						},
					},
				},
			},
		},
	}
}

func renameRemoteConfigRaw(version string) json.RawMessage {
	cfg := firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": renameRemoteConfigParam("old"),
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"group_flag": renameRemoteConfigParam("group-old"),
				},
			},
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: version},
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return raw
}

func renameRemoteConfigParam(value string) firebase.RemoteConfigParam {
	v := firebase.RemoteConfigValue{Value: value}
	return firebase.RemoteConfigParam{
		DefaultValue: &v,
		ValueType:    "STRING",
	}
}
