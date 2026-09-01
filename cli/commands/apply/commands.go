package applycmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/cli/shared"
	sharedrc "github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/publication"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

type Status string

const (
	statusUnchanged            Status = "unchanged"
	statusWouldPublish         Status = "would-publish"
	statusPublished            Status = "published"
	statusAlreadyApplied       Status = "already-applied"
	statusConflict             Status = "conflict"
	statusPublishFailed        Status = "publish-failed"
	statusPublishedHookFailed  Status = "published-hook-failed"
	statusPublishedCacheFailed Status = "published-cache-failed"
)

type targetError struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

type targetResult struct {
	Target           string       `json:"target"`
	Status           Status       `json:"status"`
	PreviousVersion  string       `json:"previous_version,omitempty"`
	PublishedVersion string       `json:"published_version,omitempty"`
	Validated        bool         `json:"validated"`
	ValidationSource string       `json:"validation_source" contract:"enum=local|firebase"`
	Error            *targetError `json:"error,omitempty"`
}

type result struct {
	PlanID    string         `json:"plan_id"`
	DryRun    bool           `json:"dry_run"`
	Published int            `json:"published_count"`
	Items     []targetResult `json:"items"`
}

type preparedTarget struct {
	plan    publication.Target
	current json.RawMessage
	etag    string
	result  targetResult
}

func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "apply <plan>", Short: "Apply an immutable publication plan", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { return run(cmd, svc, args[0]) }}
	shared.AddDryRunFlag(cmd)
	shared.AddYesFlag(cmd, "Apply the publication plan without confirmation")
	cmd.Flags().Bool("json", false, "Print publication results as JSON")
	contract.RegisterResponse(cmd, result{})
	return cmd
}

func run(cmd *cobra.Command, svc *core.Core, path string) error {
	ctx := shared.CommandContext(cmd)
	plan, err := shared.ReadPublicationPlan(cmd, path)
	if err != nil {
		return err
	}
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	if dryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	ctx, err = shared.WithChangeNote(ctx, plan.Operation.ChangeNote)
	if err != nil {
		return err
	}
	environment, err := core.PublicationEnvironmentForContext(ctx)
	if err != nil {
		return err
	}
	if environment.Policy != plan.Execution.Policy || environment.HooksEnabled != plan.Execution.HooksEnabled || environment.HookDefinitionSHA256 != plan.Execution.HookDefinitionSHA256 {
		return &machine.ConflictError{Code: "plan.stale", Resource: "publication_environment", Target: plan.PlanID, Err: fmt.Errorf("publication environment differs from the environment recorded in the plan")}
	}
	publishTargets := 0
	for _, target := range plan.Targets {
		if target.Action == publication.ActionPublish {
			publishTargets++
		}
	}
	addNonAtomicWarning(cmd, publishTargets, dryRun)
	targetIDs := make([]string, 0, len(plan.Targets))
	for _, target := range plan.Targets {
		if target.Action == publication.ActionPublish {
			targetIDs = append(targetIDs, target.Target)
		}
	}
	if len(targetIDs) > 0 {
		ctx, err = shared.FirebaseServicesContextForExecution(ctx, targetIDs)
		if err != nil {
			return err
		}
	}
	prepared, output, err := preflight(ctx, svc, plan)
	if err != nil {
		return err
	}
	if len(prepared) == 0 {
		if !dryRun {
			cleanupAlreadyAppliedDrafts(cmd, svc, plan, output.Items)
		}
		return writeResult(cmd, output)
	}

	diff := renderDiffs(prepared)
	confirmed, err := shared.PrintDiffAndConfirm(cmd, yes, cmd.ErrOrStderr(), diff, fmt.Sprintf("Apply publication plan %s?", plan.PlanID), false)
	if err != nil || !confirmed {
		return err
	}
	for index := range prepared {
		item := &prepared[index]
		itemCtx, contextErr := targetContext(ctx, item.plan)
		if contextErr != nil {
			return contextErr
		}
		if err := svc.ValidatePublicationCandidate(itemCtx, item.plan.Target, item.current, item.plan.Candidate.RemoteConfig, item.etag, "plan-apply"); err != nil {
			return err
		}
		item.result.Validated = true
		item.result.ValidationSource = core.ValidationSourceFirebase
	}
	if dryRun {
		for _, item := range prepared {
			output.Items = append(output.Items, targetResult{Target: item.plan.Target, Status: statusWouldPublish, PreviousVersion: item.plan.Base.Version, Validated: true, ValidationSource: core.ValidationSourceFirebase})
		}
		return writeResult(cmd, output)
	}
	cleanupAlreadyAppliedDrafts(cmd, svc, plan, output.Items)

	var failures []machine.BatchFailure
	for _, item := range prepared {
		itemCtx, contextErr := targetContext(ctx, item.plan)
		if contextErr != nil {
			return contextErr
		}
		publishCtx := core.WithPublicationPreHooksComplete(itemCtx)
		publishedRaw, _, publishErr := svc.PublishRemoteConfigWithETag(publishCtx, item.plan.Target, item.plan.Candidate.RemoteConfig, item.etag)
		entry, accepted := classifyPublishResult(cmd, item.plan.Target, item.result, publishedRaw, publishErr)
		if publishErr != nil {
			if accepted {
				output.Published++
			}
			failures = append(failures, machine.BatchFailure{Target: item.plan.Target, Err: publishErr})
		} else {
			output.Published++
			cleanupMatchingDraft(cmd, svc, item.plan)
		}
		output.Items = append(output.Items, entry)
	}
	if err := writeResult(cmd, output); err != nil {
		return err
	}
	if len(failures) > 0 {
		failed := make([]string, 0, len(failures))
		for _, failure := range failures {
			failed = append(failed, failure.Target)
		}
		return &machine.BatchError{Operation: "plan apply", FailedTargets: failed, Failures: failures, SuccessfulTargetCount: len(prepared) - len(failures), PublishedTargetCount: output.Published}
	}
	return nil
}

func addNonAtomicWarning(cmd *cobra.Command, publishTargets int, dryRun bool) {
	if publishTargets <= 1 || dryRun {
		return
	}
	if !contract.Enabled(cmd) {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: plan targets are published independently. Successful publications are not rolled back if another target fails.")
	}
	shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.non_atomic", Message: "Plan targets are published independently; successful publications are not rolled back when another target fails.", Details: struct {
		TargetCount int `json:"target_count"`
	}{TargetCount: publishTargets}})
}

func classifyPublishResult(cmd *cobra.Command, target string, entry targetResult, publishedRaw json.RawMessage, publishErr error) (targetResult, bool) {
	entry.Status = statusPublished
	entry.Validated = true
	entry.ValidationSource = core.ValidationSourceFirebase
	if len(publishedRaw) > 0 {
		if cfg, parseErr := firebase.ParseRemoteConfig(publishedRaw); parseErr == nil {
			entry.PublishedVersion = cfg.Version.VersionNumber
		}
	}
	if publishErr == nil {
		return entry, true
	}
	entry.Error = &targetError{Stage: "publication", Message: machine.SafeErrorText(publishErr)}
	var hookErr *core.RemoteConfigPublishedHookError
	var cacheErr *core.RemoteConfigPublishedCacheError
	switch {
	case errors.As(publishErr, &hookErr):
		entry.Status = statusPublishedHookFailed
		entry.Error.Stage = "post_publish_hook"
		shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.post_publish_hook_failed", Message: "Firebase accepted the plan target, but a post_publish hook failed.", Target: target, Details: struct {
			Stage string `json:"stage"`
		}{Stage: "post_publish_hook"}, Remediation: []shared.Remediation{{Description: "inspect hook trust and status without republishing", Strategy: shared.RemediationRunCommand, Argv: []string{"hooks", "status"}}}})
		return entry, true
	case errors.As(publishErr, &cacheErr):
		entry.Status = statusPublishedCacheFailed
		entry.Error.Stage = "cache"
		selector, _ := rctarget.ExactFilter(target)
		shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.cache_stale", Message: "Firebase accepted the plan target, but the local cache update failed.", Target: target, Details: struct {
			Stage string `json:"stage"`
		}{Stage: "cache"}, Remediation: []shared.Remediation{{Description: "refresh the successfully published target instead of reapplying the plan", Strategy: shared.RemediationRunCommand, Argv: []string{"get", "--update", "--project", selector}}}})
		return entry, true
	case sharedrc.IsRemoteConfigConflict(publishErr):
		entry.Status = statusConflict
	default:
		entry.Status = statusPublishFailed
	}
	return entry, false
}

func planTarget(plan *publication.Plan, targetID string) publication.Target {
	for _, target := range plan.Targets {
		if target.Target == targetID {
			return target
		}
	}
	return publication.Target{}
}

func cleanupAlreadyAppliedDrafts(cmd *cobra.Command, svc *core.Core, plan *publication.Plan, items []targetResult) {
	for _, item := range items {
		if item.Status == statusAlreadyApplied {
			cleanupMatchingDraft(cmd, svc, planTarget(plan, item.Target))
		}
	}
}

func cleanupMatchingDraft(cmd *cobra.Command, svc *core.Core, target publication.Target) {
	if target.Source.Kind != "draft" || target.Source.Fingerprint == "" {
		return
	}
	record, exists, err := svc.LoadDraftRecord(target.Target)
	if err != nil {
		shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.draft_cleanup_failed", Message: "The plan was applied, but its source draft could not be inspected.", Target: target.Target, Details: struct {
			Stage string `json:"stage"`
		}{Stage: "cleanup"}})
		return
	}
	if !exists {
		return
	}
	if record.UpdatedAt.UTC().Format(time.RFC3339Nano) != target.Source.Fingerprint {
		shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "plan.source_draft_changed", Message: "The source draft changed after planning and was preserved.", Target: target.Target, Details: struct {
			Stage string `json:"stage"`
		}{Stage: "source_draft"}})
		return
	}
	if err := svc.DeleteDraft(target.Target); err != nil {
		shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "publication.draft_cleanup_failed", Message: "The plan was applied, but its matching source draft could not be removed.", Target: target.Target, Details: struct {
			Stage string `json:"stage"`
		}{Stage: "cleanup"}})
	}
}

func targetContext(ctx context.Context, target publication.Target) (context.Context, error) {
	if target.ChangeNote == nil {
		return ctx, nil
	}
	return firebase.WithChangeNote(ctx, *target.ChangeNote)
}

func preflight(ctx context.Context, svc *core.Core, plan *publication.Plan) ([]preparedTarget, result, error) {
	prepared := make([]preparedTarget, 0)
	output := result{PlanID: plan.PlanID, DryRun: firebase.IsDryRun(ctx), Items: make([]targetResult, 0, len(plan.Targets))}
	stale := make([]string, 0)
	for _, target := range plan.Targets {
		if target.Action == publication.ActionNone {
			output.Items = append(output.Items, targetResult{Target: target.Target, Status: statusUnchanged, PreviousVersion: target.Base.Version, Validated: true, ValidationSource: target.Validation.Source})
			continue
		}
		cache, _, err := svc.RevalidateParameters(ctx, target.Target)
		if err != nil {
			return nil, output, err
		}
		item, entry, targetStale, err := classifyPreflightTarget(target, cache.RemoteConfig, cache.ETag)
		if err != nil {
			return nil, output, err
		}
		if entry != nil {
			output.Items = append(output.Items, *entry)
			continue
		}
		if targetStale {
			stale = append(stale, target.Target)
			continue
		}
		prepared = append(prepared, *item)
	}
	if len(stale) > 0 {
		return nil, output, &machine.ConflictError{Code: "plan.stale", Resource: "remote_config", Target: strings.Join(stale, ","), Retryable: false, Err: fmt.Errorf("publication plan is stale for %s; generate and review a new plan", strings.Join(stale, ", "))}
	}
	return prepared, output, nil
}

func classifyPreflightTarget(target publication.Target, current json.RawMessage, etag string) (*preparedTarget, *targetResult, bool, error) {
	currentDigest, err := publication.RemoteConfigDigest(current)
	if err != nil {
		return nil, nil, false, err
	}
	baseCfg, err := firebase.ParseRemoteConfig(current)
	if err != nil {
		return nil, nil, false, err
	}
	entry := targetResult{Target: target.Target, PreviousVersion: baseCfg.Version.VersionNumber}
	if currentDigest == target.Candidate.SHA256 {
		entry.Status = statusAlreadyApplied
		entry.Validated = true
		entry.ValidationSource = core.ValidationSourceLocal
		return nil, &entry, false, nil
	}
	if etag != target.Base.ETag || currentDigest != target.Base.SHA256 {
		return nil, nil, true, nil
	}
	return &preparedTarget{plan: target, current: current, etag: etag, result: entry}, nil, false, nil
}

func renderDiffs(prepared []preparedTarget) string {
	var builder strings.Builder
	for _, item := range prepared {
		base, _ := firebase.ParseRemoteConfig(item.current)
		candidate, _ := firebase.ParseRemoteConfig(item.plan.Candidate.RemoteConfig)
		diff, _ := sharedrc.RenderRemoteConfigDiff(base, candidate)
		_, _ = fmt.Fprintf(&builder, "\n%s\n%s", item.plan.Target, diff)
	}
	return builder.String()
}

func writeResult(cmd *cobra.Command, value result) error {
	if contract.Enabled(cmd) {
		return shared.WriteJSON(cmd, value)
	}
	for _, item := range value.Items {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", item.Target, item.Status)
	}
	return nil
}
