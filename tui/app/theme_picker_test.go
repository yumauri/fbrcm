package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	corestyles "github.com/yumauri/fbrcm/core/styles"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestThemePickerSelectsActiveThemeAndRestoresItOnCancel(t *testing.T) {
	setupThemePickerTest(t)
	writeThemePickerTheme(t, "firebase", "1")
	if err := coreconfig.SetConfiguredTheme("firebase", coreconfig.ThemeScopeGlobal); err != nil {
		t.Fatalf("SetConfiguredTheme = %v", err)
	}
	if _, err := coreconfig.ApplyConfiguredTheme(); err != nil {
		t.Fatalf("ApplyConfiguredTheme = %v", err)
	}

	m := viewTestModel(90, 24, panels.Projects)
	m, _, handled := m.openThemePicker()
	if !handled || !m.themePicker.IsOpen() {
		t.Fatal("theme picker did not open")
	}
	if option, ok := m.themePicker.Current(); !ok || option.Name != "firebase" {
		t.Fatalf("selected option = %#v, %t; want firebase", option, ok)
	}

	next, _, handled := m.updateThemePicker(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	if !handled || corestyles.ColorValue(corestyles.TokenPrimary) != corestyles.DefaultPalette()[corestyles.TokenPrimary] {
		t.Fatalf("built-in preview was not applied: %q", corestyles.ColorValue(corestyles.TokenPrimary))
	}
	next, _, handled = next.updateThemePicker(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	if !handled || next.themePicker.IsOpen() {
		t.Fatal("theme picker did not close on escape")
	}
	if got := corestyles.ColorValue(corestyles.TokenPrimary); got != "1" {
		t.Fatalf("cancel restored primary = %q, want 1", got)
	}
}

func TestThemePickerLivePreviewAndEnterPersistSelection(t *testing.T) {
	setupThemePickerTest(t)
	writeThemePickerTheme(t, "firebase", "2")

	m := viewTestModel(90, 24, panels.Logs)
	m, _, _ = m.openThemePicker()
	if m.mouseMode() != tea.MouseModeAllMotion {
		t.Fatalf("mouse mode = %v, want all motion while picker is open", m.mouseMode())
	}

	next, _, _ := m.updateThemePicker(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	if got := corestyles.ColorValue(corestyles.TokenPrimary); got != "2" {
		t.Fatalf("live preview primary = %q, want 2", got)
	}
	next, cmd, _ := next.updateThemePicker(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if cmd == nil || !next.themePicker.Saving() {
		t.Fatal("Enter did not start theme persistence")
	}
	saved, ok := cmd().(themeSavedMsg)
	if !ok || saved.err != nil {
		t.Fatalf("save result = %#v", saved)
	}
	next, _, handled := next.updateThemePicker(saved)
	if !handled || next.themePicker.IsOpen() {
		t.Fatal("successful save did not close picker")
	}
	global, err := coreconfig.LoadGlobalAppConfig()
	if err != nil {
		t.Fatalf("LoadGlobalAppConfig = %v", err)
	}
	if global.Theme != "firebase" {
		t.Fatalf("global theme = %q, want firebase", global.Theme)
	}
}

func TestThemePickerQuitRestoresPreviewAndQuits(t *testing.T) {
	setupThemePickerTest(t)
	writeThemePickerTheme(t, "firebase", "2")

	m := viewTestModel(90, 24, panels.Projects)
	m, _, _ = m.openThemePicker()
	m, _, _ = m.updateThemePicker(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	if got := corestyles.ColorValue(corestyles.TokenPrimary); got != "2" {
		t.Fatalf("live preview primary = %q, want 2", got)
	}

	next, cmd, handled := m.updateThemePicker(tea.KeyPressMsg(tea.Key{Code: 'q'}))
	if !handled || next.themePicker.IsOpen() {
		t.Fatal("q did not close the theme picker")
	}
	if got := corestyles.ColorValue(corestyles.TokenPrimary); got != corestyles.DefaultPalette()[corestyles.TokenPrimary] {
		t.Fatalf("q restored primary = %q, want built-in value", got)
	}
	if cmd == nil {
		t.Fatal("q did not request application quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("q command did not return tea.QuitMsg")
	}
}

func TestThemePickerReloadsSelectedThemeAndItsParents(t *testing.T) {
	setupThemePickerTest(t)
	writeThemePickerTheme(t, "base", "1")
	writeThemePickerThemeFile(t, "child", "inherits = \"base\"\n")
	if err := coreconfig.SetConfiguredTheme("child", coreconfig.ThemeScopeGlobal); err != nil {
		t.Fatalf("SetConfiguredTheme = %v", err)
	}
	if _, err := coreconfig.ApplyConfiguredTheme(); err != nil {
		t.Fatalf("ApplyConfiguredTheme = %v", err)
	}

	m := viewTestModel(90, 24, panels.Projects)
	m, watchCmd, _ := m.openThemePicker()
	defer m.stopThemeWatcher()
	if watchCmd == nil || m.themeWatcher == nil {
		t.Fatal("opening a picker with installed themes did not start the watcher")
	}
	if option, ok := m.themePicker.Current(); !ok || option.Name != "child" {
		t.Fatalf("selected option = %#v, %t; want child", option, ok)
	}

	writeThemePickerTheme(t, "base", "2")
	m, _, handled := m.updateThemePicker(themeWatchReloadMsg{
		session:  m.themeWatcher,
		revision: m.themeWatchRevision,
	})
	if !handled {
		t.Fatal("theme reload message was not handled")
	}
	if got := corestyles.ColorValue(corestyles.TokenPrimary); got != "2" {
		t.Fatalf("inherited live reload primary = %q, want 2", got)
	}
	if option, _ := m.themePicker.Current(); option.Palette[corestyles.TokenPrimary] != "2" {
		t.Fatalf("picker palette primary = %q, want 2", option.Palette[corestyles.TokenPrimary])
	}

	writeThemePickerThemeFile(t, "base", "[colors]\nprimary = \"not-a-color\"\n")
	m, _, _ = m.updateThemePicker(themeWatchReloadMsg{
		session:  m.themeWatcher,
		revision: m.themeWatchRevision,
	})
	if got := corestyles.ColorValue(corestyles.TokenPrimary); got != "2" {
		t.Fatalf("invalid reload changed primary to %q, want last valid value 2", got)
	}
	if option, _ := m.themePicker.Current(); option.Palette[corestyles.TokenPrimary] != "2" {
		t.Fatalf("invalid reload changed picker palette to %q, want 2", option.Palette[corestyles.TokenPrimary])
	}
	if view := m.themePicker.View(); !strings.Contains(view, "reload theme") {
		t.Fatalf("invalid reload error is not visible:\n%s", view)
	}

	writeThemePickerTheme(t, "base", "3")
	m, _, _ = m.updateThemePicker(themeWatchReloadMsg{
		session:  m.themeWatcher,
		revision: m.themeWatchRevision,
	})
	if got := corestyles.ColorValue(corestyles.TokenPrimary); got != "3" {
		t.Fatalf("recovered live reload primary = %q, want 3", got)
	}
	if view := m.themePicker.View(); strings.Contains(view, "reload theme") {
		t.Fatalf("successful reload did not clear the error:\n%s", view)
	}
}

func TestThemePickerDoesNotCreateMissingThemesDirectoryForWatcher(t *testing.T) {
	setupThemePickerTest(t)

	m := viewTestModel(90, 24, panels.Projects)
	m, cmd, _ := m.openThemePicker()
	if cmd != nil || m.themeWatcher != nil {
		t.Fatal("picker started a watcher without a themes directory")
	}
	if _, err := os.Stat(coreconfig.GetThemesDirPath()); !os.IsNotExist(err) {
		t.Fatalf("themes directory stat error = %v, want not exist", err)
	}
}

func TestPersistBuiltInClearsGlobalAndLocalThemeSelections(t *testing.T) {
	root := setupThemePickerTestWithLocalConfig(t)
	writeThemePickerTheme(t, "global", "1")
	writeThemePickerTheme(t, "local", "2")
	if err := coreconfig.SetConfiguredTheme("global", coreconfig.ThemeScopeGlobal); err != nil {
		t.Fatalf("set global theme = %v", err)
	}
	localPath := filepath.Join(root, "workspace", coreconfig.LocalConfigFileName)
	if err := os.WriteFile(localPath, []byte("theme = \"local\"\n"), 0o600); err != nil {
		t.Fatalf("write local config = %v", err)
	}
	coreconfig.SetLocalConfigDisabled(false)

	if err := persistEffectiveTheme(coreconfig.BuiltInThemeName); err != nil {
		t.Fatalf("persistEffectiveTheme(built-in) = %v", err)
	}
	resolved, err := coreconfig.ResolveAppConfig()
	if err != nil {
		t.Fatalf("ResolveAppConfig = %v", err)
	}
	if resolved.Global.Config.Theme != "" || resolved.Local.Config.Theme != "" || resolved.Effective.Theme != "" {
		t.Fatalf("theme selections remain: global=%q local=%q effective=%q", resolved.Global.Config.Theme, resolved.Local.Config.Theme, resolved.Effective.Theme)
	}
}

func setupThemePickerTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	coreconfig.SetLocalConfigDisabled(true)
	t.Cleanup(func() {
		coreconfig.SetLocalConfigDisabled(false)
		corestyles.ResetPalette()
	})
	if _, err := tuiconfig.Load(); err != nil {
		t.Fatalf("load TUI config = %v", err)
	}
	return root
}

func setupThemePickerTestWithLocalConfig(t *testing.T) string {
	t.Helper()
	root := setupThemePickerTest(t)
	workspace := filepath.Join(root, "workspace")
	if err := os.MkdirAll(workspace, 0o700); err != nil {
		t.Fatalf("create workspace = %v", err)
	}
	t.Chdir(workspace)
	coreconfig.SetLocalConfigDisabled(false)
	return root
}

func writeThemePickerTheme(t *testing.T, name, primary string) {
	t.Helper()
	writeThemePickerThemeFile(t, name, "[colors]\nprimary = \""+primary+"\"\n")
}

func writeThemePickerThemeFile(t *testing.T, name, contents string) {
	t.Helper()
	if err := os.MkdirAll(coreconfig.GetThemesDirPath(), 0o700); err != nil {
		t.Fatalf("create themes directory = %v", err)
	}
	path, err := coreconfig.GetThemeFilePath(name)
	if err != nil {
		t.Fatalf("GetThemeFilePath(%q) = %v", name, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write theme %q = %v", name, err)
	}
}
