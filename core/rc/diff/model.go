package diff

import (
	"reflect"

	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/strfold"
)

type ChangeKind string

const (
	ChangeAdded     ChangeKind = "added"
	ChangeRemoved   ChangeKind = "removed"
	ChangeChanged   ChangeKind = "changed"
	ChangeUnchanged ChangeKind = "unchanged"
)

type ItemKind string

const (
	ItemCondition        ItemKind = "condition"
	ItemParameter        ItemKind = "parameter"
	ItemGroupDescription ItemKind = "group_description"
)

type Summary struct {
	Added     int `json:"added"`
	Removed   int `json:"removed"`
	Changed   int `json:"changed"`
	Unchanged int `json:"unchanged"`
}

type Result struct {
	Conditions        []ConditionChange        `json:"conditions"`
	Parameters        []ParameterChange        `json:"parameters"`
	GroupDescriptions []GroupDescriptionChange `json:"group_descriptions"`
}

type ConditionChange struct {
	Name             string                          `json:"name"`
	Kind             ChangeKind                      `json:"kind"`
	PreviousPosition int                             `json:"previous_position,omitempty"`
	FinalPosition    int                             `json:"final_position,omitempty"`
	Current          *firebase.RemoteConfigCondition `json:"current,omitempty"`
	Final            *firebase.RemoteConfigCondition `json:"final,omitempty"`
}

type ParameterChange struct {
	Key           string                      `json:"key"`
	Group         string                      `json:"group,omitempty"`
	PreviousKey   string                      `json:"previous_key,omitempty"`
	PreviousGroup string                      `json:"previous_group,omitempty"`
	Kind          ChangeKind                  `json:"kind"`
	Current       *firebase.RemoteConfigParam `json:"current,omitempty"`
	Final         *firebase.RemoteConfigParam `json:"final,omitempty"`
}

type GroupDescriptionChange struct {
	Group   string     `json:"group"`
	Kind    ChangeKind `json:"kind"`
	Current string     `json:"current,omitempty"`
	Final   string     `json:"final,omitempty"`
}

func CompareRemoteConfigs(currentCfg, finalCfg *firebase.RemoteConfig) Result {
	return Result{
		Conditions:        compareConditions(currentCfg, finalCfg),
		GroupDescriptions: compareGroupDescriptions(currentCfg, finalCfg),
		Parameters:        compareParameters(currentCfg, finalCfg),
	}
}

func (r Result) HasChanges() bool {
	summary := r.TotalSummary()
	return summary.Added+summary.Removed+summary.Changed > 0
}

func (r Result) ConditionSummary() Summary {
	return summarizeConditions(r.Conditions)
}

func (r Result) ParameterSummary() Summary {
	return summarizeParameters(r.Parameters)
}

func (r Result) GroupDescriptionSummary() Summary {
	return summarizeGroupDescriptions(r.GroupDescriptions)
}

func (r Result) TotalSummary() Summary {
	out := r.ConditionSummary()
	out = addSummary(out, r.ParameterSummary())
	out = addSummary(out, r.GroupDescriptionSummary())
	return out
}

func compareConditions(currentCfg, finalCfg *firebase.RemoteConfig) []ConditionChange {
	current := make(map[string]firebase.RemoteConfigCondition, len(currentCfg.Conditions))
	final := make(map[string]firebase.RemoteConfigCondition, len(finalCfg.Conditions))
	currentPosition := make(map[string]int, len(currentCfg.Conditions))
	finalPosition := make(map[string]int, len(finalCfg.Conditions))
	keys := make([]string, 0, len(currentCfg.Conditions)+len(finalCfg.Conditions))
	seen := make(map[string]struct{})

	for index, condition := range currentCfg.Conditions {
		current[condition.Name] = condition
		currentPosition[condition.Name] = index + 1
		if _, ok := seen[condition.Name]; !ok {
			keys = append(keys, condition.Name)
			seen[condition.Name] = struct{}{}
		}
	}
	for index, condition := range finalCfg.Conditions {
		final[condition.Name] = condition
		finalPosition[condition.Name] = index + 1
		if _, ok := seen[condition.Name]; !ok {
			keys = append(keys, condition.Name)
			seen[condition.Name] = struct{}{}
		}
	}
	strfold.Sort(keys)

	changes := make([]ConditionChange, 0, len(keys))
	for _, key := range keys {
		left, hasLeft := current[key]
		right, hasRight := final[key]
		switch {
		case !hasLeft && hasRight:
			changes = append(changes, ConditionChange{Name: key, Kind: ChangeAdded, FinalPosition: finalPosition[key], Final: &right})
		case hasLeft && !hasRight:
			changes = append(changes, ConditionChange{Name: key, Kind: ChangeRemoved, PreviousPosition: currentPosition[key], Current: &left})
		case reflect.DeepEqual(left, right) && currentPosition[key] == finalPosition[key]:
			changes = append(changes, ConditionChange{Name: key, Kind: ChangeUnchanged, PreviousPosition: currentPosition[key], FinalPosition: finalPosition[key], Current: &left, Final: &right})
		default:
			changes = append(changes, ConditionChange{Name: key, Kind: ChangeChanged, PreviousPosition: currentPosition[key], FinalPosition: finalPosition[key], Current: &left, Final: &right})
		}
	}
	return changes
}

func compareGroupDescriptions(currentCfg, finalCfg *firebase.RemoteConfig) []GroupDescriptionChange {
	keys := make([]string, 0, len(currentCfg.ParameterGroups)+len(finalCfg.ParameterGroups))
	seen := make(map[string]struct{})
	for key := range currentCfg.ParameterGroups {
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	for key := range finalCfg.ParameterGroups {
		if _, ok := seen[key]; ok {
			continue
		}
		keys = append(keys, key)
	}
	strfold.Sort(keys)

	changes := make([]GroupDescriptionChange, 0, len(keys))
	for _, key := range keys {
		left, hasLeft := currentCfg.ParameterGroups[key]
		right, hasRight := finalCfg.ParameterGroups[key]
		switch {
		case !hasLeft && hasRight:
			changes = append(changes, GroupDescriptionChange{Group: key, Kind: ChangeAdded, Final: right.Description})
		case hasLeft && !hasRight:
			changes = append(changes, GroupDescriptionChange{Group: key, Kind: ChangeRemoved, Current: left.Description})
		case left.Description == right.Description:
			changes = append(changes, GroupDescriptionChange{Group: key, Kind: ChangeUnchanged, Current: left.Description, Final: right.Description})
		default:
			changes = append(changes, GroupDescriptionChange{Group: key, Kind: ChangeChanged, Current: left.Description, Final: right.Description})
		}
	}
	return changes
}

func compareParameters(currentCfg, finalCfg *firebase.RemoteConfig) []ParameterChange {
	current := rcmutate.CollectParamSlots(currentCfg)
	final := rcmutate.CollectParamSlots(finalCfg)
	moved := detectMovedParamSlots(current, final)
	keys := make([]string, 0, len(current)+len(final))
	seen := make(map[string]struct{})
	for key := range current {
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	for key := range final {
		if _, ok := seen[key]; ok {
			continue
		}
		keys = append(keys, key)
	}
	strfold.Sort(keys)

	changes := make([]ParameterChange, 0, len(keys))
	skip := make(map[string]struct{})
	for _, destination := range moved {
		skip[destination] = struct{}{}
	}
	for _, slotKey := range keys {
		if _, ok := skip[slotKey]; ok {
			continue
		}
		if nextKey, ok := moved[slotKey]; ok {
			left := current[slotKey]
			right := final[nextKey]
			key := rcmutate.SlotKeyParam(nextKey)
			changes = append(changes, ParameterChange{
				Key:           key,
				Group:         right.Group,
				PreviousKey:   rcmutate.SlotKeyParam(slotKey),
				PreviousGroup: left.Group,
				Kind:          ChangeChanged,
				Current:       &left.Param,
				Final:         &right.Param,
			})
			skip[nextKey] = struct{}{}
			continue
		}
		left, hasLeft := current[slotKey]
		right, hasRight := final[slotKey]
		key := rcmutate.SlotKeyParam(slotKey)
		group := rcmutate.SlotKeyGroup(slotKey)
		switch {
		case !hasLeft && hasRight:
			changes = append(changes, ParameterChange{Key: key, Group: group, Kind: ChangeAdded, Final: &right.Param})
		case hasLeft && !hasRight:
			changes = append(changes, ParameterChange{Key: key, Group: group, Kind: ChangeRemoved, Current: &left.Param})
		case reflect.DeepEqual(left.Param, right.Param):
			changes = append(changes, ParameterChange{Key: key, Group: group, Kind: ChangeUnchanged, Current: &left.Param, Final: &right.Param})
		default:
			changes = append(changes, ParameterChange{Key: key, Group: group, Kind: ChangeChanged, Current: &left.Param, Final: &right.Param})
		}
	}
	return changes
}

func detectMovedParamSlots(current, final map[string]rcmutate.Slot) map[string]string {
	removedByParam := make(map[string][]string)
	addedByParam := make(map[string][]string)
	for slotKey := range current {
		if _, ok := final[slotKey]; !ok {
			paramKey := rcmutate.SlotKeyParam(slotKey)
			removedByParam[paramKey] = append(removedByParam[paramKey], slotKey)
		}
	}
	for slotKey := range final {
		if _, ok := current[slotKey]; !ok {
			paramKey := rcmutate.SlotKeyParam(slotKey)
			addedByParam[paramKey] = append(addedByParam[paramKey], slotKey)
		}
	}

	out := make(map[string]string)
	for paramKey, removed := range removedByParam {
		added := addedByParam[paramKey]
		if len(removed) != 1 || len(added) != 1 {
			continue
		}
		out[removed[0]] = added[0]
	}
	return out
}

func summarizeConditions(changes []ConditionChange) Summary {
	var summary Summary
	for _, change := range changes {
		incrementSummary(&summary, change.Kind)
	}
	return summary
}

func summarizeParameters(changes []ParameterChange) Summary {
	var summary Summary
	for _, change := range changes {
		incrementSummary(&summary, change.Kind)
	}
	return summary
}

func summarizeGroupDescriptions(changes []GroupDescriptionChange) Summary {
	var summary Summary
	for _, change := range changes {
		incrementSummary(&summary, change.Kind)
	}
	return summary
}

func incrementSummary(summary *Summary, kind ChangeKind) {
	switch kind {
	case ChangeAdded:
		summary.Added++
	case ChangeRemoved:
		summary.Removed++
	case ChangeChanged:
		summary.Changed++
	case ChangeUnchanged:
		summary.Unchanged++
	}
}

func addSummary(left, right Summary) Summary {
	return Summary{
		Added:     left.Added + right.Added,
		Removed:   left.Removed + right.Removed,
		Changed:   left.Changed + right.Changed,
		Unchanged: left.Unchanged + right.Unchanged,
	}
}
