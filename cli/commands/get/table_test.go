package get

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRenderValueTreePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := renderValueTree([]valueLine{
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

	got := renderValueTree([]valueLine{{Label: "Missing values", Missing: true}}, "missing", len("Default value"), true, 80, nil)
	if got != "╌╌╌ Missing values" {
		t.Fatalf("missing value tree = %q, want plain missing label", got)
	}
}

func TestTableLayoutKeepsWideColumns(t *testing.T) {
	t.Setenv("COLUMNS", "120")

	layout := chooseTableLayout([]parameterRow{
		{
			Project: "Project A",
			Group:   defaultGroupLabel,
			Key:     "flag",
			Type:    "string",
			ValueLines: []valueLine{
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
	rows := []parameterRow{{Status: "cache"}, {Status: "missing"}}

	if !isStripedDataRow(1) || isStripedDataRow(0) || isStripedDataRow(-1) {
		t.Fatalf("isStripedDataRow parity changed")
	}
	if rowStatus(rows, -1) != "" || rowStatus(rows, 2) != "" || rowStatus(rows, 1) != "missing" {
		t.Fatalf("rowStatus returned unexpected value")
	}
	if !isErrorStatus("missing") || !isErrorStatus("staled") || isErrorStatus("cache") {
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
		{name: "in app default", value: firebase.RemoteConfigValue{UseInAppDefault: true}, want: "<in-app default>"},
		{name: "personalization", value: firebase.RemoteConfigValue{PersonalizationValue: json.RawMessage(`{"x":1}`)}, want: "<personalization>"},
		{name: "rollout", value: firebase.RemoteConfigValue{RolloutValue: json.RawMessage(`{"x":1}`)}, want: "<rollout>"},
		{name: "empty typed", value: firebase.RemoteConfigValue{}, valueType: "NUMBER", want: "(empty number)"},
		{name: "newline", value: firebase.RemoteConfigValue{Value: "a\nb"}, want: `a\nb`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatRemoteConfigValue(tc.value, tc.valueType); got != tc.want {
				t.Fatalf("formatRemoteConfigValue = %q, want %q", got, tc.want)
			}
		})
	}

	if got := valueForJSON("(empty string)"); got != nil {
		t.Fatalf("valueForJSON(empty marker) = %#v, want nil", got)
	}
	if got := valueForJSON("enabled"); got == nil || *got != "enabled" {
		t.Fatalf("valueForJSON(enabled) = %#v, want enabled pointer", got)
	}
	if emptyValueType("  ") != "string" || emptyValueType(" BOOLEAN ") != "boolean" {
		t.Fatalf("emptyValueType normalization changed")
	}
	if valueTypeKey("  ") != "string" || valueTypeKey(" JSON ") != "json" {
		t.Fatalf("valueTypeKey normalization changed")
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

	line := valueLine{Label: "beta", Value: "enabled"}
	if got := valueLineHeadWidth(line, 0, 2, len("Default value"), true); got != 20 {
		t.Fatalf("valueLineHeadWidth with names = %d, want 20", got)
	}
	if got := valueLineHeadWidth(line, 0, 2, len("Default value"), false); got != 4 {
		t.Fatalf("valueLineHeadWidth without names = %d, want 4", got)
	}
	if got := valueLineHeadWidth(valueLine{Missing: true}, 0, 1, len("Default value"), true); got != 4 {
		t.Fatalf("valueLineHeadWidth missing = %d, want 4", got)
	}
}

func TestSortingHelpers(t *testing.T) {
	if got, want := sortedStringKeys(map[string]int{"beta": 1, "Alpha": 2, "alpha": 3}), []string{"Alpha", "alpha", "beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("sortedStringKeys = %#v, want %#v", got, want)
	}

	values := map[string]firebase.RemoteConfigValue{
		"beta":  {},
		"alpha": {},
		"ga":    {},
	}
	order := map[string]int{"ga": 0}
	if got, want := sortedConditionalKeys(values, order), []string{"ga", "alpha", "beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("sortedConditionalKeys = %#v, want %#v", got, want)
	}
}

func TestPointerHelpers(t *testing.T) {
	if got := stringPtrOrNil("  "); got != nil {
		t.Fatalf("stringPtrOrNil(blank) = %#v, want nil", got)
	}
	if got := stringPtrOrNil(" value "); got == nil || *got != " value " {
		t.Fatalf("stringPtrOrNil(value) = %#v, want original value pointer", got)
	}

	if got := timePtrOrNil(time.Time{}); got != nil {
		t.Fatalf("timePtrOrNil(zero) = %#v, want nil", got)
	}
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	if got := timePtrOrNil(now); got == nil || !got.Equal(now) {
		t.Fatalf("timePtrOrNil(now) = %#v, want now pointer", got)
	}
}
