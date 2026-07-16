package draft

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

func TestMutationOperations(t *testing.T) {
	tests := []struct {
		name   string
		cache  json.RawMessage
		spec   MutationSpec
		assert func(*testing.T, json.RawMessage)
	}{
		{
			name:  "delete parameter",
			cache: remoteConfigRaw("1", map[string]string{"flag": "old"}),
			spec:  MutationSpec{Apply: DeleteParameter("", "flag")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertFlagMissing(t, raw)
			},
		},
		{
			name:  "delete group",
			cache: groupedRemoteConfigRaw("1", "group", map[string]string{"flag": "old"}),
			spec:  MutationSpec{Apply: DeleteGroup("group")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertGroupMissing(t, raw, "group")
			},
		},
		{
			name:  "delete conditional value",
			cache: conditionalRemoteConfigRaw("1", "flag", "cond", "conditional"),
			spec:  MutationSpec{Apply: DeleteConditionalValue("", "flag", "cond")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertConditionalMissing(t, raw, "flag", "cond")
			},
		},
		{
			name:  "rename parameter",
			cache: remoteConfigRaw("1", map[string]string{"flag": "old"}),
			spec:  MutationSpec{Apply: RenameParameter("", "flag", "renamed")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertFlagMissing(t, raw)
				assertParamValue(t, raw, "renamed", "old")
			},
		},
		{
			name:  "rename group",
			cache: groupedRemoteConfigRaw("1", "old_group", map[string]string{"flag": "old"}),
			spec:  MutationSpec{Apply: RenameGroup("old_group", "new_group")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertGroupMissing(t, raw, "old_group")
				assertGroupParamValue(t, raw, "new_group", "flag", "old")
			},
		},
		{
			name:  "move parameter into group",
			cache: remoteConfigRaw("1", map[string]string{"flag": "old"}),
			spec:  MutationSpec{Apply: MoveParameter("", "flag", "group")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertFlagMissing(t, raw)
				assertGroupParamValue(t, raw, "group", "flag", "old")
			},
		},
		{
			name:  "move group to root",
			cache: groupedRemoteConfigRaw("1", "group", map[string]string{"flag": "old"}),
			spec:  MutationSpec{Apply: MoveGroup("group", "")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertGroupMissing(t, raw, "group")
				assertParamValue(t, raw, "flag", "old")
			},
		},
		{
			name:  "set boolean value",
			cache: typedRemoteConfigRaw("1", "flag", "false", "BOOLEAN"),
			spec:  MutationSpec{Apply: SetBooleanParameterValue("", "flag", "default", true)},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertParamValue(t, raw, "flag", "true")
			},
		},
		{
			name:  "set number value",
			cache: typedRemoteConfigRaw("1", "flag", "1", "NUMBER"),
			spec:  MutationSpec{Apply: SetNumberParameterValue("", "flag", "default", "2")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertParamValue(t, raw, "flag", "2")
			},
		},
		{
			name:  "set string value",
			cache: remoteConfigRaw("1", map[string]string{"flag": "old"}),
			spec:  MutationSpec{Apply: SetStringParameterValue("", "flag", "default", "new")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertParamValue(t, raw, "flag", "new")
			},
		},
		{
			name:  "set json value",
			cache: typedRemoteConfigRaw("1", "flag", `{"a":1}`, "JSON"),
			spec:  MutationSpec{Apply: SetJSONParameterValue("", "flag", "default", `{"b":2}`)},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertParamValue(t, raw, "flag", `{"b":2}`)
			},
		},
		{
			name:  "duplicate parameter with provided name",
			cache: remoteConfigRaw("1", map[string]string{"flag": "old"}),
			spec:  MutationSpec{Apply: DuplicateParameterNamed("", "flag", "copy")},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertParamValue(t, raw, "flag", "old")
				assertParamValue(t, raw, "copy", "old")
			},
		},
		{
			name:  "create parameter details",
			cache: remoteConfigRaw("1", map[string]string{}),
			spec: MutationSpec{Apply: EditParameterDetails(ParameterDetailsEdit{
				Create:        true,
				NextParamKey:  "created",
				NextValueType: "STRING",
				ValueEdits:    []ParameterValueEdit{{Label: "default", NextValue: "value"}},
			})},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertParamValue(t, raw, "created", "value")
			},
		},
		{
			name:  "edit parameter details",
			cache: remoteConfigRaw("1", map[string]string{"flag": "old"}),
			spec: MutationSpec{Apply: EditParameterDetails(ParameterDetailsEdit{
				GroupKey:        "",
				ParamKey:        "flag",
				NextParamKey:    "flag",
				NextValueType:   "STRING",
				NextDescription: "updated",
				ValueEdits:      []ParameterValueEdit{{Label: "default", NextValue: "new"}},
			})},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertParamValue(t, raw, "flag", "new")
				assertParamDescription(t, raw, "flag", "updated")
			},
		},
		{
			name:  "add conditional value from parameter details",
			cache: conditionOnlyRemoteConfigRaw("1", "flag", "staff"),
			spec: MutationSpec{Apply: EditParameterDetails(ParameterDetailsEdit{
				GroupKey:      "",
				ParamKey:      "flag",
				NextParamKey:  "flag",
				NextValueType: "STRING",
				ValueEdits:    []ParameterValueEdit{{Label: "STAFF", NextValue: "assigned"}},
			})},
			assert: func(t *testing.T, raw json.RawMessage) {
				assertConditionalValue(t, raw, "flag", "staff", "assigned")
				assertParamValue(t, raw, "flag", "default")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDraftTestEnv(t)
			cache := saveParametersCache(t, "demo", "etag-1", tt.cache)
			deps := (&fakeDeps{cache: cache}).deps()

			_, hasDraft, err := Mutate(context.Background(), deps, "demo", false, tt.spec)
			if err != nil {
				t.Fatalf("Mutate returned error: %v", err)
			}
			if !hasDraft {
				t.Fatal("hasDraft = false, want true")
			}

			raw, loaded := loadDraft(t, "demo")
			if !loaded {
				t.Fatal("draft not saved")
			}
			tt.assert(t, raw)
		})
	}
}

func TestDuplicateParameterAutoNamed(t *testing.T) {
	setupDraftTestEnv(t)
	cache := saveParametersCache(t, "demo", "etag-1", remoteConfigRaw("1", map[string]string{"flag": "old"}))
	apply, nextKey := DuplicateParameterAutoNamed("", "flag")
	deps := (&fakeDeps{cache: cache}).deps()

	_, hasDraft, err := Mutate(context.Background(), deps, "demo", false, MutationSpec{Apply: apply})
	if err != nil {
		t.Fatalf("Mutate returned error: %v", err)
	}
	if !hasDraft {
		t.Fatal("hasDraft = false, want true")
	}
	if got := nextKey(); got != "flag_copy" {
		t.Fatalf("duplicate key = %q, want flag_copy", got)
	}

	raw, loaded := loadDraft(t, "demo")
	if !loaded {
		t.Fatal("draft not saved")
	}
	assertParamValue(t, raw, "flag", "old")
	assertParamValue(t, raw, "flag_copy", "old")
}

func TestMutationPreservesEmptyGroups(t *testing.T) {
	setupDraftTestEnv(t)
	baseRaw := marshalRemoteConfig(firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": remoteConfigParam("old", "STRING"),
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"empty": {},
			"ROKU":  {Description: "FLAGS FOR ROKU"},
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: "1"},
	})
	cache := saveParametersCache(t, "demo", "etag-1", baseRaw)
	deps := (&fakeDeps{cache: cache}).deps()

	result, _, err := Mutate(context.Background(), deps, "demo", false, MutationSpec{
		Apply: SetStringParameterValue("", "flag", "default", "new"),
	})
	if err != nil {
		t.Fatalf("Mutate returned error: %v", err)
	}
	assertPreservedEmptyGroups(t, result.FinalRaw)

	stored, ok, err := LoadRecord("demo")
	if err != nil || !ok {
		t.Fatalf("LoadRecord ok = %v, err = %v", ok, err)
	}
	assertPreservedEmptyGroups(t, stored.BaseRemoteConfig)
	assertPreservedEmptyGroups(t, stored.RemoteConfig)
	baseCfg, _ := firebase.ParseRemoteConfig(stored.BaseRemoteConfig)
	draftCfg, _ := firebase.ParseRemoteConfig(stored.RemoteConfig)
	groupChanges := rcdiff.CompareRemoteConfigs(baseCfg, draftCfg).GroupDescriptionSummary()
	if groupChanges.Added != 0 || groupChanges.Removed != 0 || groupChanges.Changed != 0 {
		t.Fatalf("group description changes = %+v, want no changes", groupChanges)
	}
}

func TestDeleteLastParameterPreservesGroup(t *testing.T) {
	setupDraftTestEnv(t)
	baseRaw := marshalRemoteConfig(firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group": {
				Description: "metadata",
				Parameters: map[string]firebase.RemoteConfigParam{
					"flag": remoteConfigParam("old", "STRING"),
				},
			},
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: "1"},
	})
	cache := saveParametersCache(t, "demo", "etag-1", baseRaw)
	deps := (&fakeDeps{cache: cache}).deps()

	result, _, err := Mutate(context.Background(), deps, "demo", false, MutationSpec{
		Apply: DeleteParameter("group", "flag"),
	})
	if err != nil {
		t.Fatalf("Mutate returned error: %v", err)
	}
	cfg, err := firebase.ParseRemoteConfig(result.FinalRaw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	group, ok := cfg.ParameterGroups["group"]
	if !ok {
		t.Fatal("group was removed with its last parameter")
	}
	if group.Description != "metadata" || len(group.Parameters) != 0 {
		t.Fatalf("group = %#v, want preserved empty group with metadata", group)
	}
}

func assertPreservedEmptyGroups(t *testing.T, raw json.RawMessage) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	if _, ok := cfg.ParameterGroups["empty"]; !ok {
		t.Fatal("empty group was removed")
	}
	group, ok := cfg.ParameterGroups["ROKU"]
	if !ok || group.Description != "FLAGS FOR ROKU" || len(group.Parameters) != 0 {
		t.Fatalf("ROKU group = %#v, ok = %v; want preserved description-only group", group, ok)
	}
}
