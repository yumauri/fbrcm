package updatecmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type valueSpec struct {
	value     string
	valueType string
}

type updateSpec struct {
	value                      *valueSpec
	name                       string
	group                      string
	description                string
	removeConditionalValues    []string
	nameChanged                bool
	groupChanged               bool
	descriptionChanged         bool
	removeAllConditionalValues bool
}

type updateTotals struct {
	modifiedProjects int
	updatedParams    int
}

const defaultGroupLabel = "(root)"

// New constructs the update command.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [parameter]",
		Short: "Update Remote Config parameters",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectFilters, err := cmd.Flags().GetStringArray("project")
			if err != nil {
				return err
			}
			paramExpr, err := cmd.Flags().GetString("expr")
			if err != nil {
				return err
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			paramFilters, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				return err
			}
			searchValue, err := cmd.Flags().GetString("search")
			if err != nil {
				return err
			}
			search := shared.NewParameterSearch(searchValue)
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			groupName, err := cmd.Flags().GetString("group")
			if err != nil {
				return err
			}
			noGroup, err := cmd.Flags().GetBool("no-group")
			if err != nil {
				return err
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}
			removeAllConditionalValues, err := cmd.Flags().GetBool("remove-all-conditional-values")
			if err != nil {
				return err
			}
			removeConditionalValues, err := readRemoveConditionalValues(cmd)
			if err != nil {
				return err
			}
			value, err := readValueSpec(cmd)
			if err != nil {
				return err
			}
			if len(args) > 0 {
				var err error
				paramFilters, err = shared.ResolveParameterArgFilters(args, paramFilters)
				if err != nil {
					return err
				}
			}

			groupChanged := cmd.Flags().Changed("group")
			if noGroup {
				groupChanged = true
				groupName = ""
			}
			descriptionChanged := cmd.Flags().Changed("description")
			nameChanged := cmd.Flags().Changed("name")
			groupName = strings.TrimSpace(groupName)
			name = strings.TrimSpace(name)
			if nameChanged && name == "" {
				return fmt.Errorf("--name cannot be empty")
			}

			spec := updateSpec{
				value:                      value,
				name:                       name,
				group:                      groupName,
				description:                description,
				removeConditionalValues:    removeConditionalValues,
				nameChanged:                nameChanged,
				groupChanged:               groupChanged,
				descriptionChanged:         descriptionChanged,
				removeAllConditionalValues: removeAllConditionalValues,
			}

			if shared.StdinAvailable(cmd.InOrStdin()) {
				corelog.For("update").Info("stdin mode enabled; using remote config from stdin")
				return runUpdateStdin(cmd, paramFilters, paramExpr, search, spec)
			}
			return runUpdateRemote(cmd, svc, projectFilters, paramExpr, paramFilters, search, spec, yes, dryRun)
		},
	}

	cmd.Flags().StringArrayP("project", "p", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter parameters by mode-prefixed query (^, /, ~, =); may be repeated")
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	cmd.Flags().String("search", "", "Search parameters by name, description, values, and conditions")
	cmd.Flags().Bool("dry-run", false, "Log Firebase write requests without sending them")
	cmd.Flags().BoolP("yes", "y", false, "Print diff and update without confirmation")
	cmd.Flags().String("description", "", "Parameter description")
	cmd.Flags().String("group", "", "Target parameter group")
	cmd.Flags().Bool("no-group", false, "Move parameter out of its group")
	cmd.Flags().String("name", "", "New parameter name")
	cmd.Flags().String("boolean", "", "Boolean parameter value: true or false")
	cmd.Flags().String("number", "", "Number parameter value")
	cmd.Flags().String("string", "", "String parameter value")
	cmd.Flags().String("json", "", "JSON parameter value")
	cmd.Flags().Bool("remove-all-conditional-values", false, "Remove all conditional values from matched parameters")
	cmd.Flags().StringArray("remove-conditional-value", nil, "Remove a conditional value from matched parameters; may be repeated")
	cmd.MarkFlagsMutuallyExclusive("boolean", "number", "string", "json")
	cmd.MarkFlagsMutuallyExclusive("group", "no-group")
	cmd.MarkFlagsMutuallyExclusive("remove-all-conditional-values", "remove-conditional-value")
	return cmd
}

func readRemoveConditionalValues(cmd *cobra.Command) ([]string, error) {
	values, err := cmd.Flags().GetStringArray("remove-conditional-value")
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("--remove-conditional-value cannot be empty")
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out, nil
}

func readValueSpec(cmd *cobra.Command) (*valueSpec, error) {
	value, err := shared.ReadValueFlag(cmd, false)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	return &valueSpec{value: value.Value, valueType: value.Type}, nil
}

func runUpdateRemote(cmd *cobra.Command, svc *core.Core, projectFilters []string, paramExpr string, paramFilters []string, search shared.ParameterSearch, spec updateSpec, yes bool, dryRun bool) error {
	ctx := context.Background()
	if dryRun {
		ctx = firebase.WithDryRun(ctx)
	}

	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return err
	}
	projects = shared.FilterProjects(projects, projectFilters)
	shared.SortProjects(projects)
	compiledExpr, ok := shared.CompileExpr(paramExpr, "")
	if !ok {
		return nil
	}

	var totals updateTotals
	for _, project := range projects {
		for {
			cfg, err := shared.RevalidateProjectConfig(ctx, svc, project)
			if err != nil {
				return err
			}
			matched := shared.CollectMatchingParamTargets(project, cfg.Config, paramFilters, search, compiledExpr, defaultGroupLabel)
			if len(matched) == 0 {
				break
			}
			var updatedCount int
			changedCount, retry, err := shared.PublishProjectConfigMutation(ctx, svc, cfg, "update", cmd.ErrOrStderr(), func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
				updated, finalCfg, err := confirmAndUpdateProject(cmd, project.ProjectID, current, matched, spec, yes, cmd.ErrOrStderr())
				if err != nil {
					return 0, nil, err
				}
				updatedCount = len(updated)
				return updatedCount, finalCfg, nil
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
			totals.updatedParams += updatedCount
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✏️ published: %s\n", project.ProjectID)
			break
		}
	}
	logUpdateTotals("remote", totals)
	return nil
}

func runUpdateStdin(cmd *cobra.Command, paramFilters []string, paramExpr string, search shared.ParameterSearch, spec updateSpec) error {
	cfg, _, err := shared.ReadRemoteConfigInput(cmd.InOrStdin())
	if err != nil {
		return err
	}
	compiledExpr, ok := shared.CompileExpr(paramExpr, "<stdin>")
	if !ok {
		return nil
	}

	project := core.Project{Name: "<stdin>", ProjectID: "<stdin>"}
	matched := shared.CollectMatchingParamTargets(project, cfg, paramFilters, search, compiledExpr, defaultGroupLabel)
	updated, finalCfg, err := confirmAndUpdateProject(cmd, "<stdin>", cfg, matched, spec, true, cmd.ErrOrStderr())
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(finalCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode remote config: %w", err)
	}
	if _, err := cmd.OutOrStdout().Write(out); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	totals := updateTotals{updatedParams: len(updated)}
	if len(updated) > 0 {
		totals.modifiedProjects = 1
	}
	logUpdateTotals("stdin", totals)
	return nil
}

func confirmAndUpdateProject(cmd *cobra.Command, label string, cfg *firebase.RemoteConfig, matched []shared.ParamTarget, spec updateSpec, yes bool, diffOut io.Writer) ([]shared.ParamTarget, *firebase.RemoteConfig, error) {
	finalCfg := shared.CloneRemoteConfig(cfg)
	updated := make([]shared.ParamTarget, 0, len(matched))

	for _, target := range matched {
		nextCfg := shared.CloneRemoteConfig(finalCfg)
		if err := updateParamSlot(nextCfg, target, spec); err != nil {
			return nil, nil, err
		}
		diffText, hasChanges := shared.RenderRemoteConfigDiff(finalCfg, nextCfg)
		if !hasChanges {
			continue
		}
		_, _ = fmt.Fprintln(diffOut, diffText)
		if !yes {
			ok, err := runConfirmationPrompt(
				fmt.Sprintf("Update %s in %s?", shared.FormatParameterHeader(target.Key, target.Group), label),
				cmd.OutOrStdout(),
			)
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				continue
			}
		}
		finalCfg = nextCfg
		updated = append(updated, target)
	}
	if len(updated) == 0 {
		return nil, finalCfg, nil
	}
	return updated, finalCfg, nil
}

func updateParamSlot(cfg *firebase.RemoteConfig, target shared.ParamTarget, spec updateSpec) error {
	param := target.Param
	if spec.value != nil {
		param.DefaultValue = &firebase.RemoteConfigValue{Value: spec.value.value}
		param.ValueType = spec.value.valueType
	}
	if spec.descriptionChanged {
		param.Description = spec.description
	}
	if spec.removeAllConditionalValues {
		param.ConditionalValues = nil
	} else if len(spec.removeConditionalValues) > 0 {
		for _, name := range spec.removeConditionalValues {
			delete(param.ConditionalValues, name)
		}
		if len(param.ConditionalValues) == 0 {
			param.ConditionalValues = nil
		}
	}

	nextGroup := target.Group
	if spec.groupChanged {
		nextGroup = spec.group
	}
	nextKey := target.Key
	if spec.nameChanged {
		nextKey = spec.name
	}
	if (target.Key != nextKey || target.Group != nextGroup) && shared.ParamSlotExists(cfg, nextKey, nextGroup) {
		return fmt.Errorf("parameter %s already exists", shared.FormatParameterHeader(nextKey, nextGroup))
	}
	shared.RemoveParamSlot(cfg, target.Key, target.Group)
	shared.SetParamSlot(cfg, nextKey, nextGroup, param)
	return nil
}

func runConfirmationPrompt(prompt string, fallbackOut io.Writer) (bool, error) {
	confirm := shared.NewConfirmation(prompt, confirmation.Yes, shared.ConfirmationOptions{})
	if fallbackOut != nil {
		confirm.Output = fallbackOut
	}
	return confirm.RunPrompt()
}

func logUpdateTotals(mode string, totals updateTotals) {
	corelog.For("update").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.updatedParams)
}
