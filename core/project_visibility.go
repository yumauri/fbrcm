package core

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core/config"
)

type includeHiddenProjectsKey struct{}

func WithHiddenProjects(ctx context.Context, include bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, includeHiddenProjectsKey{}, include)
}

func IncludesHiddenProjects(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	include, _ := ctx.Value(includeHiddenProjectsKey{}).(bool)
	return include
}

// ProjectHiddenError distinguishes a policy exclusion from a missing project.
type ProjectHiddenError struct {
	ProjectID string
	Reason    string
}

func (e *ProjectHiddenError) Error() string {
	return fmt.Sprintf("project %q is hidden (%s); use --include-hidden or change projects blacklist/whitelist", e.ProjectID, ProjectVisibilityReasonText(e.Reason))
}

// ProjectVisibilityForContext avoids all profile reads during stateless runs.
func ProjectVisibilityForContext(ctx context.Context) (config.ProjectVisibility, error) {
	if !ExecutionPolicyFromContext(ctx).ReadLocalState {
		return config.DefaultProjectVisibility(), nil
	}
	return config.LoadProjectVisibility()
}

// VisibleProjects applies policy before selectors, template expansion, or
// template reads. Inventory and authentication callers retain the raw registry.
func VisibleProjects(ctx context.Context, projects []Project) ([]Project, error) {
	if IncludesHiddenProjects(ctx) || !ExecutionPolicyFromContext(ctx).ReadLocalState {
		return projects, nil
	}
	policy, err := ProjectVisibilityForContext(ctx)
	if err != nil {
		return nil, err
	}
	visible := make([]Project, 0, len(projects))
	for _, project := range projects {
		if policy.HiddenReason(project.ProjectID) == "" {
			visible = append(visible, project)
		}
	}
	return visible, nil
}

func RequireProjectVisible(ctx context.Context, projectID string) error {
	if IncludesHiddenProjects(ctx) || !ExecutionPolicyFromContext(ctx).ReadLocalState {
		return nil
	}
	policy, err := ProjectVisibilityForContext(ctx)
	if err != nil {
		return err
	}
	if reason := policy.HiddenReason(projectID); reason != "" {
		return &ProjectHiddenError{ProjectID: projectID, Reason: reason}
	}
	return nil
}

// ProjectVisibilityReasonText provides consistent human labels for exclusions.
func ProjectVisibilityReasonText(reason string) string {
	if reason == "not_whitelisted" {
		return "not whitelisted"
	}
	return reason
}
