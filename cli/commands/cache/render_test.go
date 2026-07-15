package cache

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHumanSize(t *testing.T) {
	cases := []struct {
		size int64
		want string
	}{
		{size: 12, want: "12 B "},
		{size: 1536, want: "1.5 KB"},
		{size: 12 * 1024, want: "12 KB"},
		{size: 5*1024*1024 + 512*1024, want: "5.5 MB"},
		{size: 12 * 1024 * 1024, want: "12 MB"},
	}

	for _, tc := range cases {
		if got := humanSize(tc.size); got != tc.want {
			t.Fatalf("humanSize(%d) = %q, want %q", tc.size, got, tc.want)
		}
	}
}

func TestSortCacheEntries(t *testing.T) {
	entries := []cacheEntry{
		{ProjectID: "beta", Version: "1"},
		{ProjectID: "Alpha", Version: "2"},
		{ProjectID: "alpha", Version: "3"},
		{ProjectID: "beta", Version: "4"},
	}

	sortCacheEntries(entries)

	got := make([]string, len(entries))
	for i, entry := range entries {
		got[i] = entry.ProjectID + ":" + entry.Version
	}
	want := []string{"alpha:3", "Alpha:2", "beta:4", "beta:1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted entries = %#v, want %#v", got, want)
	}
}

func TestTotalCacheSize(t *testing.T) {
	got := totalCacheSize([]cacheEntry{{Size: 10}, {Size: 20}, {Size: 5}})
	if got != 35 {
		t.Fatalf("totalCacheSize = %d, want 35", got)
	}
}

func TestRenderCacheTablePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cachedAt := time.Date(2026, 6, 14, 9, 10, 11, 0, time.UTC)

	table := renderCacheTable([]cacheEntry{
		{ProjectID: "project-a", Project: "Project A", Version: "42", Size: 1536, CachedAt: &cachedAt},
		{ProjectID: "project-b", Project: "Project B", Version: "43", Size: 10},
	})

	for _, want := range []string{"Project ID", "project-a", "Project A", "42", "1.5 KB", "project-b", "43"} {
		if !strings.Contains(table, want) {
			t.Fatalf("renderCacheTable = %q, want substring %q", table, want)
		}
	}
}
