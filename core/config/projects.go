package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	corelog "github.com/yumauri/fbrcm/core/log"
)

// File holds file state used by the config package.
type File struct {
	// Projects stores projects for File.
	Projects []Project `json:"projects"`
	// SyncedAt stores synced at for File.
	SyncedAt string `json:"synced_at,omitempty"`
}

var ErrEmptyProjectsFile = errors.New("projects config is empty")

// Project holds project state used by the config package.
type Project struct {
	// Name stores name for Project.
	Name string `json:"name"`
	// ProjectID stores project id for Project.
	ProjectID string `json:"project_id"`
	// ProjectNumber stores project number for Project.
	ProjectNumber string `json:"project_number,omitempty"`
	// State stores state for Project.
	State string `json:"state,omitempty"`
	// ETag stores etag for Project.
	ETag string `json:"etag,omitempty"`
	// UpdatedAt stores updated at for Project.
	UpdatedAt string `json:"updated_at,omitempty"`
	// SyncedAt stores synced at for Project.
	SyncedAt string `json:"synced_at,omitempty"`
}

// Load list of projects from the projects file
func LoadProjects() ([]Project, error) {
	path := GetProjectsFilePath()
	logger := corelog.For("config")
	logger.Debug("read projects config", "path", path)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("projects config cache miss", "path", path)
			return nil, fmt.Errorf("read projects config: %w", err)
		}
		logger.Error("read projects config failed", "path", path, "err", err)
		return nil, fmt.Errorf("read projects config: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		logger.Warn("projects config empty", "path", path)
		return nil, fmt.Errorf("read projects config: %w", ErrEmptyProjectsFile)
	}

	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		logger.Error("decode projects config failed", "path", path, "err", err)
		return nil, fmt.Errorf("decode projects config: %w", err)
	}

	sortProjects(file.Projects)
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
		Projects: append([]Project(nil), projects...),
		SyncedAt: updatedAt.UTC().Format(time.RFC3339),
	}
	sortProjects(file.Projects)

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("encode projects config: %w", err)
	}
	data = append(data, '\n')

	path := GetProjectsFilePath()
	logger.Debug("write projects config", "path", path, "count", len(file.Projects), "synced_at", file.SyncedAt)
	if err := os.WriteFile(path, data, PrivateFileMode); err != nil {
		logger.Error("write projects config failed", "path", path, "err", err)
		return fmt.Errorf("write projects config: %w", err)
	}
	if err := EnsurePrivateFile(path); err != nil {
		return fmt.Errorf("chmod projects config: %w", err)
	}

	logger.Info("saved projects config", "path", path, "count", len(file.Projects), "synced_at", file.SyncedAt)
	return nil
}

// Delete saved projects config file if it exists.
func PurgeProjects() error {
	path := GetProjectsFilePath()
	logger := corelog.For("config")
	logger.Debug("remove projects config", "path", path)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Error("remove projects config failed", "path", path, "err", err)
		return fmt.Errorf("remove projects config: %w", err)
	}

	logger.Info("projects config removed", "path", path)
	return nil
}

// Helper to sort projects
func sortProjects(projects []Project) {
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Name == projects[j].Name {
			return projects[i].ProjectID < projects[j].ProjectID
		}
		return projects[i].Name < projects[j].Name
	})
}
