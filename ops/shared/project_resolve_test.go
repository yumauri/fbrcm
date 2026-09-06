package shared

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func TestResolveProjectTargetForExecutionRequiresLiteralTargetWithoutLocalReads(t *testing.T) {
	ctx := core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy())
	cmd := &cobra.Command{Use: "export"}

	for query, want := range map[string]string{
		"demo":        "demo",
		"client@demo": "demo",
		"SERVER@demo": "server@demo",
	} {
		project, err := ResolveProjectTargetForExecution(ctx, cmd, nil, query)
		if err != nil || project.ProjectID != want || project.Name != "demo" {
			t.Fatalf("ResolveProjectTargetForExecution(%q) = %#v, %v; want %q", query, project, err, want)
		}
	}
	for _, query := range []string{"=demo", "server@=demo", "demo project", "server@"} {
		if _, err := ResolveProjectTargetForExecution(ctx, cmd, nil, query); err == nil {
			t.Errorf("ResolveProjectTargetForExecution(%q) accepted a non-literal target", query)
		}
	}
}

func TestResolvePhysicalProjectForExecutionRequiresLiteralIDWithoutLocalReads(t *testing.T) {
	ctx := core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy())
	cmd := &cobra.Command{Use: "open"}

	project, err := ResolvePhysicalProjectForExecution(ctx, cmd, nil, "demo-project")
	if err != nil || project.ProjectID != "demo-project" || project.Name != "demo-project" {
		t.Fatalf("ResolvePhysicalProjectForExecution() = %#v, %v", project, err)
	}
	for _, query := range []string{"=demo", "client@demo", "server@demo", "demo project"} {
		if _, err := ResolvePhysicalProjectForExecution(ctx, cmd, nil, query); err == nil {
			t.Errorf("ResolvePhysicalProjectForExecution(%q) accepted a selector", query)
		}
	}
}

func TestResolveProjectScopedResourceForExecutionRejectsTemplatePrefixes(t *testing.T) {
	cmd := &cobra.Command{}
	for _, query := range []string{"client@demo-project", "server@demo-project"} {
		_, err := ResolveProjectScopedResourceForExecution(context.Background(), cmd, nil, query, "app")
		var argumentErr *ArgumentError
		if err == nil || !errors.As(err, &argumentErr) || !strings.Contains(err.Error(), "app commands are project-scoped") {
			t.Errorf("query %q error = %v", query, err)
		}
	}
}

func TestResolveProjectScopedResourceRequiresLiteralIDStateless(t *testing.T) {
	ctx := core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy())
	cmd := &cobra.Command{Use: "show"}
	project, err := ResolveProjectScopedResourceForExecution(ctx, cmd, nil, "demo-project", "app")
	if err != nil || project.ProjectID != "demo-project" {
		t.Fatalf("literal project = %#v, %v", project, err)
	}
	for _, query := range []string{"~demo", "/demo", "^demo", "=demo"} {
		if _, err := ResolveProjectScopedResourceForExecution(ctx, cmd, nil, query, "app"); err == nil {
			t.Errorf("stateless selector %q was accepted", query)
		}
	}
}

func TestResolveProjectNumberUsesConfiguredProject(t *testing.T) {
	cmd := &cobra.Command{Use: "show"}
	project, err := resolveProjectNumber(cmd, []core.Project{
		{Name: "Other", ProjectID: "other", ProjectNumber: "456"},
		{Name: "Demo", ProjectID: "demo", ProjectNumber: "123", AuthID: "main"},
	}, "123")
	if err != nil || project.ProjectID != "demo" || project.AuthID != "main" {
		t.Fatalf("resolveProjectNumber() = %#v, %v", project, err)
	}
}

func TestResolveProjectNumberReportsMissingProfileMapping(t *testing.T) {
	cmd := &cobra.Command{Use: "show"}
	cmd.SetOut(&bytes.Buffer{})
	_, err := resolveProjectNumber(cmd, []core.Project{{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"}}, "456")
	var selection *ProjectResolutionError
	if !errors.As(err, &selection) || selection.Kind != "not_found" || selection.Query != "456" || !strings.Contains(err.Error(), "--project") {
		t.Fatalf("resolveProjectNumber() error = %#v", err)
	}
}

func TestResolveProjectNumberForStatelessExecutionUsesNumberDirectly(t *testing.T) {
	ctx := core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy())
	project, err := ResolveProjectNumberForExecution(ctx, &cobra.Command{Use: "show"}, nil, "123")
	if err != nil || project.ProjectID != "123" || project.ProjectNumber != "123" {
		t.Fatalf("ResolveProjectNumberForExecution() = %#v, %v", project, err)
	}
}

func TestFirebaseServiceContextForExecutionRequiresStatelessToken(t *testing.T) {
	t.Setenv(env.GoogleAccessToken, "")
	ctx := core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy())
	_, err := FirebaseServiceContextForExecution(ctx, "demo")
	var authErr *core.AuthError
	if !errors.As(err, &authErr) || authErr.Kind != "configuration" || !strings.Contains(err.Error(), env.GoogleAccessToken) {
		t.Fatalf("FirebaseServiceContextForExecution error = %v", err)
	}

	stateful := core.WithExecutionPolicy(context.Background(), core.StatefulExecutionPolicy())
	got, err := FirebaseServiceContextForExecution(stateful, "demo")
	if err != nil || got != stateful {
		t.Fatalf("stateful context = %v, %v; want unchanged", got, err)
	}
}

func TestMatchProjectsForArgResolutionOrder(t *testing.T) {
	projects := []core.Project{
		{Name: "Production", ProjectID: "example-production-a1b2"},
		{Name: "example-production-a1b2", ProjectID: "name-collision"},
		{Name: "Production EU", ProjectID: "example-production-eu-c3d4"},
		{Name: "Staging", ProjectID: "example-staging-e5f6"},
	}

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{name: "exact id wins over name", query: "example-production-a1b2", want: []string{"example-production-a1b2"}},
		{name: "case-mismatched id does not match", query: "EXAMPLE-production-A1B2", want: nil},
		{name: "exact name", query: "Production", want: []string{"example-production-a1b2"}},
		{name: "case-mismatched name does not match", query: "production", want: nil},
		{name: "substring does not match", query: "stag", want: nil},
		{name: "missing", query: "unrelated", want: nil},
		{name: "empty", query: "  ", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := matchProjectsForArg(projects, tt.query)
			got := make([]string, len(matches))
			for i, project := range matches {
				got[i] = project.ProjectID
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("matchProjectsForArg(%q) = %#v, want %#v", tt.query, got, tt.want)
			}
		})
	}
}

func TestResolveProjectArgWithFilterUsesExactThenFilterPrecedence(t *testing.T) {
	projects := []core.Project{
		{Name: "Production", ProjectID: "production-a"},
		{Name: "Production EU", ProjectID: "production-eu"},
		{Name: "Staging", ProjectID: "staging-a"},
	}
	aliases := map[string]string{"prod": "production-a"}
	tests := []struct {
		query string
		want  string
	}{
		{query: "production-a", want: "production-a"},
		{query: "prod", want: "production-a"},
		{query: "Staging", want: "staging-a"},
		{query: "stg", want: "staging-a"},
		{query: "^stag", want: "staging-a"},
		{query: "/tion-e", want: "production-eu"},
		{query: "=production-a", want: "production-a"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			cmd := &cobra.Command{Use: "show"}
			project, err := resolveProjectArgWithAliases(cmd, projects, tt.query, aliases)
			if err != nil || project.ProjectID != tt.want {
				t.Fatalf("resolveProjectArgWithAliases(%q) = %#v, %v; want %q", tt.query, project, err, tt.want)
			}
		})
	}
}

func TestResolveProjectArgWithFilterReportsOnlyMatchingVariants(t *testing.T) {
	projects := []core.Project{
		{Name: "Production", ProjectID: "production-a"},
		{Name: "Preview", ProjectID: "preview-a"},
		{Name: "Staging", ProjectID: "staging-a"},
	}
	var output bytes.Buffer
	cmd := &cobra.Command{Use: "show"}
	cmd.SetOut(&output)
	_, err := resolveProjectArgWithAliases(cmd, projects, "pr", nil)
	var selection *ProjectResolutionError
	if !errors.As(err, &selection) || selection.Kind != "ambiguous" || len(selection.Candidates) != 2 {
		t.Fatalf("resolution error = %#v", err)
	}
	if !strings.Contains(output.String(), "production-a") || !strings.Contains(output.String(), "preview-a") || strings.Contains(output.String(), "staging-a") {
		t.Fatalf("matching variants output =\n%s", output.String())
	}
}

func TestResolveProjectArgWithFilterDoesNotFilterAliases(t *testing.T) {
	cmd := &cobra.Command{Use: "show"}
	cmd.SetOut(&bytes.Buffer{})
	_, err := resolveProjectArgWithAliases(cmd, []core.Project{{Name: "Production", ProjectID: "production-a"}}, "/shortcut", map[string]string{"shortcut": "production-a"})
	var selection *ProjectResolutionError
	if !errors.As(err, &selection) || selection.Kind != "not_found" {
		t.Fatalf("alias filter error = %#v", err)
	}
}

func TestResolveCachedProjectArgUsesRepositoryAliasPrecedence(t *testing.T) {
	root := setupProjectAliasResolutionTest(t, `[projects.aliases]
prod = "acme-production-42"
release = "acme-production-42"
`)
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	projects := []config.Project{
		{Name: "Alias name collision", ProjectID: "prod", AuthID: "main"},
		{Name: "release", ProjectID: "display-name-collision", AuthID: "main"},
		{Name: "Production", ProjectID: "acme-production-42", AuthID: "main"},
	}
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "show"}
	cmd.SetOut(&bytes.Buffer{})

	project, err := ResolveCachedProjectArg(cmd, "prod")
	if err != nil || project.ProjectID != "prod" {
		t.Fatalf("exact ID precedence = %#v, %v", project, err)
	}
	project, err = ResolveCachedProjectArg(cmd, "release")
	if err != nil || project.ProjectID != "acme-production-42" {
		t.Fatalf("alias precedence = %#v, %v", project, err)
	}
	project, err = ResolveCachedProjectArg(cmd, "RELEASE")
	if err != nil || project.ProjectID != "display-name-collision" {
		t.Fatalf("filter fallback after exact alias miss = %#v, %v", project, err)
	}

	if _, err := os.Stat(filepath.Join(root, config.LocalConfigFileName)); err != nil {
		t.Fatal(err)
	}
}

func TestResolveCachedProjectTargetArgUsesAliasAndConfiguredPrimary(t *testing.T) {
	setupProjectAliasResolutionTest(t, "[projects.aliases]\nprod = \"acme-production-42\"\n")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	project := config.Project{
		Name: "Production", ProjectID: "acme-production-42", AuthID: "main",
		Templates: []rctarget.Kind{rctarget.Client, rctarget.Server}, PrimaryTemplate: rctarget.Server,
	}
	if err := config.SaveProjects([]config.Project{project}, time.Now()); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "export"}
	cmd.SetOut(&bytes.Buffer{})
	for query, want := range map[string]string{
		"prod":                       "server@acme-production-42",
		"client@prod":                "acme-production-42",
		"server@prod":                "server@acme-production-42",
		"client@/production":         "acme-production-42",
		"server@^acme":               "server@acme-production-42",
		"server@=ACME-PRODUCTION-42": "server@acme-production-42",
		"server@production":          "server@acme-production-42",
	} {
		got, err := ResolveCachedProjectTargetArg(cmd, query)
		if err != nil || got.ProjectID != want {
			t.Fatalf("ResolveCachedProjectTargetArg(%q) = %#v, %v; want %q", query, got, err, want)
		}
	}
}

func TestResolveCachedProjectArgUsesFirebaseRCAlias(t *testing.T) {
	root := setupProjectAliasResolutionTest(t, "")
	if err := os.WriteFile(filepath.Join(root, config.FirebaseConfigFileName), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, config.FirebaseRCFileName), []byte(`{"projects":{"prod":"acme-production-42"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Production", ProjectID: "acme-production-42", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "show"}
	cmd.SetOut(&bytes.Buffer{})

	project, err := ResolveCachedProjectArg(cmd, "prod")
	if err != nil || project.ProjectID != "acme-production-42" {
		t.Fatalf("Firebase RC alias resolution = %#v, %v", project, err)
	}
}

func TestResolveCachedProjectArgReportsDanglingAlias(t *testing.T) {
	setupProjectAliasResolutionTest(t, "[projects.aliases]\nprod = \"acme-production-42\"\n")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Other", ProjectID: "other-project", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "show"}
	cmd.SetOut(&bytes.Buffer{})
	_, err := ResolveCachedProjectArg(cmd, "prod")
	if err == nil || !strings.Contains(err.Error(), `alias "prod"`) || !strings.Contains(err.Error(), "acme-production-42") || !strings.Contains(err.Error(), `profile "default"`) {
		t.Fatalf("dangling alias error = %v", err)
	}
	var selection *ProjectResolutionError
	if !errors.As(err, &selection) || selection.Resource != "project" || selection.Kind != "not_found" || selection.Query != "prod" || len(selection.Candidates) != 1 || selection.Candidates[0].ID != "other-project" {
		t.Fatalf("dangling alias selection = %#v", selection)
	}
}

func TestProjectChoiceTableIncludesAliasesAndFitsNarrowTerminal(t *testing.T) {
	setupProjectAliasResolutionTest(t, "[projects.aliases]\nproduction = \"acme-production-42\"\n")
	output := renderProjectsChoiceTableAtWidth([]core.Project{{
		Name: strings.Repeat("Production Project ", 3), ProjectID: "acme-production-42",
	}}, 44)
	if !strings.Contains(output, "Aliases") && !strings.Contains(output, "Alias…") {
		t.Fatalf("choice table lacks aliases header: %q", output)
	}
	for index, line := range strings.Split(output, "\n") {
		if width := lipgloss.Width(line); width > 44 {
			t.Fatalf("line %d width = %d, want <= 44:\n%s", index, width, output)
		}
	}
}

func setupProjectAliasResolutionTest(t *testing.T, localConfig string) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, config.LocalConfigFileName), []byte(localConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	config.SetLocalConfigDisabled(false)
	t.Cleanup(func() {
		config.SetLocalConfigDisabled(false)
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	return root
}

func TestResolveCachedProjectTargetArgUsesConfiguredPrimary(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	project := config.Project{
		Name:            "Demo",
		ProjectID:       "demo",
		AuthID:          "main",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Server,
	}
	if err := config.SaveProjects([]config.Project{project}, time.Now()); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "diff"}
	cmd.SetOut(&bytes.Buffer{})
	for input, want := range map[string]string{
		"demo":        "server@demo",
		"client@demo": "demo",
		"server@demo": "server@demo",
	} {
		got, err := ResolveCachedProjectTargetArg(cmd, input)
		if err != nil || got.ProjectID != want {
			t.Fatalf("ResolveCachedProjectTargetArg(%q) = %#v, %v; want %q", input, got, err, want)
		}
	}
}

func TestResolveCachedProjectArgUsesLocalRegistry(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "diff"}
	cmd.SetOut(&bytes.Buffer{})
	project, err := ResolveCachedProjectArg(cmd, "demo")
	if err != nil {
		t.Fatalf("resolve cached project = %v", err)
	}
	if project.ProjectID != "demo" {
		t.Fatalf("cached project = %#v", project)
	}
}

func TestResolveCachedProjectTargetArgCanonicalizesClientAndServer(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	cmd := &cobra.Command{Use: "diff"}
	cmd.SetOut(&bytes.Buffer{})
	for input, want := range map[string]string{
		"demo":        "demo",
		"client@demo": "demo",
		"server@Demo": "server@demo",
	} {
		project, err := ResolveCachedProjectTargetArg(cmd, input)
		if err != nil {
			t.Fatalf("ResolveCachedProjectTargetArg(%q) = %v", input, err)
		}
		if project.ProjectID != want {
			t.Fatalf("ResolveCachedProjectTargetArg(%q).ProjectID = %q, want %q", input, project.ProjectID, want)
		}
	}
}
