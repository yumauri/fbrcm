package projects

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
)

func New(svc *core.Core) *cobra.Command {
	projectsCmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects list",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List projects using cache-first loading",
		RunE: func(cmd *cobra.Command, args []string) error {
			forceUpdate, err := cmd.Flags().GetBool("update")
			if err != nil {
				return err
			}

			var projects []core.Project
			var source string
			if forceUpdate {
				projects, source, err = svc.SyncProjects(context.Background())
			} else {
				projects, source, err = svc.ListProjects(context.Background())
			}
			if err != nil {
				return err
			}

			return printProjects(cmd, svc, projects, source)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update projects from Firebase into cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			authID, err := cmd.Flags().GetString("auth")
			if err != nil {
				return err
			}
			var projects []core.Project
			var source string
			if authID != "" {
				projects, source, err = svc.SyncProjectsForAuth(context.Background(), authID)
			} else {
				projects, source, err = svc.SyncProjects(context.Background())
			}
			if err != nil {
				return err
			}

			return printProjects(cmd, svc, projects, source)
		},
	}

	listCmd.Flags().Bool("json", false, "Print projects as JSON")
	listCmd.Flags().StringArrayP("filter", "f", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	listCmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	listCmd.Flags().Bool("update", false, "Update projects from Firebase before printing")
	listCmd.Flags().Bool("url", false, "Include Firebase Console Remote Config URL")
	updateCmd.Flags().Bool("json", false, "Print projects as JSON")
	updateCmd.Flags().StringArrayP("filter", "f", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	updateCmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	updateCmd.Flags().Bool("url", false, "Include Firebase Console Remote Config URL")
	updateCmd.Flags().String("auth", "", "Sync projects for one auth id")

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print projects config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := config.GetProjectsFilePath()
			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(map[string]string{"path": path}); err != nil {
					return fmt.Errorf("encode projects path json: %w", err)
				}
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	pathCmd.Flags().Bool("json", false, "Print path as JSON")

	purgeCmd := &cobra.Command{
		Use:   "purge",
		Short: "Delete cached projects config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Delete cached projects config file %s?", config.GetProjectsFilePath()),
					confirmation.Yes,
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

			if err := svc.PurgeProjects(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged: %s\n", config.GetProjectsFilePath())
			return nil
		},
	}
	purgeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation dialog")

	projectsCmd.AddCommand(listCmd, updateCmd, pathCmd, purgeCmd)
	return projectsCmd
}
