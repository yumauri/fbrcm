package rc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/display"
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
	RemoteMutationDrafted              RemoteMutationStatus = "drafted"
	RemoteMutationWouldDraft           RemoteMutationStatus = "would-draft"
	RemoteMutationWouldPublish         RemoteMutationStatus = "would-publish"
	RemoteMutationDraftFailed          RemoteMutationStatus = "draft-failed"
)

// RemoteMutationResult records one selected project's final outcome.
type RemoteMutationResult struct {
	Project      core.Project
	Status       RemoteMutationStatus
	ChangedCount int
	Published    bool
	Err          error
}

// RemoteMutationTotals contains aggregate counts and ordered project results.
type RemoteMutationTotals struct {
	ModifiedProjects int
	ChangedParams    int
	Results          []RemoteMutationResult
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
type RemoteMutationPlanner func(project core.Project, cfg *ProjectConfig) (RemoteConfigMutation, error)

// RunRemoteDraftLoop applies mutations on top of each project's draft and records
// failures independently. It never writes to Firebase.
func RunRemoteDraftLoop(ctx context.Context, cmd *cobra.Command, svc *core.Core, projects []core.Project, operation string, plan RemoteMutationPlanner) (RemoteMutationTotals, error) {
	var totals RemoteMutationTotals
	for _, project := range projects {
		result := RemoteMutationResult{Project: project}
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
		var mutate RemoteConfigMutation
		if err == nil {
			mutate, err = plan(project, cfg)
		}
		if err == nil && mutate == nil {
			result.Status = RemoteMutationUnchanged
			totals.Results = append(totals.Results, result)
			continue
		}
		var finalCfg *firebase.RemoteConfig
		if err == nil {
			result.ChangedCount, finalCfg, err = mutate(cfg.Config)
		}
		if err == nil && result.ChangedCount == 0 {
			result.Status = RemoteMutationUnchanged
			totals.Results = append(totals.Results, result)
			continue
		}
		var finalRaw []byte
		if err == nil {
			finalRaw, err = firebase.MarshalRemoteConfig(finalCfg)
		}
		if err == nil && !firebase.IsDryRun(ctx) {
			err = svc.SaveDraft(project.ProjectID, finalRaw)
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
	return totals, mutationBatchError(totals)
}

// RunRemotePublishLoop validates and publishes every selected project
// independently. Project-scoped failures are reported and do not stop later
// projects. Conflicts are left for a fresh, explicitly reviewed retry.
func RunRemotePublishLoop(ctx context.Context, cmd *cobra.Command, svc *core.Core, projects []core.Project, operation, publishedEmoji string, plan RemoteMutationPlanner) (RemoteMutationTotals, error) {
	var totals RemoteMutationTotals
	if len(projects) > 1 && !firebase.IsDryRun(ctx) {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: Remote Config is published independently for each project. Some projects may succeed while others fail. For coordinated changes, consider staging with --draft first.")
	}

	for _, project := range projects {
		result := RemoteMutationResult{Project: project}
		preparationFailed := false
		hasDraft, err := svc.HasDraft(project.ProjectID)
		if err == nil && hasDraft {
			err = fmt.Errorf("project has an unpublished draft; use --draft, publish it, or discard it first")
		}
		var cfg *ProjectConfig
		if err == nil {
			cfg, err = RevalidateProjectConfig(ctx, svc, project)
		}
		var mutate RemoteConfigMutation
		if err == nil {
			mutate, err = plan(project, cfg)
		}
		preparationFailed = err != nil
		if err == nil && mutate == nil {
			result.Status = RemoteMutationUnchanged
			totals.Results = append(totals.Results, result)
			continue
		}
		var retry bool
		if err == nil {
			result.ChangedCount, retry, err = PublishProjectConfigMutation(ctx, svc, cfg, operation, nil, mutate)
		}
		switch {
		case retry:
			result.Status = RemoteMutationConflict
			result.Err = fmt.Errorf("remote config changed during %s; rerun the command to review a fresh candidate", operation)
		case err != nil:
			var cacheErr *core.RemoteConfigPublishedCacheError
			if errors.As(err, &cacheErr) {
				result.Status, result.Published, result.Err = RemoteMutationPublishedCacheFailed, true, err
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
	return totals, mutationBatchError(totals)
}

// WriteRemoteMutationResults renders a collected batch after command logging
// has finished, keeping outcomes together at the end of the run.
func WriteRemoteMutationResults(cmd *cobra.Command, totals RemoteMutationTotals, operation, publishedEmoji string) {
	if len(totals.Results) == 0 {
		return
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Results:")
	for _, result := range totals.Results {
		writeMutationResult(cmd, result, publishedEmoji)
	}
	writeMutationRecoveryHints(cmd, totals, operation)
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
	default:
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "❌ %s: %s: %v\n", strings.ReplaceAll(string(result.Status), "-", " "), projectID, result.Err)
	}
}

func writeMutationRecoveryHints(cmd *cobra.Command, totals RemoteMutationTotals, operation string) {
	failures := totals.failureCount()
	if failures == 0 {
		return
	}
	if ids := totals.failedProjectIDs(); len(ids) > 0 {
		filters := make([]string, 0, len(ids))
		for _, id := range ids {
			filters = append(filters, fmt.Sprintf("-p '=%s'", id))
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Retry projects that were not %s by rerunning the command with only these project filters:\n  %s\n", map[bool]string{true: "drafted", false: "published"}[operation == "draft"], strings.Join(filters, " "))
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
			filters = append(filters, fmt.Sprintf("-p '=%s'", id))
		}
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Firebase was updated, but local caches are stale. Refresh them instead of retrying the mutation:\n  fbrcm get --update %s\n", strings.Join(filters, " "))
	}
}

func mutationBatchError(totals RemoteMutationTotals) error {
	failures := totals.failureCount()
	if failures == 0 {
		return nil
	}
	return fmt.Errorf("%s failed", display.FormatCount(failures, "project", "projects"))
}

func batchMustStop(ctx context.Context, err error) bool {
	return err != nil && (ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
}
