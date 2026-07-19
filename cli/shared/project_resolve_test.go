package shared

import (
	"bytes"
	"slices"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
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
