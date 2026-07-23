package viewutil

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestWrapRenderedLineUsesContinuationIndentAndPreservesStyle(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	lines := WrapRenderedLine("  "+style.Render("abcdefghij"), 7, 4)
	if len(lines) != 3 {
		t.Fatalf("wrapped lines = %d, want 3: %#v", len(lines), lines)
	}
	if got := strings.Join([]string{ansi.Strip(lines[0]), ansi.Strip(lines[1]), ansi.Strip(lines[2])}, "\n"); got != "  abcde\n    fgh\n    ij" {
		t.Fatalf("wrapped text = %q", got)
	}
	if !strings.Contains(lines[1], style.Render("fgh")) {
		t.Fatalf("continuation lost foreground style: %q", lines[1])
	}
}
