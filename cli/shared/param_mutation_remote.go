package shared

import (
	"context"

	"github.com/spf13/cobra"

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

	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return rc.RemoteMutationTotals{}, err
	}
	projects = FilterProjects(projects, opts.ProjectFilters)
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })
	compiledExpr, ok := CompileExpr(opts.ParamExpr, "")
	if !ok {
		return rc.RemoteMutationTotals{}, nil
	}

	return rc.RunRemotePublishLoop(ctx, cmd, svc, projects, operation, emoji, func(project core.Project, cfg *rc.ProjectConfig) (rc.RemoteConfigMutation, error) {
		matched := CollectMatchingParamTargets(project, cfg.Config, opts.ParamFilters, opts.Search, compiledExpr, DefaultRootGroupLabel)
		if len(matched) == 0 {
			return nil, nil
		}
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			return apply(cmd, project, current, matched, opts.Yes)
		}, nil
	})
}
