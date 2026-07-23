package diffinput

import (
	"testing"

	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
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

func TestParameterEntityNameUsesGroupKeyWithoutDescription(t *testing.T) {
	if got := ParameterEntityName("WEB", "banner"); got != "Property: WEB / banner" {
		t.Fatalf("ParameterEntityName() = %q", got)
	}
	if got := ParameterEntityName("", "ungrouped"); got != "Property: ungrouped" {
		t.Fatalf("ParameterEntityName() = %q", got)
	}
}
