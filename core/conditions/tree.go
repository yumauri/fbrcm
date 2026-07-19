package conditions

import (
	"fmt"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/parameters"
	"github.com/yumauri/fbrcm/core/rc/display"
	"github.com/yumauri/fbrcm/core/rootgroup"
	"github.com/yumauri/fbrcm/core/strfold"
)

func BuildTree(remoteConfig *firebase.RemoteConfig, cachedAt time.Time, etag string) *Tree {
	tree := &Tree{CachedAt: cachedAt, ETag: etag}
	if remoteConfig == nil {
		return tree
	}

	tree.Version = remoteConfig.Version.VersionNumber
	tree.Conditions = make([]Entry, len(remoteConfig.Conditions))
	byName := make(map[string]int, len(remoteConfig.Conditions))
	for i, condition := range remoteConfig.Conditions {
		tree.Conditions[i] = Entry{
			Priority:   i + 1,
			Name:       condition.Name,
			Expression: condition.Expression,
			TagColor:   condition.TagColor,
			Usages:     []Usage{},
		}
		byName[condition.Name] = i
	}

	tree.parameters = buildParameters(remoteConfig, tree.Conditions, byName)
	return tree
}

func buildParameters(remoteConfig *firebase.RemoteConfig, conditions []Entry, byName map[string]int) []parameterState {
	seen := make(map[string]struct{})
	groupKeys := strfold.SortedKeys(remoteConfig.ParameterGroups)
	for _, groupKey := range groupKeys {
		for parameterKey := range remoteConfig.ParameterGroups[groupKey].Parameters {
			seen[parameterKey] = struct{}{}
		}
	}

	states := make([]parameterState, 0, len(remoteConfig.Parameters)+len(seen))
	rootKeys := strfold.SortedKeys(remoteConfig.Parameters)
	for _, parameterKey := range rootKeys {
		if _, grouped := seen[parameterKey]; grouped {
			continue
		}
		states = append(states, addParameterUsages(
			conditions,
			byName,
			rootgroup.TreeKey,
			rootgroup.Label,
			parameterKey,
			remoteConfig.Parameters[parameterKey],
		))
	}

	for _, groupKey := range groupKeys {
		group := remoteConfig.ParameterGroups[groupKey]
		parameterKeys := strfold.SortedKeys(group.Parameters)
		for _, parameterKey := range parameterKeys {
			states = append(states, addParameterUsages(
				conditions,
				byName,
				groupKey,
				groupKey,
				parameterKey,
				group.Parameters[parameterKey],
			))
		}
	}
	return states
}

func addParameterUsages(conditions []Entry, byName map[string]int, groupKey, groupLabel, parameterKey string, parameter firebase.RemoteConfigParam) parameterState {
	ref := ParameterRef{GroupKey: groupKey, GroupLabel: groupLabel, ParameterKey: parameterKey}
	state := parameterState{
		ref:               ref,
		hasDefault:        parameter.DefaultValue != nil,
		conditionalValues: make(map[string]struct{}, len(parameter.ConditionalValues)),
	}
	valueType := strings.ToUpper(display.EmptyValueType(parameter.ValueType))
	for conditionName, value := range parameter.ConditionalValues {
		conditionIndex, known := byName[conditionName]
		if !known {
			continue
		}
		state.conditionalValues[conditionName] = struct{}{}
		conditions[conditionIndex].Usages = append(conditions[conditionIndex].Usages, Usage{
			GroupKey:     groupKey,
			GroupLabel:   groupLabel,
			ParameterKey: parameterKey,
			Value:        parameters.FormatRemoteConfigDisplayValue(value, parameter.ValueType),
			RawValue:     value.Value,
			ValueType:    valueType,
			Plain:        !value.UseInAppDefault && len(value.PersonalizationValue) == 0 && len(value.RolloutValue) == 0,
		})
	}
	return state
}

func (t *Tree) Find(name string) (Entry, bool) {
	if t == nil {
		return Entry{}, false
	}
	for _, condition := range t.Conditions {
		if condition.Name == name {
			return condition, true
		}
	}
	return Entry{}, false
}

func (t *Tree) DeleteImpact(name string) (DeleteImpact, error) {
	condition, ok := t.Find(name)
	if !ok {
		return DeleteImpact{}, fmt.Errorf("condition %q not found", name)
	}

	impact := DeleteImpact{Condition: condition, Usages: append([]Usage(nil), condition.Usages...)}
	for _, parameter := range t.parameters {
		if _, usesCondition := parameter.conditionalValues[name]; !usesCondition {
			continue
		}
		if !parameter.hasDefault && len(parameter.conditionalValues) == 1 {
			impact.RemovedParameters = append(impact.RemovedParameters, parameter.ref)
		}
	}
	return impact, nil
}

func (t *Tree) MoveImpact(name string, toPriority int) (MoveImpact, error) {
	condition, ok := t.Find(name)
	if !ok {
		return MoveImpact{}, fmt.Errorf("condition %q not found", name)
	}
	if toPriority < 1 || toPriority > len(t.Conditions) {
		return MoveImpact{}, fmt.Errorf("condition priority must be between 1 and %d", len(t.Conditions))
	}

	impact := MoveImpact{
		Condition:    condition,
		FromPriority: condition.Priority,
		ToPriority:   toPriority,
	}
	if toPriority == condition.Priority {
		return impact, nil
	}

	from := condition.Priority - 1
	to := toPriority - 1
	if to < from {
		for i := to; i < from; i++ {
			impact.CrossedConditions = append(impact.CrossedConditions, t.Conditions[i].Name)
		}
	} else {
		for i := from + 1; i <= to; i++ {
			impact.CrossedConditions = append(impact.CrossedConditions, t.Conditions[i].Name)
		}
	}

	crossed := make(map[string]struct{}, len(impact.CrossedConditions))
	for _, crossedName := range impact.CrossedConditions {
		crossed[crossedName] = struct{}{}
	}
	for _, parameter := range t.parameters {
		if _, usesMoved := parameter.conditionalValues[name]; !usesMoved {
			continue
		}
		for conditionalName := range parameter.conditionalValues {
			if _, usesCrossed := crossed[conditionalName]; usesCrossed {
				impact.AffectedParameters = append(impact.AffectedParameters, parameter.ref)
				break
			}
		}
	}
	return impact, nil
}
