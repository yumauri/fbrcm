package app

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/messages"
)

func TestHistoryRollbackBlocksProjectsWithDraftChanges(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := New(nil)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project:  project,
		Tree:     &core.ParametersTree{},
		HasDraft: true,
	})

	next, cmd, handled := m.beginHistoryRollback(messages.HistoryRollbackRequestedMsg{
		Project: project,
		Target:  core.RemoteConfigVersionEntry{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "7"}},
	})
	if !handled || cmd != nil {
		t.Fatalf("draft guard handled=%v cmd=%v", handled, cmd)
	}
	if next.historyRollback == nil || next.historyRollback.phase != historyRollbackFailed {
		t.Fatal("draft guard did not retain a recoverable failed rollback session")
	}
	if !next.dialog.IsOpen() {
		t.Fatal("draft guard did not open an explanation dialog")
	}
	next.dialog = next.dialog.SetBounds(0, 0, 100, 24)
	if view := next.dialog.View(); !strings.Contains(view, "unpublished draft changes") {
		t.Fatalf("draft guard explanation missing from dialog:\n%s", view)
	}
}

func TestHistoryRollbackPreviewShowsMetadataAndRequiresConfirmation(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	request := messages.HistoryRollbackRequestedMsg{
		Project: project,
		Target:  core.RemoteConfigVersionEntry{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "7"}},
	}
	m := New(nil)
	m.historyRollback = &historyRollbackSession{request: request, phase: historyRollbackPreparing}

	next, cmd, handled := m.updateHistoryRollbackPreview(messages.HistoryRollbackPreviewLoadedMsg{
		Project: project,
		Current: &core.ResolvedRemoteConfigVersion{Version: firebase.RemoteConfigVersion{
			VersionNumber: "8", UpdateTime: "2026-02-13T15:45:49Z",
			UpdateUser: firebase.RemoteConfigUser{Email: "current@example.com"},
		}},
		Target: &core.ResolvedRemoteConfigVersion{Version: firebase.RemoteConfigVersion{
			VersionNumber: "7", UpdateTime: "2025-04-11T13:34:34Z",
			UpdateUser: firebase.RemoteConfigUser{Email: "author@example.com"},
		}},
		Diff:    "diff\n\nSummary:\nParameters changed: 3",
		Changed: true,
	})
	if !handled || cmd != nil {
		t.Fatalf("preview handled=%v cmd=%v", handled, cmd)
	}
	if next.historyRollback == nil || next.historyRollback.phase != historyRollbackConfirming || next.historyRollback.currentVersion != "8" {
		t.Fatalf("preview session = %#v", next.historyRollback)
	}
	next.dialog = next.dialog.SetBounds(0, 0, 120, 30)
	view := next.dialog.View()
	for _, want := range []string{"Rollback to v7?", "Current:", "Target:", "author@example.com", "Parameters changed: 3", "publishes"} {
		if !strings.Contains(view, want) {
			t.Fatalf("confirmation dialog missing %q:\n%s", want, view)
		}
	}

	publishing, publishCmd, handled := next.confirmHistoryRollback()
	if !handled || publishCmd == nil || publishing.historyRollback.phase != historyRollbackPublishing || !publishing.historyRollbackModalLocked() {
		t.Fatal("confirmation did not enter locked publishing state")
	}
}

func TestHistoryRollbackCompletionRefreshesParametersAndPreferredPair(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := New(nil)
	m.historyRollback = &historyRollbackSession{request: messages.HistoryRollbackRequestedMsg{
		Project: project,
		Target:  core.RemoteConfigVersionEntry{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "7"}},
	}}
	tree := &core.ParametersTree{Version: "9"}

	next, cmd, handled := m.updateHistoryRollbackCompleted(messages.HistoryRollbackCompletedMsg{
		Project: project,
		Result: core.VersionPublishResult{
			PreviousVersion: "8", SourceVersion: "7", PublishedVersion: "9",
		},
		Tree: tree,
	})
	if !handled || cmd == nil || next.historyRollback != nil || !next.dialog.IsOpen() {
		t.Fatal("successful rollback did not finish the session and open its result dialog")
	}
	loaded, ok := cmd().(messages.ParametersLoadedMsg)
	if !ok {
		t.Fatalf("completion emitted %T", cmd())
	}
	if loaded.Tree != tree || loaded.CacheVersion != "9" || loaded.HasDraft || loaded.Err != nil {
		t.Fatalf("refresh message = %#v", loaded)
	}
}

func TestRollbackSummaryLinesExtractsOnlySummary(t *testing.T) {
	got := rollbackSummaryLines("details\n\nSummary:\n  Added: 1\n    Removed: 2")
	want := []string{"Summary:", "Added: 1", "Removed: 2"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("summary = %#v, want %#v", got, want)
	}
}
