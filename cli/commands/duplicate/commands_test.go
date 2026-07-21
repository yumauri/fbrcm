package duplicatecmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestDuplicateParameterClonesCompleteGroupedParameter(t *testing.T) {
	original := &firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{
		"group-a": {
			Description: "Group description",
			Parameters: map[string]firebase.RemoteConfigParam{
				"Source_Flag": {
					DefaultValue: &firebase.RemoteConfigValue{Value: "default"},
					ConditionalValues: map[string]firebase.RemoteConfigValue{
						"beta": {Value: "enabled"},
					},
					Description: "Parameter description",
					ValueType:   "STRING",
				},
			},
		},
	}}

	changed, finalCfg, source, err := duplicateParameter(original, "source_flag", "target_flag")
	if err != nil || !changed {
		t.Fatalf("duplicateParameter = changed %v, err %v", changed, err)
	}
	if source.Key != "Source_Flag" || source.Group != "group-a" {
		t.Fatalf("resolved source = %#v", source)
	}
	if _, exists := original.ParameterGroups["group-a"].Parameters["target_flag"]; exists {
		t.Fatal("duplicateParameter mutated the original config")
	}
	duplicate := finalCfg.ParameterGroups["group-a"].Parameters["target_flag"]
	if duplicate.Description != "Parameter description" || duplicate.ValueType != "STRING" || duplicate.DefaultValue == nil || duplicate.DefaultValue.Value != "default" {
		t.Fatalf("duplicate metadata = %#v", duplicate)
	}
	if duplicate.ConditionalValues["beta"].Value != "enabled" {
		t.Fatalf("duplicate conditional values = %#v", duplicate.ConditionalValues)
	}
	if finalCfg.ParameterGroups["group-a"].Description != "Group description" {
		t.Fatalf("group description = %q", finalCfg.ParameterGroups["group-a"].Description)
	}
}

func TestDuplicateParameterSkipsMissingSourceAndRejectsTargetCollision(t *testing.T) {
	cfg := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"source": {},
		"target": {},
	}}
	changed, _, _, err := duplicateParameter(cfg, "missing", "copy")
	if err != nil || changed {
		t.Fatalf("missing source = changed %v, err %v", changed, err)
	}
	changed, _, _, err = duplicateParameter(cfg, "source", "TARGET")
	if err == nil || changed || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("target collision = changed %v, err %v", changed, err)
	}
}

func TestReadDuplicateOptionsIncludesProjectFiltersAndExpression(t *testing.T) {
	cmd := New(nil)
	if err := cmd.Flags().Set("project", "^prod"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("expr", `project.id contains "android"`); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("dry-run", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("draft", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}
	opts, err := readDuplicateOptions(cmd, []string{" source ", " target "})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.projectFilters) != 1 || opts.projectFilters[0] != "^prod" || opts.projectExpr != `project.id contains "android"` {
		t.Fatalf("project selection = %v, %q", opts.projectFilters, opts.projectExpr)
	}
	if !opts.dryRun || !opts.draft || !opts.yes || opts.source != "source" || opts.target != "target" {
		t.Fatalf("options = %#v", opts)
	}
}

func TestResolveSourceRejectsDuplicateKeysAcrossGroups(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{"flag": {}},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group-a": {Parameters: map[string]firebase.RemoteConfigParam{"FLAG": {}}},
		},
	}
	_, found, err := resolveSource(cfg, "flag")
	if err == nil || found || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("resolveSource = found %v, err %v", found, err)
	}
}

func TestDuplicateProjectPrintsDiffWithoutMutatingSource(t *testing.T) {
	cmd := &cobra.Command{}
	var errOut bytes.Buffer
	cmd.SetErr(&errOut)
	cmd.SetOut(&bytes.Buffer{})
	cfg := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"source": {DefaultValue: &firebase.RemoteConfigValue{Value: "on"}},
	}}
	changed, finalCfg, err := duplicateProject(cmd, core.Project{ProjectID: "demo"}, cfg, "source", "copy", true)
	if err != nil || !changed {
		t.Fatalf("duplicateProject = changed %v, err %v", changed, err)
	}
	if _, exists := finalCfg.Parameters["copy"]; !exists {
		t.Fatal("copy is missing from final config")
	}
	if _, exists := cfg.Parameters["copy"]; exists {
		t.Fatal("duplicateProject mutated source config")
	}
	if diff := ansi.Strip(errOut.String()); !strings.Contains(diff, "+ copy") || !strings.Contains(diff, "default:") || !strings.Contains(diff, "on") {
		t.Fatalf("diff = %q", diff)
	}
}
