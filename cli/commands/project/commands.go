package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	importpkg "github.com/yumauri/fbrcm/cli/commands/project/import"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

// New constructs the project command.
func New(svc *core.Core) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Export and import project Remote Config",
	}
	projectCmd.AddCommand(newExportCommand(svc), newImportCommand(svc))
	return projectCmd
}

func newExportCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <project>",
		Short: "Export project Remote Config JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := shared.ResolveProjectArg(context.Background(), cmd, svc, args[0])
			if err != nil {
				return err
			}

			raw, _, err := svc.ExportRemoteConfig(context.Background(), project.ProjectID)
			if err != nil {
				return err
			}

			toPath, err := cmd.Flags().GetString("to")
			if err != nil {
				return err
			}
			if toPath == "" {
				body := rc.TrimTrailingLineBreaks(rc.NormalizeExportJSON(raw))
				_, err = cmd.OutOrStdout().Write(body)
				return err
			}

			if err := rc.WriteRemoteConfigFile(toPath, raw); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📤 exported: %s\n", toPath)
			return nil
		},
	}
	cmd.Flags().String("to", "", "Write Remote Config JSON to file path")
	return cmd
}

func newImportCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <project>",
		Short: "Import project Remote Config JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			if dryRun {
				ctx = firebase.WithDryRun(ctx)
			}

			project, err := shared.ResolveProjectArg(ctx, cmd, svc, args[0])
			if err != nil {
				return err
			}
			return importpkg.Run(cmd, svc, project)
		},
	}
	cmd.Flags().String("from", "", "Read Remote Config JSON from file path")
	cmd.Flags().StringArray("group", nil, "Import only specified parameter group; may be repeated")
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().String("expr", "", "Filter imported config by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	cmd.Flags().Bool("remove-all-conditions", false, "Remove all conditions and conditional values from imported config")
	cmd.Flags().Bool("remove-project-specific-conditions", false, "Remove project specific conditions and their usages from imported config")
	cmd.Flags().Bool("merge", false, "Merge imported config into current project config")
	cmd.Flags().Bool("override", false, "Replace current project config with imported config")
	cmd.Flags().String("merge-resolve", "", "Conflict resolution for merge: current or import")
	cmd.MarkFlagsMutuallyExclusive("remove-all-conditions", "remove-project-specific-conditions")
	cmd.MarkFlagsMutuallyExclusive("merge", "override")
	return cmd
}
