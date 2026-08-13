package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	corelog "github.com/yumauri/fbrcm/core/log"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
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
	Name            string          `json:"name"`
	ProjectID       string          `json:"project_id"`
	ProjectNumber   string          `json:"project_number,omitempty"`
	State           string          `json:"state,omitempty"`
	ETag            string          `json:"etag,omitempty"`
	UpdatedAt       string          `json:"updated_at,omitempty"`
	SyncedAt        string          `json:"synced_at,omitempty"`
	AuthID          string          `json:"auth_id"`
	Disabled        bool            `json:"disabled,omitempty"`
	Templates       []rctarget.Kind `json:"templates,omitempty"`
	PrimaryTemplate rctarget.Kind   `json:"primary_template,omitempty"`
	// DiscoveredBy stores auth identities that discovered Project.
	DiscoveredBy []string `json:"discovered_by,omitempty"`
}

// NormalizeTemplatePreferences fills backward-compatible client defaults and
// validates a project's persisted template views.
func (p *Project) NormalizeTemplatePreferences() error {
	if len(p.Templates) == 0 {
		p.Templates = []rctarget.Kind{rctarget.Client}
	}
	seen := make(map[rctarget.Kind]struct{}, len(p.Templates))
	for _, kind := range p.Templates {
		if kind != rctarget.Client && kind != rctarget.Server {
			return fmt.Errorf("project %s has unsupported template %q", p.ProjectID, kind)
		}
		seen[kind] = struct{}{}
	}
	templates := make([]rctarget.Kind, 0, len(seen))
	for _, kind := range []rctarget.Kind{rctarget.Client, rctarget.Server} {
		if _, ok := seen[kind]; ok {
			templates = append(templates, kind)
		}
	}
	p.Templates = templates
	if p.PrimaryTemplate == "" {
		p.PrimaryTemplate = templates[0]
		if _, ok := seen[rctarget.Client]; ok {
			p.PrimaryTemplate = rctarget.Client
		}
	}
	if _, ok := seen[p.PrimaryTemplate]; !ok {
		return fmt.Errorf("project %s primary template %q is not enabled", p.ProjectID, p.PrimaryTemplate)
	}
	return nil
}

// TemplateKinds returns enabled template kinds with the primary template first.
func (p Project) TemplateKinds() []rctarget.Kind {
	if err := p.NormalizeTemplatePreferences(); err != nil {
		return []rctarget.Kind{rctarget.Client}
	}
	out := make([]rctarget.Kind, 0, len(p.Templates))
	out = append(out, p.PrimaryTemplate)
	for _, kind := range p.Templates {
		if kind != p.PrimaryTemplate {
			out = append(out, kind)
		}
	}
	return out
}

// TemplateTarget returns a copy identified by its canonical template target.
func (p Project) TemplateTarget(kind rctarget.Kind) Project {
	p.ProjectID = (rctarget.Target{Kind: kind, ProjectID: p.ProjectID}).String()
	return p
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
		return nil, invalidConfiguration(path, "decoding", fmt.Errorf("decode projects config: %w", err))
	}
	if file.Version != ProjectsConfigVersion {
		return nil, invalidConfiguration(path, "validation", fmt.Errorf("unsupported projects config version %d", file.Version))
	}
	for i := range file.Projects {
		project := &file.Projects[i]
		if strings.TrimSpace(project.ProjectID) == "" {
			return nil, invalidConfiguration(path, "validation", fmt.Errorf("projects config contains project without project_id"))
		}
		if strings.TrimSpace(project.AuthID) == "" {
			return nil, invalidConfiguration(path, "validation", fmt.Errorf("project %s missing auth_id", project.ProjectID))
		}
		if err := project.NormalizeTemplatePreferences(); err != nil {
			return nil, invalidConfiguration(path, "validation", err)
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
		if err := file.Projects[i].NormalizeTemplatePreferences(); err != nil {
			return err
		}
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

// ResetProjects deletes the saved projects registry file and reports whether
// one existed.
func ResetProjects() (bool, error) {
	path := GetProjectsFilePath()
	logger := corelog.For("config")
	logger.Debug("remove projects config", "path", path)
	if err := os.Remove(path); err != nil {
		if isNotExist(err) {
			logger.Info("projects config already absent", "path", path)
			return false, nil
		}
		logger.Error("remove projects config failed", "path", path, "err", err)
		return false, fmt.Errorf("remove projects config: %w", err)
	}

	logger.Info("projects config removed", "path", path)
	return true, nil
}
