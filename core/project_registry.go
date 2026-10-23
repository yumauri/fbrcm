package core

import (
	"context"
	"fmt"
	"slices"

	"github.com/yumauri/fbrcm/core/config"
)

type projectRegistrySnapshotKey struct{}

type projectRegistrySnapshot struct {
	path     string
	projects []Project
}

// WithProjectRegistrySnapshot shares already loaded physical project records
// with Firebase client resolution during one operation. Later operations must
// load a fresh registry, and registry mutations continue to read from disk.
func WithProjectRegistrySnapshot(ctx context.Context, projects []Project) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if !ExecutionPolicyFromContext(ctx).ReadLocalState {
		return ctx
	}
	snapshot := projectRegistrySnapshot{
		path:     config.GetProjectsFilePath(),
		projects: make([]Project, len(projects)),
	}
	for i, project := range projects {
		snapshot.projects[i] = cloneProjectRecord(project)
	}
	return context.WithValue(ctx, projectRegistrySnapshotKey{}, snapshot)
}

func (s *Core) projectByIDForContext(ctx context.Context, projectID string) (Project, error) {
	if ctx != nil && ExecutionPolicyFromContext(ctx).ReadLocalState {
		if snapshot, ok := ctx.Value(projectRegistrySnapshotKey{}).(projectRegistrySnapshot); ok && snapshot.path == config.GetProjectsFilePath() {
			return findProjectByID(snapshot.projects, projectID)
		}
	}
	return s.ProjectByID(projectID)
}

func findProjectByID(projects []Project, projectID string) (Project, error) {
	for _, project := range projects {
		if project.ProjectID == projectID {
			return cloneProjectRecord(project), nil
		}
	}
	return Project{}, fmt.Errorf("project %q is not in projects config", projectID)
}

func cloneProjectRecord(project Project) Project {
	project.Templates = slices.Clone(project.Templates)
	project.DiscoveredBy = slices.Clone(project.DiscoveredBy)
	return project
}
