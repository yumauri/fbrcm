package shared

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestCollectParamTargetsSortsByKeyThenGroup(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"beta":  {DefaultValue: &firebase.RemoteConfigValue{Value: "root-beta"}},
			"alpha": {DefaultValue: &firebase.RemoteConfigValue{Value: "root-alpha"}},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-b": {Parameters: map[string]firebase.RemoteConfigParam{"alpha": {DefaultValue: &firebase.RemoteConfigValue{Value: "group-alpha"}}}},
		},
	}

	got := CollectParamTargets(cfg)
	want := []struct {
		key   string
		group string
	}{
		{key: "alpha"},
		{key: "alpha", group: "group-b"},
		{key: "beta"},
	}
	if len(got) != len(want) {
		t.Fatalf("target count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Key != want[i].key || got[i].Group != want[i].group {
			t.Fatalf("target[%d] = (%q, %q), want (%q, %q)", i, got[i].Key, got[i].Group, want[i].key, want[i].group)
		}
	}
}

func TestRemoveParamSlotPreservesEmptyGroup(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {Description: "metadata", Parameters: map[string]firebase.RemoteConfigParam{"flag": {}}},
		},
	}

	RemoveParamSlot(cfg, "flag", "group-a")
	group, ok := cfg.ParameterGroups["group-a"]
	if !ok {
		t.Fatal("empty group was removed")
	}
	if group.Description != "metadata" || group.Parameters != nil {
		t.Fatalf("group = %#v, want preserved metadata and nil parameters", group)
	}
}

func TestSetParamSlotCreatesGroup(t *testing.T) {
	cfg := &firebase.RemoteConfig{}

	SetParamSlot(cfg, "flag", "group-a", firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: "on"}})

	group, ok := cfg.ParameterGroups["group-a"]
	if !ok {
		t.Fatalf("group-a was not created")
	}
	if group.Parameters["flag"].DefaultValue.Value != "on" {
		t.Fatalf("flag value = %q, want on", group.Parameters["flag"].DefaultValue.Value)
	}
}

func TestParamExistsFindsRootAndGroupedParams(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{"root": {}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {Parameters: map[string]firebase.RemoteConfigParam{"grouped": {}}},
		},
	}

	if !ParamExists(cfg, "root") {
		t.Fatalf("root parameter not found")
	}
	if !ParamExists(cfg, "grouped") {
		t.Fatalf("grouped parameter not found")
	}
	if ParamExists(cfg, "missing") {
		t.Fatalf("missing parameter found")
	}
}

func TestParamSlotExists(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{"root": {}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {Parameters: map[string]firebase.RemoteConfigParam{"grouped": {}}},
		},
	}

	if !ParamSlotExists(cfg, "root", "") {
		t.Fatalf("root slot not found")
	}
	if !ParamSlotExists(cfg, "grouped", "group-a") {
		t.Fatalf("grouped slot not found")
	}
	if ParamSlotExists(cfg, "missing", "") {
		t.Fatalf("missing root slot found")
	}
}
