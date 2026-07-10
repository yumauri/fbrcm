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

func TestRemoveParamSlotDropsEmptyGroup(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"g": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"only": {ValueType: "STRING"},
				},
			},
		},
	}
	rcmutate.RemoveParamSlot(cfg, "only", "g")
	if _, ok := cfg.ParameterGroups["g"]; ok {
		t.Fatalf("empty group still present: %#v", cfg.ParameterGroups)
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
	if _, ok := cfg.ParameterGroups["empty-group"]; ok {
		t.Fatalf("empty-group still present")
	}
	rootKeep := cfg.Parameters["root_keep"]
	if _, ok := rootKeep.ConditionalValues["remove"]; ok {
		t.Fatalf("root_keep still references remove condition")
	}
	if _, ok := rootKeep.ConditionalValues["keep"]; !ok {
		t.Fatalf("root_keep lost keep conditional")
	}
}

func TestRemoveEmptyGroups(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"empty": {Parameters: map[string]firebase.RemoteConfigParam{}},
		},
	}
	rcmutate.RemoveEmptyGroups(cfg)
	if cfg.ParameterGroups != nil {
		t.Fatalf("ParameterGroups = %#v, want nil", cfg.ParameterGroups)
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
			rcmutate.RemoveEmptyGroups(cfg)

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
