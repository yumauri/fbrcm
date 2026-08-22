package shared

import (
	"context"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/progress"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

// ResolveProjectTargetForExecution uses configured project resolution when
// local reads are allowed and otherwise requires one literal physical project
// target. Stateless targets default to the client template.
func ResolveProjectTargetForExecution(ctx context.Context, cmd *cobra.Command, svc *core.Core, query string) (core.Project, error) {
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return ResolveProjectTargetArg(ctx, cmd, svc, query)
	}
	target, _, err := rctarget.ParsePositionalSelector(query)
	if err != nil {
		return core.Project{}, InvalidArgument(err)
	}
	if err := config.ValidatePhysicalProjectID(target.ProjectID); err != nil {
		return core.Project{}, InvalidArgument(err)
	}
	return core.Project{
		Name:            target.ProjectID,
		ProjectID:       target.String(),
		Templates:       []rctarget.Kind{target.Kind},
		PrimaryTemplate: target.Kind,
	}, nil
}

// ResolvePhysicalProjectForExecution uses configured project resolution when
// local reads are allowed and otherwise requires one literal physical project
// ID without client/server target syntax.
func ResolvePhysicalProjectForExecution(ctx context.Context, cmd *cobra.Command, svc *core.Core, query string) (core.Project, error) {
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return ResolveProjectArg(ctx, cmd, svc, query)
	}
	if err := config.ValidatePhysicalProjectID(query); err != nil {
		return core.Project{}, InvalidArgument(err)
	}
	return core.Project{Name: query, ProjectID: query}, nil
}

// FirebaseServiceContextForExecution binds an in-memory static-token Firebase
// service when configured service resolution is disabled by the execution
// policy. Stateful execution returns the original context unchanged.
func FirebaseServiceContextForExecution(ctx context.Context, projectID string) (context.Context, error) {
	return FirebaseServicesContextForExecution(ctx, []string{projectID})
}

// FirebaseServicesContextForExecution binds one in-memory static-token
// Firebase service to every selected physical project when configured service
// resolution is disabled by the execution policy. Stateful execution returns
// the original context unchanged.
func FirebaseServicesContextForExecution(ctx context.Context, projectIDs []string) (context.Context, error) {
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return ctx, nil
	}
	service, err := staticAccessTokenFirebaseService(ctx)
	if err != nil {
		return nil, err
	}
	for _, projectID := range projectIDs {
		ctx, err = core.WithDirectFirebaseService(ctx, projectID, service)
		if err != nil {
			return nil, err
		}
	}
	return ctx, nil
}

// FirebaseProjectDiscoveryContextForExecution binds an in-memory static-token
// service for stateless project discovery. Stateful execution returns the
// original context unchanged.
func FirebaseProjectDiscoveryContextForExecution(ctx context.Context) (context.Context, error) {
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return ctx, nil
	}
	service, err := staticAccessTokenFirebaseService(ctx)
	if err != nil {
		return nil, err
	}
	return core.WithDirectFirebaseDiscoveryService(ctx, service)
}

// ResolveProjectTargetsForExecution applies normal profile-backed target
// selection when local state is enabled. In stateless execution, exact
// selectors are treated as literal project IDs while all other selectors are
// matched against one live project-discovery result without repository aliases.
func ResolveProjectTargetsForExecution(ctx context.Context, cmd *cobra.Command, svc *core.Core, rawFilters []string) ([]core.Project, context.Context, error) {
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		progress.Start("Loading projects…")
		projects, _, err := svc.ListProjects(ctx)
		if err != nil {
			return nil, ctx, err
		}
		projects, err = FilterProjectTargets(projects, rawFilters)
		return projects, ctx, err
	}

	selected := make([]core.Project, 0, len(rawFilters))
	seen := make(map[string]struct{})
	appendUnique := func(projects ...core.Project) {
		for _, project := range projects {
			if _, ok := seen[project.ProjectID]; ok {
				continue
			}
			seen[project.ProjectID] = struct{}{}
			selected = append(selected, project)
		}
	}

	discoveryFilters := make([]string, 0, len(rawFilters))
	hasFilter := false
	for _, raw := range rawFilters {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		hasFilter = true
		target, _, err := rctarget.ParseSelector(raw)
		if err != nil {
			return nil, ctx, InvalidArgument(err)
		}
		mode, query := filter.ParseModePrefixedQuery(target.ProjectID)
		if mode != filter.ModeExact {
			discoveryFilters = append(discoveryFilters, raw)
			continue
		}

		target.ProjectID = query
		project, err := ResolveProjectTargetForExecution(ctx, cmd, svc, target.String())
		if err != nil {
			return nil, ctx, err
		}
		appendUnique(project)
	}

	if !hasFilter || len(discoveryFilters) > 0 {
		discoveryCtx, err := FirebaseProjectDiscoveryContextForExecution(ctx)
		if err != nil {
			return nil, ctx, err
		}
		progress.Start("Loading projects…")
		projects, _, err := svc.ListProjectsForExecution(discoveryCtx)
		if err != nil {
			return nil, ctx, err
		}
		projects, err = FilterProjectTargetsWithAliases(projects, discoveryFilters, nil)
		if err != nil {
			return nil, ctx, err
		}
		appendUnique(projects...)
		ctx = discoveryCtx
	}

	return selected, ctx, nil
}

// ResolveProjectMutationTargetsForExecution resolves target filters using the
// active execution policy and binds direct Firebase services for every
// stateless target. Stateful execution retains configured service resolution.
func ResolveProjectMutationTargetsForExecution(ctx context.Context, cmd *cobra.Command, svc *core.Core, rawFilters []string) ([]core.Project, context.Context, error) {
	projects, ctx, err := ResolveProjectTargetsForExecution(ctx, cmd, svc, rawFilters)
	if err != nil {
		return nil, ctx, err
	}
	projectIDs := make([]string, len(projects))
	for i, project := range projects {
		projectIDs[i] = project.ProjectID
	}
	ctx, err = FirebaseServicesContextForExecution(ctx, projectIDs)
	if err != nil {
		return nil, ctx, err
	}
	return projects, ctx, nil
}

func staticAccessTokenFirebaseService(ctx context.Context) (*firebase.Service, error) {
	accessToken, ok := env.LookupNonEmpty(env.GoogleAccessToken)
	if !ok {
		return nil, &core.AuthError{
			Kind: "configuration",
			Err:  fmt.Errorf("%s is required with --stateless", env.GoogleAccessToken),
		}
	}
	service, err := firebase.NewServiceWithAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func ResolveProjectArg(ctx context.Context, cmd *cobra.Command, svc *core.Core, query string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectArg(cmd, projects, query)
}

// ResolveCachedProjectArg resolves a project using only the local projects
// registry. It never attempts project discovery.
func ResolveCachedProjectArg(cmd *cobra.Command, query string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, err := config.LoadProjects()
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectArg(cmd, projects, query)
}

// ResolveProjectTargetArg resolves the project portion of a client/server
// template target, then returns the project with its canonical target identity.
func ResolveProjectTargetArg(ctx context.Context, cmd *cobra.Command, svc *core.Core, query string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectTargetArg(cmd, projects, query)
}

// ResolveCachedProjectTargetArg is the cache-only counterpart of
// ResolveProjectTargetArg.
func ResolveCachedProjectTargetArg(cmd *cobra.Command, query string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, err := config.LoadProjects()
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectTargetArg(cmd, projects, query)
}

func resolveProjectTargetArg(cmd *cobra.Command, projects []core.Project, query string) (core.Project, error) {
	target, explicit, err := rctarget.ParsePositionalSelector(query)
	if err != nil {
		return core.Project{}, err
	}
	project, err := resolveProjectArg(cmd, projects, target.ProjectID)
	if err != nil {
		return core.Project{}, err
	}
	if !explicit {
		target.Kind = project.TemplateKinds()[0]
	}
	project.ProjectID = target.WithProjectID(project.ProjectID).String()
	return project, nil
}

func resolveProjectArg(cmd *cobra.Command, projects []core.Project, query string) (core.Project, error) {
	aliases, err := config.LoadProjectAliases()
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectArgWithAliases(cmd, projects, query, aliases)
}

func resolveProjectArgWithAliases(cmd *cobra.Command, projects []core.Project, query string, aliases map[string]string) (core.Project, error) {
	for _, project := range projects {
		if project.ProjectID == query {
			return project, nil
		}
	}
	if alias, projectID, ok := config.ResolveProjectAlias(aliases, query); ok {
		for _, project := range projects {
			if project.ProjectID == projectID {
				return project, nil
			}
		}
		return core.Project{}, &ProjectResolutionError{
			Resource:   "project",
			Kind:       "not_found",
			Query:      query,
			Candidates: selectionCandidates(projects),
			Err: fmt.Errorf(
				"project alias %q resolves to %q, but that project is not available in profile %q",
				alias,
				projectID,
				config.GetActiveProfileName(),
			),
		}
	}
	matches := matchProjectsForArg(projects, query)

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		if len(projects) > 0 && !MachineMode(cmd) {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), RenderProjectsChoiceTable(projects)); err != nil {
				return core.Project{}, err
			}
		}
		return core.Project{}, &ProjectResolutionError{Resource: "project", Kind: "not_found", Query: query, Candidates: selectionCandidates(projects)}
	default:
		if !MachineMode(cmd) {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), RenderProjectsChoiceTable(matches)); err != nil {
				return core.Project{}, err
			}
		}
		return core.Project{}, &ProjectResolutionError{Resource: "project", Kind: "ambiguous", Query: query, Candidates: selectionCandidates(matches)}
	}
}

func selectionCandidates(projects []core.Project) []SelectionCandidate {
	result := make([]SelectionCandidate, 0, len(projects))
	for _, project := range projects {
		result = append(result, SelectionCandidate{Name: project.Name, ID: project.ProjectID})
	}
	return result
}

func matchProjectsForArg(projects []core.Project, query string) []core.Project {
	if query == "" {
		return nil
	}
	for _, project := range projects {
		if project.ProjectID == query {
			return []core.Project{project}
		}
	}
	exactNames := make([]core.Project, 0, 1)
	for _, project := range projects {
		if project.Name == query {
			exactNames = append(exactNames, project)
		}
	}
	return exactNames
}

func RenderProjectsChoiceTable(projects []core.Project) string {
	return renderProjectsChoiceTableAtWidth(projects, TerminalWidth())
}

func renderProjectsChoiceTableAtWidth(projects []core.Project, terminalWidth int) string {
	aliases, _ := config.LoadProjectAliases()
	aliasesByID := config.ProjectAliasesByID(aliases)
	headers := []string{"Project", "Project ID", "Aliases"}
	rows := make([][]string, 0, len(projects))
	widths := []int{lipgloss.Width(headers[0]), lipgloss.Width(headers[1]), lipgloss.Width(headers[2])}
	for _, project := range projects {
		aliasLabel := strings.Join(aliasesByID[project.ProjectID], ", ")
		if aliasLabel == "" {
			aliasLabel = "—"
		}
		rows = append(rows, []string{project.Name, project.ProjectID, aliasLabel})
		for column, value := range rows[len(rows)-1] {
			widths[column] = max(widths[column], lipgloss.Width(value))
		}
	}
	tableWidth := func() int { return widths[0] + widths[1] + widths[2] + 10 }
	if terminalWidth > 0 && tableWidth() > terminalWidth {
		for _, minimumMode := range []bool{true, false} {
			for _, column := range []int{2, 0, 1} {
				minimum := 1
				if minimumMode {
					minimum = lipgloss.Width(headers[column])
				}
				for widths[column] > minimum && tableWidth() > terminalWidth {
					widths[column]--
				}
			}
		}
		for column := range headers {
			headers[column] = ansi.Truncate(headers[column], widths[column], "…")
		}
		for row := range rows {
			for column := range rows[row] {
				rows[row][column] = ansi.Truncate(rows[row][column], widths[column], "…")
			}
		}
	}

	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if clistyles.NoColorEnabled() {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if col == 0 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateDim)
	}

	tbl := table.New().
		Headers(headers...).
		Rows(rows...).
		Width(tableWidth()).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}
