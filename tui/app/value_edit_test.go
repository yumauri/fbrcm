package app

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
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
