package projects

import (
	"context"
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
		Short: "Manage Firebase projects for Remote Config",
	}
	projectsCmd.AddCommand(newListCommand(svc), newUpdateCommand(svc), newPathCommand(), newPurgeCommand(svc))
	return projectsCmd
}

func newListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
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
	addProjectOutputFlags(cmd)
	cmd.Flags().Bool("update", false, "Update projects from Firebase before printing")
	return cmd
}

func newUpdateCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
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
	addProjectOutputFlags(cmd)
	cmd.Flags().String("auth", "", "Sync projects for one auth id")
	return cmd
}

func addProjectOutputFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Print projects as JSON")
	shared.AddProjectListFilterFlag(cmd)
	cmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	cmd.Flags().Bool("url", false, "Include Firebase Console Remote Config URL")
}

func newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print projects config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := config.GetProjectsFilePath()
			if jsonOut {
				if err := shared.WriteJSON(cmd, map[string]string{"path": path}); err != nil {
					return fmt.Errorf("encode projects path json: %w", err)
				}
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print path as JSON")
	return cmd
}

func newPurgeCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
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
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}
