package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corestyles "github.com/yumauri/fbrcm/core/styles"
)

func TestLoadThemeResolvesInheritanceOverBuiltInPalette(t *testing.T) {
	setupTestDirs(t)
	writeFile(t, filepath.Join(GetThemesDirPath(), "base.toml"), `[colors]
primary = "#112233"
text = "245"
`, 0o600)
	writeFile(t, filepath.Join(GetThemesDirPath(), "child.toml"), `inherits = "base"

[colors]
primary = "#abcdef"
error = "160"
`, 0o600)

	resolved, err := LoadTheme("child")
	if err != nil {
		t.Fatalf("LoadTheme = %v", err)
	}
	if resolved.Name != "child" || resolved.Path != filepath.Join(GetThemesDirPath(), "child.toml") {
		t.Fatalf("resolution = %#v", resolved)
	}
	if got := resolved.Palette[corestyles.TokenPrimary]; got != "#abcdef" {
		t.Fatalf("primary = %q", got)
	}
	if got := resolved.Palette[corestyles.TokenText]; got != "245" {
		t.Fatalf("inherited text = %q", got)
	}
	if got := resolved.Palette[corestyles.TokenError]; got != "160" {
		t.Fatalf("error = %q", got)
	}
	if got := resolved.Palette[corestyles.TokenHighlight]; got != corestyles.DefaultPalette()[corestyles.TokenHighlight] {
		t.Fatalf("built-in highlight = %q", got)
	}
}

func TestConfiguredThemeUsesLocalOverride(t *testing.T) {
	setupTestDirs(t)
	writeFile(t, filepath.Join(GetThemesDirPath(), "global.toml"), "[colors]\nprimary = \"1\"\n", 0o600)
	writeFile(t, filepath.Join(GetThemesDirPath(), "local.toml"), "[colors]\nprimary = \"2\"\n", 0o600)
	writeFile(t, GetGlobalConfigFilePath(), "theme = \"global\"\n", 0o600)
	writeFile(t, filepath.Join(filepath.Dir(GetConfigRootDirPath()), ".fbrcm.toml"), "theme = \"local\"\n", 0o600)

	resolved, err := ResolveConfiguredTheme()
	if err != nil {
		t.Fatalf("ResolveConfiguredTheme = %v", err)
	}
	if resolved.Name != "local" || resolved.Palette[corestyles.TokenPrimary] != "2" {
		t.Fatalf("resolution = %#v", resolved)
	}
}

func TestResolveConfiguredThemeWithoutSelectionDoesNotCreateThemesDirectory(t *testing.T) {
	setupTestDirs(t)

	resolved, err := ResolveConfiguredTheme()
	if err != nil {
		t.Fatalf("ResolveConfiguredTheme = %v", err)
	}
	if resolved.Name != "" {
		t.Fatalf("name = %q", resolved.Name)
	}
	if _, err := os.Stat(GetThemesDirPath()); !os.IsNotExist(err) {
		t.Fatalf("themes directory stat = %v, want not exist", err)
	}
}

func TestLoadThemeRejectsInvalidFilesAndInheritance(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		load    string
		message string
	}{
		{name: "missing", load: "missing", message: "does not exist"},
		{name: "invalid TOML", files: map[string]string{"bad.toml": "[colors\n"}, load: "bad", message: "decode theme"},
		{name: "unknown color", files: map[string]string{"bad.toml": "[colors]\nunknown = \"1\"\n"}, load: "bad", message: "unknown color unknown"},
		{name: "invalid color", files: map[string]string{"bad.toml": "[colors]\nprimary = \"red\"\n"}, load: "bad", message: "must be #RGB"},
		{name: "missing parent", files: map[string]string{"bad.toml": "inherits = \"missing\"\n"}, load: "bad", message: "resolve parent"},
		{name: "cycle", files: map[string]string{"one.toml": "inherits = \"two\"\n", "two.toml": "inherits = \"one\"\n"}, load: "one", message: "cycle"},
		{name: "case exact", files: map[string]string{"Nord.toml": "[colors]\nprimary = \"1\"\n"}, load: "nord", message: "does not exist"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setupTestDirs(t)
			for name, contents := range test.files {
				writeFile(t, filepath.Join(GetThemesDirPath(), name), contents, 0o600)
			}
			_, err := LoadTheme(test.load)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("LoadTheme error = %v, want %q", err, test.message)
			}
		})
	}
}

func TestApplyConfiguredThemeFallsBackAtomically(t *testing.T) {
	setupTestDirs(t)
	corestyles.ApplyPalette(corestyles.Palette{corestyles.TokenPrimary: "1"})
	t.Cleanup(corestyles.ResetPalette)
	writeFile(t, GetGlobalConfigFilePath(), "theme = \"broken\"\n", 0o600)
	writeFile(t, filepath.Join(GetThemesDirPath(), "broken.toml"), "[colors]\nprimary = \"2\"\nerror = \"invalid\"\n", 0o600)

	if _, err := ApplyConfiguredTheme(); err == nil {
		t.Fatal("ApplyConfiguredTheme error = nil")
	}
	want := corestyles.DefaultPalette()
	got := corestyles.CurrentPalette()
	for token, value := range want {
		if got[token] != value {
			t.Fatalf("fallback %s = %q, want %q", token, got[token], value)
		}
	}
}

func TestListAndImportThemesDoNotCreateDirectoryUntilValidImport(t *testing.T) {
	setupTestDirs(t)

	themes, err := ListThemes()
	if err != nil || len(themes) != 0 {
		t.Fatalf("ListThemes = %v, %v", themes, err)
	}
	if _, err := os.Stat(GetThemesDirPath()); !os.IsNotExist(err) {
		t.Fatalf("themes directory exists before import: %v", err)
	}
	if _, err := ImportTheme("broken", []byte("[colors]\nprimary = \"red\"\n")); err == nil {
		t.Fatal("ImportTheme invalid error = nil")
	}
	if _, err := os.Stat(GetThemesDirPath()); !os.IsNotExist(err) {
		t.Fatalf("themes directory exists after invalid import: %v", err)
	}

	path, err := ImportTheme("firebase", []byte("[colors]\nprimary = \"#FFC400\"\n"))
	if err != nil {
		t.Fatalf("ImportTheme = %v", err)
	}
	if path != filepath.Join(GetThemesDirPath(), "firebase.toml") {
		t.Fatalf("import path = %q", path)
	}
	themes, err = ListThemes()
	if err != nil || len(themes) != 1 || themes[0] != "firebase" {
		t.Fatalf("ListThemes after import = %v, %v", themes, err)
	}
	if _, err := ImportTheme("firebase", []byte("[colors]\nprimary = \"1\"\n")); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate ImportTheme error = %v", err)
	}
}

func TestImportThemesResolvesInheritanceWithinBatch(t *testing.T) {
	setupTestDirs(t)
	imported, err := ImportThemes([]ThemeImport{
		{Name: "child", Data: []byte("inherits = \"base\"\n[colors]\nerror = \"1\"\n")},
		{Name: "base", Data: []byte("[colors]\nprimary = \"#FFC400\"\n")},
	})
	if err != nil {
		t.Fatalf("ImportThemes = %v", err)
	}
	if len(imported) != 2 || imported[0].Name != "child" || imported[1].Name != "base" {
		t.Fatalf("imported = %#v", imported)
	}
	resolved, err := LoadTheme("child")
	if err != nil {
		t.Fatalf("LoadTheme = %v", err)
	}
	if resolved.Palette[corestyles.TokenPrimary] != "#FFC400" || resolved.Palette[corestyles.TokenError] != "1" {
		t.Fatalf("palette = %#v", resolved.Palette)
	}
}

func TestImportThemesValidatesWholeBatchBeforeWriting(t *testing.T) {
	setupTestDirs(t)
	_, err := ImportThemes([]ThemeImport{
		{Name: "valid", Data: []byte("[colors]\nprimary = \"1\"\n")},
		{Name: "invalid", Data: []byte("[colors]\nunknown = \"2\"\n")},
	})
	if err == nil {
		t.Fatal("ImportThemes error = nil")
	}
	if _, statErr := os.Stat(GetThemesDirPath()); !os.IsNotExist(statErr) {
		t.Fatalf("themes directory exists after failed batch: %v", statErr)
	}
}

func TestImportThemesSkippingExistingUsesInstalledParent(t *testing.T) {
	setupTestDirs(t)
	writeFile(t, filepath.Join(GetThemesDirPath(), "base.toml"), "[colors]\nprimary = \"1\"\n", 0o600)

	imported, skipped, err := ImportThemesSkippingExisting([]ThemeImport{
		{Name: "base", Data: []byte("this source is intentionally invalid")},
		{Name: "child", Data: []byte("inherits = \"base\"\n[colors]\nerror = \"2\"\n")},
	})
	if err != nil {
		t.Fatalf("ImportThemesSkippingExisting = %v", err)
	}
	if len(imported) != 1 || imported[0].Name != "child" {
		t.Fatalf("imported = %#v", imported)
	}
	if len(skipped) != 1 || skipped[0].Name != "base" {
		t.Fatalf("skipped = %#v", skipped)
	}
	resolved, err := LoadTheme("child")
	if err != nil {
		t.Fatalf("LoadTheme(child) = %v", err)
	}
	if resolved.Palette[corestyles.TokenPrimary] != "1" || resolved.Palette[corestyles.TokenError] != "2" {
		t.Fatalf("resolved child palette = %#v", resolved.Palette)
	}
}

func TestSetConfiguredThemeSupportsGlobalAndLocalScopes(t *testing.T) {
	setupTestDirs(t)
	writeFile(t, filepath.Join(GetThemesDirPath(), "one.toml"), "[colors]\nprimary = \"1\"\n", 0o600)
	writeFile(t, filepath.Join(GetThemesDirPath(), "two.toml"), "[colors]\nprimary = \"2\"\n", 0o600)

	if err := SetConfiguredTheme("one", ThemeScopeGlobal); err != nil {
		t.Fatalf("SetConfiguredTheme global = %v", err)
	}
	resolved, err := ResolveAppConfig()
	if err != nil || resolved.Global.Config.Theme != "one" || resolved.Effective.Theme != "one" {
		t.Fatalf("global resolution = %#v, %v", resolved, err)
	}
	if err := SetConfiguredTheme("two", ThemeScopeLocal); err != nil {
		t.Fatalf("SetConfiguredTheme local = %v", err)
	}
	resolved, err = ResolveAppConfig()
	if err != nil || resolved.Local.Config.Theme != "two" || resolved.Effective.Theme != "two" {
		t.Fatalf("local resolution = %#v, %v", resolved, err)
	}
	if err := SetConfiguredTheme(BuiltInThemeName, ThemeScopeLocal); err != nil {
		t.Fatalf("SetConfiguredTheme built-in = %v", err)
	}
	resolved, err = ResolveAppConfig()
	if err != nil || resolved.Local.Config.Theme != "" || resolved.Effective.Theme != "one" {
		t.Fatalf("built-in switch resolution = %#v, %v", resolved, err)
	}
	if err := ResetConfiguredTheme(ThemeScopeGlobal); err != nil {
		t.Fatalf("ResetConfiguredTheme global = %v", err)
	}
	resolved, err = ResolveAppConfig()
	if err != nil || resolved.Global.Config.Theme != "" || resolved.Effective.Theme != "" {
		t.Fatalf("global reset resolution = %#v, %v", resolved, err)
	}
	if err := SetConfiguredTheme("missing", ThemeScopeGlobal); err == nil {
		t.Fatal("SetConfiguredTheme missing error = nil")
	} else {
		var themeErr *ThemeError
		if !errors.As(err, &themeErr) || themeErr.Kind != ThemeErrorNotFound {
			t.Fatalf("missing error = %#v, %v", themeErr, err)
		}
	}
}

func TestBuiltInThemeNameIsReservedForSelection(t *testing.T) {
	setupTestDirs(t)
	if err := ValidateThemeName(BuiltInThemeName); err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("ValidateThemeName built-in error = %v", err)
	}
	if _, err := ImportTheme(BuiltInThemeName, []byte("[colors]\nprimary = \"1\"\n")); err == nil {
		t.Fatal("ImportTheme built-in error = nil")
	}
}

func TestRenameThemeUpdatesSelectionsAndInheritance(t *testing.T) {
	setupTestDirs(t)
	writeFile(t, filepath.Join(GetThemesDirPath(), "base.toml"), "[colors]\nprimary = \"1\"\n", 0o640)
	writeFile(t, filepath.Join(GetThemesDirPath(), "child.toml"), "inherits = \"base\"\n\n[colors]\nerror = \"2\"\n", 0o600)
	writeFile(t, GetGlobalConfigFilePath(), "theme = \"base\"\n", 0o600)

	if err := RenameTheme("base", "foundation"); err != nil {
		t.Fatalf("RenameTheme = %v", err)
	}
	if _, err := os.Stat(filepath.Join(GetThemesDirPath(), "base.toml")); !os.IsNotExist(err) {
		t.Fatalf("old theme stat = %v", err)
	}
	assertFileMode(t, filepath.Join(GetThemesDirPath(), "foundation.toml"), 0o640)
	resolved, err := ResolveAppConfig()
	if err != nil || resolved.Global.Config.Theme != "foundation" {
		t.Fatalf("renamed selection = %#v, %v", resolved.Global.Config, err)
	}
	child, err := LoadTheme("child")
	if err != nil || child.Palette[corestyles.TokenPrimary] != "1" {
		t.Fatalf("renamed inheritance = %#v, %v", child, err)
	}
}

func TestDeleteThemeRejectsSelectedAndInheritedThemes(t *testing.T) {
	setupTestDirs(t)
	writeFile(t, filepath.Join(GetThemesDirPath(), "base.toml"), "[colors]\nprimary = \"1\"\n", 0o600)
	writeFile(t, filepath.Join(GetThemesDirPath(), "child.toml"), "inherits = \"base\"\n", 0o600)

	if err := DeleteTheme("base"); err == nil || !strings.Contains(err.Error(), "inherited by child") {
		t.Fatalf("delete inherited error = %v", err)
	}
	if err := SetConfiguredTheme("child", ThemeScopeGlobal); err != nil {
		t.Fatal(err)
	}
	if err := DeleteTheme("child"); err == nil || !strings.Contains(err.Error(), "selected theme") {
		t.Fatalf("delete selected error = %v", err)
	}
}
