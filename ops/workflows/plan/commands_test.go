package plancmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/core/rc/publication"
)

func writeTestPlan(t *testing.T) string {
	t.Helper()
	rawConfig := json.RawMessage(`{"parameters":{},"version":{"versionNumber":"1"}}`)
	plan := publication.New("test", "update", "stateless", nil)
	plan.Targets = append(plan.Targets, publication.Target{
		Target: "demo", ProjectID: "demo", Template: "client", Action: publication.ActionNone,
		Base: publication.Snapshot{Version: "1", RemoteConfig: rawConfig}, Candidate: publication.Snapshot{RemoteConfig: rawConfig},
		Validation: publication.Validation{Source: "local", ValidatedAt: plan.CreatedAt}, Source: publication.Source{Kind: "direct"},
	})
	if err := publication.Seal(plan); err != nil {
		t.Fatal(err)
	}
	raw, err := publication.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "test.fbrcm-plan.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPlanShowAndValidateJSON(t *testing.T) {
	path := writeTestPlan(t)
	for _, args := range [][]string{{"show", path, "--json"}, {"validate", path, "--json"}} {
		cmd := New()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !json.Valid(out.Bytes()) {
			t.Fatalf("%v output is not JSON: %s", args, out.String())
		}
	}
}
