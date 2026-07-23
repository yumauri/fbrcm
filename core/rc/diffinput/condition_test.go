package diffinput

import (
	"testing"

	"github.com/yumauri/fbrcm/core/dictdiff"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestConditionUsesTypedComparisonHints(t *testing.T) {
	properties := Condition(3, &firebase.RemoteConfigCondition{
		Name:       "Audience",
		Expression: "country == 'US'",
		TagColor:   "BLUE",
	})
	if properties["position"].Type != dictdiff.ValueNumber ||
		properties["position"].CompareAs != dictdiff.CompareEnum {
		t.Fatalf("position = %#v, want atomic number", properties["position"])
	}
	if properties["expression"].CompareAs != dictdiff.CompareString {
		t.Fatalf("expression = %#v, want string comparison", properties["expression"])
	}
	if properties["color"].CompareAs != dictdiff.CompareEnum {
		t.Fatalf("color = %#v, want enum comparison", properties["color"])
	}
}

func TestGroupDistinguishesAbsentFromEmptyDescription(t *testing.T) {
	if got := Group("", false); len(got) != 0 {
		t.Fatalf("absent Group() = %#v, want empty dictionary", got)
	}
	got := Group("", true)
	if value, ok := got["description"]; !ok || value.Raw != "" {
		t.Fatalf("present Group() = %#v, want empty description property", got)
	}
}
