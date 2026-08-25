package themepicker

import (
	"errors"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	corestyles "github.com/yumauri/fbrcm/core/styles"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestViewShowsPaletteAsLastFieldAndUnavailableTheme(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	palette := corestyles.DefaultPalette()
	m := New().SetBounds(0, 0, 80, 24).Open([]Option{
		{Name: "firebase", Palette: palette},
		{Name: "broken", Err: errors.New("invalid theme")},
	}, 0)

	view := m.View()
	if count := strings.Count(view, "▇"); count != len(corestyles.PreviewTokens())*paletteSwatchWidth {
		t.Fatalf("palette glyph count = %d, want %d:\n%s", count, len(corestyles.PreviewTokens())*paletteSwatchWidth, view)
	}
	if !strings.Contains(view, "broken") || !strings.Contains(view, "unavailable") {
		t.Fatalf("unavailable theme missing from view:\n%s", view)
	}
	for line := range strings.SplitSeq(view, "\n") {
		if strings.Contains(line, "firebase") && strings.LastIndex(line, "▇") < strings.Index(line, "firebase") {
			t.Fatalf("palette is not after theme name: %q", line)
		}
	}
}

func TestViewUsesSelectedHeaderStandardHelpAndExtraRightPadding(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := New().SetBounds(0, 0, 80, 24).Open([]Option{{Name: "firebase", Palette: corestyles.DefaultPalette()}}, 0)
	view := m.View()

	hint := tuiconfig.ActionKeyHint(tuiconfig.BlockGlobal, tuiconfig.ActionThemes) + " "
	title, _ := styles.PanelHeaderTab(hint, "Themes", true, true, 80)
	if !strings.Contains(view, title) {
		t.Fatalf("selected header with theme hint missing:\n%s", view)
	}
	plain := ansi.Strip(view)
	if !strings.Contains(plain, "preview") || !strings.Contains(plain, " • ") || !strings.Contains(plain, "select") || !strings.Contains(plain, "close") {
		t.Fatalf("standard help line missing:\n%s", plain)
	}
	if !strings.Contains(plain, "↑/↓ preview") {
		t.Fatalf("compact navigation help missing:\n%s", plain)
	}
	if strings.Contains(plain, "up/k") || strings.Contains(plain, "down/j") || strings.Contains(plain, "ctrl+t close") || strings.Contains(plain, "q quit") {
		t.Fatalf("alternate keys should not be shown in compact help:\n%s", plain)
	}
	for line := range strings.SplitSeq(plain, "\n") {
		if strings.Contains(line, "firebase") && !strings.HasSuffix(line, "  │ ") {
			t.Fatalf("theme row does not have two spaces before right border: %q", line)
		}
	}
}

func TestSelectedRowBackgroundCoversPalette(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	palette := corestyles.DefaultPalette()
	m := New().SetBounds(0, 0, 80, 24).Open([]Option{{Name: "firebase", Palette: palette}}, 0)

	selectionStyle := styles.TitleStyle(true)
	selectedSwatch := lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette[corestyles.PreviewTokens()[0]])).
		Background(selectionStyle.GetBackground()).
		Render(strings.Repeat(paletteSwatchGlyph, paletteSwatchWidth))
	if !strings.Contains(m.View(), selectedSwatch) {
		t.Fatalf("selected palette swatch does not use the row background:\n%s", m.View())
	}
}

func TestSelectionMovementAndMouseDoubleClick(t *testing.T) {
	options := []Option{
		{Name: "built-in", Palette: corestyles.DefaultPalette()},
		{Name: "firebase", Palette: corestyles.DefaultPalette()},
	}
	m := New().SetBounds(0, 0, 80, 24).Open(options, 1)
	if option, _ := m.Current(); option.Name != "firebase" {
		t.Fatalf("initial option = %q, want firebase", option.Name)
	}
	if !m.Move(-1) {
		t.Fatal("Move did not change selection")
	}
	if option, _ := m.Current(); option.Name != "built-in" {
		t.Fatalf("moved option = %q, want built-in", option.Name)
	}

	boxX, boxY := m.Position()
	now := time.Now()
	row := boxY + firstOptionRow + 1
	changed, double, hit := m.SelectOptionAt(boxX+2, row, now)
	if !hit || !changed || double {
		t.Fatalf("first click = changed:%t double:%t hit:%t", changed, double, hit)
	}
	changed, double, hit = m.SelectOptionAt(boxX+2, row, now.Add(time.Millisecond))
	if !hit || changed || !double {
		t.Fatalf("second click = changed:%t double:%t hit:%t", changed, double, hit)
	}
}

func TestPositionMatchesRenderedBox(t *testing.T) {
	m := New().SetBounds(0, 0, 80, 24).Open([]Option{{Name: "built-in", Palette: corestyles.DefaultPalette()}}, 0)
	w, h := m.boxSize()
	if got := lipgloss.Width(m.View()); got != w {
		t.Fatalf("view width = %d, box width = %d", got, w)
	}
	if got := lipgloss.Height(m.View()); got != h {
		t.Fatalf("view height = %d, box height = %d", got, h)
	}
}

func TestSetCurrentPaletteRefreshesOnlyHighlightedOption(t *testing.T) {
	broken := errors.New("invalid theme")
	m := New().SetBounds(0, 0, 80, 24).Open([]Option{
		{Name: "first", Palette: corestyles.DefaultPalette()},
		{Name: "second", Err: broken},
	}, 1)
	palette := corestyles.DefaultPalette()
	palette[corestyles.TokenPrimary] = "2"
	m = m.SetError(broken).SetCurrentPalette(palette)

	option, ok := m.Current()
	if !ok || option.Err != nil || option.Palette[corestyles.TokenPrimary] != "2" {
		t.Fatalf("updated option = %#v, %t", option, ok)
	}
	palette[corestyles.TokenPrimary] = "3"
	option, _ = m.Current()
	if option.Palette[corestyles.TokenPrimary] != "2" {
		t.Fatal("picker retained the caller's mutable palette map")
	}
	if strings.Contains(m.View(), broken.Error()) {
		t.Fatalf("successful palette update did not clear the error:\n%s", m.View())
	}
}
