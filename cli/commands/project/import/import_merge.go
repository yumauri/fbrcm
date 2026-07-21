package importpkg

import (
	"fmt"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/importer"
)

func mergeRemoteConfigs(cmd *cobra.Command, currentCfg, importCfg *firebase.RemoteConfig, opts importOptions) (*firebase.RemoteConfig, error) {
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

func resolveConflict(cmd *cobra.Command, opts importOptions, label string, currentValue, importValue any) (conflictResolution, error) {
	if opts.mergeResolve != "" {
		return conflictResolution(opts.mergeResolve), nil
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
	prompt.Input = cmd.InOrStdin()
	prompt.Output = cmd.ErrOrStderr()
	choice, err := prompt.RunPrompt()
	if err != nil {
		return "", err
	}
	return conflictResolution(choice.value), nil
}
