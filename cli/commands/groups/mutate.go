package groups

import (
	"context"
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
}

type groupMutation func(*firebase.RemoteConfig) (bool, error)

func newAddCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "add <name>", Short: "Add an empty parameter group across projects", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		name, err := coregroups.NormalizeName(args[0])
		if err != nil {
			return err
		}
		description, _ := cmd.Flags().GetString("description")
		return runGroupMutation(cmd, svc, readMutationOptions(cmd), "add group", "➕", false, func(cfg *firebase.RemoteConfig) (bool, error) {
			if _, exists := cfg.ParameterGroups[name]; exists {
				return false, nil
			}
			return true, coregroups.Add(cfg, coregroups.Definition{Name: name, Description: description})
		})
	}}
	cmd.Flags().String("description", "", "Group description")
	addMutationFlags(cmd)
	return cmd
}

func newEditCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "edit <group>", Short: "Edit a group description across projects", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("description") && !cmd.Flags().Changed("no-description") {
			return fmt.Errorf("one of --description or --no-description is required")
		}
		value, _ := cmd.Flags().GetString("description")
		if noDescription, _ := cmd.Flags().GetBool("no-description"); noDescription {
			value = ""
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
	cmd.MarkFlagsMutuallyExclusive("description", "no-description")
	addMutationFlags(cmd)
	return cmd
}

func newRenameCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{Use: "rename <group> <new-name>", Short: "Rename a parameter group across projects", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		nextName, err := coregroups.NormalizeName(args[1])
		if err != nil {
			return err
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
	shared.AddProjectFilterFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddDryRunFlag(cmd)
	shared.AddYesFlag(cmd, "Print diff and apply without confirmation")
}

func readMutationOptions(cmd *cobra.Command) mutationOptions {
	projectFilters, _ := cmd.Flags().GetStringArray("project")
	draft, _ := cmd.Flags().GetBool("draft")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	return mutationOptions{ProjectFilters: projectFilters, Draft: draft, DryRun: dryRun, Yes: yes}
}

func runNamedGroupMutation(cmd *cobra.Command, svc *core.Core, requested string, opts mutationOptions, operation, emoji string, destructive bool, mutate func(*firebase.RemoteConfig, string) (bool, error)) error {
	return runGroupMutation(cmd, svc, opts, operation, emoji, destructive, namedGroupMutation(requested, mutate))
}

func namedGroupMutation(requested string, mutate func(*firebase.RemoteConfig, string) (bool, error)) groupMutation {
	return func(cfg *firebase.RemoteConfig) (bool, error) {
		name, ok := coregroups.ResolveName(cfg, requested)
		if !ok {
			return false, nil
		}
		return mutate(cfg, name)
	}
}

func runGroupMutation(cmd *cobra.Command, svc *core.Core, opts mutationOptions, operation, emoji string, destructive bool, mutate groupMutation) error {
	ctx := context.Background()
	if opts.DryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return err
	}
	projects = shared.FilterProjects(projects, opts.ProjectFilters)
	strfold.SortProjects(projects, func(project core.Project) string { return project.Name }, func(project core.Project) string { return project.ProjectID })
	plan := func(project core.Project, _ *sharedrc.ProjectConfig) (sharedrc.RemoteConfigMutation, error) {
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			finalCfg, err := firebase.CloneRemoteConfig(current)
			if err != nil {
				return 0, nil, err
			}
			applicable, err := mutate(finalCfg)
			if err != nil {
				return 0, nil, err
			}
			if !applicable {
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
		}, nil
	}
	if opts.Draft {
		_, err = sharedrc.RunRemoteDraftLoop(ctx, cmd, svc, projects, operation, plan)
	} else {
		_, err = sharedrc.RunRemotePublishLoop(ctx, cmd, svc, projects, operation, emoji, plan)
	}
	return err
}
