package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if got := strings.TrimSpace(out.String()); got != "[]" {
		t.Fatalf("empty auth JSON = %q, want []", got)
	}
}

func TestAuthListJSONMarksDefaultIdentity(t *testing.T) {
	entries := []config.AuthEntry{
		{ID: "main", Type: config.AuthTypeGCloud, Label: "Main"},
		{ID: "work", Type: config.AuthTypeOAuth, Label: "Work"},
	}
	raw, err := json.Marshal(newAuthListItems(entries, "work"))
	if err != nil {
		t.Fatal(err)
	}
	var got []authListItem
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode auth list JSON: %v", err)
	}
	if len(got) != 2 || got[0].Default || !got[1].Default {
		t.Fatalf("auth list JSON = %#v", got)
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

func TestAuthBindDefaultsToAllAndSkipsInaccessibleProjects(t *testing.T) {
	svc := setupAuthCommandTest(t)
	for _, authID := range []string{"old", "main"} {
		if _, err := svc.AddGCloudAuth(authID, authID); err != nil {
			t.Fatal(err)
		}
	}
	projects := []config.Project{
		{Name: "Allowed", ProjectID: "allowed", AuthID: "old", DiscoveredBy: []string{"old", "main"}},
		{Name: "Denied", ProjectID: "denied", AuthID: "old", DiscoveredBy: []string{"old"}},
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	cmd := newBindCommand(svc)
	if err := cmd.Flags().Set("auth", "main"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("auth bind = %v", err)
	}
	if !strings.Contains(out.String(), "bound: allowed -> main") || !strings.Contains(out.String(), "Summary: 1 bound, 1 skipped") {
		t.Fatalf("output = %q", out.String())
	}
	denied, err := svc.ProjectByID("denied")
	if err != nil {
		t.Fatal(err)
	}
	if denied.AuthID != "old" {
		t.Fatalf("denied auth = %q, want old", denied.AuthID)
	}
}

func TestAuthBindProjectFlagUsesModePrefixedFiltering(t *testing.T) {
	svc := setupAuthCommandTest(t)
	for _, authID := range []string{"old", "main"} {
		if _, err := svc.AddGCloudAuth(authID, authID); err != nil {
			t.Fatal(err)
		}
	}
	projects := []config.Project{
		{Name: "Alpha App", ProjectID: "alpha", AuthID: "old", DiscoveredBy: []string{"main", "old"}},
		{Name: "Beta App", ProjectID: "beta", AuthID: "old", DiscoveredBy: []string{"main", "old"}},
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	cmd := newBindCommand(svc)
	if err := cmd.Flags().Set("auth", "main"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("project", "=alpha"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cmd.SetOut(&out)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("auth bind = %v", err)
	}
	if strings.Contains(out.String(), "beta") || !strings.Contains(out.String(), "Summary: 1 bound, 0 skipped") {
		t.Fatalf("output = %q", out.String())
	}
}
