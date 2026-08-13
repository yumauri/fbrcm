package projects

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func TestResetJSONReportsWhetherRegistryWasRemoved(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := coreconfig.SwitchProfile(coreconfig.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	execute := func() projectsResetResult {
		t.Helper()
		parent := &cobra.Command{Use: "root"}
		parent.PersistentFlags().Bool("json", false, "machine output")
		parent.AddCommand(newResetCommand(svc))
		var output bytes.Buffer
		parent.SetOut(&output)
		parent.SetErr(&output)
		parent.SetArgs([]string{"reset", "--yes", "--json"})
		if err := parent.Execute(); err != nil {
			t.Fatal(err)
		}
		var result projectsResetResult
		if err := json.Unmarshal(output.Bytes(), &result); err != nil {
			t.Fatalf("decode %q: %v", output.String(), err)
		}
		return result
	}

	if result := execute(); result.Changed || result.Status != "reset" {
		t.Fatalf("absent registry result = %#v", result)
	}
	if err := coreconfig.SaveProjects([]coreconfig.Project{{ProjectID: "demo"}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if result := execute(); !result.Changed || result.Status != "reset" {
		t.Fatalf("existing registry result = %#v", result)
	}
}
