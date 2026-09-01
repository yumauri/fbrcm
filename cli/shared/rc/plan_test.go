package rc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/core/rc/publication"
)

func TestCreatePlanFileIsPrivateAndExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "publication-plan.json")
	original := []byte(`{"plan_id":"plan_original"}`)
	if err := createPlanFile(path, original); err != nil {
		t.Fatalf("createPlanFile = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("plan permissions = %o, want 600", got)
	}

	err = createPlanFile(path, []byte(`{"plan_id":"plan_replacement"}`))
	var conflict *machine.ConflictError
	if !errors.As(err, &conflict) || conflict.Code != "plan.exists" {
		t.Fatalf("second createPlanFile error = %#v, want plan.exists conflict", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("existing plan was replaced: %s", got)
	}
}

func TestWritePublicationPlanReportsExactPrivateArtifact(t *testing.T) {
	rawConfig := json.RawMessage(`{"parameters":{}}`)
	plan := publication.New("test", "update", "stateless", nil)
	plan.Targets = append(plan.Targets, publication.Target{
		Target: "demo", ProjectID: "demo", Template: "client", Action: publication.ActionNone,
		Base: publication.Snapshot{Version: "1", RemoteConfig: rawConfig}, Candidate: publication.Snapshot{RemoteConfig: rawConfig},
		Validation: publication.Validation{Source: "local", ValidatedAt: plan.CreatedAt}, Source: publication.Source{Kind: "direct"},
	})
	path := filepath.Join(t.TempDir(), "plan.json")
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	result, err := WritePublicationPlan(cmd, plan, path)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if result.Path != path || result.Artifact.Destination != path || result.Artifact.MediaType != "application/vnd.fbrcm.publication-plan+json" || result.Artifact.Encoding != "none" || result.Artifact.Overwritten || result.Artifact.SizeBytes != int64(len(raw)) || result.Artifact.SHA256 != hex.EncodeToString(digest[:]) {
		t.Fatalf("artifact metadata = %#v for %d bytes", result.Artifact, len(raw))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("plan permissions = %o, want 600", info.Mode().Perm())
	}
}
