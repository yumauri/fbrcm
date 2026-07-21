package conditions

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
)

func TestRenderConditionsTableNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderConditionsTableAtWidth([]core.ConditionEntry{
		{Priority: 1, Name: "beta", TagColor: "BLUE", Expression: "app.version > '2'", Usages: []core.ConditionUsage{{}, {}}},
		{Priority: 2, Name: "staff", Expression: "user in staff"},
	}, 100)
	for _, want := range []string{"┌", "Priority", "Name", "beta", "app.version > '2'", "staff", "user in staff"} {
		if !strings.Contains(got, want) {
			t.Fatalf("table missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Color") || strings.Contains(got, "● BLUE") {
		t.Fatalf("table still renders a color column:\n%s", got)
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("NO_COLOR table contains ANSI escapes: %q", got)
	}
}

func TestRenderConditionsTableCropsExpressionToTerminalWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	const terminalWidth = 60
	fullExpression := "app.version.matches('a-very-long-version-pattern-that-will-not-fit')"
	got := renderConditionsTableAtWidth([]core.ConditionEntry{{
		Priority: 1, Name: "beta", Expression: fullExpression,
	}}, terminalWidth)

	if strings.Contains(got, fullExpression) || !strings.Contains(got, "…") {
		t.Fatalf("expression was not cropped with ellipsis:\n%s", got)
	}
	for i, line := range strings.Split(ansi.Strip(got), "\n") {
		if width := lipgloss.Width(line); width > terminalWidth {
			t.Fatalf("line %d width = %d, exceeds terminal width %d:\n%s", i, width, terminalWidth, got)
		}
	}
}

func TestRenderConditionDetailsShowsUsageTable(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderConditionDetails(core.ConditionEntry{
		Priority: 2, Name: "staff", Expression: "user in staff",
		Usages: []core.ConditionUsage{{GroupLabel: "(root)", ParameterKey: "welcome", ValueType: "STRING", Value: "Hello"}},
	})
	for _, want := range []string{"Priority: 2", "Name: staff", "Color: —", "Expression: user in staff", "Used by: 1 parameter", "(root)", "welcome", "STRING", "Hello"} {
		if !strings.Contains(got, want) {
			t.Fatalf("details missing %q:\n%s", want, got)
		}
	}
}

func TestRenderConditionDetailsShowsEmptyUsageTable(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := renderConditionDetails(core.ConditionEntry{Name: "staff", Expression: "true"})

	for _, want := range []string{"Used by: 0 parameters", "Group", "Parameter", "Type", "Value"} {
		if !strings.Contains(got, want) {
			t.Fatalf("empty condition details missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "No parameters use this condition") {
		t.Fatalf("condition details uses special empty-state message:\n%s", got)
	}
}

func TestRenderConditionDetailsColorsNameAndCircleLabel(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := renderConditionDetails(core.ConditionEntry{Name: "staff", TagColor: "GREEN", Expression: "true"})
	style := lipgloss.NewStyle().Foreground(clistyles.ConditionLipglossColor("GREEN"))
	for _, want := range []string{style.Render("staff"), style.Render("● GREEN")} {
		if !strings.Contains(got, want) {
			t.Fatalf("details missing colored value %q:\n%s", want, got)
		}
	}
}

func TestRenderUsagesTableUsesGetParameterAndValueStyles(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := renderUsagesTable([]core.ConditionUsage{{
		GroupLabel: "(root)", ParameterKey: "enabled", ValueType: "boolean", Value: "true",
	}})

	namePrefix := stylePrefix(lipgloss.NewStyle().Foreground(clistyles.PaletteBlueBright))
	valuePrefix := stylePrefix(clistyles.RemoteConfigValueStyle("true", "BOOLEAN"))
	if namePrefix == "" || !strings.Contains(got, namePrefix) {
		t.Fatalf("usage table does not use get parameter-name color:\n%s", got)
	}
	if valuePrefix == "" || !strings.Contains(got, valuePrefix) {
		t.Fatalf("usage table does not use get boolean value color:\n%s", got)
	}
	if !strings.Contains(ansi.Strip(got), "BOOLEAN") {
		t.Fatalf("usage table does not uppercase value type:\n%s", got)
	}
}

func stylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, ok := strings.Cut(rendered, "x")
	if !ok {
		return ""
	}
	return prefix
}
