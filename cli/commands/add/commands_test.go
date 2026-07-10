package addcmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRunAddStdinAddsGroupedParameter(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetIn(strings.NewReader(`{"parameters":{"existing":{"defaultValue":{"value":"old"}}}}`))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err := runAddStdin(cmd, "new_flag", "group-a", "New flag", addValueSpec{value: "on", valueType: "STRING"}, "")
	if err != nil {
		t.Fatalf("runAddStdin returned error: %v", err)
	}

	var cfg firebase.RemoteConfig
	if err := json.Unmarshal(out.Bytes(), &cfg); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, out.String())
	}
	group, ok := cfg.ParameterGroups["group-a"]
	if !ok {
		t.Fatalf("group-a not found in output: %s", out.String())
	}
	param, ok := group.Parameters["new_flag"]
	if !ok {
		t.Fatalf("new_flag not found in group-a: %s", out.String())
	}
	if param.DefaultValue == nil || param.DefaultValue.Value != "on" {
		t.Fatalf("new_flag default = %#v, want on", param.DefaultValue)
	}
	if param.Description != "New flag" {
		t.Fatalf("new_flag description = %q, want New flag", param.Description)
	}
	if param.ValueType != "STRING" {
		t.Fatalf("new_flag type = %q, want STRING", param.ValueType)
	}
}

func TestAddParameterClonesAndRejectsDuplicates(t *testing.T) {
	original := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"existing": {DefaultValue: &firebase.RemoteConfigValue{Value: "old"}},
		},
	}

	changed, finalCfg, err := addParameter(original, "new_flag", "", "New flag", addValueSpec{value: "on", valueType: "BOOLEAN"})
	if err != nil {
		t.Fatalf("addParameter returned error: %v", err)
	}
	if !changed {
		t.Fatalf("addParameter changed = false, want true")
	}
	if _, ok := original.Parameters["new_flag"]; ok {
		t.Fatalf("addParameter mutated original config")
	}
	param := finalCfg.Parameters["new_flag"]
	if param.DefaultValue == nil || param.DefaultValue.Value != "on" {
		t.Fatalf("new_flag default = %#v, want on", param.DefaultValue)
	}
	if param.Description != "New flag" || param.ValueType != "BOOLEAN" {
		t.Fatalf("new_flag metadata = %q/%q, want New flag/BOOLEAN", param.Description, param.ValueType)
	}

	changed, finalCfg, err = addParameter(original, "existing", "group-a", "Duplicate", addValueSpec{value: "new", valueType: "STRING"})
	if err != nil {
		t.Fatalf("addParameter returned error: %v", err)
	}
	if changed {
		t.Fatalf("duplicate addParameter changed = true, want false")
	}
	if _, ok := finalCfg.ParameterGroups["group-a"]; ok {
		t.Fatalf("duplicate addParameter created group-a")
	}
}
