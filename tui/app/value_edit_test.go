package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
)

func TestValueEditLoadedMsgDraft(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	msg := (Model{}).valueEditLoadedMsg(project, "group", "flag", &core.ParametersTree{}, true, true, false)

	if msg.Project.ProjectID != "demo" {
		t.Fatalf("ProjectID = %q, want demo", msg.Project.ProjectID)
	}
	if msg.Source != "draft" {
		t.Fatalf("Source = %q, want draft", msg.Source)
	}
	if !msg.HasDraft {
		t.Fatalf("HasDraft = false, want true")
	}
	if !msg.StaleDraft {
		t.Fatalf("StaleDraft = false, want true")
	}
	if msg.SelectGroupKey != "group" || msg.SelectParamKey != "flag" {
		t.Fatalf("selection = (%q, %q), want (group, flag)", msg.SelectGroupKey, msg.SelectParamKey)
	}
}

func TestValueEditLoadedMsgPublish(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	msg := (Model{}).valueEditLoadedMsg(project, "group", "flag", &core.ParametersTree{}, false, true, true)

	if msg.Source != "firebase" {
		t.Fatalf("Source = %q, want firebase", msg.Source)
	}
	if msg.HasDraft {
		t.Fatalf("HasDraft = true, want false")
	}
	if msg.StaleDraft {
		t.Fatalf("StaleDraft = true, want false")
	}
}

func TestUpdateOpenModalRoutesHighestPriorityOverlay(t *testing.T) {
	m := New(nil)
	m.dialog = m.dialog.Open(dialogcmp.Config{Title: "Confirm", Body: []string{"body"}})
	m.boolPicker = m.boolPicker.Open(0, 0, true)

	next, _, handled := m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	if !handled {
		t.Fatalf("updateOpenModal did not handle open dialog")
	}
	if next.dialog.IsOpen() {
		t.Fatalf("dialog is open after cancel")
	}
	if !next.boolPicker.IsOpen() {
		t.Fatalf("bool picker closed while lower-priority overlay should remain open")
	}

	next, _, handled = next.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	if !handled {
		t.Fatalf("updateOpenModal did not handle bool picker")
	}
	if next.boolPicker.IsOpen() {
		t.Fatalf("bool picker is open after cancel")
	}
}

func TestCloseOverlaysClosesOpenEditors(t *testing.T) {
	m := New(nil)
	m.dialog = m.dialog.Open(dialogcmp.Config{Title: "Confirm", Body: []string{"body"}})
	m.boolPicker = m.boolPicker.Open(0, 0, true)
	m.jsonInput, _ = m.jsonInput.Open(80, 24, `{"enabled":true}`)

	m.closeOverlays()

	if m.dialog.IsOpen() || m.boolPicker.IsOpen() || m.jsonInput.IsOpen() {
		t.Fatalf("closeOverlays left an overlay open")
	}
}
