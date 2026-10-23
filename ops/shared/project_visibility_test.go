package shared

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	charmlog "charm.land/log/v2"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func TestProjectSelectionReadsVisibilityOnce(t *testing.T) {
	setupProjectAliasResolutionTest(t, "")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	policy := config.DefaultProjectVisibility()
	policy.Blacklist.ProjectIDs = []string{"alpha"}
	if err := config.SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	t.Setenv(env.LogLevel, "debug")
	var logs bytes.Buffer
	previousLevel := corelog.CurrentLevel()
	corelog.ConfigureCLIOutput(&logs, &logs)
	corelog.InitWithDefault(corelog.ModeCLI, charmlog.DebugLevel)
	t.Cleanup(func() {
		corelog.ConfigureCLIOutput(os.Stderr, os.Stderr)
		corelog.SetLevel(previousLevel)
	})
	projects := []core.Project{{ProjectID: "alpha", ProjectNumber: "100"}, {ProjectID: "beta", ProjectNumber: "200"}}
	cmd := &cobra.Command{}
	ctx := context.Background()
	cmd.SetContext(ctx)
	for _, tc := range []struct {
		name   string
		run    func() error
		hidden bool
	}{
		{"bulk exact", func() error { _, err := VisibleProjectSelection(cmd, projects, []string{"=beta", "=beta"}); return err }, false},
		{"bulk hidden", func() error { _, err := VisibleProjectSelection(cmd, projects, []string{"=alpha"}); return err }, true},
		{"positional ID", func() error { _, err := visibleProjectArguments(cmd, projects, "beta", nil); return err }, false},
		{"positional alias", func() error {
			_, err := visibleProjectArguments(cmd, projects, "prod", map[string]string{"prod": "beta"})
			return err
		}, false},
		{"positional hidden", func() error { _, err := visibleProjectArguments(cmd, projects, "alpha", nil); return err }, true},
		{"project number", func() error { _, err := resolveProjectNumber(cmd, projects, "200"); return err }, false},
		{"hidden project number", func() error { _, err := resolveProjectNumber(cmd, projects, "100"); return err }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs.Reset()
			err := tc.run()
			var hidden *core.ProjectHiddenError
			if (err != nil) != tc.hidden || errors.As(err, &hidden) != tc.hidden {
				t.Fatalf("error = %v, want hidden %v", err, tc.hidden)
			}
			for _, record := range []string{"read project visibility config", "loaded project visibility config"} {
				if count := strings.Count(logs.String(), record); count != 1 {
					t.Errorf("%q count = %d, want 1\n%s", record, count, logs.String())
				}
			}
			if cmd.Context() != ctx {
				t.Fatal("selection retained its snapshot for later operations")
			}
		})
	}
}

func TestProjectVisibilityResolutionAndExactPrecedence(t *testing.T) {
	setupProjectAliasResolutionTest(t, "")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	policy := config.DefaultProjectVisibility()
	policy.Blacklist.ProjectIDs = []string{"alpha"}
	if err := config.SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	projects := []core.Project{{ProjectID: "alpha", Name: "Beta", ProjectNumber: "100"}, {ProjectID: "beta", Name: "Beta", ProjectNumber: "200"}}
	cmd := &cobra.Command{Use: "show"}
	cmd.SetContext(context.Background())
	aliases := map[string]string{"prod": "alpha", "beta": "alpha"}
	for _, query := range []string{"alpha", "prod", "=alpha"} {
		_, err := resolveProjectArgWithAliases(cmd, projects, query, aliases)
		if _, ok := errors.AsType[*core.ProjectHiddenError](err); !ok {
			t.Errorf("query %q error = %v, want project.hidden", query, err)
		}
	}
	// Exact physical ID still takes precedence over an alias or hidden name.
	project, err := resolveProjectArgWithAliases(cmd, projects, "beta", aliases)
	if err != nil || project.ProjectID != "beta" {
		t.Fatalf("exact ID = %#v, %v", project, err)
	}
	_, err = resolveProjectNumber(cmd, projects, "100")
	var hidden *core.ProjectHiddenError
	if !errors.As(err, &hidden) {
		t.Fatalf("App ID project number bypassed policy: %v", err)
	}
	visible, err := VisibleProjectSelection(cmd, projects, nil)
	if err != nil || len(visible) != 1 {
		t.Fatalf("bulk selection = %#v, %v", visible, err)
	}
	for i := range projects {
		projects[i].AuthID = "main"
	}
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	_, err = ResolveCachedProjectTargetArg(cmd, "server@alpha")
	if !errors.As(err, &hidden) {
		t.Fatalf("cached target bypassed policy: %v", err)
	}
	_, err = ResolveCachedProjectNumber(cmd, "100")
	if !errors.As(err, &hidden) {
		t.Fatalf("cached App ID bypassed policy: %v", err)
	}
	cmd.SetContext(core.WithHiddenProjects(context.Background(), true))
	project, err = resolveProjectArgWithAliases(cmd, projects, "prod", aliases)
	if err != nil || project.ProjectID != "alpha" {
		t.Fatalf("override = %#v, %v", project, err)
	}
}

func TestVisibilityChoiceTableNaturalAndNarrowWidths(t *testing.T) {
	setupProjectAliasResolutionTest(t, "")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NO_COLOR", "1")
	projects := []core.Project{{Name: "Unknown project (blacklisted)", ProjectID: "a-very-long-project-id"}}
	natural := renderProjectsChoiceTableAtWidth(projects, 200)
	if !strings.Contains(natural, projects[0].ProjectID) || !strings.Contains(natural, projects[0].Name) || strings.Contains(natural, "\x1b[") {
		t.Fatalf("natural table = %q", natural)
	}
	narrow := renderProjectsChoiceTableAtWidth(projects, 35)
	if !strings.Contains(narrow, "…") {
		t.Fatalf("narrow table did not crop: %q", narrow)
	}
	for line := range strings.SplitSeq(narrow, "\n") {
		if lipgloss.Width(line) > 35 {
			t.Fatalf("narrow table overflow: %q", line)
		}
	}
}
