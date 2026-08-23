package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
)

const (
	DefaultProfileName = "default"
	profilesDirName    = "profiles"
)

const (
	ProfileErrorInvalidArgument = "invalid_argument"
	ProfileErrorNotFound        = "not_found"
	ProfileErrorConflict        = "conflict"
)

// ProfileError classifies profile selection and mutation failures without
// requiring machine-output callers to interpret user-facing text.
type ProfileError struct {
	Kind    string
	Profile string
	Err     error
}

func (e *ProfileError) Error() string { return e.Err.Error() }
func (e *ProfileError) Unwrap() error { return e.Err }

func profileError(kind, profile string, err error) error {
	return &ProfileError{Kind: kind, Profile: profile, Err: err}
}

func GetActiveProfileName() string {
	return getPaths().profile
}

func EnsureActiveProfile() error {
	if profile, overridden := selectedProfileOverride(); overridden {
		if err := ValidateProfileName(profile); err != nil {
			return fmt.Errorf("selected profile: %w", err)
		}
		if !profileConfigDirExists(profile) {
			return profileError(ProfileErrorNotFound, profile, fmt.Errorf("selected profile %q does not exist; create it with `fbrcm profile switch %s`", profile, profile))
		}
		if err := ensureProfileDirs(profile, false); err != nil {
			return err
		}
		corelog.For("config").Info("current profile", "profile", profile, "override", true)
		return nil
	}
	resolved, err := ResolveAppConfig()
	if err != nil {
		return err
	}
	if !resolved.Global.Exists && !resolved.Local.Exists {
		return SwitchProfile(DefaultProfileName)
	}
	profile := resolved.Effective.Profile
	if strings.TrimSpace(profile) == "" {
		return SwitchProfile(DefaultProfileName)
	}
	if err := ValidateProfileName(profile); err != nil {
		return err
	}
	if !profileConfigDirExists(profile) {
		source := resolved.Global.Path
		if strings.TrimSpace(resolved.Local.Config.Profile) != "" {
			source = resolved.Local.Path
		}
		err := fmt.Errorf("active profile %q selected by %s does not exist in config directory; create it with `fbrcm profile switch %s`", profile, source, profile)
		corelog.For("config").Error("active profile missing", "profile", profile, "config_dir", profileConfigDir(profile), "err", err)
		return profileError(ProfileErrorNotFound, profile, err)
	}
	if err := ensureProfileDirs(profile, false); err != nil {
		return err
	}
	corelog.For("config").Info("current profile", "profile", profile)
	return nil
}

func ListProfiles() ([]string, error) {
	seen := map[string]struct{}{}
	root := filepath.Join(GetConfigRootDirPath(), profilesDirName)
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read profiles root %s: %w", root, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if ValidateProfileName(name) == nil {
			seen[name] = struct{}{}
		}
	}

	profiles := make([]string, 0, len(seen))
	for name := range seen {
		profiles = append(profiles, name)
	}
	strfold.Sort(profiles)
	return profiles, nil
}

func GetProfileConfigDirPath(name string) (string, error) {
	if err := ValidateProfileName(name); err != nil {
		return "", err
	}
	return profileConfigDir(name), nil
}

func GetProfileCacheDirPath(name string) (string, error) {
	if err := ValidateProfileName(name); err != nil {
		return "", err
	}
	return profileCacheDir(name), nil
}

func DeleteProfile(name string) error {
	if err := EnsureProfileCanDelete(name); err != nil {
		return err
	}
	if err := os.RemoveAll(profileConfigDir(name)); err != nil {
		return fmt.Errorf("remove profile config dir: %w", err)
	}
	if err := os.RemoveAll(profileCacheDir(name)); err != nil {
		return fmt.Errorf("remove profile cache dir: %w", err)
	}
	corelog.For("config").Info("profile deleted", "profile", name, "config_dir", profileConfigDir(name), "cache_dir", profileCacheDir(name))
	return nil
}

func EnsureProfileCanDelete(name string) error {
	if err := ValidateProfileName(name); err != nil {
		return err
	}
	if active, overridden := selectedProfileOverride(); overridden && active == name {
		return profileError(ProfileErrorConflict, name, fmt.Errorf("cannot delete active profile %q", name))
	}
	resolved, err := ResolveAppConfig()
	if err != nil {
		return err
	}
	if resolved.Global.Config.Profile == name || resolved.Effective.Profile == name {
		err := fmt.Errorf("cannot delete active profile %q", name)
		corelog.For("config").Error("active profile deletion rejected", "profile", name, "err", err)
		return profileError(ProfileErrorConflict, name, err)
	}
	if !profileConfigDirExists(name) {
		return profileError(ProfileErrorNotFound, name, fmt.Errorf("profile %q does not exist", name))
	}
	return nil
}

func SwitchProfile(name string) error {
	if err := ValidateProfileName(name); err != nil {
		return err
	}
	refreshRootPathsIfEnvironmentChanged()
	if err := ensureProfileDirs(name, true); err != nil {
		return err
	}
	if err := saveActiveProfile(name); err != nil {
		return err
	}
	clearSessionProfile()
	corelog.For("config").Info("current profile", "profile", name)
	return nil
}

// SwitchProfileForSession persists the global selection and keeps the chosen
// profile active for the remainder of an interactive process, above a local
// repository selection.
func SwitchProfileForSession(name string) error {
	if err := SwitchProfile(name); err != nil {
		return err
	}
	resolved, err := ResolveAppConfig()
	if err != nil {
		return err
	}
	if resolved.Local.Exists && strings.TrimSpace(resolved.Local.Config.Profile) != "" {
		setSessionProfile(name)
	}
	return nil
}

func RenameProfile(oldName, newName string) error {
	if err := ValidateProfileName(oldName); err != nil {
		return fmt.Errorf("old profile: %w", err)
	}
	if err := ValidateProfileName(newName); err != nil {
		return fmt.Errorf("new profile: %w", err)
	}
	if oldName == newName {
		if !profileConfigDirExists(oldName) {
			return profileError(ProfileErrorNotFound, oldName, fmt.Errorf("profile %q does not exist", oldName))
		}
		return nil
	}
	resolved, err := ResolveAppConfig()
	if err != nil {
		return err
	}
	if resolved.Local.Exists && strings.TrimSpace(resolved.Local.Config.Profile) == oldName {
		return profileError(ProfileErrorConflict, oldName, fmt.Errorf("profile %q is selected by local config %s; update that file before renaming the profile", oldName, resolved.Local.Path))
	}

	oldConfigDir := profileConfigDir(oldName)
	oldCacheDir := profileCacheDir(oldName)
	newConfigDir := profileConfigDir(newName)
	newCacheDir := profileCacheDir(newName)

	if !profileConfigDirExists(oldName) {
		return profileError(ProfileErrorNotFound, oldName, fmt.Errorf("profile %q does not exist", oldName))
	}
	if dirExists(newConfigDir) {
		return profileError(ProfileErrorConflict, newName, fmt.Errorf("profile %q already exists", newName))
	}

	if err := EnsurePrivateDir(GetConfigRootDirPath()); err != nil {
		return fmt.Errorf("create config root: %w", err)
	}
	if err := EnsurePrivateDir(GetCacheRootDirPath()); err != nil {
		return fmt.Errorf("create cache root: %w", err)
	}

	if dirExists(oldConfigDir) {
		if err := os.Rename(oldConfigDir, newConfigDir); err != nil {
			return fmt.Errorf("rename config profile: %w", err)
		}
	}
	if dirExists(oldCacheDir) && !dirExists(newCacheDir) {
		if err := os.Rename(oldCacheDir, newCacheDir); err != nil {
			return fmt.Errorf("rename cache profile: %w", err)
		}
	}

	active, err := loadActiveProfile()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if active == oldName || errors.Is(err, os.ErrNotExist) {
		if err := saveActiveProfile(newName); err != nil {
			return err
		}
		resetPaths()
	}
	return nil
}

func ValidateProfileName(name string) error {
	if err := validatePathSegment(name, "profile name"); err != nil {
		return profileError(ProfileErrorInvalidArgument, name, err)
	}
	return nil
}

func activeProfileOrDefault() string {
	if profile, overridden := selectedProfileOverride(); overridden {
		if err := ValidateProfileName(profile); err != nil {
			corelog.For("config").Error("selected profile invalid", "profile", profile, "err", err)
			return DefaultProfileName
		}
		return profile
	}
	if profile, err := loadActiveProfile(); err == nil && ValidateProfileName(profile) == nil {
		if profileConfigDirExists(profile) {
			return profile
		}
		corelog.For("config").Error("active profile missing", "profile", profile, "config_dir", profileConfigDir(profile))
		return profile
	}
	if err := ensureProfileDirs(DefaultProfileName, true); err != nil {
		corelog.For("config").Error("ensure default profile dirs failed", "err", err)
		return DefaultProfileName
	}
	if err := saveActiveProfile(DefaultProfileName); err != nil {
		corelog.For("config").Error("ensure default profile failed", "err", err)
	}
	corelog.For("config").Info("current profile", "profile", DefaultProfileName)
	return DefaultProfileName
}

func ensureProfileDirs(name string, createConfig bool) error {
	configAction := "secure"
	if createConfig {
		configAction = "create"
	}
	if err := EnsurePrivateDir(profileConfigDir(name)); err != nil {
		return fmt.Errorf("%s profile config dir: %w", configAction, err)
	}
	if err := EnsurePrivateDir(profileCacheDir(name)); err != nil {
		return fmt.Errorf("create profile cache dir: %w", err)
	}
	return nil
}

func loadActiveProfile() (string, error) {
	cfg, err := LoadAppConfig()
	if err != nil {
		return "", err
	}
	return cfg.Profile, nil
}

func saveActiveProfile(name string) error {
	cfg, err := LoadGlobalAppConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		cfg = &AppConfig{}
	}
	cfg.Profile = name
	return SaveAppConfig(cfg)
}

func profileConfigDir(name string) string {
	return filepath.Join(GetConfigRootDirPath(), profilesDirName, name)
}

func profileCacheDir(name string) string {
	return filepath.Join(GetCacheRootDirPath(), profilesDirName, name)
}

func profileConfigDirExists(name string) bool {
	entries, err := os.ReadDir(filepath.Join(GetConfigRootDirPath(), profilesDirName))
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.Name() == name {
			return dirExists(profileConfigDir(name))
		}
	}
	return false
}

// ProfileExists reports whether a profile config directory exists.
func ProfileExists(name string) bool {
	return profileConfigDirExists(name)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
