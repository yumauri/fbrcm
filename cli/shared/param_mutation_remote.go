package shared

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/strfold"
)

// ParameterMutationOpts holds flags shared by delete and update remote mutation commands.
type ParameterMutationOpts struct {
	ProjectFilters []string
	ParamExpr      string
	ParamFilters   []string
	Search         ParameterSearch
	Yes            bool
	DryRun         bool
	Draft          bool
	ChangeNote     *string
}

// ReadParameterMutationOpts reads project/filter/search/expr/dry-run/yes flags and resolves
// an optional positional parameter argument into filter queries.
func ReadParameterMutationOpts(cmd *cobra.Command, args []string) (ParameterMutationOpts, error) {
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
	if len(args) > 0 {
		paramFilters, err = ResolveParameterArgFilters(args, paramFilters)
		if err != nil {
			return ParameterMutationOpts{}, err
		}
	}
	return ParameterMutationOpts{
		ProjectFilters: projectFilters,
		ParamExpr:      paramExpr,
		ParamFilters:   paramFilters,
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
type ParameterMutationApplyFn func(cmd *cobra.Command, project core.Project, current *firebase.RemoteConfig, matched []ParamTarget, yes bool) (int, *firebase.RemoteConfig, error)

// RunParameterMutationRemote lists, filters, and publishes parameter mutations across projects.
func RunParameterMutationRemote(cmd *cobra.Command, svc *core.Core, opts ParameterMutationOpts, operation, emoji string, apply ParameterMutationApplyFn) (rc.RemoteMutationTotals, error) {
	ctx := context.Background()
	if opts.DryRun {
		ctx = firebase.WithDryRun(ctx)
	}
	ctx, err := WithChangeNote(ctx, opts.ChangeNote)
	if err != nil {
		return rc.RemoteMutationTotals{}, err
	}

	progress.Start("Loading projects…")
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return rc.RemoteMutationTotals{}, err
	}
	projects, err = FilterProjectTargets(projects, opts.ProjectFilters)
	if err != nil {
		return rc.RemoteMutationTotals{}, err
	}
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })
	compiledExpr, err := CompileExpr(opts.ParamExpr, "")
	if err != nil {
		return rc.RemoteMutationTotals{}, err
	}

	plan := func(project core.Project, cfg *rc.ProjectConfig) (rc.RemoteConfigMutation, error) {
		matched, err := CollectMatchingParamTargets(project, cfg.Config, opts.ParamFilters, opts.Search, compiledExpr, DefaultRootGroupLabel)
		if err != nil {
			return nil, err
		}
		if len(matched) == 0 {
			return nil, nil
		}
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			return apply(cmd, project, current, matched, opts.Yes)
		}, nil
	}
	if opts.Draft {
		progress.Start("Preparing Remote Config drafts…")
		return rc.RunRemoteDraftLoop(ctx, cmd, svc, projects, operation, plan)
	}
	progress.Start("Preparing Remote Config changes…")
	return rc.RunRemotePublishLoop(ctx, cmd, svc, projects, operation, emoji, plan)
}
