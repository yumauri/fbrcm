package config

import (
	"fmt"
	"path/filepath"
	"slices"

	corelog "github.com/yumauri/fbrcm/core/log"
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
	logger := corelog.For("config").With("path", path)
	logger.Debug("read project visibility config")
	var policy ProjectVisibility
	if err := readJSONFile(path, &policy); err != nil {
		if isNotExist(err) {
			logger.Debug("project visibility config missing; using defaults")
			return DefaultProjectVisibility(), nil
		}
		if isDecodeError(err) {
			logger.Error("decode project visibility config failed", "err", err)
			return policy, invalidConfiguration(path, "decoding", err)
		}
		logger.Error("read project visibility config failed", "err", err)
		return policy, fmt.Errorf("read project visibility: %w", err)
	}
	if err := policy.normalize(); err != nil {
		logger.Error("validate project visibility config failed", "err", err)
		return policy, invalidConfiguration(path, "validation", err)
	}
	logger.Info("loaded project visibility config",
		"blacklist_enabled", policy.Blacklist.Enabled, "blacklist_count", len(policy.Blacklist.ProjectIDs),
		"whitelist_enabled", policy.Whitelist.Enabled, "whitelist_count", len(policy.Whitelist.ProjectIDs))
	return policy, nil
}

func SaveProjectVisibility(policy ProjectVisibility) error {
	path := GetProjectVisibilityFilePath()
	logger := corelog.For("config").With("path", path)
	if err := policy.normalize(); err != nil {
		logger.Error("validate project visibility config failed", "err", err)
		return invalidConfiguration(path, "validation", err)
	}
	if err := EnsurePrivateDir(GetConfigDirPath()); err != nil {
		logger.Error("prepare project visibility config directory failed", "err", err)
		return err
	}
	logger = logger.With(
		"blacklist_enabled", policy.Blacklist.Enabled, "blacklist_count", len(policy.Blacklist.ProjectIDs),
		"whitelist_enabled", policy.Whitelist.Enabled, "whitelist_count", len(policy.Whitelist.ProjectIDs))
	logger.Debug("write project visibility config")
	if err := writeJSONFile(path, policy); err != nil {
		logger.Error("write project visibility config failed", "err", err)
		return err
	}
	logger.Info("saved project visibility config")
	return nil
}
