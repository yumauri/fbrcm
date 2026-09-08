package conditions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/yumauri/fbrcm/core"
	coreconditions "github.com/yumauri/fbrcm/core/conditions"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	sharedrc "github.com/yumauri/fbrcm/ops/shared/rc"
)

type mutationOptions struct {
	ProjectFilters       []string
	ProjectFilterChanged bool
	Draft                bool
	DryRun               bool
	Yes                  bool
	ChangeNote           *string
}

type conditionMutation func(*firebase.RemoteConfig) error

func newAddCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "add [project] <name>",
		Short: "Add a condition across projects",
		Args:  invocation.RangeArgs(1, 2),
		RunE: func(cmd invocation.Call, args []string) error {
			projectQuery, name := conditionAddArguments(args)
			expression, _ := cmd.Flags().GetString("expression")
			color, _ := cmd.Flags().GetString("color")
			priority, _ := cmd.Flags().GetInt("priority")
			definition := core.ConditionDefinition{Name: name, Expression: expression, TagColor: color}
			return runConditionMutationTargets(cmd, svc, projectQuery, readMutationOptions(cmd), "add condition", "➕", false, func(cfg *firebase.RemoteConfig) error {
				return coreconditions.Add(cfg, definition, priority)
			})
		},
	}
	shared.AddProjectTargetFilterFlag(cmd)
	cmd.Flags().String("expression", "", "Raw Firebase condition expression (required)")
	cmd.Flags().String("color", "", "Firebase display color")
	cmd.Flags().Int("priority", 0, "Evaluation priority; defaults to last")
	_ = cmd.MarkFlagRequired("expression")
	addMutationFlags(cmd)
	return cmd
}

func newEditCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "edit <project> <condition>",
		Short: "Edit a condition expression or color",
		Args:  invocation.ExactArgs(2),
		RunE: func(cmd invocation.Call, args []string) error {
			expression, color, err := readConditionEdit(cmd)
			if err != nil {
				return err
			}
			return runNamedConditionMutation(cmd, svc, args[0], args[1], readMutationOptions(cmd), "edit condition", "✏️", false, func(cfg *firebase.RemoteConfig, name string) error {
				return coreconditions.EditDefinition(cfg, name, core.ConditionEdit{Expression: expression, TagColor: color})
			})
		},
	}
	cmd.Flags().String("expression", "", "New raw Firebase condition expression")
	cmd.Flags().String("color", "", "New Firebase display color")
	cmd.Flags().Bool("no-color", false, "Remove the display color")
	addMutationFlags(cmd)
	return cmd
}

func readConditionEdit(cmd invocation.Call) (*string, *string, error) {
	expressionChanged := cmd.Flags().Changed("expression")
	colorChanged := cmd.Flags().Changed("color")
	noColor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		return nil, nil, shared.InvalidArgument(err)
	}
	if !expressionChanged && !colorChanged && !noColor {
		return nil, nil, shared.InvalidArgument(fmt.Errorf("at least one edit flag is required"))
	}
	if colorChanged && noColor {
		return nil, nil, shared.InvalidArgument(fmt.Errorf("--color and --no-color are mutually exclusive"))
	}
	var expression, color *string
	if expressionChanged {
		value, getErr := cmd.Flags().GetString("expression")
		if getErr != nil {
			return nil, nil, shared.InvalidArgument(getErr)
		}
		expression = &value
	}
	if colorChanged {
		value, getErr := cmd.Flags().GetString("color")
		if getErr != nil {
			return nil, nil, shared.InvalidArgument(getErr)
		}
		color = &value
	}
	if noColor {
		value := ""
		color = &value
	}
	return expression, color, nil
}

func newRenameCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "rename <project> <condition> <new-name>",
		Short: "Rename a condition and all of its parameter references",
		Args:  invocation.ExactArgs(3),
		RunE: func(cmd invocation.Call, args []string) error {
			return runNamedConditionMutation(cmd, svc, args[0], args[1], readMutationOptions(cmd), "rename condition", "✏️", false, func(cfg *firebase.RemoteConfig, name string) error {
				return coreconditions.Rename(cfg, name, args[2])
			})
		},
	}
	addMutationFlags(cmd)
	return cmd
}

func newMoveCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "move <project> <condition> <priority>",
		Short: "Move a condition to a new evaluation priority",
		Args:  invocation.ExactArgs(3),
		RunE: func(cmd invocation.Call, args []string) error {
			priority, err := strconv.Atoi(args[2])
			if err != nil {
				return shared.InvalidArgument(fmt.Errorf("invalid condition priority %q", args[2]))
			}
			return runNamedConditionMutation(cmd, svc, args[0], args[1], readMutationOptions(cmd), "move condition", "↕️", false, func(cfg *firebase.RemoteConfig, name string) error {
				tree := coreconditions.BuildTree(cfg, time.Time{}, "")
				impact, err := tree.MoveImpact(name, priority)
				if err != nil {
					return err
				}
				if !shared.MachineMode(cmd) {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), rcdisplay.FormatConditionMoveImpact(len(impact.CrossedConditions), len(impact.AffectedParameters)))
				}
				return coreconditions.Move(cfg, name, priority)
			})
		},
	}
	addMutationFlags(cmd)
	return cmd
}

func newDeleteCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "delete [project] [condition]",
		Short: "Delete matching conditions and their conditional values across projects",
		Args:  invocation.MaximumNArgs(2),
		RunE: func(cmd invocation.Call, args []string) error {
			projectQuery, exactName := conditionDeleteArguments(args)
			filters, _ := cmd.Flags().GetStringArray("filter")
			if exactName != nil && shared.HasFilters(filters) {
				return shared.InvalidArgument(fmt.Errorf("condition argument cannot be used together with --filter"))
			}
			search, _ := cmd.Flags().GetString("search")
			rawExpr, _ := cmd.Flags().GetString("expr")
			compiledExpr, err := shared.CompileExpr(rawExpr, "")
			if err != nil {
				return err
			}
			opts := readMutationOptions(cmd)
			return runConditionMutationPlans(cmd, svc, projectQuery, opts, "delete condition", "🗑️", func(project core.Project, cfg *sharedrc.ProjectConfig) (sharedrc.RemoteMutationPlan, error) {
				tree := coreconditions.BuildTree(cfg.Config, time.Time{}, "")
				matched, exactFound, err := selectConditionEntries(project, tree.Conditions, exactName, filters, search, compiledExpr)
				if err != nil {
					return sharedrc.RemoteMutationPlan{}, err
				}
				if exactName != nil && !exactFound {
					return sharedrc.RemoteMutationPlan{}, &shared.SelectionError{Resource: "condition", Kind: "not_found", Query: *exactName, Err: fmt.Errorf("condition %q not found", *exactName)}
				}
				if len(matched) == 0 {
					return sharedrc.RemoteMutationPlan{}, nil
				}
				names := make([]string, len(matched))
				conditionalValues := 0
				for index, entry := range matched {
					names[index] = entry.Name
					conditionalValues += len(entry.Usages)
				}
				return sharedrc.RemoteMutationPlan{MatchedItemCount: len(names), Mutation: func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
					finalCfg, err := firebase.CloneRemoteConfig(current)
					if err != nil {
						return 0, nil, err
					}
					beforeParameters := len(rcmutate.CollectParamSlots(current))
					for _, name := range names {
						if err := coreconditions.Delete(finalCfg, name); err != nil {
							return 0, nil, typedConditionMutationError(project.ProjectID, err)
						}
					}
					if !shared.MachineMode(cmd) {
						removedParameters := beforeParameters - len(rcmutate.CollectParamSlots(finalCfg))
						_, _ = fmt.Fprintln(cmd.ErrOrStderr(), rcdisplay.FormatConditionDeleteImpact(conditionalValues, removedParameters))
					}
					diffText, changed := sharedrc.RenderRemoteConfigDiff(current, finalCfg)
					if !changed {
						return 0, finalCfg, nil
					}
					confirmed, err := shared.PrintDiffAndConfirm(cmd, opts.Yes, cmd.ErrOrStderr(), diffText, "Apply condition changes to "+project.ProjectID+"?", true)
					if err != nil || !confirmed {
						return 0, finalCfg, err
					}
					return len(names), finalCfg, nil
				}}, nil
			})
		},
	}
	shared.AddProjectTargetFilterFlag(cmd)
	addConditionFilterFlags(cmd)
	addMutationFlags(cmd)
	return cmd
}

func newValidateCommandDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "validate <project>",
		Short: "Validate the current draft or published conditions with Firebase",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			ctx := shared.CommandContext(cmd)
			project, err := shared.ResolveProjectTargetForExecution(ctx, cmd, svc, args[0])
			if err != nil {
				return err
			}
			ctx, err = shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
			if err != nil {
				return err
			}
			cmd.SetContext(ctx)
			if !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
				raw, etag, err := svc.ExportRemoteConfig(ctx, project.ProjectID)
				if err != nil {
					return err
				}
				if err := svc.ValidateRemoteConfigWithETag(ctx, project.ProjectID, raw, etag); err != nil {
					return err
				}
				if shared.MachineMode(cmd) {
					return shared.WriteJSON(cmd, conditionValidationResult{Project: project, Source: "firebase", Valid: true})
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Valid: %s (%s) · firebase\n", project.Name, project.ProjectID)
				return nil
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
				return shared.WriteJSON(cmd, conditionValidationResult{Project: project, Source: source, Valid: true})
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Valid: %s (%s) · %s\n", project.Name, project.ProjectID, source)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print validation result as JSON")
	return cmd
}

type conditionValidationResult struct {
	Project core.Project `json:"project"`
	Source  string       `json:"source" contract:"enum=draft|firebase"`
	Valid   bool         `json:"valid"`
}

func addMutationFlags(cmd invocation.FlagGroups) {
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddDryRunFlag(cmd)
	shared.AddChangeNoteFlag(cmd)
	shared.AddYesFlag(cmd, "Print diff and apply without confirmation")
	shared.AddPlanOutFlag(cmd)
	cmd.Flags().Bool("json", false, "Print mutation results as JSON")
}

func readMutationOptions(cmd invocation.Call) mutationOptions {
	var projectFilters []string
	projectFilterChanged := false
	if cmd.Flags().Lookup("project") != nil {
		projectFilters, _ = cmd.Flags().GetStringArray("project")
		projectFilterChanged = cmd.Flags().Changed("project")
	}
	draft, _ := cmd.Flags().GetBool("draft")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	yes, _ := cmd.Flags().GetBool("yes")
	var changeNote *string
	if cmd.Flags().Changed("change-note") {
		value, _ := cmd.Flags().GetString("change-note")
		changeNote = &value
	}
	return mutationOptions{ProjectFilters: projectFilters, ProjectFilterChanged: projectFilterChanged, Draft: draft, DryRun: dryRun, Yes: yes, ChangeNote: changeNote}
}

func conditionAddArguments(args []string) (*string, string) {
	if len(args) == 1 {
		return nil, args[0]
	}
	return &args[0], args[1]
}

func conditionDeleteArguments(args []string) (*string, *string) {
	if len(args) == 0 {
		return nil, nil
	}
	if len(args) == 1 {
		return &args[0], nil
	}
	return &args[0], &args[1]
}

func runNamedConditionMutation(cmd invocation.Call, svc *core.Core, projectQuery, requestedName string, opts mutationOptions, operation, emoji string, destructive bool, mutate func(*firebase.RemoteConfig, string) error) error {
	return runConditionMutation(cmd, svc, projectQuery, opts, operation, emoji, destructive, func(cfg *firebase.RemoteConfig) error {
		name, ok := coreconditions.ResolveNameExact(cfg, requestedName)
		if !ok {
			return &shared.SelectionError{Resource: "condition", Kind: "not_found", Query: requestedName, Err: fmt.Errorf("condition %q not found", requestedName)}
		}
		return mutate(cfg, name)
	})
}

func runConditionMutation(cmd invocation.Call, svc *core.Core, projectQuery string, opts mutationOptions, operation, emoji string, destructive bool, mutate conditionMutation) error {
	return runConditionMutationTargets(cmd, svc, &projectQuery, opts, operation, emoji, destructive, mutate)
}

func runConditionMutationTargets(cmd invocation.Call, svc *core.Core, projectQuery *string, opts mutationOptions, operation, emoji string, destructive bool, mutate conditionMutation) error {
	return runConditionMutationPlans(cmd, svc, projectQuery, opts, operation, emoji, func(project core.Project, _ *sharedrc.ProjectConfig) (sharedrc.RemoteMutationPlan, error) {
		return sharedrc.RemoteMutationPlan{MatchedItemCount: 1, Mutation: func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			finalCfg, err := firebase.CloneRemoteConfig(current)
			if err != nil {
				return 0, nil, err
			}
			if err := mutate(finalCfg); err != nil {
				return 0, nil, typedConditionMutationError(project.ProjectID, err)
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
		}}, nil
	})
}

func runConditionMutationPlans(cmd invocation.Call, svc *core.Core, projectQuery *string, opts mutationOptions, operation, emoji string, plan sharedrc.RemoteMutationPlanner) error {
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
	projects, defaultScope, ctx, err := resolveConditionMutationProjects(ctx, cmd, svc, projectQuery, opts)
	if err != nil {
		return err
	}
	var totals sharedrc.RemoteMutationTotals
	if opts.Draft {
		totals, err = sharedrc.RunRemoteDraftLoop(ctx, cmd, svc, projects, defaultScope, operation, plan)
	} else {
		totals, err = sharedrc.RunRemotePublishLoop(ctx, cmd, svc, projects, defaultScope, operation, emoji, plan)
	}
	if writeErr := sharedrc.WriteRemoteMutationResults(cmd, totals, map[bool]string{true: "draft", false: "publish"}[opts.Draft], emoji); writeErr != nil {
		return writeErr
	}
	return err
}

func resolveConditionMutationProjects(ctx context.Context, cmd invocation.Call, svc *core.Core, projectQuery *string, opts mutationOptions) ([]core.Project, bool, context.Context, error) {
	if projectQuery != nil {
		if opts.ProjectFilterChanged {
			return nil, false, ctx, shared.InvalidArgument(fmt.Errorf("<project> and --project cannot be used together"))
		}
		project, err := shared.ResolveProjectTargetForExecution(ctx, cmd, svc, *projectQuery)
		if err != nil {
			return nil, false, ctx, err
		}
		ctx, err = shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
		if err != nil {
			return nil, false, ctx, err
		}
		return []core.Project{project}, false, ctx, nil
	}
	projects, ctx, err := shared.ResolveProjectMutationTargetsForExecution(ctx, cmd, svc, opts.ProjectFilters)
	if err != nil {
		return nil, false, ctx, err
	}
	strfold.SortProjects(projects, func(project core.Project) string { return project.Name }, func(project core.Project) string { return project.ProjectID })
	return projects, !opts.ProjectFilterChanged, ctx, nil
}

func typedConditionMutationError(projectID string, err error) error {
	var conflict *shared.ConflictError
	var selection *shared.SelectionError
	var validation *shared.ValidationError
	if errors.As(err, &conflict) || errors.As(err, &selection) || errors.As(err, &validation) {
		return err
	}
	return &shared.ValidationError{Code: "condition.invalid", Source: "command", Stage: "mutation", Target: projectID, Err: err}
}
