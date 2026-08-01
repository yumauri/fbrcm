package hooks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func setupCommandTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.NoLocalConfig, "")
	t.Setenv("FBRCM_HOOK_TRUST", "")
	config.SetLocalConfigDisabled(false)
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	if err := os.WriteFile(filepath.Join(repo, config.LocalConfigFileName), []byte(`[hooks]
timeout = "30s"
pre_publish = ["./validate"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	return repo
}

func executeCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := New()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(bytes.NewBufferString("y\n"))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestStatusAndTrustJSON(t *testing.T) {
	repo := setupCommandTest(t)
	stdout, _, err := executeCommand(t, "status", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var before statusResult
	if err := json.Unmarshal([]byte(stdout), &before); err != nil {
		t.Fatal(err)
	}
	if before.Trusted || !before.LocalHooks || before.LocalConfig != filepath.Join(repo, config.LocalConfigFileName) || before.Fingerprint == "" {
		t.Fatalf("before = %+v", before)
	}

	stdout, _, err = executeCommand(t, "trust", "--yes", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var after statusResult
	if err := json.Unmarshal([]byte(stdout), &after); err != nil {
		t.Fatal(err)
	}
	if !after.Trusted || after.Fingerprint != before.Fingerprint {
		t.Fatalf("after = %+v", after)
	}

	fingerprint, _, err := executeCommand(t, "fingerprint")
	if err != nil || fingerprint != before.Fingerprint+"\n" {
		t.Fatalf("fingerprint = %q, %v", fingerprint, err)
	}
}

func TestUntrustRemovesCurrentRecord(t *testing.T) {
	setupCommandTest(t)
	if _, _, err := executeCommand(t, "trust", "--yes"); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := executeCommand(t, "untrust", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatal(err)
	}
	if result["changed"] != true || result["trusted"] != false {
		t.Fatalf("result = %#v", result)
	}
}
