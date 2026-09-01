package draft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/fileoutput"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/core/rc/publication"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "draft", Short: "Inspect, publish, and discard Remote Config drafts"}
	cmd.AddCommand(newListCommand(), newPathCommand(), newShowCommand(), newChangeNoteCommand(svc), newDiffCommand(svc), newPublishCommand(svc), newDiscardCommand())
	contract.MustRegisterResponsePath(cmd, "list", []listItem{})
	contract.MustRegisterResponsePath(cmd, "path", shared.PathResult{})
	contract.MustRegisterResponsePath(cmd, "show", contract.ArtifactData{})
	contract.MustRegisterResponsePath(cmd, "change-note", draftChangeNoteResult{})
	contract.MustRegisterResponsePath(cmd, "diff", draftDiffResult{})
	contract.MustRegisterResponsePath(cmd, "publish", []publishResult{}, rc.PlanCreatedResult{})
	contract.MustRegisterResponsePath(cmd, "discard", []draftDiscardResult{})
	return cmd
}

func newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print Remote Config drafts directory path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := config.GetDraftsDirPath()
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			if jsonOut {
				return shared.WriteJSON(cmd, shared.PathResult{Path: path})
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print path as JSON")
	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "list", Short: "List local Remote Config drafts", Args: cobra.NoArgs, RunE: runList}
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter template targets by optional client@ or server@ project query; may be repeated")
	cmd.Flags().Bool("json", false, "Print drafts as JSON")
	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	filters, _ := cmd.Flags().GetStringArray("filter")
	items, err := loadItems(filters)
	if err != nil {
		return err
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	if jsonOut {
		return shared.WriteJSON(cmd, items)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderList(items))
	return nil
}

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "show <project>", Short: "Show a local Remote Config draft", Args: cobra.ExactArgs(1), RunE: runShow}
	cmd.Flags().Bool("raw", false, "Print the exact stored draft envelope without parsing")
	cmd.Flags().String("to", "", "Write output to file path")
	return cmd
}

func newChangeNoteCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-note <project> [<text>]",
		Short: "Show or update a Remote Config draft change note",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, _, err := resolveDraft(args[0])
			if err != nil {
				return err
			}
			clear, _ := cmd.Flags().GetBool("clear")
			if clear && len(args) == 2 {
				return shared.InvalidArgument(fmt.Errorf("change note text cannot be used with --clear"))
			}
			if clear || len(args) == 2 {
				value := ""
				if len(args) == 2 {
					value = args[1]
				}
				if err := svc.SetDraftChangeNote(projectID, value); err != nil {
					return err
				}
			}
			stored, err := config.LoadDraft(projectID)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, draftChangeNoteResult{ProjectID: projectID, ChangeNote: optionalString(stored.ChangeNote)})
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), stored.ChangeNote)
			return err
		},
	}
	cmd.Flags().Bool("clear", false, "Clear the stored change note")
	cmd.Flags().Bool("json", false, "Print the draft change note as JSON")
	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	projectID, _, err := resolveDraft(args[0])
	if err != nil {
		return err
	}
	rawOut, _ := cmd.Flags().GetBool("raw")
	var body []byte
	if rawOut {
		body, err = os.ReadFile(config.GetDraftPath(projectID))
	} else {
		stored, loadErr := config.LoadDraft(projectID)
		if loadErr != nil {
			return loadErr
		}
		cfg, parseErr := firebase.ParseRemoteConfig(stored.RemoteConfig)
		if parseErr != nil {
			return parseErr
		}
		body, err = firebase.MarshalRemoteConfig(cfg)
	}
	if err != nil {
		return err
	}
	if !rawOut {
		body = rc.TrimTrailingLineBreaks(rc.NormalizeExportJSON(body))
	}
	to, _ := cmd.Flags().GetString("to")
	if to == "" {
		if contract.Enabled(cmd) {
			target := projectID
			return shared.WriteJSON(cmd, contract.NewArtifact(&target, "application/json", body, nil, false))
		}
		_, err = cmd.OutOrStdout().Write(body)
		return err
	}
	overwrite, proceed, err := shared.ConfirmFileOverwriteWithoutBypass(cmd, to)
	if err != nil {
		return err
	}
	if !proceed {
		return nil
	}
	write := fileoutput.Create
	if overwrite {
		write = fileoutput.Write
	}
	if err := write(to, body); err != nil {
		return err
	}
	if contract.Enabled(cmd) {
		target, destination := projectID, to
		return shared.WriteJSON(cmd, contract.NewArtifact(&target, "application/json", body, &destination, overwrite))
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📤 exported draft: %s\n", to)
	return nil
}

type diffOptions struct {
	against                              string
	cached, json, parameters, conditions bool
	filters, groups                      []string
	search, expr                         string
}

func newDiffCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "diff <project>", Short: "Compare a draft with its base or current Remote Config", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return runDiff(cmd, svc, args[0])
	}}
	cmd.Flags().String("against", "base", "Comparison target: base or current")
	cmd.Flags().Bool("cached", false, "Use the latest local snapshot and do not contact Firebase")
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().StringArray("group", nil, "Select parameters in group; may be repeated")
	cmd.Flags().String("expr", "", "Filter parameter changes by expr-lang expression")
	cmd.Flags().Bool("parameters", false, "Include only parameter and group description changes")
	cmd.Flags().Bool("conditions", false, "Include only condition changes")
	cmd.Flags().Bool("json", false, "Print diff as JSON")
	cmd.MarkFlagsMutuallyExclusive("parameters", "conditions")
	return cmd
}

func runDiff(cmd *cobra.Command, svc *core.Core, query string) error {
	projectID, project, err := resolveDraft(query)
	if err != nil {
		return err
	}
	stored, err := config.LoadDraft(projectID)
	if err != nil {
		return err
	}
	opts := readDiffOptions(cmd)
	var fromRaw, toRaw json.RawMessage
	fromRaw, toRaw = stored.BaseRemoteConfig, stored.RemoteConfig
	currentVersion := ""
	if opts.against == "current" {
		var cache *core.ParametersCache
		if opts.cached {
			cache, _, err = svc.InspectParametersCache(projectID)
			if err == nil && cache == nil {
				err = &shared.SelectionError{Resource: "parameters_cache", Kind: "not_found", Query: projectID}
			}
		} else {
			cache, _, err = svc.GetParameters(shared.CommandContext(cmd), projectID, true)
		}
		if err != nil {
			return err
		}
		var changed bool
		toRaw, changed, err = core.MergeDraftWithLatest(stored.BaseRemoteConfig, stored.RemoteConfig, cache.RemoteConfig)
		if err != nil {
			return err
		}
		if !changed {
			toRaw = cache.RemoteConfig
		}
		fromRaw = cache.RemoteConfig
		currentCfg, _ := firebase.ParseRemoteConfig(cache.RemoteConfig)
		currentVersion = currentCfg.Version.VersionNumber
	} else if opts.against != "base" {
		return shared.InvalidArgument(fmt.Errorf("--against must be base or current"))
	} else if opts.cached {
		return shared.InvalidArgument(fmt.Errorf("--cached requires --against current"))
	}
	fromCfg, err := firebase.ParseRemoteConfig(fromRaw)
	if err != nil {
		return err
	}
	toCfg, err := firebase.ParseRemoteConfig(toRaw)
	if err != nil {
		return err
	}
	result, err := filterDiff(project, rcdiff.CompareRemoteConfigs(fromCfg, toCfg), fromCfg, toCfg, opts)
	if err != nil {
		return err
	}
	if opts.json {
		if err := shared.WriteJSON(cmd, draftDiffResult{Project: project, Against: opts.against, BaseVersion: stored.BaseVersion, CurrentVersion: currentVersion, Changed: result.HasChanges(), Diff: result}); err != nil {
			return err
		}
		if result.HasChanges() {
			return shared.DiffFoundError(cmd)
		}
		return nil
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (%s): %s → draft\n", project.Name, projectID, opts.against); err != nil {
		return err
	}
	text, changed := rcdiff.RenderResult(result)
	if !changed {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "🤷 No differences")
		return err
	}
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), text); err != nil {
		return err
	}
	return shared.DiffFoundError(cmd)
}

func newPublishCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "publish [project...]", Short: "Safely rebase and publish Remote Config drafts", Args: cobra.ArbitraryArgs, RunE: func(cmd *cobra.Command, args []string) error {
		return runPublish(cmd, svc, args)
	}}
	cmd.Flags().Bool("all", false, "Publish every valid draft in the active profile")
	shared.AddDryRunFlag(cmd)
	shared.AddChangeNoteFlag(cmd)
	shared.AddYesFlag(cmd, "Skip publish confirmations")
	shared.AddPlanOutFlag(cmd)
	cmd.Flags().Bool("json", false, "Print results as JSON")
	return cmd
}

type publishResult struct {
	ProjectID        string             `json:"project_id"`
	Status           string             `json:"status" contract:"enum=failed|unchanged|would-publish|already-applied|published|published-hook-failed|published-cache-failed|published-cleanup-failed|conflict"`
	BaseVersion      string             `json:"base_version,omitempty"`
	PreviousVersion  string             `json:"previous_version,omitempty"`
	PublishedVersion string             `json:"published_version,omitempty"`
	Rebased          bool               `json:"rebased"`
	Changed          bool               `json:"changed"`
	DraftDeleted     bool               `json:"draft_deleted"`
	DryRun           bool               `json:"dry_run"`
	Validated        bool               `json:"validated"`
	ValidationSource string             `json:"validation_source"`
	Error            *draftPublishError `json:"error"`
	ChangeNote       *string            `json:"change_note"`
	Cause            error              `json:"-"`
}

type draftPublishError struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

func runPublish(cmd *cobra.Command, svc *core.Core, args []string) error {
	ids, err := selectedDraftIDs(cmd, args)
	if err != nil {
		return err
	}
	dry, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	jsonOut, _ := cmd.Flags().GetBool("json")
	changeNote, err := shared.ReadChangeNoteFlag(cmd)
	if err != nil {
		return err
	}
	ctx := shared.CommandContext(cmd)
	if planPath, planning, planErr := shared.PlanOutputPath(cmd); planErr != nil {
		return planErr
	} else if planning {
		return writeDraftPublicationPlan(ctx, cmd, svc, ids, changeNote, planPath)
	}
	if dry {
		ctx = firebase.WithDryRun(ctx)
	}
	if len(ids) > 1 && !dry {
		if !jsonOut {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: drafts are published independently for each project. Failures do not roll back projects already published.")
		}
		shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.non_atomic", Message: "Drafts are published independently; successful publications are not rolled back when another target fails.", Details: struct {
			TargetCount int `json:"target_count"`
		}{TargetCount: len(ids)}})
	}
	results := make([]publishResult, 0, len(ids))
	failed := false
	for _, projectID := range ids {
		progress.Start("Preparing draft for " + projectID + "…")
		result := publishResult{ProjectID: projectID, DryRun: dry, ValidationSource: core.ValidationSourceLocal}
		plan, prepareErr := svc.PrepareDraftPublish(ctx, projectID)
		if prepareErr != nil {
			result.Status, result.Error = "failed", &draftPublishError{Stage: "preparation", Message: shared.SafeErrorText(prepareErr)}
			result.Cause = prepareErr
			results = append(results, result)
			failed = true
			continue
		}
		result.BaseVersion = plan.Draft.BaseVersion
		effectiveNote := plan.ChangeNote
		if changeNote != nil {
			effectiveNote = *changeNote
		}
		result.ChangeNote = optionalString(effectiveNote)
		result.Rebased = plan.Rebased
		result.Changed = plan.HasChanges
		result.Validated = true
		latestCfg, _ := firebase.ParseRemoteConfig(plan.Latest.RemoteConfig)
		result.PreviousVersion = latestCfg.Version.VersionNumber
		if plan.HasChanges && !jsonOut {
			fromCfg, _ := firebase.ParseRemoteConfig(plan.Latest.RemoteConfig)
			toCfg, _ := firebase.ParseRemoteConfig(plan.Candidate)
			diffText, _ := rcdiff.RenderRemoteConfigDiff(fromCfg, toCfg)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Publish draft for %s\n%s\n", projectID, diffText)
		}
		if plan.HasChanges && !dry && !yes {
			if err := shared.RequireYesInMachineMode(cmd, yes, "publishing draft "+projectID, true); err != nil {
				return err
			}
			confirm := shared.NewConfirmation(fmt.Sprintf("Publish this draft to %s?", projectID), shared.ConfirmationOptions{Destructive: true})
			confirm.Output = cmd.ErrOrStderr()
			ok, confirmErr := confirm.RunPrompt()
			if confirmErr != nil {
				return confirmErr
			}
			if !ok {
				result.Status = "canceled"
				results = append(results, result)
				continue
			}
		}
		if dry {
			progress.Start("Previewing draft for " + projectID + "…")
		} else {
			progress.Start("Publishing draft for " + projectID + "…")
		}
		if changeNote != nil {
			if dry {
				plan.ChangeNote = *changeNote
			} else {
				if setErr := svc.SetDraftChangeNote(projectID, *changeNote); setErr != nil {
					result.Status, result.Error = "failed", &draftPublishError{Stage: "draft", Message: shared.SafeErrorText(setErr)}
					result.Cause = setErr
					results = append(results, result)
					failed = true
					continue
				}
				plan, prepareErr = svc.PrepareDraftPublish(ctx, projectID)
				if prepareErr != nil {
					result.Status, result.Error = "failed", &draftPublishError{Stage: "preparation", Message: shared.SafeErrorText(prepareErr)}
					result.Cause = prepareErr
					results = append(results, result)
					failed = true
					continue
				}
			}
		}
		cache, _, publishErr := svc.ExecuteDraftPublish(ctx, projectID, plan)
		if publishErr != nil {
			result.Cause = publishErr
			if source, ok := core.RemoteConfigValidationSource(publishErr); ok {
				result.Validated = false
				result.ValidationSource = source
			}
			var cleanupErr *core.DraftPublishedCleanupError
			var cacheErr *core.RemoteConfigPublishedCacheError
			var hookErr *core.RemoteConfigPublishedHookError
			if errors.As(publishErr, &hookErr) && cache != nil {
				markDraftFirebaseValidated(&result)
				publishedCfg, _ := firebase.ParseRemoteConfig(cache.RemoteConfig)
				result.Status = "published-hook-failed"
				result.PublishedVersion = publishedCfg.Version.VersionNumber
				result.Error = &draftPublishError{Stage: "post_publish_hook", Message: shared.SafeErrorText(publishErr)}
				shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.post_publish_hook_failed", Message: "Firebase accepted the draft publication, but a post_publish hook failed.", Target: projectID, Details: struct {
					Stage string `json:"stage"`
				}{Stage: "post_publish_hook"}, Remediation: []shared.Remediation{{Description: "inspect hook trust and status without republishing", Strategy: shared.RemediationRunCommand, Argv: []string{"hooks", "status"}}}})
				results = append(results, result)
				failed = true
				continue
			}
			if errors.As(publishErr, &cleanupErr) && cache != nil {
				markDraftFirebaseValidated(&result)
				publishedCfg, _ := firebase.ParseRemoteConfig(cache.RemoteConfig)
				result.Status = "published-cleanup-failed"
				result.PublishedVersion = publishedCfg.Version.VersionNumber
				result.Error = &draftPublishError{Stage: "cleanup", Message: shared.SafeErrorText(publishErr)}
				shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.draft_cleanup_failed", Message: "Firebase accepted the publication, but the local draft could not be removed.", Target: projectID, Details: struct {
					Stage string `json:"stage"`
				}{Stage: "cleanup"}, Remediation: []shared.Remediation{{Description: "retry safe draft cleanup", Strategy: shared.RemediationRunCommand, Argv: []string{"draft", "publish", projectID}}}})
				results = append(results, result)
				failed = true
				continue
			}
			if errors.As(publishErr, &cacheErr) && cache != nil {
				markDraftFirebaseValidated(&result)
				publishedCfg, _ := firebase.ParseRemoteConfig(cache.RemoteConfig)
				result.Status = "published-cache-failed"
				result.PublishedVersion = publishedCfg.Version.VersionNumber
				result.Error = &draftPublishError{Stage: "cache", Message: shared.SafeErrorText(publishErr)}
				filter, _ := rctarget.ExactFilter(projectID)
				shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.cache_stale", Message: "Firebase accepted the draft publication, but the local cache update failed.", Target: projectID, Details: struct {
					Stage string `json:"stage"`
				}{Stage: "cache"}, Remediation: []shared.Remediation{{Description: "refresh the published target instead of republishing", Strategy: shared.RemediationRunCommand, Argv: []string{"get", "--update", "--project", filter}}}})
				results = append(results, result)
				failed = true
				continue
			}
			if rc.IsRemoteConfigConflict(publishErr) {
				result.Status = "conflict"
				result.Error = &draftPublishError{Stage: "publication", Message: shared.SafeErrorText(publishErr)}
				results = append(results, result)
				failed = true
				continue
			}
			stage := "publication"
			if !result.Validated {
				stage = "validation"
			}
			result.Status, result.Error = "failed", &draftPublishError{Stage: stage, Message: shared.SafeErrorText(publishErr)}
			results = append(results, result)
			failed = true
			continue
		}
		if plan.HasChanges {
			result.Validated = true
			result.ValidationSource = core.ValidationSourceFirebase
		}
		result.Status, result.DraftDeleted = successfulPublishStatus(dry, plan.HasChanges)
		if result.Status == "published" {
			publishedCfg, _ := firebase.ParseRemoteConfig(cache.RemoteConfig)
			result.PublishedVersion = publishedCfg.Version.VersionNumber
		}
		results = append(results, result)
	}
	if jsonOut {
		if err := shared.WriteJSON(cmd, results); err != nil {
			return err
		}
	} else {
		if len(results) > 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Results:")
		}
		for _, result := range results {
			if result.Error != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s: %s\nvalidated: %t · validation_source: %s\n", result.ProjectID, result.Error.Message, result.Validated, result.ValidationSource)
				continue
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\nvalidated: %t · validation_source: %s\n", result.Status, result.ProjectID, result.Validated, result.ValidationSource)
		}
	}
	var retryIDs []string
	for _, result := range results {
		if result.Error != nil && result.Status != "published-cleanup-failed" && result.Status != "published-cache-failed" && result.Status != "published-hook-failed" {
			retryIDs = append(retryIDs, result.ProjectID)
		}
	}
	if !jsonOut && len(retryIDs) > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Retry failed drafts:\n  fbrcm draft publish %s\n", strings.Join(retryIDs, " "))
	}
	if !jsonOut {
		var cacheFailed, cleanupFailed []string
		for _, result := range results {
			switch result.Status {
			case "published-cache-failed":
				cacheFailed = append(cacheFailed, result.ProjectID)
			case "published-cleanup-failed":
				cleanupFailed = append(cleanupFailed, result.ProjectID)
			}
		}
		if len(cacheFailed) > 0 {
			filters := make([]string, 0, len(cacheFailed))
			for _, id := range cacheFailed {
				filter, err := rctarget.ExactFilter(id)
				if err == nil {
					filters = append(filters, fmt.Sprintf("-p '%s'", filter))
				}
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Firebase was updated, but local caches are stale. Refresh them with:\n  fbrcm get --update %s\n", strings.Join(filters, " "))
		}
		if len(cleanupFailed) > 0 {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Firebase was updated, but draft cleanup failed. Safely retry cleanup with:\n  fbrcm draft publish %s\n", strings.Join(cleanupFailed, " "))
		}
	}
	if failed {
		failedTargets := make([]string, 0)
		targetFailures := make([]shared.BatchFailure, 0)
		successfulTargets := 0
		publishedTargets := 0
		for _, result := range results {
			if result.Error != nil {
				failedTargets = append(failedTargets, result.ProjectID)
				targetFailures = append(targetFailures, shared.BatchFailure{Target: result.ProjectID, Err: result.Cause})
			}
			if result.Error == nil || strings.HasPrefix(result.Status, "published-") {
				successfulTargets++
			}
			if result.Status == "published" || strings.HasPrefix(result.Status, "published-") {
				publishedTargets++
			}
		}
		remediation := []shared.Remediation{}
		if len(retryIDs) > 0 {
			remediation = append(remediation, shared.Remediation{Description: "retry only drafts that did not publish", Strategy: shared.RemediationRunCommand, Argv: append([]string{"draft", "publish"}, retryIDs...)})
		}
		return &shared.BatchError{Operation: "draft.publish", FailedTargets: failedTargets, Failures: targetFailures, SuccessfulTargetCount: successfulTargets, PublishedTargetCount: publishedTargets, Remediation: remediation, Err: fmt.Errorf("%d drafts failed", len(failedTargets))}
	}
	return nil
}

func writeDraftPublicationPlan(ctx context.Context, cmd *cobra.Command, svc *core.Core, ids []string, overrideNote *string, path string) error {
	environment, err := core.PublicationEnvironmentForContext(ctx)
	if err != nil {
		return err
	}
	publicationPlan := publication.New(cmd.Root().Version, "draft.publish", environment.Policy, overrideNote)
	publicationPlan.Execution.HooksEnabled = environment.HooksEnabled
	publicationPlan.Execution.HookDefinitionSHA256 = environment.HookDefinitionSHA256
	publicationPlan.Operation.Selection, err = json.Marshal(struct {
		Targets []string `json:"targets"`
	}{Targets: append([]string(nil), ids...)})
	if err != nil {
		return err
	}
	for _, projectID := range ids {
		progress.Start("Preparing draft publication plan for " + projectID + "…")
		draftPlan, err := svc.PrepareDraftPublish(ctx, projectID)
		if err != nil {
			return err
		}
		note := draftPlan.ChangeNote
		if overrideNote != nil {
			note = *overrideNote
		}
		notePointer := optionalString(note)
		targetCtx := ctx
		if notePointer != nil {
			targetCtx, err = firebase.WithChangeNote(targetCtx, *notePointer)
			if err != nil {
				return err
			}
		}
		action := publication.ActionNone
		validationSource := core.ValidationSourceLocal
		if draftPlan.HasChanges {
			action = publication.ActionPublish
			validationSource = core.ValidationSourceFirebase
			if err := svc.ValidatePublicationCandidate(targetCtx, projectID, draftPlan.Latest.RemoteConfig, draftPlan.Candidate, draftPlan.Latest.ETag, "draft-publish"); err != nil {
				return err
			}
		}
		target, err := rctarget.Parse(projectID)
		if err != nil {
			return err
		}
		latestCfg, err := firebase.ParseRemoteConfig(draftPlan.Latest.RemoteConfig)
		if err != nil {
			return err
		}
		publicationPlan.Targets = append(publicationPlan.Targets, publication.Target{
			Target: projectID, ProjectID: target.ProjectID, Template: string(target.Kind), Action: action, ChangeNote: notePointer,
			Base:       publication.Snapshot{Version: latestCfg.Version.VersionNumber, ETag: draftPlan.Latest.ETag, RemoteConfig: draftPlan.Latest.RemoteConfig},
			Candidate:  publication.Snapshot{RemoteConfig: draftPlan.Candidate},
			Validation: publication.Validation{Source: validationSource, ValidatedAt: publicationPlan.CreatedAt},
			Source:     publication.Source{Kind: "draft", Fingerprint: draftPlan.Draft.UpdatedAt.UTC().Format(time.RFC3339Nano)},
		})
	}
	_, err = rc.WritePublicationPlan(cmd, publicationPlan, path)
	return err
}

func markDraftFirebaseValidated(result *publishResult) {
	result.Validated = true
	result.ValidationSource = core.ValidationSourceFirebase
}

func successfulPublishStatus(dryRun, changed bool) (string, bool) {
	switch {
	case !changed && dryRun:
		return "unchanged", false
	case !changed:
		return "already-applied", true
	case dryRun:
		return "would-publish", false
	default:
		return "published", true
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func newDiscardCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "discard [project...]", Short: "Discard local Remote Config drafts", Args: cobra.ArbitraryArgs, RunE: runDiscard}
	cmd.Flags().Bool("all", false, "Discard every draft in the active profile")
	shared.AddYesFlag(cmd, "Skip destructive confirmations")
	cmd.Flags().Bool("json", false, "Print results as JSON")
	return cmd
}

func runDiscard(cmd *cobra.Command, args []string) error {
	ids, err := selectedDraftIDs(cmd, args)
	if err != nil {
		return err
	}
	yes, _ := cmd.Flags().GetBool("yes")
	jsonOut, _ := cmd.Flags().GetBool("json")
	results := make([]draftDiscardResult, 0, len(ids))
	for _, projectID := range ids {
		stored, loadErr := config.LoadDraft(projectID)
		if !jsonOut {
			if loadErr != nil {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Draft preview unavailable:", loadErr)
			} else {
				baseCfg, baseErr := firebase.ParseRemoteConfig(stored.BaseRemoteConfig)
				draftCfg, draftErr := firebase.ParseRemoteConfig(stored.RemoteConfig)
				if baseErr == nil && draftErr == nil {
					if text, changed := rcdiff.RenderRemoteConfigDiff(baseCfg, draftCfg); changed {
						_, _ = fmt.Fprintln(cmd.ErrOrStderr(), text)
					}
				}
			}
		}
		if !yes {
			if err := shared.RequireYesInMachineMode(cmd, yes, "discarding draft "+projectID, true); err != nil {
				return err
			}
			confirm := shared.NewConfirmation(fmt.Sprintf("Discard draft for %s?", projectID), shared.ConfirmationOptions{Destructive: true})
			confirm.Output = cmd.ErrOrStderr()
			ok, confirmErr := confirm.RunPrompt()
			if confirmErr != nil {
				return confirmErr
			}
			if !ok {
				results = append(results, draftDiscardResult{ProjectID: projectID, Status: "canceled"})
				continue
			}
		}
		if err := config.DeleteDraft(projectID); err != nil {
			return err
		}
		baseVersion := ""
		if stored != nil {
			baseVersion = stored.BaseVersion
		}
		results = append(results, draftDiscardResult{ProjectID: projectID, Status: "discarded", BaseVersion: baseVersion})
	}
	if jsonOut {
		return shared.WriteJSON(cmd, results)
	}
	if len(results) > 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Results:")
	}
	for _, result := range results {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", result.Status, result.ProjectID)
	}
	return nil
}

type draftChangeNoteResult struct {
	ProjectID  string  `json:"project_id"`
	ChangeNote *string `json:"change_note"`
}

type draftDiffResult struct {
	Project        core.Project  `json:"project"`
	Against        string        `json:"against" contract:"enum=base|current"`
	BaseVersion    string        `json:"base_version"`
	CurrentVersion string        `json:"current_version"`
	Changed        bool          `json:"changed"`
	Diff           rcdiff.Result `json:"diff"`
}

type draftDiscardResult struct {
	ProjectID   string `json:"project_id"`
	Status      string `json:"status" contract:"enum=discarded"`
	BaseVersion string `json:"base_version"`
}
