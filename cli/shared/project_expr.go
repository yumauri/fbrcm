package shared

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// QueryFilter holds a parsed mode-prefixed query.
type QueryFilter struct {
	Mode  filter.Mode
	Query string
}

// ParseFilters parses mode-prefixed queries and drops empty queries.
func ParseFilters(rawFilters []string) []QueryFilter {
	filters := make([]QueryFilter, 0, len(rawFilters))
	for _, raw := range rawFilters {
		mode, query := ParseFilter(raw)
		if query == "" {
			continue
		}
		filters = append(filters, QueryFilter{Mode: mode, Query: query})
	}
	return filters
}

// MatchAnyFilter reports whether value matches any filter. Empty filters match all.
func MatchAnyFilter(value string, filters []QueryFilter) bool {
	if len(filters) == 0 {
		return true
	}
	for _, item := range filters {
		match, _ := filter.Match(value, item.Query, item.Mode)
		if match {
			return true
		}
	}
	return false
}

// HighlightFilters returns merged highlight indices for every matching filter.
func HighlightFilters(value string, filters []QueryFilter) []int {
	highlightSet := make(map[int]struct{})
	for _, item := range filters {
		match, highlights := filter.Match(value, item.Query, item.Mode)
		if !match {
			continue
		}
		for _, index := range highlights {
			highlightSet[index] = struct{}{}
		}
	}
	if len(highlightSet) == 0 {
		return nil
	}

	indices := make([]int, 0, len(highlightSet))
	for index := range highlightSet {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	return indices
}

// FilterProjects filters projects by mode-prefixed queries. Multiple queries are ORed.
func FilterProjects(projects []core.Project, rawFilters []string) []core.Project {
	filters := ParseFilters(rawFilters)
	if len(filters) == 0 {
		return projects
	}

	filtered := make([]core.Project, 0, len(projects))
	for _, project := range projects {
		for _, item := range filters {
			nameMatch, _ := filter.Match(project.Name, item.Query, item.Mode)
			idMatch, _ := filter.Match(project.ProjectID, item.Query, item.Mode)
			if nameMatch || idMatch {
				filtered = append(filtered, project)
				break
			}
		}
	}
	return filtered
}

// ParseFilter parses a mode-prefixed query.
func ParseFilter(raw string) (filter.Mode, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return filter.ModeFuzzy, ""
	}

	mode, ok := filter.ModeFromLabel(string([]rune(raw)[0]))
	if !ok {
		return filter.ModeFuzzy, raw
	}
	return mode, string([]rune(raw)[1:])
}

// FilterProjectsByExpr filters projects by expr and returns the resulting value or error.
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

// MatchProjectByExpr matches project by expr and returns the resulting value or error.
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

// MatchProjectByCompiledExpr matches project by compiled expr and returns the resulting value or error.
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

// MatchParameterByCompiledExpr matches parameter by compiled expr and returns the resulting value or error.
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
