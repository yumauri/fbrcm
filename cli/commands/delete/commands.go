package deletecmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
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
	contract.RegisterResponse(cmd, []rc.RemoteMutationJSONResult{}, rc.PlanCreatedResult{}, contract.ArtifactData{})
	return cmd
}

func addDeleteFlags(cmd *cobra.Command) {
	shared.AddProjectTargetFilterFlag(cmd)
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	shared.AddChangeNoteFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddYesFlag(cmd, "Print diff and delete without confirmation")
	shared.AddPlanOutFlag(cmd)
	cmd.Flags().Bool("json", false, "Print mutation results as JSON")
}

func runDeleteCommand(cmd *cobra.Command, svc *core.Core, args []string) error {
	opts, err := readDeleteOptions(cmd, args)
	if err != nil {
		return err
	}
	if shared.StdinAvailable(cmd.InOrStdin()) {
		if err := shared.RejectChangedFlags(cmd, "stdin mode", "project", "dry-run", "draft", "change-note", "plan-out"); err != nil {
			return err
		}
		corelog.For("delete").Info("stdin mode enabled; using remote config from stdin")
		return runDeleteStdin(cmd, opts.ParamFilters, opts.ParamArgument, opts.ParamExpr, opts.Search)
	}
	return runDeleteRemote(cmd, svc, opts)
}

func readDeleteOptions(cmd *cobra.Command, args []string) (deleteOptions, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) == "" {
		return deleteOptions{}, shared.InvalidArgument(fmt.Errorf("parameter argument cannot be empty"))
	}
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
	logDeleteTotals("remote", deleteTotals{modifiedProjects: totals.ModifiedProjects, deletedParams: totals.ChangedParams})
	if writeErr := rc.WriteRemoteMutationResults(cmd, totals, map[bool]string{true: "draft", false: "publish"}[opts.Draft], "🗑️"); writeErr != nil {
		return writeErr
	}
	return err
}

func runDeleteStdin(cmd *cobra.Command, paramFilters []string, paramArgument *string, projectExpr string, search shared.ParameterSearch) error {
	cfg, remoteConfigRaw, err := rc.ReadRemoteConfigInput(cmd.InOrStdin())
	if err != nil {
		return err
	}
	compiledExpr, err := shared.CompileExpr(projectExpr, "<stdin>")
	if err != nil {
		return err
	}

	project := core.Project{Name: "<stdin>", ProjectID: "<stdin>"}
	matched, err := shared.CollectMatchingParamTargetsWithArgument(project, cfg, paramFilters, paramArgument, search, compiledExpr, shared.DefaultRootGroupLabel)
	if err != nil {
		return err
	}
	deleted, finalCfg, err := confirmAndDeleteProject(cmd, "<stdin>", cfg, matched, true, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	out, err := rc.MarshalOrderPreservingRemoteConfig(finalCfg, remoteConfigRaw, nil)
	if err != nil {
		return err
	}
	if contract.Enabled(cmd) {
		target := "<stdin>"
		if err := shared.WriteJSON(cmd, contract.NewArtifact(&target, "application/json", out, nil, false)); err != nil {
			return err
		}
	} else if err := rc.WriteOrderPreservingRemoteConfigStdout(cmd, finalCfg, remoteConfigRaw); err != nil {
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
	deleted, finalCfg, err := shared.ConfirmParamTargets(cmd, label, cfg, matched, yes, diffOut, func(target shared.ParamTarget, finalCfg *firebase.RemoteConfig) (shared.ParamTargetMutationStep, error) {
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
	if err == nil {
		err = rcmutate.EnsureOpaqueValuesUnchanged(cfg, finalCfg)
	}
	return deleted, finalCfg, err
}

func logDeleteTotals(mode string, totals deleteTotals) {
	corelog.For("delete").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.deletedParams)
}
