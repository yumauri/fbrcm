package promote

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

func TestApplyDefaultAddsAndUpdatesWithoutRemovingTargetOnly(t *testing.T) {
	source := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"add":    stringParam("source-add"),
			"change": stringParam("source-change"),
		},
	}
	target := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"change":      stringParam("target-change"),
			"target_only": stringParam("keep"),
		},
	}

	plan := BuildPlan(source, target, Options{})
	finalCfg, _, err := Apply(plan, SelectAll(plan.Items), Options{})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if got := finalCfg.Parameters["add"].DefaultValue.Value; got != "source-add" {
		t.Fatalf("add = %q, want source-add", got)
	}
	if got := finalCfg.Parameters["change"].DefaultValue.Value; got != "source-change" {
		t.Fatalf("change = %q, want source-change", got)
	}
	if got := finalCfg.Parameters["target_only"].DefaultValue.Value; got != "keep" {
		t.Fatalf("target_only = %q, want keep", got)
	}
}

func TestApplyPruneRemovesTargetOnlyParameter(t *testing.T) {
	source := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{}}
	target := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{"target_only": stringParam("remove")}}

	plan := BuildPlan(source, target, Options{Prune: true})
	finalCfg, _, err := Apply(plan, SelectAll(plan.Items), Options{Prune: true})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if _, ok := finalCfg.Parameters["target_only"]; ok {
		t.Fatalf("target_only still exists after prune")
	}
}

func TestApplySelectedParameterPullsRequiredCondition(t *testing.T) {
	source := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "beta", Expression: "app.version == '2'"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "off"},
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"beta": {Value: "on"},
				},
			},
		},
	}
	target := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{}}

	plan := BuildPlan(source, target, Options{})
	selected := map[ItemID]bool{{Kind: rcdiff.ItemParameter, Name: "flag"}: true}
	finalCfg, applied, err := Apply(plan, selected, Options{})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if len(finalCfg.Conditions) != 1 || finalCfg.Conditions[0].Name != "beta" {
		t.Fatalf("conditions = %#v, want required beta", finalCfg.Conditions)
	}
	if len(applied) != 2 {
		t.Fatalf("applied length = %d, want parameter plus required condition", len(applied))
	}
}

func TestApplySelectedGroupedParameterPullsGroupDescription(t *testing.T) {
	source := &firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"onboarding": {
				Description: "source description",
				Parameters: map[string]firebase.RemoteConfigParam{
					"flag": stringParam("on"),
				},
			},
		},
	}
	target := &firebase.RemoteConfig{}

	plan := BuildPlan(source, target, Options{})
	selected := map[ItemID]bool{{Kind: rcdiff.ItemParameter, Name: "flag", Group: "onboarding"}: true}
	finalCfg, _, err := Apply(plan, selected, Options{})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if got := finalCfg.ParameterGroups["onboarding"].Description; got != "source description" {
		t.Fatalf("group description = %q, want source description", got)
	}
}

func stringParam(value string) firebase.RemoteConfigParam {
	return firebase.RemoteConfigParam{
		DefaultValue: &firebase.RemoteConfigValue{Value: value},
		ValueType:    "STRING",
	}
}
