package shared

import (
	"context"
	"fmt"
	"strings"

	"fbrcm/core"
	"fbrcm/core/filter"
	"fbrcm/core/firebase"
	corelog "fbrcm/core/log"
)

func FilterProjectsByExpr(ctx context.Context, svc *core.Core, projects []core.Project, rawExpr string) ([]core.Project, error) {
	rawExpr = strings.TrimSpace(rawExpr)
	if rawExpr == "" {
		return projects, nil
	}

	compiled, ok := CompileExpr(rawExpr, "")
	if !ok {
		return nil, nil
	}

	filtered := make([]core.Project, 0, len(projects))
	for _, project := range projects {
		cfg, err := loadProjectExprConfig(ctx, svc, project)
		if err != nil {
			corelog.For("filter").Error("project expression context load failed; skipping project", "project_id", project.ProjectID, "expr", rawExpr, "err", err)
			continue
		}

		match, ok := MatchProjectByCompiledExpr(compiled, project, cfg)
		if ok && match {
			filtered = append(filtered, project)
		}
	}

	return filtered, nil
}

func MatchProjectByExpr(project core.Project, cfg *firebase.RemoteConfig, rawExpr string) bool {
	rawExpr = strings.TrimSpace(rawExpr)
	if rawExpr == "" {
		return true
	}

	compiled, ok := CompileExpr(rawExpr, project.ProjectID)
	if !ok {
		return false
	}

	match, ok := MatchProjectByCompiledExpr(compiled, project, cfg)
	return ok && match
}

func CompileExpr(rawExpr, projectID string) (*filter.Expression, bool) {
	rawExpr = strings.TrimSpace(rawExpr)
	if rawExpr == "" {
		return nil, true
	}

	compiled, err := filter.CompileExpression(rawExpr)
	if err != nil {
		logger := corelog.For("filter")
		if projectID == "" {
			logger.Error("expression compile failed", "expr", rawExpr, "err", err)
		} else {
			logger.Error("expression compile failed", "project_id", projectID, "expr", rawExpr, "err", err)
		}
		return nil, false
	}

	return compiled, true
}

func MatchProjectByCompiledExpr(compiled *filter.Expression, project core.Project, cfg *firebase.RemoteConfig) (bool, bool) {
	if compiled == nil {
		return true, true
	}

	match, err := compiled.MatchProject(project.ProjectID, project.Name, cfg)
	if err != nil {
		corelog.For("filter").Error("project expression evaluation failed; skipping project", "project_id", project.ProjectID, "err", err)
		return false, false
	}

	return match, true
}

func MatchParameterByCompiledExpr(compiled *filter.Expression, project core.Project, cfg *firebase.RemoteConfig, name, group string) (bool, bool) {
	if compiled == nil {
		return true, true
	}

	match, err := compiled.MatchParameter(project.ProjectID, project.Name, cfg, name, group)
	if err != nil {
		corelog.For("filter").Error("parameter expression evaluation failed; skipping parameter", "project_id", project.ProjectID, "name", name, "group", group, "err", err)
		return false, false
	}

	return match, true
}

func loadProjectExprConfig(ctx context.Context, svc *core.Core, project core.Project) (*firebase.RemoteConfig, error) {
	cache, _, err := svc.GetParameters(ctx, project.ProjectID, false)
	if err != nil {
		return nil, err
	}

	cfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, fmt.Errorf("decode remote config: %w", err)
	}
	return cfg, nil
}
