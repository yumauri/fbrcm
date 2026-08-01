package shared

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
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
	got, err := FilterProjects(projects, []string{"=alpha"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ProjectID != "alpha" {
		t.Fatalf("FilterProjects = %+v", got)
	}
}

func TestProjectFiltersMatchRepositoryAliases(t *testing.T) {
	setupProjectAliasResolutionTest(t, `[projects.aliases]
prod = "acme-production-42"
production = "acme-production-42"
stage = "acme-staging-42"
`)
	projects := []core.Project{
		{Name: "Production", ProjectID: "acme-production-42"},
		{Name: "Staging", ProjectID: "acme-staging-42"},
	}
	for _, raw := range []string{"=prod", "^prod", "/duct", "~prd"} {
		got, err := FilterProjects(projects, []string{raw})
		if err != nil || len(got) != 1 || got[0].ProjectID != "acme-production-42" {
			t.Fatalf("FilterProjects(%q) = %#v, %v", raw, got, err)
		}
	}
}

func TestTargetFiltersMatchAliasesAndDeduplicateTemplates(t *testing.T) {
	setupProjectAliasResolutionTest(t, `[projects.aliases]
prod = "acme-production-42"
production = "acme-production-42"
`)
	projects := []core.Project{{
		Name: "Production", ProjectID: "acme-production-42",
		Templates: []rctarget.Kind{rctarget.Client, rctarget.Server}, PrimaryTemplate: rctarget.Client,
	}}
	got, err := FilterProjectTargets(projects, []string{"=prod", "=production"})
	if err != nil || len(got) != 2 || got[0].ProjectID != "acme-production-42" || got[1].ProjectID != "server@acme-production-42" {
		t.Fatalf("unqualified aliases = %#v, %v", got, err)
	}
	got, err = FilterProjectTargets(projects, []string{"server@=prod"})
	if err != nil || len(got) != 1 || got[0].ProjectID != "server@acme-production-42" {
		t.Fatalf("server alias = %#v, %v", got, err)
	}
}

func TestFilterProjectTargets(t *testing.T) {
	projects := []core.Project{
		{Name: "Mobile", ProjectID: "mobile-prod"},
		{Name: "API", ProjectID: "api-prod"},
	}
	got, err := FilterProjectTargets(projects, []string{"client@=mobile-prod", "server@=api-prod", "server@=api-prod"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ProjectID != "mobile-prod" || got[1].ProjectID != "server@api-prod" {
		t.Fatalf("targets = %#v", got)
	}
}

func TestFilterProjectTargetsDefaultsToClient(t *testing.T) {
	projects := []core.Project{{Name: "Mobile", ProjectID: "mobile-prod"}}
	got, err := FilterProjectTargets(projects, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ProjectID != "mobile-prod" {
		t.Fatalf("targets = %#v", got)
	}
}

func TestFilterProjectTargetsUsesConfiguredViewsAndExplicitOverride(t *testing.T) {
	projects := []core.Project{{
		Name:            "Demo",
		ProjectID:       "demo",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Server,
	}}
	got, err := FilterProjectTargets(projects, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ProjectID != "server@demo" || got[1].ProjectID != "demo" {
		t.Fatalf("configured targets = %#v, want server then client", got)
	}
	got, err = FilterProjectTargets(projects, []string{"=demo"})
	if err != nil || len(got) != 2 || got[0].ProjectID != "server@demo" || got[1].ProjectID != "demo" {
		t.Fatalf("unqualified targets = %#v, %v", got, err)
	}
	got, err = FilterProjectTargets(projects, []string{"client@=demo"})
	if err != nil || len(got) != 1 || got[0].ProjectID != "demo" {
		t.Fatalf("explicit client targets = %#v, %v", got, err)
	}
}

func TestMatchProjectTargetSeparatesTemplateKinds(t *testing.T) {
	server := core.Project{Name: "Demo", ProjectID: "server@demo"}
	for filter, want := range map[string]bool{
		"server@=demo": true,
		"client@=demo": false,
		"=demo":        false,
	} {
		got, err := MatchProjectTarget(server, []string{filter})
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("MatchProjectTarget(server, %q) = %t, want %t", filter, got, want)
		}
	}
}

func TestMatchProjectTargetUsesConfiguredViewsForUnqualifiedFilter(t *testing.T) {
	server := core.Project{
		Name:            "Demo",
		ProjectID:       "server@demo",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Server,
	}
	got, err := MatchProjectTarget(server, []string{"=demo"})
	if err != nil || !got {
		t.Fatalf("MatchProjectTarget configured server = %t, %v", got, err)
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

func TestProjectExpressionUsesUnderlyingProjectIDForServerTarget(t *testing.T) {
	compiled, ok := CompileExpr(`project_id == "demo"`, "")
	if !ok {
		t.Fatal("CompileExpr failed")
	}
	match, valid := MatchProjectByCompiledExpr(compiled, core.Project{ProjectID: "server@demo", Name: "Demo"}, &firebase.RemoteConfig{})
	if !valid || !match {
		t.Fatalf("MatchProjectByCompiledExpr = %t/%t", match, valid)
	}
}

func TestMatchConditionByCompiledExpr(t *testing.T) {
	compiled, ok := CompileExpr(`usage_count == 0 && priority == 1`, "demo")
	if !ok || compiled == nil {
		t.Fatal("CompileExpr failed")
	}
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	entry := core.ConditionEntry{Name: "unused", Priority: 1}
	match, ok := MatchConditionByCompiledExpr(compiled, project, entry)
	if !ok || !match {
		t.Fatalf("MatchConditionByCompiledExpr = %v/%v", match, ok)
	}
}

func TestHighlightFilters(t *testing.T) {
	filters := ParseFilters([]string{"feat"})
	got := HighlightFilters("feature_login", filters)
	if len(got) == 0 {
		t.Fatal("HighlightFilters returned no indices")
	}
}
