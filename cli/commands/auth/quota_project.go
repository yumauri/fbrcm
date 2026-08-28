package auth

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

type quotaProjectResult struct {
	AuthID                   string                      `json:"auth_id"`
	ConfiguredQuotaProjectID *string                     `json:"configured_quota_project_id"`
	EffectiveQuotaProjectID  *string                     `json:"effective_quota_project_id"`
	Source                   firebase.QuotaProjectSource `json:"source" contract:"enum=environment|auth|credentials|unresolved"`
	PreviousQuotaProjectID   *string                     `json:"previous_quota_project_id,omitempty"`
	Changed                  bool                        `json:"changed"`
	Status                   string                      `json:"status" contract:"enum=shown|set|unset|unchanged"`
}

func newQuotaProjectCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "quota-project", Short: "Manage auth quota projects"}
	cmd.AddCommand(newQuotaProjectShowCommand(svc), newQuotaProjectSetCommand(svc), newQuotaProjectUnsetCommand(svc))
	return cmd
}

func newQuotaProjectShowCommand(svc *core.Core) *cobra.Command {
	return &cobra.Command{
		Use:   "show <auth-id>",
		Short: "Show an auth identity's quota-project resolution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, selection, err := svc.ResolveAuthQuotaProject(shared.CommandContext(cmd), args[0])
			if err != nil && !isQuotaProjectRequired(err) {
				return err
			}
			return writeQuotaProjectResult(cmd, quotaResult(entry.ID, entry.QuotaProjectID, "", selection, false, "shown"))
		},
	}
}

func newQuotaProjectSetCommand(svc *core.Core) *cobra.Command {
	return &cobra.Command{
		Use:   "set <auth-id> <quota-project-id>",
		Short: "Set an auth identity's quota project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ValidateQuotaProjectID(args[1]); err != nil {
				return shared.InvalidArgument(err)
			}
			entry, previous, changed, err := svc.SetAuthQuotaProject(args[0], args[1])
			if err != nil {
				return err
			}
			_, selection, resolveErr := svc.ResolveAuthQuotaProject(shared.CommandContext(cmd), entry.ID)
			if resolveErr != nil && !isQuotaProjectRequired(resolveErr) {
				return resolveErr
			}
			status := "set"
			if !changed {
				status = "unchanged"
			}
			return writeQuotaProjectResult(cmd, quotaResult(entry.ID, entry.QuotaProjectID, previous, selection, changed, status))
		},
	}
}

func newQuotaProjectUnsetCommand(svc *core.Core) *cobra.Command {
	return &cobra.Command{
		Use:   "unset <auth-id>",
		Short: "Remove an auth identity's quota-project setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, previous, changed, err := svc.SetAuthQuotaProject(args[0], "")
			if err != nil {
				return err
			}
			_, selection, resolveErr := svc.ResolveAuthQuotaProject(shared.CommandContext(cmd), entry.ID)
			if resolveErr != nil && !isQuotaProjectRequired(resolveErr) {
				return resolveErr
			}
			status := "unset"
			if !changed {
				status = "unchanged"
			}
			return writeQuotaProjectResult(cmd, quotaResult(entry.ID, entry.QuotaProjectID, previous, selection, changed, status))
		},
	}
}

func quotaResult(authID, configured, previous string, selection firebase.QuotaProjectSelection, changed bool, status string) quotaProjectResult {
	return quotaProjectResult{
		AuthID: authID, ConfiguredQuotaProjectID: optionalQuotaProject(configured), EffectiveQuotaProjectID: optionalQuotaProject(selection.ProjectID),
		Source: selection.Source, PreviousQuotaProjectID: optionalQuotaProject(previous), Changed: changed, Status: status,
	}
}

func optionalQuotaProject(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func isQuotaProjectRequired(err error) bool {
	var required *firebase.QuotaProjectRequiredError
	return errors.As(err, &required)
}

func writeQuotaProjectResult(cmd *cobra.Command, result quotaProjectResult) error {
	if contract.Enabled(cmd) {
		return shared.WriteJSON(cmd, result)
	}
	configured := "—"
	if result.ConfiguredQuotaProjectID != nil {
		configured = *result.ConfiguredQuotaProjectID
	}
	effective := "unresolved"
	if result.EffectiveQuotaProjectID != nil {
		effective = *result.EffectiveQuotaProjectID
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "Auth: %s\nConfigured quota project: %s\nEffective quota project: %s\nSource: %s\n", result.AuthID, configured, effective, result.Source)
	return err
}
