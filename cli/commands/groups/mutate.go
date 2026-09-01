package groups

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	sharedrc "github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	coregroups "github.com/yumauri/fbrcm/core/groups"
	"github.com/yumauri/fbrcm/core/strfold"
)

type mutationOptions struct {
	ProjectFilters []string
	Draft          bool
	DryRun         bool
	Yes            bool
	ChangeNote     *string
}

type groupMutationResult struct {
	matched    bool
	applicable bool
}

type groupMutation func(*firebase.RemoteConfig) (groupMutationResult, error)

func newAddCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "add <name>", Short: "Add an empty parameter group across projects", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		name, err := coregroups.NormalizeName(args[0])
		if err != nil {
			return shared.InvalidArgument(err)
		}
		description, _ := cmd.Flags().GetString("description")
		return runGroupMutation(cmd, svc, readMutationOptions(cmd), "add group", "➕", false, func(cfg *firebase.RemoteConfig) (groupMutationResult, error) {
			if _, exists := cfg.ParameterGroups[name]; exists {
				return groupMutationResult{matched: true}, nil
			}
			err := coregroups.Add(cfg, coregroups.Definition{Name: name, Description: description})
			return groupMutationResult{matched: true, applicable: true}, err
		})
	}}
	cmd.Flags().String("description", "", "Group description")
	addMutationFlags(cmd)
	return cmd
}

func newEditCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "edit <group>", Short: "Edit a group description across projects", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		value, err := readDescriptionEdit(cmd)
		if err != nil {
			return err
		}
		return runNamedGroupMutation(cmd, svc, args[0], readMutationOptions(cmd), "edit group", "✏️", false, func(cfg *firebase.RemoteConfig, name string) (bool, error) {
			if cfg.ParameterGroups[name].Description == strings.TrimSpace(value) {
				return false, nil
			}
			return true, coregroups.EditMetadata(cfg, name, coregroups.Edit{Description: &value})
		})
	}}
	cmd.Flags().String("description", "", "New group description")
	cmd.Flags().Bool("no-description", false, "Remove the group description")
	addMutationFlags(cmd)
	return cmd
}

func readDescriptionEdit(cmd *cobra.Command) (string, error) {
	descriptionChanged := cmd.Flags().Changed("description")
	noDescription, err := cmd.Flags().GetBool("no-description")
	if err != nil {
		return "", shared.InvalidArgument(err)
	}
	if !descriptionChanged && !noDescription {
		return "", shared.InvalidArgument(fmt.Errorf("one of --description or --no-description is required"))
	}
	if descriptionChanged && noDescription {
		return "", shared.InvalidArgument(fmt.Errorf("--description and --no-description are mutually exclusive"))
	}
	if noDescription {
		return "", nil
	}
	value, err := cmd.Flags().GetString("description")
	if err != nil {
		return "", shared.InvalidArgument(err)
	}
	return value, nil
}

func newRenameCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "rename <group> <new-name>", Short: "Rename a parameter group across projects", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		nextName, err := coregroups.NormalizeName(args[1])
		if err != nil {
			return shared.InvalidArgument(err)
		}
		return runNamedGroupMutation(cmd, svc, args[0], readMutationOptions(cmd), "rename group", "✏️", false, func(cfg *firebase.RemoteConfig, name string) (bool, error) {
			if name == nextName {
				return false, nil
			}
			return true, coregroups.Rename(cfg, name, nextName)
		})
	}}
	addMutationFlags(cmd)
	return cmd
}

func newDeleteCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "delete <group>", Short: "Delete a group across projects", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return runNamedGroupMutation(cmd, svc, args[0], readMutationOptions(cmd), "delete group", "🗑️", true, func(cfg *firebase.RemoteConfig, name string) (bool, error) {
			return true, coregroups.Delete(cfg, name)
		})
	}}
	addMutationFlags(cmd)
	return cmd
}

func addMutationFlags(cmd *cobra.Command) {
	shared.AddProjectTargetFilterFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddDryRunFlag(cmd)
	shared.AddChangeNoteFlag(cmd)
	shared.AddYesFlag(cmd, "Print diff and apply without confirmation")
	shared.AddPlanOutFlag(cmd)
	cmd.Flags().Bool("json", false, "Print mutation results as JSON")
}

func readMutationOptions(cmd *cobra.Command) mutationOptions {
	projectFilters, _ := cmd.Flags().GetStringArray("project")
	draft, _ := cmd.Flags().GetBool("draft")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	var changeNote *string
	if cmd.Flags().Changed("change-note") {
		value, _ := cmd.Flags().GetString("change-note")
		changeNote = &value
	}
	return mutationOptions{ProjectFilters: projectFilters, Draft: draft, DryRun: dryRun, Yes: yes, ChangeNote: changeNote}
}

func runNamedGroupMutation(cmd *cobra.Command, svc *core.Core, requested string, opts mutationOptions, operation, emoji string, destructive bool, mutate func(*firebase.RemoteConfig, string) (bool, error)) error {
	return runGroupMutation(cmd, svc, opts, operation, emoji, destructive, namedGroupMutation(requested, mutate))
}

func namedGroupMutation(requested string, mutate func(*firebase.RemoteConfig, string) (bool, error)) groupMutation {
	return func(cfg *firebase.RemoteConfig) (groupMutationResult, error) {
		name, ok := coregroups.ResolveName(cfg, requested)
		if !ok {
			return groupMutationResult{}, nil
		}
		applicable, err := mutate(cfg, name)
		return groupMutationResult{matched: true, applicable: applicable}, err
	}
}

func runGroupMutation(cmd *cobra.Command, svc *core.Core, opts mutationOptions, operation, emoji string, destructive bool, mutate groupMutation) error {
	ctx := shared.CommandContext(cmd)
	if err := shared.RejectStatelessDraft(ctx, opts.Draft); err != nil {
		return err
	}
	if opts.DryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	var err error
	ctx, err = shared.WithChangeNote(ctx, opts.ChangeNote)
	if err != nil {
		return err
	}
	projects, ctx, err := shared.ResolveProjectMutationTargetsForExecution(ctx, cmd, svc, opts.ProjectFilters)
	if err != nil {
		return err
	}
	strfold.SortProjects(projects, func(project core.Project) string { return project.Name }, func(project core.Project) string { return project.ProjectID })
	plan := func(project core.Project, cfg *sharedrc.ProjectConfig) (sharedrc.RemoteMutationPlan, error) {
		probe, err := firebase.CloneRemoteConfig(cfg.Config)
		if err != nil {
			return sharedrc.RemoteMutationPlan{}, err
		}
		probeResult, err := mutate(probe)
		if err != nil {
			return sharedrc.RemoteMutationPlan{}, err
		}
		if !probeResult.matched {
			return sharedrc.RemoteMutationPlan{}, nil
		}
		return sharedrc.RemoteMutationPlan{MatchedItemCount: 1, Mutation: func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			finalCfg, err := firebase.CloneRemoteConfig(current)
			if err != nil {
				return 0, nil, err
			}
			result, err := mutate(finalCfg)
			if err != nil {
				return 0, nil, err
			}
			if !result.applicable {
				return 0, finalCfg, nil
			}
			diffText, changed := sharedrc.RenderRemoteConfigDiff(current, finalCfg)
			if !changed {
				return 0, finalCfg, nil
			}
			confirmed, err := shared.PrintDiffAndConfirm(cmd, opts.Yes, cmd.ErrOrStderr(), diffText, "Apply group changes to "+project.ProjectID+"?", destructive)
			if err != nil || !confirmed {
				return 0, finalCfg, err
			}
			return 1, finalCfg, nil
		}}, nil
	}
	var totals sharedrc.RemoteMutationTotals
	if opts.Draft {
		totals, err = sharedrc.RunRemoteDraftLoop(ctx, cmd, svc, projects, len(opts.ProjectFilters) == 0, operation, plan)
	} else {
		totals, err = sharedrc.RunRemotePublishLoop(ctx, cmd, svc, projects, len(opts.ProjectFilters) == 0, operation, emoji, plan)
	}
	if writeErr := sharedrc.WriteRemoteMutationResults(cmd, totals, map[bool]string{true: "draft", false: "publish"}[opts.Draft], emoji); writeErr != nil {
		return writeErr
	}
	return err
}
