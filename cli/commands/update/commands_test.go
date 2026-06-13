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

func TestUpdateParamSlotRenamesMovesAndEditsParameter(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {
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
	if _, ok := cfg.ParameterGroups["group-a"]; ok {
		t.Fatalf("group-a still present after moving last parameter to root")
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
