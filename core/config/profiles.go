package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	corelog "github.com/yumauri/fbrcm/core/log"
)

const DefaultProfileName = "default"

func GetActiveProfileName() string {
	return getPaths().profile
}

func EnsureActiveProfile() error {
	profile, err := loadActiveProfile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SwitchProfile(DefaultProfileName)
		}
		return err
	}
	if strings.TrimSpace(profile) == "" {
		return SwitchProfile(DefaultProfileName)
	}
	if err := ValidateProfileName(profile); err != nil {
		return err
	}
	if !profileConfigDirExists(profile) {
		err := fmt.Errorf("active profile %q does not exist in config directory", profile)
		corelog.For("config").Error("active profile missing", "profile", profile, "config_dir", profileConfigDir(profile), "err", err)
		return err
	}
	if err := ensureExistingProfileDirs(profile); err != nil {
		return err
	}
	corelog.For("config").Info("current profile", "profile", profile)
	return nil
}

func ListProfiles() ([]string, error) {
	seen := map[string]struct{}{}
	root := GetConfigRootDirPath()
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
	sort.Strings(profiles)
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

func PurgeProfile(name string) error {
	if err := EnsureProfileCanPurge(name); err != nil {
		return err
	}
	if err := os.RemoveAll(profileConfigDir(name)); err != nil {
		return fmt.Errorf("remove profile config dir: %w", err)
	}
	if err := os.RemoveAll(profileCacheDir(name)); err != nil {
		return fmt.Errorf("remove profile cache dir: %w", err)
	}
	corelog.For("config").Info("profile purged", "profile", name, "config_dir", profileConfigDir(name), "cache_dir", profileCacheDir(name))
	return nil
}

// EnsureProfileCanPurge validates profile purge safety and returns the resulting value or error.
func EnsureProfileCanPurge(name string) error {
	if err := ValidateProfileName(name); err != nil {
		return err
	}
	active, err := loadActiveProfile()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if active == name {
		err := fmt.Errorf("cannot purge active profile %q", name)
		corelog.For("config").Error("active profile purge rejected", "profile", name, "err", err)
		return err
	}
	if !profileConfigDirExists(name) {
		return fmt.Errorf("profile %q does not exist", name)
	}
	return nil
}

func SwitchProfile(name string) error {
	if err := ValidateProfileName(name); err != nil {
		return err
	}
	if err := ensureProfileDirs(name); err != nil {
		return err
	}
	if err := saveActiveProfile(name); err != nil {
		return err
	}
	resetPaths()
	corelog.For("config").Info("current profile", "profile", name)
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
		return nil
	}

	oldConfigDir := filepath.Join(GetConfigRootDirPath(), oldName)
	oldCacheDir := filepath.Join(GetCacheRootDirPath(), oldName)
	newConfigDir := filepath.Join(GetConfigRootDirPath(), newName)
	newCacheDir := filepath.Join(GetCacheRootDirPath(), newName)

	if !dirExists(oldConfigDir) {
		return fmt.Errorf("profile %q does not exist", oldName)
	}
	if dirExists(newConfigDir) {
		return fmt.Errorf("profile %q already exists", newName)
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
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if trimmed != name {
		return fmt.Errorf("profile name cannot have leading or trailing whitespace")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("profile name %q is reserved", name)
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("profile name cannot contain path separators")
	}
	if filepath.Clean(name) != name {
		return fmt.Errorf("profile name must be a single path segment")
	}
	return nil
}

func activeProfileOrDefault() string {
	if profile, err := loadActiveProfile(); err == nil && ValidateProfileName(profile) == nil {
		if profileConfigDirExists(profile) {
			return profile
		}
		corelog.For("config").Error("active profile missing", "profile", profile, "config_dir", profileConfigDir(profile))
		return profile
	}
	if err := ensureProfileDirs(DefaultProfileName); err != nil {
		corelog.For("config").Error("ensure default profile dirs failed", "err", err)
		return DefaultProfileName
	}
	if err := saveActiveProfile(DefaultProfileName); err != nil {
		corelog.For("config").Error("ensure default profile failed", "err", err)
	}
	corelog.For("config").Info("current profile", "profile", DefaultProfileName)
	return DefaultProfileName
}

func ensureProfileDirs(name string) error {
	if err := EnsurePrivateDir(profileConfigDir(name)); err != nil {
		return fmt.Errorf("create profile config dir: %w", err)
	}
	if err := EnsurePrivateDir(profileCacheDir(name)); err != nil {
		return fmt.Errorf("create profile cache dir: %w", err)
	}
	return nil
}

func ensureExistingProfileDirs(name string) error {
	if err := EnsurePrivateDir(profileConfigDir(name)); err != nil {
		return fmt.Errorf("secure profile config dir: %w", err)
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
	cfg, err := LoadAppConfig()
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
	return filepath.Join(GetConfigRootDirPath(), name)
}

func profileCacheDir(name string) string {
	return filepath.Join(GetCacheRootDirPath(), name)
}

func profileConfigDirExists(name string) bool {
	return dirExists(profileConfigDir(name))
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
