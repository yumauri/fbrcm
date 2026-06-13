package projects

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestProjectsJSONCopiesFieldsAndAddsURL(t *testing.T) {
	projects := []core.Project{
		{
			Name:          "Project A",
			ProjectID:     "project-a",
			ProjectNumber: "123",
			State:         "ACTIVE",
			ETag:          "etag",
			AuthID:        "auth-main",
			DiscoveredBy:  []string{"auth-main"},
			UpdatedAt:     "2026-06-14T09:10:11Z",
			SyncedAt:      "2026-06-14T10:11:12Z",
		},
	}

	got := projectsJSON(projects, true)
	if len(got) != 1 {
		t.Fatalf("projectsJSON length = %d, want 1", len(got))
	}
	row := got[0]
	if row.Project != "Project A" || row.ProjectID != "project-a" || row.Number != "123" || row.AuthID != "auth-main" {
		t.Fatalf("project row = %#v, want Project A/project-a/123/auth-main", row)
	}
	if row.URL != firebase.RemoteConfigConsoleURL("project-a") {
		t.Fatalf("url = %q, want Remote Config console URL", row.URL)
	}
	projects[0].DiscoveredBy[0] = "changed"
	if row.DiscoveredBy[0] != "auth-main" {
		t.Fatalf("DiscoveredBy was not copied: %#v", row.DiscoveredBy)
	}
}

func TestHumanDateTime(t *testing.T) {
	if got := humanDateTime(""); got != "" {
		t.Fatalf("humanDateTime(empty) = %q, want empty", got)
	}
	if got := humanDateTime("not-a-date"); got != "not-a-date" {
		t.Fatalf("humanDateTime(invalid) = %q, want original", got)
	}
	if got := humanDateTime("2026-06-14T09:10:11Z"); !strings.Contains(got, "2026-06-14") || !strings.HasSuffix(got, ":10:11") {
		t.Fatalf("humanDateTime(valid) = %q, want formatted local date/time", got)
	}
}

func TestIndicesSet(t *testing.T) {
	got := indicesSet([]int{2, 4, 2})
	want := map[int]bool{2: true, 4: true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("indicesSet = %#v, want %#v", got, want)
	}
}

func TestRenderProjectsTablePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	table := renderProjectsTable([]core.Project{
		{Name: "Project A", ProjectID: "project-a", ProjectNumber: "123", AuthID: "auth-main", UpdatedAt: "2026-06-14T09:10:11Z", SyncedAt: "bad-date"},
	}, nil, true)

	for _, want := range []string{"Project", "Project ID", "Project A", "project-a", "123", "auth-main", "bad-date", firebase.RemoteConfigConsoleURL("project-a")} {
		if !strings.Contains(table, want) {
			t.Fatalf("renderProjectsTable = %q, want substring %q", table, want)
		}
	}
}
