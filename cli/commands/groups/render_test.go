package groups

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
)

func TestRenderGroupsTableUsesNaturalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderGroupsTableAtWidth([]projectGroup{{Group: core.ParametersGroup{Key: "checkout", Description: "Checkout flags", Parameters: []core.ParametersEntry{{Key: "enabled"}}}}}, false, 120)
	for _, want := range []string{"Name", "Parameters", "Description", "checkout", "Checkout flags"} {
		if !strings.Contains(got, want) {
			t.Fatalf("table missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("NO_COLOR table contains ANSI: %q", got)
	}
}

func TestRenderGroupsTableCropsDescriptionOnNarrowTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	const width = 45
	got := renderGroupsTableAtWidth([]projectGroup{{Group: core.ParametersGroup{Key: "checkout", Description: "A deliberately long group description that cannot fit"}}}, false, width)
	if !strings.Contains(got, "…") {
		t.Fatalf("description was not cropped:\n%s", got)
	}
	for index, line := range strings.Split(ansi.Strip(got), "\n") {
		if gotWidth := lipgloss.Width(line); gotWidth > width {
			t.Fatalf("line %d width = %d, want <= %d:\n%s", index, gotWidth, width, got)
		}
	}
}

func TestRenderGroupsTableIncludesProjectForMultipleProjects(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderGroupsTableAtWidth([]projectGroup{
		{Project: core.Project{ProjectID: "project-a"}, Group: core.ParametersGroup{Key: "shared"}},
		{Project: core.Project{ProjectID: "project-b"}, Group: core.ParametersGroup{Key: "shared"}},
	}, true, 100)
	for _, want := range []string{"Project", "project-a", "project-b", "shared"} {
		if !strings.Contains(got, want) {
			t.Fatalf("multi-project table missing %q:\n%s", want, got)
		}
	}
}

func TestRenderGroupsTableCropsMultiProjectFlexibleColumns(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	const width = 52
	got := renderGroupsTableAtWidth([]projectGroup{{
		Project: core.Project{ProjectID: "an-extremely-long-project-identifier"},
		Group:   core.ParametersGroup{Key: "an-extremely-long-group-name", Description: "A long description that is cropped first"},
	}}, true, width)
	if !strings.Contains(got, "…") {
		t.Fatalf("multi-project table did not crop flexible columns:\n%s", got)
	}
	for index, line := range strings.Split(ansi.Strip(got), "\n") {
		if gotWidth := lipgloss.Width(line); gotWidth > width {
			t.Fatalf("line %d width = %d, want <= %d:\n%s", index, gotWidth, width, got)
		}
	}
}
