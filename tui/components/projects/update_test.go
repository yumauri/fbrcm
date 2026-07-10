package projects

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestProjectsViewSnapshotLoaded(t *testing.T) {
	m := loadedProjectsModel()

	got := testutil.NormalizeViewSnapshot(m.View(true))
	if got != loadedProjectsSnapshot {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, loadedProjectsSnapshot)
	}
}

func TestProjectsKeyMovementAndSelection(t *testing.T) {
	m := loadedProjectsModel()

	m, _ = m.Update(keyPress(tea.KeyDown))
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}

	m, cmd := m.Update(keyPress(tea.KeyEnter))
	if cmd == nil {
		t.Fatalf("selection command is nil")
	}
	msgs := runBatch(t, cmd)
	selection := findMsg[messages.ProjectsSelectionChangedMsg](msgs)
	if len(selection.Projects) != 1 || selection.Projects[0].ProjectID != "beta" {
		t.Fatalf("selected projects = %+v, want beta", selection.Projects)
	}
	active := findMsg[messages.SetActivePanelMsg](msgs)
	if active.Panel != panels.Parameters {
		t.Fatalf("active panel = %v, want Parameters", active.Panel)
	}
}

func TestProjectsUpdateIgnoredWhileLoading(t *testing.T) {
	m := loadedProjectsModel()
	m.loading = true

	next, cmd := m.Update(keyPress('u'))

	if cmd != nil {
		t.Fatalf("update returned command while loading")
	}
	if !next.loading {
		t.Fatalf("loading changed to false")
	}
}

func TestProjectsFilterApplySelectsCurrentAndReleasesKeyboard(t *testing.T) {
	m := loadedProjectsModel()
	m, cmd := m.Update(keyText("/"))
	if cmd == nil {
		t.Fatalf("filter activation command is nil")
	}
	if !m.filter.Focused() {
		t.Fatalf("filter is not focused")
	}

	m, _ = m.Update(tea.PasteMsg{Content: "beta"})
	if len(m.projects) != 1 || m.projects[0].ProjectID != "beta" {
		t.Fatalf("filtered projects = %+v, want beta", m.projects)
	}

	m, cmd = m.Update(keyPress(tea.KeyEnter))
	msgs := runBatch(t, cmd)
	selection := findMsg[messages.ProjectsSelectionChangedMsg](msgs)
	if len(selection.Projects) != 1 || selection.Projects[0].ProjectID != "beta" {
		t.Fatalf("selected projects = %+v, want beta", selection.Projects)
	}
	capture := findMsg[messages.KeyboardCaptureMsg](msgs)
	if capture.Enabled {
		t.Fatalf("keyboard capture enabled, want false")
	}
	active := findMsg[messages.SetActivePanelMsg](msgs)
	if active.Panel != panels.Parameters {
		t.Fatalf("active panel = %v, want Parameters", active.Panel)
	}
}

func TestProjectsMouseWheelMovesCursor(t *testing.T) {
	m := loadedProjectsModel()

	m, _ = m.Update(tea.MouseWheelMsg{X: 2, Y: 2, Button: tea.MouseWheelDown})

	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
}

func loadedProjectsModel() Model {
	m := New(nil).SetBounds(0, 0, 32, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{
		Projects: []core.Project{
			{Name: "Alpha Project", ProjectID: "alpha"},
			{Name: "Beta Project", ProjectID: "beta"},
			{Name: "Gamma Project", ProjectID: "gamma"},
		},
		Source: "cache",
	})
	return m
}

func keyPress(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func keyText(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text})
}

func runBatch(t *testing.T, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		out := make([]tea.Msg, 0, len(batch))
		for _, item := range batch {
			if item == nil {
				continue
			}
			out = append(out, item())
		}
		return out
	}
	return []tea.Msg{msg}
}

func findMsg[T any](msgs []tea.Msg) T {
	for _, msg := range msgs {
		if typed, ok := msg.(T); ok {
			return typed
		}
	}
	var zero T
	return zero
}

const loadedProjectsSnapshot = `── [1] Projects ─────────── 3 ─╮
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
