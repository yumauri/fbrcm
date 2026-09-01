package rc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/rc/publication"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

// PlanCreatedResult is the machine result returned by mutation commands in
// --plan-out mode.
type PlanCreatedResult struct {
	PlanID               string                   `json:"plan_id"`
	Path                 string                   `json:"path"`
	Artifact             publication.FileArtifact `json:"artifact"`
	CreatedAt            time.Time                `json:"created_at"`
	CommandID            string                   `json:"command_id"`
	TargetCount          int                      `json:"target_count"`
	PublishTargetCount   int                      `json:"publish_target_count"`
	UnchangedTargetCount int                      `json:"unchanged_target_count"`
}

func runRemotePlanLoop(ctx context.Context, cmd *cobra.Command, svc *core.Core, projects []core.Project, defaultScope bool, operation, path string, planner RemoteMutationPlanner) (RemoteMutationTotals, error) {
	totals := RemoteMutationTotals{DefaultScope: defaultScope, ResolvedTargets: len(projects)}
	environment, err := core.PublicationEnvironmentForContext(ctx)
	if err != nil {
		return totals, err
	}
	plan := publication.New(cmd.Root().Version, operation, environment.Policy, remoteMutationChangeNote(ctx))
	plan.Execution.HooksEnabled = environment.HooksEnabled
	plan.Execution.HookDefinitionSHA256 = environment.HookDefinitionSHA256
	selection, err := json.Marshal(SelectionMetadata{DefaultScope: defaultScope, ResolvedTargetCount: len(projects)})
	if err != nil {
		return totals, err
	}
	plan.Operation.Selection = selection

	for _, project := range projects {
		progress.Start("Preparing publication plan for " + project.ProjectID + "…")
		if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
			hasDraft, draftErr := svc.HasDraft(project.ProjectID)
			if draftErr != nil {
				return totals, draftErr
			}
			if hasDraft {
				return totals, &machine.ConflictError{Code: "draft.exists", Resource: "draft", Target: project.ProjectID, Err: fmt.Errorf("project has an unpublished draft; create the plan from draft publish instead")}
			}
		}
		cfg, err := RevalidateProjectConfig(ctx, svc, project)
		if err != nil {
			return totals, err
		}
		mutationPlan, err := planner(project, cfg)
		if err != nil {
			return totals, err
		}
		candidate := append(json.RawMessage(nil), cfg.Cache.RemoteConfig...)
		changedCount := 0
		if mutationPlan.Mutation != nil {
			var finalCfg *firebase.RemoteConfig
			var mutationErr error
			changedCount, finalCfg, mutationErr = mutationPlan.Mutation(cfg.Config)
			if mutationErr != nil {
				return totals, mutationErr
			}
			if err := rcmutate.EnsureOpaqueValuesUnchanged(cfg.Config, finalCfg); err != nil {
				return totals, err
			}
			candidate, err = firebase.MarshalRemoteConfig(finalCfg)
			if err != nil {
				return totals, err
			}
		}
		targetAction := publication.ActionNone
		validationSource := core.ValidationSourceLocal
		if changedCount > 0 {
			targetAction = publication.ActionPublish
			validationSource = core.ValidationSourceFirebase
			progress.Start("Validating publication plan for " + project.ProjectID + "…")
			if err := svc.ValidatePublicationCandidate(ctx, project.ProjectID, cfg.Cache.RemoteConfig, candidate, cfg.Cache.ETag, operation); err != nil {
				return totals, err
			}
			totals.ModifiedProjects++
			totals.ChangedParams += changedCount
		}
		target, err := rctarget.Parse(project.ProjectID)
		if err != nil {
			return totals, err
		}
		plan.Targets = append(plan.Targets, publication.Target{
			Target: project.ProjectID, ProjectID: target.ProjectID, Template: string(target.Kind), Action: targetAction, ChangeNote: remoteMutationChangeNote(ctx),
			Base:       publication.Snapshot{Version: cfg.Config.Version.VersionNumber, ETag: cfg.Cache.ETag, RemoteConfig: append(json.RawMessage(nil), cfg.Cache.RemoteConfig...)},
			Candidate:  publication.Snapshot{RemoteConfig: candidate},
			Validation: publication.Validation{Source: validationSource, ValidatedAt: time.Now().UTC()},
			Source:     publication.Source{Kind: "direct"},
		})
	}
	if err := publication.Seal(plan); err != nil {
		return totals, err
	}
	raw, err := publication.Marshal(plan)
	if err != nil {
		return totals, err
	}
	if err := createPlanFile(path, raw); err != nil {
		return totals, err
	}
	result := &PlanCreatedResult{PlanID: plan.PlanID, Path: path, Artifact: publication.NewFileArtifact(path, raw), CreatedAt: plan.CreatedAt, CommandID: operation, TargetCount: len(plan.Targets)}
	for _, target := range plan.Targets {
		if target.Action == publication.ActionPublish {
			result.PublishTargetCount++
		} else {
			result.UnchangedTargetCount++
		}
	}
	totals.Plan = result
	return totals, nil
}

// WritePublicationPlan seals and writes one immutable plan, then emits the
// command result in human or machine form.
func WritePublicationPlan(cmd *cobra.Command, plan *publication.Plan, path string) (*PlanCreatedResult, error) {
	if err := publication.Seal(plan); err != nil {
		return nil, err
	}
	raw, err := publication.Marshal(plan)
	if err != nil {
		return nil, err
	}
	if err := createPlanFile(path, raw); err != nil {
		return nil, err
	}
	result := &PlanCreatedResult{PlanID: plan.PlanID, Path: path, Artifact: publication.NewFileArtifact(path, raw), CreatedAt: plan.CreatedAt, CommandID: plan.Operation.CommandID, TargetCount: len(plan.Targets)}
	for _, target := range plan.Targets {
		if target.Action == publication.ActionPublish {
			result.PublishTargetCount++
		} else {
			result.UnchangedTargetCount++
		}
	}
	jsonOut, _ := cmd.Flags().GetBool("json")
	if jsonOut {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			return nil, err
		}
	} else {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Publication plan written to %s\nPlan ID: %s\nTargets: %d to publish, %d unchanged\n", result.Path, result.PlanID, result.PublishTargetCount, result.UnchangedTargetCount); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func createPlanFile(path string, raw []byte) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create publication plan directory: %w", err)
		}
	}
	if err := config.WritePrivateFileExclusive(path, raw); err != nil {
		if errors.Is(err, os.ErrExist) {
			return &machine.ConflictError{Code: "plan.exists", Resource: "publication_plan", Target: path, Err: fmt.Errorf("publication plan already exists: %s", path)}
		}
		return fmt.Errorf("write publication plan: %w", err)
	}
	return nil
}
