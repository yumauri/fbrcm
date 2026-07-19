package dialog

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/styles"
)

type pressedMsg string

func TestSuccessToneUsesSuccessBorder(t *testing.T) {
	m := New().Open(Config{Tone: ToneSuccess})
	if got := m.borderStyle().GetForeground(); got != styles.PaletteSuccess {
		t.Fatalf("success border foreground = %v, want %v", got, styles.PaletteSuccess)
	}
	m = m.Close()
	if got := m.borderStyle().GetForeground(); got != styles.PaletteError {
		t.Fatalf("closed dialog border foreground = %v, want default %v", got, styles.PaletteError)
	}
}

func TestNavigationWrapsAndScrollClamps(t *testing.T) {
	m := New().SetBounds(0, 0, 80, 16).Open(Config{
		Body:    []string{"1", "2", "3", "4", "5", "6", "7", "8"},
		Buttons: []Button{{Label: "One"}, {Label: "Two"}},
	})

	m.move(-1)
	if m.selected != 1 {
		t.Fatalf("wrapped selection = %d, want 1", m.selected)
	}
	m.scrollBy(100)
	if m.scroll != m.maxScroll() {
		t.Fatalf("scroll = %d, want max %d", m.scroll, m.maxScroll())
	}
	m.scrollBy(-100)
	if m.scroll != 0 {
		t.Fatalf("scroll = %d, want 0", m.scroll)
	}
}

func TestMouseButtonClickRunsCommandAndCloses(t *testing.T) {
	m := New().SetBounds(0, 0, 100, 30).Open(Config{
		Title: "Confirm",
		Body:  []string{"body"},
		Buttons: []Button{{
			Label:   "Apply",
			OnPress: func() tea.Msg { return pressedMsg("apply") },
		}},
	})
	boxX, boxY, _, _ := m.boxGeometry()
	contentWidth := m.contentWidth()
	buttonX := boxX + 4 + max(contentWidth-printableWidth(m.renderButtons()), 0)
	buttonY := boxY + m.bodyHeight() + 3

	next, cmd := m.Update(tea.MouseClickMsg{X: buttonX, Y: buttonY, Button: tea.MouseLeft})
	if next.IsOpen() || cmd == nil {
		t.Fatalf("click left dialog open=%v cmd nil=%v", next.IsOpen(), cmd == nil)
	}
	if got := cmd(); got != pressedMsg("apply") {
		t.Fatalf("command returned %#v", got)
	}
}

func TestEveryRenderedButtonIsClickable(t *testing.T) {
	buttons := []Button{
		{Label: "Save Draft", OnPress: func() tea.Msg { return pressedMsg("draft") }},
		{Label: "Publish Now", OnPress: func() tea.Msg { return pressedMsg("publish") }},
		{Label: "Cancel", OnPress: func() tea.Msg { return pressedMsg("cancel") }},
	}
	for index, button := range buttons {
		m := New().SetBounds(0, 0, 120, 30).Open(Config{Title: "Import", Body: []string{"body"}, Buttons: buttons})
		x, y := renderedTextPoint(t, m, button.Label)
		if got, ok := m.buttonIndexAt(x, y); !ok || got != index {
			boxX, boxY, boxWidth, boxHeight := m.boxGeometry()
			t.Fatalf("button %q point (%d,%d) resolves to index=%d ok=%v; box=(%d,%d %dx%d):\n%s", button.Label, x, y, got, ok, boxX, boxY, boxWidth, boxHeight, ansi.Strip(m.View()))
		}
		next, cmd := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
		if next.IsOpen() || cmd == nil {
			t.Fatalf("button %q click left dialog open=%v cmd nil=%v", button.Label, next.IsOpen(), cmd == nil)
		}
		if got := cmd(); got != pressedMsg([]string{"draft", "publish", "cancel"}[index]) {
			t.Fatalf("button %q command returned %#v", button.Label, got)
		}
	}
}

func TestMultilineBodyIsWrappedInsideFrame(t *testing.T) {
	m := New().SetBounds(0, 0, 100, 24).Open(Config{
		Title: "Import Failed",
		Body: []string{
			"Project: Demo (demo)",
			"",
			"firebase error: validate remote config api returned 400 Bad Request: {\n  \"error\": {\n    \"code\": 400,\n    \"message\": \"A validation result with a deliberately long explanation that must remain inside the dialog frame.\"\n  }\n}",
		},
		Buttons: []Button{{Label: "Close"}},
	})
	view := ansi.Strip(m.View())
	lines := strings.Split(view, "\n")
	wantWidth := lipgloss.Width(lines[0])
	for index, line := range lines {
		if got := lipgloss.Width(line); got != wantWidth {
			t.Fatalf("line %d width = %d, want %d:\n%s", index, got, wantWidth, view)
		}
	}
	for _, want := range []string{"firebase error:", "\"error\": {", "\"code\": 400"} {
		if !strings.Contains(view, want) {
			t.Fatalf("wrapped dialog missing %q:\n%s", want, view)
		}
	}
}

func renderedTextPoint(t *testing.T, m Model, label string) (int, int) {
	t.Helper()
	x, y := m.Position()
	for row, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		if before, _, found := strings.Cut(line, label); found {
			return x + lipgloss.Width(before), y + row
		}
	}
	t.Fatalf("dialog does not render button %q", label)
	return 0, 0
}

func TestMouseDragClampsDialogToBounds(t *testing.T) {
	m := New().SetBounds(10, 5, 70, 20).Open(Config{
		Title:   "Drag",
		Body:    []string{"body"},
		Buttons: []Button{{Label: "OK"}},
	})
	boxX, boxY, _, _ := m.boxGeometry()
	m, _ = m.Update(tea.MouseClickMsg{X: boxX, Y: boxY, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseMotionMsg{X: -100, Y: -100, Button: tea.MouseLeft})
	x, y, _, _ := m.boxGeometry()
	if x != 10 || y != 5 {
		t.Fatalf("dragged position = (%d,%d), want (10,5)", x, y)
	}
	m, _ = m.Update(tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft})
	if m.dragging {
		t.Fatal("mouse release should stop dragging")
	}
}

func TestNarrowBoundsRemainContained(t *testing.T) {
	m := New().SetBounds(3, 4, 8, 5).Open(Config{Title: "Long title", Body: []string{"body"}})
	x, y, _, _ := m.boxGeometry()
	if x != 3 || y != 4 || !m.Contains(3, 4) {
		t.Fatalf("geometry origin = (%d,%d), contains origin=%v", x, y, m.Contains(3, 4))
	}
}
