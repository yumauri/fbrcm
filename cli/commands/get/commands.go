package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
)

const defaultGroupLabel = "(root)"

func New(svc *core.Core) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get [parameter]",
		Short: "Get parameters from all projects",
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
			paramFilters, err := cmd.Flags().GetStringArray("filter")
			if err != nil {
				return err
			}
			searchValue, err := cmd.Flags().GetString("search")
			if err != nil {
				return err
			}
			search := shared.NewParameterSearch(searchValue)
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			update, err := cmd.Flags().GetBool("update")
			if err != nil {
				return err
			}
			all, err := cmd.Flags().GetBool("all")
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
				if handled, rows, err := loadStdinDirectoryParameterRows(cmd, projectExpr, search); handled || err != nil {
					if err != nil {
						return err
					}
					rows = filterParameterRows(rows, paramFilters)
					sortParameterRows(rows)

					if jsonOut {
						return writeRowsJSON(cmd, rows)
					}

					_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderParametersTable(rows, shared.ParseFilters(paramFilters), false, true))
					logGetTotals("table-stdin-dir", rows)
					return nil
				}
				corelog.For("get").Info("stdin mode enabled; using remote config from stdin")
				compiledExpr, ok := shared.CompileExpr(projectExpr, "<stdin>")
				if !ok {
					return nil
				}
				_, rows, err := loadStdinParameterRows(cmd, compiledExpr, search)
				if err != nil {
					return err
				}
				rows = filterParameterRows(rows, paramFilters)
				sortParameterRows(rows)

				if jsonOut {
					return writeRowsJSON(cmd, rows)
				}

				_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderParametersTable(rows, shared.ParseFilters(paramFilters), false, false))
				logGetTotals("table-stdin", rows)
				return nil
			}

			projects, _, err := svc.ListProjects(context.Background())
			if err != nil {
				return err
			}
			projects = shared.FilterProjects(projects, projectFilters)
			shared.SortProjects(projects)

			loaded, err := loadProjectsParameters(context.Background(), svc, projects, update)
			if err != nil {
				return err
			}
			compiledExpr, ok := shared.CompileExpr(projectExpr, "")
			if !ok {
				return nil
			}

			rows := make([]parameterRow, 0)
			for _, item := range loaded {
				if item.cfg == nil || item.cache == nil {
					continue
				}
				rows = append(rows, flattenParameters(item.project, item.cfg, item.cache.CachedAt, item.status, "", compiledExpr, search)...)
			}

			rows = filterParameterRows(rows, paramFilters)
			sortParameterRows(rows)

			if jsonOut {
				if err := writeRowsJSON(cmd, rows); err != nil {
					return err
				}
				logGetTotals("json", rows)
				return nil
			}

			projectExact := singleExactProjectFilter(projectFilters)
			paramExact := singleExactParameterFilter(paramFilters)
			tableRows := rows
			if all {
				tableRows = buildTableRows(loaded, rows)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderParametersTable(tableRows, shared.ParseFilters(paramFilters), paramExact, !projectExact))
			logGetTotals("table", tableRows)
			return nil
		},
	}

	getCmd.Flags().Bool("json", false, "Print parameters as JSON")
	getCmd.Flags().StringArrayP("project", "p", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	getCmd.Flags().StringArrayP("filter", "f", nil, "Filter parameters by mode-prefixed query (^, /, ~, =); may be repeated")
	getCmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	getCmd.Flags().String("search", "", "Search parameters by name, description, values, and conditions")
	getCmd.Flags().Bool("all", false, "Include projects with no matching parameters")
	getCmd.Flags().Bool("update", false, "Revalidate cached parameters before printing")
	return getCmd
}
