package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func newShowCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <project>",
		Short: "Show project details and auth access",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			update, err := cmd.Flags().GetBool("update")
			if err != nil {
				return err
			}
			if update {
				if _, _, err := svc.SyncProjects(ctx); err != nil {
					return err
				}
			}

			project, err := shared.ResolveProjectArg(ctx, cmd, svc, args[0])
			if err != nil {
				return err
			}
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			if jsonOut {
				aliases, aliasErr := projectAliases(project.ProjectID)
				if aliasErr != nil {
					return aliasErr
				}
				return shared.WriteJSON(cmd, shared.NewProjectJSONWithAliases(project, aliases, true))
			}

			aliases, aliasErr := projectAliases(project.ProjectID)
			if aliasErr != nil {
				return aliasErr
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), renderProjectDetailsWithAliases(project, aliases))
			return err
		},
	}
	cmd.Flags().Bool("update", false, "Update projects from Firebase before printing")
	cmd.Flags().Bool("json", false, "Print project details as JSON")
	return cmd
}

func renderProjectDetailsWithAliases(project core.Project, aliases []string) string {
	_ = project.NormalizeTemplatePreferences()
	authIdentities := strings.Join(project.DiscoveredBy, ", ")
	if authIdentities == "" {
		authIdentities = "none recorded"
	}
	aliasLabel := strings.Join(aliases, ", ")
	if aliasLabel == "" {
		aliasLabel = "—"
	}

	return strings.Join([]string{
		"Project: " + displayProjectValue(project.Name),
		"Project ID: " + project.ProjectID,
		"Aliases: " + aliasLabel,
		"Status: " + projectStatus(project),
		"Number: " + displayProjectValue(project.ProjectNumber),
		"State: " + displayProjectValue(project.State),
		"Selected auth: " + displayProjectValue(project.AuthID),
		"Auth identities: " + authIdentities,
		"Enabled templates: " + projectTemplatesLabel(project),
		"Primary template: " + string(project.PrimaryTemplate),
		"Updated at: " + displayProjectValue(shared.FormatDateTime(project.UpdatedAt)),
		"Synced at: " + displayProjectValue(shared.FormatDateTime(project.SyncedAt)),
		"ETag: " + displayProjectValue(project.ETag),
		"URL: " + firebase.RemoteConfigConsoleURL(project.ProjectID),
	}, "\n")
}

func projectAliases(projectID string) ([]string, error) {
	aliases, err := config.LoadProjectAliases()
	if err != nil {
		return nil, err
	}
	return config.ProjectAliasesByID(aliases)[projectID], nil
}

func projectTemplatesLabel(project core.Project) string {
	if err := project.NormalizeTemplatePreferences(); err != nil {
		return "—"
	}
	templates := make([]string, len(project.Templates))
	for i, kind := range project.Templates {
		templates[i] = string(kind)
	}
	return strings.Join(templates, ", ")
}

func projectStatus(project core.Project) string {
	if project.Disabled {
		return "disabled"
	}
	return "enabled"
}

func displayProjectValue(value string) string {
	if value == "" {
		return "—"
	}
	return value
}
