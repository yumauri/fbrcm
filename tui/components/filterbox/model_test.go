package filterbox

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestFilterboxViewActive(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New()
	m.Activate(filter.ModeFuzzy)
	m.input.SetValue("login")

	lines := m.View(40, true, 3)
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(lines))
	}
	got := testutil.NormalizeViewSnapshot(strings.Join(lines, "\n"))
	if !strings.Contains(got, "~") || !strings.Contains(got, "login") || !strings.Contains(got, "3") {
		t.Fatalf("view = %q", got)
	}
}

func TestFilterboxClearAndBlur(t *testing.T) {
	m := New()
	m.Activate(filter.ModeIncludes)
	m.input.SetValue("x")
	m.ClearAndBlur()
	if m.Visible() {
		t.Fatal("filter should be hidden after clear")
	}
}

func TestFilterboxPasteSetsValue(t *testing.T) {
	m := New()
	m.Activate(filter.ModeExact)
	updated, _ := m.Update(tea.PasteMsg{Content: "demo"})
	if updated.Value() != "demo" {
		t.Fatalf("value = %q, want demo", updated.Value())
	}
}
