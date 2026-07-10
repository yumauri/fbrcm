package promote

import (
	"fmt"
	"reflect"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/strfold"
)

type ItemID struct {
	Kind  rcdiff.ItemKind `json:"kind"`
	Name  string          `json:"name"`
	Group string          `json:"group,omitempty"`
}

type Item struct {
	ID       ItemID            `json:"id"`
	Kind     rcdiff.ItemKind   `json:"kind"`
	Change   rcdiff.ChangeKind `json:"change"`
	Label    string            `json:"label"`
	Required bool              `json:"required,omitempty"`
}

type Options struct {
	Prune bool
}

type Plan struct {
	Source *firebase.RemoteConfig
	Target *firebase.RemoteConfig
	Diff   rcdiff.Result
	Items  []Item
}

func BuildPlan(source, target *firebase.RemoteConfig, opts Options) Plan {
	result := rcdiff.CompareRemoteConfigs(target, source)
	items := make([]Item, 0, len(result.Parameters)+len(result.Conditions)+len(result.GroupDescriptions))

	for _, change := range result.Parameters {
		if !eligible(change.Kind, opts.Prune) {
			continue
		}
		items = append(items, Item{
			ID:     ItemID{Kind: rcdiff.ItemParameter, Name: change.Key, Group: change.Group},
			Kind:   rcdiff.ItemParameter,
			Change: change.Kind,
			Label:  formatParamLabel(change.Key, change.Group),
		})
	}
	for _, change := range result.GroupDescriptions {
		if !eligible(change.Kind, opts.Prune) {
			continue
		}
		items = append(items, Item{
			ID:     ItemID{Kind: rcdiff.ItemGroupDescription, Name: change.Group},
			Kind:   rcdiff.ItemGroupDescription,
			Change: change.Kind,
			Label:  "group " + change.Group,
		})
	}
	for _, change := range result.Conditions {
		if !eligible(change.Kind, opts.Prune) {
			continue
		}
		items = append(items, Item{
			ID:     ItemID{Kind: rcdiff.ItemCondition, Name: change.Name},
			Kind:   rcdiff.ItemCondition,
			Change: change.Kind,
			Label:  "condition " + change.Name,
		})
	}

	return Plan{Source: source, Target: target, Diff: result, Items: items}
}

func SelectAll(items []Item) map[ItemID]bool {
	selected := make(map[ItemID]bool, len(items))
	for _, item := range items {
		selected[item.ID] = true
	}
	return selected
}

func Apply(plan Plan, selected map[ItemID]bool, opts Options) (*firebase.RemoteConfig, []Item, error) {
	finalCfg, err := firebase.CloneRemoteConfig(plan.Target)
	if err != nil {
		return nil, nil, err
	}

	selected = withDependencies(plan, selected)
	applied := make([]Item, 0, len(selected))
	appliedIDs := make(map[ItemID]struct{}, len(selected))

	for _, item := range plan.Items {
		if !selected[item.ID] {
			continue
		}
		if err := applyItem(finalCfg, plan.Source, item, opts); err != nil {
			return nil, nil, err
		}
		applied = append(applied, item)
		appliedIDs[item.ID] = struct{}{}
	}
	for id := range selected {
		if id.Kind != rcdiff.ItemCondition {
			continue
		}
		if _, ok := appliedIDs[id]; ok {
			continue
		}
		item := Item{ID: id, Kind: rcdiff.ItemCondition, Change: conditionChangeKind(plan.Target, plan.Source, id.Name), Label: "condition " + id.Name, Required: true}
		if err := applyItem(finalCfg, plan.Source, item, opts); err != nil {
			return nil, nil, err
		}
		applied = append(applied, item)
	}
	for id := range selected {
		if id.Kind != rcdiff.ItemGroupDescription {
			continue
		}
		if _, ok := appliedIDs[id]; ok {
			continue
		}
		item := Item{ID: id, Kind: rcdiff.ItemGroupDescription, Change: groupDescriptionChangeKind(plan.Target, plan.Source, id.Name), Label: "group " + id.Name, Required: true}
		if err := applyItem(finalCfg, plan.Source, item, opts); err != nil {
			return nil, nil, err
		}
		applied = append(applied, item)
	}

	rcmutate.DropUnknownConditionReferences(finalCfg)
	rcmutate.RemoveEmptyGroups(finalCfg)
	return finalCfg, applied, nil
}

func withDependencies(plan Plan, selected map[ItemID]bool) map[ItemID]bool {
	out := make(map[ItemID]bool, len(selected))
	for id, ok := range selected {
		if ok {
			out[id] = true
		}
	}

	sourceConditions := conditionMap(plan.Source)
	targetConditions := conditionMap(plan.Target)
	for _, item := range plan.Items {
		if item.Kind != rcdiff.ItemParameter || !out[item.ID] {
			continue
		}
		if item.ID.Group != "" && groupDescriptionChangeKind(plan.Target, plan.Source, item.ID.Group) != rcdiff.ChangeUnchanged {
			out[ItemID{Kind: rcdiff.ItemGroupDescription, Name: item.ID.Group}] = true
		}
		param := sourceParam(plan.Source, item.ID.Name, item.ID.Group)
		for conditionName := range param.ConditionalValues {
			sourceCondition, ok := sourceConditions[conditionName]
			if !ok {
				continue
			}
			targetCondition, exists := targetConditions[conditionName]
			if exists && reflect.DeepEqual(sourceCondition, targetCondition) {
				continue
			}
			out[ItemID{Kind: rcdiff.ItemCondition, Name: conditionName}] = true
		}
	}
	return out
}

func applyItem(target, source *firebase.RemoteConfig, item Item, opts Options) error {
	switch item.Kind {
	case rcdiff.ItemParameter:
		return applyParameter(target, source, item, opts)
	case rcdiff.ItemGroupDescription:
		return applyGroupDescription(target, source, item, opts)
	case rcdiff.ItemCondition:
		return applyCondition(target, source, item, opts)
	default:
		return fmt.Errorf("unknown promotion item kind %q", item.Kind)
	}
}

func applyParameter(target, source *firebase.RemoteConfig, item Item, opts Options) error {
	if item.Change == rcdiff.ChangeRemoved {
		if !opts.Prune {
			return nil
		}
		rcmutate.RemoveParamSlot(target, item.ID.Name, item.ID.Group)
		return nil
	}
	param := sourceParam(source, item.ID.Name, item.ID.Group)
	removeParamEverywhere(target, item.ID.Name)
	rcmutate.SetParamSlot(target, item.ID.Name, rcmutate.Slot{Group: item.ID.Group, Param: param})
	return nil
}

func applyGroupDescription(target, source *firebase.RemoteConfig, item Item, opts Options) error {
	if target.ParameterGroups == nil {
		target.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}
	if item.Change == rcdiff.ChangeRemoved {
		if !opts.Prune {
			return nil
		}
		group := target.ParameterGroups[item.ID.Name]
		group.Description = ""
		target.ParameterGroups[item.ID.Name] = group
		return nil
	}
	sourceGroup := source.ParameterGroups[item.ID.Name]
	targetGroup := target.ParameterGroups[item.ID.Name]
	targetGroup.Description = sourceGroup.Description
	target.ParameterGroups[item.ID.Name] = targetGroup
	return nil
}

func applyCondition(target, source *firebase.RemoteConfig, item Item, opts Options) error {
	conditions := conditionMap(target)
	sourceConditions := conditionMap(source)
	if item.Change == rcdiff.ChangeRemoved {
		if !opts.Prune {
			return nil
		}
		delete(conditions, item.ID.Name)
		target.Conditions = sortedConditions(conditions)
		return nil
	}
	sourceCondition, ok := sourceConditions[item.ID.Name]
	if !ok {
		return nil
	}
	conditions[item.ID.Name] = sourceCondition
	target.Conditions = sortedConditions(conditions)
	return nil
}

func sourceParam(source *firebase.RemoteConfig, key, group string) firebase.RemoteConfigParam {
	if group == "" {
		return source.Parameters[key]
	}
	return source.ParameterGroups[group].Parameters[key]
}

func removeParamEverywhere(cfg *firebase.RemoteConfig, key string) {
	rcmutate.RemoveParamSlot(cfg, key, "")
	for groupName := range cfg.ParameterGroups {
		rcmutate.RemoveParamSlot(cfg, key, groupName)
	}
}

func conditionMap(cfg *firebase.RemoteConfig) map[string]firebase.RemoteConfigCondition {
	out := make(map[string]firebase.RemoteConfigCondition, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		out[condition.Name] = condition
	}
	return out
}

func sortedConditions(values map[string]firebase.RemoteConfigCondition) []firebase.RemoteConfigCondition {
	keys := strfold.SortedKeys(values)
	out := make([]firebase.RemoteConfigCondition, 0, len(keys))
	for _, key := range keys {
		out = append(out, values[key])
	}
	return out
}

func eligible(kind rcdiff.ChangeKind, prune bool) bool {
	switch kind {
	case rcdiff.ChangeAdded, rcdiff.ChangeChanged:
		return true
	case rcdiff.ChangeRemoved:
		return prune
	default:
		return false
	}
}

func conditionChangeKind(target, source *firebase.RemoteConfig, name string) rcdiff.ChangeKind {
	targetConditions := conditionMap(target)
	sourceConditions := conditionMap(source)
	sourceCondition, sourceOK := sourceConditions[name]
	targetCondition, targetOK := targetConditions[name]
	switch {
	case sourceOK && !targetOK:
		return rcdiff.ChangeAdded
	case !sourceOK && targetOK:
		return rcdiff.ChangeRemoved
	case sourceOK && targetOK && !reflect.DeepEqual(sourceCondition, targetCondition):
		return rcdiff.ChangeChanged
	default:
		return rcdiff.ChangeUnchanged
	}
}

func groupDescriptionChangeKind(target, source *firebase.RemoteConfig, name string) rcdiff.ChangeKind {
	sourceGroup, sourceOK := source.ParameterGroups[name]
	targetGroup, targetOK := target.ParameterGroups[name]
	switch {
	case sourceOK && !targetOK:
		return rcdiff.ChangeAdded
	case !sourceOK && targetOK:
		return rcdiff.ChangeRemoved
	case sourceOK && targetOK && sourceGroup.Description != targetGroup.Description:
		return rcdiff.ChangeChanged
	default:
		return rcdiff.ChangeUnchanged
	}
}

func formatParamLabel(key, group string) string {
	if group == "" {
		return key
	}
	return group + "/" + key
}
