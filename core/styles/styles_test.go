package styles

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestPrecomputedStyleReadsRuntimePalette(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(ColorBlueBright)
	before := style.Render("themed")
	ApplyPalette(Palette{TokenPrimary: "#010203"})
	t.Cleanup(ResetPalette)
	after := style.Render("themed")
	if before == after {
		t.Fatalf("rendered style did not change: %q", after)
	}

	red, green, blue, alpha := style.GetForeground().RGBA()
	if red>>8 != 1 || green>>8 != 2 || blue>>8 != 3 || alpha != 0xffff {
		t.Fatalf("foreground RGBA = %04x %04x %04x %04x", red, green, blue, alpha)
	}
}

func TestApplyPaletteFillsMissingBuiltInColors(t *testing.T) {
	ApplyPalette(Palette{TokenPrimary: "1"})
	t.Cleanup(ResetPalette)
	got := CurrentPalette()
	if got[TokenPrimary] != "1" {
		t.Fatalf("primary = %q", got[TokenPrimary])
	}
	if got[TokenError] != DefaultPalette()[TokenError] {
		t.Fatalf("error = %q", got[TokenError])
	}
}
