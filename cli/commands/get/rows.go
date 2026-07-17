package get

import (
	"slices"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/cli/commands/get/table"
	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
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
	groupKeys := strfold.SortedKeys(cfg.ParameterGroups)
	for _, groupKey := range groupKeys {
		group := cfg.ParameterGroups[groupKey]
		paramKeys := strfold.SortedKeys(group.Parameters)
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
	for _, key := range strfold.SortedKeys(rootParams) {
		param := rootParams[key]
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, key, shared.DefaultRootGroupLabel)
		if !ok || !match {
			continue
		}
		if !shared.MatchParameterSearch(key, param, cfg, search) {
			continue
		}
		rows = append(rows, buildParameterRow(project, shared.DefaultRootGroupLabel, key, param, version, cachedAt, status, conditionOrder, conditionColors))
	}

	return rows
}

func buildParameterRow(project core.Project, group, key string, param firebase.RemoteConfigParam, version string, cachedAt time.Time, status string, conditionOrder map[string]int, conditionColors map[string]string) parameterRow {
	conditions := make([]parameterConditionJSON, 0, len(param.ConditionalValues))
	valueLines := make([]valueLine, 0, len(param.ConditionalValues)+1)

	for _, name := range table.SortedConditionalKeys(param.ConditionalValues, conditionOrder) {
		value := core.FormatRemoteConfigDisplayValue(param.ConditionalValues[name], param.ValueType)
		conditions = append(conditions, parameterConditionJSON{
			Name:  name,
			Value: table.ValueForJSON(value),
		})
		valueLines = append(valueLines, valueLine{
			Label:     name,
			Value:     value,
			Color:     clistyles.ConditionLipglossColor(conditionColors[name]),
			IsDefault: false,
			ValueType: table.ValueTypeKey(param.ValueType),
		})
	}

	var defaultValue *string
	if param.DefaultValue != nil {
		formatted := core.FormatRemoteConfigDisplayValue(*param.DefaultValue, param.ValueType)
		defaultValue = table.ValueForJSON(formatted)
		valueLines = append(valueLines, valueLine{
			Label:     "Default value",
			Value:     formatted,
			IsDefault: true,
			ValueType: table.ValueTypeKey(param.ValueType),
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
	return shared.SingleExactFilter(rawFilters)
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

func filterParameterRowsByProject(rows []parameterRow, rawFilters []string) []parameterRow {
	if len(shared.ParseFilters(rawFilters)) == 0 {
		return rows
	}

	filtered := make([]parameterRow, 0, len(rows))
	for _, row := range rows {
		project := core.Project{Name: row.Project, ProjectID: row.ProjectID}
		if len(shared.FilterProjects([]core.Project{project}, rawFilters)) > 0 {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

// singleExactParameterFilter reports whether table output can hide exact key column.
func singleExactParameterFilter(rawFilters []string) bool {
	return shared.SingleExactFilter(rawFilters)
}

func sortParameterRows(rows []parameterRow) {
	slices.SortFunc(rows, func(left, right parameterRow) int {
		if cmp := strfold.CompareProjects(left.Project, left.ProjectID, right.Project, right.ProjectID); cmp != 0 {
			return cmp
		}
		if cmp := strfold.CompareFolded(left.Group, right.Group); cmp != 0 {
			return cmp
		}
		return strfold.CompareFolded(left.Key, right.Key)
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
