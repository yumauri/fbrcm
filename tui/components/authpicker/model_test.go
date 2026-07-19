package authpicker

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestPickerSelectsMovesAndWraps(t *testing.T) {
	m := New().SetBounds(0, 0, 80, 24).Open("Bind authentication", []string{"Project: demo"}, []Option{
		{Key: "main", Label: "main", Detail: "gcloud ADC · verified"},
		{Key: "work", Label: "work", Detail: "OAuth · verified 0/1"},
	}, 1)
	if current, ok := m.Current(); !ok || current.Key != "work" {
		t.Fatalf("current = %+v, %v; want work", current, ok)
	}
	m.Move(1)
	if current, _ := m.Current(); current.Key != "main" {
		t.Fatalf("wrapped current = %+v, want main", current)
	}
}

func TestPickerViewShowsIdentityProvenanceAndHelp(t *testing.T) {
	m := New().SetBounds(0, 0, 80, 24).Open("Bind authentication", []string{"Project: demo"}, []Option{
		{Key: "main", Label: "main", Detail: "gcloud ADC · verified"},
	}, 0)
	view := ansi.Strip(m.View())
	for _, want := range []string{"Bind authentication", "Project: demo", "main  ·  gcloud ADC · verified", "enter bind", "esc cancel"} {
		if !strings.Contains(view, want) {
			t.Fatalf("picker view missing %q:\n%s", want, view)
		}
	}
}
