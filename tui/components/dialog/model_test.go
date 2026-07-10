package dialog

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

type pressedMsg string

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
	buttonX := boxX + 3 + max(contentWidth-printableWidth(m.renderButtons()), 0)
	buttonY := boxY + m.bodyHeight() + 3

	next, cmd := m.Update(tea.MouseClickMsg{X: buttonX, Y: buttonY, Button: tea.MouseLeft})
	if next.IsOpen() || cmd == nil {
		t.Fatalf("click left dialog open=%v cmd nil=%v", next.IsOpen(), cmd == nil)
	}
	if got := cmd(); got != pressedMsg("apply") {
		t.Fatalf("command returned %#v", got)
	}
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
