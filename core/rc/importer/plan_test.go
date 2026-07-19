package importer

import (
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestParseSourceRecognizesCacheAndSummarizes(t *testing.T) {
	source, err := ParseSource([]byte(`{"remote_config":{"conditions":[{"name":"ios","expression":"device.os == 'ios'"}],"parameterGroups":{"empty":{"description":"keep me"}},"parameters":{"flag":{"defaultValue":{"value":"on"}}}}}`))
	if err != nil {
		t.Fatalf("ParseSource = %v", err)
	}
	summary := Summarize(source.Config, source.WrappedCache)
	if !summary.WrappedCache || summary.Parameters() != 1 || summary.Groups != 1 || summary.Conditions != 1 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestSummarizeClassifiesPortableConditions(t *testing.T) {
	cfg := &firebase.RemoteConfig{Conditions: []firebase.RemoteConfigCondition{
		{Name: "portable_percent", Expression: "percent <= 10"},
		{Name: "portable_first_open", Expression: "app.firstOpenTimestamp > ('2026-01-01T00:00:00Z')"},
		{Name: "audience", Expression: "app.audiences.inAtLeastOne(['paid'])"},
		{Name: "property", Expression: "app.userProperty['tier'] == 'paid'"},
		{Name: "app", Expression: "app.id == '1:123:ios:abc'"},
		{Name: "signal", Expression: "app.customSignal['tier'] == 'paid'"},
		{Name: "installation", Expression: "app.firebaseInstallationId in ['abc']"},
		{Name: "experiment", Expression: "inExperiment('checkout')"},
	}}
	summary := Summarize(cfg, false)
	if summary.Conditions != 8 || summary.PortableConditions() != 2 || summary.NonPortableConditions != 6 {
		t.Fatalf("condition summary = %+v", summary)
	}
}

func TestParseSourceNormalizesConditionsForFirebaseUpdate(t *testing.T) {
	source, err := ParseSource([]byte(`{"conditions":[{"name":"staff","expression":"true","tagColor":"deep_orange"}]}`))
	if err != nil {
		t.Fatalf("ParseSource = %v", err)
	}
	condition := source.Config.Conditions[0]
	if condition.TagColor != "DEEP_ORANGE" {
		t.Fatalf("normalized condition = %#v", condition)
	}

	_, err = ParseSource([]byte(`{"conditions":[{"name":"staff","expression":"true","description":"unsupported"}]}`))
	if err == nil || !strings.Contains(err.Error(), `unknown field "description"`) {
		t.Fatalf("unsupported condition field error = %v", err)
	}

	_, err = ParseSource([]byte(`{"conditions":[{"name":"staff","expression":"true","tagColor":"RED"}]}`))
	if err == nil || !strings.Contains(err.Error(), `condition "staff"`) {
		t.Fatalf("unsupported color error = %v", err)
	}
}

func TestBuildPlanMergeDiscoversAndResolvesConflicts(t *testing.T) {
	current := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "current"}},
	}}
	source := &ParsedSource{Config: &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "import"}},
		"new":  {DefaultValue: &firebase.RemoteConfigValue{Value: "added"}},
	}}}
	plan, err := BuildPlan("demo", "Demo", current, source, Options{Strategy: StrategyMerge, DefaultResolution: ResolutionCurrent})
	if err != nil {
		t.Fatalf("BuildPlan = %v", err)
	}
	if len(plan.Conflicts) != 1 {
		t.Fatalf("conflicts = %+v", plan.Conflicts)
	}
	if got := plan.Final.Parameters["flag"].DefaultValue.Value; got != "current" {
		t.Fatalf("current resolution value = %q", got)
	}
	resolved, err := BuildPlan("demo", "Demo", current, source, Options{
		Strategy:    StrategyMerge,
		Resolutions: map[string]Resolution{plan.Conflicts[0].ID: ResolutionImport},
	})
	if err != nil {
		t.Fatalf("resolved BuildPlan = %v", err)
	}
	if got := resolved.Final.Parameters["flag"].DefaultValue.Value; got != "import" {
		t.Fatalf("import resolution value = %q", got)
	}
	if got := resolved.Final.Parameters["new"].DefaultValue.Value; got != "added" {
		t.Fatalf("added value = %q", got)
	}
}

func TestBuildPlanPreservesDescriptionOnlyGroup(t *testing.T) {
	current := &firebase.RemoteConfig{}
	source := &ParsedSource{Config: &firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{
		"empty": {Description: "keep me"},
	}}}
	plan, err := BuildPlan("demo", "Demo", current, source, Options{Strategy: StrategyReplace})
	if err != nil {
		t.Fatalf("BuildPlan = %v", err)
	}
	group, ok := plan.Final.ParameterGroups["empty"]
	if !ok || group.Description != "keep me" {
		t.Fatalf("empty group = %+v, %v", group, ok)
	}
}

func TestTransformAppliesScopeAndConditionPolicy(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "audience", Expression: "inUserAudience('paid')"},
			{Name: "platform", Expression: "device.os == 'ios'"},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"rollout": {Parameters: map[string]firebase.RemoteConfigParam{
				"feature": {DefaultValue: &firebase.RemoteConfigValue{Value: "on"}, ConditionalValues: map[string]firebase.RemoteConfigValue{"audience": {Value: "off"}, "platform": {Value: "ios"}}},
			}},
			"other": {Parameters: map[string]firebase.RemoteConfigParam{"skip": {DefaultValue: &firebase.RemoteConfigValue{Value: "x"}}}},
		},
	}
	err := Transform("demo", "Demo", cfg, Options{Groups: []string{"rollout"}, Search: "feature", ConditionPolicy: ConditionPolicyKeepPortableOnly})
	if err != nil {
		t.Fatalf("Transform = %v", err)
	}
	if _, ok := cfg.ParameterGroups["other"]; ok {
		t.Fatal("unselected group retained")
	}
	param := cfg.ParameterGroups["rollout"].Parameters["feature"]
	if _, ok := param.ConditionalValues["audience"]; ok {
		t.Fatal("non-portable conditional value retained")
	}
	if _, ok := param.ConditionalValues["platform"]; !ok {
		t.Fatal("portable conditional value removed")
	}
}
