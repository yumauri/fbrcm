package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/tui/components/setup"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

func TestQuitPromptsOnlyForDirtyDetails(t *testing.T) {
	clean := conditionalValueDetailsTestModel()
	_, cmd, handled := clean.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	if !handled || cmd == nil || clean.dialog.IsOpen() {
		t.Fatalf("clean quit = handled:%v cmd:%v dialog:%v; want immediate quit", handled, cmd != nil, clean.dialog.IsOpen())
	}

	dirty := conditionalValueDetailsTestModel()
	dirty.details, _ = dirty.details.ActivateName()
	dirty.details, _ = dirty.details.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Text: "x"}))
	dirty.details = dirty.details.DeactivateField()
	if !dirty.details.Dirty() {
		t.Fatal("test Details form is not dirty")
	}
	dirty, cmd, handled = dirty.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	if !handled || cmd != nil || !dirty.dialog.IsOpen() {
		t.Fatalf("dirty quit = handled:%v cmd:%v dialog:%v; want confirmation", handled, cmd != nil, dirty.dialog.IsOpen())
	}
}

func TestDirtyDetailsQuitPromptsWhenAnotherPanelIsActive(t *testing.T) {
	m := conditionalValueDetailsTestModel()
	m.details, _ = m.details.ActivateName()
	m.details, _ = m.details.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Text: "x"}))
	m.details = m.details.DeactivateField()
	m.setActive(m.selectedParametersTab())

	m, cmd, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'q', Text: "q"}))
	if !handled || cmd != nil || !m.dialog.IsOpen() {
		t.Fatalf("background dirty quit = handled:%v cmd:%v dialog:%v", handled, cmd != nil, m.dialog.IsOpen())
	}
}

func TestForceQuitBypassesDirtyDetailsAndOpenDialog(t *testing.T) {
	m := conditionalValueDetailsTestModel()
	m.details, _ = m.details.ActivateName()
	m.details, _ = m.details.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Text: "x"}))
	m.details = m.details.DeactivateField()
	m.requestQuit()
	if !m.dialog.IsOpen() {
		t.Fatal("quit confirmation did not open")
	}

	next, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	if cmd == nil || !next.(Model).dialog.IsOpen() {
		t.Fatalf("force quit = cmd:%v dialog:%v; want immediate quit without modal processing", cmd != nil, next.(Model).dialog.IsOpen())
	}
}

func TestDirtyDetailsBlocksAccountsAndSetupQuitUsesGuard(t *testing.T) {
	m := conditionalValueDetailsTestModel()
	m.details, _ = m.details.ActivateName()
	m.details, _ = m.details.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Text: "x"}))
	m.details = m.details.DeactivateField()

	m, cmd, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'A', Text: "A"}))
	if !handled || cmd != nil || !m.dialog.IsOpen() || m.setup.IsOpen() {
		t.Fatalf("dirty accounts = handled:%v cmd:%v dialog:%v setup:%v", handled, cmd != nil, m.dialog.IsOpen(), m.setup.IsOpen())
	}
	if view := m.dialog.View(); !strings.Contains(view, "Unsaved Details") || !strings.Contains(view, "Save or discard") {
		t.Fatalf("dirty accounts dialog missing guidance:\n%s", view)
	}
	if enabled, reason := m.globalHelpActionAvailability(tuiconfig.ActionAccounts); enabled || !strings.Contains(reason, "save or discard") {
		t.Fatalf("accounts availability = %v, %q", enabled, reason)
	}

	m.dialog = m.dialog.Close()
	next, quitCmd := m.Update(setup.QuitRequestedMsg{})
	m = next.(Model)
	if quitCmd != nil || !m.dialog.IsOpen() {
		t.Fatalf("setup quit with dirty Details = cmd:%v dialog:%v", quitCmd != nil, m.dialog.IsOpen())
	}
}
