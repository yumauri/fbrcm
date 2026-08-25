package about

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	corestyles "github.com/yumauri/fbrcm/core/styles"
)

func TestBuildInfoTextMatchesVersionOutput(t *testing.T) {
	info := BuildInfo{Version: "1.2.3", Commit: "abc123", Date: "2026-06-14"}
	want := Logo + "\nfbrcm 1.2.3 (commit abc123, built 2026-06-14)\n" + Author + "\n"
	if got := info.Text(false); got != want {
		t.Fatalf("plain About text = %q, want %q", got, want)
	}
	colored := info.Text(true)
	if !strings.Contains(colored, "\x1b[") || ansi.Strip(colored) != want {
		t.Fatalf("colored About text = %q, want styled %q", colored, want)
	}
}

func TestLogoGradientUsesRuntimeTheme(t *testing.T) {
	corestyles.ApplyPalette(corestyles.Palette{
		corestyles.TokenLogoStart:  "#010203",
		corestyles.TokenLogoMiddle: "#040506",
		corestyles.TokenLogoEnd:    "#070809",
	})
	t.Cleanup(corestyles.ResetPalette)

	if got := interpolateLogoColor(0, 101); got != (logoColor{red: 1, green: 2, blue: 3}) {
		t.Fatalf("start = %#v", got)
	}
	if got := interpolateLogoColor(20, 101); got != (logoColor{red: 4, green: 5, blue: 6}) {
		t.Fatalf("middle = %#v", got)
	}
	if got := interpolateLogoColor(100, 101); got != (logoColor{red: 7, green: 8, blue: 9}) {
		t.Fatalf("end = %#v", got)
	}
}

func TestLogoGradientUsesAttachedLogoPaletteAndRedBias(t *testing.T) {
	want := []struct {
		column int
		color  logoColor
	}{
		{column: 0, color: logoColor{red: 0xff, green: 0xc4, blue: 0x00}},
		{column: 20, color: logoColor{red: 0xff, green: 0x91, blue: 0x00}},
		{column: 100, color: logoColor{red: 0xdd, green: 0x2c, blue: 0x00}},
	}
	for _, expected := range want {
		if got := interpolateLogoColor(expected.column, 101); got != expected.color {
			t.Fatalf("gradient column %d = %#v, want %#v", expected.column, got, expected.color)
		}
	}
}
