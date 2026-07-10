package diff

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

type diffCounts struct {
	added     int
	removed   int
	changed   int
	unchanged int
}

// ParamSlotPreview identifies a parameter slot for import conflict previews.
type ParamSlotPreview struct {
	Group string
	Param firebase.RemoteConfigParam
}

type paramView struct {
	Group string
	Param firebase.RemoteConfigParam
}

// RenderRemoteConfigDiff renders a human-readable diff between two Remote Config
// snapshots. The returned bool is true when any condition or parameter changed.
func RenderRemoteConfigDiff(currentCfg, finalCfg *firebase.RemoteConfig) (string, bool) {
	return RenderResult(CompareRemoteConfigs(currentCfg, finalCfg))
}

func RenderResult(result Result) (string, bool) {
	var sections []string

	conditionsText, conditionCounts := renderConditionsDiff(result)
	groupsText, groupCounts := renderGroupDescriptionsDiff(result)
	paramsText, parameterCounts := renderParametersDiff(result)

	if conditionsText != "" {
		sections = append(sections, conditionsText)
	}
	if groupsText != "" {
		sections = append(sections, groupsText)
	}
	if paramsText != "" {
		sections = append(sections, paramsText)
	}

	hasChanges := conditionCounts.added+conditionCounts.removed+conditionCounts.changed+
		groupCounts.added+groupCounts.removed+groupCounts.changed+
		parameterCounts.added+parameterCounts.removed+parameterCounts.changed > 0
	if !hasChanges {
		return "", false
	}

	summary := fmt.Sprintf(
		"Summary:\n  %s condition added, %s removed, %s changed, %s unchanged\n  %s group description added, %s removed, %s changed, %s unchanged\n  %s parameter added, %s removed, %s changed, %s unchanged",
		formatCount(conditionCounts.added),
		formatCount(conditionCounts.removed),
		formatCount(conditionCounts.changed),
		formatCount(conditionCounts.unchanged),
		formatCount(groupCounts.added),
		formatCount(groupCounts.removed),
		formatCount(groupCounts.changed),
		formatCount(groupCounts.unchanged),
		formatCount(parameterCounts.added),
		formatCount(parameterCounts.removed),
		formatCount(parameterCounts.changed),
		formatCount(parameterCounts.unchanged),
	)
	sections = append(sections, summary)
	return "\n" + strings.Join(sections, "\n\n") + "\n", true
}

// RenderConflictPreview renders a single import conflict line for interactive merge.
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
			colorRemoved(rcdisplay.FormatPlainValue(current)),
			colorAdded(rcdisplay.FormatPlainValue(incoming)),
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
	return fmt.Sprintf("current:\n%s\nimport:\n%s", string(normalizeJSON(currentJSON)), string(normalizeJSON(importJSON)))
}

// RenderConflictChoiceValue summarizes a conflict choice for prompt labels.
func RenderConflictChoiceValue(value any) string {
	switch v := value.(type) {
	case ParamSlotPreview:
		return summarizeParamSlotChoice(v)
	case firebase.RemoteConfigCondition:
		return trimPreview(formatConditionSummary(v))
	case string:
		return trimPreview(rcdisplay.FormatPlainValue(v))
	default:
		body, err := json.Marshal(value)
		if err != nil {
			return trimPreview(fmt.Sprintf("%v", value))
		}
		return trimPreview(string(normalizeJSON(body)))
	}
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

func summarizeParamSlotChoice(slot ParamSlotPreview) string {
	parts := make([]string, 0, 4)
	if slot.Group != "" {
		parts = append(parts, "group="+formatGroupValue(slot.Group))
	}
	if slot.Param.DefaultValue != nil {
		parts = append(parts, "default="+rcdisplay.FormatDiff(*slot.Param.DefaultValue))
	}
	for _, condition := range sortedConditionalNames(slot.Param.ConditionalValues) {
		parts = append(parts, condition+"="+rcdisplay.FormatDiff(slot.Param.ConditionalValues[condition]))
	}
	if strings.TrimSpace(slot.Param.ValueType) != "" {
		parts = append(parts, "type="+slot.Param.ValueType)
	}
	if strings.TrimSpace(slot.Param.Description) != "" {
		parts = append(parts, "desc="+rcdisplay.FormatPlainValue(slot.Param.Description))
	}
	if len(parts) == 0 {
		return "(empty)"
	}
	return trimPreview(strings.Join(parts, " | "))
}

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

func formatCount(count int) string {
	return fmt.Sprintf("%d", count)
}
