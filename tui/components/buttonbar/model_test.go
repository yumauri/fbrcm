package buttonbar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestButtonHitAreasMatchRenderedButtons(t *testing.T) {
	m := New([]Button{{Label: "Continue", Variant: VariantAccent}, {Label: "Cancel", Variant: VariantAccent}}).
		SetFocused(true)
	view := ansi.Strip(m.View())
	for index, label := range []string{"Continue", "Cancel"} {
		lines := strings.Split(view, "\n")
		for y, line := range lines {
			before, _, found := strings.Cut(line, label)
			if !found {
				continue
			}
			if got, ok := m.IndexAt(len([]rune(before)), y); !ok || got != index {
				t.Fatalf("%s hit = %d ok=%v, want %d", label, got, ok, index)
			}
			break
		}
	}
}

func TestButtonSelectionWraps(t *testing.T) {
	m := New([]Button{{Label: "One"}, {Label: "Two"}})
	m.Move(-1)
	if m.Selected() != 1 {
		t.Fatalf("selected = %d, want 1", m.Selected())
	}
}
