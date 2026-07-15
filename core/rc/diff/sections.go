package diff

import (
	"fmt"
	"strings"
)

func renderConditionsDiff(result Result) (string, diffCounts) {
	var lines []string
	var counts diffCounts
	for _, change := range result.Conditions {
		switch change.Kind {
		case ChangeAdded:
			counts.added++
			lines = append(lines, fmt.Sprintf("  + %-15s %s", colorAdded(change.Name), formatConditionSummary(*change.Final)))
		case ChangeRemoved:
			counts.removed++
			lines = append(lines, fmt.Sprintf("  - %-15s %s", colorRemoved(change.Name), formatConditionSummary(*change.Current)))
		case ChangeUnchanged:
			counts.unchanged++
		case ChangeChanged:
			counts.changed++
			currentSummary := formatConditionSummary(*change.Current)
			finalSummary := formatConditionSummary(*change.Final)
			detail := fmt.Sprintf("%s → %s", colorRemoved(currentSummary), colorAdded(finalSummary))
			if currentSummary == finalSummary && change.PreviousPosition != change.FinalPosition {
				detail = fmt.Sprintf("position %s → %s", colorRemoved(fmt.Sprint(change.PreviousPosition)), colorAdded(fmt.Sprint(change.FinalPosition)))
			} else if change.PreviousPosition != change.FinalPosition {
				detail += fmt.Sprintf("; position %s → %s", colorRemoved(fmt.Sprint(change.PreviousPosition)), colorAdded(fmt.Sprint(change.FinalPosition)))
			}
			lines = append(lines, fmt.Sprintf("  ~ %-15s %s", colorChanged(change.Name), detail))
		}
	}

	if len(lines) == 0 {
		return "", counts
	}
	return "Conditions:\n" + strings.Join(lines, "\n"), counts
}

func renderGroupDescriptionsDiff(result Result) (string, diffCounts) {
	var lines []string
	var counts diffCounts
	for _, change := range result.GroupDescriptions {
		switch change.Kind {
		case ChangeAdded:
			counts.added++
			lines = append(lines, fmt.Sprintf("  + %-15s %s", colorAdded(formatGroupValue(change.Group)), change.Final))
		case ChangeRemoved:
			counts.removed++
			lines = append(lines, fmt.Sprintf("  - %-15s %s", colorRemoved(formatGroupValue(change.Group)), change.Current))
		case ChangeUnchanged:
			counts.unchanged++
		case ChangeChanged:
			counts.changed++
			lines = append(lines, fmt.Sprintf("  ~ %-15s %s → %s", colorChanged(formatGroupValue(change.Group)), colorRemoved(change.Current), colorAdded(change.Final)))
		}
	}
	if len(lines) == 0 {
		return "", counts
	}
	return "Group descriptions:\n" + strings.Join(lines, "\n"), counts
}

func renderParametersDiff(result Result) (string, diffCounts) {
	var lines []string
	var counts diffCounts
	for _, change := range result.Parameters {
		switch change.Kind {
		case ChangeAdded:
			counts.added++
			lines = append(lines, renderAddedParameter(change.Key, paramView{Group: change.Group, Param: *change.Final})...)
		case ChangeRemoved:
			counts.removed++
			lines = append(lines, renderRemovedParameter(change.Key, paramView{Group: change.Group, Param: *change.Current})...)
		case ChangeUnchanged:
			counts.unchanged++
		case ChangeChanged:
			counts.changed++
			leftGroup := change.Group
			if change.PreviousGroup != "" || change.PreviousKey != "" {
				leftGroup = change.PreviousGroup
			}
			lines = append(lines, renderChangedParameter(change.Key, paramView{Group: leftGroup, Param: *change.Current}, paramView{Group: change.Group, Param: *change.Final})...)
		}
	}
	if len(lines) == 0 {
		return "", counts
	}
	return "Parameters:\n" + strings.Join(lines, "\n"), counts
}
