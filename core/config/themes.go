package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	corestyles "github.com/yumauri/fbrcm/core/styles"
)

const (
	themesDirName           = "themes"
	themeFileExtension      = ".toml"
	maximumThemeInheritance = 16
	BuiltInThemeName        = "built-in"
)

const (
	ThemeErrorInvalidArgument = "invalid_argument"
	ThemeErrorNotFound        = "not_found"
	ThemeErrorConflict        = "conflict"
)

// ThemeError classifies theme selection and mutation failures without
// requiring machine-output callers to interpret user-facing text.
type ThemeError struct {
	Kind  string
	Theme string
	Err   error
}

func (e *ThemeError) Error() string { return e.Err.Error() }
func (e *ThemeError) Unwrap() error { return e.Err }

func themeError(kind, theme string, err error) error {
	return &ThemeError{Kind: kind, Theme: theme, Err: err}
}

var hexThemeColor = regexp.MustCompile(`^#[0-9A-Fa-f]{3}([0-9A-Fa-f]{3})?$`)

type themeFile struct {
	Inherits string            `toml:"inherits,omitempty"`
	Colors   map[string]string `toml:"colors,omitempty"`
}

type ThemeResolution struct {
	Name    string
	Path    string
	Palette corestyles.Palette
}

func GetThemesDirPath() string {
	return filepath.Join(GetConfigRootDirPath(), themesDirName)
}

func GetThemeFilePath(name string) (string, error) {
	if err := ValidateThemeName(name); err != nil {
		return "", err
	}
	return filepath.Join(GetThemesDirPath(), name+themeFileExtension), nil
}

func ValidateThemeName(name string) error {
	if name == BuiltInThemeName {
		return themeError(ThemeErrorInvalidArgument, name, fmt.Errorf("theme name %q is reserved", name))
	}
	if err := validatePathSegment(name, "theme name"); err != nil {
		return themeError(ThemeErrorInvalidArgument, name, err)
	}
	return nil
}

// LoadTheme resolves one theme and its inheritance chain over the built-in
// palette. It reads only the selected files and never creates theme state.
func LoadTheme(name string) (ThemeResolution, error) {
	if err := ValidateThemeName(name); err != nil {
		return ThemeResolution{}, err
	}
	palette := corestyles.DefaultPalette()
	path, err := loadThemeInto(name, palette, map[string]bool{}, 0)
	if err != nil {
		return ThemeResolution{}, err
	}
	return ThemeResolution{Name: name, Path: path, Palette: palette}, nil
}

func loadThemeInto(name string, palette corestyles.Palette, visiting map[string]bool, depth int) (string, error) {
	if depth >= maximumThemeInheritance {
		return "", fmt.Errorf("theme inheritance exceeds %d levels at %q", maximumThemeInheritance, name)
	}
	if err := ValidateThemeName(name); err != nil {
		return "", fmt.Errorf("invalid inherited theme %q: %w", name, err)
	}
	if visiting[name] {
		return "", fmt.Errorf("theme inheritance cycle includes %q", name)
	}
	visiting[name] = true
	defer delete(visiting, name)

	path, raw, err := readExactThemeFile(name)
	if err != nil {
		return "", err
	}
	var file themeFile
	if err := decodeTOMLWithOptions(raw, &file, true); err != nil {
		return "", fmt.Errorf("decode theme %q at %s: %w", name, path, err)
	}
	if file.Inherits != "" {
		if _, err := loadThemeInto(file.Inherits, palette, visiting, depth+1); err != nil {
			return "", fmt.Errorf("resolve parent of theme %q: %w", name, err)
		}
	}
	if err := applyThemeColors(name, file.Colors, palette); err != nil {
		return "", err
	}
	return path, nil
}

func readExactThemeFile(name string) (string, []byte, error) {
	directory := GetThemesDirPath()
	entries, err := os.ReadDir(directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, fmt.Errorf("themes directory does not exist: %s", directory)
		}
		return "", nil, fmt.Errorf("read themes directory %s: %w", directory, err)
	}
	filename := name + themeFileExtension
	for _, entry := range entries {
		if entry.Name() != filename {
			continue
		}
		path := filepath.Join(directory, filename)
		if entry.Type()&os.ModeSymlink != 0 {
			return "", nil, fmt.Errorf("theme %q is a symbolic link", name)
		}
		info, err := entry.Info()
		if err != nil {
			return "", nil, fmt.Errorf("inspect theme %q at %s: %w", name, path, err)
		}
		if !info.Mode().IsRegular() {
			return "", nil, fmt.Errorf("theme %q is not a regular file", name)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return "", nil, fmt.Errorf("read theme %q at %s: %w", name, path, err)
		}
		return path, raw, nil
	}
	return "", nil, fmt.Errorf("theme %q does not exist in %s", name, directory)
}

func applyThemeColors(name string, colors map[string]string, palette corestyles.Palette) error {
	known := make(map[string]struct{}, len(corestyles.SupportedTokens()))
	for _, token := range corestyles.SupportedTokens() {
		known[token] = struct{}{}
	}
	unknown := make([]string, 0)
	for token, value := range colors {
		if _, ok := known[token]; !ok {
			unknown = append(unknown, token)
			continue
		}
		if err := validateThemeColor(value); err != nil {
			return fmt.Errorf("theme %q color %q: %w", name, token, err)
		}
		palette[token] = value
	}
	if len(unknown) > 0 {
		slices.Sort(unknown)
		label := "color"
		if len(unknown) > 1 {
			label = "colors"
		}
		return fmt.Errorf("theme %q has unknown %s %s", name, label, strings.Join(unknown, ", "))
	}
	return nil
}

func validateThemeColor(value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("value cannot have surrounding whitespace")
	}
	if hexThemeColor.MatchString(value) {
		return nil
	}
	index, err := strconv.Atoi(value)
	if err == nil && index >= 0 && index <= 255 {
		return nil
	}
	return fmt.Errorf("value must be #RGB, #RRGGBB, or an ANSI index from 0 through 255")
}

// ResolveConfiguredTheme resolves the effective global/local theme setting.
// An absent setting selects the built-in palette without reading themes/.
func ResolveConfiguredTheme() (ThemeResolution, error) {
	resolved, err := ResolveAppConfig()
	if err != nil {
		return ThemeResolution{}, err
	}
	name := resolved.Effective.Theme
	if name == "" {
		return ThemeResolution{Palette: corestyles.DefaultPalette()}, nil
	}
	return LoadTheme(name)
}

// ApplyConfiguredTheme resets the current palette first, guaranteeing that
// every error leaves the application on the complete built-in palette.
func ApplyConfiguredTheme() (ThemeResolution, error) {
	corestyles.ResetPalette()
	resolved, err := ResolveConfiguredTheme()
	if err != nil {
		return ThemeResolution{Palette: corestyles.DefaultPalette()}, err
	}
	corestyles.ApplyPalette(resolved.Palette)
	return resolved, nil
}

func ValidateConfiguredTheme(name string) error {
	if name == "" {
		return nil
	}
	_, err := LoadTheme(name)
	return err
}
