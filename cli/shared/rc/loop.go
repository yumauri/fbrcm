package rc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/display"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

// RemoteMutationStatus is the final per-project outcome of a mutation batch.
type RemoteMutationStatus string

const (
	RemoteMutationUnchanged            RemoteMutationStatus = "unchanged"
	RemoteMutationPreparationFailed    RemoteMutationStatus = "preparation-failed"
	RemoteMutationPublished            RemoteMutationStatus = "published"
	RemoteMutationValidationFailed     RemoteMutationStatus = "validation-failed"
	RemoteMutationConflict             RemoteMutationStatus = "conflict"
	RemoteMutationPublishFailed        RemoteMutationStatus = "publish-failed"
	RemoteMutationPublishedCacheFailed RemoteMutationStatus = "published-cache-failed"
	RemoteMutationPublishedHookFailed  RemoteMutationStatus = "published-hook-failed"
	RemoteMutationDrafted              RemoteMutationStatus = "drafted"
	RemoteMutationWouldDraft           RemoteMutationStatus = "would-draft"
	RemoteMutationWouldPublish         RemoteMutationStatus = "would-publish"
	RemoteMutationDraftFailed          RemoteMutationStatus = "draft-failed"
)

// RemoteMutationResult records one selected project's final outcome.
type RemoteMutationResult struct {
	Project          core.Project
	Status           RemoteMutationStatus
	ChangedCount     int
	PreviousVersion  string
	PublishedVersion string
	ErrorStage       string
	Validated        bool
	ValidationSource string
	Published        bool
	Err              error
	ChangeNote       *string
	MatchedItemCount int
	NoOpReason       *NoOpReason
}

// RemoteMutationTotals contains aggregate counts and ordered project results.
type RemoteMutationTotals struct {
	ModifiedProjects int
	ChangedParams    int
	Results          []RemoteMutationResult
	DefaultScope     bool
	ResolvedTargets  int
}

func (t RemoteMutationTotals) failedProjectIDs() []string {
	ids := make([]string, 0)
	for _, result := range t.Results {
		if result.Err != nil && !result.Published {
			ids = append(ids, result.Project.ProjectID)
		}
	}
	return ids
}

func (t RemoteMutationTotals) failureCount() int {
	count := 0
	for _, result := range t.Results {
		if result.Err != nil {
			count++
		}
	}
	return count
}

// RemoteMutationPlanner builds the per-project mutation from a freshly revalidated
// config. Returning a nil mutation leaves the project untouched.
type RemoteMutationPlanner func(project core.Project, cfg *ProjectConfig) (RemoteMutationPlan, error)

// NoOpReason distinguishes an empty selection from an already-applied change.
type NoOpReason string

const (
	NoOpNoMatch        NoOpReason = "no_match"
	NoOpAlreadyApplied NoOpReason = "already_applied"
)

// RunRemoteDraftLoop applies mutations on top of each project's draft and records
// failures independently. It never writes to Firebase.
func RunRemoteDraftLoop(ctx context.Context, cmd *cobra.Command, svc *core.Core, projects []core.Project, defaultScope bool, operation string, plan RemoteMutationPlanner) (RemoteMutationTotals, error) {
	totals := RemoteMutationTotals{DefaultScope: defaultScope, ResolvedTargets: len(projects)}
	for _, project := range projects {
		progress.Start("Preparing draft for " + project.ProjectID + "…")
		result := RemoteMutationResult{Project: project, ChangeNote: remoteMutationChangeNote(ctx), ValidationSource: core.ValidationSourceLocal}
		cfg, err := RevalidateProjectConfig(ctx, svc, project)
		if err == nil {
			if draftRaw, hasDraft, loadErr := svc.LoadDraft(project.ProjectID); loadErr != nil {
				err = loadErr
			} else if hasDraft {
				var draftCfg *firebase.RemoteConfig
				draftCfg, err = firebase.ParseRemoteConfig(draftRaw)
				if err == nil {
					cfg.Config = draftCfg
				}
			}
		}
		if cfg != nil && cfg.Config != nil {
			result.PreviousVersion = cfg.Config.Version.VersionNumber
		}
		var mutationPlan RemoteMutationPlan
		if err == nil {
			mutationPlan, err = plan(project, cfg)
			result.MatchedItemCount = mutationPlan.MatchedItemCount
		}
		if err == nil && mutationPlan.Mutation == nil {
			result.Validated = true
			result.Status = RemoteMutationUnchanged
			reason := NoOpNoMatch
			result.NoOpReason = &reason
			totals.Results = append(totals.Results, result)
			continue
		}
		var finalCfg *firebase.RemoteConfig
		if err == nil {
			result.ChangedCount, finalCfg, err = mutationPlan.Mutation(cfg.Config)
		}
		if err == nil {
			err = rcmutate.EnsureOpaqueValuesUnchanged(cfg.Config, finalCfg)
		}
		if err == nil && result.ChangedCount == 0 {
			result.Validated = true
			result.Status = RemoteMutationUnchanged
			reason := NoOpAlreadyApplied
			result.NoOpReason = &reason
			totals.Results = append(totals.Results, result)
			continue
		}
		var finalRaw []byte
		if err == nil {
			finalRaw, err = firebase.MarshalRemoteConfig(finalCfg)
		}
		if err == nil {
			result.Validated = true
		}
		if err == nil && !firebase.IsDryRun(ctx) {
			progress.Start("Saving draft for " + project.ProjectID + "…")
			if changeNote, set := firebase.ChangeNoteFromContext(ctx); set {
				err = svc.SaveDraftWithChangeNote(project.ProjectID, finalRaw, core.DraftChangeNoteUpdate{Set: true, Value: changeNote})
			} else {
				err = svc.SaveDraft(project.ProjectID, finalRaw)
			}
		}
		if err != nil {
			result.Status, result.Err = RemoteMutationDraftFailed, err
		} else {
			result.Status = RemoteMutationDrafted
			if firebase.IsDryRun(ctx) {
				result.Status = RemoteMutationWouldDraft
			}
			totals.ModifiedProjects++
			totals.ChangedParams += result.ChangedCount
		}
		totals.Results = append(totals.Results, result)
		if batchMustStop(ctx, err) {
			break
		}
	}
	return totals, mutationBatchError(totals, operation)
}

// RunRemotePublishLoop validates and publishes every selected project
// independently. Project-scoped failures are reported and do not stop later
// projects. Conflicts are left for a fresh, explicitly reviewed retry.
func RunRemotePublishLoop(ctx context.Context, cmd *cobra.Command, svc *core.Core, projects []core.Project, defaultScope bool, operation, publishedEmoji string, plan RemoteMutationPlanner) (RemoteMutationTotals, error) {
	totals := RemoteMutationTotals{DefaultScope: defaultScope, ResolvedTargets: len(projects)}
	if len(projects) > 1 && !firebase.IsDryRun(ctx) {
		jsonOut, _ := cmd.Flags().GetBool("json")
		message := "Warning: Remote Config is published independently for each project. Some projects may succeed while others fail."
		var remediation []machine.Remediation
		if core.ExecutionPolicyFromContext(ctx).WriteLocalState {
			message += " For coordinated changes, consider staging with --draft first."
			remediation = []machine.Remediation{{Description: "stage the changes as drafts before publishing", Strategy: machine.RemediationRetryWithArguments, Argv: []string{"--draft"}}}
		}
		if !jsonOut {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), message)
		}
		machine.FromContext(cmd.Context()).AddWarning(machine.Warning{Code: "publication.non_atomic", Message: "Remote Config is published independently for each target; successful publications are not rolled back when another target fails.", Details: struct {
			TargetCount int `json:"target_count"`
		}{TargetCount: len(projects)}, Remediation: remediation})
	}

	for _, project := range projects {
		progress.Start("Preparing Remote Config for " + project.ProjectID + "…")
		result := RemoteMutationResult{Project: project, ChangeNote: remoteMutationChangeNote(ctx), ValidationSource: core.ValidationSourceLocal}
		preparationFailed := false
		var err error
		if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
			hasDraft, draftErr := svc.HasDraft(project.ProjectID)
			err = draftErr
			if err == nil && hasDraft {
				err = &machine.ConflictError{Code: "draft.exists", Resource: "draft", Target: project.ProjectID, Remediation: []machine.Remediation{
					{Description: "apply the mutation to the existing draft", Strategy: machine.RemediationRetryWithArguments, Argv: []string{"--draft"}},
					{Description: "publish the existing draft", Strategy: machine.RemediationRunCommand, Argv: []string{"draft", "publish", project.ProjectID}},
					{Description: "discard the existing draft", Strategy: machine.RemediationRunCommand, Argv: []string{"draft", "discard", project.ProjectID}},
				}, Err: fmt.Errorf("project has an unpublished draft; use --draft, publish it, or discard it first")}
			}
		}
		var cfg *ProjectConfig
		if err == nil {
			cfg, err = RevalidateProjectConfig(ctx, svc, project)
		}
		if cfg != nil && cfg.Config != nil {
			result.PreviousVersion = cfg.Config.Version.VersionNumber
		}
		var mutationPlan RemoteMutationPlan
		if err == nil {
			mutationPlan, err = plan(project, cfg)
			result.MatchedItemCount = mutationPlan.MatchedItemCount
		}
		preparationFailed = err != nil
		if err == nil && mutationPlan.Mutation == nil {
			result.Validated = true
			result.Status = RemoteMutationUnchanged
			reason := NoOpNoMatch
			result.NoOpReason = &reason
			totals.Results = append(totals.Results, result)
			continue
		}
		var publishResult RemoteConfigMutationPublishResult
		if err == nil {
			publishResult, err = PublishProjectConfigMutationResult(ctx, svc, cfg, operation, nil, mutationPlan.Mutation)
			result.ChangedCount = publishResult.ChangedCount
			result.PublishedVersion = publishResult.PublishedVersion
			result.ErrorStage = publishResult.FailureStage
			result.Validated = publishResult.Validated
			result.ValidationSource = publishResult.ValidationSource
		}
		switch {
		case publishResult.Retry:
			result.Status = RemoteMutationConflict
			result.Err = &machine.ConflictError{Code: "remote_config.conflict", Resource: "remote_config", Target: project.ProjectID, Retryable: true, Err: fmt.Errorf("remote config changed during %s; rerun the command to review a fresh candidate", operation)}
		case err != nil:
			var hookPublishErr *core.RemoteConfigPublishedHookError
			var cacheErr *core.RemoteConfigPublishedCacheError
			if errors.As(err, &hookPublishErr) {
				result.Status, result.Published, result.Err = RemoteMutationPublishedHookFailed, true, err
				machine.FromContext(cmd.Context()).AddWarning(machine.Warning{Code: "publication.post_publish_hook_failed", Message: "Firebase accepted the publication, but a post_publish hook failed.", Target: project.ProjectID, Details: struct {
					Stage string `json:"stage"`
				}{Stage: "post_publish_hook"}, Remediation: []machine.Remediation{{Description: "inspect hook trust and status without republishing", Strategy: machine.RemediationRunCommand, Argv: []string{"hooks", "status"}}}})
				totals.ModifiedProjects++
				totals.ChangedParams += result.ChangedCount
			} else if errors.As(err, &cacheErr) {
				result.Status, result.Published, result.Err = RemoteMutationPublishedCacheFailed, true, err
				selector, _ := rctarget.ExactFilter(project.ProjectID)
				machine.FromContext(cmd.Context()).AddWarning(machine.Warning{Code: "publication.cache_stale", Message: "Firebase accepted the publication, but the local cache update failed.", Target: project.ProjectID, Details: struct {
					Stage string `json:"stage"`
				}{Stage: "cache"}, Remediation: []machine.Remediation{{Description: "refresh the successfully published target instead of retrying the mutation", Strategy: machine.RemediationRunCommand, Argv: []string{"get", "--update", "--project", selector}}}})
				totals.ModifiedProjects++
				totals.ChangedParams += result.ChangedCount
			} else if preparationFailed || IsPreparationError(err) {
				result.Status, result.Err = RemoteMutationPreparationFailed, err
			} else if IsRemoteConfigConflict(err) {
				result.Status, result.Err = RemoteMutationConflict, err
			} else if IsValidationError(err) {
				result.Status, result.Err = RemoteMutationValidationFailed, err
			} else {
				result.Status, result.Err = RemoteMutationPublishFailed, err
			}
		case result.ChangedCount == 0:
			result.Status = RemoteMutationUnchanged
			reason := NoOpAlreadyApplied
			result.NoOpReason = &reason
		default:
			result.Status, result.Published = RemoteMutationPublished, true
			if firebase.IsDryRun(ctx) {
				result.Status, result.Published = RemoteMutationWouldPublish, false
			}
			totals.ModifiedProjects++
			totals.ChangedParams += result.ChangedCount
		}
		totals.Results = append(totals.Results, result)
		if batchMustStop(ctx, err) {
			break
		}
	}
	return totals, mutationBatchError(totals, operation)
}

// RemoteMutationJSONError is the stable structured failure payload for a
// direct Remote Config mutation result.
type RemoteMutationJSONError struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

// RemoteMutationJSONResult is the stable per-target automation contract shared
// by add, update, delete, duplicate, condition, and group mutations.
type RemoteMutationJSONResult struct {
	Target           string                   `json:"target"`
	Status           RemoteMutationStatus     `json:"status"`
	ChangedItemCount int                      `json:"changed_item_count"`
	PreviousVersion  *string                  `json:"previous_version"`
	PublishedVersion *string                  `json:"published_version"`
	Draft            bool                     `json:"draft"`
	DryRun           bool                     `json:"dry_run"`
	Validated        bool                     `json:"validated"`
	ValidationSource string                   `json:"validation_source"`
	Error            *RemoteMutationJSONError `json:"error"`
	RetrySelector    *string                  `json:"retry_selector"`
	ChangeNote       *string                  `json:"change_note"`
	Selection        SelectionMetadata        `json:"selection"`
	NoOpReason       *NoOpReason              `json:"no_op_reason"`
}

// SelectionMetadata explains the scope and cardinality resolved for a result.
type SelectionMetadata struct {
	DefaultScope        bool `json:"default_scope"`
	ResolvedTargetCount int  `json:"resolved_target_count"`
	MatchedItemCount    int  `json:"matched_item_count"`
}

// WriteRemoteMutationResults renders a collected batch after command logging
// has finished, keeping outcomes together at the end of the run.
func WriteRemoteMutationResults(cmd *cobra.Command, totals RemoteMutationTotals, operation, publishedEmoji string) error {
	jsonOut, _ := cmd.Flags().GetBool("json")
	if jsonOut {
		draft, _ := cmd.Flags().GetBool("draft")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return writeRemoteMutationJSON(cmd, totals, draft, dryRun)
	}
	if len(totals.Results) == 0 {
		return nil
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Results:")
	for _, result := range totals.Results {
		writeMutationResult(cmd, result, publishedEmoji)
	}
	writeMutationRecoveryHints(cmd, totals, operation)
	return nil
}

func writeRemoteMutationJSON(cmd *cobra.Command, totals RemoteMutationTotals, draft, dryRun bool) error {
	results := make([]RemoteMutationJSONResult, 0, len(totals.Results))
	for _, result := range totals.Results {
		item := RemoteMutationJSONResult{
			Target:           result.Project.ProjectID,
			Status:           result.Status,
			ChangedItemCount: result.ChangedCount,
			Draft:            draft,
			DryRun:           dryRun,
			Validated:        result.Validated,
			ValidationSource: result.ValidationSource,
			ChangeNote:       result.ChangeNote,
			Selection:        SelectionMetadata{DefaultScope: totals.DefaultScope, ResolvedTargetCount: totals.ResolvedTargets, MatchedItemCount: result.MatchedItemCount},
			NoOpReason:       result.NoOpReason,
		}
		if result.PreviousVersion != "" {
			item.PreviousVersion = &result.PreviousVersion
		}
		if result.PublishedVersion != "" {
			item.PublishedVersion = &result.PublishedVersion
		}
		if result.Err != nil {
			item.Error = &RemoteMutationJSONError{
				Stage:   remoteMutationErrorStage(result),
				Message: machine.SafeErrorText(result.Err),
			}
		}
		if result.Err != nil && !result.Published {
			if selector, err := rctarget.ExactFilter(result.Project.ProjectID); err == nil {
				item.RetrySelector = &selector
			}
		}
		results = append(results, item)
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func remoteMutationChangeNote(ctx context.Context) *string {
	value, ok := firebase.ChangeNoteFromContext(ctx)
	if !ok {
		return nil
	}
	return &value
}

func remoteMutationErrorStage(result RemoteMutationResult) string {
	if result.ErrorStage != "" {
		return result.ErrorStage
	}
	switch result.Status {
	case RemoteMutationValidationFailed:
		return "validation"
	case RemoteMutationPublishFailed, RemoteMutationConflict:
		return "publication"
	case RemoteMutationPublishedCacheFailed:
		return "cache"
	case RemoteMutationPublishedHookFailed:
		return "post_publish_hook"
	case RemoteMutationDraftFailed:
		return "draft"
	default:
		return "preparation"
	}
}

func writeMutationResult(cmd *cobra.Command, result RemoteMutationResult, publishedEmoji string) {
	projectID := result.Project.ProjectID
	switch result.Status {
	case RemoteMutationPublished:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s published: %s\n", publishedEmoji, projectID)
	case RemoteMutationDrafted, RemoteMutationWouldDraft:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📝 %s: %s\n", strings.ReplaceAll(string(result.Status), "-", " "), projectID)
	case RemoteMutationWouldPublish:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧪 would publish: %s\n", projectID)
	case RemoteMutationUnchanged:
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "⏭️ unchanged: %s\n", projectID)
	case RemoteMutationPublishedCacheFailed:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "⚠️ published, cache update failed: %s: %v\n", projectID, result.Err)
	case RemoteMutationPublishedHookFailed:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "⚠️ published, post_publish hook failed: %s: %v\n", projectID, result.Err)
	default:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "❌ %s: %s: %v\n", strings.ReplaceAll(string(result.Status), "-", " "), projectID, result.Err)
	}
	out := cmd.OutOrStdout()
	if result.Err != nil {
		out = cmd.ErrOrStderr()
	}
	_, _ = fmt.Fprintf(out, "   validated: %t · validation_source: %s\n", result.Validated, result.ValidationSource)
}

func writeMutationRecoveryHints(cmd *cobra.Command, totals RemoteMutationTotals, operation string) {
	failures := totals.failureCount()
	if failures == 0 {
		return
	}
	if ids := totals.failedProjectIDs(); len(ids) > 0 {
		filters := make([]string, 0, len(ids))
		for _, id := range ids {
			filter, err := rctarget.ExactFilter(id)
			if err == nil {
				filters = append(filters, fmt.Sprintf("-p '%s'", filter))
			}
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Retry template targets that were not %s by rerunning the command with only these target filters:\n  %s\n", map[bool]string{true: "drafted", false: "published"}[operation == "draft"], strings.Join(filters, " "))
	}
	cacheFailed := make([]string, 0)
	for _, result := range totals.Results {
		if result.Status == RemoteMutationPublishedCacheFailed {
			cacheFailed = append(cacheFailed, result.Project.ProjectID)
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
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Firebase was updated, but local caches are stale. Refresh them instead of retrying the mutation:\n  fbrcm get --update %s\n", strings.Join(filters, " "))
	}
}

func mutationBatchError(totals RemoteMutationTotals, operation string) error {
	failures := totals.failureCount()
	if failures == 0 {
		return nil
	}
	if len(totals.Results) == 1 && totals.Results[0].Err != nil && !totals.Results[0].Published {
		return totals.Results[0].Err
	}
	failedTargets := make([]string, 0, failures)
	targetFailures := make([]machine.BatchFailure, 0, failures)
	publishedTargets := 0
	successfulTargets := 0
	for _, result := range totals.Results {
		if result.Published {
			publishedTargets++
		}
		if result.Err != nil {
			failedTargets = append(failedTargets, result.Project.ProjectID)
			targetFailures = append(targetFailures, machine.BatchFailure{Target: result.Project.ProjectID, Err: result.Err})
		}
		if result.Err == nil || result.Published {
			successfulTargets++
		}
	}
	remediation := make([]machine.Remediation, 0, 1)
	if ids := totals.failedProjectIDs(); len(ids) > 0 {
		argv := make([]string, 0, len(ids)*2)
		for _, id := range ids {
			selector, err := rctarget.ExactFilter(id)
			if err == nil {
				argv = append(argv, "--project", selector)
			}
		}
		if len(argv) > 0 {
			remediation = append(remediation, machine.Remediation{Description: "retry only targets that did not complete", Strategy: machine.RemediationRetryWithArguments, Argv: argv})
		}
	}
	return &machine.BatchError{Operation: operation, FailedTargets: failedTargets, Failures: targetFailures, SuccessfulTargetCount: successfulTargets, PublishedTargetCount: publishedTargets, Remediation: remediation, Err: fmt.Errorf("%s failed", display.FormatCount(failures, "template target", "template targets"))}
}

func batchMustStop(ctx context.Context, err error) bool {
	return err != nil && (ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}
