package config

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestProjectVisibilityDefaultsAndRoundTrip(t *testing.T) {
	setupTestDirs(t)
	policy, err := LoadProjectVisibility()
	if err != nil || !reflect.DeepEqual(policy, DefaultProjectVisibility()) {
		t.Fatalf("default = %#v, %v", policy, err)
	}
	if _, err := os.Stat(GetProjectVisibilityFilePath()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read created preferences: %v", err)
	}
	policy.Blacklist.ProjectIDs = []string{"zeta", "alpha", "alpha"}
	policy.Whitelist.Enabled = true
	policy.Whitelist.ProjectIDs = []string{"alpha", "beta"}
	if err := SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadProjectVisibility()
	if err != nil || !reflect.DeepEqual(loaded.Blacklist.ProjectIDs, []string{"alpha", "zeta"}) {
		t.Fatalf("round trip = %#v, %v", loaded, err)
	}
	assertFileMode(t, GetProjectVisibilityFilePath(), PrivateFileMode)
	for id, want := range map[string]string{"alpha": "blacklisted", "server@alpha": "blacklisted", "beta": "", "server@beta": "", "gamma": "not_whitelisted"} {
		if got := loaded.HiddenReason(id); got != want {
			t.Errorf("HiddenReason(%q) = %q, want %q", id, got, want)
		}
	}
	loaded.Whitelist.ProjectIDs = []string{}
	if loaded.HiddenReason("beta") != "not_whitelisted" {
		t.Fatal("empty enabled whitelist must hide every project")
	}
	loaded.Whitelist.Enabled = false
	loaded.Blacklist.Enabled = false
	if loaded.HiddenReason("alpha") != "" {
		t.Fatal("disabled lists must retain membership without hiding projects")
	}
}

func TestProjectVisibilitySurvivesResetAndIsProfileScoped(t *testing.T) {
	setupTestDirs(t)
	policy := DefaultProjectVisibility()
	policy.Blacklist.ProjectIDs = []string{"alpha"}
	if err := SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	if _, err := ResetProjects(); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadProjectVisibility()
	if err != nil || loaded.HiddenReason("alpha") != "blacklisted" {
		t.Fatalf("reset erased visibility: %#v, %v", loaded, err)
	}
	if err := SwitchProfile("other"); err != nil {
		t.Fatal(err)
	}
	loaded, err = LoadProjectVisibility()
	if err != nil || loaded.HiddenReason("alpha") != "" {
		t.Fatalf("preferences leaked across profiles: %#v, %v", loaded, err)
	}
}

func TestProjectVisibilityRejectsInvalidFilesAndIDs(t *testing.T) {
	setupTestDirs(t)
	for _, raw := range []string{"", "null", "{", `{"version":2}`, `{"version":1,"blacklist":{"project_ids":["server@alpha"]}}`} {
		writeFile(t, GetProjectVisibilityFilePath(), raw, PrivateFileMode)
		if _, err := LoadProjectVisibility(); err == nil {
			t.Errorf("accepted malformed preferences %q", raw)
		}
	}
	for _, id := range []string{"", " alpha", "alpha beta", "=alpha", "client@alpha"} {
		policy := DefaultProjectVisibility()
		policy.Blacklist.ProjectIDs = []string{id}
		if err := SaveProjectVisibility(policy); err == nil {
			t.Errorf("accepted ID %q", id)
		}
	}
}
