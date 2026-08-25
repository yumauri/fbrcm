package theme

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	corestyles "github.com/yumauri/fbrcm/core/styles"
)

func TestRenderThemesTableUsesNaturalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderThemesTableAtWidth([]themeListItem{{Theme: "firebase", Active: true}, {Theme: "nord"}}, 120)
	if !strings.Contains(got, "firebase") || !strings.Contains(got, "✓") {
		t.Fatalf("table =\n%s", got)
	}
	if strings.Contains(got, "Palette") || strings.Contains(got, "\x1b[") {
		t.Fatalf("NO_COLOR table contains palette presentation: %q", got)
	}
	for line := range strings.SplitSeq(got, "\n") {
		if width := lipgloss.Width(line); width > 22 {
			t.Fatalf("natural table line width = %d, want <= 22:\n%s", width, got)
		}
	}
}

func TestRenderThemesTableIncludesEightColorPalette(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := renderThemesTableAtWidth([]themeListItem{{Theme: coreconfig.BuiltInThemeName, Active: true, BuiltIn: true}}, 120)
	plain := ansi.Strip(got)
	if !strings.Contains(plain, "Palette") || !strings.Contains(plain, strings.Repeat(paletteSwatchGlyph, paletteSwatchCount*paletteSwatchWidth)) {
		t.Fatalf("table has no eight-color palette:\n%s", got)
	}
	if strings.Index(plain, "Active") > strings.Index(plain, "Palette") {
		t.Fatalf("palette is not the last column:\n%s", got)
	}
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("palette table has no ANSI colors: %q", got)
	}
	for line := range strings.SplitSeq(got, "\n") {
		if width := lipgloss.Width(line); width > 41 {
			t.Fatalf("natural palette table line width = %d, want <= 41:\n%s", width, got)
		}
	}
}

func TestRenderThemesTableShrinksPaletteInNarrowTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := renderThemesTableAtWidth([]themeListItem{{Theme: coreconfig.BuiltInThemeName, Active: true, BuiltIn: true}}, 24)
	if !strings.Contains(ansi.Strip(got), strings.Repeat(paletteSwatchGlyph, paletteSwatchWidth)) {
		t.Fatalf("narrow table has no palette swatch:\n%s", got)
	}
	for line := range strings.SplitSeq(got, "\n") {
		if width := lipgloss.Width(line); width > 24 {
			t.Fatalf("narrow table line width = %d, exceeds 24:\n%s", width, got)
		}
	}
}

func TestRenderCurrentThemeIncludesPaletteOnlyWhenColorsEnabled(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	colored := renderCurrentTheme("", 120)
	if ansi.Strip(colored) != builtInThemeLabel+"  "+strings.Repeat(paletteSwatchGlyph, paletteSwatchCount*paletteSwatchWidth) || !strings.Contains(colored, "\x1b[") {
		t.Fatalf("colored current theme = %q", colored)
	}

	t.Setenv("NO_COLOR", "1")
	if got := renderCurrentTheme("", 120); got != builtInThemeLabel {
		t.Fatalf("NO_COLOR current theme = %q", got)
	}
}

func TestThemeJSONOutputDoesNotIncludePalettePreview(t *testing.T) {
	setupThemeCommandTest(t)
	cmd := New()
	cmd.Flags().Bool("json", true, "")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	if strings.Contains(output.String(), "palette") || strings.Contains(output.String(), "Palette") {
		t.Fatalf("JSON output contains palette preview: %s", output.String())
	}
}

func TestThemeListJSONOutputDoesNotIncludePalettePreview(t *testing.T) {
	setupThemeCommandTest(t)
	cmd := newListCommand()
	cmd.Flags().Bool("json", true, "")
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute = %v", err)
	}
	if strings.Contains(output.String(), "palette") || strings.Contains(output.String(), "Palette") {
		t.Fatalf("JSON output contains palette preview: %s", output.String())
	}
}

func TestRenderPalettePreviewUsesRepresentativeSemanticColors(t *testing.T) {
	palette := corestyles.DefaultPalette()
	preview := renderPalettePreview(palette, paletteSwatchCount*paletteSwatchWidth)
	if lipgloss.Width(preview) != paletteSwatchCount*paletteSwatchWidth {
		t.Fatalf("preview width = %d, want %d", lipgloss.Width(preview), paletteSwatchCount*paletteSwatchWidth)
	}
	for _, token := range corestyles.PreviewTokens() {
		if palette[token] == "" {
			t.Fatalf("preview token %q is absent from palette", token)
		}
	}
}

func TestRenderThemesTableCropsNameInNarrowTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderThemesTableAtWidth([]themeListItem{{Theme: strings.Repeat("firebase", 10), Active: true}}, 24)
	if !strings.Contains(ansi.Strip(got), "…") {
		t.Fatalf("cropped table has no ellipsis:\n%s", got)
	}
	for line := range strings.SplitSeq(got, "\n") {
		if width := lipgloss.Width(line); width > 24 {
			t.Fatalf("narrow table line width = %d, exceeds 24:\n%s", width, got)
		}
	}
}
