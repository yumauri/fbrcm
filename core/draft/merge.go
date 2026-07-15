package draft

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/strfold"
)

func MergeWithLatest(baseRaw, draftRaw, latestRaw json.RawMessage) (json.RawMessage, bool, error) {
	baseCfg, err := firebase.ParseRemoteConfig(baseRaw)
	if err != nil {
		return nil, false, fmt.Errorf("decode base remote config: %w", err)
	}
	draftCfg, err := firebase.ParseRemoteConfig(draftRaw)
	if err != nil {
		return nil, false, fmt.Errorf("decode draft remote config: %w", err)
	}
	latestCfg, err := firebase.ParseRemoteConfig(latestRaw)
	if err != nil {
		return nil, false, fmt.Errorf("decode latest remote config: %w", err)
	}

	merged, err := firebase.CloneRemoteConfig(latestCfg)
	if err != nil {
		return nil, false, fmt.Errorf("clone latest remote config: %w", err)
	}
	baseSlots := rcmutate.CollectParamSlots(baseCfg)
	draftSlots := rcmutate.CollectParamSlots(draftCfg)
	latestSlots := rcmutate.CollectParamSlots(latestCfg)

	for _, key := range sortedSlotKeys(baseSlots, draftSlots, latestSlots) {
		baseSlot, inBase := baseSlots[key]
		draftSlot, inDraft := draftSlots[key]
		latestSlot, inLatest := latestSlots[key]

		localChanged := !equalParamState(baseSlot, inBase, draftSlot, inDraft)
		if !localChanged {
			continue
		}
		remoteChanged := !equalParamState(baseSlot, inBase, latestSlot, inLatest)
		if !remoteChanged {
			applyMergedSlot(merged, key, baseSlot, inBase, draftSlot, inDraft)
			continue
		}
		if equalParamState(draftSlot, inDraft, latestSlot, inLatest) {
			continue
		}
		return nil, false, fmt.Errorf("draft conflict on %s", rcmutate.SlotDisplayKey(key))
	}
	if err := mergeGroupDescriptions(baseCfg, draftCfg, latestCfg, merged); err != nil {
		return nil, false, err
	}
	if err := mergeConditions(baseCfg, draftCfg, latestCfg, merged); err != nil {
		return nil, false, err
	}
	rcmutate.DropUnknownConditionReferences(merged)
	rcmutate.NormalizeEmptyParameterMaps(merged)

	if reflect.DeepEqual(latestCfg, merged) {
		return nil, false, nil
	}
	raw, err := firebase.MarshalRemoteConfig(merged)
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
}

type groupDescriptionState struct {
	Description string
	Present     bool
}

func mergeGroupDescriptions(base, draft, latest, merged *firebase.RemoteConfig) error {
	keys := make(map[string]struct{})
	for name := range base.ParameterGroups {
		keys[name] = struct{}{}
	}
	for name := range draft.ParameterGroups {
		keys[name] = struct{}{}
	}
	for name := range latest.ParameterGroups {
		keys[name] = struct{}{}
	}
	names := make([]string, 0, len(keys))
	for name := range keys {
		names = append(names, name)
	}
	strfold.Sort(names)
	for _, name := range names {
		baseState := groupDescription(base, name)
		draftState := groupDescription(draft, name)
		latestState := groupDescription(latest, name)
		if reflect.DeepEqual(baseState, draftState) {
			continue
		}
		if !reflect.DeepEqual(baseState, latestState) && !reflect.DeepEqual(draftState, latestState) {
			return fmt.Errorf("draft conflict on group description %s", name)
		}
		if !draftState.Present {
			delete(merged.ParameterGroups, name)
			continue
		}
		group, ok := merged.ParameterGroups[name]
		if !ok {
			if merged.ParameterGroups == nil {
				merged.ParameterGroups = make(map[string]firebase.RemoteConfigGroup)
			}
			group = firebase.RemoteConfigGroup{}
		}
		group.Description = draftState.Description
		merged.ParameterGroups[name] = group
	}
	return nil
}

func groupDescription(cfg *firebase.RemoteConfig, name string) groupDescriptionState {
	group, ok := cfg.ParameterGroups[name]
	return groupDescriptionState{Description: group.Description, Present: ok}
}

func mergeConditions(base, draft, latest, merged *firebase.RemoteConfig) error {
	baseByName := conditionsByName(base.Conditions)
	draftByName := conditionsByName(draft.Conditions)
	latestByName := conditionsByName(latest.Conditions)
	keys := make(map[string]struct{})
	for name := range baseByName {
		keys[name] = struct{}{}
	}
	for name := range draftByName {
		keys[name] = struct{}{}
	}
	for name := range latestByName {
		keys[name] = struct{}{}
	}
	names := make([]string, 0, len(keys))
	for name := range keys {
		names = append(names, name)
	}
	strfold.Sort(names)
	mergedByName := conditionsByName(merged.Conditions)
	for _, name := range names {
		baseCondition, inBase := baseByName[name]
		draftCondition, inDraft := draftByName[name]
		latestCondition, inLatest := latestByName[name]
		localChanged := inBase != inDraft || !reflect.DeepEqual(baseCondition, draftCondition)
		if !localChanged {
			continue
		}
		remoteChanged := inBase != inLatest || !reflect.DeepEqual(baseCondition, latestCondition)
		if remoteChanged && (inDraft != inLatest || !reflect.DeepEqual(draftCondition, latestCondition)) {
			return fmt.Errorf("draft conflict on condition %s", name)
		}
		if inDraft {
			mergedByName[name] = draftCondition
		} else {
			delete(mergedByName, name)
		}
	}

	baseOrder := conditionOrder(base.Conditions)
	draftOrder := conditionOrder(draft.Conditions)
	latestOrder := conditionOrder(latest.Conditions)
	localOrderChanged := !slices.Equal(baseOrder, draftOrder)
	remoteOrderChanged := !slices.Equal(baseOrder, latestOrder)
	if localOrderChanged && remoteOrderChanged && !slices.Equal(draftOrder, latestOrder) {
		return fmt.Errorf("draft conflict on condition order")
	}
	order := latestOrder
	if localOrderChanged {
		order = draftOrder
	}
	merged.Conditions = make([]firebase.RemoteConfigCondition, 0, len(mergedByName))
	seen := make(map[string]bool, len(mergedByName))
	for _, name := range order {
		if condition, ok := mergedByName[name]; ok {
			merged.Conditions = append(merged.Conditions, condition)
			seen[name] = true
		}
	}
	remaining := make([]string, 0)
	for name := range mergedByName {
		if !seen[name] {
			remaining = append(remaining, name)
		}
	}
	strfold.Sort(remaining)
	for _, name := range remaining {
		merged.Conditions = append(merged.Conditions, mergedByName[name])
	}
	if len(merged.Conditions) == 0 {
		merged.Conditions = nil
	}
	return nil
}

func conditionsByName(conditions []firebase.RemoteConfigCondition) map[string]firebase.RemoteConfigCondition {
	out := make(map[string]firebase.RemoteConfigCondition, len(conditions))
	for _, condition := range conditions {
		out[condition.Name] = condition
	}
	return out
}

func conditionOrder(conditions []firebase.RemoteConfigCondition) []string {
	out := make([]string, 0, len(conditions))
	for _, condition := range conditions {
		out = append(out, condition.Name)
	}
	return out
}

func sortedSlotKeys(maps ...map[string]rcmutate.Slot) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, items := range maps {
		for key := range items {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, key)
		}
	}
	strfold.Sort(out)
	return out
}

func equalParamState(left rcmutate.Slot, leftOK bool, right rcmutate.Slot, rightOK bool) bool {
	if leftOK != rightOK {
		return false
	}
	if !leftOK {
		return true
	}
	return reflect.DeepEqual(left, right)
}

func applyMergedSlot(cfg *firebase.RemoteConfig, key string, baseSlot rcmutate.Slot, inBase bool, draftSlot rcmutate.Slot, inDraft bool) {
	paramKey := rcmutate.SlotKeyParam(key)
	if !inDraft {
		group := rcmutate.SlotKeyGroup(key)
		if inBase {
			group = baseSlot.Group
		}
		rcmutate.RemoveParamSlot(cfg, paramKey, group)
		return
	}
	if inBase {
		rcmutate.RemoveParamSlot(cfg, paramKey, baseSlot.Group)
	}
	rcmutate.SetParamSlot(cfg, paramKey, draftSlot)
}
