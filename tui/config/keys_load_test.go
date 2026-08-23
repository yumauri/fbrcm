package config

import (
	"errors"
	"os"
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

func TestMergeKeyMapIgnoresLegacyHistoryPairBindings(t *testing.T) {
	defaults := DefaultKeyMap()
	configured := map[string]map[string][]string{
		string(BlockHistory): {
			"both_older": {"-"},
			"both_newer": {"+"},
		},
	}
	merged := merge(defaults, configured)
	if got := merged[BlockHistory][ActionHistoryBothOlder]; len(got) != 1 || got[0] != "," {
		t.Fatalf("older pair keys = %v, want [,]", got)
	}
	if got := merged[BlockHistory][ActionHistoryBothNewer]; len(got) != 1 || got[0] != "." {
		t.Fatalf("newer pair keys = %v, want [.]", got)
	}
}

func TestCleanKeysDedupesAndDropsEmpty(t *testing.T) {
	got := cleanKeys([]string{"a", "", "a", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("cleanKeys = %v", got)
	}
}

func TestLoadMergesDefaultsWithoutPersistingThem(t *testing.T) {
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

	if _, err := os.Stat(config.GetGlobalConfigFilePath()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load persisted generated defaults: %v", err)
	}
}

func TestLoadMigratesGeneratedAdministrationShortcuts(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	configured := toConfigMap(DefaultKeyMap())
	delete(configured[string(BlockGlobal)], string(ActionProfiles))
	configured[string(BlockGlobal)][string(ActionAccounts)] = []string{"A"}
	configured[string(BlockParameters)][string(ActionPublishAll)] = []string{"P"}
	configured[string(BlockConditions)][string(ActionPublishAll)] = []string{"P"}
	if err := config.SaveAppConfig(&config.AppConfig{Keys: configured}); err != nil {
		t.Fatalf("SaveAppConfig = %v", err)
	}

	state, err := Load()
	if err != nil {
		t.Fatalf("Load = %v", err)
	}
	if !state.Matches(BlockGlobal, ActionAccounts, "ctrl+a") {
		t.Fatal("global ctrl+a did not activate Accounts after migration")
	}
	if !state.Matches(BlockGlobal, ActionProfiles, "ctrl+p") {
		t.Fatal("global ctrl+p did not activate Profiles after migration")
	}
	for _, block := range []Block{BlockParameters, BlockConditions} {
		if !state.Matches(block, ActionPublishAll, "P") {
			t.Fatalf("%s publish all did not retain P", block)
		}
	}
}

func TestMigrateAdminShortcutsPreservesCustomizedBindings(t *testing.T) {
	configured := map[string]map[string][]string{
		string(BlockGlobal): {
			string(ActionAccounts): {"alt+a"},
			string(ActionProfiles): {"alt+p"},
		},
		string(BlockParameters): {
			string(ActionPublishAll): {"P"},
		},
	}
	if migrateAdminShortcuts(configured) {
		t.Fatal("migration changed customized administration shortcuts")
	}
	if got := configured[string(BlockParameters)][string(ActionPublishAll)]; len(got) != 1 || got[0] != "P" {
		t.Fatalf("customized publish all binding = %v, want [P]", got)
	}
}

func TestLoadRespectsDisabledPowerlineGlyphs(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Cleanup(func() { powerlineGlyphs = true })

	disabled := false
	if err := config.SaveAppConfig(&config.AppConfig{PowerlineGlyphs: &disabled}); err != nil {
		t.Fatalf("SaveAppConfig = %v", err)
	}
	if _, err := Load(); err != nil {
		t.Fatalf("Load = %v", err)
	}
	if PowerlineGlyphsEnabled() {
		t.Fatal("PowerlineGlyphsEnabled() = true, want false")
	}
}

func TestDefaultKeyMapIncludesAllHistoryPickerBindings(t *testing.T) {
	actions := DefaultKeyMap()[BlockHistoryPicker]
	for _, action := range []Action{ActionCancel, ActionToggle, ActionLeft, ActionRight, ActionHistoryBothOlder, ActionHistoryBothNewer, ActionHistoryRollback, ActionReset, ActionUp, ActionDown, ActionPageUp, ActionPageDown, ActionHome, ActionEnd, ActionSubmit} {
		if len(actions[action]) == 0 {
			t.Fatalf("history picker action %q has no default binding", action)
		}
	}
}

func TestDefaultKeyMapIncludesHistoryChangesToggle(t *testing.T) {
	if got := DefaultKeyMap()[BlockHistory][ActionHistoryChanges]; len(got) != 1 || got[0] != "c" {
		t.Fatalf("history changes toggle = %v, want [c]", got)
	}
	if got := DefaultKeyMap()[BlockHistory][ActionSubmit]; len(got) != 1 || got[0] != "enter" {
		t.Fatalf("history diff = %v, want [enter]", got)
	}
}

func TestDefaultKeyMapIncludesHelpPaletteBindings(t *testing.T) {
	if got := DefaultKeyMap()[BlockGlobal][ActionHelp]; len(got) != 1 || got[0] != "?" {
		t.Fatalf("help action = %v, want [?]", got)
	}
	for _, action := range []Action{ActionCancel, ActionSubmit, ActionUp, ActionDown, ActionPageUp, ActionPageDown, ActionHome, ActionEnd} {
		if len(DefaultKeyMap()[BlockHelp][action]) == 0 {
			t.Fatalf("help palette action %q has no default binding", action)
		}
	}
}

func TestAboutHasNoDirectKeyBinding(t *testing.T) {
	if _, ok := DefaultKeyMap()[BlockGlobal][ActionAbout]; ok {
		t.Fatal("About has a direct global key binding")
	}
	if keys := Keys(BlockGlobal, ActionAbout); len(keys) != 0 {
		t.Fatalf("About keys = %v, want none", keys)
	}
}

func TestDefaultKeyMapIncludesWorkspaceMenuBindings(t *testing.T) {
	if got := DefaultKeyMap()[BlockGlobal][ActionWorkspaceMenu]; len(got) != 1 || got[0] != "\\" {
		t.Fatalf("workspace menu action = %v, want [\\\\]", got)
	}
	for _, action := range []Action{ActionCancel, ActionSubmit, ActionUp, ActionDown, ActionHome, ActionEnd} {
		if len(DefaultKeyMap()[BlockWorkspaceMenu][action]) == 0 {
			t.Fatalf("workspace menu action %q has no default binding", action)
		}
	}
}

func TestDefaultKeyMapIncludesAccountsBinding(t *testing.T) {
	if got := DefaultKeyMap()[BlockGlobal][ActionAccounts]; len(got) != 1 || got[0] != "ctrl+a" {
		t.Fatalf("accounts action = %v, want [ctrl+a]", got)
	}
}

func TestDefaultKeyMapIncludesProfilesAndProjectAuthBindings(t *testing.T) {
	if got := DefaultKeyMap()[BlockGlobal][ActionProfiles]; len(got) != 1 || got[0] != "ctrl+p" {
		t.Fatalf("profiles action = %v, want [ctrl+p]", got)
	}
	if got := DefaultKeyMap()[BlockProjects][ActionBindAuth]; len(got) != 1 || got[0] != "b" {
		t.Fatalf("project auth action = %v, want [b]", got)
	}
}

func TestDefaultKeyMapIncludesProjectDefaultsBinding(t *testing.T) {
	if got := DefaultKeyMap()[BlockProjects][ActionDefaults]; len(got) != 1 || got[0] != "d" {
		t.Fatalf("projects defaults keys = %v, want [d]", got)
	}
}

func TestDefaultKeyMapIncludesProjectTemplateBindings(t *testing.T) {
	if got := DefaultKeyMap()[BlockProjects][ActionToggleTemplates]; len(got) != 1 || got[0] != "t" {
		t.Fatalf("projects template toggle keys = %v, want [t]", got)
	}
	if got := DefaultKeyMap()[BlockProjects][ActionMakePrimary]; len(got) != 1 || got[0] != "p" {
		t.Fatalf("projects make-primary keys = %v, want [p]", got)
	}
}

func TestDefaultKeyMapIncludesInAppDefaultToggle(t *testing.T) {
	got := DefaultKeyMap()[BlockDetails][ActionToggleInAppDefault]
	if len(got) != 1 || got[0] != "d" {
		t.Fatalf("Details in-app-default binding = %v, want [d]", got)
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
