package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestViewSnapshots(t *testing.T) {
	tests := []struct {
		name      string
		model     Model
		offline   bool
		mouseMode tea.MouseMode
		want      string
	}{
		{
			name:      "min size",
			model:     viewTestModel(20, 5, panels.Projects),
			mouseMode: tea.MouseModeNone,
			want:      minSizeViewSnapshot,
		},
		{
			name:      "base empty app",
			model:     viewTestModel(90, 24, panels.Projects),
			mouseMode: tea.MouseModeAllMotion,
			want:      baseEmptyAppViewSnapshot,
		},
		{
			name:      "logs active disables mouse",
			model:     viewTestModel(90, 24, panels.Logs),
			mouseMode: tea.MouseModeNone,
			want:      logsActiveViewSnapshot,
		},
		{
			name:      "offline badge",
			model:     viewTestModel(90, 24, panels.Parameters),
			offline:   true,
			mouseMode: tea.MouseModeAllMotion,
			want:      offlineBadgeViewSnapshot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firebase.SetOfflineMode(tt.offline)
			t.Cleanup(func() { firebase.SetOfflineMode(false) })

			view := tt.model.View()
			if view.MouseMode != tt.mouseMode {
				t.Fatalf("MouseMode = %v, want %v", view.MouseMode, tt.mouseMode)
			}
			got := testutil.NormalizeViewSnapshot(view.Content)
			if got != tt.want {
				t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, tt.want)
			}
		})
	}
}

func TestWindowSizeUpdateEnablesNormalView(t *testing.T) {
	model := New(nil)

	next, _ := model.Update(tea.WindowSizeMsg{Width: 90, Height: 24})
	model, ok := next.(Model)
	if !ok {
		t.Fatalf("model type = %T, want Model", next)
	}
	if model.width != 90 || model.height != 24 {
		t.Fatalf("size = %dx%d, want 90x24", model.width, model.height)
	}

	got := testutil.NormalizeViewSnapshot(model.View().Content)
	if strings.Contains(got, "Terminal too small") {
		t.Fatalf("view still renders min-size screen after WindowSizeMsg:\n%s", got)
	}
	if !strings.Contains(got, "[1] Projects") || !strings.Contains(got, "[2] Parameters") {
		t.Fatalf("view does not render main panels after WindowSizeMsg:\n%s", got)
	}
}

const minSizeViewSnapshot = `
 Terminal too small
 Minimum size 80x20`

const baseEmptyAppViewSnapshot = `── [1] Projects ──── | ─╮╭─ [2] Parameters ──────────────────────────────────────────────╮
 Loading projects...    ││Select project in Projects panel.                              │
                        ││                                                               │
                        ││Selected project will appear here immediately.                 │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
────────────────────────╯╰───────────────────────────────────────────────────────────────╯
── [0] Logs ──DEBU─INFO─WARN─ERRO─FATA─SLNT─────────────────────────────────────── live ──
No logs yet.




──────────────────────────────────────────────────────────────────────────────────────────
q quit • c collapse • enter select • space mark • o open • u update • ~/^///= filter`

const logsActiveViewSnapshot = `── [1] Projects ──── | ─╮╭─ [2] Parameters ──────────────────────────────────────────────╮
 Loading projects...    ││Select project in Projects panel.                              │
                        ││                                                               │
                        ││Selected project will appear here immediately.                 │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
────────────────────────╯╰───────────────────────────────────────────────────────────────╯
── [0] Logs ──DEBU─INFO─WARN─ERRO─FATA─SLNT─────────────────────────────────────── live ──
No logs yet.




──────────────────────────────────────────────────────────────────────────────────────────
q quit • c collapse • [/] level • -/_/=/+ resize`

const offlineBadgeViewSnapshot = `── [1] Projects ──── | ─╮╭─ [2] Parameters ──────────────────────────────────────────────╮
 Loading projects...    ││Select project in Projects panel.                              │
                        ││                                                               │
                        ││Selected project will appear here immediately.                 │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
                        ││                                                               │
────────────────────────╯╰───────────────────────────────────────────────────────────────╯
── [0] Logs ──DEBU─INFO─WARN─ERRO─FATA─SLNT─────────────────────────────────────── live ──
No logs yet.




──────────────────────────────────────────────────────────────────────────────────────────
q quit • z maximize • r rename • e edit • a new • c duplicate • m move •  /space  OFFLINE`

func viewTestModel(width, height int, active panels.ID) Model {
	m := New(nil)
	m.width = width
	m.height = height
	m.logsSized = true
	m.logsHeight = defaultLogsPanelHeight
	m.setActive(active)
	m.applyLayout()
	return m
}
