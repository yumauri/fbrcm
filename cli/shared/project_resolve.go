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
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

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
