package config

import (
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func TestMergeKeyMapUsesConfiguredKeys(t *testing.T) {
	defaults := DefaultKeyMap()
	configured := map[string]map[string][]string{
		string(BlockGlobal): {
			string(ActionQuit): {"x"},
		},
	}
	merged := merge(defaults, configured)
	if got := merged[BlockGlobal][ActionQuit]; len(got) != 1 || got[0] != "x" {
		t.Fatalf("merged quit keys = %v, want [x]", got)
	}
}

func TestCleanKeysDedupesAndDropsEmpty(t *testing.T) {
	got := cleanKeys([]string{"a", "", "a", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("cleanKeys = %v", got)
	}
}

func TestLoadMergesAndPersistsMissingKeys(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))

	state, err := Load()
	if err != nil {
		t.Fatalf("Load = %v", err)
	}
	if !state.Matches(BlockGlobal, ActionQuit, "q") {
		t.Fatal("expected default quit binding after load")
	}

	cfg, err := config.LoadAppConfig()
	if err != nil {
		t.Fatalf("LoadAppConfig = %v", err)
	}
	if _, ok := cfg.Keys[string(BlockGlobal)]; !ok {
		t.Fatalf("config keys = %+v, want global block persisted", cfg.Keys)
	}
}

func TestConflictsReportsDisabledActions(t *testing.T) {
	keys := Clone(DefaultKeyMap())
	keys[BlockProjects][ActionRefresh] = []string{"enter"}
	keys[BlockProjects][ActionSelect] = []string{"enter"}
	state := validate(keys)
	if len(conflicts(state)) == 0 {
		t.Fatal("expected conflicts for duplicate enter binding")
	}
}
