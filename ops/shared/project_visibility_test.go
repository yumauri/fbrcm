package shared

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
)

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
