package importpkg

import (
	"strings"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
)

func normalizeGroups(groups []string) []string {
	seen := make(map[string]struct{}, len(groups))
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		out = append(out, group)
	}
	return out
}

func transformImportConfig(project core.Project, cfg *firebase.RemoteConfig, opts importOptions) error {
	if len(opts.groups) > 0 {
		if err := filterImportGroups(cfg, opts.groups); err != nil {
			return err
		}
		pruneUnusedConditions(cfg)
	}
	if len(shared.ParseFilters(opts.paramFilters)) > 0 {
		filterImportParameters(cfg, opts.paramFilters)
		pruneUnusedConditions(cfg)
	}
	if !opts.search.Empty() {
		filterImportParametersBySearch(cfg, opts.search)
		pruneUnusedConditions(cfg)
	}
	if opts.expr != "" {
		compiledExpr, ok := shared.CompileExpr(opts.expr, project.ProjectID)
		if !ok {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
			cfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
			cfg.Conditions = nil
			return nil
		}
		filterImportParametersByExpr(project, cfg, compiledExpr)
		pruneUnusedConditions(cfg)
	}

	switch {
	case opts.removeAllConditions:
		removeAllConditions(cfg)
	case opts.removeProjectSpecificConditions:
		removeProjectSpecificConditions(cfg)
	}

	pruneUnusedConditions(cfg)
	dropUnknownConditionReferences(cfg)
	removeEmptyGroups(cfg)
	return nil
}

func filterImportGroups(cfg *firebase.RemoteConfig, groups []string) error {
	selected := make(map[string]firebase.RemoteConfigGroup, len(groups))
	missing := make([]string, 0)
	for _, group := range groups {
		value, ok := cfg.ParameterGroups[group]
		if !ok {
			missing = append(missing, group)
			continue
		}
		selected[group] = value
	}
	if len(missing) > 0 {
		return &missingImportGroupsError{
			missing:   append([]string(nil), missing...),
			available: summarizeGroups(cfg.ParameterGroups),
		}
	}
	cfg.Parameters = nil
	cfg.ParameterGroups = selected
	return nil
}

func filterImportParameters(cfg *firebase.RemoteConfig, rawFilters []string) {
	filters := shared.ParseFilters(rawFilters)
	if len(filters) == 0 {
		return
	}

	cfg.Parameters = filterImportParamMap(cfg.Parameters, filters)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = filterImportParamMap(group.Parameters, filters)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func filterImportParametersBySearch(cfg *firebase.RemoteConfig, search shared.ParameterSearch) {
	if search.Empty() {
		return
	}

	cfg.Parameters = filterImportParamMapBySearch(cfg, cfg.Parameters, search)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = filterImportParamMapBySearch(cfg, group.Parameters, search)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func filterImportParametersByExpr(project core.Project, cfg *firebase.RemoteConfig, compiledExpr *filter.Expression) {
	if compiledExpr == nil {
		return
	}

	cfg.Parameters = filterImportParamMapByExpr(project, cfg, cfg.Parameters, shared.DefaultRootGroupLabel, compiledExpr)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = filterImportParamMapByExpr(project, cfg, group.Parameters, groupName, compiledExpr)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func filterImportParamMap(params map[string]firebase.RemoteConfigParam, filters []shared.QueryFilter) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}

	filtered := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		if shared.MatchAnyFilter(key, filters) {
			filtered[key] = param
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterImportParamMapBySearch(cfg *firebase.RemoteConfig, params map[string]firebase.RemoteConfigParam, search shared.ParameterSearch) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}

	filtered := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		if shared.MatchParameterSearch(key, param, cfg, search) {
			filtered[key] = param
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterImportParamMapByExpr(project core.Project, cfg *firebase.RemoteConfig, params map[string]firebase.RemoteConfigParam, groupName string, compiledExpr *filter.Expression) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}

	filtered := make(map[string]firebase.RemoteConfigParam, len(params))
	for key := range params {
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, key, groupName)
		if ok && match {
			filtered[key] = params[key]
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func removeAllConditions(cfg *firebase.RemoteConfig) {
	cfg.Conditions = nil
	cfg.Parameters = stripAllConditionalValues(cfg.Parameters, nil)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = stripAllConditionalValues(group.Parameters, nil)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func removeProjectSpecificConditions(cfg *firebase.RemoteConfig) {
	deleted := make(map[string]struct{})
	kept := make([]firebase.RemoteConfigCondition, 0, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		if isProjectSpecificCondition(condition.Expression) {
			deleted[condition.Name] = struct{}{}
			continue
		}
		kept = append(kept, condition)
	}
	cfg.Conditions = kept
	cfg.Parameters = stripAllConditionalValues(cfg.Parameters, deleted)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = stripAllConditionalValues(group.Parameters, deleted)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func stripAllConditionalValues(params map[string]firebase.RemoteConfigParam, deleted map[string]struct{}) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		if len(param.ConditionalValues) > 0 {
			filtered := make(map[string]firebase.RemoteConfigValue, len(param.ConditionalValues))
			for cond, value := range param.ConditionalValues {
				if deleted == nil {
					continue
				}
				if _, ok := deleted[cond]; ok {
					continue
				}
				filtered[cond] = value
			}
			if len(filtered) > 0 {
				param.ConditionalValues = filtered
			} else {
				param.ConditionalValues = nil
			}
		}
		if param.DefaultValue == nil && len(param.ConditionalValues) == 0 {
			continue
		}
		out[key] = param
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isProjectSpecificCondition(expr string) bool {
	for _, needle := range []string{
		"inExperiment",
		"inUserAudience",
		"app.id",
		"app.userProperty[",
		"app.firebaseInstallationId",
		"app.instanceId",
		"app.instance_id",
	} {
		if strings.Contains(expr, needle) {
			return true
		}
	}
	return false
}
