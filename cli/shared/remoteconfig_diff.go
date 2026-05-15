package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/firebase"
)

// ParamSlotPreview holds param slot preview state used by the shared package.
type ParamSlotPreview struct {
	// Group stores group for ParamSlotPreview.
	Group string
	// Param stores param for ParamSlotPreview.
	Param firebase.RemoteConfigParam
}

// diffCounts holds diff counts state used by the shared package.
type diffCounts struct {
	// added stores added for diffCounts.
	added int
	// removed stores removed for diffCounts.
	removed int
	// changed stores changed for diffCounts.
	changed int
	// unchanged stores unchanged for diffCounts.
	unchanged int
}

// paramView holds param view state used by the shared package.
type paramView struct {
	// Group stores group for paramView.
	Group string
	// Param stores param for paramView.
	Param firebase.RemoteConfigParam
}

// RenderRemoteConfigDiff renders remote config diff and returns the resulting value or error.
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
		"Summary:\n  %s condition added, %s removed, %s changed, %s unchanged\n  %s parameter added, %s removed, %s changed, %s unchanged",
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
	return "\n" + strings.Join(sections, "\n\n") + "\n", true
}

// RenderConflictPreview renders conflict preview and returns the resulting value or error.
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

// RenderConflictChoiceValue renders conflict choice value and returns the resulting value or error.
func RenderConflictChoiceValue(value any) string {
	switch v := value.(type) {
	case ParamSlotPreview:
		return summarizeParamSlotChoice(v)
	case firebase.RemoteConfigCondition:
		return trimPreview(formatConditionSummary(v))
	case string:
		return trimPreview(formatPlainValue(v))
	default:
		body, err := json.Marshal(value)
		if err != nil {
			return trimPreview(fmt.Sprintf("%v", value))
		}
		return trimPreview(string(NormalizeExportJSON(body)))
	}
}

// renderConditionsDiff renders render conditions diff and returns the resulting value or error.
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
	return "Conditions:\n" + strings.Join(lines, "\n"), counts
}

// renderParametersDiff renders render parameters diff and returns the resulting value or error.
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
	return "Parameters:\n" + strings.Join(lines, "\n"), counts
}

// collectParamViews handles collect param views and returns the resulting value or error.
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

// renderAddedParameter renders render added parameter and returns the resulting value or error.
func renderAddedParameter(key string, value paramView) []string {
	lines := []string{fmt.Sprintf("  + %s", colorAdded(formatParameterHeader(key, value.Group)))}
	lines = append(lines, renderAddedParameterDetails(value.Param)...)
	return lines
}

// renderRemovedParameter renders render removed parameter and returns the resulting value or error.
func renderRemovedParameter(key string, value paramView) []string {
	return []string{fmt.Sprintf("  - %s", colorRemoved(formatParameterHeader(key, value.Group)))}
}

// renderChangedParameter renders render changed parameter and returns the resulting value or error.
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

// renderAddedParameterDetails renders render added parameter details and returns the resulting value or error.
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

// renderParameterValueChanges renders render parameter value changes and returns the resulting value or error.
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

// unionConditionalNames handles union conditional names and returns the resulting value or error.
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

// sortedConditionalNames handles sorted conditional names and returns the resulting value or error.
func sortedConditionalNames(values map[string]firebase.RemoteConfigValue) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// formatConditionSummary formats format condition summary and returns the resulting value or error.
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

// formatParameterHeader formats format parameter header and returns the resulting value or error.
func formatParameterHeader(key, group string) string {
	if group == "" {
		return key
	}
	return fmt.Sprintf("%s [%s]", key, group)
}

// summarizeParamSlotChoice handles summarize param slot choice and returns the resulting value or error.
func summarizeParamSlotChoice(slot ParamSlotPreview) string {
	parts := make([]string, 0, 4)
	if slot.Group != "" {
		parts = append(parts, "group="+formatGroupValue(slot.Group))
	}
	if slot.Param.DefaultValue != nil {
		parts = append(parts, "default="+formatRemoteValue(*slot.Param.DefaultValue))
	}
	for _, condition := range sortedConditionalNames(slot.Param.ConditionalValues) {
		parts = append(parts, condition+"="+formatRemoteValue(slot.Param.ConditionalValues[condition]))
	}
	if strings.TrimSpace(slot.Param.ValueType) != "" {
		parts = append(parts, "type="+slot.Param.ValueType)
	}
	if strings.TrimSpace(slot.Param.Description) != "" {
		parts = append(parts, "desc="+formatPlainValue(slot.Param.Description))
	}
	if len(parts) == 0 {
		return "(empty)"
	}
	return trimPreview(strings.Join(parts, " | "))
}

// trimPreview handles trim preview and returns the resulting value or error.
func trimPreview(value string) string {
	const maxLen = 72
	value = strings.TrimSpace(value)
	if value == "" {
		return "(empty)"
	}
	if len(value) <= maxLen {
		return value
	}
	return strings.TrimSpace(value[:maxLen-1]) + "…"
}

// formatRemoteValue formats format remote value and returns the resulting value or error.
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

// formatPlainValue formats format plain value and returns the resulting value or error.
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

// isSimpleToken reports is simple token and returns the resulting value or error.
func isSimpleToken(value string) bool {
	for _, r := range value {
		if r == ' ' || r == '\t' || r == '\n' || r == '"' {
			return false
		}
	}
	return true
}

// formatGroupValue formats format group value and returns the resulting value or error.
func formatGroupValue(group string) string {
	if group == "" {
		return "(root)"
	}
	return "[" + group + "]"
}

// emptyAsDash handles empty as dash and returns the resulting value or error.
func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(empty)"
	}
	return value
}

// formatCount formats format count and returns the resulting value or error.
func formatCount(count int) string {
	return fmt.Sprintf("%d", count)
}

// colorAdded handles color added and returns the resulting value or error.
func colorAdded(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorAdded).Render(value)
}

// colorRemoved handles color removed and returns the resulting value or error.
func colorRemoved(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorRemoved).Render(value)
}

// colorChanged handles color changed and returns the resulting value or error.
func colorChanged(value string) string {
	if clistyles.NoColorEnabled() || value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(clistyles.ColorChanged).Render(value)
}
