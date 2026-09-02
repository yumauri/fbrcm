package project

import (
	"fmt"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/browser"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/ops/shared/rc"
	importpkg "github.com/yumauri/fbrcm/ops/workflows/project/import"
)

type projectOpenResult struct {
	ProjectID string `json:"project_id"`
	URL       string `json:"url" contract:"format=uri"`
	Opened    bool   `json:"opened"`
}

// New constructs the project command.
func NewDefinition(svc *core.Core) *invocation.Definition {
	projectCmd := &invocation.Definition{
		Use:   "project",
		Short: "Manage individual project Remote Config",
	}
	projectCmd.AddCommand(
		newShowCommandDefinition(svc),
		newTemplatesCommandDefinition(svc),
		newQuotaProjectCommandDefinition(svc),
		newOpenCommandDefinition(svc, browser.OpenURL),
		newExportCommandDefinition(svc),
		newImportCommandDefinition(svc),
		newDefaultsCommandDefinition(svc),
	)
	invocation.MustRegisterResponsePath(projectCmd, "show", shared.ProjectJSON{})
	invocation.MustRegisterResponsePath(projectCmd, "templates show", projectTemplatesJSON{})
	invocation.MustRegisterResponsePath(projectCmd, "templates set", projectTemplatesJSON{})
	for _, path := range []string{"quota-project show", "quota-project set", "quota-project unset"} {
		invocation.MustRegisterResponsePath(projectCmd, path, projectQuotaProjectResult{})
	}
	invocation.MustRegisterResponsePath(projectCmd, "open", projectOpenResult{})
	invocation.MustRegisterResponsePath(projectCmd, "export", contract.ArtifactData{})
	invocation.MustRegisterResponsePath(projectCmd, "import", importpkg.Result{}, rc.PlanCreatedResult{})
	invocation.MustRegisterResponsePath(projectCmd, "defaults", contract.ArtifactData{})
	return projectCmd
}

func newOpenCommandDefinition(svc *core.Core, openURL func(string) error) *invocation.Definition {
	return &invocation.Definition{
		Use:   "open <project>",
		Short: "Open project Remote Config in the Firebase console",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			project, err := shared.ResolvePhysicalProjectForExecution(shared.CommandContext(cmd), cmd, svc, args[0])
			if err != nil {
				return err
			}
			url := firebase.RemoteConfigConsoleURL(project.ProjectID)
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, projectOpenResult{ProjectID: project.ProjectID, URL: url, Opened: false})
			}
			if firebase.IsOffline() {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), url)
				return err
			}
			if err := openURL(url); err != nil {
				return err
			}
			return nil
		},
	}
}

func newExportCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "export <project>",
		Short: "Export project Remote Config JSON",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
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

			raw, _, err := svc.ExportRemoteConfig(ctx, project.ProjectID)
			if err != nil {
				return err
			}
			body := rc.NormalizeExportBytes(raw)
			if toPath == "" {
				if contract.Enabled(cmd) {
					target := project.ProjectID
					return shared.WriteJSON(cmd, contract.NewArtifact(&target, "application/json", body, nil, false))
				}
				_, err = cmd.OutOrStdout().Write(body)
				return err
			}

			write := rc.CreateRemoteConfigFile
			if overwrite {
				write = rc.WriteRemoteConfigFile
			}
			if err := write(toPath, body); err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				target, destination := project.ProjectID, toPath
				return shared.WriteJSON(cmd, contract.NewArtifact(&target, "application/json", body, &destination, overwrite))
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📤 exported: %s\n", toPath)
			return nil
		},
	}
	cmd.Flags().String("to", "", "Write Remote Config JSON to file path")
	shared.AddYesFlag(cmd, "Overwrite an existing destination without confirmation")
	return cmd
}

func newImportCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "import <project>",
		Short: "Import project Remote Config JSON",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			ctx := shared.CommandContext(cmd)
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			if dryRun {
				ctx = firebase.WithDryRun(ctx)
			}

			project, err := shared.ResolveProjectTargetForExecution(ctx, cmd, svc, args[0])
			if err != nil {
				return err
			}
			ctx, err = shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
			if err != nil {
				return err
			}
			cmd.SetContext(ctx)
			return importpkg.Run(cmd, svc, project)
		},
	}
	cmd.Flags().String("from", "", "Read Remote Config JSON from file path")
	cmd.Flags().StringArray("group", nil, "Import only specified parameter group; may be repeated")
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().String("expr", "", "Filter imported config by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	shared.AddChangeNoteFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	cmd.Flags().Bool("remove-all-conditions", false, "Remove all conditions and conditional values from imported config")
	cmd.Flags().Bool("keep-portable-conditions-only", false, "Keep only portable conditions and remove destination-specific usages")
	cmd.Flags().Bool("merge", false, "Merge imported config into current project config")
	cmd.Flags().Bool("override", false, "Replace current project config with imported config")
	cmd.Flags().String("merge-resolve", "", "Conflict resolution for merge: current or import")
	shared.AddYesFlag(cmd, "Skip final import confirmation")
	shared.AddPlanOutFlag(cmd)
	cmd.Flags().Bool("json", false, "Print import result as JSON")
	cmd.MarkFlagsMutuallyExclusive("remove-all-conditions", "keep-portable-conditions-only")
	cmd.MarkFlagsMutuallyExclusive("merge", "override")
	return cmd
}
