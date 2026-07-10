package diff

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
)

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
	strfold.Sort(keys)

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
	strfold.Sort(keys)

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
