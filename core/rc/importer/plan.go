package importer

import (
	"fmt"
	"reflect"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/strfold"
)

func BuildPlan(projectID, projectName string, current *firebase.RemoteConfig, source *ParsedSource, opts Options) (*Plan, error) {
	if source == nil || source.Config == nil {
		return nil, fmt.Errorf("import source is empty")
	}
	imported, err := firebase.CloneRemoteConfig(source.Config)
	if err != nil {
		return nil, err
	}
	if err := Transform(projectID, projectName, imported, opts); err != nil {
		return nil, err
	}
	if current == nil || !ConfigHasContent(current) || opts.Strategy == StrategyReplace {
		final, cloneErr := firebase.CloneRemoteConfig(imported)
		if cloneErr != nil {
			return nil, cloneErr
		}
		Cleanup(final)
		return &Plan{Imported: imported, Final: final, Summary: Summarize(imported, source.WrappedCache)}, nil
	}
	final, conflicts, err := merge(current, imported, opts)
	if err != nil {
		return nil, err
	}
	Cleanup(final)
	return &Plan{Imported: imported, Final: final, Conflicts: conflicts, Summary: Summarize(imported, source.WrappedCache)}, nil
}

// MergeConfigs merges an already transformed import into current config.
func MergeConfigs(current, imported *firebase.RemoteConfig, opts Options) (*firebase.RemoteConfig, []Conflict, error) {
	return merge(current, imported, opts)
}

func merge(current, imported *firebase.RemoteConfig, opts Options) (*firebase.RemoteConfig, []Conflict, error) {
	final, err := firebase.CloneRemoteConfig(current)
	if err != nil {
		return nil, nil, err
	}
	if final.Parameters == nil {
		final.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	if final.ParameterGroups == nil {
		final.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}
	conflicts := make([]Conflict, 0)

	conditionIndex := make(map[string]int, len(final.Conditions))
	for index, condition := range final.Conditions {
		conditionIndex[condition.Name] = index
	}
	for _, condition := range imported.Conditions {
		index, ok := conditionIndex[condition.Name]
		if !ok {
			final.Conditions = append(final.Conditions, condition)
			conditionIndex[condition.Name] = len(final.Conditions) - 1
			continue
		}
		if reflect.DeepEqual(final.Conditions[index], condition) {
			continue
		}
		conflict := Conflict{ID: "condition:" + condition.Name, Kind: ConflictCondition, Label: "condition " + condition.Name, Current: final.Conditions[index], Import: condition}
		conflicts = append(conflicts, conflict)
		if resolutionFor(conflict.ID, opts) == ResolutionImport {
			final.Conditions[index] = condition
		}
	}

	for _, groupName := range strfold.SortedKeys(imported.ParameterGroups) {
		importGroup := imported.ParameterGroups[groupName]
		currentGroup, ok := final.ParameterGroups[groupName]
		if !ok {
			final.ParameterGroups[groupName] = firebase.RemoteConfigGroup{Description: importGroup.Description, Parameters: map[string]firebase.RemoteConfigParam{}}
			continue
		}
		if currentGroup.Description != importGroup.Description {
			conflict := Conflict{ID: "group:" + groupName, Kind: ConflictGroupDescription, Label: "group description " + groupName, Current: currentGroup.Description, Import: importGroup.Description}
			conflicts = append(conflicts, conflict)
			if resolutionFor(conflict.ID, opts) == ResolutionImport {
				currentGroup.Description = importGroup.Description
				final.ParameterGroups[groupName] = currentGroup
			}
		}
	}

	currentSlots := rcmutate.CollectParamSlots(final)
	importSlots := rcmutate.CollectParamSlots(imported)
	for _, key := range strfold.SortedKeys(importSlots) {
		importSlot := importSlots[key]
		currentSlot, ok := currentSlots[key]
		if !ok {
			rcmutate.SetParamSlot(final, rcmutate.SlotKeyParam(key), importSlot)
			currentSlots[key] = importSlot
			continue
		}
		if currentSlot.Group == importSlot.Group && reflect.DeepEqual(currentSlot.Param, importSlot.Param) {
			continue
		}
		conflict := Conflict{ID: "parameter:" + key, Kind: ConflictParameter, Label: "parameter " + rcmutate.SlotDisplayKey(key), Current: rcdiff.ParamSlotPreview{Group: currentSlot.Group, Param: currentSlot.Param}, Import: rcdiff.ParamSlotPreview{Group: importSlot.Group, Param: importSlot.Param}}
		conflicts = append(conflicts, conflict)
		if resolutionFor(conflict.ID, opts) == ResolutionImport {
			if currentSlot.Group != importSlot.Group {
				rcmutate.RemoveParamSlot(final, rcmutate.SlotKeyParam(key), currentSlot.Group)
			}
			rcmutate.SetParamSlot(final, rcmutate.SlotKeyParam(key), importSlot)
			currentSlots[key] = importSlot
		}
	}
	return final, conflicts, nil
}

func resolutionFor(id string, opts Options) Resolution {
	if value, ok := opts.Resolutions[id]; ok {
		return value
	}
	if opts.DefaultResolution == ResolutionImport {
		return ResolutionImport
	}
	return ResolutionCurrent
}
