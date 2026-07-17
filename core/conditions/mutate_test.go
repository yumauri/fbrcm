package conditions

import (
	"slices"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestDefinitionMutationsPreservePriorityAndReferences(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "first", Expression: "true"},
			{Name: "second", Expression: "false"},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {ConditionalValues: map[string]firebase.RemoteConfigValue{"second": {Value: "yes"}}},
		},
	}
	if err := Add(cfg, Definition{Name: "middle", Expression: " app.version == '2' ", Description: " Mobile users ", TagColor: "green"}, 2); err != nil {
		t.Fatal(err)
	}
	if got := conditionNames(cfg); !slices.Equal(got, []string{"first", "middle", "second"}) {
		t.Fatalf("after add = %v", got)
	}
	if cfg.Conditions[1].Expression != "app.version == '2'" || cfg.Conditions[1].Description != "Mobile users" || cfg.Conditions[1].TagColor != "GREEN" {
		t.Fatalf("normalized condition = %#v", cfg.Conditions[1])
	}
	if err := Rename(cfg, "second", "renamed"); err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Parameters["flag"].ConditionalValues["renamed"]; !ok {
		t.Fatalf("conditional values were not renamed: %#v", cfg.Parameters["flag"].ConditionalValues)
	}
	if err := Move(cfg, "renamed", 1); err != nil {
		t.Fatal(err)
	}
	if got := conditionNames(cfg); !slices.Equal(got, []string{"renamed", "first", "middle"}) {
		t.Fatalf("after move = %v", got)
	}
}

func TestEditDetailsAtomicallyUpdatesDefinitionPriorityAndReferences(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "first", Expression: "true"},
			{Name: "second", Expression: "false", Description: "old", TagColor: "BLUE"},
			{Name: "third", Expression: "true"},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {ConditionalValues: map[string]firebase.RemoteConfigValue{"second": {Value: "yes"}}},
		},
	}
	err := EditDetails(cfg, DetailsEdit{
		Name:            "second",
		NextName:        "renamed",
		NextExpression:  " app.version == '2' ",
		NextDescription: " Employees ",
		NextTagColor:    "green",
		NextPriority:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := conditionNames(cfg); !slices.Equal(got, []string{"renamed", "first", "third"}) {
		t.Fatalf("condition order = %v", got)
	}
	if got := cfg.Conditions[0]; got.Expression != "app.version == '2'" || got.Description != "Employees" || got.TagColor != "GREEN" {
		t.Fatalf("edited condition = %#v", got)
	}
	values := cfg.Parameters["flag"].ConditionalValues
	if _, ok := values["renamed"]; !ok || len(values) != 1 {
		t.Fatalf("renamed conditional values = %#v", values)
	}
}

func TestEditDetailsUpdatesConditionalUsageWithoutDefinitionChange(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "staff", Expression: "true", TagColor: "GREEN"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"enabled": {
				ValueType: "BOOLEAN",
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"staff": {Value: "true"},
				},
			},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"checkout": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"limit": {
						ValueType: "NUMBER",
						ConditionalValues: map[string]firebase.RemoteConfigValue{
							"staff": {Value: "1"},
						},
					},
				},
			},
		},
	}
	err := EditDetails(cfg, DetailsEdit{
		Name: "staff", NextName: "staff", NextExpression: "true", NextTagColor: "GREEN", NextPriority: 1,
		ValueEdits: []UsageValueEdit{
			{GroupKey: "__default__", ParameterKey: "enabled", NextValue: "false"},
			{GroupKey: "checkout", ParameterKey: "limit", NextValue: "2"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Parameters["enabled"].ConditionalValues["staff"].Value; got != "false" {
		t.Fatalf("root conditional value = %q, want false", got)
	}
	if got := cfg.ParameterGroups["checkout"].Parameters["limit"].ConditionalValues["staff"].Value; got != "2" {
		t.Fatalf("group conditional value = %q, want 2", got)
	}
}

func TestEditDetailsRejectsInvalidUsageValueAtomically(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "staff", Expression: "true"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"enabled": {ValueType: "BOOLEAN", ConditionalValues: map[string]firebase.RemoteConfigValue{"staff": {Value: "true"}}},
			"limit":   {ValueType: "NUMBER", ConditionalValues: map[string]firebase.RemoteConfigValue{"staff": {Value: "1"}}},
		},
	}
	err := EditDetails(cfg, DetailsEdit{
		Name: "staff", NextName: "staff", NextExpression: "true", NextPriority: 1,
		ValueEdits: []UsageValueEdit{
			{ParameterKey: "enabled", NextValue: "false"},
			{ParameterKey: "limit", NextValue: "invalid"},
		},
	})
	if err == nil {
		t.Fatal("EditDetails accepted invalid number")
	}
	if got := cfg.Parameters["enabled"].ConditionalValues["staff"].Value; got != "true" {
		t.Fatalf("first value changed after later validation error: %q", got)
	}
}

func TestDeleteConditionCleansValuesAndPreservesEmptyGroup(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "remove", Expression: "true"}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"metadata": {
				Description: "keep",
				Parameters: map[string]firebase.RemoteConfigParam{
					"only": {ConditionalValues: map[string]firebase.RemoteConfigValue{"remove": {Value: "x"}}},
				},
			},
		},
	}
	if err := Delete(cfg, "remove"); err != nil {
		t.Fatal(err)
	}
	group, ok := cfg.ParameterGroups["metadata"]
	if !ok || group.Description != "keep" || group.Parameters != nil {
		t.Fatalf("description-only group was not preserved: %#v", cfg.ParameterGroups)
	}
}

func TestConditionDefinitionValidation(t *testing.T) {
	if _, err := NormalizeName(""); err == nil {
		t.Fatal("empty condition name accepted")
	}
	if _, err := NormalizeName(string(make([]rune, MaxNameLength+1))); err == nil {
		t.Fatal("long condition name accepted")
	}
	if _, err := NormalizeExpression("  "); err == nil {
		t.Fatal("empty condition expression accepted")
	}
	if got, err := NormalizeTagColor("deep_orange"); err != nil || got != "DEEP_ORANGE" {
		t.Fatalf("NormalizeTagColor = %q, %v", got, err)
	}
	if _, err := NormalizeTagColor("red"); err == nil {
		t.Fatal("unsupported color accepted")
	}
}

func TestResolveNameUsesExactThenCaseInsensitiveMatching(t *testing.T) {
	cfg := &firebase.RemoteConfig{Conditions: []firebase.RemoteConfigCondition{{Name: "Beta Users"}}}
	if got, ok := ResolveName(cfg, "Beta Users"); !ok || got != "Beta Users" {
		t.Fatalf("exact ResolveName = %q, %v", got, ok)
	}
	if got, ok := ResolveName(cfg, "beta users"); !ok || got != "Beta Users" {
		t.Fatalf("case-insensitive ResolveName = %q, %v", got, ok)
	}
}

func conditionNames(cfg *firebase.RemoteConfig) []string {
	names := make([]string, len(cfg.Conditions))
	for index, condition := range cfg.Conditions {
		names[index] = condition.Name
	}
	return names
}
