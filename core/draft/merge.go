package draft

import (
	"encoding/json"
	"fmt"
	"reflect"

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
	rcmutate.RemoveEmptyGroups(merged)
	rcmutate.DropUnknownConditionReferences(merged)

	if reflect.DeepEqual(latestCfg, merged) {
		return nil, false, nil
	}
	raw, err := firebase.MarshalRemoteConfig(merged)
	if err != nil {
		return nil, false, err
	}
	return raw, true, nil
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
