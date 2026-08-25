package projects

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/styles"
	"github.com/yumauri/fbrcm/tui/testutil"
)

// parityTestModel builds a representative projects panel used to lock in
// rendered output before view_render splits.
func parityTestModel() Model {
	return loadedProjectsModel()
}

func TestRefreshThemeRebuildsCachedProjectRows(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := parityTestModel()
	before := m.View(true)
	corestyles.ApplyPalette(corestyles.Palette{corestyles.TokenSelection: "1"})
	t.Cleanup(corestyles.ResetPalette)
	want := projectStylePrefix(itemStyle.Inherit(styles.ProjectStateStyle(true, false))) + "A"
	if strings.Contains(before, want) {
		t.Fatal("pre-theme project row unexpectedly uses preview palette")
	}
	if stale := m.View(true); strings.Contains(stale, want) {
		t.Fatal("cached project row changed without a viewport refresh")
	}
	m = m.RefreshTheme()
	if got := m.View(true); !strings.Contains(got, want) {
		t.Fatalf("refreshed project row does not use preview selection color:\n%s", got)
	}
}

func projectStylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, _ := strings.Cut(rendered, "x")
	return prefix
}

func TestProjectsViewSnapshot(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := testutil.NormalizeViewSnapshot(parityTestModel().View(true))
	if got != projectsViewSnapshot {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, projectsViewSnapshot)
	}
}

const projectsViewSnapshot = `── ¹Projects ────────────── 3 ─╮
 Alpha Project                 │
  alpha                        │
                               │
 Beta Project                  │
  beta                         │
                               │
 Gamma Project                 │
  gamma                        │
                               │
                               │
───────────────────────────────╯`
