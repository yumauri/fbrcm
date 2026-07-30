package updatecmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRunUpdateStdinUpdatesRootParameterValue(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetIn(strings.NewReader(`{"parameters":{"flag":{"defaultValue":{"value":"old"},"valueType":"STRING"}}}`))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	spec := updateSpec{
		value: &valueSpec{value: "new", valueType: "STRING"},
	}
	err := runUpdateStdin(cmd, []string{"=flag"}, "", shared.ParameterSearch{}, spec)
	if err != nil {
		t.Fatalf("runUpdateStdin returned error: %v", err)
	}

	var cfg firebase.RemoteConfig
	if err := json.Unmarshal(out.Bytes(), &cfg); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, out.String())
	}
	param := cfg.Parameters["flag"]
	if param.DefaultValue == nil || param.DefaultValue.Value != "new" {
		t.Fatalf("flag default = %#v, want new", param.DefaultValue)
	}
	if param.ValueType != "STRING" {
		t.Fatalf("flag type = %q, want STRING", param.ValueType)
	}
}

func TestRunUpdateStdinRejectsManagedValue(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetIn(strings.NewReader(`{"parameters":{"flag":{"defaultValue":{"rolloutValue":{"rolloutId":"rollout-1","value":"20","percent":10}},"valueType":"NUMBER"}}}`))
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})

	err := runUpdateStdin(cmd, []string{"=flag"}, "", shared.ParameterSearch{}, updateSpec{
		value: &valueSpec{value: "30", valueType: "NUMBER"},
	})
	if err == nil || !strings.Contains(err.Error(), "rollout value is read-only") {
		t.Fatalf("error = %v, want read-only rollout error", err)
	}
	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", out.String())
	}
}

func TestUpdateParamSlotRenamesMovesAndEditsParameter(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {
				Description: "group metadata",
				Parameters: map[string]firebase.RemoteConfigParam{
					"old_flag": {
						DefaultValue: &firebase.RemoteConfigValue{Value: "old"},
						Description:  "old description",
						ValueType:    "STRING",
						ConditionalValues: map[string]firebase.RemoteConfigValue{
							"beta": {Value: "beta"},
							"ga":   {Value: "ga"},
						},
					},
				},
			},
		},
	}
	target := shared.ParamTarget{
		Key:   "old_flag",
		Group: "group-a",
		Param: cfg.ParameterGroups["group-a"].Parameters["old_flag"],
	}
	spec := updateSpec{
		value:                   &valueSpec{value: "new", valueType: "JSON"},
		name:                    "new_flag",
		group:                   "",
		description:             "new description",
		removeConditionalValues: []string{"beta"},
		nameChanged:             true,
		groupChanged:            true,
		descriptionChanged:      true,
	}

	if err := updateParamSlot(cfg, target, spec); err != nil {
		t.Fatalf("updateParamSlot returned error: %v", err)
	}
	group, ok := cfg.ParameterGroups["group-a"]
	if !ok {
		t.Fatal("group-a was removed after moving its last parameter to root")
	}
	if group.Description != "group metadata" || group.Parameters != nil {
		t.Fatalf("group-a = %#v, want preserved metadata and nil parameters", group)
	}
	param, ok := cfg.Parameters["new_flag"]
	if !ok {
		t.Fatalf("new_flag missing from root parameters")
	}
	if param.DefaultValue == nil || param.DefaultValue.Value != "new" {
		t.Fatalf("new_flag default = %#v, want new", param.DefaultValue)
	}
	if param.ValueType != "JSON" || param.Description != "new description" {
		t.Fatalf("new_flag metadata = %q/%q, want JSON/new description", param.ValueType, param.Description)
	}
	if _, ok := param.ConditionalValues["beta"]; ok {
		t.Fatalf("beta conditional value still present")
	}
	if got := param.ConditionalValues["ga"].Value; got != "ga" {
		t.Fatalf("ga conditional value = %q, want ga", got)
	}
}

func TestUpdateParamSlotRejectsDestinationCollision(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"source": {DefaultValue: &firebase.RemoteConfigValue{Value: "source"}},
			"target": {DefaultValue: &firebase.RemoteConfigValue{Value: "target"}},
		},
	}
	target := shared.ParamTarget{Key: "source", Param: cfg.Parameters["source"]}
	spec := updateSpec{name: "target", nameChanged: true}

	if err := updateParamSlot(cfg, target, spec); err == nil {
		t.Fatalf("updateParamSlot accepted destination collision")
	}
	if _, ok := cfg.Parameters["source"]; !ok {
		t.Fatalf("source was removed after rejected collision")
	}
	if got := cfg.Parameters["target"].DefaultValue.Value; got != "target" {
		t.Fatalf("target default = %q, want target", got)
	}
}

func TestUpdateParamSlotRemovesAllConditionalValues(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "default"},
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"beta": {Value: "beta"},
				},
			},
		},
	}
	target := shared.ParamTarget{Key: "flag", Param: cfg.Parameters["flag"]}

	if err := updateParamSlot(cfg, target, updateSpec{removeAllConditionalValues: true}); err != nil {
		t.Fatalf("updateParamSlot returned error: %v", err)
	}
	if cfg.Parameters["flag"].ConditionalValues != nil {
		t.Fatalf("conditional values = %#v, want nil", cfg.Parameters["flag"].ConditionalValues)
	}
}

func TestUpdateParamSlotAssignsConditionalValueWithoutChangingDefault(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "Beta Users", Expression: "true"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{Value: "default"},
				ValueType:    "STRING",
			},
		},
	}
	target := shared.ParamTarget{Key: "flag", Param: cfg.Parameters["flag"]}
	spec := updateSpec{
		value:     &valueSpec{value: "enabled", valueType: "STRING"},
		condition: "beta users",
	}

	if err := updateParamSlot(cfg, target, spec); err != nil {
		t.Fatal(err)
	}
	param := cfg.Parameters["flag"]
	if param.DefaultValue == nil || param.DefaultValue.Value != "default" {
		t.Fatalf("default value = %#v, want unchanged", param.DefaultValue)
	}
	if got := param.ConditionalValues["Beta Users"].Value; got != "enabled" {
		t.Fatalf("conditional value = %q, want enabled", got)
	}
}

func TestReadUpdateSpecConditionRequiresValue(t *testing.T) {
	cmd := New(nil)
	if err := cmd.Flags().Set("condition", "beta"); err != nil {
		t.Fatal(err)
	}
	if _, err := readUpdateSpec(cmd); err == nil || !strings.Contains(err.Error(), "requires one value flag") {
		t.Fatalf("readUpdateSpec error = %v", err)
	}
}

func TestUpdateParamSlotSetsInAppDefaultWithoutChangingType(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "Beta", Expression: "true"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"payload": {
				DefaultValue: &firebase.RemoteConfigValue{Value: `{"default":true}`},
				ValueType:    "JSON",
			},
		},
	}
	target := shared.ParamTarget{Key: "payload", Param: cfg.Parameters["payload"]}
	spec := updateSpec{
		value:     &valueSpec{useInAppDefault: true},
		condition: "beta",
	}

	if err := updateParamSlot(cfg, target, spec); err != nil {
		t.Fatal(err)
	}
	param := cfg.Parameters["payload"]
	if param.ValueType != "JSON" {
		t.Fatalf("type = %q, want preserved JSON", param.ValueType)
	}
	if value := param.ConditionalValues["Beta"]; !value.UseInAppDefault || value.Value != "" {
		t.Fatalf("conditional value = %#v, want useInAppDefault", value)
	}
	if param.DefaultValue == nil || param.DefaultValue.Value != `{"default":true}` {
		t.Fatalf("default value = %#v, want unchanged", param.DefaultValue)
	}
}

func TestUpdateParamSlotConcreteValueClearsInAppDefault(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				DefaultValue: &firebase.RemoteConfigValue{UseInAppDefault: true},
				ValueType:    "BOOLEAN",
			},
		},
	}
	target := shared.ParamTarget{Key: "flag", Param: cfg.Parameters["flag"]}

	if err := updateParamSlot(cfg, target, updateSpec{value: &valueSpec{value: "true", valueType: "BOOLEAN"}}); err != nil {
		t.Fatal(err)
	}
	value := cfg.Parameters["flag"].DefaultValue
	if value == nil || value.UseInAppDefault || value.Value != "true" {
		t.Fatalf("default value = %#v, want plain true", value)
	}
}
