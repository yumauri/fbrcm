package versions

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRenderVersionsTablePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	table := renderVersionsTableAtWidth([]core.RemoteConfigVersionEntry{{
		RemoteConfigVersion: firebase.RemoteConfigVersion{VersionNumber: "42", UpdateTime: "2026-07-11T12:10:11Z", UpdateUser: firebase.RemoteConfigUser{Email: "a@example.com"}, UpdateOrigin: "REST_API", UpdateType: "ROLLBACK", ChangeNote: "known good"},
		Current:             true, Cached: true,
	}}, false, 200)
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

func TestRenderVersionsTableCropsChangeNoteOnNarrowTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	output := renderVersionsTableAtWidth([]core.RemoteConfigVersionEntry{{
		RemoteConfigVersion: firebase.RemoteConfigVersion{
			VersionNumber: "42",
			UpdateTime:    "2026-07-11T12:10:11Z",
			UpdateUser:    firebase.RemoteConfigUser{Email: "release-manager@example.com"},
			ChangeNote:    "Enable checkout version two for the production environment",
		},
	}}, false, 80)
	for index, line := range strings.Split(output, "\n") {
		if got := lipgloss.Width(line); got > 80 {
			t.Fatalf("line %d width = %d, want <= 80:\n%s", index, got, output)
		}
	}
	if !strings.Contains(output, "…") {
		t.Fatalf("narrow table did not ellipsize flexible content:\n%s", output)
	}
}
