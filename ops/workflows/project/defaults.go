package project

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/ops/shared/fileoutput"
)

type defaultsDownloader func(context.Context, string, firebase.DefaultsFormat) ([]byte, error)

func newDefaultsCommandDefinition(svc *core.Core) *invocation.Definition {
	return newDefaultsCommandWithDownloaderDefinition(svc, func(ctx context.Context, projectID string, format firebase.DefaultsFormat) ([]byte, error) {
		return svc.DownloadRemoteConfigDefaults(ctx, projectID, format)
	})
}

func newDefaultsCommandWithDownloaderDefinition(svc *core.Core, download defaultsDownloader) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "defaults <project>",
		Short: "Download template application defaults",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
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
