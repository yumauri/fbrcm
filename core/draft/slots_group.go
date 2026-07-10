package draft

import (
	"fmt"
	"maps"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
)

func removeGroupSlot(cfg *firebase.RemoteConfig, groupName string) error {
	if groupName == "" {
		return fmt.Errorf("default group cannot be removed")
	}
	if _, ok := cfg.ParameterGroups[groupName]; !ok {
		return fmt.Errorf("group not found")
	}
	delete(cfg.ParameterGroups, groupName)
	return nil
}

func renameGroupSlot(cfg *firebase.RemoteConfig, key, nextKey string) error {
	nextKey = strings.TrimSpace(nextKey)
	if key == "" {
		return fmt.Errorf("default group cannot be renamed")
	}
	if nextKey == "" {
		return fmt.Errorf("group name is empty")
	}
	if key == nextKey {
		return fmt.Errorf("group not changed")
	}
	group, ok := cfg.ParameterGroups[key]
	if !ok {
		return fmt.Errorf("group not found")
	}
	if _, exists := cfg.ParameterGroups[nextKey]; exists {
		return fmt.Errorf("group %q already exists", nextKey)
	}
	delete(cfg.ParameterGroups, key)
	cfg.ParameterGroups[nextKey] = group
	return nil
}

func moveGroupSlot(cfg *firebase.RemoteConfig, currentGroup, nextGroup string) error {
	if currentGroup == nextGroup {
		return fmt.Errorf("group already moved to %q", nextGroup)
	}
	if currentGroup == "" {
		if nextGroup == "" {
			return fmt.Errorf("default group cannot be moved to default group")
		}
		destGroup := cfg.ParameterGroups[nextGroup]
		for key := range cfg.Parameters {
			if _, exists := destGroup.Parameters[key]; exists {
				return fmt.Errorf("parameter %q already exists", key)
			}
		}
		rootParams := make(map[string]firebase.RemoteConfigParam, len(cfg.Parameters))
		maps.Copy(rootParams, cfg.Parameters)
		for key, param := range rootParams {
			rcmutate.RemoveParamSlot(cfg, key, "")
			rcmutate.SetParamSlot(cfg, key, rcmutate.Slot{Group: nextGroup, Param: param})
		}
		return nil
	}
	group, ok := cfg.ParameterGroups[currentGroup]
	if !ok {
		return fmt.Errorf("group not found")
	}
	for key := range group.Parameters {
		if nextGroup == "" {
			if _, exists := cfg.Parameters[key]; exists {
				return fmt.Errorf("parameter %q already exists", key)
			}
			continue
		}
		destGroup := cfg.ParameterGroups[nextGroup]
		if _, exists := destGroup.Parameters[key]; exists {
			return fmt.Errorf("parameter %q already exists", key)
		}
	}
	for key, param := range group.Parameters {
		rcmutate.RemoveParamSlot(cfg, key, currentGroup)
		rcmutate.SetParamSlot(cfg, key, rcmutate.Slot{Group: nextGroup, Param: param})
	}
	delete(cfg.ParameterGroups, currentGroup)
	if nextGroup != "" {
		if group, ok := cfg.ParameterGroups[nextGroup]; ok {
			cfg.ParameterGroups[nextGroup] = group
		}
	}
	return nil
}
