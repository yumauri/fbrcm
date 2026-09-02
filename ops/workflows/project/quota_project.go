package project

import (
	"fmt"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
)

type projectQuotaProjectResult struct {
	Project                  string                      `json:"project"`
	ProjectID                string                      `json:"project_id"`
	AuthID                   string                      `json:"auth_id"`
	ConfiguredQuotaProjectID *string                     `json:"configured_quota_project_id"`
	EffectiveQuotaProjectID  *string                     `json:"effective_quota_project_id"`
	Source                   firebase.QuotaProjectSource `json:"source" contract:"enum=environment|project|auth|credentials|target"`
	PreviousQuotaProjectID   *string                     `json:"previous_quota_project_id,omitempty"`
	Changed                  bool                        `json:"changed"`
	Status                   string                      `json:"status" contract:"enum=shown|set|unset|unchanged"`
}

func newQuotaProjectCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{Use: "quota-project", Short: "Manage a project's quota-project override"}
	cmd.AddCommand(newProjectQuotaProjectShowCommandDefinition(svc), newProjectQuotaProjectSetCommandDefinition(svc), newProjectQuotaProjectUnsetCommandDefinition(svc))
	return cmd
}

func newProjectQuotaProjectShowCommandDefinition(svc *core.Core) *invocation.Definition {
	return &invocation.Definition{
		Use:   "show <project>",
		Short: "Show a project's quota-project resolution",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			project, err := resolveTemplatePreferencesProject(cmd, args[0])
			if err != nil {
				return err
			}
			project, selection, err := svc.ResolveProjectQuotaProject(shared.CommandContext(cmd), project.ProjectID)
			if err != nil {
				return err
			}
			return writeProjectQuotaProjectResult(cmd, newProjectQuotaResult(project, "", selection, false, "shown"))
		},
	}
}

func newProjectQuotaProjectSetCommandDefinition(svc *core.Core) *invocation.Definition {
	return &invocation.Definition{
		Use:   "set <project> <quota-project-id>",
		Short: "Set a project's quota-project override",
		Args:  invocation.ExactArgs(2),
		RunE: func(cmd invocation.Call, args []string) error {
			if err := config.ValidateQuotaProjectID(args[1]); err != nil {
				return shared.InvalidArgument(err)
			}
			project, err := resolveTemplatePreferencesProject(cmd, args[0])
			if err != nil {
				return err
			}
			project, previous, changed, err := svc.SetProjectQuotaProject(project.ProjectID, args[1])
			if err != nil {
				return err
			}
			_, selection, err := svc.ResolveProjectQuotaProject(shared.CommandContext(cmd), project.ProjectID)
			if err != nil {
				return err
			}
			status := "set"
			if !changed {
				status = "unchanged"
			}
			return writeProjectQuotaProjectResult(cmd, newProjectQuotaResult(project, previous, selection, changed, status))
		},
	}
}

func newProjectQuotaProjectUnsetCommandDefinition(svc *core.Core) *invocation.Definition {
	return &invocation.Definition{
		Use:   "unset <project>",
		Short: "Remove a project's quota-project override",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			project, err := resolveTemplatePreferencesProject(cmd, args[0])
			if err != nil {
				return err
			}
			project, previous, changed, err := svc.SetProjectQuotaProject(project.ProjectID, "")
			if err != nil {
				return err
			}
			_, selection, err := svc.ResolveProjectQuotaProject(shared.CommandContext(cmd), project.ProjectID)
			if err != nil {
				return err
			}
			status := "unset"
			if !changed {
				status = "unchanged"
			}
			return writeProjectQuotaProjectResult(cmd, newProjectQuotaResult(project, previous, selection, changed, status))
		},
	}
}

func newProjectQuotaResult(project core.Project, previous string, selection firebase.QuotaProjectSelection, changed bool, status string) projectQuotaProjectResult {
	return projectQuotaProjectResult{
		Project: project.Name, ProjectID: project.ProjectID, AuthID: project.AuthID,
		ConfiguredQuotaProjectID: quotaProjectPointer(project.QuotaProjectID), EffectiveQuotaProjectID: quotaProjectPointer(selection.ProjectID),
		Source: selection.Source, PreviousQuotaProjectID: quotaProjectPointer(previous), Changed: changed, Status: status,
	}
}

func quotaProjectPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func writeProjectQuotaProjectResult(cmd invocation.Call, result projectQuotaProjectResult) error {
	if contract.Enabled(cmd) {
		return shared.WriteJSON(cmd, result)
	}
	configured := "—"
	if result.ConfiguredQuotaProjectID != nil {
		configured = *result.ConfiguredQuotaProjectID
	}
	effective := "unresolved"
	if result.EffectiveQuotaProjectID != nil {
		effective = *result.EffectiveQuotaProjectID
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\nProject ID: %s\nConfigured quota project: %s\nEffective quota project: %s\nSource: %s\n", result.Project, result.ProjectID, configured, effective, result.Source)
	return err
}
