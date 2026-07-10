package deletecmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRunDeleteStdinDeletesRootParameter(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetIn(strings.NewReader(`{"parameters":{"keep":{"defaultValue":{"value":"yes"}},"remove_me":{"defaultValue":{"value":"no"}}}}`))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err := runDeleteStdin(cmd, []string{"=remove_me"}, "", shared.ParameterSearch{})
	if err != nil {
		t.Fatalf("runDeleteStdin returned error: %v", err)
	}

	var cfg firebase.RemoteConfig
	if err := json.Unmarshal(out.Bytes(), &cfg); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, out.String())
	}
	if _, ok := cfg.Parameters["remove_me"]; ok {
		t.Fatalf("remove_me still present in output: %s", out.String())
	}
	if _, ok := cfg.Parameters["keep"]; !ok {
		t.Fatalf("keep missing from output: %s", out.String())
	}
}

func TestConfirmAndDeleteProjectRemovesMatchedTargets(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cfg := &firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			"keep":      {DefaultValue: &firebase.RemoteConfigValue{Value: "yes"}},
			"remove_me": {DefaultValue: &firebase.RemoteConfigValue{Value: "no"}},
		},
	}
	matched := []shared.ParamTarget{{Key: "remove_me", Param: cfg.Parameters["remove_me"]}}

	deleted, finalCfg, err := confirmAndDeleteProject(cmd, "project-a", cfg, matched, true, &errOut)
	if err != nil {
		t.Fatalf("confirmAndDeleteProject returned error: %v", err)
	}
	if len(deleted) != 1 || deleted[0].Key != "remove_me" {
		t.Fatalf("deleted = %#v, want remove_me", deleted)
	}
	if _, ok := finalCfg.Parameters["remove_me"]; ok {
		t.Fatalf("remove_me still present in final config")
	}
	if _, ok := cfg.Parameters["remove_me"]; !ok {
		t.Fatalf("confirmAndDeleteProject mutated original config")
	}
	if !strings.Contains(errOut.String(), "remove_me") {
		t.Fatalf("diff output = %q, want remove_me", errOut.String())
	}
}

func TestRenderDeletedParameterIncludesDetails(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	target := shared.ParamTarget{
		Key:   "flag",
		Group: "group-a",
		Param: firebase.RemoteConfigParam{
			DefaultValue: &firebase.RemoteConfigValue{Value: "default"},
			Description:  "description",
			ValueType:    "STRING",
			ConditionalValues: map[string]firebase.RemoteConfigValue{
				"beta": {Value: "beta"},
			},
		},
	}

	got := rc.RenderRemovedParameterDetail(target.Key, target.Group, target.Param)
	for _, want := range []string{"flag [group-a]", "type:", "STRING", "description:", "default:", "cond beta:"} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderRemovedParameterDetail = %q, want substring %q", got, want)
		}
	}
}
