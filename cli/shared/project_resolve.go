package shared

import (
	"context"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
)

func ResolveProjectArg(ctx context.Context, cmd *cobra.Command, svc *core.Core, query string) (core.Project, error) {
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return core.Project{}, err
	}

	for _, project := range projects {
		if strings.EqualFold(project.ProjectID, query) {
			return project, nil
		}
	}

	matches := make([]core.Project, 0, 1)
	for _, project := range projects {
		if strings.EqualFold(project.Name, query) {
			matches = append(matches, project)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		if len(projects) > 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), RenderProjectsChoiceTable(projects))
		}
		return core.Project{}, fmt.Errorf("no project matches %q", query)
	default:
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), RenderProjectsChoiceTable(matches))
		return core.Project{}, fmt.Errorf("several projects match %q", query)
	}
}

func RenderProjectsChoiceTable(projects []core.Project) string {
	rows := make([][]string, 0, len(projects))
	projectWidth := lipgloss.Width("Project")
	idWidth := lipgloss.Width("Project ID")
	for _, project := range projects {
		rows = append(rows, []string{project.Name, project.ProjectID})
		projectWidth = max(projectWidth, lipgloss.Width(project.Name))
		idWidth = max(idWidth, lipgloss.Width(project.ProjectID))
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
		Headers("Project", "Project ID").
		Rows(rows...).
		Width(projectWidth + idWidth + 7).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}
