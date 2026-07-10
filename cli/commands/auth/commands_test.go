package auth

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func setupAuthCommandTest(t *testing.T) *core.Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestAuthListEmptyJSON(t *testing.T) {
	svc := setupAuthCommandTest(t)
	listCmd, _, err := New(svc).Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	if err := listCmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	listCmd.SetOut(&out)
	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatalf("auth list = %v", err)
	}
	if !strings.Contains(out.String(), `"auth"`) {
		t.Fatalf("output = %q", out.String())
	}
}

func TestAuthAddGCloudAndList(t *testing.T) {
	svc := setupAuthCommandTest(t)
	addCmd, _, err := New(svc).Find([]string{"add", "gcloud", "main"})
	if err != nil {
		t.Fatal(err)
	}
	if err := addCmd.RunE(addCmd, []string{"main"}); err != nil {
		t.Fatalf("auth add gcloud = %v", err)
	}

	listCmd, _, err := New(svc).Find([]string{"list"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	listCmd.SetOut(&out)
	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatalf("auth list = %v", err)
	}
	if !strings.Contains(out.String(), "main") {
		t.Fatalf("output = %q, want main auth", out.String())
	}
}

func TestAuthPathCommand(t *testing.T) {
	svc := setupAuthCommandTest(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatal(err)
	}

	pathCmd, _, err := New(svc).Find([]string{"path", "main"})
	if err != nil {
		t.Fatal(err)
	}
	if err := pathCmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	pathCmd.SetOut(&out)
	if err := pathCmd.RunE(pathCmd, []string{"main"}); err != nil {
		t.Fatalf("auth path = %v", err)
	}
	if !strings.Contains(out.String(), "auth-config.json") {
		t.Fatalf("output = %q", out.String())
	}
}
