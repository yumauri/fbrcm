package shared

import (
	"context"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
	"github.com/yumauri/fbrcm/ops/invocation"
)

// ResolveProjectTargetForExecution uses configured project resolution when
// local reads are allowed and otherwise requires one literal physical project
// target. Stateless targets default to the client template.
func ResolveProjectTargetForExecution(ctx context.Context, cmd invocation.Call, svc *core.Core, query string) (core.Project, error) {
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

// ResolvePhysicalProjectForExecution uses exact-then-filtered configured
// project resolution when local reads are allowed and otherwise requires one
// literal physical project ID without client/server target syntax.
func ResolvePhysicalProjectForExecution(ctx context.Context, cmd invocation.Call, svc *core.Core, query string) (core.Project, error) {
	if core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return ResolveProjectArg(ctx, cmd, svc, query)
	}
	if err := config.ValidatePhysicalProjectID(query); err != nil {
		return core.Project{}, InvalidArgument(err)
	}
	return core.Project{Name: query, ProjectID: query}, nil
}

// ResolveProjectScopedResourceForExecution resolves a physical project for a
// Firebase resource that is not attached to a Remote Config client/server
// template. Explicit target prefixes are rejected with resource-specific text.
func ResolveProjectScopedResourceForExecution(ctx context.Context, cmd invocation.Call, svc *core.Core, query, resource string) (core.Project, error) {
	target, explicit, err := rctarget.ParsePositionalSelector(query)
	if err != nil {
		return core.Project{}, InvalidArgument(err)
	}
	if explicit {
		return core.Project{}, InvalidArgument(fmt.Errorf("%s commands are project-scoped; omit the %s@ prefix", resource, target.Kind))
	}
	return ResolvePhysicalProjectForExecution(ctx, cmd, svc, target.ProjectID)
}

// ResolveCachedProjectScopedResource resolves a physical project exclusively
// from the local registry. It is used by commands whose explicit cache-only
// mode must not trigger project discovery.
func ResolveCachedProjectScopedResource(cmd invocation.Call, query, resource string) (core.Project, error) {
	target, explicit, err := rctarget.ParsePositionalSelector(query)
	if err != nil {
		return core.Project{}, InvalidArgument(err)
	}
	if explicit {
		return core.Project{}, InvalidArgument(fmt.Errorf("%s commands are project-scoped; omit the %s@ prefix", resource, target.Kind))
	}
	return ResolveCachedProjectArg(cmd, target.ProjectID)
}

// ResolveProjectNumberForExecution resolves the numeric project embedded in a
// Firebase App ID. Stateful execution maps it back to the configured project
// so the project's bound identity and quota project remain authoritative.
// Stateless execution can address the Firebase Management API by project
// number directly.
func ResolveProjectNumberForExecution(ctx context.Context, cmd invocation.Call, svc *core.Core, projectNumber string) (core.Project, error) {
	if !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return core.Project{Name: projectNumber, ProjectID: projectNumber, ProjectNumber: projectNumber}, nil
	}
	progress.Start("Resolving project…")
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectNumber(cmd, projects, projectNumber)
}

// ResolveCachedProjectNumber maps an App ID's embedded project number using
// only the local project registry.
func ResolveCachedProjectNumber(cmd invocation.Call, projectNumber string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, err := config.LoadProjects()
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectNumber(cmd, projects, projectNumber)
}

func resolveProjectNumber(cmd invocation.Call, projects []core.Project, projectNumber string) (core.Project, error) {
	matches := make([]core.Project, 0, 1)
	for _, project := range projects {
		if project.ProjectNumber == projectNumber {
			matches = append(matches, project)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		if len(projects) > 0 && !MachineMode(cmd) {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), RenderProjectsChoiceTable(projects)); err != nil {
				return core.Project{}, err
			}
		}
		return core.Project{}, &ProjectResolutionError{
			Resource:   "project",
			Kind:       "not_found",
			Query:      projectNumber,
			Candidates: selectionCandidates(projects),
			Err: fmt.Errorf(
				"project number %q from the Firebase App ID is not available in profile %q; pass --project or run projects update",
				projectNumber,
				config.GetActiveProfileName(),
			),
		}
	default:
		if !MachineMode(cmd) {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), RenderProjectsChoiceTable(matches)); err != nil {
				return core.Project{}, err
			}
		}
		return core.Project{}, &ProjectResolutionError{Resource: "project", Kind: "ambiguous", Query: projectNumber, Candidates: selectionCandidates(matches)}
	}
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
func ResolveProjectTargetsForExecution(ctx context.Context, cmd invocation.Call, svc *core.Core, rawFilters []string) ([]core.Project, context.Context, error) {
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

// ResolveListProjects preserves scalar positional resolution and otherwise uses
// bulk filters. Resource-scoped lists select physical projects; an empty
// resource selects Remote Config template targets. Cached mode never discovers.
func ResolveListProjects(cmd invocation.Call, svc *core.Core, args []string, resource string, cached bool) ([]core.Project, context.Context, error) {
	ctx := CommandContext(cmd)
	filters, err := cmd.Flags().GetStringArray("project")
	if err != nil {
		return nil, ctx, err
	}
	if len(args) > 0 && cmd.Flags().Changed("project") {
		return nil, ctx, InvalidArgument(fmt.Errorf("<project> and --project cannot be used together"))
	}
	var projects []core.Project
	if len(args) > 0 {
		var project core.Project
		if resource == "" {
			project, err = ResolveProjectTargetForExecution(ctx, cmd, svc, args[0])
		} else if cached {
			project, err = ResolveCachedProjectScopedResource(cmd, args[0], resource)
		} else {
			project, err = ResolveProjectScopedResourceForExecution(ctx, cmd, svc, args[0], resource)
		}
		if err == nil {
			projects = []core.Project{project}
		}
	} else if resource == "" {
		projects, ctx, err = ResolveProjectTargetsForExecution(ctx, cmd, svc, filters)
	} else {
		if err = RejectTemplateProjectFilters(filters); err != nil {
			return nil, ctx, err
		}
		if !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
			projects, ctx, err = ResolveProjectTargetsForExecution(ctx, cmd, svc, filters)
		} else {
			if cached {
				projects, err = config.LoadProjects()
			} else {
				projects, _, err = svc.ListProjects(ctx)
			}
			if err == nil {
				projects, err = FilterProjects(projects, filters)
			}
		}
	}
	if err != nil {
		return nil, ctx, err
	}
	strfold.SortProjects(projects, func(p core.Project) string { return p.Name }, func(p core.Project) string { return p.ProjectID })
	return projects, ctx, nil
}

// ResolveProjectMutationTargetsForExecution resolves target filters using the
// active execution policy and binds direct Firebase services for every
// stateless target. Stateful execution retains configured service resolution.
func ResolveProjectMutationTargetsForExecution(ctx context.Context, cmd invocation.Call, svc *core.Core, rawFilters []string) ([]core.Project, context.Context, error) {
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

func ResolveProjectArg(ctx context.Context, cmd invocation.Call, svc *core.Core, query string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectArg(cmd, projects, query)
}

// ResolveCachedProjectArg resolves a project using only the local projects
// registry. It never attempts project discovery.
func ResolveCachedProjectArg(cmd invocation.Call, query string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, err := config.LoadProjects()
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectArg(cmd, projects, query)
}

// ResolveProjectTargetArg resolves the project portion of a client/server
// template target, then returns the project with its canonical target identity.
func ResolveProjectTargetArg(ctx context.Context, cmd invocation.Call, svc *core.Core, query string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectTargetArg(cmd, projects, query)
}

// ResolveCachedProjectTargetArg is the cache-only counterpart of
// ResolveProjectTargetArg.
func ResolveCachedProjectTargetArg(cmd invocation.Call, query string) (core.Project, error) {
	progress.Start("Resolving project…")
	projects, err := config.LoadProjects()
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectTargetArg(cmd, projects, query)
}

func resolveProjectTargetArg(cmd invocation.Call, projects []core.Project, query string) (core.Project, error) {
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

func resolveProjectArg(cmd invocation.Call, projects []core.Project, query string) (core.Project, error) {
	aliases, err := config.LoadProjectAliases()
	if err != nil {
		return core.Project{}, err
	}
	return resolveProjectArgWithAliases(cmd, projects, query, aliases)
}

func resolveProjectArgWithAliases(cmd invocation.Call, projects []core.Project, query string, aliases map[string]string) (core.Project, error) {
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
	if len(matches) == 0 {
		mode, filterQuery := filter.ParseModePrefixedQuery(query)
		if strings.TrimSpace(filterQuery) == "" {
			return core.Project{}, InvalidArgument(fmt.Errorf("project selector requires a non-empty query"))
		}
		matches = filterProjectsWithAliases(projects, []QueryFilter{{Mode: mode, Query: filterQuery}}, nil)
	}

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
