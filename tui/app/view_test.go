package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/components/setup"
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
	if !strings.Contains(got, "¹Projects") || !strings.Contains(got, "²Parameters") {
		t.Fatalf("view does not render main panels after WindowSizeMsg:\n%s", got)
	}
}

func TestPopupWindowDimsBasePanelBorders(t *testing.T) {
	m := viewTestModel(90, 24, panels.Parameters)
	m.boolPicker = m.boolPicker.Open(10, 5, true)

	if !m.popupWindowOpen() {
		t.Fatal("popupWindowOpen() = false with boolean picker open")
	}

	wantTop := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.projects.ViewWithBorder(false, false),
		m.parameters.ViewWithBorder(false, false),
	)
	want := lipgloss.JoinVertical(
		lipgloss.Left,
		wantTop,
		m.logs.ViewWithBorder(false, false),
		m.helpView(),
	)
	if got := m.baseView(); got != want {
		t.Fatal("base view does not render inactive panel borders while popup is open")
	}

	m.boolPicker = m.boolPicker.Close()
	if m.popupWindowOpen() {
		t.Fatal("popupWindowOpen() = true after boolean picker closed")
	}
	wantTop = lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.projects.ViewWithBorder(false, false),
		m.parameters.ViewWithBorder(true, true),
	)
	want = lipgloss.JoinVertical(
		lipgloss.Left,
		wantTop,
		m.logs.ViewWithBorder(false, false),
		m.helpView(),
	)
	if got := m.baseView(); got != want {
		t.Fatal("base view does not restore only the focused panel border after popup closes")
	}
}

func TestAccountsPopupOverlaysWorkspace(t *testing.T) {
	svc := newRenameTestService(t)
	m := viewTestModel(90, 24, panels.Projects)
	m.svc = svc
	m.setup = setup.New(svc)
	var cmd tea.Cmd
	m.setup, cmd = m.setup.OpenAccounts()
	if cmd == nil || !m.setup.IsPopup() {
		t.Fatalf("OpenAccounts = cmd:%v popup:%v", cmd != nil, m.setup.IsPopup())
	}

	view := testutil.NormalizeViewSnapshot(m.View().Content)
	if !strings.Contains(view, "¹Projects") || !strings.Contains(view, "Starting fbrcm") {
		t.Fatalf("popup did not retain workspace beneath setup:\n%s", view)
	}
	if m.View().MouseMode != tea.MouseModeNone {
		t.Fatalf("popup mouse mode = %v, want none", m.View().MouseMode)
	}
}

func TestDialogEnablesMouseWhenLogsWereActive(t *testing.T) {
	m := viewTestModel(100, 30, panels.Logs)
	m.openErrorDialog("Import Failed", core.Project{Name: "Demo", ProjectID: "demo"}, "validation failed")
	if got := m.View().MouseMode; got != tea.MouseModeAllMotion {
		t.Fatalf("dialog mouse mode = %v, want all motion", got)
	}
}

func TestGlobalProfilesShortcutOpensPopupDirectly(t *testing.T) {
	svc := newRenameTestService(t)
	m := viewTestModel(90, 24, panels.Projects)
	m.svc = svc
	m.setup = setup.New(svc)

	next, cmd, handled := m.updateGlobalKeyMessage("ctrl+p")
	if !handled || cmd == nil || !next.setup.IsPopup() {
		t.Fatalf("global ctrl+p = handled:%v cmd:%v popup:%v", handled, cmd != nil, next.setup.IsPopup())
	}
}

func TestPopupWindowDimsDetailsPanelBorder(t *testing.T) {
	m := viewTestModel(90, 24, panels.Details)
	m.detailsVisible = true
	m.boolPicker = m.boolPicker.Open(10, 5, true)

	if got, want := m.detailsPanelView(), m.details.ViewWithBorder(false); got != want {
		t.Fatal("details panel does not render an inactive border while popup is open")
	}
}

const minSizeViewSnapshot = `
 Terminal too small
 Minimum size 80x20`

const baseEmptyAppViewSnapshot = `── ¹Projects ─────── | ─╮╭─ ²Parameters ── ³Conditions ── ⁴History ──────────────────────╮
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
── ⁰Logs ──DEBU─INFO─WARN─ERRO─FATA─SLNT────────────────────────────────────────── live ──
No logs yet.




──────────────────────────────────────────────────────────────────────────────────────────
q quit • ? help • c collapse • enter select • space mark • o open • u update …`

const logsActiveViewSnapshot = `── ¹Projects ─────── | ─╮╭─ ²Parameters ── ³Conditions ── ⁴History ──────────────────────╮
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
── ⁰Logs ──DEBU─INFO─WARN─ERRO─FATA─SLNT────────────────────────────────────────── live ──
No logs yet.




──────────────────────────────────────────────────────────────────────────────────────────
q quit • ? help • c collapse • [/] level • -/_/=/+ resize`

const offlineBadgeViewSnapshot = `── ¹Projects ─────── | ─╮╭─ ²Parameters ── ³Conditions ── ⁴History ──────────────────────╮
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
── ⁰Logs ──DEBU─INFO─WARN─ERRO─FATA─SLNT────────────────────────────────────────── live ──
No logs yet.




──────────────────────────────────────────────────────────────────────────────────────────
q quit • ? help • z maximize • r rename • e edit • a new • c duplicate • m move … OFFLINE`

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
