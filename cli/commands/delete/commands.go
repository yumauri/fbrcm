package deletecmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

type deleteOptions struct {
	shared.ParameterMutationOpts
}

type deleteTotals struct {
	modifiedProjects int
	deletedParams    int
}

// New constructs the delete command.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [parameter]",
		Short: "Delete Remote Config parameters",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteCommand(cmd, svc, args)
		},
	}

	addDeleteFlags(cmd)
	return cmd
}

func addDeleteFlags(cmd *cobra.Command) {
	shared.AddProjectFilterFlag(cmd)
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	shared.AddYesFlag(cmd, "Print diff and delete without confirmation")
}

func runDeleteCommand(cmd *cobra.Command, svc *core.Core, args []string) error {
	opts, err := readDeleteOptions(cmd, args)
	if err != nil {
		return err
	}
	if shared.StdinAvailable(cmd.InOrStdin()) {
		corelog.For("delete").Info("stdin mode enabled; using remote config from stdin")
		return runDeleteStdin(cmd, opts.ParamFilters, opts.ParamExpr, opts.Search)
	}
	return runDeleteRemote(cmd, svc, opts)
}

func readDeleteOptions(cmd *cobra.Command, args []string) (deleteOptions, error) {
	opts, err := shared.ReadParameterMutationOpts(cmd, args)
	if err != nil {
		return deleteOptions{}, err
	}
	return deleteOptions{ParameterMutationOpts: opts}, nil
}

func runDeleteRemote(cmd *cobra.Command, svc *core.Core, opts deleteOptions) error {
	totals, err := shared.RunParameterMutationRemote(cmd, svc, opts.ParameterMutationOpts, "delete", "🗑️", func(cmd *cobra.Command, project core.Project, current *firebase.RemoteConfig, matched []shared.ParamTarget, yes bool) (int, *firebase.RemoteConfig, error) {
		deleted, finalCfg, err := confirmAndDeleteProject(cmd, project.ProjectID, current, matched, yes, cmd.ErrOrStderr())
		if err != nil {
			return 0, nil, err
		}
		return len(deleted), finalCfg, nil
	})
	if err != nil {
		return err
	}

	logDeleteTotals("remote", deleteTotals{modifiedProjects: totals.ModifiedProjects, deletedParams: totals.ChangedParams})
	return nil
}

func runDeleteStdin(cmd *cobra.Command, paramFilters []string, projectExpr string, search shared.ParameterSearch) error {
	cfg, remoteConfigRaw, err := rc.ReadRemoteConfigInput(cmd.InOrStdin())
	if err != nil {
		return err
	}
	compiledExpr, ok := shared.CompileExpr(projectExpr, "<stdin>")
	if !ok {
		return nil
	}

	project := core.Project{Name: "<stdin>", ProjectID: "<stdin>"}
	matched := shared.CollectMatchingParamTargets(project, cfg, paramFilters, search, compiledExpr, shared.DefaultRootGroupLabel)
	deleted, finalCfg, err := confirmAndDeleteProject(cmd, "<stdin>", cfg, matched, true, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	if err := rc.WriteOrderPreservingRemoteConfigStdout(cmd, finalCfg, remoteConfigRaw); err != nil {
		return err
	}

	totals := deleteTotals{deletedParams: len(deleted)}
	if len(deleted) > 0 {
		totals.modifiedProjects = 1
	}
	logDeleteTotals("stdin", totals)
	return nil
}

func confirmAndDeleteProject(cmd *cobra.Command, label string, cfg *firebase.RemoteConfig, matched []shared.ParamTarget, yes bool, diffOut io.Writer) ([]shared.ParamTarget, *firebase.RemoteConfig, error) {
	return shared.ConfirmParamTargets(cmd, label, cfg, matched, yes, diffOut, func(target shared.ParamTarget, finalCfg *firebase.RemoteConfig) (shared.ParamTargetMutationStep, error) {
		return shared.ParamTargetMutationStep{
			DiffText:    rc.RenderRemovedParameterDetail(target.Key, target.Group, target.Param),
			Prompt:      fmt.Sprintf("Delete %s from %s?", rcdisplay.FormatParameterHeader(target.Key, target.Group), label),
			Destructive: true,
			Apply: func(cfg *firebase.RemoteConfig) (*firebase.RemoteConfig, error) {
				shared.RemoveParamSlot(cfg, target.Key, target.Group)
				return nil, nil
			},
		}, nil
	})
}

func logDeleteTotals(mode string, totals deleteTotals) {
	corelog.For("delete").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.deletedParams)
}
