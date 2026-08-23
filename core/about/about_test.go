package about

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
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
