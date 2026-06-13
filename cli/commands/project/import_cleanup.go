package project

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/firebase"
)

func configHasContent(cfg *firebase.RemoteConfig) bool {
	return cfg != nil && (len(cfg.Conditions) > 0 || len(cfg.Parameters) > 0 || len(cfg.ParameterGroups) > 0)
}

func pruneUnusedConditions(cfg *firebase.RemoteConfig) {
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

func dropUnknownConditionReferences(cfg *firebase.RemoteConfig) {
	allowed := make(map[string]struct{}, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		allowed[condition.Name] = struct{}{}
	}
	cfg.Parameters = stripUnknownConditionRefs(cfg.Parameters, allowed)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = stripUnknownConditionRefs(group.Parameters, allowed)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func stripUnknownConditionRefs(params map[string]firebase.RemoteConfigParam, allowed map[string]struct{}) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		if len(param.ConditionalValues) > 0 {
			filtered := make(map[string]firebase.RemoteConfigValue, len(param.ConditionalValues))
			for cond, value := range param.ConditionalValues {
				if _, ok := allowed[cond]; !ok {
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

type groupSummary struct {
	Name       string
	Parameters int
}

func summarizeGroups(groups map[string]firebase.RemoteConfigGroup) []groupSummary {
	names := shared.SortedStringKeys(groups)
	out := make([]groupSummary, 0, len(names))
	for _, name := range names {
		out = append(out, groupSummary{
			Name:       name,
			Parameters: len(groups[name].Parameters),
		})
	}
	return out
}

func renderGroupsTable(groups []groupSummary) string {
	rows := make([][]string, 0, len(groups))
	groupWidth := lipgloss.Width("Group")
	parametersWidth := lipgloss.Width("Parameters")
	for _, group := range groups {
		count := fmt.Sprintf("%d", group.Parameters)
		rows = append(rows, []string{group.Name, count})
		groupWidth = max(groupWidth, lipgloss.Width(group.Name))
		parametersWidth = max(parametersWidth, lipgloss.Width(count))
	}

	styleFunc := func(row, col int) lipgloss.Style {
		_ = col
		style := lipgloss.NewStyle().Padding(0, 1)
		if clistyles.NoColorEnabled() {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateBright)
	}

	tbl := table.New().
		Headers("Group", "Parameters").
		Rows(rows...).
		Width(groupWidth + parametersWidth + 7).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func removeEmptyGroups(cfg *firebase.RemoteConfig) {
	for groupName, group := range cfg.ParameterGroups {
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
		}
	}
	if len(cfg.ParameterGroups) == 0 {
		cfg.ParameterGroups = nil
	}
	if len(cfg.Parameters) == 0 {
		cfg.Parameters = nil
	}
}
