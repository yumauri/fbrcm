package importpkg

import (
	"fmt"

	"github.com/erikgeiser/promptkit/selection"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/importer"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/ops/shared/rc"
)

func mergeRemoteConfigs(cmd invocation.Call, currentCfg, importCfg *firebase.RemoteConfig, opts importOptions) (*firebase.RemoteConfig, error) {
	plannerOpts := opts.plannerOptions()
	finalCfg, conflicts, err := importer.MergeConfigs(currentCfg, importCfg, plannerOpts)
	if err != nil {
		return nil, err
	}
	if opts.mergeResolve != "" || len(conflicts) == 0 {
		return finalCfg, nil
	}
	plannerOpts.Resolutions = make(map[string]importer.Resolution, len(conflicts))
	for _, conflict := range conflicts {
		resolution, resolveErr := resolveConflict(cmd, opts, conflict.Label, conflict.Current, conflict.Import)
		if resolveErr != nil {
			return nil, resolveErr
		}
		plannerOpts.Resolutions[conflict.ID] = importer.Resolution(resolution)
	}
	finalCfg, _, err = importer.MergeConfigs(currentCfg, importCfg, plannerOpts)
	return finalCfg, err
}

func resolveConflict(cmd invocation.Call, opts importOptions, label string, currentValue, importValue any) (conflictResolution, error) {
	if opts.mergeResolve != "" {
		return conflictResolution(opts.mergeResolve), nil
	}
	if shared.MachineMode(cmd) {
		return "", shared.InteractionRequiredWithArguments("a merge conflict resolution is required; pass --merge-resolve=current or --merge-resolve=import", "selection_required", false, "--merge-resolve")
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nConflict: %s\n", label)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), rc.RenderConflictPreview(label, currentValue, importValue))
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())

	prompt := selection.New("Choose value", []mergeChoice{
		{label: fmt.Sprintf("Use import value (%s)", rc.RenderConflictChoiceValue(importValue)), value: string(conflictResolutionImport)},
		{label: fmt.Sprintf("Keep current value (%s)", rc.RenderConflictChoiceValue(currentValue)), value: string(conflictResolutionCurrent)},
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
	promptInput, closePromptInput, err := shared.OpenPromptInput(cmd.InOrStdin())
	if err != nil {
		return "", err
	}
	defer closePromptInput()
	prompt.Input = promptInput
	prompt.Output = cmd.ErrOrStderr()
	choice, err := prompt.RunPrompt()
	if err != nil {
		return "", err
	}
	return conflictResolution(choice.value), nil
}
