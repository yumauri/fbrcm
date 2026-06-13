package addcmd

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type addValueSpec struct {
	value     string
	valueType string
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
			projectFilters, err := cmd.Flags().GetStringArray("project")
			if err != nil {
				return err
			}
			projectExpr, err := cmd.Flags().GetString("expr")
			if err != nil {
				return err
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			groupName, err := cmd.Flags().GetString("group")
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}
			spec, err := readAddValueSpec(cmd)
			if err != nil {
				return err
			}

			key := strings.TrimSpace(args[0])
			if key == "" {
				return fmt.Errorf("parameter key cannot be empty")
			}
			groupName = strings.TrimSpace(groupName)

			if shared.StdinAvailable(cmd.InOrStdin()) {
				corelog.For("add").Info("stdin mode enabled; using remote config from stdin")
				return runAddStdin(cmd, key, groupName, description, spec, projectExpr)
			}
			return runAddRemote(cmd, svc, key, projectFilters, projectExpr, groupName, description, spec, dryRun)
		},
	}

	cmd.Flags().StringArrayP("project", "p", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	cmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	cmd.Flags().Bool("dry-run", false, "Log Firebase write requests without sending them")
	cmd.Flags().String("description", "", "Parameter description")
	cmd.Flags().String("group", "", "Target parameter group")
	cmd.Flags().String("boolean", "", "Boolean parameter value: true or false")
	cmd.Flags().String("number", "", "Number parameter value")
	cmd.Flags().String("string", "", "String parameter value")
	cmd.Flags().String("json", "", "JSON parameter value")
	cmd.MarkFlagsMutuallyExclusive("boolean", "number", "string", "json")
	return cmd
}

func readAddValueSpec(cmd *cobra.Command) (addValueSpec, error) {
	value, err := shared.ReadValueFlag(cmd, true)
	if err != nil {
		return addValueSpec{}, err
	}
	return addValueSpec{value: value.Value, valueType: value.Type}, nil
}

func runAddRemote(cmd *cobra.Command, svc *core.Core, key string, projectFilters []string, projectExpr, groupName, description string, spec addValueSpec, dryRun bool) error {
	ctx := context.Background()
	if dryRun {
		ctx = firebase.WithDryRun(ctx)
	}

	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return err
	}
	projects = shared.FilterProjects(projects, projectFilters)
	projects, err = shared.FilterProjectsByExpr(ctx, svc, projects, projectExpr)
	if err != nil {
		return err
	}
	shared.SortProjects(projects)

	var totals addTotals
	for _, project := range projects {
		for {
			cfg, err := shared.RevalidateProjectConfig(ctx, svc, project)
			if err != nil {
				return err
			}

			changedCount, retry, err := shared.PublishProjectConfigMutation(ctx, svc, cfg, "add", cmd.ErrOrStderr(), func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
				changed, finalCfg := addParameter(current, key, groupName, description, spec)
				if !changed {
					corelog.For("add").Error("parameter already exists; skipping", "project_id", project.ProjectID, "parameter", key)
					return 0, finalCfg, nil
				}
				diffText, hasChanges := shared.RenderRemoteConfigDiff(current, finalCfg)
				if !hasChanges {
					return 0, finalCfg, nil
				}
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
				return 1, finalCfg, nil
			})
			if err != nil {
				return err
			}
			if changedCount == 0 {
				break
			}
			if retry {
				continue
			}

			totals.modifiedProjects++
			totals.addedParams += changedCount
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "➕ published: %s\n", project.ProjectID)
			break
		}
	}

	logAddTotals("remote", totals)
	return nil
}

func runAddStdin(cmd *cobra.Command, key, groupName, description string, spec addValueSpec, projectExpr string) error {
	cfg, remoteConfigRaw, err := shared.ReadRemoteConfigInput(cmd.InOrStdin())
	if err != nil {
		return err
	}

	if !shared.MatchProjectByExpr(core.Project{Name: "<stdin>", ProjectID: "<stdin>"}, cfg, projectExpr) {
		return nil
	}

	order, err := shared.ParseRemoteConfigOrder(remoteConfigRaw)
	if err != nil {
		return fmt.Errorf("parse stdin remote config order: %w", err)
	}

	changed, finalCfg := addParameter(cfg, key, groupName, description, spec)
	if !changed {
		corelog.For("add").Error("parameter already exists; skipping", "project_id", "<stdin>", "parameter", key)
	} else {
		diffText, hasChanges := shared.RenderRemoteConfigDiff(cfg, finalCfg)
		if hasChanges {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
		}
		if groupName == "" {
			order.Parameters = append(order.Parameters, key)
		} else {
			if !slices.Contains(order.Groups, groupName) {
				order.Groups = append(order.Groups, groupName)
			}
			order.GroupParameters[groupName] = append(order.GroupParameters[groupName], key)
		}
	}

	out, err := shared.MarshalPrettyRemoteConfigWithOrder(finalCfg, order)
	if err != nil {
		return err
	}
	if _, err := cmd.OutOrStdout().Write(out); err != nil {
		return err
	}
	if len(out) == 0 || out[len(out)-1] != '\n' {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	totals := addTotals{}
	if changed {
		totals.modifiedProjects = 1
		totals.addedParams = 1
	}
	logAddTotals("stdin", totals)
	return nil
}

func addParameter(cfg *firebase.RemoteConfig, key, groupName, description string, spec addValueSpec) (bool, *firebase.RemoteConfig) {
	finalCfg := shared.CloneRemoteConfig(cfg)
	if shared.ParamExists(finalCfg, key) {
		return false, finalCfg
	}

	param := firebase.RemoteConfigParam{
		DefaultValue: &firebase.RemoteConfigValue{Value: spec.value},
		Description:  description,
		ValueType:    spec.valueType,
	}

	shared.SetParamSlot(finalCfg, key, groupName, param)
	return true, finalCfg
}

func logAddTotals(mode string, totals addTotals) {
	corelog.For("add").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.addedParams)
}
