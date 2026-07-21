package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	importpkg "github.com/yumauri/fbrcm/cli/commands/project/import"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/browser"
	"github.com/yumauri/fbrcm/core/firebase"
)

// New constructs the project command.
func New(svc *core.Core) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Manage individual project Remote Config",
	}
	projectCmd.AddCommand(newShowCommand(svc), newOpenCommand(svc, browser.OpenURL), newExportCommand(svc), newImportCommand(svc), newDefaultsCommand(svc))
	return projectCmd
}

func newOpenCommand(svc *core.Core, openURL func(string) error) *cobra.Command {
	return &cobra.Command{
		Use:   "open <project>",
		Short: "Open project Remote Config in the Firebase console",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := shared.ResolveProjectArg(context.Background(), cmd, svc, args[0])
			if err != nil {
				return err
			}
			return openURL(firebase.RemoteConfigConsoleURL(project.ProjectID))
		},
	}
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

			toPath, err := cmd.Flags().GetString("to")
			if err != nil {
				return err
			}
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			overwrite := false
			if toPath != "" {
				var proceed bool
				overwrite, proceed, err = shared.ConfirmFileOverwrite(cmd, toPath, yes)
				if err != nil || !proceed {
					return err
				}
			}

			raw, _, err := svc.ExportRemoteConfig(context.Background(), project.ProjectID)
			if err != nil {
				return err
			}
			if toPath == "" {
				body := rc.TrimTrailingLineBreaks(rc.NormalizeExportJSON(raw))
				_, err = cmd.OutOrStdout().Write(body)
				return err
			}

			write := rc.CreateRemoteConfigFile
			if overwrite {
				write = rc.WriteRemoteConfigFile
			}
			if err := write(toPath, raw); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📤 exported: %s\n", toPath)
			return nil
		},
	}
	cmd.Flags().String("to", "", "Write Remote Config JSON to file path")
	shared.AddYesFlag(cmd, "Overwrite an existing destination without confirmation")
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
	cmd.Flags().Bool("keep-portable-conditions-only", false, "Keep only portable conditions and remove destination-specific usages")
	cmd.Flags().Bool("merge", false, "Merge imported config into current project config")
	cmd.Flags().Bool("override", false, "Replace current project config with imported config")
	cmd.Flags().String("merge-resolve", "", "Conflict resolution for merge: current or import")
	shared.AddYesFlag(cmd, "Skip final import confirmation")
	cmd.Flags().Bool("json", false, "Print import result as JSON")
	cmd.MarkFlagsMutuallyExclusive("remove-all-conditions", "keep-portable-conditions-only")
	cmd.MarkFlagsMutuallyExclusive("merge", "override")
	return cmd
}
