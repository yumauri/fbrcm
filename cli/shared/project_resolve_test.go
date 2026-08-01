package shared

import (
	"bytes"
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

func TestMatchProjectsForArgResolutionOrder(t *testing.T) {
	projects := []core.Project{
		{Name: "Production", ProjectID: "setplex-production-a1b2"},
		{Name: "setplex-production-a1b2", ProjectID: "name-collision"},
		{Name: "Production EU", ProjectID: "setplex-production-eu-c3d4"},
		{Name: "Staging", ProjectID: "setplex-staging-e5f6"},
	}

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{name: "exact id wins over name", query: "SETplex-production-A1B2", want: []string{"setplex-production-a1b2"}},
		{name: "exact name precedes substring", query: "production", want: []string{"setplex-production-a1b2"}},
		{name: "single substring", query: "stag", want: []string{"setplex-staging-e5f6"}},
		{name: "ambiguous substring", query: "prod", want: []string{"setplex-production-a1b2", "name-collision", "setplex-production-eu-c3d4"}},
		{name: "fuzzy-only does not match", query: "stg", want: nil},
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
	project, err = ResolveCachedProjectArg(cmd, "RELEASE")
	if err != nil || project.ProjectID != "acme-production-42" {
		t.Fatalf("alias precedence = %#v, %v", project, err)
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
		"prod": "server@acme-production-42", "client@prod": "acme-production-42", "server@prod": "server@acme-production-42",
	} {
		got, err := ResolveCachedProjectTargetArg(cmd, query)
		if err != nil || got.ProjectID != want {
			t.Fatalf("ResolveCachedProjectTargetArg(%q) = %#v, %v; want %q", query, got, err, want)
		}
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
