package shared

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
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
			colorRemoved(FormatPlainValue(current)),
			colorAdded(FormatPlainValue(incoming)),
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

func RenderConflictChoiceValue(value any) string {
	switch v := value.(type) {
	case ParamSlotPreview:
		return summarizeParamSlotChoice(v)
	case firebase.RemoteConfigCondition:
		return trimPreview(formatConditionSummary(v))
	case string:
		return trimPreview(FormatPlainValue(v))
	default:
		body, err := json.Marshal(value)
		if err != nil {
			return trimPreview(fmt.Sprintf("%v", value))
		}
		return trimPreview(string(NormalizeExportJSON(body)))
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

func formatParameterHeader(key, group string) string {
	if group == "" {
		return key
	}
	return fmt.Sprintf("%s [%s]", key, group)
}

func summarizeParamSlotChoice(slot ParamSlotPreview) string {
	parts := make([]string, 0, 4)
	if slot.Group != "" {
		parts = append(parts, "group="+formatGroupValue(slot.Group))
	}
	if slot.Param.DefaultValue != nil {
		parts = append(parts, "default="+FormatRemoteValue(*slot.Param.DefaultValue))
	}
	for _, condition := range sortedConditionalNames(slot.Param.ConditionalValues) {
		parts = append(parts, condition+"="+FormatRemoteValue(slot.Param.ConditionalValues[condition]))
	}
	if strings.TrimSpace(slot.Param.ValueType) != "" {
		parts = append(parts, "type="+slot.Param.ValueType)
	}
	if strings.TrimSpace(slot.Param.Description) != "" {
		parts = append(parts, "desc="+FormatPlainValue(slot.Param.Description))
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
