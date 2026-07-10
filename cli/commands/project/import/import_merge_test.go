package importpkg

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestMergeRemoteConfigsKeepsCurrentConflicts(t *testing.T) {
	finalCfg, err := mergeRemoteConfigs(&cobra.Command{}, mergeCurrentFixture(), mergeImportFixture(), importOptions{
		mergeResolve: string(conflictResolutionCurrent),
	})
	if err != nil {
		t.Fatalf("mergeRemoteConfigs returned error: %v", err)
	}

	if got := finalCfg.Conditions[0].Expression; got != `app.version == "1.0"` {
		t.Fatalf("condition expression = %q, want current", got)
	}
	if got := finalCfg.ParameterGroups["group-a"].Description; got != "current group" {
		t.Fatalf("group description = %q, want current", got)
	}
	if got := finalCfg.ParameterGroups["group-a"].Parameters["flag"].DefaultValue.Value; got != "current" {
		t.Fatalf("flag value = %q, want current", got)
	}
	if got := finalCfg.Parameters["root_new"].DefaultValue.Value; got != "new" {
		t.Fatalf("root_new value = %q, want new import param", got)
	}
}

func TestMergeRemoteConfigsUsesImportConflicts(t *testing.T) {
	finalCfg, err := mergeRemoteConfigs(&cobra.Command{}, mergeCurrentFixture(), mergeImportFixture(), importOptions{
		mergeResolve: string(conflictResolutionImport),
	})
	if err != nil {
		t.Fatalf("mergeRemoteConfigs returned error: %v", err)
	}

	if got := finalCfg.Conditions[0].Expression; got != `app.version == "2.0"` {
		t.Fatalf("condition expression = %q, want import", got)
	}
	if got := finalCfg.ParameterGroups["group-a"].Description; got != "import group" {
		t.Fatalf("group description = %q, want import", got)
	}
	if got := finalCfg.ParameterGroups["group-a"].Parameters["flag"].DefaultValue.Value; got != "import" {
		t.Fatalf("flag value = %q, want import", got)
	}
	if got := finalCfg.Parameters["root_new"].DefaultValue.Value; got != "new" {
		t.Fatalf("root_new value = %q, want new import param", got)
	}
}

func mergeCurrentFixture() *firebase.RemoteConfig {
	return &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "condition-a", Expression: `app.version == "1.0"`},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {
				Description: "current group",
				Parameters: map[string]firebase.RemoteConfigParam{
					"flag": {
						DefaultValue: &firebase.RemoteConfigValue{Value: "current"},
						ValueType:    "STRING",
					},
				},
			},
		},
	}
}

func mergeImportFixture() *firebase.RemoteConfig {
	return &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "condition-a", Expression: `app.version == "2.0"`},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"root_new": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "new"},
				ValueType:    "STRING",
			},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {
				Description: "import group",
				Parameters: map[string]firebase.RemoteConfigParam{
					"flag": {
						DefaultValue: &firebase.RemoteConfigValue{Value: "import"},
						ValueType:    "STRING",
					},
				},
			},
		},
	}
}
