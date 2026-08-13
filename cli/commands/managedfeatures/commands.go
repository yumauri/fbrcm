package managedfeatures

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
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
	contract.MustRegisterResponsePath(cmd, "list", []core.ExperimentEntry{})
	contract.MustRegisterResponsePath(cmd, "show", experimentShowResult{})
	contract.MustRegisterResponsePath(cmd, "delete", managedFeatureDeleteResult{})
	return cmd
}

func NewRollouts(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollouts",
		Short: "Inspect Firebase Remote Config rollouts",
		Long:  "Inspect rollout metadata and its parameter bindings in the published Remote Config template, or explicitly delete a rollout.",
	}
	cmd.AddCommand(newRolloutsListCommand(svc), newRolloutsShowCommand(svc), newRolloutsDeleteCommand(svc))
	contract.MustRegisterResponsePath(cmd, "list", []core.RolloutEntry{})
	contract.MustRegisterResponsePath(cmd, "show", rolloutShowResult{})
	contract.MustRegisterResponsePath(cmd, "delete", managedFeatureDeleteResult{})
	return cmd
}

func NewPersonalizations(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "personalizations",
		Short: "Inspect Firebase Remote Config personalizations",
		Long:  "Inspect read-only personalization IDs and their parameter bindings in the published Remote Config template.",
	}
	cmd.AddCommand(newPersonalizationsListCommand(svc), newPersonalizationsShowCommand(svc))
	contract.MustRegisterResponsePath(cmd, "list", []core.PersonalizationEntry{})
	contract.MustRegisterResponsePath(cmd, "show", personalizationShowResult{})
	return cmd
}

type experimentShowResult struct {
	Project    core.Project                `json:"project"`
	Template   core.ManagedFeatureTemplate `json:"template"`
	Experiment core.ExperimentEntry        `json:"experiment"`
}

type rolloutShowResult struct {
	Project  core.Project                `json:"project"`
	Template core.ManagedFeatureTemplate `json:"template"`
	Rollout  core.RolloutEntry           `json:"rollout"`
}

type personalizationShowResult struct {
	Project         core.Project                `json:"project"`
	Template        core.ManagedFeatureTemplate `json:"template"`
	Personalization core.PersonalizationEntry   `json:"personalization"`
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
			result, err := svc.ListRemoteConfigExperiments(shared.CommandContext(cmd), project, update)
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
			experiment, template, err := svc.GetRemoteConfigExperiment(shared.CommandContext(cmd), project, args[1], update)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, experimentShowResult{Project: project, Template: template, Experiment: experiment})
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
			experiment, err := svc.GetRemoteConfigExperimentMetadata(shared.CommandContext(cmd), project, args[1])
			if err != nil {
				return err
			}
			id := firebase.ManagedFeatureID(experiment.Name)
			identity := managedFeatureIdentity(experiment.Definition.DisplayName, id)
			yes, _ := cmd.Flags().GetBool("yes")
			if contract.Enabled(cmd) && !yes {
				if err := shared.WriteJSON(cmd, managedFeatureDeleteResult{Kind: "experiment", ID: id, DisplayName: experiment.Definition.DisplayName, ProjectID: project.ProjectID, Status: "would-delete"}); err != nil {
					return err
				}
				return shared.InteractionRequired("confirmation is required before deleting the experiment", true, "--yes")
			}
			confirmed, err := confirmManagedFeatureDelete(cmd, yes, fmt.Sprintf("Delete experiment %s from %s?", identity, project.ProjectID))
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
			progress.Start("Deleting experiment from " + project.ProjectID + "…")
			if err := svc.DeleteRemoteConfigExperiment(shared.CommandContext(cmd), project, id); err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, managedFeatureDeleteResult{Kind: "experiment", ID: id, DisplayName: experiment.Definition.DisplayName, ProjectID: project.ProjectID, Status: "deleted"})
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
			result, err := svc.ListRemoteConfigRollouts(shared.CommandContext(cmd), project, update)
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
			rollout, template, err := svc.GetRemoteConfigRollout(shared.CommandContext(cmd), project, args[1], update)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, rolloutShowResult{Project: project, Template: template, Rollout: rollout})
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
			rollout, err := svc.GetRemoteConfigRolloutMetadata(shared.CommandContext(cmd), project, args[1])
			if err != nil {
				return err
			}
			id := firebase.ManagedFeatureID(rollout.Name)
			identity := managedFeatureIdentity(rollout.Definition.DisplayName, id)
			yes, _ := cmd.Flags().GetBool("yes")
			if contract.Enabled(cmd) && !yes {
				if err := shared.WriteJSON(cmd, managedFeatureDeleteResult{Kind: "rollout", ID: id, DisplayName: rollout.Definition.DisplayName, ProjectID: project.ProjectID, Status: "would-delete"}); err != nil {
					return err
				}
				return shared.InteractionRequired("confirmation is required before deleting the rollout", true, "--yes")
			}
			confirmed, err := confirmManagedFeatureDelete(cmd, yes, fmt.Sprintf("Delete rollout %s from %s?", identity, project.ProjectID))
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
			progress.Start("Deleting rollout from " + project.ProjectID + "…")
			if err := svc.DeleteRemoteConfigRollout(shared.CommandContext(cmd), project, id); err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, managedFeatureDeleteResult{Kind: "rollout", ID: id, DisplayName: rollout.Definition.DisplayName, ProjectID: project.ProjectID, Status: "deleted"})
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
			result, err := svc.ListRemoteConfigPersonalizations(shared.CommandContext(cmd), project, update)
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
			personalization, template, err := svc.GetRemoteConfigPersonalization(shared.CommandContext(cmd), project, args[1], update)
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, personalizationShowResult{Project: project, Template: template, Personalization: personalization})
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
	target, explicit, err := rctarget.ParsePositionalSelector(query)
	if err != nil {
		return core.Project{}, shared.InvalidArgument(err)
	}
	if explicit {
		if target.Kind == rctarget.Server {
			return core.Project{}, shared.InvalidArgument(fmt.Errorf(
				"managed features support only the client Remote Config namespace; omit the server@ prefix",
			))
		}
		return core.Project{}, shared.InvalidArgument(fmt.Errorf(
			"managed-feature commands are project-scoped; omit the client@ prefix",
		))
	}
	ctx := shared.CommandContext(cmd)
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
	if err := shared.RequireYesInMachineMode(cmd, yes, "deleting the managed feature", true); err != nil {
		return false, err
	}
	confirm := shared.NewConfirmation(prompt, shared.ConfirmationOptions{Destructive: true})
	confirm.Input = cmd.InOrStdin()
	confirm.Output = cmd.ErrOrStderr()
	return confirm.RunPrompt()
}

type managedFeatureDeleteResult struct {
	Kind        string `json:"kind" contract:"enum=experiment|rollout"`
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	ProjectID   string `json:"project_id"`
	Status      string `json:"status" contract:"enum=deleted|would-delete"`
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
