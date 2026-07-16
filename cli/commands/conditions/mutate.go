package conditions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	sharedrc "github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	coreconditions "github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

type mutationOptions struct {
	Draft  bool
	DryRun bool
	Yes    bool
}

type conditionMutation func(*firebase.RemoteConfig) error

func newAddCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <project> <name>",
		Short: "Add a condition",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			expression, _ := cmd.Flags().GetString("expression")
			color, _ := cmd.Flags().GetString("color")
			priority, _ := cmd.Flags().GetInt("priority")
			definition := core.ConditionDefinition{Name: args[1], Expression: expression, TagColor: color}
			return runConditionMutation(cmd, svc, args[0], readMutationOptions(cmd), "add condition", "➕", false, func(cfg *firebase.RemoteConfig) error {
				return coreconditions.Add(cfg, definition, priority)
			})
		},
	}
	cmd.Flags().String("expression", "", "Raw Firebase condition expression (required)")
	cmd.Flags().String("color", "", "Firebase display color")
	cmd.Flags().Int("priority", 0, "Evaluation priority; defaults to last")
	_ = cmd.MarkFlagRequired("expression")
	addMutationFlags(cmd)
	return cmd
}

func newEditCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <project> <condition>",
		Short: "Edit a condition expression or color",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("expression") && !cmd.Flags().Changed("color") && !cmd.Flags().Changed("no-color") {
				return fmt.Errorf("at least one of --expression, --color, or --no-color is required")
			}
			var expression, color *string
			if cmd.Flags().Changed("expression") {
				value, _ := cmd.Flags().GetString("expression")
				expression = &value
			}
			if cmd.Flags().Changed("color") {
				value, _ := cmd.Flags().GetString("color")
				color = &value
			}
			if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
				value := ""
				color = &value
			}
			return runNamedConditionMutation(cmd, svc, args[0], args[1], readMutationOptions(cmd), "edit condition", "✏️", false, func(cfg *firebase.RemoteConfig, name string) error {
				return coreconditions.EditDefinition(cfg, name, core.ConditionEdit{Expression: expression, TagColor: color})
			})
		},
	}
	cmd.Flags().String("expression", "", "New raw Firebase condition expression")
	cmd.Flags().String("color", "", "New Firebase display color")
	cmd.Flags().Bool("no-color", false, "Remove the display color")
	cmd.MarkFlagsMutuallyExclusive("color", "no-color")
	addMutationFlags(cmd)
	return cmd
}

func newRenameCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <project> <condition> <new-name>",
		Short: "Rename a condition and all of its parameter references",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNamedConditionMutation(cmd, svc, args[0], args[1], readMutationOptions(cmd), "rename condition", "✏️", false, func(cfg *firebase.RemoteConfig, name string) error {
				return coreconditions.Rename(cfg, name, args[2])
			})
		},
	}
	addMutationFlags(cmd)
	return cmd
}

func newMoveCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <project> <condition> <priority>",
		Short: "Move a condition to a new evaluation priority",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			priority, err := strconv.Atoi(args[2])
			if err != nil {
				return fmt.Errorf("invalid condition priority %q", args[2])
			}
			return runNamedConditionMutation(cmd, svc, args[0], args[1], readMutationOptions(cmd), "move condition", "↕️", false, func(cfg *firebase.RemoteConfig, name string) error {
				tree := coreconditions.BuildTree(cfg, time.Time{}, "")
				impact, err := tree.MoveImpact(name, priority)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), rcdisplay.FormatConditionMoveImpact(len(impact.CrossedConditions), len(impact.AffectedParameters)))
				return coreconditions.Move(cfg, name, priority)
			})
		},
	}
	addMutationFlags(cmd)
	return cmd
}

func newDeleteCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <project> <condition>",
		Short: "Delete a condition and its conditional values",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNamedConditionMutation(cmd, svc, args[0], args[1], readMutationOptions(cmd), "delete condition", "🗑️", true, func(cfg *firebase.RemoteConfig, name string) error {
				tree := coreconditions.BuildTree(cfg, time.Time{}, "")
				impact, err := tree.DeleteImpact(name)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), rcdisplay.FormatConditionDeleteImpact(len(impact.Usages), len(impact.RemovedParameters)))
				return coreconditions.Delete(cfg, name)
			})
		},
	}
	addMutationFlags(cmd)
	return cmd
}

func newValidateCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <project>",
		Short: "Validate the current draft or published conditions with Firebase",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			project, err := shared.ResolveProjectArg(ctx, cmd, svc, args[0])
			if err != nil {
				return err
			}
			hasDraft, err := svc.HasDraft(project.ProjectID)
			if err != nil {
				return err
			}
			var cache *core.ParametersCache
			var raw json.RawMessage
			source := "firebase"
			if hasDraft {
				plan, err := svc.PrepareDraftPublish(ctx, project.ProjectID)
				if err != nil {
					return err
				}
				raw = plan.Candidate
				cache = plan.Latest
				source = "draft"
			} else {
				cache, _, err = svc.RevalidateParameters(ctx, project.ProjectID)
				if err != nil {
					return err
				}
				raw = cache.RemoteConfig
			}
			if err := svc.ValidateRemoteConfigWithETag(ctx, project.ProjectID, raw, cache.ETag); err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, map[string]any{"project": project, "source": source, "valid": true})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Valid: %s (%s) · %s\n", project.Name, project.ProjectID, source)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print validation result as JSON")
	return cmd
}

func addMutationFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddDryRunFlag(cmd)
	shared.AddYesFlag(cmd, "Print diff and apply without confirmation")
}

func readMutationOptions(cmd *cobra.Command) mutationOptions {
	draft, _ := cmd.Flags().GetBool("draft")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	return mutationOptions{Draft: draft, DryRun: dryRun, Yes: yes}
}

func runNamedConditionMutation(cmd *cobra.Command, svc *core.Core, projectQuery, requestedName string, opts mutationOptions, operation, emoji string, destructive bool, mutate func(*firebase.RemoteConfig, string) error) error {
	return runConditionMutation(cmd, svc, projectQuery, opts, operation, emoji, destructive, func(cfg *firebase.RemoteConfig) error {
		name, ok := coreconditions.ResolveName(cfg, requestedName)
		if !ok {
			return fmt.Errorf("condition %q not found", requestedName)
		}
		return mutate(cfg, name)
	})
}

func runConditionMutation(cmd *cobra.Command, svc *core.Core, projectQuery string, opts mutationOptions, operation, emoji string, destructive bool, mutate conditionMutation) error {
	ctx := context.Background()
	if opts.DryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	project, err := shared.ResolveProjectArg(ctx, cmd, svc, projectQuery)
	if err != nil {
		return err
	}
	plan := func(project core.Project, _ *sharedrc.ProjectConfig) (sharedrc.RemoteConfigMutation, error) {
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			finalCfg, err := firebase.CloneRemoteConfig(current)
			if err != nil {
				return 0, nil, err
			}
			if err := mutate(finalCfg); err != nil {
				return 0, nil, err
			}
			diffText, changed := sharedrc.RenderRemoteConfigDiff(current, finalCfg)
			if !changed {
				return 0, finalCfg, nil
			}
			confirmed, err := shared.PrintDiffAndConfirm(cmd, opts.Yes, cmd.ErrOrStderr(), diffText, "Apply condition changes to "+project.ProjectID+"?", destructive)
			if err != nil || !confirmed {
				return 0, finalCfg, err
			}
			return 1, finalCfg, nil
		}, nil
	}
	projects := []core.Project{project}
	if opts.Draft {
		_, err = sharedrc.RunRemoteDraftLoop(ctx, cmd, svc, projects, operation, plan)
	} else {
		_, err = sharedrc.RunRemotePublishLoop(ctx, cmd, svc, projects, operation, emoji, plan)
	}
	return err
}
