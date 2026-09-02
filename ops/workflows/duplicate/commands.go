package duplicatecmd

import (
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/draft"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	sharedrc "github.com/yumauri/fbrcm/ops/shared/rc"
)

type duplicateOptions struct {
	projectFilters []string
	projectExpr    string
	dryRun         bool
	draft          bool
	yes            bool
	source         string
	target         string
	changeNote     *string
}

// New constructs the duplicate command.
func NewDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "duplicate <source> <target>",
		Short: "Duplicate a Remote Config parameter",
		Args:  invocation.ExactArgs(2),
		RunE: func(cmd invocation.Call, args []string) error {
			opts, err := readDuplicateOptions(cmd, args)
			if err != nil {
				return err
			}
			return runDuplicateRemote(cmd, svc, opts)
		},
	}
	shared.AddProjectTargetFilterFlag(cmd)
	cmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	shared.AddChangeNoteFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddYesFlag(cmd, "Print diff and duplicate without confirmation")
	shared.AddPlanOutFlag(cmd)
	cmd.Flags().Bool("json", false, "Print mutation results as JSON")
	invocation.RegisterResponse(cmd, []sharedrc.RemoteMutationJSONResult{}, sharedrc.PlanCreatedResult{})
	return cmd
}

func readDuplicateOptions(cmd invocation.Call, args []string) (duplicateOptions, error) {
	projectFilters, err := cmd.Flags().GetStringArray("project")
	if err != nil {
		return duplicateOptions{}, err
	}
	projectExpr, err := cmd.Flags().GetString("expr")
	if err != nil {
		return duplicateOptions{}, err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return duplicateOptions{}, err
	}
	draft, err := cmd.Flags().GetBool("draft")
	if err != nil {
		return duplicateOptions{}, err
	}
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return duplicateOptions{}, err
	}
	changeNote, err := shared.ReadChangeNoteFlag(cmd)
	if err != nil {
		return duplicateOptions{}, err
	}
	source := args[0]
	target := strings.TrimSpace(args[1])
	if strings.TrimSpace(source) == "" {
		return duplicateOptions{}, shared.InvalidArgument(fmt.Errorf("source parameter key cannot be empty"))
	}
	if target == "" {
		return duplicateOptions{}, shared.InvalidArgument(fmt.Errorf("target parameter key cannot be empty"))
	}
	if source == target {
		return duplicateOptions{}, shared.InvalidArgument(fmt.Errorf("source and target parameter keys must differ"))
	}
	return duplicateOptions{
		projectFilters: projectFilters,
		projectExpr:    projectExpr,
		dryRun:         dryRun,
		draft:          draft,
		yes:            yes,
		source:         source,
		target:         target,
		changeNote:     changeNote,
	}, nil
}

func runDuplicateRemote(cmd invocation.Call, svc *core.Core, opts duplicateOptions) error {
	ctx := shared.CommandContext(cmd)
	if err := shared.RejectStatelessDraft(ctx, opts.draft); err != nil {
		return err
	}
	if opts.dryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	ctx, err := shared.WithChangeNote(ctx, opts.changeNote)
	if err != nil {
		return err
	}
	projects, ctx, err := shared.ResolveProjectMutationTargetsForExecution(ctx, cmd, svc, opts.projectFilters)
	if err != nil {
		return err
	}
	projects, err = shared.FilterProjectsByExpr(ctx, svc, projects, opts.projectExpr)
	if err != nil {
		return err
	}
	strfold.SortProjects(projects, func(project core.Project) string { return project.Name }, func(project core.Project) string { return project.ProjectID })

	plan := func(project core.Project, cfg *sharedrc.ProjectConfig) (sharedrc.RemoteMutationPlan, error) {
		_, found, err := resolveSource(cfg.Config, opts.source)
		if err != nil {
			return sharedrc.RemoteMutationPlan{}, err
		}
		if !found {
			return sharedrc.RemoteMutationPlan{}, nil
		}
		return sharedrc.RemoteMutationPlan{MatchedItemCount: 1, Mutation: func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			changed, finalCfg, err := duplicateProject(cmd, project, current, opts.source, opts.target, opts.yes)
			if err != nil {
				return 0, nil, err
			}
			if !changed {
				return 0, finalCfg, nil
			}
			return 1, finalCfg, nil
		}}, nil
	}
	var totals sharedrc.RemoteMutationTotals
	if opts.draft {
		totals, err = sharedrc.RunRemoteDraftLoop(ctx, cmd, svc, projects, len(opts.projectFilters) == 0 && strings.TrimSpace(opts.projectExpr) == "", "duplicate", plan)
	} else {
		totals, err = sharedrc.RunRemotePublishLoop(ctx, cmd, svc, projects, len(opts.projectFilters) == 0 && strings.TrimSpace(opts.projectExpr) == "", "duplicate", "📋", plan)
	}
	corelog.For("duplicate").Info("total", "projects", totals.ModifiedProjects, "parameters", totals.ChangedParams)
	if writeErr := sharedrc.WriteRemoteMutationResults(cmd, totals, map[bool]string{true: "draft", false: "publish"}[opts.draft], "📋"); writeErr != nil {
		return writeErr
	}
	return err
}

func duplicateProject(cmd invocation.Call, project core.Project, current *firebase.RemoteConfig, source, target string, yes bool) (bool, *firebase.RemoteConfig, error) {
	changed, finalCfg, sourceTarget, err := duplicateParameter(current, source, target)
	if err != nil || !changed {
		return changed, finalCfg, err
	}
	diffText, hasChanges := sharedrc.RenderRemoteConfigDiff(current, finalCfg)
	if !hasChanges {
		return false, finalCfg, nil
	}
	prompt := fmt.Sprintf(
		"Duplicate %s as %s in %s?",
		shared.FormatParameterHeader(sourceTarget.Key, sourceTarget.Group),
		target,
		project.ProjectID,
	)
	confirmed, err := shared.PrintDiffAndConfirm(cmd, yes, cmd.ErrOrStderr(), diffText, prompt, false)
	if err != nil || !confirmed {
		return false, finalCfg, err
	}
	return true, finalCfg, nil
}

func duplicateParameter(cfg *firebase.RemoteConfig, source, target string) (bool, *firebase.RemoteConfig, shared.ParamTarget, error) {
	finalCfg, err := firebase.CloneRemoteConfig(cfg)
	if err != nil {
		return false, nil, shared.ParamTarget{}, err
	}
	sourceTarget, found, err := resolveSource(finalCfg, source)
	if err != nil {
		return false, finalCfg, sourceTarget, err
	}
	if !found {
		return false, finalCfg, sourceTarget, nil
	}
	if paramExists(finalCfg, target) {
		return false, finalCfg, sourceTarget, &shared.ConflictError{Code: "parameter.exists", Resource: "parameter", Target: target, Err: fmt.Errorf("target parameter %q already exists", target)}
	}
	if err := draft.DuplicateParameterNamed(sourceTarget.Group, sourceTarget.Key, target)(finalCfg); err != nil {
		return false, finalCfg, sourceTarget, err
	}
	return true, finalCfg, sourceTarget, nil
}

func paramExists(cfg *firebase.RemoteConfig, requested string) bool {
	for _, target := range shared.CollectParamTargets(cfg) {
		if target.Key == requested {
			return true
		}
	}
	return false
}

func resolveSource(cfg *firebase.RemoteConfig, requested string) (shared.ParamTarget, bool, error) {
	var matches []shared.ParamTarget
	for _, target := range shared.CollectParamTargets(cfg) {
		if target.Key == requested {
			matches = append(matches, target)
		}
	}
	if len(matches) == 0 {
		return shared.ParamTarget{}, false, nil
	}
	if len(matches) > 1 {
		candidates := make([]shared.SelectionCandidate, 0, len(matches))
		for _, match := range matches {
			candidates = append(candidates, shared.SelectionCandidate{Name: shared.FormatParameterHeader(match.Key, match.Group), ID: match.Key})
		}
		return shared.ParamTarget{}, false, &shared.SelectionError{Resource: "parameter", Kind: "ambiguous", Query: requested, Candidates: candidates, Err: fmt.Errorf("source parameter %q is ambiguous across groups", requested)}
	}
	return matches[0], true, nil
}
