package renameinput

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestRenameinputOpenAndView(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m, _ := New().Open(4, 6, 8, 20, "feature_flag")
	view := testutil.NormalizeViewSnapshot(m.View())
	if !strings.Contains(view, "feature_flag") {
		t.Fatalf("view = %q", view)
	}
	if m.Value() != "feature_flag" {
		t.Fatalf("value = %q", m.Value())
	}
}

func TestRenameinputCloseClearsValue(t *testing.T) {
	m, _ := New().Open(0, 0, 5, 10, "x")
	m = m.Close()
	if m.IsOpen() || m.Value() != "" {
		t.Fatalf("close = open=%v value=%q", m.IsOpen(), m.Value())
	}
}
