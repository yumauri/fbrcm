package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
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
