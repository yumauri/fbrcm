package plancmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	sharedrc "github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/publication"
)

type targetSummary struct {
	Target           string             `json:"target"`
	Action           publication.Action `json:"action"`
	BaseVersion      string             `json:"base_version"`
	BaseSHA256       string             `json:"base_sha256"`
	CandidateSHA256  string             `json:"candidate_sha256"`
	ValidationSource string             `json:"validation_source" contract:"enum=local|firebase"`
}

type summary struct {
	PlanID             string          `json:"plan_id"`
	CreatedAt          string          `json:"created_at"`
	ProducerVersion    string          `json:"producer_version"`
	CommandID          string          `json:"command_id"`
	ExecutionPolicy    string          `json:"execution_policy" contract:"enum=stateful|stateless"`
	HooksEnabled       bool            `json:"hooks_enabled"`
	TargetCount        int             `json:"target_count"`
	PublishTargetCount int             `json:"publish_target_count"`
	Targets            []targetSummary `json:"targets"`
}

type validationResult struct {
	PlanID      string `json:"plan_id"`
	Valid       bool   `json:"valid"`
	TargetCount int    `json:"target_count"`
}

func New() *cobra.Command {
	cmd := &cobra.Command{Use: "plan", Short: "Inspect immutable publication plans"}
	show := &cobra.Command{Use: "show <plan>", Short: "Show a publication plan and its diffs", Args: cobra.ExactArgs(1), RunE: runShow}
	show.Flags().Bool("json", false, "Print publication plan summary as JSON")
	validate := &cobra.Command{Use: "validate <plan>", Short: "Validate publication plan structure and integrity", Args: cobra.ExactArgs(1), RunE: runValidate}
	validate.Flags().Bool("json", false, "Print validation result as JSON")
	cmd.AddCommand(show, validate)
	contract.RegisterResponse(show, summary{})
	contract.RegisterResponse(validate, validationResult{})
	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	plan, err := shared.ReadPublicationPlan(cmd, args[0])
	if err != nil {
		return err
	}
	result := summarize(plan)
	if contract.Enabled(cmd) {
		return shared.WriteJSON(cmd, result)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Plan: %s\nCreated: %s\nProducer: %s %s\nOperation: %s\nExecution: %s\nTargets: %d to publish, %d unchanged\n", plan.PlanID, plan.CreatedAt.Format("2006-01-02 15:04:05Z07:00"), plan.Producer.Name, plan.Producer.Version, plan.Operation.CommandID, plan.Execution.Policy, result.PublishTargetCount, result.TargetCount-result.PublishTargetCount)
	for _, target := range plan.Targets {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\n%s [%s]\n", target.Target, target.Action)
		base, parseErr := firebase.ParseRemoteConfig(target.Base.RemoteConfig)
		if parseErr != nil {
			return parseErr
		}
		candidate, parseErr := firebase.ParseRemoteConfig(target.Candidate.RemoteConfig)
		if parseErr != nil {
			return parseErr
		}
		diff, changed := sharedrc.RenderRemoteConfigDiff(base, candidate)
		if changed {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), diff)
		} else {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  No changes")
		}
	}
	return nil
}

func runValidate(cmd *cobra.Command, args []string) error {
	plan, err := shared.ReadPublicationPlan(cmd, args[0])
	if err != nil {
		return err
	}
	result := validationResult{PlanID: plan.PlanID, Valid: true, TargetCount: len(plan.Targets)}
	if contract.Enabled(cmd) {
		return shared.WriteJSON(cmd, result)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Publication plan %s is valid (%d targets).\n", plan.PlanID, len(plan.Targets))
	return err
}

func summarize(plan *publication.Plan) summary {
	result := summary{PlanID: plan.PlanID, CreatedAt: plan.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), ProducerVersion: plan.Producer.Version, CommandID: plan.Operation.CommandID, ExecutionPolicy: plan.Execution.Policy, HooksEnabled: plan.Execution.HooksEnabled, TargetCount: len(plan.Targets), Targets: make([]targetSummary, 0, len(plan.Targets))}
	for _, target := range plan.Targets {
		if target.Action == publication.ActionPublish {
			result.PublishTargetCount++
		}
		result.Targets = append(result.Targets, targetSummary{Target: target.Target, Action: target.Action, BaseVersion: target.Base.Version, BaseSHA256: target.Base.SHA256, CandidateSHA256: target.Candidate.SHA256, ValidationSource: target.Validation.Source})
	}
	return result
}
