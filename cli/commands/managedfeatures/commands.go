package managedfeatures

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func NewExperiments(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "experiments",
		Short: "Inspect Firebase Remote Config experiments",
		Long:  "Inspect experiment metadata and its parameter bindings in the published Remote Config template, or explicitly delete an experiment.",
	}
	cmd.AddCommand(newExperimentsListCommand(svc), newExperimentsShowCommand(svc), newExperimentsDeleteCommand(svc))
	return cmd
}

func NewRollouts(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollouts",
		Short: "Inspect Firebase Remote Config rollouts",
		Long:  "Inspect rollout metadata and its parameter bindings in the published Remote Config template, or explicitly delete a rollout.",
	}
	cmd.AddCommand(newRolloutsListCommand(svc), newRolloutsShowCommand(svc), newRolloutsDeleteCommand(svc))
	return cmd
}

func NewPersonalizations(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "personalizations",
		Short: "Inspect Firebase Remote Config personalizations",
		Long:  "Inspect read-only personalization IDs and their parameter bindings in the published Remote Config template.",
	}
	cmd.AddCommand(newPersonalizationsListCommand(svc), newPersonalizationsShowCommand(svc))
	return cmd
}

func newExperimentsListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List experiments and their parameter bindings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			update, _ := cmd.Flags().GetBool("update")
			result, err := svc.ListRemoteConfigExperiments(cmd.Context(), project, update)
			if err != nil {
				return err
			}
			rawFilters, _ := cmd.Flags().GetStringArray("filter")
			result.Experiments = filterExperimentsByName(result.Experiments, rawFilters)
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, result.Experiments)
			}
			printTemplateContext(cmd, project, result.Template)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderExperimentsTable(result.Experiments))
			return nil
		},
	}
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter experiments by mode-prefixed display-name query (^, /, ~, =); may be repeated")
	addTemplateReadFlags(cmd, "experiments")
	return cmd
}

func filterExperimentsByName(experiments []core.ExperimentEntry, rawFilters []string) []core.ExperimentEntry {
	filters := shared.ParseFilters(rawFilters)
	if len(filters) == 0 {
		return experiments
	}
	filtered := make([]core.ExperimentEntry, 0, len(experiments))
	for _, experiment := range experiments {
		if shared.MatchAnyFilter(experiment.Definition.DisplayName, filters) {
			filtered = append(filtered, experiment)
		}
	}
	return filtered
}

func newExperimentsShowCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <project> <experiment-id>",
		Short: "Show experiment details and parameter bindings",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			update, _ := cmd.Flags().GetBool("update")
			experiment, template, err := svc.GetRemoteConfigExperiment(cmd.Context(), project, args[1], update)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, map[string]any{
					"project": project, "template": template, "experiment": experiment,
				})
			}
			printTemplateContext(cmd, project, template)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderExperimentDetails(experiment))
			return nil
		},
	}
	addTemplateReadFlags(cmd, "experiment details")
	return cmd
}

func newExperimentsDeleteCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <project> <experiment-id>",
		Short: "Delete an experiment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			experiment, err := svc.GetRemoteConfigExperimentMetadata(cmd.Context(), project, args[1])
			if err != nil {
				return err
			}
			id := firebase.ManagedFeatureID(experiment.Name)
			identity := managedFeatureIdentity(experiment.Definition.DisplayName, id)
			yes, _ := cmd.Flags().GetBool("yes")
			confirmed, err := confirmManagedFeatureDelete(cmd, yes, fmt.Sprintf("Delete experiment %s from %s?", identity, project.ProjectID))
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
			progress.Start("Deleting experiment from " + project.ProjectID + "…")
			if err := svc.DeleteRemoteConfigExperiment(cmd.Context(), project, id); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🗑️ deleted experiment: %s from %s\n", identity, project.ProjectID)
			return nil
		},
	}
	shared.AddYesFlag(cmd, "Delete without confirmation")
	return cmd
}

func newRolloutsListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List rollouts and their parameter bindings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			update, _ := cmd.Flags().GetBool("update")
			result, err := svc.ListRemoteConfigRollouts(cmd.Context(), project, update)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, result.Rollouts)
			}
			printTemplateContext(cmd, project, result.Template)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderRolloutsTable(result.Rollouts))
			return nil
		},
	}
	addTemplateReadFlags(cmd, "rollouts")
	return cmd
}

func newRolloutsShowCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <project> <rollout-id>",
		Short: "Show rollout details and parameter bindings",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			update, _ := cmd.Flags().GetBool("update")
			rollout, template, err := svc.GetRemoteConfigRollout(cmd.Context(), project, args[1], update)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, map[string]any{
					"project": project, "template": template, "rollout": rollout,
				})
			}
			printTemplateContext(cmd, project, template)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderRolloutDetails(rollout))
			return nil
		},
	}
	addTemplateReadFlags(cmd, "rollout details")
	return cmd
}

func newRolloutsDeleteCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <project> <rollout-id>",
		Short: "Delete a rollout",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			rollout, err := svc.GetRemoteConfigRolloutMetadata(cmd.Context(), project, args[1])
			if err != nil {
				return err
			}
			id := firebase.ManagedFeatureID(rollout.Name)
			identity := managedFeatureIdentity(rollout.Definition.DisplayName, id)
			yes, _ := cmd.Flags().GetBool("yes")
			confirmed, err := confirmManagedFeatureDelete(cmd, yes, fmt.Sprintf("Delete rollout %s from %s?", identity, project.ProjectID))
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
			progress.Start("Deleting rollout from " + project.ProjectID + "…")
			if err := svc.DeleteRemoteConfigRollout(cmd.Context(), project, id); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🗑️ deleted rollout: %s from %s\n", identity, project.ProjectID)
			return nil
		},
	}
	shared.AddYesFlag(cmd, "Delete without confirmation")
	return cmd
}

func newPersonalizationsListCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <project>",
		Short: "List personalizations and their parameter bindings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			update, _ := cmd.Flags().GetBool("update")
			result, err := svc.ListRemoteConfigPersonalizations(cmd.Context(), project, update)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, result.Personalizations)
			}
			printTemplateContext(cmd, project, result.Template)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderPersonalizationsTable(result.Personalizations))
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nFirebase exposes personalization IDs and bindings in the template, but not value candidates or result metrics through this API.")
			return nil
		},
	}
	addTemplateReadFlags(cmd, "personalizations")
	return cmd
}

func newPersonalizationsShowCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <project> <personalization-id>",
		Short: "Show personalization parameter bindings",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			update, _ := cmd.Flags().GetBool("update")
			personalization, template, err := svc.GetRemoteConfigPersonalization(cmd.Context(), project, args[1], update)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, map[string]any{
					"project": project, "template": template, "personalization": personalization,
				})
			}
			printTemplateContext(cmd, project, template)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderPersonalizationDetails(personalization))
			return nil
		},
	}
	addTemplateReadFlags(cmd, "personalization details")
	return cmd
}

func resolveProject(cmd *cobra.Command, svc *core.Core, query string) (core.Project, error) {
	target, explicit, err := rctarget.ParseSelector(query)
	if err != nil {
		return core.Project{}, err
	}
	if explicit {
		if target.Kind == rctarget.Server {
			return core.Project{}, fmt.Errorf(
				"managed features support only the client Remote Config namespace; omit the server@ prefix",
			)
		}
		return core.Project{}, fmt.Errorf(
			"managed-feature commands are project-scoped; omit the client@ prefix",
		)
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return shared.ResolveProjectArg(ctx, cmd, svc, target.ProjectID)
}

func addTemplateReadFlags(cmd *cobra.Command, noun string) {
	cmd.Flags().Bool("update", false, "Revalidate cached Remote Config before reading parameter bindings")
	cmd.Flags().Bool("json", false, "Print "+noun+" as JSON")
}

func confirmManagedFeatureDelete(cmd *cobra.Command, yes bool, prompt string) (bool, error) {
	if yes {
		return true, nil
	}
	confirm := shared.NewConfirmation(prompt, shared.ConfirmationOptions{Destructive: true})
	confirm.Input = cmd.InOrStdin()
	confirm.Output = cmd.ErrOrStderr()
	return confirm.RunPrompt()
}

func managedFeatureIdentity(displayName, id string) string {
	if displayName == "" {
		return id
	}
	return fmt.Sprintf("%q (%s)", displayName, id)
}

func printTemplateContext(cmd *cobra.Command, project core.Project, template core.ManagedFeatureTemplate) {
	_, _ = fmt.Fprintf(
		cmd.OutOrStdout(),
		"Project: %s (%s)\nRemote Config: version %s · Source: %s\n\n",
		project.Name,
		project.ProjectID,
		template.Version,
		template.Source,
	)
}
