package importer

import (
	"strings"
	"unicode"

	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/core/strfold"
)

func Transform(projectID, projectName string, cfg *firebase.RemoteConfig, opts Options) error {
	if cfg == nil {
		return nil
	}
	groups := normalizeNames(opts.Groups)
	if len(groups) > 0 {
		if err := filterGroups(cfg, groups); err != nil {
			return err
		}
		PruneUnusedConditions(cfg)
	}
	filters := parseFilters(opts.Filters)
	if len(filters) > 0 {
		filterParameters(cfg, func(name string, _ firebase.RemoteConfigParam) bool {
			return matchAnyFilter(name, filters)
		})
		PruneUnusedConditions(cfg)
	}
	if search := newSearch(opts.Search); !search.empty() {
		filterParameters(cfg, func(name string, param firebase.RemoteConfigParam) bool {
			return matchSearch(name, param, cfg, search)
		})
		PruneUnusedConditions(cfg)
	}
	if rawExpr := strings.TrimSpace(opts.Expr); rawExpr != "" {
		compiled, err := filter.CompileExpression(rawExpr)
		if err != nil {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
			cfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
			cfg.Conditions = nil
			return nil
		}
		filterParameterMaps(cfg, func(name, group string, _ firebase.RemoteConfigParam) bool {
			match, matchErr := compiled.MatchParameter(projectID, projectName, cfg, name, group)
			return matchErr == nil && match
		})
		PruneUnusedConditions(cfg)
	}

	switch opts.ConditionPolicy {
	case ConditionPolicyRemoveAll:
		removeAllConditions(cfg)
	case ConditionPolicyKeepPortableOnly:
		keepPortableConditionsOnly(cfg)
	}

	Cleanup(cfg)
	return nil
}

func filterGroups(cfg *firebase.RemoteConfig, groups []string) error {
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
		return &MissingGroupsError{Missing: missing, Available: SummarizeGroups(cfg.ParameterGroups)}
	}
	cfg.Parameters = nil
	cfg.ParameterGroups = selected
	return nil
}

type queryFilter struct {
	mode  filter.Mode
	query string
}

func parseFilters(raw []string) []queryFilter {
	out := make([]queryFilter, 0, len(raw))
	for _, value := range raw {
		mode, query := filter.ParseModePrefixedQuery(value)
		if query != "" {
			out = append(out, queryFilter{mode: mode, query: query})
		}
	}
	return out
}

func matchAnyFilter(value string, filters []queryFilter) bool {
	for _, item := range filters {
		if match, _ := filter.Match(value, item.query, item.mode); match {
			return true
		}
	}
	return len(filters) == 0
}

func filterParameters(cfg *firebase.RemoteConfig, keep func(string, firebase.RemoteConfigParam) bool) {
	filterParameterMaps(cfg, func(name, _ string, param firebase.RemoteConfigParam) bool { return keep(name, param) })
}

func filterParameterMaps(cfg *firebase.RemoteConfig, keep func(name, group string, param firebase.RemoteConfigParam) bool) {
	cfg.Parameters = filterParameterMap(cfg.Parameters, rootgroup.Label, keep)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = filterParameterMap(group.Parameters, groupName, keep)
		cfg.ParameterGroups[groupName] = group
	}
}

func filterParameterMap(params map[string]firebase.RemoteConfigParam, group string, keep func(name, group string, param firebase.RemoteConfigParam) bool) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]firebase.RemoteConfigParam, len(params))
	for name, param := range params {
		if keep(name, group, param) {
			out[name] = param
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

type parameterSearch struct{ raw, normalized string }

func newSearch(raw string) parameterSearch {
	return parameterSearch{raw: collapseSpaces(raw), normalized: normalizeSearch(raw)}
}

func (s parameterSearch) empty() bool { return s.raw == "" && s.normalized == "" }

func matchSearch(name string, param firebase.RemoteConfigParam, cfg *firebase.RemoteConfig, search parameterSearch) bool {
	conditionByName := make(map[string]firebase.RemoteConfigCondition, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		conditionByName[condition.Name] = condition
	}
	conditionNames := strfold.SortedKeys(param.ConditionalValues)
	normalizedParts := []string{name, param.Description}
	rawParts := make([]string, 0, 1+len(conditionNames)*2)
	if param.DefaultValue != nil {
		rawParts = append(rawParts, param.DefaultValue.Value)
	}
	for _, conditionName := range conditionNames {
		normalizedParts = append(normalizedParts, conditionName)
		rawParts = append(rawParts, param.ConditionalValues[conditionName].Value)
		if condition, ok := conditionByName[conditionName]; ok {
			rawParts = append(rawParts, condition.Expression)
		}
	}
	return search.normalized != "" && strings.Contains(normalizeSearch(strings.Join(normalizedParts, " ")), search.normalized) ||
		search.raw != "" && strings.Contains(strings.Join(rawParts, " "), search.raw)
}

func normalizeSearch(value string) string {
	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(' ')
		}
	}
	return collapseSpaces(b.String())
}

func collapseSpaces(value string) string { return strings.Join(strings.Fields(value), " ") }

func removeAllConditions(cfg *firebase.RemoteConfig) {
	cfg.Conditions = nil
	cfg.Parameters = stripConditionalValues(cfg.Parameters, nil)
	for name, group := range cfg.ParameterGroups {
		group.Parameters = stripConditionalValues(group.Parameters, nil)
		cfg.ParameterGroups[name] = group
	}
}

func keepPortableConditionsOnly(cfg *firebase.RemoteConfig) {
	deleted := make(map[string]struct{})
	kept := make([]firebase.RemoteConfigCondition, 0, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		if isNonPortableCondition(condition.Expression) {
			deleted[condition.Name] = struct{}{}
		} else {
			kept = append(kept, condition)
		}
	}
	cfg.Conditions = kept
	cfg.Parameters = stripConditionalValues(cfg.Parameters, deleted)
	for name, group := range cfg.ParameterGroups {
		group.Parameters = stripConditionalValues(group.Parameters, deleted)
		cfg.ParameterGroups[name] = group
	}
}

func stripConditionalValues(params map[string]firebase.RemoteConfigParam, deleted map[string]struct{}) map[string]firebase.RemoteConfigParam {
	out := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		filtered := make(map[string]firebase.RemoteConfigValue)
		for name, value := range param.ConditionalValues {
			if deleted != nil {
				if _, remove := deleted[name]; !remove {
					filtered[name] = value
				}
			}
		}
		param.ConditionalValues = filtered
		if len(filtered) == 0 {
			param.ConditionalValues = nil
		}
		if param.DefaultValue != nil || len(param.ConditionalValues) > 0 {
			out[key] = param
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isNonPortableCondition(expression string) bool {
	for _, needle := range []string{
		"inExperiment",
		"inUserAudience",
		"app.audiences",
		"app.id",
		"app.userProperty[",
		"app.customSignal[",
		"app.firebaseInstallationId",
		"app.instanceId",
		"app.instance_id",
	} {
		if strings.Contains(expression, needle) {
			return true
		}
	}
	return false
}

func normalizeNames(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; !ok {
			seen[value] = struct{}{}
			out = append(out, value)
		}
	}
	return out
}

func Cleanup(cfg *firebase.RemoteConfig) {
	PruneUnusedConditions(cfg)
	DropUnknownConditionReferences(cfg)
	NormalizeEmptyParameterMaps(cfg)
}

func DropUnknownConditionReferences(cfg *firebase.RemoteConfig) {
	rcmutate.DropUnknownConditionReferences(cfg)
}

func NormalizeEmptyParameterMaps(cfg *firebase.RemoteConfig) {
	rcmutate.NormalizeEmptyParameterMaps(cfg)
}

func PruneUnusedConditions(cfg *firebase.RemoteConfig) {
	if cfg == nil || len(cfg.Conditions) == 0 {
		return
	}
	used := make(map[string]struct{})
	collectUsedConditions(used, cfg.Parameters)
	for _, group := range cfg.ParameterGroups {
		collectUsedConditions(used, group.Parameters)
	}
	kept := make([]firebase.RemoteConfigCondition, 0, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		if _, ok := used[condition.Name]; ok {
			kept = append(kept, condition)
		}
	}
	cfg.Conditions = kept
}

func collectUsedConditions(used map[string]struct{}, params map[string]firebase.RemoteConfigParam) {
	for _, param := range params {
		for condition := range param.ConditionalValues {
			used[condition] = struct{}{}
		}
	}
}

func ConfigHasContent(cfg *firebase.RemoteConfig) bool {
	return cfg != nil && (len(cfg.Conditions) > 0 || len(cfg.Parameters) > 0 || len(cfg.ParameterGroups) > 0)
}

func SummarizeGroups(groups map[string]firebase.RemoteConfigGroup) []GroupSummary {
	names := strfold.SortedKeys(groups)
	out := make([]GroupSummary, 0, len(names))
	for _, name := range names {
		out = append(out, GroupSummary{Name: name, Parameters: len(groups[name].Parameters)})
	}
	return out
}

func joinNames(values []string) string { return strings.Join(values, ", ") }
