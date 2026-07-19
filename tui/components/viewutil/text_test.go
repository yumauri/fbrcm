package viewutil

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestIndentLines(t *testing.T) {
	if got, want := IndentLines("first\nsecond", 1), " first\n second"; got != want {
		t.Fatalf("IndentLines() = %q, want %q", got, want)
	}
}

func TestSelectorLineUsesUnstyledTwoCellInset(t *testing.T) {
	for _, selected := range []bool{false, true} {
		got := SelectorLine("Profile", selected)
		if !strings.HasPrefix(got, "  ") || strings.HasPrefix(got, "   ") {
			t.Fatalf("SelectorLine(selected=%v) = %q, want exactly two leading spaces", selected, got)
		}
		if plain := ansi.Strip(got); plain != "  Profile" {
			t.Fatalf("SelectorLine(selected=%v) = %q after styling", selected, plain)
		}
	}
}
