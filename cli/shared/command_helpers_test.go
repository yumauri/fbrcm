package shared

import (
	"io"
	"testing"

	"github.com/yumauri/fbrcm/core"
)

func TestResolveParameterArgFilters(t *testing.T) {
	got, err := ResolveParameterArgFilters([]string{"flag"}, nil)
	if err != nil {
		t.Fatalf("ResolveParameterArgFilters returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "=flag" {
		t.Fatalf("filters = %#v, want =flag", got)
	}
}

func TestResolveParameterArgFiltersRejectsExistingFilter(t *testing.T) {
	_, err := ResolveParameterArgFilters([]string{"flag"}, []string{"foo"})
	if err == nil {
		t.Fatalf("ResolveParameterArgFilters accepted parameter arg with existing filter")
	}
}

func TestResolveParameterArgFiltersKeepsFiltersWithoutArg(t *testing.T) {
	in := []string{"foo"}
	got, err := ResolveParameterArgFilters(nil, in)
	if err != nil {
		t.Fatalf("ResolveParameterArgFilters returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "foo" {
		t.Fatalf("filters = %#v, want original filters", got)
	}
}

func TestHasFiltersDropsEmptyQueries(t *testing.T) {
	if HasFilters([]string{"", "  "}) {
		t.Fatalf("HasFilters returned true for empty queries")
	}
	if !HasFilters([]string{"flag"}) {
		t.Fatalf("HasFilters returned false for non-empty query")
	}
}

func TestSortProjects(t *testing.T) {
	projects := []core.Project{
		{Name: "", ProjectID: "zeta"},
		{Name: "Alpha", ProjectID: "project-b"},
		{Name: "alpha", ProjectID: "project-a"},
		{Name: "Beta", ProjectID: "project-c"},
	}

	SortProjects(projects)

	got := []string{projects[0].ProjectID, projects[1].ProjectID, projects[2].ProjectID, projects[3].ProjectID}
	want := []string{"project-a", "project-b", "project-c", "zeta"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sorted IDs = %#v, want %#v", got, want)
		}
	}
}

func TestStdinAvailableRejectsNonFileReader(t *testing.T) {
	if StdinAvailable(io.NopCloser(nil)) {
		t.Fatalf("StdinAvailable returned true for non-file reader")
	}
}
