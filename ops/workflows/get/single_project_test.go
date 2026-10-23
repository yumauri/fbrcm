package get

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func TestGetOmitsProjectColumnForOneVisibleTarget(t *testing.T) {
	for _, test := range []struct {
		name        string
		projectIDs  []string
		whitelist   bool
		blacklist   bool
		bothKinds   bool
		args        []string
		wantProject bool
	}{
		{name: "only configured project", projectIDs: []string{"alpha"}},
		{name: "only whitelisted project", projectIDs: []string{"alpha", "beta"}, whitelist: true},
		{name: "only non-blacklisted project", projectIDs: []string{"alpha", "beta"}, blacklist: true},
		{name: "one fuzzy match", projectIDs: []string{"alpha", "beta"}, args: []string{"--project", "^alpha"}},
		{name: "one exact match", projectIDs: []string{"alpha", "beta"}, args: []string{"--project", "=alpha"}},
		{name: "several projects", projectIDs: []string{"alpha", "beta"}, wantProject: true},
		{name: "client and server targets", projectIDs: []string{"alpha"}, bothKinds: true, wantProject: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(env.ConfigDir, t.TempDir())
			t.Setenv(env.CacheDir, t.TempDir())
			t.Setenv(env.NoColor, "1")
			t.Chdir(t.TempDir())
			if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
				t.Fatal(err)
			}
			projects := make([]config.Project, 0, len(test.projectIDs))
			for _, id := range test.projectIDs {
				project := config.Project{ProjectID: id, Name: id + " workspace", AuthID: "main"}
				if test.bothKinds {
					project.Templates = []rctarget.Kind{rctarget.Client, rctarget.Server}
				}
				projects = append(projects, project)
				for _, kind := range project.TemplateKinds() {
					if err := config.SaveParametersCache(project.TemplateTarget(kind).ProjectID, &config.ParametersCache{
						CachedAt: time.Now(), RemoteConfig: json.RawMessage(`{"parameters":{"flag":{"defaultValue":{"value":"a moderately long value for narrow terminals"}}}}`),
					}); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := config.SaveProjects(projects, time.Now()); err != nil {
				t.Fatal(err)
			}
			policy := config.DefaultProjectVisibility()
			policy.Whitelist.Enabled = test.whitelist
			policy.Whitelist.ProjectIDs = []string{"alpha"}
			if test.blacklist {
				policy.Blacklist.ProjectIDs = []string{"beta"}
			}
			if err := config.SaveProjectVisibility(policy); err != nil {
				t.Fatal(err)
			}
			svc, err := core.NewService(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			for _, width := range []string{"200", "60"} {
				t.Setenv("COLUMNS", width)
				cmd := New(svc)
				var out bytes.Buffer
				cmd.SetOut(&out)
				cmd.SetErr(&bytes.Buffer{})
				cmd.SetIn(strings.NewReader(""))
				cmd.SetArgs(append([]string{"--cached"}, test.args...))
				if err := cmd.Execute(); err != nil {
					t.Fatal(err)
				}
				if strings.Contains(out.String(), "Project") != test.wantProject {
					t.Fatalf("width %s project column = %t, want %t:\n%s", width, !test.wantProject, test.wantProject, out.String())
				}
				if !test.wantProject && strings.Contains(out.String(), "alpha workspace") {
					t.Fatalf("single-project output still contains project identity:\n%s", out.String())
				}
				if width == "60" {
					for line := range strings.SplitSeq(out.String(), "\n") {
						if lipgloss.Width(line) > 60 {
							t.Fatalf("narrow table overflows: %q", line)
						}
					}
				}
			}
			// Project identity remains present in the machine output.
			cmd := New(svc)
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(strings.NewReader(""))
			cmd.SetArgs(append([]string{"--cached", "--json"}, test.args...))
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			var rows []parameterRowJSON
			if err := json.Unmarshal(out.Bytes(), &rows); err != nil || len(rows) == 0 || rows[0].ProjectID != "alpha" || rows[0].Project != "alpha workspace" {
				t.Fatalf("JSON identity = %s, %v", out.String(), err)
			}
		})
	}
}
