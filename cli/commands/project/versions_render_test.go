package project

import (
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRenderVersionsTablePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	table := renderVersionsTable([]core.RemoteConfigVersionEntry{{
		RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "42", UpdateTime: "2026-07-11T12:10:11Z", UpdateUser: firebase.RemoteConfigUser{Email: "a@example.com"}, UpdateOrigin: "REST_API", UpdateType: "ROLLBACK", Description: "known good"},
		Current:             true, Cached: true,
	}}, false)
	for _, want := range []string{"Version", "Published", "Updated By", "42", "current", "a@example.com", "REST API", "Rollback", "yes", "known good", "┌", "┘"} {
		if !strings.Contains(table, want) {
			t.Fatalf("renderVersionsTable = %q, want substring %q", table, want)
		}
	}
}

func TestRenderCachedVersionsTablePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	cachedAt := time.Date(2026, 7, 11, 12, 10, 11, 0, time.UTC)
	table := renderVersionsTable([]core.RemoteConfigVersionEntry{{RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "7"}, Cached: true, CachedAt: cachedAt, Size: 1536}}, true)
	for _, want := range []string{"Version", "Cached At", "Size", "7", "1.5 KB"} {
		if !strings.Contains(table, want) {
			t.Fatalf("renderVersionsTable = %q, want substring %q", table, want)
		}
	}
}
