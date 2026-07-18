package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
)

func TestInspectStartupStateEmptyProfile(t *testing.T) {
	svc := setupCoreTestEnv(t)

	state, err := svc.InspectStartupState()
	if err != nil {
		t.Fatalf("InspectStartupState = %v", err)
	}
	if state.Profile != config.DefaultProfileName {
		t.Fatalf("profile = %q, want %q", state.Profile, config.DefaultProfileName)
	}
	if len(state.Profiles) != 1 || state.Profiles[0] != config.DefaultProfileName {
		t.Fatalf("profiles = %v, want [default]", state.Profiles)
	}
	if len(state.Auth) != 0 || len(state.Projects) != 0 || state.DefaultAuthID != "" {
		t.Fatalf("state = %+v, want empty auth and projects", state)
	}
}

func TestSwitchProfileKeepsFirebaseClientCacheProfileScoped(t *testing.T) {
	svc := setupCoreTestEnv(t)
	defaultKey := firebaseClientKey("main")
	if err := svc.SwitchProfile("work"); err != nil {
		t.Fatalf("SwitchProfile work = %v", err)
	}
	workKey := firebaseClientKey("main")
	if defaultKey == workKey {
		t.Fatalf("firebase client keys collide across profiles: %q", defaultKey)
	}
	state, err := svc.InspectStartupState()
	if err != nil {
		t.Fatalf("InspectStartupState work = %v", err)
	}
	if state.Profile != "work" || len(state.Profiles) != 2 {
		t.Fatalf("state = %+v, want active work with two profiles", state)
	}
}

func TestInspectStartupStateReturnsCachedState(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")

	state, err := svc.InspectStartupState()
	if err != nil {
		t.Fatalf("InspectStartupState = %v", err)
	}
	if len(state.Auth) != 1 || state.Auth[0].ID != "main" || state.DefaultAuthID != "main" {
		t.Fatalf("auth state = %+v default=%q, want main", state.Auth, state.DefaultAuthID)
	}
	if len(state.Projects) != 1 || state.Projects[0].ProjectID != "demo" {
		t.Fatalf("projects = %+v, want demo", state.Projects)
	}
}

func TestInspectStartupStateRejectsCorruptProjectsCache(t *testing.T) {
	svc := setupCoreTestEnv(t)
	path := config.GetProjectsFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir projects dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write corrupt projects cache: %v", err)
	}

	if _, err := svc.InspectStartupState(); err == nil {
		t.Fatal("InspectStartupState corrupt projects = nil, want error")
	}
}

func TestInspectStartupStateAcceptsEmptyProjectsFile(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects(nil, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects empty = %v", err)
	}

	state, err := svc.InspectStartupState()
	if err != nil {
		t.Fatalf("InspectStartupState = %v", err)
	}
	if len(state.Projects) != 0 {
		t.Fatalf("projects = %+v, want empty", state.Projects)
	}
}
