package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/tui/components/setup"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestAccountPurgeUsesSharedDialogAboveManagementPopup(t *testing.T) {
	m := openAccountsForAdminTest(t)

	next, cmd := m.Update(keyPress('x'))
	m = next.(Model)
	if cmd == nil {
		t.Fatal("purge did not request confirmation")
	}
	next, _ = m.Update(cmd())
	m = next.(Model)
	if !m.dialog.IsOpen() || !m.setup.IsOpen() {
		t.Fatalf("purge overlay = dialog:%v setup:%v", m.dialog.IsOpen(), m.setup.IsOpen())
	}
	if setupView := ansi.Strip(m.setup.PopupViewWithFocus(m.width, m.height, false)); !strings.Contains(setupView, "Accounts") || !strings.Contains(setupView, "Profiles") {
		t.Fatalf("management popup changed while dialog opened:\n%s", setupView)
	}
	view := ansi.Strip(m.View().Content)
	for _, want := range []string{"Purge Authentication?", "Purge", "Cancel"} {
		if !strings.Contains(view, want) {
			t.Fatalf("purge overlay missing %q:\n%s", want, view)
		}
	}
}

func TestActiveProfilePurgeUsesSingleButtonErrorDialog(t *testing.T) {
	m := openAccountsForAdminTest(t)
	m = updateAdminTestMessage(t, m, tea.KeyPressMsg(tea.Key{Code: 'p', Mod: tea.ModCtrl}))
	m = updateAdminTestMessage(t, m, keyPress('x'))

	if !m.dialog.IsOpen() || !m.setup.IsOpen() {
		t.Fatalf("active-profile error = dialog:%v setup:%v", m.dialog.IsOpen(), m.setup.IsOpen())
	}
	view := ansi.Strip(m.dialog.View())
	for _, want := range []string{"Cannot Purge Active Profile", "is active", "Close"} {
		if !strings.Contains(view, want) {
			t.Fatalf("active-profile dialog missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Cancel") {
		t.Fatalf("active-profile error has confirmation actions:\n%s", view)
	}
}

func TestProfileRenameUsesSharedInlineInputAboveProfilesPopup(t *testing.T) {
	m := openAccountsForAdminTest(t)
	m = updateAdminTestMessage(t, m, tea.KeyPressMsg(tea.Key{Code: 'p', Mod: tea.ModCtrl}))
	m = updateAdminTestMessage(t, m, keyPress('r'))

	if !m.renameInput.IsOpen() || m.profileRename == nil || !m.setup.IsOpen() {
		t.Fatalf("profile rename = input:%v session:%v setup:%v", m.renameInput.IsOpen(), m.profileRename != nil, m.setup.IsOpen())
	}
	if got := m.renameInput.Value(); got != config.DefaultProfileName {
		t.Fatalf("rename value = %q, want %q", got, config.DefaultProfileName)
	}
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "Profiles") || !strings.Contains(view, "╭") {
		t.Fatalf("inline rename is not composed over Profiles popup:\n%s", view)
	}
}

func TestProfileRenameSubmitsThroughExistingInlineEditor(t *testing.T) {
	m := openAccountsForAdminTest(t)
	m = updateAdminTestMessage(t, m, tea.KeyPressMsg(tea.Key{Code: 'p', Mod: tea.ModCtrl}))
	m = updateAdminTestMessage(t, m, keyPress('r'))
	x, y := m.renameInput.Position()
	m.renameInput, _ = m.renameInput.Open(x, y, len("renamed"), 40, "renamed")

	cmd := m.submitRenameInput()
	if cmd == nil || m.renameInput.IsOpen() || m.profileRename != nil {
		t.Fatalf("rename submit = cmd:%v input:%v session:%v", cmd != nil, m.renameInput.IsOpen(), m.profileRename != nil)
	}
	result, ok := cmd().(profileRenameCompletedMsg)
	if !ok || result.err != nil {
		t.Fatalf("rename result = %#v", result)
	}
	m, refresh := m.updateProfileRenameCompleted(result)
	if refresh == nil || !m.setup.IsOpen() || config.GetActiveProfileName() != "renamed" {
		t.Fatalf("rename refresh = cmd:%v setup:%v active:%q", refresh != nil, m.setup.IsOpen(), config.GetActiveProfileName())
	}
}

func openAccountsForAdminTest(t *testing.T) Model {
	t.Helper()
	svc := newRenameTestService(t)
	if _, err := svc.AddGCloudAuth("main", "main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	m := viewTestModel(90, 24, panels.Projects)
	m.svc = svc
	m.setup = setup.New(svc)
	var cmd tea.Cmd
	m.setup, cmd = m.setup.OpenAccounts()
	return finishSetupInspection(t, m, cmd)
}

func updateAdminTestMessage(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	next, cmd := m.Update(msg)
	m = next.(Model)
	if cmd != nil {
		next, _ = m.Update(cmd())
		m = next.(Model)
	}
	return m
}
