package projects

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
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
	aliasesByID := map[string][]string{}
	if core.ExecutionPolicyFromContext(shared.CommandContext(cmd)).ReadLocalState {
		aliases, err := config.LoadProjectAliases()
		if err != nil {
			return err
		}
		aliasesByID = config.ProjectAliasesByID(aliases)
	}

	projects = shared.FilterProjectsWithAliases(projects, filterValues, aliasesByID)
	projects, err = shared.FilterProjectsByExpr(shared.CommandContext(cmd), svc, projects, projectExpr)
	if err != nil {
		return err
	}
	highlightFilters := shared.ParseFilters(filterValues)

	if jsonOut {
		if err := shared.WriteJSON(cmd, projectsJSONWithAliases(projects, withURL, aliasesByID)); err != nil {
			return fmt.Errorf("encode projects json: %w", err)
		}
		logProjectsTotal(projects)
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderProjectsTableWithAliases(projects, highlightFilters, withURL, aliasesByID))
	logProjectsTotal(projects)
	return nil
}

func logProjectsTotal(projects []core.Project) {
	corelog.For("projects").Info("total", "projects", len(projects))
}

func renderProjectsTableWithAliases(projects []core.Project, highlightFilters []shared.QueryFilter, withURL bool, aliasesByID map[string][]string) string {
	return renderProjectsTableAtWidth(projects, highlightFilters, withURL, aliasesByID, shared.TerminalWidth())
}

func renderProjectsTableAtWidth(projects []core.Project, highlightFilters []shared.QueryFilter, withURL bool, aliasesByID map[string][]string, terminalWidth int) string {
	noColor := clistyles.NoColorEnabled()
	headers := []string{"Project", "Project ID", "Aliases", "Number", "Auth", "Updated At", "Synced At"}
	if withURL {
		headers = append(headers, "URL")
	}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = lipgloss.Width(header)
	}
	rawRows := make([][]string, 0, len(projects))
	for _, project := range projects {
		aliases := strings.Join(aliasesByID[project.ProjectID], ", ")
		if aliases == "" {
			aliases = "—"
		}
		updatedAt := shared.FormatDateTime(project.UpdatedAt)
		syncedAt := shared.FormatDateTime(project.SyncedAt)
		row := []string{
			project.Name,
			project.ProjectID,
			aliases,
			project.ProjectNumber,
			projectAuthLabel(project),
			updatedAt,
			syncedAt,
		}
		if withURL {
			row = append(row, firebase.RemoteConfigConsoleURL(project.ProjectID))
		}
		for column, value := range row {
			widths[column] = max(widths[column], lipgloss.Width(value))
		}
		rawRows = append(rawRows, row)
	}
	tableWidth := func() int {
		width := 3*len(headers) + 1
		for _, columnWidth := range widths {
			width += columnWidth
		}
		return width
	}
	if terminalWidth > 0 && tableWidth() > terminalWidth {
		priority := []int{2}
		if withURL {
			priority = append(priority, 7)
		}
		priority = append(priority, 0, 5, 6, 4, 3, 1)
		for _, minimumMode := range []bool{true, false} {
			for _, column := range priority {
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
		for row := range rawRows {
			for column := range rawRows[row] {
				rawRows[row][column] = ansi.Truncate(rawRows[row][column], widths[column], "…")
			}
		}
	}

	rows := make([][]string, 0, len(rawRows))
	for rowIndex, raw := range rawRows {
		var rowBG color.Color
		if !noColor && rowIndex%2 == 1 {
			rowBG = clistyles.ColorRowStripe
		}
		row := append([]string(nil), raw...)
		row[0] = renderHighlightedText(raw[0], clistyles.PanelText, shared.HighlightFilters(raw[0], highlightFilters), rowBG)
		row[1] = renderHighlightedText(raw[1], clistyles.PanelMuted, shared.HighlightFilters(raw[1], highlightFilters), rowBG)
		row[2] = renderHighlightedText(raw[2], clistyles.PanelMuted, shared.HighlightFilters(raw[2], highlightFilters), rowBG)
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
		if withURL && col == 7 {
			return style.Foreground(clistyles.PaletteBlueBright)
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
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func projectAuthLabel(project core.Project) string {
	if project.Disabled {
		return project.AuthID + " (disabled)"
	}
	return project.AuthID
}

type projectJSON = shared.ProjectJSON

func projectsJSON(projects []core.Project, withURL bool) []projectJSON {
	aliases, _ := config.LoadProjectAliases()
	return projectsJSONWithAliases(projects, withURL, config.ProjectAliasesByID(aliases))
}

func projectsJSONWithAliases(projects []core.Project, withURL bool, aliasesByID map[string][]string) []projectJSON {
	out := make([]projectJSON, len(projects))
	for i, project := range projects {
		out[i] = shared.NewProjectJSONWithAliases(project, aliasesByID[project.ProjectID], withURL)
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
