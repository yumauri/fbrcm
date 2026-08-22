package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

func New(svc *core.Core) *cobra.Command {
	projectsCmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage Firebase projects for Remote Config",
	}
	projectsCmd.AddCommand(newListCommand(svc), newUpdateCommand(svc), newForgetCommand(svc), newDiffCommand(svc), newPromoteCommand(svc), newAliasesCommand(), newPathCommand(), newResetCommand(svc))
	contract.MustRegisterResponsePath(projectsCmd, "list", []shared.ProjectJSON{})
	contract.MustRegisterResponsePath(projectsCmd, "update", []shared.ProjectJSON{})
	contract.MustRegisterResponsePath(projectsCmd, "forget", projectsForgetResult{})
	contract.MustRegisterResponsePath(projectsCmd, "diff", compareResult{})
	contract.MustRegisterResponsePath(projectsCmd, "promote", promoteResult{})
	contract.MustRegisterResponsePath(projectsCmd, "aliases list", []projectAliasRow{})
	contract.MustRegisterResponsePath(projectsCmd, "aliases set", projectAliasSetResult{})
	contract.MustRegisterResponsePath(projectsCmd, "aliases remove", projectAliasRemoveResult{})
	contract.MustRegisterResponsePath(projectsCmd, "aliases import", projectAliasImportResult{})
	contract.MustRegisterResponsePath(projectsCmd, "path", shared.PathResult{})
	contract.MustRegisterResponsePath(projectsCmd, "reset", projectsResetResult{})
	return projectsCmd
}

func newForgetCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forget",
		Short: "Forget locally tracked projects and delete their local data",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := config.LoadProjects()
			if err != nil {
				return err
			}
			filters, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				return err
			}
			projects, err = shared.FilterProjects(projects, filters)
			if err != nil {
				return err
			}
			rawExpr, err := cmd.Flags().GetString("expr")
			if err != nil {
				return err
			}
			projects, err = shared.FilterProjectsByCachedExpr(svc, projects, rawExpr)
			if err != nil {
				return err
			}
			if len(projects) == 0 {
				return &shared.SelectionError{Resource: "project", Kind: "not_found", Query: fmt.Sprintf("filters=%v expression=%q", filters, rawExpr), Err: fmt.Errorf("no projects matched")}
			}

			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				if err := shared.RequireYesInMachineMode(cmd, yes, "forgetting matched projects and deleting their local data", true); err != nil {
					return err
				}
				confirm := shared.NewConfirmation(
					projectForgetConfirmationPrompt(len(projects)),
					shared.ConfirmationOptions{
						Destructive: true,
						Notes:       []shared.ConfirmationNote{{Text: "Firebase projects will not be deleted."}},
					},
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

			projectIDs := make([]string, 0, len(projects))
			for _, project := range projects {
				projectIDs = append(projectIDs, project.ProjectID)
			}
			deleted, err := svc.DeleteProjectIDs(projectIDs)
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				items := make([]forgottenProject, 0, len(deleted))
				for _, project := range deleted {
					items = append(items, forgottenProject{Project: project.Name, ProjectID: project.ProjectID, Status: "forgotten"})
				}
				return shared.WriteJSON(cmd, projectsForgetResult{Count: len(items), Items: items})
			}
			for _, project := range deleted {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 forgot project: %s (%s)\n", project.Name, project.ProjectID)
			}
			return nil
		},
	}
	shared.AddProjectListFilterFlag(cmd)
	cmd.Flags().String("expr", "", "Filter projects by expr-lang expression using local cache only")
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}

func projectForgetConfirmationPrompt(count int) string {
	return fmt.Sprintf(
		"Forget %s and delete all associated local caches, versions, and drafts?",
		rcdisplay.FormatCount(count, "project", "projects"),
	)
}

func newListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects using cache-first loading",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := shared.CommandContext(cmd)
			forceUpdate, err := cmd.Flags().GetBool("update")
			if err != nil {
				return err
			}
			if !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
				if forceUpdate {
					return shared.InvalidArgument(fmt.Errorf("--update cannot be used with --stateless; project discovery is already live"))
				}
				ctx, err = shared.FirebaseProjectDiscoveryContextForExecution(ctx)
				if err != nil {
					return err
				}
				cmd.SetContext(ctx)
			}

			var projects []core.Project
			var source string
			if forceUpdate {
				progress.Start("Syncing projects…")
				projects, source, err = svc.SyncProjects(ctx)
			} else {
				projects, source, err = svc.ListProjectsForExecution(ctx)
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
				projects, source, err = svc.SyncProjectsForAuth(shared.CommandContext(cmd), authID)
			} else {
				projects, source, err = svc.SyncProjects(shared.CommandContext(cmd))
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
				if err := shared.WriteJSON(cmd, shared.PathResult{Path: path}); err != nil {
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

func newResetCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the cached projects registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				if err := shared.RequireYesInMachineMode(cmd, yes, "resetting the cached projects registry", true); err != nil {
					return err
				}
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Reset cached projects registry by deleting %s?", config.GetProjectsFilePath()),
					shared.ConfirmationOptions{
						Destructive: true,
						Notes:       []shared.ConfirmationNote{{Text: "Remote Config snapshots, cached versions, and drafts will not be deleted."}},
					},
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

			changed, err := svc.ResetProjects()
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, projectsResetResult{Path: config.GetProjectsFilePath(), Status: "reset", Changed: changed})
			}

			if changed {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 reset projects registry: %s\n", config.GetProjectsFilePath())
			} else {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "projects registry already reset: %s\n", config.GetProjectsFilePath())
			}
			return nil
		},
	}
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}

type forgottenProject struct {
	Project   string `json:"project"`
	ProjectID string `json:"project_id"`
	Status    string `json:"status" contract:"enum=forgotten"`
}

type projectsForgetResult struct {
	Count int                `json:"count"`
	Items []forgottenProject `json:"items"`
}

type projectsResetResult struct {
	Path    string `json:"path"`
	Status  string `json:"status" contract:"enum=reset"`
	Changed bool   `json:"changed"`
}
