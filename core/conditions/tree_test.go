package conditions

import (
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rootgroup"
)

func TestBuildTreePreservesConditionPriorityAndBuildsUsage(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Version: firebase.RemoteConfigVersion{VersionNumber: "42"},
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "beta", Expression: "app.version > '2'", TagColor: "BLUE"},
			{Name: "staff", Expression: "user in staff", Description: " Employees "},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"welcome": {ValueType: "STRING", ConditionalValues: map[string]firebase.RemoteConfigValue{
				"staff": {Value: "Hello team"},
			}},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"checkout": {Parameters: map[string]firebase.RemoteConfigParam{
				"button": {ValueType: "BOOLEAN", ConditionalValues: map[string]firebase.RemoteConfigValue{
					"beta": {Value: "true"},
				}},
			}},
		},
	}

	tree := BuildTree(cfg, time.Unix(123, 0), "etag")
	if got := []string{tree.Conditions[0].Name, tree.Conditions[1].Name}; got[0] != "beta" || got[1] != "staff" {
		t.Fatalf("condition order = %v", got)
	}
	if tree.Conditions[0].Priority != 1 || tree.Conditions[1].Priority != 2 {
		t.Fatalf("priorities = %d, %d", tree.Conditions[0].Priority, tree.Conditions[1].Priority)
	}
	if tree.Conditions[1].Description != "Employees" {
		t.Fatalf("trimmed description = %q", tree.Conditions[1].Description)
	}
	staffUsage := tree.Conditions[1].Usages
	if len(staffUsage) != 1 || staffUsage[0].GroupKey != rootgroup.TreeKey || staffUsage[0].ParameterKey != "welcome" {
		t.Fatalf("staff usage = %#v", staffUsage)
	}
	betaUsage := tree.Conditions[0].Usages
	if len(betaUsage) != 1 || betaUsage[0].GroupKey != "checkout" || betaUsage[0].Value != "true" || betaUsage[0].ValueType != "BOOLEAN" {
		t.Fatalf("beta usage = %#v", betaUsage)
	}
}

func TestDeleteImpactReportsValuesAndParametersThatBecomeEmpty(t *testing.T) {
	tree := BuildTree(&firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "staff"}, {Name: "beta"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"removed":           {ConditionalValues: map[string]firebase.RemoteConfigValue{"staff": {Value: "x"}}},
			"kept_by_default":   {DefaultValue: &firebase.RemoteConfigValue{Value: "x"}, ConditionalValues: map[string]firebase.RemoteConfigValue{"staff": {Value: "y"}}},
			"kept_by_condition": {ConditionalValues: map[string]firebase.RemoteConfigValue{"staff": {Value: "x"}, "beta": {Value: "y"}}},
		},
	}, time.Time{}, "")

	impact, err := tree.DeleteImpact("staff")
	if err != nil {
		t.Fatal(err)
	}
	if len(impact.Usages) != 3 {
		t.Fatalf("usage count = %d, want 3", len(impact.Usages))
	}
	if len(impact.RemovedParameters) != 1 || impact.RemovedParameters[0].ParameterKey != "removed" {
		t.Fatalf("removed parameters = %#v", impact.RemovedParameters)
	}
}

func TestMoveImpactReportsOnlyParametersUsingMovedAndCrossedConditions(t *testing.T) {
	tree := BuildTree(&firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"affected_b":  {ConditionalValues: map[string]firebase.RemoteConfigValue{"d": {}, "b": {}}},
			"affected_c":  {ConditionalValues: map[string]firebase.RemoteConfigValue{"d": {}, "c": {}}},
			"not_crossed": {ConditionalValues: map[string]firebase.RemoteConfigValue{"d": {}, "a": {}}},
			"no_moved":    {ConditionalValues: map[string]firebase.RemoteConfigValue{"b": {}, "c": {}}},
		},
	}, time.Time{}, "")

	impact, err := tree.MoveImpact("d", 2)
	if err != nil {
		t.Fatal(err)
	}
	if got := impact.CrossedConditions; len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Fatalf("crossed conditions = %v", got)
	}
	if got := impact.AffectedParameters; len(got) != 2 || got[0].ParameterKey != "affected_b" || got[1].ParameterKey != "affected_c" {
		t.Fatalf("affected parameters = %#v", got)
	}
	if _, err := tree.MoveImpact("d", 0); err == nil {
		t.Fatal("expected invalid priority error")
	}
}
