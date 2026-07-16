package draft

import (
	"fmt"
	"strings"

	coreconditions "github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	rcvalue "github.com/yumauri/fbrcm/core/rc/value"
)

func renameParamSlot(cfg *firebase.RemoteConfig, key, nextKey, groupName string) error {
	nextKey = strings.TrimSpace(nextKey)
	if nextKey == "" {
		return fmt.Errorf("parameter name is empty")
	}
	if key == nextKey {
		return fmt.Errorf("parameter not changed")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	if lookupAnyParamSlot(cfg, nextKey) {
		return fmt.Errorf("parameter %q already exists", nextKey)
	}
	rcmutate.RemoveParamSlot(cfg, key, groupName)
	rcmutate.SetParamSlot(cfg, nextKey, slot)
	return nil
}

func moveParamSlot(cfg *firebase.RemoteConfig, key, currentGroup, nextGroup string) error {
	nextGroup = strings.TrimSpace(nextGroup)
	if currentGroup == nextGroup {
		return fmt.Errorf("parameter already in group %q", nextGroup)
	}
	slot, ok := lookupParamSlot(cfg, key, currentGroup)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	rcmutate.RemoveParamSlot(cfg, key, currentGroup)
	slot.Group = nextGroup
	rcmutate.SetParamSlot(cfg, key, slot)
	return nil
}

func applyParameterDetailsEdit(cfg *firebase.RemoteConfig, edit ParameterDetailsEdit) error {
	if edit.Create {
		return createParameterDetailsSlot(cfg, edit)
	}
	return editParameterDetailsSlot(cfg, edit)
}

func createParameterDetailsSlot(cfg *firebase.RemoteConfig, edit ParameterDetailsEdit) error {
	nextGroup := NormalizeGroupKey(edit.NextGroupKey)
	nextKey := strings.TrimSpace(edit.NextParamKey)
	if nextKey == "" {
		return fmt.Errorf("parameter name is empty")
	}
	if lookupAnyParamSlot(cfg, nextKey) {
		return fmt.Errorf("parameter %q already exists", nextKey)
	}
	param := firebase.RemoteConfigParam{
		Description:  strings.TrimSpace(edit.NextDescription),
		ValueType:    normalizeParameterValueType(edit.NextValueType),
		DefaultValue: &firebase.RemoteConfigValue{Value: ""},
	}
	slot := rcmutate.Slot{Group: nextGroup, Param: param}
	for _, valueEdit := range edit.ValueEdits {
		if err := setRawParamValue(cfg, &slot.Param, valueEdit.Label, valueEdit.NextValue, slot.Param.ValueType); err != nil {
			return err
		}
	}
	rcmutate.SetParamSlot(cfg, nextKey, slot)
	return nil
}

func editParameterDetailsSlot(cfg *firebase.RemoteConfig, edit ParameterDetailsEdit) error {
	currentGroup := NormalizeGroupKey(edit.GroupKey)
	nextGroup := NormalizeGroupKey(edit.NextGroupKey)
	nextKey := strings.TrimSpace(edit.NextParamKey)
	if nextKey == "" {
		return fmt.Errorf("parameter name is empty")
	}
	slot, ok := lookupParamSlot(cfg, edit.ParamKey, currentGroup)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	if nextKey != edit.ParamKey {
		if lookupAnyParamSlot(cfg, nextKey) {
			return fmt.Errorf("parameter %q already exists", nextKey)
		}
	}

	slot.Param.Description = strings.TrimSpace(edit.NextDescription)
	slot.Param.ValueType = normalizeParameterValueType(edit.NextValueType)
	for _, valueEdit := range edit.ValueEdits {
		if err := setRawParamValue(cfg, &slot.Param, valueEdit.Label, valueEdit.NextValue, slot.Param.ValueType); err != nil {
			return err
		}
	}
	slot.Group = nextGroup
	rcmutate.RemoveParamSlot(cfg, edit.ParamKey, currentGroup)
	rcmutate.SetParamSlot(cfg, nextKey, slot)
	return nil
}

func setRawParamValue(cfg *firebase.RemoteConfig, param *firebase.RemoteConfigParam, valueLabel, nextValue, valueType string) error {
	if err := rcvalue.ValidateRawValueForType(nextValue, valueType); err != nil {
		return err
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("value editor supports only plain values")
		}
		value.Value = nextValue
		return value, nil
	}
	if valueLabel == "default" {
		if param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*param.DefaultValue)
		if err != nil {
			return err
		}
		param.DefaultValue = &next
		return nil
	}
	canonicalLabel, ok := coreconditions.ResolveName(cfg, valueLabel)
	if !ok {
		return fmt.Errorf("condition %q not found", valueLabel)
	}
	if param.ConditionalValues == nil {
		param.ConditionalValues = make(map[string]firebase.RemoteConfigValue)
	}
	current := param.ConditionalValues[canonicalLabel]
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	param.ConditionalValues[canonicalLabel] = next
	return nil
}

func normalizeParameterValueType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "string":
		return "STRING"
	case "boolean", "bool":
		return "BOOLEAN"
	case "number":
		return "NUMBER"
	case "json":
		return "JSON"
	default:
		return strings.TrimSpace(value)
	}
}

func lookupParamSlot(cfg *firebase.RemoteConfig, key, groupName string) (rcmutate.Slot, bool) {
	if groupName == "" {
		param, ok := cfg.Parameters[key]
		return rcmutate.Slot{Group: "", Param: param}, ok
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return rcmutate.Slot{}, false
	}
	param, ok := group.Parameters[key]
	return rcmutate.Slot{Group: groupName, Param: param}, ok
}

func lookupAnyParamSlot(cfg *firebase.RemoteConfig, key string) bool {
	if _, ok := cfg.Parameters[key]; ok {
		return true
	}
	for _, group := range cfg.ParameterGroups {
		if _, ok := group.Parameters[key]; ok {
			return true
		}
	}
	return false
}
