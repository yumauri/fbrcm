package projects

import (
	"reflect"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/yumauri/fbrcm/cli/shared"
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
			Disabled:      true,
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
	if row.Project != "Project A" || row.ProjectID != "project-a" || row.Number != "123" || row.AuthID != "auth-main" || !row.Disabled {
		t.Fatalf("project row = %#v, want Project A/project-a/123/auth-main", row)
	}
	if row.URL != firebase.RemoteConfigConsoleURL("project-a") {
		t.Fatalf("url = %q, want Remote Config console URL", row.URL)
	}
	if row.Aliases == nil || len(row.Aliases) != 0 {
		t.Fatalf("aliases = %#v, want empty array", row.Aliases)
	}
	if len(row.Templates) != 1 || row.Templates[0] != "client" || row.PrimaryTemplate != "client" {
		t.Fatalf("templates = %v/%q, want client/client", row.Templates, row.PrimaryTemplate)
	}
	projects[0].DiscoveredBy[0] = "changed"
	if row.DiscoveredBy[0] != "auth-main" {
		t.Fatalf("DiscoveredBy was not copied: %#v", row.DiscoveredBy)
	}
}

func TestProjectsOutputIncludesRepositoryAliases(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	project := core.Project{Name: "Production", ProjectID: "acme-production-42", AuthID: "main"}
	aliasesByID := map[string][]string{"acme-production-42": {"prod", "production"}}

	rows := projectsJSONWithAliases([]core.Project{project}, false, aliasesByID)
	if len(rows) != 1 || !reflect.DeepEqual(rows[0].Aliases, []string{"prod", "production"}) {
		t.Fatalf("project aliases JSON = %#v", rows)
	}
	table := renderProjectsTableAtWidth([]core.Project{project}, nil, false, aliasesByID, 200)
	if !strings.Contains(table, "Aliases") || !strings.Contains(table, "prod, production") {
		t.Fatalf("project aliases table = %q", table)
	}
}

func TestProjectsTableFitsNarrowTerminalWithAliases(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	project := core.Project{
		Name: strings.Repeat("Production Project ", 4), ProjectID: "acme-production-project-42", ProjectNumber: "123456789012345",
		AuthID: "long-auth-identity", UpdatedAt: "2026-06-14T09:10:11Z", SyncedAt: "2026-06-14T10:11:12Z",
	}
	aliasesByID := map[string][]string{project.ProjectID: {strings.Repeat("production-alias-", 4)}}
	output := renderProjectsTableAtWidth([]core.Project{project}, nil, false, aliasesByID, 90)
	for index, line := range strings.Split(output, "\n") {
		if width := lipgloss.Width(line); width > 90 {
			t.Fatalf("line %d width = %d, want <= 90:\n%s", index, width, output)
		}
	}
	if !strings.Contains(output, "…") {
		t.Fatalf("narrow table did not crop:\n%s", output)
	}
}

func TestFormatDateTime(t *testing.T) {
	if got := shared.FormatDateTime(""); got != "" {
		t.Fatalf("FormatDateTime(empty) = %q, want empty", got)
	}
	if got := shared.FormatDateTime("not-a-date"); got != "not-a-date" {
		t.Fatalf("FormatDateTime(invalid) = %q, want original", got)
	}
	if got := shared.FormatDateTime("2026-06-14T09:10:11Z"); !strings.Contains(got, "2026-06-14") || !strings.HasSuffix(got, ":10:11") {
		t.Fatalf("FormatDateTime(valid) = %q, want formatted local date/time", got)
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

	table := renderProjectsTableAtWidth([]core.Project{
		{Name: "Project A", ProjectID: "project-a", ProjectNumber: "123", AuthID: "auth-main", Disabled: true, UpdatedAt: "2026-06-14T09:10:11Z", SyncedAt: "bad-date"},
	}, nil, true, nil, 500)

	for _, want := range []string{"Project", "Project ID", "Project A", "project-a", "123", "auth-main (disabled)", "bad-date", firebase.RemoteConfigConsoleURL("project-a")} {
		if !strings.Contains(table, want) {
			t.Fatalf("renderProjectsTable = %q, want substring %q", table, want)
		}
	}
}
