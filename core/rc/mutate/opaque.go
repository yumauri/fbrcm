package mutate

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
)

type opaqueValueSlot struct {
	Group     string
	Parameter string
	Condition string
}

type opaqueParameterSlot struct {
	Group     string
	Parameter string
}

// EnsureOpaqueValuesUnchanged rejects mutations that add, remove, replace, or
// relocate Firebase-managed and unknown future Remote Config value options.
// These union members are read-only to fbrcm and must be managed by Firebase.
func EnsureOpaqueValuesUnchanged(current, final *firebase.RemoteConfig) error {
	currentValues := collectOpaqueValues(current)
	finalValues := collectOpaqueValues(final)
	keys := make([]opaqueValueSlot, 0, len(currentValues)+len(finalValues))
	seen := make(map[opaqueValueSlot]struct{}, len(currentValues)+len(finalValues))
	for key := range currentValues {
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	for key := range finalValues {
		if _, ok := seen[key]; !ok {
			keys = append(keys, key)
		}
	}
	slices.SortFunc(keys, compareOpaqueValueSlots)

	for _, key := range keys {
		before, beforeOK := currentValues[key]
		after, afterOK := finalValues[key]
		if beforeOK && afterOK && reflect.DeepEqual(before, after) {
			continue
		}
		value := before
		if !beforeOK {
			value = after
		}
		return fmt.Errorf("%s is read-only and must be managed by Firebase", opaqueValueLabel(key, value))
	}
	currentTypes := collectOpaqueParameterTypes(current)
	typeKeys := make([]opaqueParameterSlot, 0, len(currentTypes))
	for key := range currentTypes {
		typeKeys = append(typeKeys, key)
	}
	slices.SortFunc(typeKeys, compareOpaqueParameterSlots)
	for _, key := range typeKeys {
		beforeType := currentTypes[key]
		afterType, ok := parameterTypeAt(final, key)
		if ok && beforeType != afterType {
			return fmt.Errorf(
				"parameter %q%s value type is read-only because it contains a Firebase-managed or unknown value",
				key.Parameter,
				opaqueGroupSuffix(key.Group),
			)
		}
	}
	return nil
}

func collectOpaqueValues(cfg *firebase.RemoteConfig) map[opaqueValueSlot]firebase.RemoteConfigValue {
	values := make(map[opaqueValueSlot]firebase.RemoteConfigValue)
	if cfg == nil {
		return values
	}
	for key, param := range cfg.Parameters {
		collectOpaqueParameterValues(values, "", key, param)
	}
	for groupName, group := range cfg.ParameterGroups {
		for key, param := range group.Parameters {
			collectOpaqueParameterValues(values, groupName, key, param)
		}
	}
	return values
}

func collectOpaqueParameterValues(values map[opaqueValueSlot]firebase.RemoteConfigValue, group, key string, param firebase.RemoteConfigParam) {
	if param.DefaultValue != nil && param.DefaultValue.IsOpaque() {
		values[opaqueValueSlot{Group: group, Parameter: key}] = *param.DefaultValue
	}
	for condition, value := range param.ConditionalValues {
		if value.IsOpaque() {
			values[opaqueValueSlot{Group: group, Parameter: key, Condition: condition}] = value
		}
	}
}

func collectOpaqueParameterTypes(cfg *firebase.RemoteConfig) map[opaqueParameterSlot]string {
	types := make(map[opaqueParameterSlot]string)
	if cfg == nil {
		return types
	}
	for key, param := range cfg.Parameters {
		if remoteConfigParamHasOpaqueValue(param) {
			types[opaqueParameterSlot{Parameter: key}] = param.ValueType
		}
	}
	for groupName, group := range cfg.ParameterGroups {
		for key, param := range group.Parameters {
			if remoteConfigParamHasOpaqueValue(param) {
				types[opaqueParameterSlot{Group: groupName, Parameter: key}] = param.ValueType
			}
		}
	}
	return types
}

func remoteConfigParamHasOpaqueValue(param firebase.RemoteConfigParam) bool {
	if param.DefaultValue != nil && param.DefaultValue.IsOpaque() {
		return true
	}
	for _, value := range param.ConditionalValues {
		if value.IsOpaque() {
			return true
		}
	}
	return false
}

func parameterTypeAt(cfg *firebase.RemoteConfig, key opaqueParameterSlot) (string, bool) {
	if cfg == nil {
		return "", false
	}
	if key.Group == "" {
		param, ok := cfg.Parameters[key.Parameter]
		return param.ValueType, ok
	}
	group, ok := cfg.ParameterGroups[key.Group]
	if !ok {
		return "", false
	}
	param, ok := group.Parameters[key.Parameter]
	return param.ValueType, ok
}

func compareOpaqueValueSlots(left, right opaqueValueSlot) int {
	for _, pair := range [][2]string{
		{left.Group, right.Group},
		{left.Parameter, right.Parameter},
		{left.Condition, right.Condition},
	} {
		if result := strings.Compare(pair[0], pair[1]); result != 0 {
			return result
		}
	}
	return 0
}

func compareOpaqueParameterSlots(left, right opaqueParameterSlot) int {
	if result := strings.Compare(left.Group, right.Group); result != 0 {
		return result
	}
	return strings.Compare(left.Parameter, right.Parameter)
}

func opaqueValueLabel(slot opaqueValueSlot, value firebase.RemoteConfigValue) string {
	location := fmt.Sprintf("parameter %q", slot.Parameter)
	if slot.Group != "" {
		location = fmt.Sprintf("parameter %q in group %q", slot.Parameter, slot.Group)
	}
	valueSlot := "default value"
	if slot.Condition != "" {
		valueSlot = fmt.Sprintf("value for condition %q", slot.Condition)
	}
	return fmt.Sprintf("%s %s %s", location, valueSlot, opaqueValueKind(value))
}

func opaqueValueKind(value firebase.RemoteConfigValue) string {
	switch {
	case len(value.PersonalizationValue) > 0:
		return "personalization value"
	case len(value.ExperimentValue) > 0:
		return "A/B test value"
	case len(value.RolloutValue) > 0:
		return "rollout value"
	case value.UnknownValueOption != "":
		return fmt.Sprintf("unknown value option %q", value.UnknownValueOption)
	default:
		return "opaque value"
	}
}

func opaqueGroupSuffix(group string) string {
	if group == "" {
		return ""
	}
	return fmt.Sprintf(" in group %q", group)
}
