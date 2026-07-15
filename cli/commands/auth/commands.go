package auth

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
)

// New constructs auth command.
func New(svc *core.Core) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage auth identities",
	}
	authCmd.AddCommand(newListCommand(svc), newAddCommand(svc), newLoginCommand(svc), newPathCommand(svc), newPurgeCommand(svc), newBindCommand(svc))
	return authCmd
}

func newListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List auth identities",
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
				return shared.WriteJSON(cmd, map[string]any{"default_auth_id": defaultAuthID, "auth": entries})
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderAuthTable(entries, defaultAuthID))
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
	cmd.AddCommand(newAddOAuthCommand(svc), newAddServiceAccountCommand(svc), newAddGCloudCommand(svc))
	return cmd
}

func newAddOAuthCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oauth <auth-id>",
		Short: "Add OAuth auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			entry, err := svc.AddOAuthAuth(args[0], label, data)
			if err != nil {
				return err
			}
			_, paths, err := svc.AuthPaths(entry.ID)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "secret: %s\n", paths.ClientSecretPath)
			return nil
		},
	}
	cmd.Flags().String("from", "", "Import OAuth client secret from file path; if omitted, read stdin or open file picker")
	cmd.Flags().String("label", "", "Auth identity label")
	return cmd
}

func newAddServiceAccountCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service-account <auth-id>",
		Short: "Add service account auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			entry, err := svc.AddServiceAccountAuth(args[0], label, data)
			if err != nil {
				return err
			}
			_, paths, err := svc.AuthPaths(entry.ID)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "service account: %s\n", paths.ServiceAccountPath)
			return nil
		},
	}
	cmd.Flags().String("from", "", "Import service account key from file path; if omitted, read stdin or open file picker")
	cmd.Flags().String("label", "", "Auth identity label")
	return cmd
}

func newAddGCloudCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcloud <auth-id>",
		Short: "Add gcloud ADC auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label, err := cmd.Flags().GetString("label")
			if err != nil {
				return err
			}
			entry, err := svc.AddGCloudAuth(args[0], label)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "adc: application default credentials")
			return nil
		},
	}
	cmd.Flags().String("label", "", "Auth identity label")
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
			if err := svc.EnsureAuthLogin(context.Background(), args[0], noOpen); err != nil {
				return err
			}
			auth, paths, err := svc.AuthPaths(args[0])
			if err != nil {
				return err
			}
			switch auth.Type {
			case config.AuthTypeOAuth:
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

func newPurgeCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge <auth-id>",
		Short: "Delete auth identity files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Delete auth identity %s and its files?", args[0]),
					shared.ConfirmationOptions{Destructive: true},
				)
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}
			auth, paths, err := svc.PurgeAuth(args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged auth: %s\n", auth.ID)
			for _, path := range authPathLines(auth, paths) {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged: %s\n", path)
			}
			return nil
		},
	}
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}

func newBindCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind <project-query>",
		Short: "Bind projects to auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			authID, err := cmd.Flags().GetString("auth")
			if err != nil {
				return err
			}
			if authID == "" {
				return fmt.Errorf("--auth is required")
			}
			projects, err := svc.BindProjectsAuth([]string{args[0]}, authID)
			if err != nil {
				return err
			}
			for _, project := range projects {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔗 bound: %s -> %s\n", project.ProjectID, authID)
			}
			return nil
		},
	}
	cmd.Flags().String("auth", "", "Auth id to bind")
	return cmd
}
