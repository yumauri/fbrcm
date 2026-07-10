package draft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	rcvalue "github.com/yumauri/fbrcm/core/rc/value"
)

func setBooleanParamValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel string, nextValue bool) error {
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	target := "false"
	if nextValue {
		target = "true"
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("boolean editor supports only plain values")
		}
		if strings.EqualFold(value.Value, target) {
			return firebase.RemoteConfigValue{}, fmt.Errorf("parameter value not changed")
		}
		value.Value = target
		return value, nil
	}

	if valueLabel == "default" {
		if slot.Param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*slot.Param.DefaultValue)
		if err != nil {
			return err
		}
		slot.Param.DefaultValue = &next
		rcmutate.SetParamSlot(cfg, key, slot)
		return nil
	}

	if slot.Param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := slot.Param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	slot.Param.ConditionalValues[valueLabel] = next
	rcmutate.SetParamSlot(cfg, key, slot)
	return nil
}

func setNumberParamValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel, nextValue string) error {
	nextValue = strings.TrimSpace(nextValue)
	if !rcvalue.IsJSONNumber(nextValue) {
		return fmt.Errorf("invalid number")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("number editor supports only plain values")
		}
		if strings.TrimSpace(value.Value) == nextValue {
			return firebase.RemoteConfigValue{}, fmt.Errorf("parameter value not changed")
		}
		value.Value = nextValue
		return value, nil
	}

	if valueLabel == "default" {
		if slot.Param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*slot.Param.DefaultValue)
		if err != nil {
			return err
		}
		slot.Param.DefaultValue = &next
		rcmutate.SetParamSlot(cfg, key, slot)
		return nil
	}

	if slot.Param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := slot.Param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	slot.Param.ConditionalValues[valueLabel] = next
	rcmutate.SetParamSlot(cfg, key, slot)
	return nil
}

func setStringParamValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel, nextValue string) error {
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("string editor supports only plain values")
		}
		if value.Value == nextValue {
			return firebase.RemoteConfigValue{}, fmt.Errorf("parameter value not changed")
		}
		value.Value = nextValue
		return value, nil
	}

	if valueLabel == "default" {
		if slot.Param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*slot.Param.DefaultValue)
		if err != nil {
			return err
		}
		slot.Param.DefaultValue = &next
		rcmutate.SetParamSlot(cfg, key, slot)
		return nil
	}

	if slot.Param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := slot.Param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	slot.Param.ConditionalValues[valueLabel] = next
	rcmutate.SetParamSlot(cfg, key, slot)
	return nil
}

func setJSONParamValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel, nextValue string) error {
	nextValue = strings.TrimSpace(nextValue)
	if !json.Valid([]byte(nextValue)) {
		return fmt.Errorf("invalid json")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(nextValue)); err != nil {
		return fmt.Errorf("invalid json")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	updateValue := func(value firebase.RemoteConfigValue) (firebase.RemoteConfigValue, error) {
		if value.UseInAppDefault || len(value.PersonalizationValue) > 0 || len(value.RolloutValue) > 0 {
			return firebase.RemoteConfigValue{}, fmt.Errorf("json editor supports only plain values")
		}
		if value.Value == compact.String() {
			return firebase.RemoteConfigValue{}, fmt.Errorf("parameter value not changed")
		}
		value.Value = compact.String()
		return value, nil
	}

	if valueLabel == "default" {
		if slot.Param.DefaultValue == nil {
			return fmt.Errorf("default value not found")
		}
		next, err := updateValue(*slot.Param.DefaultValue)
		if err != nil {
			return err
		}
		slot.Param.DefaultValue = &next
		rcmutate.SetParamSlot(cfg, key, slot)
		return nil
	}

	if slot.Param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	current, ok := slot.Param.ConditionalValues[valueLabel]
	if !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	next, err := updateValue(current)
	if err != nil {
		return err
	}
	slot.Param.ConditionalValues[valueLabel] = next
	rcmutate.SetParamSlot(cfg, key, slot)
	return nil
}

func deleteConditionalValueSlot(cfg *firebase.RemoteConfig, key, groupName, valueLabel string) error {
	if valueLabel == "default" || strings.TrimSpace(valueLabel) == "" {
		return fmt.Errorf("conditional value not found")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	if slot.Param.ConditionalValues == nil {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	if _, ok := slot.Param.ConditionalValues[valueLabel]; !ok {
		return fmt.Errorf("conditional value %q not found", valueLabel)
	}
	delete(slot.Param.ConditionalValues, valueLabel)
	if len(slot.Param.ConditionalValues) == 0 {
		slot.Param.ConditionalValues = nil
	}
	rcmutate.SetParamSlot(cfg, key, slot)
	return nil
}

func duplicateParamSlot(cfg *firebase.RemoteConfig, key, groupName string) (string, error) {
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return "", fmt.Errorf("parameter not found")
	}
	nextKey := nextDuplicateParamKey(cfg, key+"_copy")
	rcmutate.SetParamSlot(cfg, nextKey, slot)
	return nextKey, nil
}

func duplicateParamSlotAs(cfg *firebase.RemoteConfig, key, nextKey, groupName string) error {
	nextKey = strings.TrimSpace(nextKey)
	if nextKey == "" {
		return fmt.Errorf("invalid name")
	}
	slot, ok := lookupParamSlot(cfg, key, groupName)
	if !ok {
		return fmt.Errorf("parameter not found")
	}
	if lookupAnyParamSlot(cfg, nextKey) {
		return fmt.Errorf("parameter %q already exists", nextKey)
	}
	rcmutate.SetParamSlot(cfg, nextKey, slot)
	return nil
}

func nextDuplicateParamKey(cfg *firebase.RemoteConfig, base string) string {
	if !lookupAnyParamSlot(cfg, base) {
		return base
	}
	for i := 2; ; i++ {
		next := fmt.Sprintf("%s__dup__%d", base, i)
		if !lookupAnyParamSlot(cfg, next) {
			return next
		}
	}
}
