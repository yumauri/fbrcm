package styles

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestBackgroundRowPadsWrapsAndKeepsForegroundAcrossLines(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	blue := lipgloss.NewStyle().Foreground(PaletteBlueBright)
	for _, width := range []int{24, 8} {
		got := RenderBackgroundRow(width, ColorInactiveSelection, RowSpan{Text: "prefix"}, RowSpan{Text: "ABCDEFGHIJK", Style: blue})
		for line := range strings.SplitSeq(got, "\n") {
			if lipgloss.Width(line) != width {
				t.Fatalf("row does not fill %d columns: %q", width, line)
			}
		}
		if plain := strings.ReplaceAll(ansi.Strip(got), "\n", ""); strings.ReplaceAll(plain, " ", "") != "prefixABCDEFGHIJK" {
			t.Fatalf("wrapping changed text: %q", got)
		}
		fragments := []string{"ABCDEFGHIJK"}
		if width == 8 {
			fragments = []string{"AB", "CDEFGHIJ", "K"}
		}
		for _, fragment := range fragments {
			if !strings.Contains(got, blue.Background(ColorInactiveSelection).Render(fragment)) {
				t.Fatalf("foreground or background missing from %q:\n%q", fragment, got)
			}
		}
	}
}

func TestBackgroundRowPreservesUnicodeAndExplicitNewlines(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := RenderBackgroundRow(8, ColorHistoryAddedBackground, RowSpan{Text: "条件\n"}, RowSpan{Text: "abcdefghi", Style: PanelMuted})
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("lost explicit newline or wrapping: %q", got)
	}
	for _, line := range lines {
		if lipgloss.Width(line) != 8 {
			t.Fatalf("Unicode row width: %q", line)
		}
	}
	if strings.TrimSpace(ansi.Strip(lines[0])) != "条件" || strings.TrimSpace(ansi.Strip(lines[1])) != "abcdefgh" || strings.TrimSpace(ansi.Strip(lines[2])) != "i" {
		t.Fatalf("Unicode or wrapped text changed: %q", got)
	}
}

func TestBackgroundRowRespectsNoColorWithoutPadding(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := RenderBackgroundRow(20, ColorHistoryRemovedBackground, RowSpan{Text: "heading", Style: PanelMuted.Bold(true)})
	if got != "heading" {
		t.Fatalf("NO_COLOR retained styling or padding: %q", got)
	}
}
