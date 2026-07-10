package boolpicker

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestBoolpickerOpenAndMove(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := New().Open(10, 5, true)
	m.Move(1)

	got, ok := m.Current()
	if !ok || got {
		t.Fatalf("current = %v/%v, want false", got, ok)
	}

	view := testutil.NormalizeViewSnapshot(m.View())
	if !strings.Contains(view, "false") || !strings.Contains(view, "true") {
		t.Fatalf("view = %q", view)
	}
}

func TestBoolpickerClose(t *testing.T) {
	m := New().Open(0, 0, false).Close()
	if m.IsOpen() || m.View() != "" {
		t.Fatal("closed picker should not render")
	}
}
