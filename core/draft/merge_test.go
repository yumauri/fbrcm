package draft

import (
	"encoding/json"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestMergeWithLatest(t *testing.T) {
	tests := []struct {
		name       string
		base       json.RawMessage
		draft      json.RawMessage
		latest     json.RawMessage
		wantChange bool
		wantValue  string
		wantErr    bool
		wantAbsent bool
	}{
		{
			name:       "keeps local change when remote unchanged",
			base:       remoteConfigRaw("1", map[string]string{"flag": "old"}),
			draft:      remoteConfigRaw("1", map[string]string{"flag": "local"}),
			latest:     remoteConfigRaw("2", map[string]string{"flag": "old"}),
			wantChange: true,
			wantValue:  "local",
		},
		{
			name:       "obsolete draft removed when only remote changed",
			base:       remoteConfigRaw("1", map[string]string{"flag": "old"}),
			draft:      remoteConfigRaw("1", map[string]string{"flag": "old"}),
			latest:     remoteConfigRaw("2", map[string]string{"flag": "remote"}),
			wantChange: false,
		},
		{
			name:    "reports conflict when local and remote both changed",
			base:    remoteConfigRaw("1", map[string]string{"flag": "old"}),
			draft:   remoteConfigRaw("1", map[string]string{"flag": "local"}),
			latest:  remoteConfigRaw("2", map[string]string{"flag": "remote"}),
			wantErr: true,
		},
		{
			name:       "keeps local deletion when remote unchanged",
			base:       remoteConfigRaw("1", map[string]string{"flag": "old"}),
			draft:      remoteConfigRaw("1", map[string]string{}),
			latest:     remoteConfigRaw("2", map[string]string{"flag": "old"}),
			wantChange: true,
			wantAbsent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := MergeWithLatest(tt.base, tt.draft, tt.latest)
			if tt.wantErr {
				if err == nil {
					t.Fatal("MergeWithLatest returned nil error")
				}
				return
			}
			if err != nil {
				t.Fatalf("MergeWithLatest returned error: %v", err)
			}
			if changed != tt.wantChange {
				t.Fatalf("changed = %v, want %v", changed, tt.wantChange)
			}
			if !changed {
				return
			}
			if tt.wantAbsent {
				assertFlagMissing(t, got)
				return
			}
			assertParamValue(t, got, "flag", tt.wantValue)
		})
	}
}

func TestMergeWithLatestMergesCompleteRemoteConfig(t *testing.T) {
	base := completeConfigRaw("1", "base description", []firebase.RemoteConfigCondition{{Name: "beta", Expression: "true"}}, "old")
	draftRaw := completeConfigRaw("1", "local description", []firebase.RemoteConfigCondition{{Name: "beta", Expression: "app.version == '2'"}}, "old")
	latest := completeConfigRaw("2", "base description", []firebase.RemoteConfigCondition{{Name: "beta", Expression: "true"}}, "remote")

	mergedRaw, changed, err := MergeWithLatest(base, draftRaw, latest)
	if err != nil {
		t.Fatalf("MergeWithLatest returned error: %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	merged, err := firebase.ParseRemoteConfig(mergedRaw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	if got := merged.ParameterGroups["group"].Description; got != "local description" {
		t.Fatalf("group description = %q", got)
	}
	if got := merged.Conditions[0].Expression; got != "app.version == '2'" {
		t.Fatalf("condition expression = %q", got)
	}
	assertParamValue(t, mergedRaw, "remote_flag", "remote")
}

func TestMergeWithLatestConditionConflict(t *testing.T) {
	base := completeConfigRaw("1", "description", []firebase.RemoteConfigCondition{{Name: "beta", Expression: "true"}}, "old")
	draftRaw := completeConfigRaw("1", "description", []firebase.RemoteConfigCondition{{Name: "beta", Expression: "local"}}, "old")
	latest := completeConfigRaw("2", "description", []firebase.RemoteConfigCondition{{Name: "beta", Expression: "remote"}}, "old")
	if _, _, err := MergeWithLatest(base, draftRaw, latest); err == nil {
		t.Fatal("MergeWithLatest returned nil conflict error")
	}
}

func TestMergeWithLatestPreservesDescriptionOnlyGroup(t *testing.T) {
	base := configWithDescriptionOnlyGroupRaw("1", true, "old")
	draftRaw := configWithDescriptionOnlyGroupRaw("1", true, "local")
	latest := configWithDescriptionOnlyGroupRaw("2", true, "old")

	mergedRaw, changed, err := MergeWithLatest(base, draftRaw, latest)
	if err != nil {
		t.Fatalf("MergeWithLatest returned error: %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	merged, err := firebase.ParseRemoteConfig(mergedRaw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	group, ok := merged.ParameterGroups["ROKU"]
	if !ok || group.Description != "FLAGS FOR ROKU" || len(group.Parameters) != 0 {
		t.Fatalf("ROKU group = %#v, ok = %v; want preserved description-only group", group, ok)
	}
}

func TestMergeWithLatestAppliesExplicitEmptyGroupDeletion(t *testing.T) {
	base := configWithDescriptionOnlyGroupRaw("1", true, "old")
	draftRaw := configWithDescriptionOnlyGroupRaw("1", false, "old")
	latest := configWithDescriptionOnlyGroupRaw("2", true, "old")

	mergedRaw, changed, err := MergeWithLatest(base, draftRaw, latest)
	if err != nil {
		t.Fatalf("MergeWithLatest returned error: %v", err)
	}
	if !changed {
		t.Fatal("changed = false, want true")
	}
	merged, err := firebase.ParseRemoteConfig(mergedRaw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	if _, ok := merged.ParameterGroups["ROKU"]; ok {
		t.Fatalf("ROKU group still present after explicit deletion: %#v", merged.ParameterGroups["ROKU"])
	}
}

func configWithDescriptionOnlyGroupRaw(version string, includeGroup bool, value string) json.RawMessage {
	cfg := firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": remoteConfigParam(value, "STRING"),
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: version},
	}
	if includeGroup {
		cfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{
			"ROKU": {Description: "FLAGS FOR ROKU"},
		}
	}
	return marshalRemoteConfig(cfg)
}

func completeConfigRaw(version, groupDescription string, conditions []firebase.RemoteConfigCondition, remoteValue string) json.RawMessage {
	value := firebase.RemoteConfigValue{Value: remoteValue}
	cfg := firebase.RemoteConfig{
		Version:    firebase.RemoteConfigVersion{VersionNumber: version},
		Conditions: conditions,
		Parameters: map[string]firebase.RemoteConfigParam{"remote_flag": {DefaultValue: &value, ValueType: "STRING"}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group": {Description: groupDescription, Parameters: map[string]firebase.RemoteConfigParam{"group_flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "stable"}, ValueType: "STRING"}}},
		},
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return raw
}
