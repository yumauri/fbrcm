package authpicker

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestPickerSelectsMovesAndWraps(t *testing.T) {
	m := New().SetBounds(0, 0, 80, 24).Open("Bind authentication", []string{"Project: demo"}, []Option{
		{Key: "main", Label: "main", Detail: "gcloud ADC"},
		{Key: "work", Label: "work", Detail: "OAuth"},
	}, 1)
	if current, ok := m.Current(); !ok || current.Key != "work" {
		t.Fatalf("current = %+v, %v; want work", current, ok)
	}
	m.Move(1)
	if current, _ := m.Current(); current.Key != "main" {
		t.Fatalf("wrapped current = %+v, want main", current)
	}
}

func TestPickerViewUsesDialogButtonsAndCompactSeparators(t *testing.T) {
	m := New().SetBounds(0, 0, 80, 24).Open("Bind authentication", []string{"Project: demo"}, []Option{
		{Key: "main", Label: "main", Detail: "gcloud ADC"},
	}, 0)
	view := ansi.Strip(m.View())
	for _, want := range []string{"Bind authentication", "Project: demo", "main · gcloud ADC", "Bind", "Cancel"} {
		if !strings.Contains(view, want) {
			t.Fatalf("picker view missing %q:\n%s", want, view)
		}
	}
	for _, unwanted := range []string{"verified", "main  ·", "enter bind", "esc cancel"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("picker view contains %q:\n%s", unwanted, view)
		}
	}
	lines := strings.Split(view, "\n")
	wantWidth := lipgloss.Width(lines[0])
	for index, line := range lines {
		if got := lipgloss.Width(line); got != wantWidth {
			t.Fatalf("line %d width = %d, want %d:\n%s", index, got, wantWidth, view)
		}
	}
}

func TestPickerWidthFitsContentBeforeSharedFramePadding(t *testing.T) {
	m := New().SetBounds(0, 0, 120, 30).Open("Bind", []string{"Project: demo"}, []Option{
		{Key: "main", Label: "main", Detail: "OAuth"},
	}, 0)
	natural := max(
		lipgloss.Width("Bind"),
		lipgloss.Width("Project: demo"),
		lipgloss.Width("main · OAuth"),
		lipgloss.Width(m.buttons.View()),
	)
	if got := m.contentWidth(); got != natural {
		t.Fatalf("content width = %d, want natural %d", got, natural)
	}
}

func TestPickerCancelButtonHitAreaMatchesView(t *testing.T) {
	m := New().SetBounds(10, 5, 80, 24).Open("Bind authentication", []string{"Project: demo"}, []Option{
		{Key: "main", Label: "main", Detail: "OAuth"},
	}, 0)
	boxX, boxY := m.Position()
	for row, line := range strings.Split(ansi.Strip(m.View()), "\n") {
		before, _, found := strings.Cut(line, "Cancel")
		if !found {
			continue
		}
		if !m.SelectButtonAt(boxX+lipgloss.Width(before), boxY+row) || m.SelectedButton() != 1 {
			t.Fatalf("Cancel button did not select at rendered position")
		}
		return
	}
	t.Fatal("Cancel button not found")
}
