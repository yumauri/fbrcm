package get

import (
	"sort"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
)

func flattenParameters(project core.Project, cfg *firebase.RemoteConfig, cachedAt time.Time, status, version string, compiledExpr *filter.Expression, search shared.ParameterSearch) []parameterRow {
	if cfg == nil {
		return nil
	}

	if version == "" {
		version = strings.TrimSpace(cfg.Version.VersionNumber)
	}

	conditionOrder := make(map[string]int, len(cfg.Conditions))
	conditionColors := make(map[string]string, len(cfg.Conditions))
	for i, condition := range cfg.Conditions {
		conditionOrder[condition.Name] = i
		conditionColors[condition.Name] = condition.TagColor
	}

	rows := make([]parameterRow, 0)
	seen := make(map[string]struct{})
	groupKeys := sortedStringKeys(cfg.ParameterGroups)
	for _, groupKey := range groupKeys {
		group := cfg.ParameterGroups[groupKey]
		paramKeys := sortedStringKeys(group.Parameters)
		for _, key := range paramKeys {
			param := group.Parameters[key]
			match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, key, groupKey)
			if !ok || !match {
				continue
			}
			if !shared.MatchParameterSearch(key, param, cfg, search) {
				continue
			}
			seen[key] = struct{}{}
			rows = append(rows, buildParameterRow(project, groupKey, key, param, version, cachedAt, status, conditionOrder, conditionColors))
		}
	}

	rootParams := make(map[string]firebase.RemoteConfigParam)
	for key, param := range cfg.Parameters {
		if _, ok := seen[key]; ok {
			continue
		}
		rootParams[key] = param
	}
	for _, key := range sortedStringKeys(rootParams) {
		param := rootParams[key]
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, key, defaultGroupLabel)
		if !ok || !match {
			continue
		}
		if !shared.MatchParameterSearch(key, param, cfg, search) {
			continue
		}
		rows = append(rows, buildParameterRow(project, defaultGroupLabel, key, param, version, cachedAt, status, conditionOrder, conditionColors))
	}

	return rows
}

func buildParameterRow(project core.Project, group, key string, param firebase.RemoteConfigParam, version string, cachedAt time.Time, status string, conditionOrder map[string]int, conditionColors map[string]string) parameterRow {
	conditions := make([]parameterConditionJSON, 0, len(param.ConditionalValues))
	valueLines := make([]valueLine, 0, len(param.ConditionalValues)+1)

	for _, name := range sortedConditionalKeys(param.ConditionalValues, conditionOrder) {
		value := formatRemoteConfigValue(param.ConditionalValues[name], param.ValueType)
		conditions = append(conditions, parameterConditionJSON{
			Name:  name,
			Value: valueForJSON(value),
		})
		valueLines = append(valueLines, valueLine{
			Label:     name,
			Value:     value,
			Color:     clistyles.ConditionLipglossColor(conditionColors[name]),
			IsDefault: false,
			ValueType: valueTypeKey(param.ValueType),
		})
	}

	var defaultValue *string
	if param.DefaultValue != nil {
		formatted := formatRemoteConfigValue(*param.DefaultValue, param.ValueType)
		defaultValue = valueForJSON(formatted)
		valueLines = append(valueLines, valueLine{
			Label:     "Default value",
			Value:     formatted,
			IsDefault: true,
			ValueType: valueTypeKey(param.ValueType),
		})
	}

	valueType := strings.TrimSpace(param.ValueType)
	if valueType == "" {
		valueType = "string"
	}

	return parameterRow{
		Project:      project.Name,
		ProjectID:    project.ProjectID,
		Group:        group,
		Key:          key,
		Description:  strings.TrimSpace(param.Description),
		DefaultValue: defaultValue,
		Conditional:  len(conditions) > 0,
		Conditions:   conditions,
		Type:         valueType,
		Version:      version,
		CachedAt:     cachedAt,
		Status:       status,
		ValueLines:   valueLines,
	}
}

// singleExactProjectFilter reports whether table output can omit project columns.
func singleExactProjectFilter(rawFilters []string) bool {
	return singleExactFilter(rawFilters)
}

func filterParameterRows(rows []parameterRow, rawFilters []string) []parameterRow {
	filters := shared.ParseFilters(rawFilters)
	if len(filters) == 0 {
		return rows
	}

	filtered := make([]parameterRow, 0, len(rows))
	for _, row := range rows {
		if shared.MatchAnyFilter(row.Key, filters) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

// singleExactParameterFilter reports whether table output can hide exact key column.
func singleExactParameterFilter(rawFilters []string) bool {
	return singleExactFilter(rawFilters)
}

func singleExactFilter(rawFilters []string) bool {
	exact := false
	for _, raw := range rawFilters {
		mode, query := parseFilter(raw)
		if strings.TrimSpace(query) == "" {
			continue
		}
		if mode != filter.ModeExact {
			return false
		}
		if exact {
			return false
		}
		exact = true
	}
	return exact
}

func parseFilter(raw string) (filter.Mode, string) {
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

func sortParameterRows(rows []parameterRow) {
	sort.Slice(rows, func(i, j int) bool {
		leftProject := strings.ToLower(strings.TrimSpace(rows[i].Project))
		rightProject := strings.ToLower(strings.TrimSpace(rows[j].Project))
		if leftProject == "" {
			leftProject = strings.ToLower(rows[i].ProjectID)
		}
		if rightProject == "" {
			rightProject = strings.ToLower(rows[j].ProjectID)
		}
		switch {
		case leftProject != rightProject:
			return leftProject < rightProject
		case !strings.EqualFold(rows[i].ProjectID, rows[j].ProjectID):
			return strings.ToLower(rows[i].ProjectID) < strings.ToLower(rows[j].ProjectID)
		case !strings.EqualFold(rows[i].Group, rows[j].Group):
			return strings.ToLower(rows[i].Group) < strings.ToLower(rows[j].Group)
		default:
			return strings.ToLower(rows[i].Key) < strings.ToLower(rows[j].Key)
		}
	})
}

func buildTableRows(projects []loadedProjectParameters, rows []parameterRow) []parameterRow {
	rowsByProject := make(map[string][]parameterRow, len(projects))
	for _, row := range rows {
		rowsByProject[row.ProjectID] = append(rowsByProject[row.ProjectID], row)
	}

	out := make([]parameterRow, 0, len(rows)+len(projects))
	for _, project := range projects {
		projectRows := rowsByProject[project.project.ProjectID]
		if len(projectRows) == 0 {
			out = append(out, parameterRow{
				Project:    project.project.Name,
				ProjectID:  project.project.ProjectID,
				Status:     project.status,
				ValueLines: []valueLine{{Label: "Missing values", Missing: true}},
			})
			continue
		}
		out = append(out, projectRows...)
	}

	return out
}
