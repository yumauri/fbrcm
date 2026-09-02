package shared

import (
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared/rc"
)

// ParameterMutationOpts holds flags shared by delete and update remote mutation commands.
type ParameterMutationOpts struct {
	ProjectFilters []string
	ParamExpr      string
	ParamFilters   []string
	ParamArgument  *string
	Search         ParameterSearch
	Yes            bool
	DryRun         bool
	Draft          bool
	ChangeNote     *string
}

// ReadParameterMutationOpts reads project/filter/search/expr/dry-run/yes flags and resolves
// an optional positional parameter argument into filter queries.
func ReadParameterMutationOpts(cmd invocation.Call, args []string) (ParameterMutationOpts, error) {
	projectFilters, err := cmd.Flags().GetStringArray("project")
	if err != nil {
		return ParameterMutationOpts{}, err
	}
	paramExpr, err := cmd.Flags().GetString("expr")
	if err != nil {
		return ParameterMutationOpts{}, err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return ParameterMutationOpts{}, err
	}
	draftMode := false
	if cmd.Flags().Lookup("draft") != nil {
		draftMode, err = cmd.Flags().GetBool("draft")
		if err != nil {
			return ParameterMutationOpts{}, err
		}
	}
	paramFilters, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return ParameterMutationOpts{}, err
	}
	searchValue, err := cmd.Flags().GetString("search")
	if err != nil {
		return ParameterMutationOpts{}, err
	}
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return ParameterMutationOpts{}, err
	}
	changeNote, err := ReadChangeNoteFlag(cmd)
	if err != nil {
		return ParameterMutationOpts{}, err
	}
	paramArgument, err := ResolveParameterArgument(args, paramFilters)
	if err != nil {
		return ParameterMutationOpts{}, err
	}
	return ParameterMutationOpts{
		ProjectFilters: projectFilters,
		ParamExpr:      paramExpr,
		ParamFilters:   paramFilters,
		ParamArgument:  paramArgument,
		Search:         NewParameterSearch(searchValue),
		Yes:            yes,
		DryRun:         dryRun,
		Draft:          draftMode,
		ChangeNote:     changeNote,
	}, nil
}

// ParameterMutationApplyFn mutates matched parameter targets in a project config after
// optional per-target confirmation. It returns the number of applied changes and the
// resulting config snapshot.
type ParameterMutationApplyFn func(cmd invocation.Call, project core.Project, current *firebase.RemoteConfig, matched []ParamTarget, yes bool) (int, *firebase.RemoteConfig, error)

// RunParameterMutationRemote lists, filters, and publishes parameter mutations across projects.
func RunParameterMutationRemote(cmd invocation.Call, svc *core.Core, opts ParameterMutationOpts, operation, emoji string, apply ParameterMutationApplyFn) (rc.RemoteMutationTotals, error) {
	ctx := CommandContext(cmd)
	if err := RejectStatelessDraft(ctx, opts.Draft); err != nil {
		return rc.RemoteMutationTotals{}, err
	}
	if opts.DryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	ctx, err := WithChangeNote(ctx, opts.ChangeNote)
	if err != nil {
		return rc.RemoteMutationTotals{}, err
	}

	projects, ctx, err := ResolveProjectMutationTargetsForExecution(ctx, cmd, svc, opts.ProjectFilters)
	if err != nil {
		return rc.RemoteMutationTotals{}, err
	}
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })
	compiledExpr, err := CompileExpr(opts.ParamExpr, "")
	if err != nil {
		return rc.RemoteMutationTotals{}, err
	}

	plan := func(project core.Project, cfg *rc.ProjectConfig) (rc.RemoteMutationPlan, error) {
		matched, err := CollectMatchingParamTargetsWithArgument(project, cfg.Config, opts.ParamFilters, opts.ParamArgument, opts.Search, compiledExpr, DefaultRootGroupLabel)
		if err != nil {
			return rc.RemoteMutationPlan{}, err
		}
		if len(matched) == 0 {
			return rc.RemoteMutationPlan{}, nil
		}
		return rc.RemoteMutationPlan{MatchedItemCount: len(matched), Mutation: func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			return apply(cmd, project, current, matched, opts.Yes)
		}}, nil
	}
	defaultScope := len(opts.ProjectFilters) == 0
	if opts.Draft {
		progress.Start("Preparing Remote Config drafts…")
		return rc.RunRemoteDraftLoop(ctx, cmd, svc, projects, defaultScope, operation, plan)
	}
	progress.Start("Preparing Remote Config changes…")
	return rc.RunRemotePublishLoop(ctx, cmd, svc, projects, defaultScope, operation, emoji, plan)
}
