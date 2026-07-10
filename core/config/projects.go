package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
)

type File struct {
	// Version stores projects config version.
	Version  int       `json:"version"`
	Projects []Project `json:"projects"`
	SyncedAt string    `json:"synced_at,omitempty"`
}

var ErrEmptyProjectsFile = errors.New("projects config is empty")

const ProjectsConfigVersion = 2

type Project struct {
	Name          string `json:"name"`
	ProjectID     string `json:"project_id"`
	ProjectNumber string `json:"project_number,omitempty"`
	State         string `json:"state,omitempty"`
	ETag          string `json:"etag,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
	SyncedAt      string `json:"synced_at,omitempty"`
	AuthID        string `json:"auth_id"`
	// DiscoveredBy stores auth identities that discovered Project.
	DiscoveredBy []string `json:"discovered_by,omitempty"`
}

// Load list of projects from the projects file
func LoadProjects() ([]Project, error) {
	path := GetProjectsFilePath()
	logger := corelog.For("config")
	logger.Debug("read projects config", "path", path)

	data, err := readFileBytes(path)
	if err != nil {
		if isNotExist(err) {
			logger.Warn("projects config cache miss", "path", path)
		} else {
			logger.Error("read projects config failed", "path", path, "err", err)
		}
		return nil, fmt.Errorf("read projects config: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		logger.Warn("projects config empty", "path", path)
		return nil, fmt.Errorf("read projects config: %w", ErrEmptyProjectsFile)
	}

	var file File
	if err := decodeJSON(data, &file); err != nil {
		logger.Error("decode projects config failed", "path", path, "err", err)
		return nil, fmt.Errorf("decode projects config: %w", err)
	}
	if file.Version != ProjectsConfigVersion {
		return nil, fmt.Errorf("unsupported projects config version %d", file.Version)
	}
	for _, project := range file.Projects {
		if strings.TrimSpace(project.ProjectID) == "" {
			return nil, fmt.Errorf("projects config contains project without project_id")
		}
		if strings.TrimSpace(project.AuthID) == "" {
			return nil, fmt.Errorf("project %s missing auth_id", project.ProjectID)
		}
	}

	strfold.SortProjects(file.Projects, func(p Project) string { return p.Name }, func(p Project) string { return p.ProjectID })
	logger.Info("loaded projects config", "path", path, "count", len(file.Projects), "synced_at", file.SyncedAt)
	return file.Projects, nil
}

// Save list of projects to the projects file
func SaveProjects(projects []Project, updatedAt time.Time) error {
	logger := corelog.For("config")
	if err := EnsurePrivateDir(GetConfigDirPath()); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	file := File{
		Version:  ProjectsConfigVersion,
		Projects: append([]Project(nil), projects...),
		SyncedAt: updatedAt.UTC().Format(time.RFC3339),
	}
	for i := range file.Projects {
		strfold.Sort(file.Projects[i].DiscoveredBy)
	}
	strfold.SortProjects(file.Projects, func(p Project) string { return p.Name }, func(p Project) string { return p.ProjectID })

	path := GetProjectsFilePath()
	logger.Debug("write projects config", "path", path, "count", len(file.Projects), "synced_at", file.SyncedAt)
	if err := writeJSONFile(path, file); err != nil {
		if isEncodeError(err) {
			return fmt.Errorf("encode projects config: %w", err)
		}
		logger.Error("write projects config failed", "path", path, "err", err)
		return fmt.Errorf("write projects config: %w", err)
	}

	logger.Info("saved projects config", "path", path, "count", len(file.Projects), "synced_at", file.SyncedAt)
	return nil
}

// Delete saved projects config file if it exists.
func PurgeProjects() error {
	path := GetProjectsFilePath()
	logger := corelog.For("config")
	logger.Debug("remove projects config", "path", path)
	if err := os.Remove(path); err != nil && !isNotExist(err) {
		logger.Error("remove projects config failed", "path", path, "err", err)
		return fmt.Errorf("remove projects config: %w", err)
	}

	logger.Info("projects config removed", "path", path)
	return nil
}
