package details

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func conditionDetailsTestModel() Model {
	return New().SetBounds(0, 0, 60, 40).SetActive(true).SetConditionData(&messages.ConditionViewData{
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Condition: core.ConditionEntry{
			Priority:   1,
			Name:       "staff",
			TagColor:   "GREEN",
			Expression: "true",
			Usages: []core.ConditionUsage{
				{GroupLabel: "checkout", ParameterKey: "enabled", ValueType: "BOOLEAN", Value: "true", RawValue: "true", Plain: true},
				{GroupLabel: "(root)", ParameterKey: "payload", ValueType: "JSON", Value: `{"enabled":true,"limit":2}`, RawValue: `{"enabled":true,"limit":2}`, Plain: true},
			},
		},
	})
}

func TestConditionDetailsRendersStyledUsageStructureAndPrettyJSON(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	got := testutil.NormalizeViewSnapshot(ansi.Strip(conditionDetailsTestModel().View()))

	for _, want := range []string{
		"Color\n │ ● GREEN",
		"Used by 2 parameters",
		"checkout / enabled",
		"│   true",
		"(root) / payload",
		"│   {",
		`"enabled": true,`,
		`"limit": 2`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("condition Details missing %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"BOOLEAN ·", "JSON ·"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("condition Details still renders parameter type %q:\n%s", unwanted, got)
		}
	}
}

func TestConditionDetailsUsesGroupAndTypedValueStyles(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	got := strings.Join(conditionDetailsTestModel().renderContentLines(), "\n")

	for name, style := range map[string]lipgloss.Style{
		"group": groupValueStyle,
		"value": corestyles.ValueTextStyle("true", "BOOLEAN"),
	} {
		prefix := renderedStylePrefix(style)
		if prefix == "" || !strings.Contains(got, prefix) {
			t.Fatalf("condition Details does not use %s style:\n%s", name, got)
		}
	}
}

func renderedStylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, ok := strings.Cut(rendered, "x")
	if !ok {
		return ""
	}
	return prefix
}
