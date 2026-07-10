package profile

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func setupProfileTest(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
}

func TestProfileRootPrintsActiveProfile(t *testing.T) {
	setupProfileTest(t)
	if err := config.SwitchProfile("work"); err != nil {
		t.Fatal(err)
	}

	cmd := New()
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("profile root = %v", err)
	}
	if strings.TrimSpace(out.String()) != "work" {
		t.Fatalf("output = %q, want work", out.String())
	}
}

func TestProfileListJSON(t *testing.T) {
	setupProfileTest(t)
	if err := config.SwitchProfile("alpha"); err != nil {
		t.Fatal(err)
	}
	if err := config.SwitchProfile("beta"); err != nil {
		t.Fatal(err)
	}
	if err := config.SwitchProfile("alpha"); err != nil {
		t.Fatal(err)
	}

	listCmd, _, err := New().Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	if err := listCmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	listCmd.SetOut(&out)
	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatalf("profile list = %v", err)
	}
	if !strings.Contains(out.String(), `"profile": "alpha"`) || !strings.Contains(out.String(), `"active": true`) {
		t.Fatalf("json output = %s", out.String())
	}
}

func TestProfilePathCommand(t *testing.T) {
	setupProfileTest(t)
	if err := config.SwitchProfile("paths"); err != nil {
		t.Fatal(err)
	}

	pathCmd, _, err := New().Find([]string{"path", "paths"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	pathCmd.SetOut(&out)
	if err := pathCmd.RunE(pathCmd, []string{"paths"}); err != nil {
		t.Fatalf("profile path = %v", err)
	}
	text := out.String()
	if !strings.Contains(text, filepath.Join("config", "paths")) {
		t.Fatalf("output = %q, want config path", text)
	}
}
