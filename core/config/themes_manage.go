package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yumauri/fbrcm/core/strfold"
	corestyles "github.com/yumauri/fbrcm/core/styles"
)

const (
	ThemeScopeGlobal = "global"
	ThemeScopeLocal  = "local"
)

// ThemeImport describes one theme file to install.
type ThemeImport struct {
	Name string
	Data []byte
}

// ImportedTheme identifies one successfully installed theme.
type ImportedTheme struct {
	Name string
	Path string
}

// SkippedThemeImport identifies a batch theme import whose destination was
// already present and was therefore left unchanged.
type SkippedThemeImport struct {
	Name string
	Path string
}

// InspectThemeDestination reports whether any filesystem entry already
// occupies the canonical destination for a theme import.
func InspectThemeDestination(name string) (path string, exists bool, err error) {
	path, err = GetThemeFilePath(name)
	if err != nil {
		return "", false, err
	}
	if _, statErr := os.Lstat(path); statErr == nil {
		return path, true, nil
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return "", false, fmt.Errorf("inspect theme destination %s: %w", path, statErr)
	}
	return path, false, nil
}

// ListThemes returns installable regular .toml theme files. Invalid theme
// contents do not prevent discovery; validation occurs when a theme is used.
func ListThemes() ([]string, error) {
	entries, err := os.ReadDir(GetThemesDirPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read themes directory %s: %w", GetThemesDirPath(), err)
	}
	themes := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() || filepath.Ext(entry.Name()) != themeFileExtension {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil || !info.Mode().IsRegular() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), themeFileExtension)
		if ValidateThemeName(name) == nil {
			themes = append(themes, name)
		}
	}
	strfold.Sort(themes)
	return themes, nil
}

// ValidateThemeData validates an uninstalled theme against the built-in
// palette and resolves any installed parent themes without writing files.
func ValidateThemeData(name string, raw []byte) error {
	_, err := prepareThemeImports([]ThemeImport{{Name: name, Data: raw}})
	return err
}

// ImportTheme validates and installs one theme without replacing an existing
// destination. The themes directory is created only after validation passes.
func ImportTheme(name string, raw []byte) (string, error) {
	imported, err := ImportThemes([]ThemeImport{{Name: name, Data: raw}})
	if err != nil {
		return "", err
	}
	return imported[0].Path, nil
}

// ImportThemes validates and installs a set of themes as one operation. Theme
// inheritance may refer to another theme in the same set. Existing files are
// checked before any writes, and newly written files are removed on failure.
func ImportThemes(imports []ThemeImport) ([]ImportedTheme, error) {
	prepared, err := prepareThemeImports(imports)
	if err != nil {
		return nil, err
	}
	for _, item := range prepared {
		if _, err := os.Lstat(item.Path); err == nil {
			return nil, themeError(ThemeErrorConflict, item.Name, fmt.Errorf("theme %q already exists", item.Name))
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("inspect theme destination %s: %w", item.Path, err)
		}
	}
	if len(prepared) == 0 {
		return []ImportedTheme{}, nil
	}
	if err := EnsurePrivateDir(GetThemesDirPath()); err != nil {
		return nil, fmt.Errorf("create themes directory: %w", err)
	}
	result := make([]ImportedTheme, 0, len(prepared))
	for _, item := range prepared {
		if err := WritePrivateFileExclusive(item.Path, item.Data); err != nil {
			for _, written := range result {
				_ = os.Remove(written.Path)
			}
			if errors.Is(err, os.ErrExist) {
				return nil, themeError(ThemeErrorConflict, item.Name, fmt.Errorf("theme %q already exists", item.Name))
			}
			return nil, fmt.Errorf("write theme %q: %w", item.Name, err)
		}
		result = append(result, ImportedTheme{Name: item.Name, Path: item.Path})
	}
	return result, nil
}

// ImportThemesSkippingExisting imports a batch without replacing existing
// destinations. Existing items are removed from the batch before validation,
// allowing new themes to inherit the installed version of a skipped parent.
func ImportThemesSkippingExisting(imports []ThemeImport) ([]ImportedTheme, []SkippedThemeImport, error) {
	pending := make([]ThemeImport, 0, len(imports))
	skipped := make([]SkippedThemeImport, 0)
	seen := make(map[string]struct{}, len(imports))
	for _, item := range imports {
		if err := ValidateThemeName(item.Name); err != nil {
			return nil, nil, err
		}
		if _, exists := seen[item.Name]; exists {
			return nil, nil, themeError(ThemeErrorConflict, item.Name, fmt.Errorf("theme %q appears more than once in the import", item.Name))
		}
		seen[item.Name] = struct{}{}
		path, exists, err := InspectThemeDestination(item.Name)
		if err != nil {
			return nil, nil, err
		}
		if exists {
			skipped = append(skipped, SkippedThemeImport{Name: item.Name, Path: path})
			continue
		}
		pending = append(pending, item)
	}
	imported, err := ImportThemes(pending)
	if err != nil {
		return nil, nil, err
	}
	return imported, skipped, nil
}

type preparedThemeImport struct {
	ThemeImport
	Path string
	File themeFile
}

func prepareThemeImports(imports []ThemeImport) ([]preparedThemeImport, error) {
	prepared := make([]preparedThemeImport, 0, len(imports))
	byName := make(map[string]int, len(imports))
	for _, item := range imports {
		if err := ValidateThemeName(item.Name); err != nil {
			return nil, err
		}
		if _, exists := byName[item.Name]; exists {
			return nil, themeError(ThemeErrorConflict, item.Name, fmt.Errorf("theme %q appears more than once in the import", item.Name))
		}
		var file themeFile
		if err := decodeTOMLWithOptions(item.Data, &file, true); err != nil {
			return nil, themeError(ThemeErrorInvalidArgument, item.Name, fmt.Errorf("decode theme %q: %w", item.Name, err))
		}
		path, err := GetThemeFilePath(item.Name)
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, preparedThemeImport{ThemeImport: item, Path: path, File: file})
		byName[item.Name] = len(prepared) - 1
	}
	for index := range prepared {
		palette := corestyles.DefaultPalette()
		if err := validatePreparedTheme(prepared[index].Name, palette, prepared, byName, map[string]bool{}, 0); err != nil {
			return nil, themeError(ThemeErrorInvalidArgument, prepared[index].Name, err)
		}
	}
	return prepared, nil
}

func validatePreparedTheme(name string, palette corestyles.Palette, pending []preparedThemeImport, byName map[string]int, visiting map[string]bool, depth int) error {
	if depth >= maximumThemeInheritance {
		return fmt.Errorf("theme inheritance exceeds %d levels at %q", maximumThemeInheritance, name)
	}
	index, pendingTheme := byName[name]
	if !pendingTheme {
		_, err := loadThemeInto(name, palette, visiting, depth)
		return err
	}
	if visiting[name] {
		return fmt.Errorf("theme inheritance cycle includes %q", name)
	}
	visiting[name] = true
	defer delete(visiting, name)

	item := pending[index]
	if item.File.Inherits != "" {
		if err := validatePreparedTheme(item.File.Inherits, palette, pending, byName, visiting, depth+1); err != nil {
			return fmt.Errorf("resolve parent of theme %q: %w", name, err)
		}
	}
	if err := applyThemeColors(name, item.File.Colors, palette); err != nil {
		return err
	}
	return nil
}

// SetConfiguredTheme selects an installed theme in global or current local
// configuration. Theme files always remain in the global themes directory.
func SetConfiguredTheme(name, scope string) error {
	if name == BuiltInThemeName {
		return ResetConfiguredTheme(scope)
	}
	if _, _, err := readExactThemeFile(name); err != nil {
		return themeError(ThemeErrorNotFound, name, err)
	}
	if _, err := LoadTheme(name); err != nil {
		return themeError(ThemeErrorInvalidArgument, name, err)
	}
	resolved, err := ResolveAppConfig()
	if err != nil {
		return err
	}
	switch scope {
	case ThemeScopeGlobal:
		candidate := CloneAppConfig(resolved.Global.Config)
		if candidate.Theme == name {
			return nil
		}
		candidate.Theme = name
		return SaveAppConfig(candidate)
	case ThemeScopeLocal:
		if LocalConfigDisabled() {
			return fmt.Errorf("local configuration is disabled")
		}
		candidate := CloneAppConfig(resolved.Local.Config)
		if candidate.Theme == name {
			return nil
		}
		candidate.Theme = name
		raw, err := MarshalAppConfig(candidate)
		if err != nil {
			return fmt.Errorf("encode local config: %w", err)
		}
		return SaveLocalAppConfigRaw(resolved.Local.Path, raw)
	default:
		return themeError(ThemeErrorInvalidArgument, name, fmt.Errorf("unsupported theme scope %q", scope))
	}
}

// ResetConfiguredTheme removes the selected theme from one configuration
// layer, revealing the next layer or the built-in palette.
func ResetConfiguredTheme(scope string) error {
	resolved, err := ResolveAppConfig()
	if err != nil {
		return err
	}
	switch scope {
	case ThemeScopeGlobal:
		candidate := CloneAppConfig(resolved.Global.Config)
		if candidate.Theme == "" {
			return nil
		}
		candidate.Theme = ""
		return SaveAppConfig(candidate)
	case ThemeScopeLocal:
		if LocalConfigDisabled() {
			return fmt.Errorf("local configuration is disabled")
		}
		candidate := CloneAppConfig(resolved.Local.Config)
		if candidate.Theme == "" {
			return nil
		}
		candidate.Theme = ""
		raw, err := MarshalAppConfig(candidate)
		if err != nil {
			return fmt.Errorf("encode local config: %w", err)
		}
		return SaveLocalAppConfigRaw(resolved.Local.Path, raw)
	default:
		return themeError(ThemeErrorInvalidArgument, BuiltInThemeName, fmt.Errorf("unsupported theme scope %q", scope))
	}
}

// EnsureThemeCanDelete prevents removal of a theme selected by the current
// global/local configuration or inherited by another installed theme.
func EnsureThemeCanDelete(name string) error {
	if err := ValidateThemeName(name); err != nil {
		return err
	}
	if _, _, err := readExactThemeFile(name); err != nil {
		return themeError(ThemeErrorNotFound, name, err)
	}
	resolved, err := ResolveAppConfig()
	if err != nil {
		return err
	}
	if resolved.Global.Config.Theme == name || resolved.Local.Config.Theme == name {
		return themeError(ThemeErrorConflict, name, fmt.Errorf("cannot delete selected theme %q; switch themes first", name))
	}
	refs, err := themeInheritanceReferences(name)
	if err != nil {
		return err
	}
	if len(refs) > 0 {
		return themeError(ThemeErrorConflict, name, fmt.Errorf("cannot delete theme %q; inherited by %s", name, strings.Join(refs, ", ")))
	}
	return nil
}

func DeleteTheme(name string) error {
	if err := EnsureThemeCanDelete(name); err != nil {
		return err
	}
	path, err := GetThemeFilePath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete theme %q: %w", name, err)
	}
	return nil
}

func themeInheritanceReferences(parent string) ([]string, error) {
	themes, err := ListThemes()
	if err != nil {
		return nil, err
	}
	refs := make([]string, 0)
	for _, name := range themes {
		if name == parent {
			continue
		}
		path, raw, err := readExactThemeFile(name)
		if err != nil {
			return nil, err
		}
		var file themeFile
		if err := decodeTOMLWithOptions(raw, &file, true); err != nil {
			return nil, themeError(ThemeErrorInvalidArgument, name, fmt.Errorf("decode theme %q at %s: %w", name, path, err))
		}
		if file.Inherits == parent {
			refs = append(refs, name)
		}
	}
	return refs, nil
}

// RenameTheme renames a valid theme and updates direct inheritance references
// plus the current global and local selections. Updated files are rolled back
// if any step fails.
func RenameTheme(oldName, newName string) error {
	if err := ValidateThemeName(oldName); err != nil {
		return fmt.Errorf("old theme: %w", err)
	}
	if err := ValidateThemeName(newName); err != nil {
		return fmt.Errorf("new theme: %w", err)
	}
	oldPath, _, err := readExactThemeFile(oldName)
	if err != nil {
		return themeError(ThemeErrorNotFound, oldName, err)
	}
	if oldName == newName {
		return nil
	}
	if _, err := LoadTheme(oldName); err != nil {
		return themeError(ThemeErrorInvalidArgument, oldName, err)
	}
	newPath, err := GetThemeFilePath(newName)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(newPath); err == nil {
		return themeError(ThemeErrorConflict, newName, fmt.Errorf("theme %q already exists", newName))
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect theme destination %s: %w", newPath, err)
	}

	mutations, err := themeRenameMutations(oldName, newName)
	if err != nil {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename theme %q to %q: %w", oldName, newName, err)
	}
	applied := 0
	for index, mutation := range mutations {
		if err := writeFileAtomicMode(mutation.path, mutation.after, mutation.mode); err != nil {
			for rollback := applied - 1; rollback >= 0; rollback-- {
				_ = writeFileAtomicMode(mutations[rollback].path, mutations[rollback].before, mutations[rollback].mode)
			}
			_ = os.Rename(newPath, oldPath)
			clearSessionAppConfigResolution()
			return fmt.Errorf("update reference in %s: %w", mutations[index].path, err)
		}
		applied++
	}
	clearSessionAppConfigResolution()
	return nil
}

type themeFileMutation struct {
	path   string
	before []byte
	after  []byte
	mode   os.FileMode
}

func themeRenameMutations(oldName, newName string) ([]themeFileMutation, error) {
	themes, err := ListThemes()
	if err != nil {
		return nil, err
	}
	mutations := make([]themeFileMutation, 0)
	for _, name := range themes {
		if name == oldName {
			continue
		}
		path, raw, err := readExactThemeFile(name)
		if err != nil {
			return nil, err
		}
		var file themeFile
		if err := decodeTOMLWithOptions(raw, &file, true); err != nil {
			return nil, themeError(ThemeErrorInvalidArgument, name, fmt.Errorf("decode theme %q at %s: %w", name, path, err))
		}
		if file.Inherits != oldName {
			continue
		}
		file.Inherits = newName
		after, err := encodeTOML(file)
		if err != nil {
			return nil, fmt.Errorf("encode theme %q: %w", name, err)
		}
		mutation, err := newThemeFileMutation(path, raw, after)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, mutation)
	}

	resolved, err := ResolveAppConfig()
	if err != nil {
		return nil, err
	}
	configs := []AppConfigLayer{resolved.Global, resolved.Local}
	for _, item := range configs {
		if !item.Exists || item.Config.Theme != oldName {
			continue
		}
		candidate := CloneAppConfig(item.Config)
		candidate.Theme = newName
		after, err := MarshalAppConfig(candidate)
		if err != nil {
			return nil, fmt.Errorf("encode config %s: %w", item.Path, err)
		}
		before, err := os.ReadFile(item.Path)
		if err != nil {
			return nil, fmt.Errorf("read config %s: %w", item.Path, err)
		}
		mutation, err := newThemeFileMutation(item.Path, before, after)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, mutation)
	}
	return mutations, nil
}

func newThemeFileMutation(path string, before, after []byte) (themeFileMutation, error) {
	info, err := os.Stat(path)
	if err != nil {
		return themeFileMutation{}, fmt.Errorf("inspect file %s: %w", path, err)
	}
	return themeFileMutation{path: path, before: before, after: after, mode: info.Mode().Perm()}, nil
}
