package project

import (
	"fmt"
	"slices"
	"strings"

	"github.com/yumauri/fbrcm/core"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
)

type projectTemplatesJSON struct {
	Project         string          `json:"project"`
	ProjectID       string          `json:"project_id"`
	Templates       []rctarget.Kind `json:"templates"`
	PrimaryTemplate rctarget.Kind   `json:"primary_template"`
}

func newTemplatesCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "templates",
		Short: "Manage a project's default template selection",
	}
	cmd.AddCommand(newTemplatesShowCommandDefinition(), newTemplatesSetCommandDefinition(svc))
	return cmd
}

func newTemplatesShowCommandDefinition() *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "show <project>",
		Short: "Show enabled and primary templates",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			project, err := resolveTemplatePreferencesProject(cmd, args[0])
			if err != nil {
				return err
			}
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			return writeProjectTemplates(cmd, project, jsonOut)
		},
	}
	cmd.Flags().Bool("json", false, "Print template preferences as JSON")
	return cmd
}

func newTemplatesSetCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "set <project>",
		Short: "Set enabled or primary templates",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			templatesChanged := cmd.Flags().Changed("templates")
			primaryChanged := cmd.Flags().Changed("primary")
			if !templatesChanged && !primaryChanged {
				return shared.InvalidArgument(fmt.Errorf("at least one of --templates or --primary is required"))
			}

			project, err := resolveTemplatePreferencesProject(cmd, args[0])
			if err != nil {
				return err
			}
			templates := append([]rctarget.Kind(nil), project.Templates...)
			primary := project.PrimaryTemplate

			if templatesChanged {
				rawTemplates, err := cmd.Flags().GetStringSlice("templates")
				if err != nil {
					return err
				}
				templates, err = parseTemplateKinds(rawTemplates)
				if err != nil {
					return err
				}
			}
			if primaryChanged {
				rawPrimary, err := cmd.Flags().GetString("primary")
				if err != nil {
					return err
				}
				primary, err = parseTemplateKind(rawPrimary)
				if err != nil {
					return shared.InvalidArgument(fmt.Errorf("invalid --primary: %w", err))
				}
			}

			if len(templates) == 1 && !primaryChanged {
				primary = templates[0]
			}
			if !slices.Contains(templates, primary) {
				return shared.InvalidArgument(fmt.Errorf("primary template %q is not enabled by --templates", primary))
			}

			updated, err := svc.SetProjectTemplatePreferences(project.ProjectID, templates, primary)
			if err != nil {
				return err
			}
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			return writeProjectTemplates(cmd, updated, jsonOut)
		},
	}
	cmd.Flags().StringSlice("templates", nil, "Replace enabled templates with client, server, or both")
	cmd.Flags().String("primary", "", "Set the primary template: client or server")
	cmd.Flags().Bool("json", false, "Print updated template preferences as JSON")
	return cmd
}

func resolveTemplatePreferencesProject(cmd invocation.Call, query string) (core.Project, error) {
	target, explicit, err := rctarget.ParsePositionalSelector(query)
	if err != nil {
		return core.Project{}, shared.InvalidArgument(err)
	}
	if explicit {
		return core.Project{}, shared.InvalidArgument(fmt.Errorf(
			"template preferences belong to physical project %q; omit the %s@ prefix",
			target.ProjectID,
			target.Kind,
		))
	}
	return shared.ResolveCachedProjectArg(cmd, target.ProjectID)
}

func parseTemplateKinds(values []string) ([]rctarget.Kind, error) {
	seen := make(map[rctarget.Kind]struct{}, len(values))
	for _, value := range values {
		kind, err := parseTemplateKind(value)
		if err != nil {
			return nil, shared.InvalidArgument(fmt.Errorf("invalid --templates value: %w", err))
		}
		seen[kind] = struct{}{}
	}
	if len(seen) == 0 {
		return nil, shared.InvalidArgument(fmt.Errorf("--templates requires client, server, or both"))
	}
	templates := make([]rctarget.Kind, 0, len(seen))
	for _, kind := range []rctarget.Kind{rctarget.Client, rctarget.Server} {
		if _, ok := seen[kind]; ok {
			templates = append(templates, kind)
		}
	}
	return templates, nil
}

func parseTemplateKind(value string) (rctarget.Kind, error) {
	switch rctarget.Kind(strings.ToLower(strings.TrimSpace(value))) {
	case rctarget.Client:
		return rctarget.Client, nil
	case rctarget.Server:
		return rctarget.Server, nil
	default:
		return "", fmt.Errorf("template must be client or server, got %q", value)
	}
}

func writeProjectTemplates(cmd invocation.Call, project core.Project, jsonOut bool) error {
	if err := project.NormalizeTemplatePreferences(); err != nil {
		return err
	}
	if jsonOut {
		return shared.WriteJSON(cmd, projectTemplatesJSON{
			Project:         project.Name,
			ProjectID:       project.ProjectID,
			Templates:       append([]rctarget.Kind(nil), project.Templates...),
			PrimaryTemplate: project.PrimaryTemplate,
		})
	}
	_, err := fmt.Fprintln(cmd.OutOrStdout(), renderProjectTemplates(project))
	return err
}

func renderProjectTemplates(project core.Project) string {
	return strings.Join([]string{
		"Project: " + displayProjectValue(project.Name),
		"Project ID: " + project.ProjectID,
		"Enabled templates: " + projectTemplatesLabel(project),
		"Primary template: " + string(project.PrimaryTemplate),
	}, "\n")
}
