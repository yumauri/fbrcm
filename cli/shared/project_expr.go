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
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
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
		mode, query := filter.ParseModePrefixedQuery(raw)
		if query == "" {
			continue
		}
		filters = append(filters, QueryFilter{Mode: mode, Query: query})
	}
	return filters
}

// SingleExactFilter reports whether filters contain exactly one non-empty exact query.
func SingleExactFilter(rawFilters []string) bool {
	exact := false
	for _, raw := range rawFilters {
		mode, query := filter.ParseModePrefixedQuery(raw)
		if strings.TrimSpace(query) == "" {
			continue
		}
		if mode != filter.ModeExact || exact {
			return false
		}
		exact = true
	}
	return exact
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

// FilterProjectTargets applies project filters after parsing an optional
// client@ or server@ template prefix. Explicit prefixes select one template;
// unqualified filters expand to each project's configured template views.
func FilterProjectTargets(projects []core.Project, rawFilters []string) ([]core.Project, error) {
	type targetFilter struct {
		target   rctarget.Target
		explicit bool
		filter   QueryFilter
	}
	filters := make([]targetFilter, 0, len(rawFilters))
	for _, raw := range rawFilters {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		target, explicit, err := rctarget.ParseSelector(raw)
		if err != nil {
			return nil, err
		}
		mode, query := filter.ParseModePrefixedQuery(target.ProjectID)
		if strings.TrimSpace(query) == "" {
			return nil, fmt.Errorf("%s template target requires a project filter", target.Kind)
		}
		filters = append(filters, targetFilter{
			target:   target,
			explicit: explicit,
			filter:   QueryFilter{Mode: mode, Query: query},
		})
	}

	filtered := make([]core.Project, 0, len(projects)*2)
	seen := make(map[string]struct{})
	for _, project := range projects {
		if len(filters) == 0 {
			appendConfiguredProjectTargets(&filtered, seen, project)
			continue
		}
		for _, item := range filters {
			nameMatch, _ := filter.Match(project.Name, item.filter.Query, item.filter.Mode)
			idMatch, _ := filter.Match(project.ProjectID, item.filter.Query, item.filter.Mode)
			if !nameMatch && !idMatch {
				continue
			}
			if !item.explicit {
				appendConfiguredProjectTargets(&filtered, seen, project)
			} else {
				appendProjectTarget(&filtered, seen, project, item.target.Kind)
			}
		}
	}
	return filtered, nil
}

func appendConfiguredProjectTargets(out *[]core.Project, seen map[string]struct{}, project core.Project) {
	for _, kind := range project.TemplateKinds() {
		appendProjectTarget(out, seen, project, kind)
	}
}

func appendProjectTarget(out *[]core.Project, seen map[string]struct{}, project core.Project, kind rctarget.Kind) {
	matched := project.TemplateTarget(kind)
	if _, ok := seen[matched.ProjectID]; ok {
		return
	}
	seen[matched.ProjectID] = struct{}{}
	*out = append(*out, matched)
}

// MatchProjectTarget reports whether an already-canonical target project
// matches any target-aware project filter.
func MatchProjectTarget(project core.Project, rawFilters []string) (bool, error) {
	hasFilter := false
	for _, raw := range rawFilters {
		if strings.TrimSpace(raw) != "" {
			hasFilter = true
			break
		}
	}
	if !hasFilter {
		return true, nil
	}
	current, err := rctarget.Parse(project.ProjectID)
	if err != nil {
		return false, err
	}
	base := project
	base.ProjectID = current.ProjectID
	matches, err := FilterProjectTargets([]core.Project{base}, rawFilters)
	if err != nil {
		return false, err
	}
	for _, match := range matches {
		if match.ProjectID == current.String() {
			return true, nil
		}
	}
	return false, nil
}

// SingleExactProjectTargetFilter reports whether project selection contains
// exactly one exact target filter.
func SingleExactProjectTargetFilter(rawFilters []string) bool {
	exact := false
	for _, raw := range rawFilters {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		target, err := rctarget.Parse(raw)
		if err != nil {
			return false
		}
		mode, query := filter.ParseModePrefixedQuery(target.ProjectID)
		if strings.TrimSpace(query) == "" {
			continue
		}
		if mode != filter.ModeExact || exact {
			return false
		}
		exact = true
	}
	return exact
}

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

// FilterProjectsByCachedExpr evaluates project expressions using only locally
// stored Remote Config. A missing cache supplies an empty config context, so
// expressions over project fields remain usable without contacting Firebase.
func FilterProjectsByCachedExpr(svc *core.Core, projects []core.Project, rawExpr string) ([]core.Project, error) {
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
		cfg, err := loadCachedProjectExprConfig(svc, project)
		if err != nil {
			corelog.For("filter").Error("cached project expression context load failed; skipping project", "project_id", project.ProjectID, "expr", rawExpr, "err", err)
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

	match, err := compiled.MatchProject(templateProjectID(project.ProjectID), project.Name, cfg)
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

	match, err := compiled.MatchParameter(templateProjectID(project.ProjectID), project.Name, cfg, name, group)
	if err != nil {
		corelog.For("filter").Error("parameter expression evaluation failed; skipping parameter", "project_id", project.ProjectID, "name", name, "group", group, "err", err)
		return false, false
	}

	return match, true
}

func MatchConditionByCompiledExpr(compiled *filter.Expression, project core.Project, entry core.ConditionEntry) (bool, bool) {
	if compiled == nil {
		return true, true
	}

	match, err := compiled.MatchCondition(templateProjectID(project.ProjectID), project.Name, entry)
	if err != nil {
		corelog.For("filter").Error("condition expression evaluation failed; skipping condition", "project_id", project.ProjectID, "condition", entry.Name, "err", err)
		return false, false
	}

	return match, true
}

func templateProjectID(value string) string {
	target, err := rctarget.Parse(value)
	if err != nil {
		return value
	}
	return target.ProjectID
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

func loadCachedProjectExprConfig(svc *core.Core, project core.Project) (*firebase.RemoteConfig, error) {
	return svc.LoadCachedRemoteConfig(project.ProjectID)
}
