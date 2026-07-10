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
	value     string
	valueType string
}

type addOptions struct {
	projectFilters []string
	projectExpr    string
	dryRun         bool
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
	shared.AddProjectFilterFlag(cmd)
	cmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	cmd.Flags().String("description", "", "Parameter description")
	cmd.Flags().String("group", "", "Target parameter group")
	cmd.Flags().String("boolean", "", "Boolean parameter value: true or false")
	cmd.Flags().String("number", "", "Number parameter value")
	cmd.Flags().String("string", "", "String parameter value")
	cmd.Flags().String("json", "", "JSON parameter value")
	cmd.MarkFlagsMutuallyExclusive("boolean", "number", "string", "json")
}

func runAddCommand(cmd *cobra.Command, svc *core.Core, args []string) error {
	opts, err := readAddOptions(cmd, args)
	if err != nil {
		return err
	}
	if shared.StdinAvailable(cmd.InOrStdin()) {
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
	return addValueSpec{value: value.Value, valueType: value.Type}, nil
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
	projects = shared.FilterProjects(projects, opts.projectFilters)
	projects, err = shared.FilterProjectsByExpr(ctx, svc, projects, opts.projectExpr)
	if err != nil {
		return err
	}
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })

	totals, err := rc.RunRemotePublishLoop(ctx, cmd, svc, projects, "add", "➕", func(project core.Project, _ *rc.ProjectConfig) (rc.RemoteConfigMutation, error) {
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			changed, finalCfg, err := addParameter(current, opts.key, opts.groupName, opts.description, opts.valueSpec)
			if err != nil {
				return 0, nil, err
			}
			if !changed {
				corelog.For("add").Error("parameter already exists; skipping", "project_id", project.ProjectID, "parameter", opts.key)
				return 0, finalCfg, nil
			}
			diffText, hasChanges := rc.RenderRemoteConfigDiff(current, finalCfg)
			if !hasChanges {
				return 0, finalCfg, nil
			}
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
			return 1, finalCfg, nil
		}, nil
	})
	if err != nil {
		return err
	}

	logAddTotals("remote", addTotals{modifiedProjects: totals.ModifiedProjects, addedParams: totals.ChangedParams})
	return nil
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
		DefaultValue: &firebase.RemoteConfigValue{Value: spec.value},
		Description:  description,
		ValueType:    spec.valueType,
	}

	shared.SetParamSlot(finalCfg, key, groupName, param)
	return true, finalCfg, nil
}

func logAddTotals(mode string, totals addTotals) {
	corelog.For("add").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.addedParams)
}
