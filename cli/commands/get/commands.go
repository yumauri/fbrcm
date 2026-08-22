package get

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/commands/get/table"
	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	corelog "github.com/yumauri/fbrcm/core/log"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/core/strfold"
)

type getOptions struct {
	projectFilters []string
	projectExpr    string
	paramFilters   []string
	paramArgument  *string
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
	contract.RegisterResponse(getCmd, []parameterRowJSON{})
	return getCmd
}

func addGetFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Print parameters as JSON")
	shared.AddProjectTargetFilterFlag(cmd)
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
	stdinAvailable := shared.StdinAvailable(cmd.InOrStdin())
	if opts.update && !core.ExecutionPolicyFromContext(shared.CommandContext(cmd)).ReadLocalState {
		return shared.InvalidArgument(fmt.Errorf("--update cannot be used with --stateless; Remote Config reads are already live"))
	}
	if opts.update && stdinAvailable {
		return shared.InvalidArgument(fmt.Errorf("--update cannot be used with stdin; stdin Remote Config is already the complete input"))
	}
	if stdinAvailable {
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
	paramArgument, err := shared.ResolveParameterArgument(args, paramFilters)
	if err != nil {
		return getOptions{}, err
	}
	return getOptions{
		projectFilters: projectFilters,
		projectExpr:    projectExpr,
		paramFilters:   paramFilters,
		paramArgument:  paramArgument,
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
		return printGetRows(cmd, "table-stdin-dir", rows, opts.paramFilters, opts.paramArgument, opts.jsonOut, false, true)
	}
	corelog.For("get").Info("stdin mode enabled; using remote config from stdin")
	compiledExpr, err := shared.CompileExpr(opts.projectExpr, "<stdin>")
	if err != nil {
		return err
	}
	_, rows, err := loadStdinParameterRows(cmd, compiledExpr, opts.search)
	if err != nil {
		return err
	}
	rows = filterParameterRowsByProject(rows, opts.projectFilters)
	return printGetRows(cmd, "table-stdin", rows, opts.paramFilters, opts.paramArgument, opts.jsonOut, false, false)
}

func runGetRemote(cmd *cobra.Command, svc *core.Core, opts getOptions) error {
	ctx := shared.CommandContext(cmd)
	projects, ctx, err := resolveGetProjectsForExecution(ctx, cmd, svc, opts.projectFilters)
	if err != nil {
		return err
	}
	cmd.SetContext(ctx)
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })

	if opts.update {
		progress.Start("Revalidating Remote Config…")
	} else {
		progress.Start("Loading Remote Config…")
	}
	compiledExpr, err := shared.CompileExpr(opts.projectExpr, "")
	if err != nil {
		return err
	}
	loaded, err := loadProjectsParameters(ctx, svc, projects, opts.update)
	if err != nil {
		return err
	}

	rows := make([]parameterRow, 0)
	for _, item := range loaded {
		if item.status == "stale" {
			selector, _ := rctarget.ExactFilter(item.project.ProjectID)
			shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "cache.stale", Message: "The command used a stale local Remote Config cache after refresh failed.", Target: item.project.ProjectID, Details: struct {
				Source string `json:"source"`
			}{Source: item.source}, Remediation: []shared.Remediation{{Description: "refresh the stale target", Strategy: shared.RemediationRunCommand, Argv: []string{"get", "--update", "--project", selector}}}})
		}
		if item.cfg == nil || item.cache == nil {
			continue
		}
		cachedAt := item.cache.CachedAt
		if !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
			cachedAt = time.Time{}
		}
		projectRows, err := flattenParameters(item.project, item.cfg, cachedAt, item.status, "", compiledExpr, opts.search)
		if err != nil {
			return err
		}
		rows = append(rows, projectRows...)
	}

	rows = filterParameterRows(rows, opts.paramFilters, opts.paramArgument)
	sortParameterRows(rows)
	if opts.jsonOut {
		if err := writeRowsJSON(cmd, rows); err != nil {
			return err
		}
		logGetTotals("json", rows)
		return nil
	}

	projectExact := singleExactProjectFilter(opts.projectFilters) && len(loaded) == 1
	paramExact := opts.paramArgument != nil || singleExactParameterFilter(opts.paramFilters)
	tableRows := rows
	if opts.all {
		tableRows = buildTableRows(loaded, rows)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), table.Render(tableRows, shared.ParseFilters(opts.paramFilters), paramExact, !projectExact))
	logGetTotals("table", tableRows)
	return nil
}

func resolveGetProjectsForExecution(ctx context.Context, cmd *cobra.Command, svc *core.Core, projectFilters []string) ([]core.Project, context.Context, error) {
	return shared.ResolveProjectTargetsForExecution(ctx, cmd, svc, projectFilters)
}

func printGetRows(cmd *cobra.Command, source string, rows []parameterRow, paramFilters []string, paramArgument *string, jsonOut bool, allowHideKey, includeProject bool) error {
	rows = filterParameterRows(rows, paramFilters, paramArgument)
	sortParameterRows(rows)
	if jsonOut {
		return writeRowsJSON(cmd, rows)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), table.Render(rows, shared.ParseFilters(paramFilters), allowHideKey, includeProject))
	logGetTotals(source, rows)
	return nil
}
