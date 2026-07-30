package addcmd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
)

type addValueSpec struct {
	value           string
	valueType       string
	useInAppDefault bool
}

type addOptions struct {
	projectFilters []string
	projectExpr    string
	dryRun         bool
	draft          bool
	yes            bool
	groupName      string
	description    string
	valueSpec      addValueSpec
	key            string
}

type addTotals struct {
	modifiedProjects int
	addedParams      int
}

// New constructs the add command.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <parameter>",
		Short: "Add Remote Config parameter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddCommand(cmd, svc, args)
		},
	}

	addFlags(cmd)
	return cmd
}

func addFlags(cmd *cobra.Command) {
	shared.AddProjectTargetFilterFlag(cmd)
	cmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	cmd.Flags().Bool("draft", false, "Save changes to a local draft instead of publishing")
	shared.AddYesFlag(cmd, "Print diff and add without confirmation")
	cmd.Flags().String("description", "", "Parameter description")
	cmd.Flags().String("group", "", "Target parameter group")
	cmd.Flags().String("type", "", "Parameter type: string, boolean, number, or json")
	cmd.Flags().String("value", "", "Parameter value interpreted according to --type")
	cmd.Flags().Bool("use-in-app-default", false, "Use the application's default value")
	cmd.Flags().Bool("json", false, "Print mutation results as JSON")
	cmd.MarkFlagsMutuallyExclusive("value", "use-in-app-default")
}

func runAddCommand(cmd *cobra.Command, svc *core.Core, args []string) error {
	opts, err := readAddOptions(cmd, args)
	if err != nil {
		return err
	}
	if shared.StdinAvailable(cmd.InOrStdin()) {
		if opts.draft {
			return fmt.Errorf("--draft is unavailable in stdin mode")
		}
		corelog.For("add").Info("stdin mode enabled; using remote config from stdin")
		return runAddStdin(cmd, opts.key, opts.groupName, opts.description, opts.valueSpec, opts.projectExpr)
	}
	return runAddRemote(cmd, svc, opts)
}

func readAddOptions(cmd *cobra.Command, args []string) (addOptions, error) {
	projectFilters, err := cmd.Flags().GetStringArray("project")
	if err != nil {
		return addOptions{}, err
	}
	projectExpr, err := cmd.Flags().GetString("expr")
	if err != nil {
		return addOptions{}, err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return addOptions{}, err
	}
	draftMode, err := cmd.Flags().GetBool("draft")
	if err != nil {
		return addOptions{}, err
	}
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return addOptions{}, err
	}
	groupName, err := cmd.Flags().GetString("group")
	if err != nil {
		return addOptions{}, err
	}
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return addOptions{}, err
	}
	spec, err := readAddValueSpec(cmd)
	if err != nil {
		return addOptions{}, err
	}

	key := strings.TrimSpace(args[0])
	if key == "" {
		return addOptions{}, fmt.Errorf("parameter key cannot be empty")
	}

	return addOptions{
		projectFilters: projectFilters,
		projectExpr:    projectExpr,
		dryRun:         dryRun,
		draft:          draftMode,
		yes:            yes,
		groupName:      strings.TrimSpace(groupName),
		description:    description,
		valueSpec:      spec,
		key:            key,
	}, nil
}

func readAddValueSpec(cmd *cobra.Command) (addValueSpec, error) {
	value, err := shared.ReadValueFlag(cmd, true)
	if err != nil {
		return addValueSpec{}, err
	}
	return addValueSpec{value: value.Value, valueType: value.Type, useInAppDefault: value.UseInAppDefault}, nil
}

func runAddRemote(cmd *cobra.Command, svc *core.Core, opts addOptions) error {
	ctx := context.Background()
	if opts.dryRun {
		ctx = firebase.WithDryRun(ctx)
	}

	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return err
	}
	projects, err = shared.FilterProjectTargets(projects, opts.projectFilters)
	if err != nil {
		return err
	}
	projects, err = shared.FilterProjectsByExpr(ctx, svc, projects, opts.projectExpr)
	if err != nil {
		return err
	}
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })

	plan := func(project core.Project, _ *rc.ProjectConfig) (rc.RemoteConfigMutation, error) {
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			if shared.ParamExists(current, opts.key) {
				corelog.For("add").Error("parameter already exists; skipping", "project_id", project.ProjectID, "parameter", opts.key)
				return 0, current, nil
			}
			changed, finalCfg, err := addProject(cmd, project, current, opts.key, opts.groupName, opts.description, opts.valueSpec, opts.yes)
			if err != nil {
				return 0, nil, err
			}
			if !changed {
				return 0, finalCfg, nil
			}
			return 1, finalCfg, nil
		}, nil
	}
	var totals rc.RemoteMutationTotals
	if opts.draft {
		totals, err = rc.RunRemoteDraftLoop(ctx, cmd, svc, projects, "add", plan)
	} else {
		totals, err = rc.RunRemotePublishLoop(ctx, cmd, svc, projects, "add", "➕", plan)
	}
	logAddTotals("remote", addTotals{modifiedProjects: totals.ModifiedProjects, addedParams: totals.ChangedParams})
	if writeErr := rc.WriteRemoteMutationResults(cmd, totals, map[bool]string{true: "draft", false: "publish"}[opts.draft], "➕"); writeErr != nil {
		return writeErr
	}
	return err
}

func addProject(cmd *cobra.Command, project core.Project, current *firebase.RemoteConfig, key, groupName, description string, spec addValueSpec, yes bool) (bool, *firebase.RemoteConfig, error) {
	changed, finalCfg, err := addParameter(current, key, groupName, description, spec)
	if err != nil || !changed {
		return changed, finalCfg, err
	}
	diffText, hasChanges := rc.RenderRemoteConfigDiff(current, finalCfg)
	if !hasChanges {
		return false, finalCfg, nil
	}
	prompt := fmt.Sprintf("Add %s to %s?", shared.FormatParameterHeader(key, groupName), project.ProjectID)
	confirmed, err := shared.PrintDiffAndConfirm(cmd, yes, cmd.ErrOrStderr(), diffText, prompt, false)
	if err != nil || !confirmed {
		return false, finalCfg, err
	}
	return true, finalCfg, nil
}

func runAddStdin(cmd *cobra.Command, key, groupName, description string, spec addValueSpec, projectExpr string) error {
	cfg, remoteConfigRaw, err := rc.ReadRemoteConfigInput(cmd.InOrStdin())
	if err != nil {
		return err
	}

	if !shared.MatchProjectByExpr(core.Project{Name: "<stdin>", ProjectID: "<stdin>"}, cfg, projectExpr) {
		return nil
	}

	changed, finalCfg, err := addParameter(cfg, key, groupName, description, spec)
	if err != nil {
		return err
	}
	if !changed {
		corelog.For("add").Error("parameter already exists; skipping", "project_id", "<stdin>", "parameter", key)
	} else {
		diffText, hasChanges := rc.RenderRemoteConfigDiff(cfg, finalCfg)
		if hasChanges {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
		}
	}

	var mutate rc.OrderMutator
	if changed {
		mutate = func(order *rc.RemoteConfigOrder) {
			if groupName == "" {
				order.Parameters = append(order.Parameters, key)
				return
			}
			if !slices.Contains(order.Groups, groupName) {
				order.Groups = append(order.Groups, groupName)
			}
			order.GroupParameters[groupName] = append(order.GroupParameters[groupName], key)
		}
	}
	if err := rc.WriteOrderPreservingRemoteConfigStdoutWithOrder(cmd, finalCfg, remoteConfigRaw, mutate); err != nil {
		return err
	}

	totals := addTotals{}
	if changed {
		totals.modifiedProjects = 1
		totals.addedParams = 1
	}
	logAddTotals("stdin", totals)
	return nil
}

func addParameter(cfg *firebase.RemoteConfig, key, groupName, description string, spec addValueSpec) (bool, *firebase.RemoteConfig, error) {
	finalCfg, err := firebase.CloneRemoteConfig(cfg)
	if err != nil {
		return false, nil, err
	}
	if shared.ParamExists(finalCfg, key) {
		return false, finalCfg, nil
	}

	param := firebase.RemoteConfigParam{
		DefaultValue: &firebase.RemoteConfigValue{Value: spec.value, UseInAppDefault: spec.useInAppDefault},
		Description:  description,
		ValueType:    spec.valueType,
	}

	shared.SetParamSlot(finalCfg, key, groupName, param)
	return true, finalCfg, nil
}

func logAddTotals(mode string, totals addTotals) {
	corelog.For("add").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.addedParams)
}
