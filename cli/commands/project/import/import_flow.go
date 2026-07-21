package importpkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/importer"
)

type importResult struct {
	ProjectID string `json:"project_id"`
	Status    string `json:"status"`
	Changed   bool   `json:"changed"`
	Draft     bool   `json:"draft"`
	DryRun    bool   `json:"dry_run"`
}

// Run executes the project import command pipeline.
func Run(cmd *cobra.Command, svc *core.Core, project core.Project) error {
	opts, err := readImportOptions(cmd)
	if err != nil {
		return err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	draftMode, err := cmd.Flags().GetBool("draft")
	if err != nil {
		return err
	}
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	result := importResult{ProjectID: project.ProjectID, Draft: draftMode, DryRun: dryRun}
	ctx := context.Background()
	if dryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	if !draftMode {
		hasDraft, draftErr := svc.HasDraft(project.ProjectID)
		if draftErr != nil {
			return draftErr
		}
		if hasDraft {
			return fmt.Errorf("project %s has an unpublished draft; use --draft, publish it, or discard it first", project.ProjectID)
		}
	}

	raw, err := readRemoteConfig(cmd)
	if err != nil {
		return err
	}
	if raw == nil {
		result.Status = "canceled"
		return writeImportResult(cmd, jsonOut, result)
	}
	source, err := importer.ParseSource(raw)
	if err != nil {
		return err
	}
	importCfg := source.Config
	sourceConditionCount := len(importCfg.Conditions)

	if err := transformImportConfig(project, importCfg, opts); err != nil {
		var missingErr *importer.MissingGroupsError
		if !jsonOut && errors.As(err, &missingErr) && len(missingErr.Available) > 0 {
			groups := make([]groupSummary, 0, len(missingErr.Available))
			for _, group := range missingErr.Available {
				groups = append(groups, groupSummary{Name: group.Name, Parameters: group.Parameters})
			}
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), renderGroupsTable(groups))
		}
		return err
	}
	if !jsonOut {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), importConditionCountLine(sourceConditionCount, len(importCfg.Conditions)))
	}

	var currentRaw json.RawMessage
	var currentETag string
	if draftMode {
		cache, _, loadErr := svc.GetParameters(ctx, project.ProjectID, true)
		if loadErr != nil {
			return loadErr
		}
		currentRaw, currentETag = cache.RemoteConfig, cache.ETag
		if draftRaw, hasDraft, loadErr := svc.LoadDraft(project.ProjectID); loadErr != nil {
			return loadErr
		} else if hasDraft {
			currentRaw = draftRaw
		}
	} else {
		currentRaw, currentETag, err = svc.ExportRemoteConfig(ctx, project.ProjectID)
		if err != nil {
			return err
		}
	}
	currentCfg, err := firebase.ParseCloneRemoteConfig(currentRaw)
	if err != nil {
		return fmt.Errorf("decode current remote config: %w", err)
	}
	currentVersion := currentCfg.Version
	currentCfg.Version = firebase.RemoteConfigVersion{}

	finalCfg, err := buildFinalImportConfig(cmd, currentCfg, importCfg, opts)
	if err != nil {
		return err
	}
	finalCfg.Version = firebase.RemoteConfigVersion{}
	if draftMode {
		finalCfg.Version = currentVersion
	}
	pruneUnusedConditions(finalCfg)
	dropUnknownConditionReferences(finalCfg)
	normalizeEmptyParameterMaps(finalCfg)

	finalRaw, err := firebase.MarshalRemoteConfig(finalCfg)
	if err != nil {
		return err
	}

	diffText, hasChanges := rc.RenderRemoteConfigDiff(currentCfg, finalCfg)
	if !hasChanges {
		result.Status = "unchanged"
		return writeImportResult(cmd, jsonOut, result)
	}
	result.Changed = true
	if !draftMode {
		if err := svc.ValidateRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, currentETag); err != nil {
			return err
		}
	}

	if !jsonOut {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
	}

	action := "Publish Remote Config changes to"
	if draftMode {
		action = "Save Remote Config draft for"
	}
	if !yes {
		confirm := shared.NewConfirmation(
			fmt.Sprintf("%s %s?", action, project.ProjectID),
			shared.ConfirmationOptions{},
		)
		confirm.Input = cmd.InOrStdin()
		confirm.Output = cmd.ErrOrStderr()
		ok, err := confirm.RunPrompt()
		if err != nil {
			return err
		}
		if !ok {
			result.Status = "canceled"
			return writeImportResult(cmd, jsonOut, result)
		}
	}
	if draftMode {
		if !dryRun {
			if err := svc.SaveDraft(project.ProjectID, finalRaw); err != nil {
				return err
			}
		}
		result.Status = "drafted"
		if dryRun {
			result.Status = "would-draft"
		}
		return writeImportResult(cmd, jsonOut, result)
	}

	if _, _, err := svc.PublishRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, currentETag); err != nil {
		return err
	}
	result.Status = "imported"
	if dryRun {
		result.Status = "would-import"
	}
	return writeImportResult(cmd, jsonOut, result)
}

func writeImportResult(cmd *cobra.Command, jsonOut bool, result importResult) error {
	if jsonOut {
		return shared.WriteJSON(cmd, result)
	}
	switch result.Status {
	case "unchanged":
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🤷 No changes")
	case "drafted", "would-draft":
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📝 %s: %s\n", strings.ReplaceAll(result.Status, "-", " "), result.ProjectID)
	case "imported", "would-import":
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📥 %s: %s\n", strings.ReplaceAll(result.Status, "-", " "), result.ProjectID)
	}
	return nil
}

func importConditionCountLine(sourceCount, keptCount int) string {
	return fmt.Sprintf("Import conditions: %d kept · %d removed", keptCount, max(sourceCount-keptCount, 0))
}

func readImportOptions(cmd *cobra.Command) (importOptions, error) {
	var opts importOptions
	var err error

	opts.groups, err = cmd.Flags().GetStringArray("group")
	if err != nil {
		return opts, err
	}
	opts.paramFilters, err = cmd.Flags().GetStringArray("filter")
	if err != nil {
		return opts, err
	}
	opts.expr, err = cmd.Flags().GetString("expr")
	if err != nil {
		return opts, err
	}
	searchValue, err := cmd.Flags().GetString("search")
	if err != nil {
		return opts, err
	}
	opts.search = shared.NewParameterSearch(searchValue)
	opts.removeAllConditions, err = cmd.Flags().GetBool("remove-all-conditions")
	if err != nil {
		return opts, err
	}
	opts.keepPortableConditionsOnly, err = cmd.Flags().GetBool("keep-portable-conditions-only")
	if err != nil {
		return opts, err
	}
	opts.merge, err = cmd.Flags().GetBool("merge")
	if err != nil {
		return opts, err
	}
	opts.override, err = cmd.Flags().GetBool("override")
	if err != nil {
		return opts, err
	}
	opts.mergeResolve, err = cmd.Flags().GetString("merge-resolve")
	if err != nil {
		return opts, err
	}
	opts.mergeResolve = strings.TrimSpace(strings.ToLower(opts.mergeResolve))
	if opts.mergeResolve != "" && opts.mergeResolve != string(conflictResolutionCurrent) && opts.mergeResolve != string(conflictResolutionImport) {
		return opts, fmt.Errorf("invalid --merge-resolve value %q; expected current or import", opts.mergeResolve)
	}
	if opts.mergeResolve != "" && !opts.merge {
		return opts, fmt.Errorf("--merge-resolve requires --merge")
	}

	opts.groups = normalizeGroups(opts.groups)
	opts.expr = strings.TrimSpace(opts.expr)
	return opts, nil
}

func readRemoteConfig(cmd *cobra.Command) ([]byte, error) {
	fromPath, err := cmd.Flags().GetString("from")
	if err != nil {
		return nil, err
	}
	return shared.ReadJSONInput(cmd, fromPath, "remote config", nil)
}

func buildFinalImportConfig(cmd *cobra.Command, currentCfg, importCfg *firebase.RemoteConfig, opts importOptions) (*firebase.RemoteConfig, error) {
	if !configHasContent(currentCfg) {
		return firebase.CloneRemoteConfig(importCfg)
	}

	strategy, err := chooseImportStrategy(cmd, opts)
	if err != nil {
		return nil, err
	}
	if strategy == importStrategyOverride {
		return firebase.CloneRemoteConfig(importCfg)
	}

	return mergeRemoteConfigs(cmd, currentCfg, importCfg, opts)
}

func chooseImportStrategy(cmd *cobra.Command, opts importOptions) (importStrategy, error) {
	switch {
	case opts.override:
		return importStrategyOverride, nil
	case opts.merge:
		return importStrategyMerge, nil
	default:
		prompt := selection.New("Current config exists. How to apply import?", []mergeChoice{
			{label: "Merge imported config into current config", value: string(importStrategyMerge)},
			{label: "Override current config with imported config", value: string(importStrategyOverride)},
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
   {{- print (SelectedMarker $choice) (Selected $choice) "\n" }}
  {{- else }}
    {{- print "  " (Unselected $choice) "\n" }}
  {{- end }}
{{- end}}`
		prompt.SelectedChoiceStyle = styleImportStrategySelectedChoice
		prompt.UnselectedChoiceStyle = styleImportStrategyUnselectedChoice
		prompt.FinalChoiceStyle = styleImportStrategyFinalChoice
		prompt.ExtendedTemplateFuncs["SelectedMarker"] = styleImportStrategySelectedMarker
		prompt.Input = cmd.InOrStdin()
		prompt.Output = cmd.ErrOrStderr()
		choice, err := prompt.RunPrompt()
		if err != nil {
			return "", err
		}
		return importStrategy(choice.value), nil
	}
}
