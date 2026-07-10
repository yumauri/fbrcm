package importpkg

import (
	"fmt"
	"reflect"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/strfold"
)

func mergeRemoteConfigs(cmd *cobra.Command, currentCfg, importCfg *firebase.RemoteConfig, opts importOptions) (*firebase.RemoteConfig, error) {
	finalCfg, err := firebase.CloneRemoteConfig(currentCfg)
	if err != nil {
		return nil, err
	}
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

	for _, groupName := range strfold.SortedKeys(importCfg.ParameterGroups) {
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

	currentSlots := rcmutate.CollectParamSlots(finalCfg)
	importSlots := rcmutate.CollectParamSlots(importCfg)
	for _, key := range strfold.SortedKeys(importSlots) {
		importSlot := importSlots[key]
		currentSlot, ok := currentSlots[key]
		if !ok {
			rcmutate.SetParamSlot(finalCfg, rcmutate.SlotKeyParam(key), importSlot)
			currentSlots[key] = importSlot
			continue
		}
		if currentSlot.Group == importSlot.Group && reflect.DeepEqual(currentSlot.Param, importSlot.Param) {
			continue
		}

		resolution, err := resolveConflict(cmd, opts, "parameter "+rcmutate.SlotDisplayKey(key), currentSlot, importSlot)
		if err != nil {
			return nil, err
		}
		if resolution == conflictResolutionImport {
			if currentSlot.Group != importSlot.Group {
				rcmutate.RemoveParamSlot(finalCfg, rcmutate.SlotKeyParam(key), currentSlot.Group)
			}
			rcmutate.SetParamSlot(finalCfg, rcmutate.SlotKeyParam(key), importSlot)
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
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), rc.RenderConflictPreview(label, toSharedConflictValue(currentValue), toSharedConflictValue(importValue)))
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())

	prompt := selection.New("Choose value", []mergeChoice{
		{label: fmt.Sprintf("Use import value (%s)", rc.RenderConflictChoiceValue(toSharedConflictValue(importValue))), value: string(conflictResolutionImport)},
		{label: fmt.Sprintf("Keep current value (%s)", rc.RenderConflictChoiceValue(toSharedConflictValue(currentValue))), value: string(conflictResolutionCurrent)},
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
	slot, ok := value.(rcmutate.Slot)
	if !ok {
		return value
	}
	return rc.ParamSlotPreview{
		Group: slot.Group,
		Param: slot.Param,
	}
}
