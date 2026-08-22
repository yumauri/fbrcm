package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/fileoutput"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

type defaultsDownloader func(context.Context, string, firebase.DefaultsFormat) ([]byte, error)

func newDefaultsCommand(svc *core.Core) *cobra.Command {
	return newDefaultsCommandWithDownloader(svc, func(ctx context.Context, projectID string, format firebase.DefaultsFormat) ([]byte, error) {
		return svc.DownloadRemoteConfigDefaults(ctx, projectID, format)
	})
}

func newDefaultsCommandWithDownloader(svc *core.Core, download defaultsDownloader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "defaults <project>",
		Short: "Download template application defaults",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			formatName, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}
			format, err := firebase.ParseDefaultsFormat(formatName)
			if err != nil {
				return shared.InvalidArgument(err)
			}
			ctx := shared.CommandContext(cmd)
			project, err := shared.ResolveProjectTargetForExecution(ctx, cmd, svc, args[0])
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
			ctx, err = shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
			if err != nil {
				return err
			}
			cmd.SetContext(ctx)

			defaults, err := download(ctx, project.ProjectID, format)
			if err != nil {
				return err
			}
			if toPath == "" {
				if contract.Enabled(cmd) {
					target := project.ProjectID
					return shared.WriteJSON(cmd, contract.NewArtifact(&target, defaultsMediaType(format), defaults, nil, false))
				}
				_, err = cmd.OutOrStdout().Write(defaults)
				return err
			}

			write := fileoutput.Create
			if overwrite {
				write = fileoutput.Write
			}
			if err := write(toPath, defaults); err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				target, destination := project.ProjectID, toPath
				return shared.WriteJSON(cmd, contract.NewArtifact(&target, defaultsMediaType(format), defaults, &destination, overwrite))
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "downloaded defaults: %s\n", toPath)
			return err
		},
	}
	cmd.Flags().String("format", "json", "Defaults format: json, xml, or plist")
	cmd.Flags().String("to", "", "Write application defaults to file path")
	shared.AddYesFlag(cmd, "Overwrite an existing destination without confirmation")
	return cmd
}

func defaultsMediaType(format firebase.DefaultsFormat) string {
	switch format {
	case firebase.DefaultsFormatXML:
		return "application/xml"
	case firebase.DefaultsFormatPlist:
		return "application/x-plist"
	default:
		return "application/json"
	}
}
