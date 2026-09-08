package diffinput

import (
	"testing"

	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

func TestParameterPreservesRemoteConfigComparisonSemantics(t *testing.T) {
	parameter := &firebase.RemoteConfigParam{
		ValueType:   "BOOLEAN",
		Description: "Feature state",
		DefaultValue: &firebase.RemoteConfigValue{
			Value: "true",
		},
		ConditionalValues: map[string]firebase.RemoteConfigValue{
			"Store": {UseInAppDefault: true},
		},
	}
	properties := Parameter("WEB", parameter)
	if properties["type"].CompareAs != dictdiff.CompareEnum ||
		properties["group"].CompareAs != dictdiff.CompareEnum ||
		properties["description"].CompareAs != dictdiff.CompareString {
		t.Fatalf("parameter metadata comparison hints = %#v", properties)
	}
	if value := properties["value · default"]; value.Type != dictdiff.ValueBoolean ||
		value.CompareAs != dictdiff.CompareEnum || !value.Boolean {
		t.Fatalf("boolean value = %#v", value)
	}
	if value := properties["value · Store"]; value.CompareAs != dictdiff.CompareEnum ||
		value.Raw != "in-app default" {
		t.Fatalf("special value = %#v", value)
	}
}

func TestValueKeepsInvalidJSONForRawComparison(t *testing.T) {
	value := Value(firebase.RemoteConfigValue{Value: `{"broken"`}, "JSON")
	if value.Type != dictdiff.ValueJSON || value.CompareAs != dictdiff.CompareJSON ||
		value.Raw != `{"broken"` {
		t.Fatalf("invalid JSON adapter value = %#v", value)
	}
}

func TestValuePreservesUnknownValueOptionForComparison(t *testing.T) {
	value := Value(firebase.RemoteConfigValue{
		UnknownValueOption: "someFutureValue",
		UnknownValue:       []byte(`{"id":"future-1"}`),
	}, "STRING")
	if value.Type != dictdiff.ValueJSON || value.CompareAs != dictdiff.CompareJSON ||
		value.Raw != `{"someFutureValue":{"id":"future-1"}}` {
		t.Fatalf("unknown value adapter = %#v", value)
	}
}

func TestParameterEntityNameUsesGroupKeyWithoutDescription(t *testing.T) {
	if got := ParameterEntityName("WEB", "banner"); got != "Parameter: WEB / banner" {
		t.Fatalf("ParameterEntityName() = %q", got)
	}
	if got := ParameterEntityName("", "ungrouped"); got != "Parameter: ungrouped" {
		t.Fatalf("ParameterEntityName() = %q", got)
	}
}

func TestParameterChangePreservesMovedParameterIdentity(t *testing.T) {
	before := &firebase.RemoteConfigParam{ValueType: "STRING", DefaultValue: &firebase.RemoteConfigValue{Value: "old"}}
	after := &firebase.RemoteConfigParam{ValueType: "STRING", DefaultValue: &firebase.RemoteConfigValue{Value: "new"}}
	input := ParameterChange(rcdiff.ParameterChange{
		Key: "new_name", Group: "NEW", PreviousKey: "old_name", PreviousGroup: "OLD",
		Kind: rcdiff.ChangeChanged, Current: before, Final: after,
	}, "Earlier version: v1", "Later version: v2")

	if input.EntityName != "Parameter: NEW / new_name" ||
		input.Left.Name != "Earlier version: v1" || input.Right.Name != "Later version: v2" {
		t.Fatalf("parameter change identity = %#v", input)
	}
	if input.Left.Properties["group"].Raw != "OLD" || input.Right.Properties["group"].Raw != "NEW" ||
		input.Left.Properties["name"].Raw != "old_name" || input.Right.Properties["name"].Raw != "new_name" {
		t.Fatalf("parameter move properties = left:%#v right:%#v", input.Left.Properties, input.Right.Properties)
	}
}
