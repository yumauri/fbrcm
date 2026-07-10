package stringinput

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestStringinputOpenAndView(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m, _ := New().Open(2, 3, 10, 30, 80, 24, "hello", false, false)
	view := testutil.NormalizeViewSnapshot(m.View())
	if !strings.Contains(view, "hello") {
		t.Fatalf("view = %q, want hello", view)
	}
	if m.Value() != "hello" {
		t.Fatalf("value = %q", m.Value())
	}
}

func TestStringinputToggleExpanded(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m, _ := New().Open(2, 3, 10, 30, 80, 24, "line", false, false)
	expanded, _ := m.ToggleExpanded()
	if !expanded.IsExpanded() {
		t.Fatal("expected expanded mode")
	}
	if expanded.View() == "" {
		t.Fatal("expanded view should render")
	}
}
