package deletecmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type deleteTotals struct {
	modifiedProjects int
	deletedParams    int
}

const defaultDeleteGroupLabel = "(root)"

// New constructs the delete command.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [parameter]",
		Short: "Delete Remote Config parameters",
		Args:  cobra.MaximumNArgs(1),
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
			if len(args) > 0 {
				var err error
				paramFilters, err = shared.ResolveParameterArgFilters(args, paramFilters)
				if err != nil {
					return err
				}
			}

			if shared.StdinAvailable(cmd.InOrStdin()) {
				corelog.For("delete").Info("stdin mode enabled; using remote config from stdin")
				return runDeleteStdin(cmd, paramFilters, projectExpr, search)
			}
			return runDeleteRemote(cmd, svc, projectFilters, projectExpr, paramFilters, search, yes, dryRun)
		},
	}

	cmd.Flags().StringArrayP("project", "p", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter parameters by mode-prefixed query (^, /, ~, =); may be repeated")
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	cmd.Flags().String("search", "", "Search parameters by name, description, values, and conditions")
	cmd.Flags().Bool("dry-run", false, "Log Firebase write requests without sending them")
	cmd.Flags().BoolP("yes", "y", false, "Print diff and delete without confirmation")
	return cmd
}

func runDeleteRemote(cmd *cobra.Command, svc *core.Core, projectFilters []string, projectExpr string, paramFilters []string, search shared.ParameterSearch, yes bool, dryRun bool) error {
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
	compiledExpr, ok := shared.CompileExpr(projectExpr, "")
	if !ok {
		return nil
	}

	var totals deleteTotals
	for _, project := range projects {
		for {
			cfg, err := shared.RevalidateProjectConfig(ctx, svc, project)
			if err != nil {
				return err
			}

			matched := shared.CollectMatchingParamTargets(project, cfg.Config, paramFilters, search, compiledExpr, defaultDeleteGroupLabel)
			if len(matched) == 0 {
				break
			}

			var deletedCount int
			changedCount, retry, err := shared.PublishProjectConfigMutation(ctx, svc, cfg, "delete", cmd.ErrOrStderr(), func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
				deleted, finalCfg, err := confirmAndDeleteProject(cmd, project.ProjectID, current, matched, yes, cmd.ErrOrStderr())
				if err != nil {
					return 0, nil, err
				}
				deletedCount = len(deleted)
				return deletedCount, finalCfg, nil
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
			totals.deletedParams += deletedCount
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🗑️ published: %s\n", project.ProjectID)
			break
		}
	}

	logDeleteTotals("remote", totals)
	return nil
}

func runDeleteStdin(cmd *cobra.Command, paramFilters []string, projectExpr string, search shared.ParameterSearch) error {
	cfg, remoteConfigRaw, err := shared.ReadRemoteConfigInput(cmd.InOrStdin())
	if err != nil {
		return err
	}
	compiledExpr, ok := shared.CompileExpr(projectExpr, "<stdin>")
	if !ok {
		return nil
	}

	order, err := shared.ParseRemoteConfigOrder(remoteConfigRaw)
	if err != nil {
		return fmt.Errorf("parse stdin remote config order: %w", err)
	}

	project := core.Project{Name: "<stdin>", ProjectID: "<stdin>"}
	matched := shared.CollectMatchingParamTargets(project, cfg, paramFilters, search, compiledExpr, defaultDeleteGroupLabel)
	deleted, finalCfg, err := confirmAndDeleteProject(cmd, "<stdin>", cfg, matched, true, cmd.ErrOrStderr())
	if err != nil {
		return err
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

	totals := deleteTotals{deletedParams: len(deleted)}
	if len(deleted) > 0 {
		totals.modifiedProjects = 1
	}
	logDeleteTotals("stdin", totals)
	return nil
}

func confirmAndDeleteProject(cmd *cobra.Command, label string, cfg *firebase.RemoteConfig, matched []shared.ParamTarget, yes bool, diffOut io.Writer) ([]shared.ParamTarget, *firebase.RemoteConfig, error) {
	finalCfg := shared.CloneRemoteConfig(cfg)
	deleted := make([]shared.ParamTarget, 0, len(matched))

	for _, target := range matched {
		diffText := renderDeletedParameter(target)
		_, _ = fmt.Fprintln(diffOut, diffText)

		if !yes {
			ok, err := runConfirmationPrompt(
				fmt.Sprintf("Delete %s from %s?", shared.FormatParameterHeader(target.Key, target.Group), label),
				cmd.OutOrStdout(),
			)
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				continue
			}
		}

		shared.RemoveParamSlot(finalCfg, target.Key, target.Group)
		deleted = append(deleted, target)
	}

	if len(deleted) == 0 {
		return nil, finalCfg, nil
	}

	return deleted, finalCfg, nil
}

func runConfirmationPrompt(prompt string, fallbackOut io.Writer) (bool, error) {
	confirm := shared.NewConfirmation(prompt, confirmation.Yes, shared.ConfirmationOptions{
		Destructive: true,
	})
	if fallbackOut != nil {
		confirm.Output = fallbackOut
	}
	return confirm.RunPrompt()
}

func renderDeletedParameter(target shared.ParamTarget) string {
	lines := []string{fmt.Sprintf("  - %s", colorRemoved(shared.FormatParameterHeader(target.Key, target.Group)))}
	if strings.TrimSpace(target.Param.ValueType) != "" {
		lines = append(lines, fmt.Sprintf("      - type:                %s", colorRemoved(target.Param.ValueType)))
	}
	if strings.TrimSpace(target.Param.Description) != "" {
		lines = append(lines, fmt.Sprintf("      - description:         %s", colorRemoved(shared.FormatPlainValue(target.Param.Description))))
	}
	if target.Param.DefaultValue != nil {
		lines = append(lines, fmt.Sprintf("      - default:             %s", colorRemoved(shared.FormatRemoteValue(*target.Param.DefaultValue))))
	}
	for _, condition := range shared.SortedStringKeys(target.Param.ConditionalValues) {
		lines = append(lines, fmt.Sprintf("      - cond %-15s %s", condition+":", colorRemoved(shared.FormatRemoteValue(target.Param.ConditionalValues[condition]))))
	}
	return "\n" + strings.Join(lines, "\n")
}

func colorRemoved(value string) string {
	if value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(value)
}

func logDeleteTotals(mode string, totals deleteTotals) {
	corelog.For("delete").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.deletedParams)
}
