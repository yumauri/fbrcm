package diff

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/strfold"
)

// RenderRemovedParameterDetail renders a red-colored preview of a parameter that
// is about to be deleted, including its type, description, default, and
// conditional values.
func RenderRemovedParameterDetail(key, group string, param firebase.RemoteConfigParam) string {
	lines := []string{fmt.Sprintf("  - %s", colorRemoved(rcdisplay.FormatParameterHeader(key, group)))}
	if strings.TrimSpace(param.ValueType) != "" {
		lines = append(lines, fmt.Sprintf("      - type:                %s", colorRemoved(param.ValueType)))
	}
	if strings.TrimSpace(param.Description) != "" {
		lines = append(lines, fmt.Sprintf("      - description:         %s", colorRemoved(rcdisplay.FormatPlainValue(param.Description))))
	}
	if param.DefaultValue != nil {
		lines = append(lines, fmt.Sprintf("      - default:             %s", colorRemoved(rcdisplay.FormatDiff(*param.DefaultValue))))
	}
	for _, condition := range strfold.SortedKeys(param.ConditionalValues) {
		lines = append(lines, fmt.Sprintf("      - cond %-15s %s", condition+":", colorRemoved(rcdisplay.FormatDiff(param.ConditionalValues[condition]))))
	}
	return "\n" + strings.Join(lines, "\n")
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
	lines := []string{fmt.Sprintf("  + %s", colorAdded(rcdisplay.FormatParameterHeader(key, value.Group)))}
	lines = append(lines, renderAddedParameterDetails(value.Param)...)
	return lines
}

func renderRemovedParameter(key string, value paramView) []string {
	return []string{fmt.Sprintf("  - %s", colorRemoved(rcdisplay.FormatParameterHeader(key, value.Group)))}
}

func renderChangedParameter(key string, left, right paramView) []string {
	lines := []string{fmt.Sprintf("  ~ %s", colorChanged(rcdisplay.FormatParameterHeader(key, right.Group)))}
	if left.Group != right.Group {
		lines = append(lines, fmt.Sprintf("      ~ group:               %s → %s", colorRemoved(formatGroupValue(left.Group)), colorAdded(formatGroupValue(right.Group))))
	}
	if left.Param.ValueType != right.Param.ValueType {
		lines = append(lines, fmt.Sprintf("      ~ type:                %s → %s", colorRemoved(emptyAsDash(left.Param.ValueType)), colorAdded(emptyAsDash(right.Param.ValueType))))
	}
	if left.Param.Description != right.Param.Description {
		lines = append(lines, fmt.Sprintf("      ~ description:         %s → %s", colorRemoved(rcdisplay.FormatPlainValue(left.Param.Description)), colorAdded(rcdisplay.FormatPlainValue(right.Param.Description))))
	}
	lines = append(lines, renderParameterValueChanges(left.Param, right.Param)...)
	return lines
}

func renderAddedParameterDetails(param firebase.RemoteConfigParam) []string {
	var lines []string
	if param.DefaultValue != nil {
		lines = append(lines, fmt.Sprintf("      + default:             %s", rcdisplay.FormatDiff(*param.DefaultValue)))
	}
	for _, condition := range sortedConditionalNames(param.ConditionalValues) {
		lines = append(lines, fmt.Sprintf("      + cond %-15s %s", condition+":", rcdisplay.FormatDiff(param.ConditionalValues[condition])))
	}
	return lines
}

func renderParameterValueChanges(left, right firebase.RemoteConfigParam) []string {
	var lines []string
	switch {
	case left.DefaultValue == nil && right.DefaultValue != nil:
		lines = append(lines, fmt.Sprintf("      + default:             %s", colorAdded(rcdisplay.FormatDiff(*right.DefaultValue))))
	case left.DefaultValue != nil && right.DefaultValue == nil:
		lines = append(lines, fmt.Sprintf("      - default:             %s", colorRemoved(rcdisplay.FormatDiff(*left.DefaultValue))))
	case left.DefaultValue != nil && right.DefaultValue != nil && !reflect.DeepEqual(*left.DefaultValue, *right.DefaultValue):
		lines = append(lines, fmt.Sprintf("      ~ default:             %s → %s", colorRemoved(rcdisplay.FormatDiff(*left.DefaultValue)), colorAdded(rcdisplay.FormatDiff(*right.DefaultValue))))
	}
	condKeys := unionConditionalNames(left.ConditionalValues, right.ConditionalValues)
	for _, condition := range condKeys {
		lv, hasLeft := left.ConditionalValues[condition]
		rv, hasRight := right.ConditionalValues[condition]
		label := fmt.Sprintf("cond %-15s", condition+":")
		switch {
		case !hasLeft && hasRight:
			lines = append(lines, fmt.Sprintf("      + %s %s", label, colorAdded(rcdisplay.FormatDiff(rv))))
		case hasLeft && !hasRight:
			lines = append(lines, fmt.Sprintf("      - %s %s", label, colorRemoved(rcdisplay.FormatDiff(lv))))
		case hasLeft && hasRight && !reflect.DeepEqual(lv, rv):
			lines = append(lines, fmt.Sprintf("      ~ %s %s → %s", label, colorRemoved(rcdisplay.FormatDiff(lv)), colorAdded(rcdisplay.FormatDiff(rv))))
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
	strfold.Sort(keys)
	return keys
}

func sortedConditionalNames(values map[string]firebase.RemoteConfigValue) []string {
	return strfold.SortedKeys(values)
}
