package parameters

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/core/strfold"
)

const defaultGroupKey = rootgroup.TreeKey

func BuildTree(remoteConfig *firebase.RemoteConfig, cachedAt time.Time, etag string) *Tree {
	if remoteConfig == nil {
		return &Tree{
			CachedAt: cachedAt,
			ETag:     etag,
		}
	}

	return &Tree{
		Version:      remoteConfig.Version.VersionNumber,
		CachedAt:     cachedAt,
		ETag:         etag,
		Conditions:   buildConditions(remoteConfig),
		Groups:       buildGroups(remoteConfig),
		remoteConfig: remoteConfig,
	}
}

func buildConditions(remoteConfig *firebase.RemoteConfig) []Condition {
	conditions := make([]Condition, 0, len(remoteConfig.Conditions))
	for _, condition := range remoteConfig.Conditions {
		conditions = append(conditions, Condition{
			Name:       condition.Name,
			Expression: condition.Expression,
			Color:      condition.TagColor,
		})
	}
	return conditions
}

func buildGroups(remoteConfig *firebase.RemoteConfig) []Group {
	conditionColors := firebase.ConditionTagColorsByName(remoteConfig.Conditions)
	conditionOrder := make(map[string]int, len(remoteConfig.Conditions))
	for i, condition := range remoteConfig.Conditions {
		conditionOrder[condition.Name] = i
	}

	groupKeys := strfold.SortedKeys(remoteConfig.ParameterGroups)
	seen := make(map[string]struct{})
	groups := make([]Group, 0, len(groupKeys)+1)

	for _, groupKey := range groupKeys {
		group := remoteConfig.ParameterGroups[groupKey]
		params := buildEntries(group.Parameters, conditionColors, conditionOrder)
		for key := range group.Parameters {
			seen[key] = struct{}{}
		}
		groups = append(groups, Group{
			Key:         groupKey,
			Label:       groupKey,
			Description: strings.TrimSpace(group.Description),
			Parameters:  params,
		})
	}

	rootParams := make(map[string]firebase.RemoteConfigParam)
	for key, param := range remoteConfig.Parameters {
		if _, ok := seen[key]; ok {
			continue
		}
		rootParams[key] = param
	}
	if len(rootParams) > 0 {
		groups = append([]Group{{
			Key:        defaultGroupKey,
			Label:      rootgroup.Label,
			Parameters: buildEntries(rootParams, conditionColors, conditionOrder),
		}}, groups...)
	}

	return groups
}

func buildEntries(params map[string]firebase.RemoteConfigParam, conditionColors map[string]string, conditionOrder map[string]int) []Entry {
	keys := strfold.SortedKeys(params)
	out := make([]Entry, 0, len(keys))
	for _, key := range keys {
		param := params[key]
		values := make([]Value, 0, len(param.ConditionalValues)+1)
		conditionKeys := sortedConditionalKeys(param.ConditionalValues, conditionOrder)
		for _, condition := range conditionKeys {
			rawValue := param.ConditionalValues[condition]
			display := SummarizeRemoteConfigDisplayValue(rawValue, param.ValueType)
			values = append(values, Value{
				Label:           condition,
				Value:           display.Text,
				Display:         display,
				RawValue:        rawValue.Value,
				ValueType:       rcdisplay.EmptyValueType(param.ValueType),
				Empty:           isEmptyRemoteConfigValue(rawValue),
				EmptyType:       rcdisplay.EmptyValueType(param.ValueType),
				Color:           conditionColors[condition],
				Plain:           rawValue.IsPlain(),
				UseInAppDefault: rawValue.UseInAppDefault,
			})
		}
		if param.DefaultValue != nil {
			rawValue := *param.DefaultValue
			display := SummarizeRemoteConfigDisplayValue(rawValue, param.ValueType)
			values = append(values, Value{
				Label:           "default",
				Value:           display.Text,
				Display:         display,
				RawValue:        rawValue.Value,
				ValueType:       rcdisplay.EmptyValueType(param.ValueType),
				Empty:           isEmptyRemoteConfigValue(rawValue),
				EmptyType:       rcdisplay.EmptyValueType(param.ValueType),
				Plain:           rawValue.IsPlain(),
				UseInAppDefault: rawValue.UseInAppDefault,
			})
		}

		out = append(out, Entry{
			Key:         key,
			Description: strings.TrimSpace(param.Description),
			Summary:     summarizeParameterValues(values),
			Values:      values,
		})
	}
	return out
}

func summarizeParameterValues(values []Value) string {
	if len(values) == 0 {
		return "no values"
	}
	if len(values) == 1 {
		return values[0].Value
	}
	return fmt.Sprintf("%d values", len(values))
}

func isEmptyRemoteConfigValue(value firebase.RemoteConfigValue) bool {
	return value.IsPlain() && value.Value == ""
}

func sortedConditionalKeys(items map[string]firebase.RemoteConfigValue, order map[string]int) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	slices.SortFunc(keys, func(leftKey, rightKey string) int {
		left, leftOK := order[leftKey]
		right, rightOK := order[rightKey]
		switch {
		case leftOK && rightOK && left != right:
			if left < right {
				return -1
			}
			return 1
		case leftOK != rightOK:
			if leftOK {
				return -1
			}
			return 1
		default:
			return strfold.Compare(leftKey, rightKey)
		}
	})

	return keys
}

// OrderedConditionalKeys returns conditional value names in Remote Config
// evaluation order. Conditions absent from the config are placed afterward in
// stable case-folded order.
func OrderedConditionalKeys(items map[string]firebase.RemoteConfigValue, conditions []firebase.RemoteConfigCondition) []string {
	order := make(map[string]int, len(conditions))
	for index, condition := range conditions {
		order[condition.Name] = index
	}
	return sortedConditionalKeys(items, order)
}
