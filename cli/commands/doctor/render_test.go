package doctor

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/env"
)

func TestRenderDoctorTableUsesSharedTableShape(t *testing.T) {
	t.Setenv(env.NoColor, "1")
	got := renderDoctorTable([]core.DoctorCheck{
		{Status: core.DoctorPass, Check: "Profile", Detail: "default"},
		{Status: core.DoctorFail, Check: "Credentials", Detail: "missing"},
	})
	for _, want := range []string{"Status", "Check", "Detail", "PASS", "FAIL", "Profile", "Credentials", "┌", "┬"} {
		if !strings.Contains(got, want) {
			t.Fatalf("table missing %q:\n%s", want, got)
		}
	}
}

func TestRenderDoctorTableWrapsDetailsWithinTerminalWidth(t *testing.T) {
	t.Setenv(env.NoColor, "1")
	const terminalWidth = 72
	longDetail := "oauth: /Users/example/.config/fbrcm/production/auth/default/client-secret.json; refresh token available"
	got := renderDoctorTableAtWidth([]core.DoctorCheck{{
		Status: core.DoctorWarn,
		Check:  "Credentials (production-default)",
		Detail: longDetail,
	}}, terminalWidth)

	lines := strings.Split(ansi.Strip(got), "\n")
	for i, line := range lines {
		if width := lipgloss.Width(line); width > terminalWidth {
			t.Fatalf("line %d width = %d, exceeds terminal width %d:\n%s", i, width, terminalWidth, got)
		}
	}
	if len(lines) <= 5 {
		t.Fatalf("long detail did not produce a multiline table row:\n%s", got)
	}
	if !strings.Contains(got, "Credentials (production-default)") {
		t.Fatalf("Check column was wrapped or cropped:\n%s", got)
	}
	for _, want := range []string{"client-", "secret.json", "refresh", "token available", "production"} {
		if !strings.Contains(got, want) {
			t.Fatalf("wrapped table lost %q:\n%s", want, got)
		}
	}
}

func TestRenderDoctorTableUsesConfiguredTerminalWidth(t *testing.T) {
	t.Setenv(env.NoColor, "1")
	t.Setenv("COLUMNS", "52")
	got := renderDoctorTable([]core.DoctorCheck{{
		Status: core.DoctorFail,
		Check:  "Firebase permissions",
		Detail: "missing: cloudconfig.configs.get, cloudconfig.configs.update",
	}})
	for i, line := range strings.Split(ansi.Strip(got), "\n") {
		if width := lipgloss.Width(line); width > 52 {
			t.Fatalf("line %d width = %d, exceeds COLUMNS=52:\n%s", i, width, got)
		}
	}
}

func TestRenderDoctorTableUsesNaturalWidthWhenContentFits(t *testing.T) {
	t.Setenv(env.NoColor, "1")
	const terminalWidth = 120
	got := renderDoctorTableAtWidth([]core.DoctorCheck{{
		Status: core.DoctorPass,
		Check:  "Profile",
		Detail: "default",
	}}, terminalWidth)

	maxWidth := 0
	for line := range strings.SplitSeq(ansi.Strip(got), "\n") {
		maxWidth = max(maxWidth, lipgloss.Width(line))
	}
	if maxWidth >= terminalWidth {
		t.Fatalf("table width = %d, should use natural width below terminal width %d:\n%s", maxWidth, terminalWidth, got)
	}
	if maxWidth != 30 {
		t.Fatalf("table width = %d, want natural width 30:\n%s", maxWidth, got)
	}
}
