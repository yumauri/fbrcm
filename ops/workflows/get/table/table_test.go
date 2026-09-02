package table

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestRenderValueTreePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := renderValueTree([]ValueLine{
		{Label: "beta", Value: "enabled"},
		{Label: "Default value", Value: "disabled", IsDefault: true},
	}, "", len("Default value"), true, 80, nil)

	for _, want := range []string{"╌┬╌ beta", "enabled", " ╰╌ Default value", "disabled"} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderValueTree = %q, want substring %q", got, want)
		}
	}
}

func TestRenderValueTreeMissingPlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := renderValueTree([]ValueLine{{Label: "Missing values", Missing: true}}, "missing", len("Default value"), true, 80, nil)
	if got != "╌╌╌ Missing values" {
		t.Fatalf("missing value tree = %q, want plain missing label", got)
	}
}

func TestTableLayoutKeepsWideColumns(t *testing.T) {
	t.Setenv("COLUMNS", "120")

	layout := chooseTableLayout([]Row{
		{
			Project: "Project A",
			Group:   shared.DefaultRootGroupLabel,
			Key:     "flag",
			Type:    "string",
			ValueLines: []ValueLine{
				{Label: "Default value", Value: "enabled", IsDefault: true},
			},
		},
	}, len("Default value"), true, true)

	if !layout.includeProject || !layout.includeGroup || !layout.includeKey || !layout.includeType || !layout.showNames {
		t.Fatalf("wide layout = %#v, want all columns and names visible", layout)
	}
	if layout.valueWidth < len("Values") {
		t.Fatalf("wide value width = %d, want at least Values width", layout.valueWidth)
	}
}

func TestTableHelpers(t *testing.T) {
	rows := []Row{{Status: "cache"}, {Status: "missing"}}

	if !isStripedDataRow(1) || isStripedDataRow(0) || isStripedDataRow(-1) {
		t.Fatalf("isStripedDataRow parity changed")
	}
	if rowStatus(rows, -1) != "" || rowStatus(rows, 2) != "" || rowStatus(rows, 1) != "missing" {
		t.Fatalf("rowStatus returned unexpected value")
	}
	if !isErrorStatus("missing") || !isErrorStatus("stale") || isErrorStatus("cache") {
		t.Fatalf("isErrorStatus classification changed")
	}
	if tableOverhead(3) != 10 {
		t.Fatalf("tableOverhead(3) = %d, want 10", tableOverhead(3))
	}
}

func TestValueFormattingHelpers(t *testing.T) {
	cases := []struct {
		name      string
		value     firebase.RemoteConfigValue
		valueType string
		want      string
	}{
		{name: "in app default", value: firebase.RemoteConfigValue{UseInAppDefault: true}, want: "(in-app default)"},
		{name: "personalization", value: firebase.RemoteConfigValue{PersonalizationValue: json.RawMessage(`{"x":1}`)}, want: "◈ (personalization)"},
		{name: "experiment", value: firebase.RemoteConfigValue{ExperimentValue: json.RawMessage(`{"variantValue":[{"value":"true"},{"value":"false"}],"exposurePercent":15}`)}, valueType: "BOOLEAN", want: "⚗ 15% : true | false"},
		{name: "opaque experiment", value: firebase.RemoteConfigValue{ExperimentValue: json.RawMessage(`{"x":1}`)}, want: "⚗ (a/b test)"},
		{name: "rollout", value: firebase.RemoteConfigValue{RolloutValue: json.RawMessage(`{"value":"20","percent":10}`)}, want: "◐ 10% → 20 | (no change)"},
		{name: "unknown", value: firebase.RemoteConfigValue{UnknownValueOption: "futureValue", UnknownValue: json.RawMessage(`{}`)}, want: "(futureValue)"},
		{name: "empty typed", value: firebase.RemoteConfigValue{}, valueType: "NUMBER", want: "(empty number)"},
		{name: "newline", value: firebase.RemoteConfigValue{Value: "a\nb"}, want: `a\nb`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := core.FormatRemoteConfigDisplayValue(tc.value, tc.valueType); got != tc.want {
				t.Fatalf("FormatRemoteConfigDisplayValue = %q, want %q", got, tc.want)
			}
		})
	}

	if got := ValueForJSON("(empty string)"); got != nil {
		t.Fatalf("ValueForJSON(empty marker) = %#v, want nil", got)
	}
	if got := ValueForJSON("enabled"); got == nil || *got != "enabled" {
		t.Fatalf("ValueForJSON(enabled) = %#v, want enabled pointer", got)
	}
	if got := core.FormatRemoteConfigDisplayValue(firebase.RemoteConfigValue{}, "  "); got != "(empty string)" {
		t.Fatalf("FormatRemoteConfigDisplayValue(empty) = %q, want (empty string)", got)
	}
	if got := core.FormatRemoteConfigDisplayValue(firebase.RemoteConfigValue{}, " BOOLEAN "); got != "(empty boolean)" {
		t.Fatalf("FormatRemoteConfigDisplayValue(empty boolean) = %q, want (empty boolean)", got)
	}
	if ValueTypeKey("  ") != "string" || ValueTypeKey(" JSON ") != "json" {
		t.Fatalf("ValueTypeKey normalization changed")
	}
}

func TestInAppDefaultUsesEmptyValueStyle(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	const label = "(in-app default)"

	got := renderValueText(label, "BOOLEAN", nil)
	want := corestyles.EmptyValueStyle().Render(label)
	if got != want {
		t.Fatalf("renderValueText = %q, want shared empty-value style %q", got, want)
	}
	if style := clistyles.RemoteConfigValueStyle(label, "BOOLEAN"); style.Render(label) != want {
		t.Fatalf("RemoteConfigValueStyle did not return the shared empty-value style")
	}
}

func TestManagedValuesUseStructuredFragmentStyles(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	rollout := core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
		RolloutValue: json.RawMessage(`{"value":"20","percent":10}`),
	}, "NUMBER")
	got := renderValueLineText(ValueLine{
		Value: rollout.Text, ValueType: "NUMBER", Display: rollout,
	}, rollout.Text, nil)
	for fragment, want := range map[string]string{
		"icon":       clistyles.PanelMuted.Render("◐ "),
		"percentage": lipgloss.NewStyle().Foreground(clistyles.PaletteGold).Render("10%"),
		"arrow":      clistyles.PanelMuted.Render(" → "),
		"value":      corestyles.ValueTextStyle("20", "NUMBER").Render("20"),
		"separator":  clistyles.PanelMuted.Render(" | "),
		"control":    clistyles.PanelMuted.Render("(no change)"),
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rollout does not use the %s style: %q", fragment, got)
		}
	}

	experiment := core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
		ExperimentValue: json.RawMessage(`{"variantValue":[{"value":""},{"noChange":true},{"value":"new value"}],"exposurePercent":15}`),
	}, "STRING")
	got = renderValueLineText(ValueLine{
		Value: experiment.Text, ValueType: "STRING", Display: experiment,
	}, experiment.Text, nil)
	wantExperiment := clistyles.PanelMuted.Render("⚗ ") +
		lipgloss.NewStyle().Foreground(clistyles.PaletteGold).Render("15%") +
		clistyles.PanelMuted.Render(" : ") +
		clistyles.PanelMuted.Render("(empty string)") +
		clistyles.PanelMuted.Render(" | ") +
		clistyles.PanelMuted.Render("(no change)") +
		clistyles.PanelMuted.Render(" | ") +
		corestyles.ValueTextStyle("new value", "STRING").Render("new value")
	if got != wantExperiment {
		t.Fatalf("experiment render = %q, want %q", got, wantExperiment)
	}

	for name, display := range map[string]rcdisplay.ValueSummary{
		"personalization": core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
			PersonalizationValue: json.RawMessage(`{}`),
		}, "STRING"),
		"experiment": core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
			ExperimentValue: json.RawMessage(`{}`),
		}, "STRING"),
		"unknown": core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
			UnknownValueOption: "futureValue", UnknownValue: json.RawMessage(`{}`),
		}, "STRING"),
	} {
		got := renderValueLineText(ValueLine{Value: display.Text, Display: display}, display.Text, nil)
		if want := clistyles.PanelMuted.Render(display.Text); got != want {
			t.Fatalf("%s render = %q, want muted %q", name, got, want)
		}
	}
}

func TestManagedValuesRespectNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	display := core.SummarizeRemoteConfigDisplayValue(firebase.RemoteConfigValue{
		RolloutValue: json.RawMessage(`{"value":"20","percent":10}`),
	}, "NUMBER")
	got := renderValueLineText(ValueLine{
		Value: display.Text, ValueType: "NUMBER", Display: display,
	}, display.Text, nil)
	if got != "◐ 10% → 20 | (no change)" {
		t.Fatalf("no-color rollout = %q", got)
	}
}

func TestManagedValueTablesRespectNaturalAndNarrowWidths(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     firebase.RemoteConfigValue
		valueType string
	}{
		{
			name: "rollout", key: "funding_minimum_amount", valueType: "NUMBER",
			value: firebase.RemoteConfigValue{RolloutValue: json.RawMessage(`{"value":"20","percent":10}`)},
		},
		{
			name: "experiment", key: "show_income_question", valueType: "BOOLEAN",
			value: firebase.RemoteConfigValue{ExperimentValue: json.RawMessage(`{"variantValue":[{"value":"true"},{"value":"false"},{"value":"true"}],"exposurePercent":15}`)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", "1")
			display := core.SummarizeRemoteConfigDisplayValue(test.value, test.valueType)
			rows := []Row{{
				Group: "onboarding",
				Key:   test.key,
				Type:  test.valueType,
				ValueLines: []ValueLine{{
					Label: "android_beta_10", Value: display.Text, ValueType: test.valueType, Display: display,
				}},
			}}

			t.Setenv("COLUMNS", "120")
			natural := Render(rows, nil, true, false)
			if !strings.Contains(natural, display.Text) {
				t.Fatalf("natural %s table missing complete summary:\n%s", test.name, natural)
			}
			for index, line := range strings.Split(natural, "\n") {
				if width := lipgloss.Width(line); width > 120 {
					t.Fatalf("natural line %d width = %d, want at most 120:\n%s", index, width, natural)
				}
			}

			t.Setenv("COLUMNS", "24")
			narrow := Render(rows, nil, true, false)
			for index, line := range strings.Split(narrow, "\n") {
				if width := lipgloss.Width(line); width > 24 {
					t.Fatalf("narrow line %d width = %d, want at most 24:\n%s", index, width, narrow)
				}
			}
			if !strings.Contains(narrow, "…") {
				t.Fatalf("narrow %s table does not crop with an ellipsis:\n%s", test.name, narrow)
			}
		})
	}
}

func TestClippingAndValueLineWidths(t *testing.T) {
	if got := clipPlainText("abcdef", 4); got != "abc…" {
		t.Fatalf("clipPlainText = %q, want abc…", got)
	}
	if got := clipPlainText("abcdef", 1); got != "…" {
		t.Fatalf("clipPlainText width 1 = %q, want …", got)
	}
	if got := clipPlainText("abcdef", 0); got != "" {
		t.Fatalf("clipPlainText width 0 = %q, want empty", got)
	}
	if got := clipStyledLine("abcdef", 4); got != "abc…" {
		t.Fatalf("clipStyledLine = %q, want abc…", got)
	}

	line := ValueLine{Label: "beta", Value: "enabled"}
	if got := valueLineHeadWidth(line, 0, 2, len("Default value"), true); got != 20 {
		t.Fatalf("valueLineHeadWidth with names = %d, want 20", got)
	}
	if got := valueLineHeadWidth(line, 0, 2, len("Default value"), false); got != 4 {
		t.Fatalf("valueLineHeadWidth without names = %d, want 4", got)
	}
	if got := valueLineHeadWidth(ValueLine{Missing: true}, 0, 1, len("Default value"), true); got != 4 {
		t.Fatalf("valueLineHeadWidth missing = %d, want 4", got)
	}
}

func TestSortingHelpers(t *testing.T) {
	values := map[string]firebase.RemoteConfigValue{
		"beta":  {},
		"alpha": {},
		"ga":    {},
	}
	order := map[string]int{"ga": 0}
	if got, want := SortedConditionalKeys(values, order), []string{"ga", "alpha", "beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SortedConditionalKeys = %#v, want %#v", got, want)
	}
}
