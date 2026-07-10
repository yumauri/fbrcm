package updatecmd

import (
	"github.com/spf13/cobra"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func runUpdateRemote(cmd *cobra.Command, svc *core.Core, opts updateOptions) error {
	totals, err := shared.RunParameterMutationRemote(cmd, svc, opts.ParameterMutationOpts, "update", "✏️", func(cmd *cobra.Command, project core.Project, current *firebase.RemoteConfig, matched []shared.ParamTarget, yes bool) (int, *firebase.RemoteConfig, error) {
		updated, finalCfg, err := confirmAndUpdateProject(cmd, project.ProjectID, current, matched, opts.spec, yes, cmd.ErrOrStderr())
		if err != nil {
			return 0, nil, err
		}
		return len(updated), finalCfg, nil
	})
	if err != nil {
		return err
	}
	logUpdateTotals("remote", updateTotals{modifiedProjects: totals.ModifiedProjects, updatedParams: totals.ChangedParams})
	return nil
}

func runUpdateStdin(cmd *cobra.Command, paramFilters []string, paramExpr string, search shared.ParameterSearch, spec updateSpec) error {
	cfg, raw, err := rc.ReadRemoteConfigInput(cmd.InOrStdin())
	if err != nil {
		return err
	}
	compiledExpr, ok := shared.CompileExpr(paramExpr, "<stdin>")
	if !ok {
		return nil
	}
	project := core.Project{Name: "<stdin>", ProjectID: "<stdin>"}
	matched := shared.CollectMatchingParamTargets(project, cfg, paramFilters, search, compiledExpr, shared.DefaultRootGroupLabel)
	updated, finalCfg, err := confirmAndUpdateProject(cmd, "<stdin>", cfg, matched, spec, true, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	if err := rc.WriteOrderPreservingRemoteConfigStdout(cmd, finalCfg, raw); err != nil {
		return err
	}
	totals := updateTotals{updatedParams: len(updated)}
	if len(updated) > 0 {
		totals.modifiedProjects = 1
	}
	logUpdateTotals("stdin", totals)
	return nil
}

func logUpdateTotals(mode string, totals updateTotals) {
	corelog.For("update").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.updatedParams)
}
