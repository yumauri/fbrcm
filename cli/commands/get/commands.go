package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/commands/get/table"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
)

type getOptions struct {
	projectFilters []string
	projectExpr    string
	paramFilters   []string
	search         shared.ParameterSearch
	jsonOut        bool
	update         bool
	all            bool
}

func New(svc *core.Core) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get [parameter]",
		Short: "Get parameters from all projects",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetCommand(cmd, svc, args)
		},
	}

	addGetFlags(getCmd)
	return getCmd
}

func addGetFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Print parameters as JSON")
	shared.AddProjectFilterFlag(cmd)
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	cmd.Flags().Bool("all", false, "Include projects with no matching parameters")
	cmd.Flags().Bool("update", false, "Revalidate cached parameters before printing")
}

func runGetCommand(cmd *cobra.Command, svc *core.Core, args []string) error {
	opts, err := readGetOptions(cmd, args)
	if err != nil {
		return err
	}
	if shared.StdinAvailable(cmd.InOrStdin()) {
		return runGetStdin(cmd, opts)
	}
	return runGetRemote(cmd, svc, opts)
}

func readGetOptions(cmd *cobra.Command, args []string) (getOptions, error) {
	projectFilters, err := cmd.Flags().GetStringArray("project")
	if err != nil {
		return getOptions{}, err
	}
	projectExpr, err := cmd.Flags().GetString("expr")
	if err != nil {
		return getOptions{}, err
	}
	paramFilters, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return getOptions{}, err
	}
	searchValue, err := cmd.Flags().GetString("search")
	if err != nil {
		return getOptions{}, err
	}
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return getOptions{}, err
	}
	update, err := cmd.Flags().GetBool("update")
	if err != nil {
		return getOptions{}, err
	}
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return getOptions{}, err
	}
	if len(args) > 0 {
		paramFilters, err = shared.ResolveParameterArgFilters(args, paramFilters)
		if err != nil {
			return getOptions{}, err
		}
	}
	return getOptions{
		projectFilters: projectFilters,
		projectExpr:    projectExpr,
		paramFilters:   paramFilters,
		search:         shared.NewParameterSearch(searchValue),
		jsonOut:        jsonOut,
		update:         update,
		all:            all,
	}, nil
}

func runGetStdin(cmd *cobra.Command, opts getOptions) error {
	if handled, rows, err := loadStdinDirectoryParameterRows(cmd, opts.projectFilters, opts.projectExpr, opts.search); handled || err != nil {
		if err != nil {
			return err
		}
		return printGetRows(cmd, "table-stdin-dir", rows, opts.paramFilters, opts.jsonOut, false, true)
	}
	corelog.For("get").Info("stdin mode enabled; using remote config from stdin")
	compiledExpr, ok := shared.CompileExpr(opts.projectExpr, "<stdin>")
	if !ok {
		return nil
	}
	_, rows, err := loadStdinParameterRows(cmd, compiledExpr, opts.search)
	if err != nil {
		return err
	}
	rows = filterParameterRowsByProject(rows, opts.projectFilters)
	return printGetRows(cmd, "table-stdin", rows, opts.paramFilters, opts.jsonOut, false, false)
}

func runGetRemote(cmd *cobra.Command, svc *core.Core, opts getOptions) error {
	projects, _, err := svc.ListProjects(context.Background())
	if err != nil {
		return err
	}
	projects = shared.FilterProjects(projects, opts.projectFilters)
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })

	loaded, err := loadProjectsParameters(context.Background(), svc, projects, opts.update)
	if err != nil {
		return err
	}
	compiledExpr, ok := shared.CompileExpr(opts.projectExpr, "")
	if !ok {
		return nil
	}

	rows := make([]parameterRow, 0)
	for _, item := range loaded {
		if item.cfg == nil || item.cache == nil {
			continue
		}
		rows = append(rows, flattenParameters(item.project, item.cfg, item.cache.CachedAt, item.status, "", compiledExpr, opts.search)...)
	}

	rows = filterParameterRows(rows, opts.paramFilters)
	sortParameterRows(rows)
	if opts.jsonOut {
		if err := writeRowsJSON(cmd, rows); err != nil {
			return err
		}
		logGetTotals("json", rows)
		return nil
	}

	projectExact := singleExactProjectFilter(opts.projectFilters)
	paramExact := singleExactParameterFilter(opts.paramFilters)
	tableRows := rows
	if opts.all {
		tableRows = buildTableRows(loaded, rows)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), table.Render(tableRows, shared.ParseFilters(opts.paramFilters), paramExact, !projectExact))
	logGetTotals("table", tableRows)
	return nil
}

func printGetRows(cmd *cobra.Command, source string, rows []parameterRow, paramFilters []string, jsonOut bool, allowHideKey, includeProject bool) error {
	rows = filterParameterRows(rows, paramFilters)
	sortParameterRows(rows)
	if jsonOut {
		return writeRowsJSON(cmd, rows)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), table.Render(rows, shared.ParseFilters(paramFilters), allowHideKey, includeProject))
	logGetTotals(source, rows)
	return nil
}
