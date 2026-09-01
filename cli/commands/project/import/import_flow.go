package importpkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corehooks "github.com/yumauri/fbrcm/core/hooks"
	"github.com/yumauri/fbrcm/core/rc/importer"
	"github.com/yumauri/fbrcm/core/rc/publication"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

type importResult struct {
	ProjectID        string       `json:"project_id"`
	Status           string       `json:"status" contract:"enum=unchanged|validation-failed|drafted|would-draft|imported|would-import|imported-hook-failed|imported-cache-failed"`
	Changed          bool         `json:"changed"`
	Draft            bool         `json:"draft"`
	DryRun           bool         `json:"dry_run"`
	Validated        bool         `json:"validated"`
	ValidationSource string       `json:"validation_source"`
	ChangeNote       *string      `json:"change_note"`
	Published        bool         `json:"published,omitempty"`
	Error            *importError `json:"error"`
}

// Result is the machine response DTO registered by the parent project command.
type Result = importResult

type importError struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
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
	result := importResult{ProjectID: project.ProjectID, Draft: draftMode, DryRun: dryRun, ValidationSource: core.ValidationSourceLocal}
	ctx := shared.CommandContext(cmd)
	if err := shared.RejectStatelessDraft(ctx, draftMode); err != nil {
		return err
	}
	if dryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	changeNote, err := shared.ReadChangeNoteFlag(cmd)
	if err != nil {
		return err
	}
	ctx, err = shared.WithChangeNote(ctx, changeNote)
	if err != nil {
		return err
	}
	result.ChangeNote = changeNote
	if !draftMode && core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		hasDraft, draftErr := svc.HasDraft(project.ProjectID)
		if draftErr != nil {
			return draftErr
		}
		if hasDraft {
			return &shared.ConflictError{Code: "draft.exists", Resource: "draft", Target: project.ProjectID, Remediation: []shared.Remediation{
				{Description: "import into the existing draft", Strategy: shared.RemediationRetryWithArguments, Argv: []string{"--draft"}},
				{Description: "publish the existing draft", Strategy: shared.RemediationRunCommand, Argv: []string{"draft", "publish", project.ProjectID}},
				{Description: "discard the existing draft", Strategy: shared.RemediationRunCommand, Argv: []string{"draft", "discard", project.ProjectID}},
			}, Err: fmt.Errorf("project %s has an unpublished draft; use --draft, publish it, or discard it first", project.ProjectID)}
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
		return shared.InvalidInput("remote_config.invalid", "stdin", err)
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

	progress.Start("Loading current Remote Config for " + project.ProjectID + "…")
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
	result.Validated = true

	diffText, hasChanges := rc.RenderRemoteConfigDiff(currentCfg, finalCfg)
	if planPath, planning, planErr := shared.PlanOutputPath(cmd); planErr != nil {
		return planErr
	} else if planning {
		environment, environmentErr := core.PublicationEnvironmentForContext(ctx)
		if environmentErr != nil {
			return environmentErr
		}
		action := publication.ActionNone
		validationSource := core.ValidationSourceLocal
		if hasChanges {
			action = publication.ActionPublish
			validationSource = core.ValidationSourceFirebase
			if validationErr := svc.ValidatePublicationCandidate(ctx, project.ProjectID, currentRaw, finalRaw, currentETag, "import"); validationErr != nil {
				return validationErr
			}
		}
		target, targetErr := rctarget.Parse(project.ProjectID)
		if targetErr != nil {
			return targetErr
		}
		publicationPlan := publication.New(cmd.Root().Version, "project.import", environment.Policy, changeNote)
		publicationPlan.Execution.HooksEnabled = environment.HooksEnabled
		publicationPlan.Execution.HookDefinitionSHA256 = environment.HookDefinitionSHA256
		plannerOptions := opts.plannerOptions()
		publicationPlan.Operation.Selection, targetErr = json.Marshal(struct {
			Groups            []string `json:"groups"`
			Filters           []string `json:"filters"`
			Search            string   `json:"search"`
			Expr              string   `json:"expr"`
			Strategy          string   `json:"strategy"`
			ConditionPolicy   string   `json:"condition_policy"`
			DefaultResolution string   `json:"default_resolution"`
		}{
			Groups: plannerOptions.Groups, Filters: plannerOptions.Filters, Search: plannerOptions.Search,
			Expr: plannerOptions.Expr, Strategy: string(plannerOptions.Strategy), ConditionPolicy: string(plannerOptions.ConditionPolicy),
			DefaultResolution: string(plannerOptions.DefaultResolution),
		})
		if targetErr != nil {
			return targetErr
		}
		sourceRaw, targetErr := firebase.MarshalRemoteConfig(importCfg)
		if targetErr != nil {
			return targetErr
		}
		sourceDigest, targetErr := publication.RemoteConfigDigest(sourceRaw)
		if targetErr != nil {
			return targetErr
		}
		publicationPlan.Targets = append(publicationPlan.Targets, publication.Target{
			Target: project.ProjectID, ProjectID: target.ProjectID, Template: string(target.Kind), Action: action, ChangeNote: changeNote,
			Base:       publication.Snapshot{Version: currentVersion.VersionNumber, ETag: currentETag, RemoteConfig: currentRaw},
			Candidate:  publication.Snapshot{RemoteConfig: finalRaw},
			Validation: publication.Validation{Source: validationSource, ValidatedAt: publicationPlan.CreatedAt},
			Source:     publication.Source{Kind: "import", Fingerprint: sourceDigest},
		})
		_, writeErr := rc.WritePublicationPlan(cmd, publicationPlan, planPath)
		return writeErr
	}
	if !hasChanges {
		result.Status = "unchanged"
		return writeImportResult(cmd, jsonOut, result)
	}
	result.Changed = true
	if !draftMode {
		progress.Start("Validating Remote Config for " + project.ProjectID + "…")
		result.Validated = false
		result.ValidationSource = core.ValidationSourceFirebase
		if err := svc.ValidateRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, currentETag); err != nil {
			if source, ok := core.RemoteConfigValidationSource(err); ok {
				result.ValidationSource = source
			}
			result.Status = "validation-failed"
			result.Error = &importError{Stage: "validation", Message: shared.SafeErrorText(err)}
			if writeErr := writeImportResult(cmd, jsonOut, result); writeErr != nil {
				return writeErr
			}
			return err
		}
		result.Validated = true
	}

	if !jsonOut {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
	}

	action := "Publish Remote Config changes to"
	if draftMode {
		action = "Save Remote Config draft for"
	}
	if !yes {
		if err := shared.RequireYesInMachineMode(cmd, yes, strings.ToLower(action)+" "+project.ProjectID, true); err != nil {
			return err
		}
		confirm := shared.NewConfirmation(
			fmt.Sprintf("%s %s?", action, project.ProjectID),
			shared.ConfirmationOptions{Destructive: true},
		)
		promptInput, closePromptInput, err := shared.OpenPromptInput(cmd.InOrStdin())
		if err != nil {
			return err
		}
		defer closePromptInput()
		confirm.Input = promptInput
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
		progress.Start("Saving draft for " + project.ProjectID + "…")
		if !dryRun {
			var saveErr error
			if note, set := firebase.ChangeNoteFromContext(ctx); set {
				saveErr = svc.SaveDraftWithChangeNote(project.ProjectID, finalRaw, core.DraftChangeNoteUpdate{Set: true, Value: note})
			} else {
				saveErr = svc.SaveDraft(project.ProjectID, finalRaw)
			}
			if saveErr != nil {
				return saveErr
			}
		}
		result.Status = "drafted"
		if dryRun {
			result.Status = "would-draft"
		}
		return writeImportResult(cmd, jsonOut, result)
	}

	if dryRun {
		progress.Start("Previewing Remote Config import for " + project.ProjectID + "…")
	} else {
		progress.Start("Publishing Remote Config import for " + project.ProjectID + "…")
	}
	ctx = corehooks.WithOperation(ctx, "import")
	publishedRaw, _, publishErr := svc.PublishRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, currentETag)
	if publishErr != nil {
		var hookErr *core.RemoteConfigPublishedHookError
		var cacheErr *core.RemoteConfigPublishedCacheError
		if len(publishedRaw) == 0 || !errors.As(publishErr, &hookErr) && !errors.As(publishErr, &cacheErr) {
			return publishErr
		}
		result.Published = true
		if errors.As(publishErr, &hookErr) {
			result.Status = "imported-hook-failed"
			result.Error = &importError{Stage: "post_publish_hook", Message: shared.SafeErrorText(publishErr)}
			shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.post_publish_hook_failed", Message: "Firebase accepted the import, but a post_publish hook failed.", Target: project.ProjectID, Details: struct {
				Stage string `json:"stage"`
			}{Stage: "post_publish_hook"}, Remediation: []shared.Remediation{{Description: "inspect hook trust and status without republishing", Strategy: shared.RemediationRunCommand, Argv: []string{"hooks", "status"}}}})
		} else {
			result.Status = "imported-cache-failed"
			result.Error = &importError{Stage: "cache", Message: shared.SafeErrorText(publishErr)}
			filter, _ := rctarget.ExactFilter(project.ProjectID)
			shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.cache_stale", Message: "Firebase accepted the import, but the local cache update failed.", Target: project.ProjectID, Details: struct {
				Stage string `json:"stage"`
			}{Stage: "cache"}, Remediation: []shared.Remediation{{Description: "refresh the imported target instead of republishing", Strategy: shared.RemediationRunCommand, Argv: []string{"get", "--update", "--project", filter}}}})
		}
		if writeErr := writeImportResult(cmd, jsonOut, result); writeErr != nil {
			return writeErr
		}
		return publishErr
	}
	result.Status = "imported"
	result.Published = !dryRun
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
	case "imported-hook-failed":
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "⚠️ imported, post_publish hook failed: %s: %s\n", result.ProjectID, result.Error.Message)
	case "validation-failed":
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "❌ validation failed: %s: %s\n", result.ProjectID, result.Error.Message)
	}
	out := cmd.OutOrStdout()
	if result.Error != nil {
		out = cmd.ErrOrStderr()
	}
	_, _ = fmt.Fprintf(out, "validated: %t · validation_source: %s\n", result.Validated, result.ValidationSource)
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
		return opts, shared.InvalidArgument(fmt.Errorf("invalid --merge-resolve value %q; expected current or import", opts.mergeResolve))
	}
	if opts.mergeResolve != "" && !opts.merge {
		return opts, shared.InvalidArgument(fmt.Errorf("--merge-resolve requires --merge"))
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
		if shared.MachineMode(cmd) {
			return "", shared.InteractionRequiredWithArguments("an import strategy is required; pass --merge or --override", "selection_required", false, "--merge", "--merge")
		}
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
		return importStrategy(choice.value), nil
	}
}
