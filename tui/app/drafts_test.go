package app

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
)

func TestDraftMutationCmdDraftSelection(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	tree := &core.ParametersTree{}

	msg := runDraftMutationCmd(t, (Model{}).draftMutationCmd(project, false, "group", "flag", false, func(context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return &core.ParametersCache{ETag: "etag"}, tree, true, nil
	}))

	if msg.Project.ProjectID != "demo" {
		t.Fatalf("ProjectID = %q, want demo", msg.Project.ProjectID)
	}
	if msg.Tree != tree {
		t.Fatalf("Tree was not propagated")
	}
	if msg.Source != "draft" {
		t.Fatalf("Source = %q, want draft", msg.Source)
	}
	if msg.CacheSource != "cache" {
		t.Fatalf("CacheSource = %q, want cache", msg.CacheSource)
	}
	if !msg.HasDraft {
		t.Fatalf("HasDraft = false, want true")
	}
	if msg.StaleDraft {
		t.Fatalf("StaleDraft = true, want false for non-stale model")
	}
	if msg.SelectGroupKey != "group" || msg.SelectParamKey != "flag" {
		t.Fatalf("selection = (%q, %q), want (group, flag)", msg.SelectGroupKey, msg.SelectParamKey)
	}
}

func TestDraftMutationCmdPublishSource(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}

	msg := runDraftMutationCmd(t, (Model{}).draftMutationCmd(project, true, "", "", false, func(context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return &core.ParametersCache{ETag: "etag"}, &core.ParametersTree{}, false, nil
	}))

	if msg.Source != "firebase" {
		t.Fatalf("Source = %q, want firebase", msg.Source)
	}
	if msg.HasDraft {
		t.Fatalf("HasDraft = true, want false")
	}
	if msg.StaleDraft {
		t.Fatalf("StaleDraft = true, want false after publish")
	}
}

func TestDraftMutationCmdErrorKeepsDialogCloseIntent(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	wantErr := errors.New("boom")

	msg := runDraftMutationCmd(t, (Model{}).draftMutationCmd(project, false, "", "", true, func(context.Context) (*core.ParametersCache, *core.ParametersTree, bool, error) {
		return nil, nil, false, wantErr
	}))

	if !errors.Is(msg.Err, wantErr) {
		t.Fatalf("Err = %v, want %v", msg.Err, wantErr)
	}
	if !msg.CloseDetails {
		t.Fatalf("CloseDetails = false, want true")
	}
	if msg.HasDraft {
		t.Fatalf("HasDraft = true, want false for empty model")
	}
}

func runDraftMutationCmd(t *testing.T, cmd tea.Cmd) messages.ParametersLoadedMsg {
	t.Helper()
	msg, ok := cmd().(messages.ParametersLoadedMsg)
	if !ok {
		t.Fatalf("message type = %T, want ParametersLoadedMsg", msg)
	}
	return msg
}
