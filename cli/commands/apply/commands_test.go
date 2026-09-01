package applycmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rc/publication"
)

func TestApplyNoChangePlanSucceedsWithoutFirebase(t *testing.T) {
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

	cmd := New(nil)
	cmd.SetContext(core.WithExecutionPolicy(t.Context(), core.StatelessExecutionPolicy()))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{path, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"status": "unchanged"`)) {
		t.Fatalf("output = %s", out.String())
	}
}
