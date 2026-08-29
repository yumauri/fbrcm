package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// New constructs auth command.
func New(svc *core.Core) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage auth identities",
	}
	authCmd.AddCommand(newListCommand(svc), newAddCommand(svc), newLoginCommand(svc), newPathCommand(svc), newDeleteCommand(svc), newBindCommand(svc), newQuotaProjectCommand(svc))
	contract.MustRegisterResponsePath(authCmd, "list", []authListItem{})
	for _, path := range []string{"add google", "add oauth", "add service-account", "add gcloud"} {
		contract.MustRegisterResponsePath(authCmd, path, authMutationResult{})
	}
	contract.MustRegisterResponsePath(authCmd, "login", authLoginResult{})
	contract.MustRegisterResponsePath(authCmd, "path", authPathResult{})
	contract.MustRegisterResponsePath(authCmd, "delete", authDeleteResult{})
	contract.MustRegisterResponsePath(authCmd, "bind", authBindResult{})
	for _, path := range []string{"quota-project show", "quota-project set", "quota-project unset"} {
		contract.MustRegisterResponsePath(authCmd, path, quotaProjectResult{})
	}
	return authCmd
}

func newListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List auth identities",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			entries, defaultAuthID, err := svc.ListAuth()
			if err != nil {
				return err
			}
			if jsonOut {
				return shared.WriteJSON(cmd, newAuthListItems(entries, defaultAuthID))
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderAuthTable(entries, defaultAuthID, shared.TerminalWidth()))
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print auth identities as JSON")
	return cmd
}

func newAddCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add auth identity",
	}
	cmd.AddCommand(newAddGoogleCommand(svc), newAddOAuthCommand(svc), newAddServiceAccountCommand(svc), newAddGCloudCommand(svc))
	return cmd
}

func newAddGoogleCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "google <auth-id>",
		Short: "Add Google OAuth auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ValidateAuthID(args[0]); err != nil {
				return shared.InvalidArgument(err)
			}
			label, err := cmd.Flags().GetString("label")
			if err != nil {
				return err
			}
			var entry config.AuthEntry
			if cmd.Flags().Changed("quota-project") {
				quotaProjectID, flagErr := cmd.Flags().GetString("quota-project")
				if flagErr != nil {
					return flagErr
				}
				entry, err = svc.AddGoogleAuthWithQuotaProject(args[0], label, quotaProjectID)
			} else {
				entry, err = svc.AddGoogleAuth(args[0], label)
			}
			if err != nil {
				return err
			}
			_, paths, err := svc.AuthPaths(entry.ID)
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, newAuthMutationResult(entry, "added", paths))
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "oauth client: built into fbrcm")
			return nil
		},
	}
	cmd.Flags().String("label", "", "Auth identity label")
	cmd.Flags().String("quota-project", "", "Persist the Google Cloud quota project for this auth identity")
	return cmd
}

func newAddOAuthCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oauth <auth-id>",
		Short: "Add OAuth auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ValidateAuthID(args[0]); err != nil {
				return shared.InvalidArgument(err)
			}
			fromPath, err := cmd.Flags().GetString("from")
			if err != nil {
				return err
			}
			label, err := cmd.Flags().GetString("label")
			if err != nil {
				return err
			}
			data, err := shared.ReadJSONInput(cmd, fromPath, "client secret", shared.ErrNoJSONSelection)
			if err != nil {
				return err
			}
			if err := firebase.ValidateOAuthClientSecret(data); err != nil {
				return &shared.ValidationError{Code: "auth.credentials_invalid", Source: "auth", Stage: "input", Target: args[0], Err: err}
			}
			var entry config.AuthEntry
			if cmd.Flags().Changed("quota-project") {
				quotaProjectID, flagErr := cmd.Flags().GetString("quota-project")
				if flagErr != nil {
					return flagErr
				}
				entry, err = svc.AddOAuthAuthWithQuotaProject(args[0], label, data, quotaProjectID)
			} else {
				entry, err = svc.AddOAuthAuth(args[0], label, data)
			}
			if err != nil {
				return err
			}
			_, paths, err := svc.AuthPaths(entry.ID)
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, newAuthMutationResult(entry, "added", paths))
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "secret: %s\n", paths.ClientSecretPath)
			return nil
		},
	}
	cmd.Flags().String("from", "", "Import OAuth client secret from file path; if omitted, read stdin or open file picker")
	cmd.Flags().String("label", "", "Auth identity label")
	cmd.Flags().String("quota-project", "", "Persist the Google Cloud quota project for this auth identity")
	return cmd
}

func newAddServiceAccountCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service-account <auth-id>",
		Short: "Add service account auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ValidateAuthID(args[0]); err != nil {
				return shared.InvalidArgument(err)
			}
			fromPath, err := cmd.Flags().GetString("from")
			if err != nil {
				return err
			}
			label, err := cmd.Flags().GetString("label")
			if err != nil {
				return err
			}
			data, err := shared.ReadJSONInput(cmd, fromPath, "service account key", shared.ErrNoJSONSelection)
			if err != nil {
				return err
			}
			if err := firebase.ValidateServiceAccountKey(data); err != nil {
				return &shared.ValidationError{Code: "auth.credentials_invalid", Source: "auth", Stage: "input", Target: args[0], Err: err}
			}
			var entry config.AuthEntry
			if cmd.Flags().Changed("quota-project") {
				quotaProjectID, flagErr := cmd.Flags().GetString("quota-project")
				if flagErr != nil {
					return flagErr
				}
				entry, err = svc.AddServiceAccountAuthWithQuotaProject(args[0], label, data, quotaProjectID)
			} else {
				entry, err = svc.AddServiceAccountAuth(args[0], label, data)
			}
			if err != nil {
				return err
			}
			_, paths, err := svc.AuthPaths(entry.ID)
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, newAuthMutationResult(entry, "added", paths))
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "service account: %s\n", paths.ServiceAccountPath)
			return nil
		},
	}
	cmd.Flags().String("from", "", "Import service account key from file path; if omitted, read stdin or open file picker")
	cmd.Flags().String("label", "", "Auth identity label")
	cmd.Flags().String("quota-project", "", "Persist the Google Cloud quota project for this auth identity")
	return cmd
}

func newAddGCloudCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcloud <auth-id>",
		Short: "Add gcloud ADC auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ValidateAuthID(args[0]); err != nil {
				return shared.InvalidArgument(err)
			}
			label, err := cmd.Flags().GetString("label")
			if err != nil {
				return err
			}
			var entry config.AuthEntry
			if cmd.Flags().Changed("quota-project") {
				quotaProjectID, flagErr := cmd.Flags().GetString("quota-project")
				if flagErr != nil {
					return flagErr
				}
				entry, err = svc.AddGCloudAuthWithQuotaProject(args[0], label, quotaProjectID)
			} else {
				entry, err = svc.AddGCloudAuth(args[0], label)
			}
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, newAuthMutationResult(entry, "added", core.AuthPaths{}))
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "adc: application default credentials")
			return nil
		},
	}
	cmd.Flags().String("label", "", "Auth identity label")
	cmd.Flags().String("quota-project", "", "Persist the Google Cloud quota project for this auth identity")
	return cmd
}

func newLoginCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login <auth-id>",
		Short: "Authenticate auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noOpen, err := cmd.Flags().GetBool("noopen")
			if err != nil {
				return err
			}
			if err := svc.EnsureAuthLogin(shared.CommandContext(cmd), args[0], noOpen); err != nil {
				return err
			}
			auth, paths, err := svc.AuthPaths(args[0])
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, authLoginResult{AuthID: auth.ID, Type: auth.Type, Status: "authenticated", Paths: authPathPayload(auth, paths)})
			}
			switch auth.Type {
			case config.AuthTypeOAuth, config.AuthTypeGoogle:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔑 authenticated: %s\n", paths.TokenPath)
			case config.AuthTypeServiceAccount:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔑 authenticated: %s\n", paths.ServiceAccountPath)
			case config.AuthTypeGCloud:
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🔑 authenticated: application default credentials")
			default:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔑 authenticated: %s\n", auth.ID)
			}
			return nil
		},
	}
	cmd.Flags().Bool("noopen", false, "Do not open browser automatically")
	return cmd
}

func newPathCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path <auth-id>",
		Short: "Print auth file paths",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			auth, paths, err := svc.AuthPaths(args[0])
			if err != nil {
				return err
			}
			payload := authPathPayload(auth, paths)
			if jsonOut {
				return shared.WriteJSON(cmd, payload)
			}
			for _, path := range authPathLines(auth, paths) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print paths as JSON")
	return cmd
}

func newDeleteCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <auth-id>",
		Short: "Delete an auth identity and its files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				if err := shared.RequireYesInMachineMode(cmd, yes, "deleting auth identity "+args[0], true); err != nil {
					return err
				}
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Delete auth identity %s and its files?", args[0]),
					shared.ConfirmationOptions{Destructive: true},
				)
				confirm.Input = cmd.InOrStdin()
				confirm.Output = cmd.ErrOrStderr()
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}
			auth, paths, err := svc.DeleteAuth(args[0])
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, authDeleteResult{AuthID: auth.ID, Type: auth.Type, Status: "deleted", DeletedPaths: authPathLines(auth, paths)})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 deleted auth: %s\n", auth.ID)
			for _, path := range authPathLines(auth, paths) {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 deleted: %s\n", path)
			}
			return nil
		},
	}
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}

func newBindCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Bind projects to auth identity",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			authID, err := cmd.Flags().GetString("auth")
			if err != nil {
				return err
			}
			if authID == "" {
				return shared.InvalidArgument(fmt.Errorf("--auth is required"))
			}
			filters, err := cmd.Flags().GetStringArray("project")
			if err != nil {
				return err
			}
			if err := shared.RejectTemplateProjectFilters(filters); err != nil {
				return err
			}
			result, err := svc.BindProjectsAuth(filters, authID)
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				items := make([]authBindItem, 0, len(result.Bound)+len(result.Skipped))
				for _, project := range result.Bound {
					items = append(items, authBindItem{ProjectID: project.ProjectID, Status: "bound", Reason: nil})
				}
				for _, skipped := range result.Skipped {
					reason := skipped.Reason
					items = append(items, authBindItem{ProjectID: skipped.Project.ProjectID, Status: "skipped", Reason: &reason})
				}
				return shared.WriteJSON(cmd, authBindResult{AuthID: authID, Bound: len(result.Bound), Skipped: len(result.Skipped), Items: items})
			}
			for _, project := range result.Bound {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔗 bound: %s -> %s\n", project.ProjectID, authID)
			}
			logger := corelog.For("auth")
			for _, skipped := range result.Skipped {
				logger.Error("project auth bind skipped", "project_id", skipped.Project.ProjectID, "auth_id", authID, "err", skipped.Reason)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Summary: %d bound, %d skipped\n", len(result.Bound), len(result.Skipped))
			return nil
		},
	}
	cmd.Flags().String("auth", "", "Auth id to bind")
	_ = cmd.MarkFlagRequired("auth")
	shared.AddProjectFilterFlag(cmd)
	return cmd
}

type authMutationResult struct {
	AuthID         string         `json:"auth_id"`
	Type           string         `json:"type" contract:"enum=google|oauth|service-account|gcloud"`
	Label          string         `json:"label"`
	QuotaProjectID string         `json:"quota_project_id,omitempty"`
	Status         string         `json:"status" contract:"enum=added"`
	Paths          authPathResult `json:"paths"`
}

func newAuthMutationResult(entry config.AuthEntry, status string, paths core.AuthPaths) authMutationResult {
	return authMutationResult{AuthID: entry.ID, Type: entry.Type, Label: entry.Label, QuotaProjectID: entry.QuotaProjectID, Status: status, Paths: authPathPayload(entry, paths)}
}

type authLoginResult struct {
	AuthID string         `json:"auth_id"`
	Type   string         `json:"type" contract:"enum=google|oauth|service-account|gcloud"`
	Status string         `json:"status" contract:"enum=authenticated"`
	Paths  authPathResult `json:"paths"`
}

type authDeleteResult struct {
	AuthID       string   `json:"auth_id"`
	Type         string   `json:"type" contract:"enum=google|oauth|service-account|gcloud"`
	Status       string   `json:"status" contract:"enum=deleted"`
	DeletedPaths []string `json:"deleted_paths"`
}

type authBindItem struct {
	ProjectID string  `json:"project_id"`
	Status    string  `json:"status" contract:"enum=bound|skipped"`
	Reason    *string `json:"reason"`
}

type authBindResult struct {
	AuthID  string         `json:"auth_id"`
	Bound   int            `json:"bound"`
	Skipped int            `json:"skipped"`
	Items   []authBindItem `json:"items"`
}
