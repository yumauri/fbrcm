package project

import (
	"fmt"
	"reflect"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/firebase"
)

func mergeRemoteConfigs(cmd *cobra.Command, currentCfg, importCfg *firebase.RemoteConfig, opts importOptions) (*firebase.RemoteConfig, error) {
	finalCfg := shared.CloneRemoteConfig(currentCfg)
	if finalCfg.Parameters == nil {
		finalCfg.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	if finalCfg.ParameterGroups == nil {
		finalCfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}

	conditionIndex := make(map[string]int, len(finalCfg.Conditions))
	for i, condition := range finalCfg.Conditions {
		conditionIndex[condition.Name] = i
	}
	for _, condition := range importCfg.Conditions {
		index, ok := conditionIndex[condition.Name]
		if !ok {
			finalCfg.Conditions = append(finalCfg.Conditions, condition)
			conditionIndex[condition.Name] = len(finalCfg.Conditions) - 1
			continue
		}
		if reflect.DeepEqual(finalCfg.Conditions[index], condition) {
			continue
		}
		resolution, err := resolveConflict(cmd, opts, "condition "+condition.Name, finalCfg.Conditions[index], condition)
		if err != nil {
			return nil, err
		}
		if resolution == conflictResolutionImport {
			finalCfg.Conditions[index] = condition
		}
	}

	for _, groupName := range shared.SortedStringKeys(importCfg.ParameterGroups) {
		importGroup := importCfg.ParameterGroups[groupName]
		currentGroup, ok := finalCfg.ParameterGroups[groupName]
		if !ok {
			finalCfg.ParameterGroups[groupName] = firebase.RemoteConfigGroup{
				Description: importGroup.Description,
				Parameters:  map[string]firebase.RemoteConfigParam{},
			}
			continue
		}
		if currentGroup.Description != importGroup.Description {
			resolution, err := resolveConflict(cmd, opts, "group description "+groupName, currentGroup.Description, importGroup.Description)
			if err != nil {
				return nil, err
			}
			if resolution == conflictResolutionImport {
				currentGroup.Description = importGroup.Description
				finalCfg.ParameterGroups[groupName] = currentGroup
			}
		}
	}

	currentSlots := collectParamSlots(finalCfg)
	importSlots := collectParamSlots(importCfg)
	for _, key := range shared.SortedStringKeys(importSlots) {
		importSlot := importSlots[key]
		currentSlot, ok := currentSlots[key]
		if !ok {
			setParamSlot(finalCfg, key, importSlot)
			currentSlots[key] = importSlot
			continue
		}
		if currentSlot.group == importSlot.group && reflect.DeepEqual(currentSlot.param, importSlot.param) {
			continue
		}

		resolution, err := resolveConflict(cmd, opts, "parameter "+key, currentSlot, importSlot)
		if err != nil {
			return nil, err
		}
		if resolution == conflictResolutionImport {
			if currentSlot.group != importSlot.group {
				removeParamSlot(finalCfg, key, currentSlot.group)
			}
			setParamSlot(finalCfg, key, importSlot)
			currentSlots[key] = importSlot
		}
	}

	return finalCfg, nil
}

func resolveConflict(cmd *cobra.Command, opts importOptions, label string, currentValue, importValue any) (conflictResolution, error) {
	if opts.mergeResolve != "" {
		return conflictResolution(opts.mergeResolve), nil
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nConflict: %s\n", label)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), shared.RenderConflictPreview(label, toSharedConflictValue(currentValue), toSharedConflictValue(importValue)))
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())

	prompt := selection.New("Choose value", []mergeChoice{
		{label: fmt.Sprintf("Use import value (%s)", shared.RenderConflictChoiceValue(toSharedConflictValue(importValue))), value: string(conflictResolutionImport)},
		{label: fmt.Sprintf("Keep current value (%s)", shared.RenderConflictChoiceValue(toSharedConflictValue(currentValue))), value: string(conflictResolutionCurrent)},
	})
	prompt.Template = `
{{- if .Prompt -}}
  {{ Bold .Prompt }}
{{ end -}}

{{- range  $i, $choice := .Choices }}
  {{- if IsScrollUpHintPosition $i }}
    {{- "⇡ " -}}
  {{- else if IsScrollDownHintPosition $i -}}
    {{- "⇣ " -}}
  {{- else -}}
    {{- "  " -}}
  {{- end -}}

  {{- if eq $.SelectedIndex $i }}
   {{- print (Foreground "32" (Bold "▸ ")) (Selected $choice) "\n" }}
  {{- else }}
    {{- print "  " (Unselected $choice) "\n" }}
  {{- end }}
{{- end}}`
	prompt.SelectedChoiceStyle = styleConflictSelectedChoice
	prompt.UnselectedChoiceStyle = styleConflictUnselectedChoice
	prompt.FinalChoiceStyle = styleConflictFinalChoice
	choice, err := prompt.RunPrompt()
	if err != nil {
		return "", err
	}
	return conflictResolution(choice.value), nil
}

func toSharedConflictValue(value any) any {
	slot, ok := value.(paramSlot)
	if !ok {
		return value
	}
	return shared.ParamSlotPreview{
		Group: slot.group,
		Param: slot.param,
	}
}

func collectParamSlots(cfg *firebase.RemoteConfig) map[string]paramSlot {
	out := make(map[string]paramSlot)
	for key, param := range cfg.Parameters {
		out[key] = paramSlot{param: param}
	}
	for groupName, group := range cfg.ParameterGroups {
		for key, param := range group.Parameters {
			out[key] = paramSlot{group: groupName, param: param}
		}
	}
	return out
}

func setParamSlot(cfg *firebase.RemoteConfig, key string, slot paramSlot) {
	if slot.group == "" {
		if cfg.Parameters == nil {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
		}
		cfg.Parameters[key] = slot.param
		return
	}

	group := cfg.ParameterGroups[slot.group]
	if group.Parameters == nil {
		group.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	group.Parameters[key] = slot.param
	cfg.ParameterGroups[slot.group] = group
}

func removeParamSlot(cfg *firebase.RemoteConfig, key, groupName string) {
	if groupName == "" {
		delete(cfg.Parameters, key)
		return
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return
	}
	delete(group.Parameters, key)
	if len(group.Parameters) == 0 {
		delete(cfg.ParameterGroups, groupName)
		return
	}
	cfg.ParameterGroups[groupName] = group
}
