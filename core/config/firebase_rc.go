package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/tailscale/hujson"
)

const (
	FirebaseRCFileName     = ".firebaserc"
	FirebaseConfigFileName = "firebase.json"
)

// ProjectAliasSource identifies the repository file that defines an alias.
type ProjectAliasSource string

const (
	ProjectAliasSourceFBRCM    ProjectAliasSource = "fbrcm"
	ProjectAliasSourceFirebase ProjectAliasSource = "firebase"
	ProjectAliasSourceBoth     ProjectAliasSource = "both"
)

// ProjectAliasEntry is one effective repository alias with source metadata.
type ProjectAliasEntry struct {
	Alias     string             `json:"alias"`
	ProjectID string             `json:"project_id"`
	Source    ProjectAliasSource `json:"source"`
}

// ProjectAliasRegistry contains effective aliases and their repository sources.
type ProjectAliasRegistry struct {
	Aliases      map[string]string
	Entries      map[string]ProjectAliasEntry
	FBRCMPath    string
	FirebasePath string
}

// SortedEntries returns effective aliases ordered by alias name.
func (r ProjectAliasRegistry) SortedEntries() []ProjectAliasEntry {
	entries := make([]ProjectAliasEntry, 0, len(r.Entries))
	for _, entry := range r.Entries {
		entries = append(entries, entry)
	}
	slices.SortFunc(entries, func(left, right ProjectAliasEntry) int {
		if left.Alias < right.Alias {
			return -1
		}
		if left.Alias > right.Alias {
			return 1
		}
		return 0
	})
	return entries
}

// FindFirebaseProjectRoot searches for the nearest ancestor containing
// firebase.json, matching Firebase CLI project-root discovery.
func FindFirebaseProjectRoot(startDir string) (string, bool, error) {
	path, found, err := findAncestorFile(startDir, FirebaseConfigFileName)
	if err != nil || !found {
		return "", found, err
	}
	return filepath.Dir(path), true, nil
}

// GetFirebaseRCFilePath resolves the .firebaserc used by Firebase CLI. It is
// beside the nearest firebase.json, or in the current directory when no
// Firebase project root exists.
func GetFirebaseRCFilePath() (string, bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false, fmt.Errorf("resolve current working directory: %w", err)
	}
	base := cwd
	if root, found, findErr := FindFirebaseProjectRoot(cwd); findErr != nil {
		return "", false, findErr
	} else if found {
		base = root
	}
	path := filepath.Join(base, FirebaseRCFileName)
	info, err := os.Stat(path)
	switch {
	case err == nil:
		if !info.Mode().IsRegular() {
			return "", false, fmt.Errorf("firebase RC is not a regular file: %s", path)
		}
		return path, true, nil
	case errors.Is(err, os.ErrNotExist):
		return path, false, nil
	default:
		return "", false, fmt.Errorf("inspect Firebase RC %s: %w", path, err)
	}
}

// LoadFirebaseProjectAliasesFile reads the top-level projects map from a
// Firebase CLI RC file. Firebase-compatible comments and trailing commas are
// accepted; unrelated RC fields are ignored.
func LoadFirebaseProjectAliasesFile(path string) (map[string]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve Firebase RC path %s: %w", path, err)
	}
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read Firebase RC %s: %w", absPath, err)
	}
	standard, err := hujson.Standardize(raw)
	if err != nil {
		return nil, fmt.Errorf("decode Firebase RC %s: %w", absPath, err)
	}
	var rc struct {
		Projects map[string]string `json:"projects"`
	}
	if err := json.Unmarshal(standard, &rc); err != nil {
		return nil, fmt.Errorf("decode Firebase RC %s: %w", absPath, err)
	}
	if rc.Projects == nil {
		return map[string]string{}, nil
	}
	if err := ValidateProjectAliases(rc.Projects); err != nil {
		return nil, fmt.Errorf("validate Firebase RC %s: %w", absPath, err)
	}
	return maps.Clone(rc.Projects), nil
}

// LoadProjectAliasRegistry combines native fbrcm aliases with Firebase CLI
// aliases. Identical definitions are shared; conflicting definitions fail.
func LoadProjectAliasRegistry() (ProjectAliasRegistry, error) {
	registry := ProjectAliasRegistry{
		Aliases: make(map[string]string),
		Entries: make(map[string]ProjectAliasEntry),
	}
	if LocalConfigDisabled() {
		return registry, nil
	}

	resolved, err := ResolveAppConfig()
	if err != nil {
		return ProjectAliasRegistry{}, err
	}
	registry.FBRCMPath = resolved.Local.Path
	native := CloneProjectAliases(resolved.Local.Config)
	for alias, projectID := range native {
		registry.Aliases[alias] = projectID
		registry.Entries[alias] = ProjectAliasEntry{Alias: alias, ProjectID: projectID, Source: ProjectAliasSourceFBRCM}
	}

	firebasePath, exists, err := GetFirebaseRCFilePath()
	if err != nil {
		return ProjectAliasRegistry{}, err
	}
	registry.FirebasePath = firebasePath
	if !exists {
		return registry, nil
	}
	firebaseAliases, err := LoadFirebaseProjectAliasesFile(firebasePath)
	if err != nil {
		return ProjectAliasRegistry{}, err
	}
	for alias, projectID := range firebaseAliases {
		if nativeProjectID, ok := registry.Aliases[alias]; ok {
			if nativeProjectID != projectID {
				return ProjectAliasRegistry{}, fmt.Errorf(
					"project alias %q conflicts: %s maps it to %q, while %s maps it to %q",
					alias,
					registry.FBRCMPath,
					nativeProjectID,
					firebasePath,
					projectID,
				)
			}
			entry := registry.Entries[alias]
			entry.Source = ProjectAliasSourceBoth
			registry.Entries[alias] = entry
			continue
		}
		registry.Aliases[alias] = projectID
		registry.Entries[alias] = ProjectAliasEntry{Alias: alias, ProjectID: projectID, Source: ProjectAliasSourceFirebase}
	}
	return registry, nil
}
