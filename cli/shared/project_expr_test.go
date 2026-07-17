package shared

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestParseFiltersAndMatchAnyFilter(t *testing.T) {
	filters := ParseFilters([]string{"=demo", "  ", "^prod"})
	if len(filters) != 2 {
		t.Fatalf("ParseFilters = %+v, want 2 entries", filters)
	}
	if filters[0].Mode != filter.ModeExact || filters[0].Query != "demo" {
		t.Fatalf("first filter = %+v", filters[0])
	}
	if !MatchAnyFilter("demo", filters) {
		t.Fatal("MatchAnyFilter should match demo exact")
	}
	if MatchAnyFilter("other", filters) {
		t.Fatal("MatchAnyFilter should not match other")
	}
}

func TestFilterProjects(t *testing.T) {
	projects := []core.Project{
		{Name: "Alpha", ProjectID: "alpha"},
		{Name: "Beta Prod", ProjectID: "beta"},
	}
	got := FilterProjects(projects, []string{"=alpha"})
	if len(got) != 1 || got[0].ProjectID != "alpha" {
		t.Fatalf("FilterProjects = %+v", got)
	}
}

func TestSingleExactFilter(t *testing.T) {
	if !SingleExactFilter([]string{"=alpha"}) {
		t.Fatal("single exact filter was not recognized")
	}
	for _, filters := range [][]string{{"alpha"}, {"=alpha", "=beta"}, {"", "  "}} {
		if SingleExactFilter(filters) {
			t.Fatalf("SingleExactFilter(%q) = true", filters)
		}
	}
}

func TestCompileExprAndMatchProjectByExpr(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "ios"}},
	}
	project := core.Project{ProjectID: "demo", Name: "Demo"}

	compiled, ok := CompileExpr(`project_id == "demo"`, "")
	if !ok || compiled == nil {
		t.Fatal("CompileExpr failed")
	}
	match, ok := MatchProjectByCompiledExpr(compiled, project, cfg)
	if !ok || !match {
		t.Fatalf("MatchProjectByCompiledExpr = %v/%v", match, ok)
	}
	if !MatchProjectByExpr(project, cfg, `project_id == "demo"`) {
		t.Fatal("MatchProjectByExpr should match")
	}
	if MatchProjectByExpr(project, cfg, `project_id == "other"`) {
		t.Fatal("MatchProjectByExpr should not match other id")
	}
}

func TestHighlightFilters(t *testing.T) {
	filters := ParseFilters([]string{"feat"})
	got := HighlightFilters("feature_login", filters)
	if len(got) == 0 {
		t.Fatal("HighlightFilters returned no indices")
	}
}
