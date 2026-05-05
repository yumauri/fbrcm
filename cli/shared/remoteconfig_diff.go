package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"

	clistyles "fbrcm/cli/styles"
	"fbrcm/core/firebase"
)

type ParamSlotPreview struct {
	Group string
	Param firebase.RemoteConfigParam
}

type diffCounts struct {
	added     int
	removed   int
	changed   int
	unchanged int
}

type paramView struct {
	Group string
	Param firebase.RemoteConfigParam
}

func RenderRemoteConfigDiff(currentCfg, finalCfg *firebase.RemoteConfig) (string, bool) {
	var sections []string

	conditionsText, conditionCounts := renderConditionsDiff(currentCfg, finalCfg)
	paramsText, parameterCounts := renderParametersDiff(currentCfg, finalCfg)

	if conditionsText != "" {
		sections = append(sections, conditionsText)
	}
	if paramsText != "" {
		sections = append(sections, paramsText)
	}

	hasChanges := conditionCounts.added+conditionCounts.removed+conditionCounts.changed+
		parameterCounts.added+parameterCounts.removed+parameterCounts.changed > 0
	if !hasChanges {
		return "", false
	}

	summary := fmt.Sprintf(
		"SUMMARY\n  %s condition added, %s removed, %s changed, %s unchanged\n  %s parameter added, %s removed, %s changed, %s unchanged",
		formatCount(conditionCounts.added),
		formatCount(conditionCounts.removed),
		formatCount(conditionCounts.changed),
		formatCount(conditionCounts.unchanged),
		formatCount(parameterCounts.added),
		formatCount(parameterCounts.removed),
		formatCount(parameterCounts.changed),
		formatCount(parameterCounts.unchanged),
	)
	sections = append(sections, summary)
	return strings.Join(sections, "\n\n"), true
}

func RenderConflictPreview(label string, currentValue, importValue any) string {
	switch current := currentValue.(type) {
	case firebase.RemoteConfigCondition:
		incoming, ok := importValue.(firebase.RemoteConfigCondition)
		if !ok {
			break
		}
		name := strings.TrimPrefix(label, "condition ")
		return fmt.Sprintf(
			"  ~ %-15s %s → %s",
			colorChanged(name),
			colorRemoved(formatConditionSummary(current)),
			colorAdded(formatConditionSummary(incoming)),
		)
	case ParamSlotPreview:
		incoming, ok := importValue.(ParamSlotPreview)
		if !ok {
			break
		}
		name := strings.TrimPrefix(label, "parameter ")
		lines := renderChangedParameter(name, paramView(current), paramView(incoming))
		return strings.Join(lines, "\n")
	case string:
		incoming, ok := importValue.(string)
		if !ok {
			break
		}
		name := strings.TrimSpace(strings.TrimPrefix(label, "group description "))
		if name == label {
			name = label
		}
		return fmt.Sprintf(
			"  ~ %-15s %s → %s",
			colorChanged(name),
			colorRemoved(formatPlainValue(current)),
			colorAdded(formatPlainValue(incoming)),
		)
	}

	currentJSON, err := json.MarshalIndent(currentValue, "", "  ")
	if err != nil {
		return fmt.Sprintf("current: %v\nimport: %v", currentValue, importValue)
	}
	importJSON, err := json.MarshalIndent(importValue, "", "  ")
	if err != nil {
		return fmt.Sprintf("current:\n%s\nimport: %v", string(currentJSON), importValue)
	}
	return fmt.Sprintf("current:\n%s\nimport:\n%s", string(NormalizeExportJSON(currentJSON)), string(NormalizeExportJSON(importJSON)))
}

func renderConditionsDiff(currentCfg, finalCfg *firebase.RemoteConfig) (string, diffCounts) {
	current := make(map[string]firebase.RemoteConfigCondition, len(currentCfg.Conditions))
	final := make(map[string]firebase.RemoteConfigCondition, len(finalCfg.Conditions))
	keys := make([]string, 0, len(currentCfg.Conditions)+len(finalCfg.Conditions))
	seen := make(map[string]struct{})

	for _, condition := range currentCfg.Conditions {
		current[condition.Name] = condition
		if _, ok := seen[condition.Name]; !ok {
			keys = append(keys, condition.Name)
			seen[condition.Name] = struct{}{}
		}
	}
	for _, condition := range finalCfg.Conditions {
		final[condition.Name] = condition
		if _, ok := seen[condition.Name]; !ok {
			keys = append(keys, condition.Name)
			seen[condition.Name] = struct{}{}
		}
	}
	sort.Strings(keys)

	var lines []string
	var counts diffCounts
	for _, key := range keys {
		left, hasLeft := current[key]
		right, hasRight := final[key]
		switch {
		case !hasLeft && hasRight:
			counts.added++
			lines = append(lines, fmt.Sprintf("  + %-15s %s", colorAdded(key), formatConditionSummary(right)))
		case hasLeft && !hasRight:
			counts.removed++
			lines = append(lines, fmt.Sprintf("  - %-15s %s", colorRemoved(key), formatConditionSummary(left)))
		case reflect.DeepEqual(left, right):
			counts.unchanged++
		default:
			counts.changed++
			lines = append(lines, fmt.Sprintf("  ~ %-15s %s → %s", colorChanged(key), colorRemoved(formatConditionSummary(left)), colorAdded(formatConditionSummary(right))))
		}
	}

	if len(lines) == 0 {
		return "", counts
	}
	return "CONDITIONS\n" + strings.Join(lines, "\n"), counts
}

func renderParametersDiff(currentCfg, finalCfg *firebase.RemoteConfig) (string, diffCounts) {
	current := collectParamViews(currentCfg)
	final := collectParamViews(finalCfg)
	keys := make([]string, 0, len(current)+len(final))
	seen := make(map[string]struct{})
	for key := range current {
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	for key := range final {
		if _, ok := seen[key]; ok {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var lines []string
	var counts diffCounts
	for _, key := range keys {
		left, hasLeft := current[key]
		right, hasRight := final[key]
		switch {
		case !hasLeft && hasRight:
			counts.added++
			lines = append(lines, renderAddedParameter(key, right)...)
		case hasLeft && !hasRight:
			counts.removed++
			lines = append(lines, renderRemovedParameter(key, left)...)
		case reflect.DeepEqual(left, right):
			counts.unchanged++
		default:
			counts.changed++
			lines = append(lines, renderChangedParameter(key, left, right)...)
		}
	}
	if len(lines) == 0 {
		return "", counts
	}
	return "PARAMETERS\n" + strings.Join(lines, "\n"), counts
}

func collectParamViews(cfg *firebase.RemoteConfig) map[string]paramView {
	out := make(map[string]paramView)
	for key, param := range cfg.Parameters {
		out[key] = paramView{Param: param}
	}
	for groupName, group := range cfg.ParameterGroups {
		for key, param := range group.Parameters {
			out[key] = paramView{Group: groupName, Param: param}
		}
	}
	return out
}

func renderAddedParameter(key string, value paramView) []string {
	lines := []string{fmt.Sprintf("  + %s", colorAdded(formatParameterHeader(key, value.Group)))}
	lines = append(lines, renderAddedParameterDetails(value.Param)...)
	return lines
}

func renderRemovedParameter(key string, value paramView) []string {
	return []string{fmt.Sprintf("  - %s", colorRemoved(formatParameterHeader(key, value.Group)))}
}

func renderChangedParameter(key string, left, right paramView) []string {
	lines := []string{fmt.Sprintf("  ~ %s", colorChanged(formatParameterHeader(key, right.Group)))}
	if left.Group != right.Group {
		lines = append(lines, fmt.Sprintf("      ~ group:               %s → %s", colorRemoved(formatGroupValue(left.Group)), colorAdded(formatGroupValue(right.Group))))
	}
	if left.Param.ValueType != right.Param.ValueType {
		lines = append(lines, fmt.Sprintf("      ~ type:                %s → %s", colorRemoved(emptyAsDash(left.Param.ValueType)), colorAdded(emptyAsDash(right.Param.ValueType))))
	}
	if left.Param.Description != right.Param.Description {
		lines = append(lines, fmt.Sprintf("      ~ description:         %s → %s", colorRemoved(formatPlainValue(left.Param.Description)), colorAdded(formatPlainValue(right.Param.Description))))
	}
	lines = append(lines, renderParameterValueChanges(left.Param, right.Param)...)
	return lines
}

func renderAddedParameterDetails(param firebase.RemoteConfigParam) []string {
	var lines []string
	if param.DefaultValue != nil {
		lines = append(lines, fmt.Sprintf("      + default:             %s", formatRemoteValue(*param.DefaultValue)))
	}
	for _, condition := range sortedConditionalNames(param.ConditionalValues) {
		lines = append(lines, fmt.Sprintf("      + cond %-15s %s", condition+":", formatRemoteValue(param.ConditionalValues[condition])))
	}
	return lines
}

func renderParameterValueChanges(left, right firebase.RemoteConfigParam) []string {
	var lines []string
	switch {
	case left.DefaultValue == nil && right.DefaultValue != nil:
		lines = append(lines, fmt.Sprintf("      + default:             %s", colorAdded(formatRemoteValue(*right.DefaultValue))))
	case left.DefaultValue != nil && right.DefaultValue == nil:
		lines = append(lines, fmt.Sprintf("      - default:             %s", colorRemoved(formatRemoteValue(*left.DefaultValue))))
	case left.DefaultValue != nil && right.DefaultValue != nil && !reflect.DeepEqual(*left.DefaultValue, *right.DefaultValue):
		lines = append(lines, fmt.Sprintf("      ~ default:             %s → %s", colorRemoved(formatRemoteValue(*left.DefaultValue)), colorAdded(formatRemoteValue(*right.DefaultValue))))
	}
	condKeys := unionConditionalNames(left.ConditionalValues, right.ConditionalValues)
	for _, condition := range condKeys {
		lv, hasLeft := left.ConditionalValues[condition]
		rv, hasRight := right.ConditionalValues[condition]
		label := fmt.Sprintf("cond %-15s", condition+":")
		switch {
		case !hasLeft && hasRight:
			lines = append(lines, fmt.Sprintf("      + %s %s", label, colorAdded(formatRemoteValue(rv))))
		case hasLeft && !hasRight:
			lines = append(lines, fmt.Sprintf("      - %s %s", label, colorRemoved(formatRemoteValue(lv))))
		case hasLeft && hasRight && !reflect.DeepEqual(lv, rv):
			lines = append(lines, fmt.Sprintf("      ~ %s %s → %s", label, colorRemoved(formatRemoteValue(lv)), colorAdded(formatRemoteValue(rv))))
		}
	}
	return lines
}

func unionConditionalNames(left, right map[string]firebase.RemoteConfigValue) []string {
	keys := make([]string, 0, len(left)+len(right))
	seen := make(map[string]struct{})
	for key := range left {
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	for key := range right {
		if _, ok := seen[key]; ok {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedConditionalNames(values map[string]firebase.RemoteConfigValue) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func formatConditionSummary(condition firebase.RemoteConfigCondition) string {
	parts := make([]string, 0, 3)
	if strings.TrimSpace(condition.Expression) != "" {
		parts = append(parts, condition.Expression)
	}
	if strings.TrimSpace(condition.Description) != "" {
		parts = append(parts, "desc="+condition.Description)
	}
	if strings.TrimSpace(condition.TagColor) != "" {
		parts = append(parts, "color="+condition.TagColor)
	}
	if len(parts) == 0 {
		return "(empty)"
	}
	return strings.Join(parts, " | ")
}

func formatParameterHeader(key, group string) string {
	if group == "" {
		return key
	}
	return fmt.Sprintf("%s [%s]", key, group)
}

func formatRemoteValue(value firebase.RemoteConfigValue) string {
	switch {
	case len(value.PersonalizationValue) > 0:
		return string(NormalizeExportJSON(bytes.TrimSpace(value.PersonalizationValue)))
	case len(value.RolloutValue) > 0:
		return string(NormalizeExportJSON(bytes.TrimSpace(value.RolloutValue)))
	case value.UseInAppDefault:
		return "useInAppDefault"
	default:
		return formatPlainValue(value.Value)
	}
}

func formatPlainValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(empty)"
	}
	quoted, err := json.Marshal(value)
	if err != nil {
		return value
	}
	if isSimpleToken(value) {
		return value
	}
	return string(quoted)
}

func isSimpleToken(value string) bool {
	for _, r := range value {
		if r == ' ' || r == '\t' || r == '\n' || r == '"' {
			return false
		}
	}
	return true
}

func formatGroupValue(group string) string {
	if group == "" {
		return "(root)"
	}
	return "[" + group + "]"
}

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(empty)"
	}
	return value
}

func formatCount(count int) string {
	return fmt.Sprintf("%d", count)
}

func colorAdded(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorAdded).Render(value)
}

func colorRemoved(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorRemoved).Render(value)
}

func colorChanged(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorChanged).Render(value)
}
