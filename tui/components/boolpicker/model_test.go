package boolpicker

import (
	"strings"
	"testing"
	"time"

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

func TestBoolpickerRowsSelectAndReportDoubleClick(t *testing.T) {
	m := New().Open(10, 5, true)
	x, y := m.Position()
	now := time.Unix(100, 0)
	if double, hit := m.SelectAt(x+1, y+2, now); !hit || double {
		t.Fatalf("single click = hit:%v double:%v", hit, double)
	}
	if value, _ := m.Current(); value {
		t.Fatal("single click did not select false")
	}
	x, y = m.Position()
	if double, hit := m.SelectAt(x+1, y+2, now.Add(time.Millisecond)); !hit || !double {
		t.Fatalf("second click = hit:%v double:%v", hit, double)
	}
}
