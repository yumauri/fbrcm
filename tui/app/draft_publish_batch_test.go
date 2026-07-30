package app

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func TestDraftPublishBatchReviewsEveryPreparedProject(t *testing.T) {
	m := New(nil)
	items := make([]draftPublishItem, 0, 4)
	for _, id := range []string{"a", "b", "c", "d"} {
		items = append(items, draftPublishItem{
			project: core.Project{ProjectID: id},
			plan:    testDraftPublishPlan("old", "new"),
		})
	}
	m.draftPublish = &draftPublishBatch{phase: draftPublishPreparing}

	var handled bool
	m, _, handled = m.updateDraftPublishPrepared(draftPublishPreparedMsg{items: items})
	if !handled || !m.dialog.IsOpen() {
		t.Fatal("first review dialog was not opened")
	}
	for index := 0; index < len(items); index++ {
		note := "note-" + items[index].project.ProjectID
		m, _, handled = m.updateDraftPublishDecision(draftPublishDecisionMsg{decision: "approve", changeNote: note})
		if !handled {
			t.Fatalf("decision %d was not handled", index)
		}
	}
	if m.draftPublish.phase != draftPublishPublishing {
		t.Fatalf("phase = %v, want publishing", m.draftPublish.phase)
	}
	for index, item := range m.draftPublish.items {
		if !item.approved {
			t.Fatalf("item %d (%s) was not approved", index, item.project.ProjectID)
		}
		if item.plan.ChangeNote != "note-"+item.project.ProjectID {
			t.Fatalf("item %d change note = %q", index, item.plan.ChangeNote)
		}
	}
}

func TestDraftPublishBatchReviewUsesPreparedCandidate(t *testing.T) {
	m := New(nil)
	m.dialog = m.dialog.SetBounds(0, 0, 120, 40)
	m.draftPublish = &draftPublishBatch{phase: draftPublishPreparing}
	item := draftPublishItem{project: core.Project{Name: "Demo", ProjectID: "demo"}, plan: testDraftPublishPlan("old", "prepared-new")}
	item.plan.ChangeNote = "Prepared release"

	m, _, _ = m.updateDraftPublishPrepared(draftPublishPreparedMsg{items: []draftPublishItem{item}})
	view := m.dialog.View()
	if !strings.Contains(view, "prepared-new") || !strings.Contains(view, "old") {
		t.Fatalf("review view does not contain prepared diff:\n%s", view)
	}
	if m.dialog.InputValue() != "Prepared release" {
		t.Fatalf("review change note = %q", m.dialog.InputValue())
	}
}

func TestDraftPublishResultsKeepProjectOrder(t *testing.T) {
	m := New(nil)
	m.draftPublish = &draftPublishBatch{
		items: []draftPublishItem{{project: core.Project{ProjectID: "a"}}, {project: core.Project{ProjectID: "b"}}, {project: core.Project{ProjectID: "c"}}},
		results: []draftPublishResult{
			{project: core.Project{ProjectID: "b"}, status: "failed"},
			{project: core.Project{ProjectID: "a"}, status: "published"},
			{project: core.Project{ProjectID: "c"}, status: "published"},
		},
	}
	got := m.orderedDraftPublishResults()
	if len(got) != 3 || got[0].project.ProjectID != "a" || got[1].project.ProjectID != "b" || got[2].project.ProjectID != "c" {
		t.Fatalf("ordered results = %+v", got)
	}
}

func testDraftPublishPlan(from, to string) *core.DraftPublishPlan {
	return &core.DraftPublishPlan{
		Latest:     &core.ParametersCache{RemoteConfig: testDraftRemoteConfig(from)},
		Candidate:  testDraftRemoteConfig(to),
		HasChanges: from != to,
	}
}

func testDraftRemoteConfig(value string) json.RawMessage {
	return json.RawMessage(`{"version":{"versionNumber":"1"},"parameters":{"flag":{"defaultValue":{"value":"` + value + `"}}}}`)
}
