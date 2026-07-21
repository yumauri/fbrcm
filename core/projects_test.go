package core

import (
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
)

func TestMergeProjectsReturnsSortedProjects(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)

	got := mergeProjects(nil, []config.Project{
		{Name: "Beta", ProjectID: "beta-2", DiscoveredBy: []string{"main"}},
		{Name: "TIGER", ProjectID: "tiger", DiscoveredBy: []string{"main"}},
		{Name: "Alpha", ProjectID: "alpha", DiscoveredBy: []string{"main"}},
		{Name: "Tango", ProjectID: "tango", DiscoveredBy: []string{"main"}},
		{Name: "Beta", ProjectID: "beta-1", DiscoveredBy: []string{"main"}},
	}, "main", []string{"main"}, "", now)

	wantIDs := []string{"alpha", "beta-1", "beta-2", "tango", "tiger"}
	if len(got) != len(wantIDs) {
		t.Fatalf("projects length = %d, want %d", len(got), len(wantIDs))
	}
	for i, want := range wantIDs {
		if got[i].ProjectID != want {
			t.Fatalf("project[%d] = %q, want %q; got order %+v", i, got[i].ProjectID, want, got)
		}
	}
}

func TestMergeProjectsDisablesProjectMissingFromFullDiscovery(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	existing := []config.Project{{
		Name: "Missing", ProjectID: "missing", AuthID: "main", DiscoveredBy: []string{"main"}, SyncedAt: "old",
	}}

	got := mergeProjects(existing, nil, "main", []string{"main"}, "", now)
	if len(got) != 1 || !got[0].Disabled {
		t.Fatalf("merged = %+v, want retained disabled project", got)
	}
	if len(got[0].DiscoveredBy) != 0 || got[0].AuthID != "main" {
		t.Fatalf("disabled project = %+v, want cleared discovery and retained auth", got[0])
	}
	if got[0].SyncedAt != now.Format(time.RFC3339) {
		t.Fatalf("SyncedAt = %q, want %q", got[0].SyncedAt, now.Format(time.RFC3339))
	}
}

func TestMergeProjectsRebindsAndEnablesProjectDiscoveredByAnotherAuth(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	existing := []config.Project{{
		Name: "Demo", ProjectID: "demo", AuthID: "main", Disabled: true,
	}}
	incoming := []config.Project{{
		Name: "Demo", ProjectID: "demo", DiscoveredBy: []string{"work"},
	}}

	got := mergeProjects(existing, incoming, "main", []string{"main", "work"}, "", now)
	if len(got) != 1 || got[0].Disabled || got[0].AuthID != "work" {
		t.Fatalf("merged = %+v, want enabled project bound to work", got)
	}
}

func TestMergeProjectsForAuthOnlyDisablesAssignedMissingProject(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	existing := []config.Project{
		{Name: "Main only", ProjectID: "main-only", AuthID: "main", DiscoveredBy: []string{"main"}},
		{Name: "Shared", ProjectID: "shared", AuthID: "main", DiscoveredBy: []string{"main", "work"}},
		{Name: "Work", ProjectID: "work", AuthID: "work", DiscoveredBy: []string{"work"}},
	}

	got := mergeProjects(existing, nil, "main", []string{"main", "work"}, "main", now)
	byID := make(map[string]config.Project, len(got))
	for _, project := range got {
		byID[project.ProjectID] = project
	}
	if !byID["main-only"].Disabled || byID["main-only"].AuthID != "main" {
		t.Fatalf("main-only = %+v, want disabled with retained auth", byID["main-only"])
	}
	if byID["shared"].Disabled || byID["shared"].AuthID != "work" {
		t.Fatalf("shared = %+v, want rebound to work", byID["shared"])
	}
	if byID["work"].Disabled || byID["work"].AuthID != "work" {
		t.Fatalf("work = %+v, want unchanged", byID["work"])
	}
}
