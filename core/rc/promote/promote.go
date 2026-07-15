package promote

import (
	"fmt"
	"reflect"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
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
	applySelectedConditionOrder(finalCfg, plan.Source, selected)

	rcmutate.DropUnknownConditionReferences(finalCfg)
	removeExplicitlyPrunedGroups(finalCfg, plan.Source, selected)
	rcmutate.NormalizeEmptyParameterMaps(finalCfg)
	return finalCfg, applied, nil
}

func removeExplicitlyPrunedGroups(cfg, source *firebase.RemoteConfig, selected map[ItemID]bool) {
	for groupName, group := range cfg.ParameterGroups {
		if len(group.Parameters) > 0 {
			continue
		}
		if _, exists := source.ParameterGroups[groupName]; exists {
			continue
		}
		id := ItemID{Kind: rcdiff.ItemGroupDescription, Name: groupName}
		if selected[id] {
			delete(cfg.ParameterGroups, groupName)
		}
	}
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
	sourceConditions := conditionMap(source)
	if item.Change == rcdiff.ChangeRemoved {
		if !opts.Prune {
			return nil
		}
		removeCondition(target, item.ID.Name)
		return nil
	}
	sourceCondition, ok := sourceConditions[item.ID.Name]
	if !ok {
		return nil
	}
	for i := range target.Conditions {
		if target.Conditions[i].Name == item.ID.Name {
			target.Conditions[i] = sourceCondition
			return nil
		}
	}
	target.Conditions = append(target.Conditions, sourceCondition)
	return nil
}

func removeCondition(cfg *firebase.RemoteConfig, name string) {
	for i := range cfg.Conditions {
		if cfg.Conditions[i].Name != name {
			continue
		}
		cfg.Conditions = append(cfg.Conditions[:i], cfg.Conditions[i+1:]...)
		return
	}
}

// applySelectedConditionOrder moves only selected source conditions into their
// source-relative positions. Unselected and target-only conditions retain their
// relative order, while a fully selected promotion adopts the source order.
func applySelectedConditionOrder(target, source *firebase.RemoteConfig, selected map[ItemID]bool) {
	sourcePosition := make(map[string]int, len(source.Conditions))
	selectedSource := make(map[string]struct{})
	for i, condition := range source.Conditions {
		sourcePosition[condition.Name] = i
		if selected[ItemID{Kind: rcdiff.ItemCondition, Name: condition.Name}] {
			selectedSource[condition.Name] = struct{}{}
		}
	}
	if len(selectedSource) == 0 {
		return
	}

	remaining := make([]firebase.RemoteConfigCondition, 0, len(target.Conditions))
	selectedConditions := make(map[string]firebase.RemoteConfigCondition, len(selectedSource))
	for _, condition := range target.Conditions {
		if _, selected := selectedSource[condition.Name]; selected {
			selectedConditions[condition.Name] = condition
			continue
		}
		remaining = append(remaining, condition)
	}

	for _, sourceCondition := range source.Conditions {
		condition, selected := selectedConditions[sourceCondition.Name]
		if !selected {
			continue
		}
		remaining = insertConditionBySourceOrder(remaining, condition, sourcePosition)
	}
	target.Conditions = remaining
}

func insertConditionBySourceOrder(conditions []firebase.RemoteConfigCondition, condition firebase.RemoteConfigCondition, sourcePosition map[string]int) []firebase.RemoteConfigCondition {
	position := sourcePosition[condition.Name]
	afterIndex := -1
	afterPosition := len(sourcePosition) + 1
	beforeIndex := -1
	beforePosition := -1
	for i, current := range conditions {
		currentPosition, inSource := sourcePosition[current.Name]
		if !inSource {
			continue
		}
		if currentPosition > position && currentPosition < afterPosition {
			afterIndex = i
			afterPosition = currentPosition
		}
		if currentPosition < position && currentPosition > beforePosition {
			beforeIndex = i
			beforePosition = currentPosition
		}
	}

	insertAt := len(conditions)
	if afterIndex >= 0 {
		insertAt = afterIndex
	} else if beforeIndex >= 0 {
		insertAt = beforeIndex + 1
	}
	conditions = append(conditions, firebase.RemoteConfigCondition{})
	copy(conditions[insertAt+1:], conditions[insertAt:])
	conditions[insertAt] = condition
	return conditions
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
