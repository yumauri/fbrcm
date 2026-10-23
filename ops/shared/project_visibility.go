package shared

import (
	"context"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/filter"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/ops/invocation"
)

func projectSelectionVisibilityContext(ctx context.Context) (context.Context, error) {
	if core.IncludesHiddenProjects(ctx) || !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return ctx, nil
	}
	return core.WithProjectVisibilitySnapshot(ctx)
}

// VisibleProjectSelection applies the profile policy before normal filtering.
// Explicit exact filters report hidden targets instead of silently skipping them.
func VisibleProjectSelection(cmd invocation.Call, projects []core.Project, rawFilters []string) ([]core.Project, error) {
	ctx, err := projectSelectionVisibilityContext(CommandContext(cmd))
	if err != nil {
		return nil, err
	}
	visible, err := core.VisibleProjects(ctx, projects)
	if err != nil || core.IncludesHiddenProjects(ctx) || !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return visible, err
	}
	for _, raw := range rawFilters {
		target, _, err := rctarget.ParseSelector(raw)
		if err != nil {
			return nil, InvalidArgument(err)
		}
		mode, _ := filter.ParseModePrefixedQuery(target.ProjectID)
		if mode != filter.ModeExact {
			continue
		}
		matches, err := FilterProjectTargets(projects, []string{raw})
		if err != nil {
			return nil, err
		}
		for _, project := range matches {
			if err := core.RequireProjectVisible(ctx, project.ProjectID); err != nil {
				return nil, err
			}
		}
	}
	return visible, nil
}

func visibleProjectArguments(cmd invocation.Call, projects []core.Project, query string, aliases map[string]string) ([]core.Project, error) {
	ctx, err := projectSelectionVisibilityContext(CommandContext(cmd))
	if err != nil {
		return nil, err
	}
	visible, err := core.VisibleProjects(ctx, projects)
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if project.ProjectID == query {
			if err := core.RequireProjectVisible(ctx, project.ProjectID); err != nil {
				return nil, err
			}
			return visible, nil
		}
	}
	if _, projectID, ok := config.ResolveProjectAlias(aliases, query); ok {
		if err := core.RequireProjectVisible(ctx, projectID); err != nil {
			return nil, err
		}
		return visible, nil
	}
	if len(matchProjectsForArg(visible, query)) > 0 {
		return visible, nil
	}
	for _, project := range matchProjectsForArg(projects, query) {
		if err := core.RequireProjectVisible(ctx, project.ProjectID); err != nil {
			return nil, err
		}
	}
	mode, value := filter.ParseModePrefixedQuery(query)
	if mode == filter.ModeExact {
		for _, project := range filterProjectsWithAliases(projects, []QueryFilter{{Mode: mode, Query: value}}, nil) {
			if err := core.RequireProjectVisible(ctx, project.ProjectID); err != nil {
				return nil, err
			}
		}
	}
	return visible, nil
}
