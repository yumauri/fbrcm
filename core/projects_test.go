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
