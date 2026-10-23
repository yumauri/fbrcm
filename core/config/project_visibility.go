package config

import (
	"fmt"
	"path/filepath"
	"slices"

	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/core/strfold"
)

// ProjectVisibility is profile-owned preference data, independent of the
// rebuildable discovery registry and of repository aliases.
type ProjectVisibility struct {
	Version   int                   `json:"version"`
	Blacklist ProjectVisibilityList `json:"blacklist"`
	Whitelist ProjectVisibilityList `json:"whitelist"`
}

type ProjectVisibilityList struct {
	Enabled    bool     `json:"enabled"`
	ProjectIDs []string `json:"project_ids"`
}

func DefaultProjectVisibility() ProjectVisibility {
	return ProjectVisibility{
		Version:   1,
		Blacklist: ProjectVisibilityList{Enabled: true, ProjectIDs: []string{}},
		Whitelist: ProjectVisibilityList{ProjectIDs: []string{}},
	}
}

func GetProjectVisibilityFilePath() string {
	return filepath.Join(GetConfigDirPath(), "project-visibility.json")
}

// HiddenReason accepts either a physical project ID or a template target.
// Blacklist membership takes precedence over whitelist exclusion.
func (p ProjectVisibility) HiddenReason(projectID string) string {
	if target, err := rctarget.Parse(projectID); err == nil {
		projectID = target.ProjectID
	}
	if p.Blacklist.Enabled && slices.Contains(p.Blacklist.ProjectIDs, projectID) {
		return "blacklisted"
	}
	if p.Whitelist.Enabled && !slices.Contains(p.Whitelist.ProjectIDs, projectID) {
		return "not_whitelisted"
	}
	return ""
}

func (p *ProjectVisibility) normalize() error {
	if p.Version != 1 {
		return fmt.Errorf("unsupported project visibility version %d", p.Version)
	}
	for _, list := range []*ProjectVisibilityList{&p.Blacklist, &p.Whitelist} {
		for _, id := range list.ProjectIDs {
			if err := ValidatePhysicalProjectID(id); err != nil {
				return err
			}
		}
		list.ProjectIDs = append([]string{}, list.ProjectIDs...)
		strfold.Sort(list.ProjectIDs)
		list.ProjectIDs = slices.Compact(list.ProjectIDs)
	}
	return nil
}

func LoadProjectVisibility() (ProjectVisibility, error) {
	path := GetProjectVisibilityFilePath()
	var policy ProjectVisibility
	if err := readJSONFile(path, &policy); err != nil {
		if isNotExist(err) {
			return DefaultProjectVisibility(), nil
		}
		if isDecodeError(err) {
			return policy, invalidConfiguration(path, "decoding", err)
		}
		return policy, fmt.Errorf("read project visibility: %w", err)
	}
	if err := policy.normalize(); err != nil {
		return policy, invalidConfiguration(path, "validation", err)
	}
	return policy, nil
}

func SaveProjectVisibility(policy ProjectVisibility) error {
	if err := policy.normalize(); err != nil {
		return invalidConfiguration(GetProjectVisibilityFilePath(), "validation", err)
	}
	if err := EnsurePrivateDir(GetConfigDirPath()); err != nil {
		return err
	}
	return writeJSONFile(GetProjectVisibilityFilePath(), policy)
}
