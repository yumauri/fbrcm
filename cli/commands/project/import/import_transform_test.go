package importpkg

import (
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestTransformImportConfigRemovesProjectSpecificConditions(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "keep", Expression: `app.version == "1.0"`},
			{Name: "project_only", Expression: `app.id == "com.example.app"`},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"root_keep": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "default"},
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"keep":         {Value: "yes"},
					"project_only": {Value: "no"},
				},
			},
			"root_drop": {
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"project_only": {Value: "drop"},
				},
			},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"group_keep": {
						ConditionalValues: map[string]firebase.RemoteConfigValue{
							"keep": {Value: "group"},
						},
					},
				},
			},
			"group-empty": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"group_drop": {
						ConditionalValues: map[string]firebase.RemoteConfigValue{
							"project_only": {Value: "drop"},
						},
					},
				},
			},
		},
	}

	err := transformImportConfig(core.Project{ProjectID: "test-project"}, cfg, importOptions{
		removeProjectSpecificConditions: true,
	})
	if err != nil {
		t.Fatalf("transformImportConfig returned error: %v", err)
	}

	if len(cfg.Conditions) != 1 || cfg.Conditions[0].Name != "keep" {
		t.Fatalf("conditions = %#v, want only keep", cfg.Conditions)
	}
	if _, ok := cfg.Parameters["root_drop"]; ok {
		t.Fatalf("root_drop still present: %#v", cfg.Parameters["root_drop"])
	}
	rootKeep := cfg.Parameters["root_keep"]
	if _, ok := rootKeep.ConditionalValues["project_only"]; ok {
		t.Fatalf("root_keep still references project_only: %#v", rootKeep.ConditionalValues)
	}
	if _, ok := rootKeep.ConditionalValues["keep"]; !ok {
		t.Fatalf("root_keep lost keep conditional: %#v", rootKeep.ConditionalValues)
	}
	if _, ok := cfg.ParameterGroups["group-empty"]; ok {
		t.Fatalf("group-empty still present: %#v", cfg.ParameterGroups["group-empty"])
	}
	if _, ok := cfg.ParameterGroups["group-a"]; !ok {
		t.Fatalf("group-a missing: %#v", cfg.ParameterGroups)
	}
}
