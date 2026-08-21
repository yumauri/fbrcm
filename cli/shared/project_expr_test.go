package shared

import (
	"strings"
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

func TestFilterProjectsWithAliasesCanExcludeLocalAliases(t *testing.T) {
	projects := []core.Project{{Name: "Production", ProjectID: "acme-production-42"}}
	if got := FilterProjectsWithAliases(projects, []string{"=prod"}, nil); len(got) != 0 {
		t.Fatalf("filter without aliases = %+v, want no match", got)
	}
	aliases := map[string][]string{"acme-production-42": {"prod"}}
	if got := FilterProjectsWithAliases(projects, []string{"=prod"}, aliases); len(got) != 1 {
		t.Fatalf("filter with aliases = %+v, want one match", got)
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

func TestFilterProjectTargetsWithAliasesCanExcludeLocalAliases(t *testing.T) {
	projects := []core.Project{{Name: "Production", ProjectID: "acme-production-42"}}
	got, err := FilterProjectTargetsWithAliases(projects, []string{"=prod"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("filter without aliases = %+v, want no match", got)
	}

	got, err = FilterProjectTargetsWithAliases(projects, []string{"server@=prod"}, map[string][]string{"acme-production-42": {"prod"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ProjectID != "server@acme-production-42" {
		t.Fatalf("filter with aliases = %+v, want server target", got)
	}
}

func TestFilterProjectTargetsWithAliasesPreservesRemoteFilterModes(t *testing.T) {
	projects := []core.Project{{Name: "My Test Application", ProjectID: "my-test-project"}}
	for _, selector := range []string{"test", "^my-test", "/Test App", "~mtpr", "server@test"} {
		got, err := FilterProjectTargetsWithAliases(projects, []string{selector}, nil)
		if err != nil {
			t.Fatalf("selector %q: %v", selector, err)
		}
		wantID := "my-test-project"
		if strings.HasPrefix(selector, "server@") {
			wantID = "server@my-test-project"
		}
		if len(got) != 1 || got[0].ProjectID != wantID {
			t.Fatalf("selector %q = %+v, want %q", selector, got, wantID)
		}
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

	compiled, err := CompileExpr(`project_id == "demo"`, "")
	if err != nil || compiled == nil {
		t.Fatalf("CompileExpr failed: %v", err)
	}
	match, err := MatchProjectByCompiledExpr(compiled, project, cfg)
	if err != nil || !match {
		t.Fatalf("MatchProjectByCompiledExpr = %v, %v", match, err)
	}
	if match, err = MatchProjectByExpr(project, cfg, `project_id == "demo"`); err != nil || !match {
		t.Fatal("MatchProjectByExpr should match")
	}
	if match, err = MatchProjectByExpr(project, cfg, `project_id == "other"`); err != nil || match {
		t.Fatal("MatchProjectByExpr should not match other id")
	}
}

func TestProjectExpressionUsesUnderlyingProjectIDForServerTarget(t *testing.T) {
	compiled, err := CompileExpr(`project_id == "demo"`, "")
	if err != nil {
		t.Fatalf("CompileExpr failed: %v", err)
	}
	match, err := MatchProjectByCompiledExpr(compiled, core.Project{ProjectID: "server@demo", Name: "Demo"}, &firebase.RemoteConfig{})
	if err != nil || !match {
		t.Fatalf("MatchProjectByCompiledExpr = %t, %v", match, err)
	}
}

func TestMatchConditionByCompiledExpr(t *testing.T) {
	compiled, err := CompileExpr(`usage_count == 0 && priority == 1`, "demo")
	if err != nil || compiled == nil {
		t.Fatalf("CompileExpr failed: %v", err)
	}
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	entry := core.ConditionEntry{Name: "unused", Priority: 1}
	match, err := MatchConditionByCompiledExpr(compiled, project, entry)
	if err != nil || !match {
		t.Fatalf("MatchConditionByCompiledExpr = %v, %v", match, err)
	}
}

func TestCompileExprReturnsError(t *testing.T) {
	if _, err := CompileExpr(`project_id ==`, "demo"); err == nil {
		t.Fatal("CompileExpr error is nil")
	}
}

func TestMatchParameterByCompiledExprReturnsEvaluationError(t *testing.T) {
	compiled, err := CompileExpr(`jq(value, ".[" ) == true`, "demo")
	if err != nil {
		t.Fatalf("CompileExpr: %v", err)
	}
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	cfg := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: `{}`}},
	}}
	if _, err := MatchParameterByCompiledExpr(compiled, project, cfg, "flag", DefaultRootGroupLabel); err == nil {
		t.Fatal("evaluation error is nil")
	}
}

func TestHighlightFilters(t *testing.T) {
	filters := ParseFilters([]string{"feat"})
	got := HighlightFilters("feature_login", filters)
	if len(got) == 0 {
		t.Fatal("HighlightFilters returned no indices")
	}
}
