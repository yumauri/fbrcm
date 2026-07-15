package mutate_test

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
)

func TestSlotKeyRoundTrip(t *testing.T) {
	tests := []struct {
		group string
		param string
	}{
		{"", "feature_login"},
		{"checkout", "tax_rate"},
	}
	for _, tc := range tests {
		key := rcmutate.SlotKey(tc.group, tc.param)
		if got := rcmutate.SlotKeyGroup(key); got != tc.group {
			t.Fatalf("SlotKeyGroup(%q) = %q, want %q", key, got, tc.group)
		}
		if got := rcmutate.SlotKeyParam(key); got != tc.param {
			t.Fatalf("SlotKeyParam(%q) = %q, want %q", key, got, tc.param)
		}
	}
}

func TestSetParamSlotInitializesMaps(t *testing.T) {
	cfg := &firebase.RemoteConfig{}
	slot := rcmutate.Slot{
		Group: "group-a",
		Param: firebase.RemoteConfigParam{
			DefaultValue: &firebase.RemoteConfigValue{Value: "v"},
			ValueType:    "STRING",
		},
	}
	rcmutate.SetParamSlot(cfg, "flag", slot)

	if cfg.ParameterGroups == nil {
		t.Fatal("ParameterGroups is nil after setParamSlot")
	}
	if got := cfg.ParameterGroups["group-a"].Parameters["flag"].DefaultValue.Value; got != "v" {
		t.Fatalf("group param value = %q, want v", got)
	}

	rootSlot := rcmutate.Slot{
		Param: firebase.RemoteConfigParam{
			DefaultValue: &firebase.RemoteConfigValue{Value: "root"},
			ValueType:    "STRING",
		},
	}
	rcmutate.SetParamSlot(cfg, "root_flag", rootSlot)
	if cfg.Parameters == nil {
		t.Fatal("Parameters is nil after root setParamSlot")
	}
	if got := cfg.Parameters["root_flag"].DefaultValue.Value; got != "root" {
		t.Fatalf("root param value = %q, want root", got)
	}
}

func TestCollectParamSlotsUsesCompositeKeys(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"root": {ValueType: "STRING"},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"g": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"flag": {ValueType: "BOOLEAN"},
				},
			},
		},
	}
	slots := rcmutate.CollectParamSlots(cfg)
	if len(slots) != 2 {
		t.Fatalf("slot count = %d, want 2", len(slots))
	}
	if _, ok := slots[rcmutate.SlotKey("", "root")]; !ok {
		t.Fatalf("missing root slot key %q", rcmutate.SlotKey("", "root"))
	}
	if _, ok := slots[rcmutate.SlotKey("g", "flag")]; !ok {
		t.Fatalf("missing group slot key %q", rcmutate.SlotKey("g", "flag"))
	}
}

func TestRemoveParamSlotPreservesEmptyGroup(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"g": {
				Description: "metadata",
				Parameters: map[string]firebase.RemoteConfigParam{
					"only": {ValueType: "STRING"},
				},
			},
		},
	}
	rcmutate.RemoveParamSlot(cfg, "only", "g")
	group, ok := cfg.ParameterGroups["g"]
	if !ok {
		t.Fatalf("empty group was removed: %#v", cfg.ParameterGroups)
	}
	if group.Description != "metadata" || group.Parameters != nil {
		t.Fatalf("group = %#v, want preserved description and nil parameters", group)
	}
}

func TestDropUnknownConditionReferences(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "keep"},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"root_keep": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "default"},
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"keep":   {Value: "yes"},
					"remove": {Value: "no"},
				},
			},
			"root_drop": {
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"remove": {Value: "drop"},
				},
			},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"empty-group": {
				Description: "keep metadata",
				Parameters: map[string]firebase.RemoteConfigParam{
					"group_drop": {
						ConditionalValues: map[string]firebase.RemoteConfigValue{
							"remove": {Value: "drop"},
						},
					},
				},
			},
		},
	}

	rcmutate.DropUnknownConditionReferences(cfg)

	if _, ok := cfg.Parameters["root_drop"]; ok {
		t.Fatalf("root_drop still present")
	}
	group, ok := cfg.ParameterGroups["empty-group"]
	if !ok {
		t.Fatal("empty-group was removed")
	}
	if group.Description != "keep metadata" || group.Parameters != nil {
		t.Fatalf("empty-group = %#v, want preserved metadata and nil parameters", group)
	}
	rootKeep := cfg.Parameters["root_keep"]
	if _, ok := rootKeep.ConditionalValues["remove"]; ok {
		t.Fatalf("root_keep still references remove condition")
	}
	if _, ok := rootKeep.ConditionalValues["keep"]; !ok {
		t.Fatalf("root_keep lost keep conditional")
	}
}

func TestNormalizeEmptyParameterMapsPreservesGroups(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"empty": {Description: "metadata", Parameters: map[string]firebase.RemoteConfigParam{}},
		},
	}
	rcmutate.NormalizeEmptyParameterMaps(cfg)
	group, ok := cfg.ParameterGroups["empty"]
	if !ok {
		t.Fatalf("ParameterGroups = %#v, want empty group preserved", cfg.ParameterGroups)
	}
	if group.Description != "metadata" || group.Parameters != nil {
		t.Fatalf("group = %#v, want preserved metadata and nil parameters", group)
	}
	if cfg.Parameters != nil {
		t.Fatalf("Parameters = %#v, want nil", cfg.Parameters)
	}
}

func TestFixtureRoundTripMutateNormalize(t *testing.T) {
	fixtures := []string{
		"root_params.json",
		"grouped_params.json",
		"with_conditions.json",
		"empty_values.json",
	}

	dir := fixtureDir(t)
	for _, name := range fixtures {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			cfg, err := firebase.ParseRemoteConfig(raw)
			if err != nil {
				t.Fatalf("ParseRemoteConfig: %v", err)
			}

			rcmutate.DropUnknownConditionReferences(cfg)
			rcmutate.NormalizeEmptyParameterMaps(cfg)

			out, err := firebase.MarshalRemoteConfig(cfg)
			if err != nil {
				t.Fatalf("MarshalRemoteConfig: %v", err)
			}
			roundTrip, err := firebase.ParseRemoteConfig(out)
			if err != nil {
				t.Fatalf("round-trip parse: %v", err)
			}
			if !reflect.DeepEqual(normalizeRemoteConfig(cfg), normalizeRemoteConfig(roundTrip)) {
				t.Fatal("round-trip config differs after cleanup")
			}
		})
	}
}

func normalizeRemoteConfig(cfg *firebase.RemoteConfig) *firebase.RemoteConfig {
	if cfg == nil {
		return nil
	}
	clone, err := firebase.CloneRemoteConfig(cfg)
	if err != nil {
		panic(err)
	}
	return clone
}

func fixtureDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "testdata", "remoteconfig")
}
