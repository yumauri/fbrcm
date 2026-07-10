package projects

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func printProjects(cmd *cobra.Command, svc *core.Core, projects []core.Project, source string) error {
	_ = source
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}

	filterValues, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return err
	}
	projectExpr, err := cmd.Flags().GetString("expr")
	if err != nil {
		return err
	}
	withURL, err := cmd.Flags().GetBool("url")
	if err != nil {
		return err
	}

	projects = shared.FilterProjects(projects, filterValues)
	projects, err = shared.FilterProjectsByExpr(context.Background(), svc, projects, projectExpr)
	if err != nil {
		return err
	}
	highlightFilters := shared.ParseFilters(filterValues)

	if jsonOut {
		if err := shared.WriteJSON(cmd, projectsJSON(projects, withURL)); err != nil {
			return fmt.Errorf("encode projects json: %w", err)
		}
		logProjectsTotal(projects)
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderProjectsTable(projects, highlightFilters, withURL))
	logProjectsTotal(projects)
	return nil
}

func logProjectsTotal(projects []core.Project) {
	corelog.For("projects").Info("total", "projects", len(projects))
}

func renderProjectsTable(projects []core.Project, highlightFilters []shared.QueryFilter, withURL bool) string {
	noColor := clistyles.NoColorEnabled()
	rows := make([][]string, 0, len(projects))
	projectWidth := lipgloss.Width("Project")
	idWidth := lipgloss.Width("Project ID")
	numberWidth := lipgloss.Width("Number")
	authWidth := lipgloss.Width("Auth")
	updatedAtWidth := lipgloss.Width("Updated At")
	syncedAtWidth := lipgloss.Width("Synced At")
	linkWidth := lipgloss.Width("URL")
	for _, project := range projects {
		rowIndex := len(rows)
		var rowBG color.Color
		if !noColor && rowIndex >= 0 && rowIndex%2 == 1 {
			rowBG = clistyles.ColorRowStripe
		}
		nameHighlights := shared.HighlightFilters(project.Name, highlightFilters)
		idHighlights := shared.HighlightFilters(project.ProjectID, highlightFilters)

		projectCell := renderHighlightedText(project.Name, clistyles.PanelText, nameHighlights, rowBG)
		idCell := renderHighlightedText(project.ProjectID, clistyles.PanelMuted, idHighlights, rowBG)
		updatedAt := humanDateTime(project.UpdatedAt)
		syncedAt := humanDateTime(project.SyncedAt)

		row := []string{
			projectCell,
			idCell,
			project.ProjectNumber,
			project.AuthID,
			updatedAt,
			syncedAt,
		}
		projectWidth = max(projectWidth, lipgloss.Width(project.Name))
		idWidth = max(idWidth, lipgloss.Width(project.ProjectID))
		numberWidth = max(numberWidth, lipgloss.Width(project.ProjectNumber))
		authWidth = max(authWidth, lipgloss.Width(project.AuthID))
		updatedAtWidth = max(updatedAtWidth, lipgloss.Width(updatedAt))
		syncedAtWidth = max(syncedAtWidth, lipgloss.Width(syncedAt))
		if withURL {
			link := firebase.RemoteConfigConsoleURL(project.ProjectID)
			linkCell := link
			if !noColor {
				linkCell = applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteBlueBright), rowBG).Render(link)
			}
			row = append(row, linkCell)
			linkWidth = max(linkWidth, lipgloss.Width(link))
		}
		rows = append(rows, row)
	}

	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if noColor {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if row >= 0 && row%2 == 1 {
			style = style.Background(clistyles.ColorRowStripe)
		}
		if col == 0 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		if withURL && col == 6 {
			return style.Foreground(clistyles.PaletteBlueBright)
		}
		return style.Foreground(clistyles.PaletteSlateDim)
	}

	headers := []string{"Project", "Project ID", "Number", "Auth", "Updated At", "Synced At"}
	width := projectWidth + idWidth + numberWidth + authWidth + updatedAtWidth + syncedAtWidth + 19
	if withURL {
		headers = append(headers, "URL")
		width += linkWidth + 3
	}

	tbl := table.New().
		Headers(headers...).
		Rows(rows...).
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

type projectJSON struct {
	Project      string   `json:"project"`
	ProjectID    string   `json:"project_id"`
	Number       string   `json:"number,omitempty"`
	State        string   `json:"state,omitempty"`
	ETag         string   `json:"etag,omitempty"`
	AuthID       string   `json:"auth_id"`
	DiscoveredBy []string `json:"discovered_by,omitempty"`
	UpdatedAt    string   `json:"updated_at,omitempty"`
	SyncedAt     string   `json:"synced_at,omitempty"`
	URL          string   `json:"url,omitempty"`
}

func projectsJSON(projects []core.Project, withURL bool) []projectJSON {
	out := make([]projectJSON, len(projects))
	for i, project := range projects {
		out[i] = projectJSON{
			Project:      project.Name,
			ProjectID:    project.ProjectID,
			Number:       project.ProjectNumber,
			State:        project.State,
			ETag:         project.ETag,
			AuthID:       project.AuthID,
			DiscoveredBy: append([]string(nil), project.DiscoveredBy...),
			UpdatedAt:    project.UpdatedAt,
			SyncedAt:     project.SyncedAt,
		}
		if withURL {
			out[i].URL = firebase.RemoteConfigConsoleURL(project.ProjectID)
		}
	}
	return out
}

func renderHighlightedText(value string, base lipgloss.Style, indices []int, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return value
	}
	if len(indices) == 0 {
		return applyBackground(base, rowBG).Render(value)
	}

	highlighted := indicesSet(indices)
	highlightStyle := base.Foreground(clistyles.PaletteYellow)
	base = applyBackground(base, rowBG)
	highlightStyle = applyBackground(highlightStyle, rowBG)

	var b strings.Builder
	for i, r := range []rune(value) {
		style := base
		if highlighted[i] {
			style = highlightStyle
		}
		b.WriteString(style.Render(string(r)))
	}
	return b.String()
}

func indicesSet(indices []int) map[int]bool {
	set := make(map[int]bool, len(indices))
	for _, index := range indices {
		set[index] = true
	}
	return set
}

func applyBackground(style lipgloss.Style, bg color.Color) lipgloss.Style {
	if bg == nil {
		return style
	}
	return style.Background(bg)
}

func humanDateTime(value string) string {
	if value == "" {
		return ""
	}

	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}

	return t.Local().Format("2006-01-02 15:04:05")
}
