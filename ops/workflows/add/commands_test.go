package addcmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
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

func TestReadAddValueSpecRequiresTypeForInAppDefault(t *testing.T) {
	cmd := New(nil)
	if err := cmd.Flags().Set("use-in-app-default", "true"); err != nil {
		t.Fatal(err)
	}
	if _, err := readAddValueSpec(cmd); err == nil || !strings.Contains(err.Error(), "--type is required") {
		t.Fatalf("readAddValueSpec error = %v, want required type", err)
	}
	if err := cmd.Flags().Set("type", "json"); err != nil {
		t.Fatal(err)
	}
	spec, err := readAddValueSpec(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !spec.useInAppDefault || spec.valueType != "JSON" {
		t.Fatalf("readAddValueSpec = %+v, want JSON in-app default", spec)
	}
}

func TestAddParameterSupportsInAppDefault(t *testing.T) {
	cfg := &firebase.RemoteConfig{}
	changed, finalCfg, err := addParameter(cfg, "payload", "", "", addValueSpec{
		valueType: "JSON", useInAppDefault: true,
	})
	if err != nil || !changed {
		t.Fatalf("addParameter = changed:%v err:%v", changed, err)
	}
	param := finalCfg.Parameters["payload"]
	if param.ValueType != "JSON" || param.DefaultValue == nil || !param.DefaultValue.UseInAppDefault || param.DefaultValue.Value != "" {
		t.Fatalf("payload = %#v, want JSON useInAppDefault", param)
	}
}

func TestReadAddOptionsIncludesYes(t *testing.T) {
	cmd := New(nil)
	for flag, value := range map[string]string{
		"type":  "boolean",
		"value": "true",
		"yes":   "true",
	} {
		if err := cmd.Flags().Set(flag, value); err != nil {
			t.Fatal(err)
		}
	}
	opts, err := readAddOptions(cmd, []string{" feature_enabled "})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.yes || opts.key != "feature_enabled" {
		t.Fatalf("options = %#v, want yes and trimmed key", opts)
	}
}

func TestAddProjectPrintsDiffWithoutMutatingSource(t *testing.T) {
	cmd := &cobra.Command{}
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	cmd.SetOut(&bytes.Buffer{})
	cfg := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"existing": {DefaultValue: &firebase.RemoteConfigValue{Value: "old"}},
	}}

	changed, finalCfg, err := addProject(
		cmd,
		core.Project{ProjectID: "demo"},
		cfg,
		"new_flag",
		"group-a",
		"New flag",
		addValueSpec{value: "true", valueType: "BOOLEAN"},
		true,
	)
	if err != nil || !changed {
		t.Fatalf("addProject = changed %v, err %v", changed, err)
	}
	if _, exists := finalCfg.ParameterGroups["group-a"].Parameters["new_flag"]; !exists {
		t.Fatal("new_flag is missing from final config")
	}
	if _, exists := cfg.ParameterGroups["group-a"]; exists {
		t.Fatal("addProject mutated source config")
	}
	diff := ansi.Strip(errOut.String())
	for _, want := range []string{"+ new_flag [group-a]", "default:", "true"} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff = %q, want %q", diff, want)
		}
	}
}
