package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// New constructs new and returns the resulting value or error.
func New(svc *core.Core) *cobra.Command {
	projectsCmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects list",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List projects using cache-first loading",
		RunE: func(cmd *cobra.Command, args []string) error {
			forceUpdate, err := cmd.Flags().GetBool("update")
			if err != nil {
				return err
			}

			var projects []core.Project
			var source string
			if forceUpdate {
				projects, source, err = svc.SyncProjects(context.Background())
			} else {
				projects, source, err = svc.ListProjects(context.Background())
			}
			if err != nil {
				return err
			}

			return printProjects(cmd, svc, projects, source)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update projects from Firebase into cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			authID, err := cmd.Flags().GetString("auth")
			if err != nil {
				return err
			}
			var projects []core.Project
			var source string
			if authID != "" {
				projects, source, err = svc.SyncProjectsForAuth(context.Background(), authID)
			} else {
				projects, source, err = svc.SyncProjects(context.Background())
			}
			if err != nil {
				return err
			}

			return printProjects(cmd, svc, projects, source)
		},
	}

	listCmd.Flags().Bool("json", false, "Print projects as JSON")
	listCmd.Flags().StringArrayP("filter", "f", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	listCmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	listCmd.Flags().Bool("update", false, "Update projects from Firebase before printing")
	listCmd.Flags().Bool("url", false, "Include Firebase Console Remote Config URL")
	updateCmd.Flags().Bool("json", false, "Print projects as JSON")
	updateCmd.Flags().StringArrayP("filter", "f", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	updateCmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	updateCmd.Flags().Bool("url", false, "Include Firebase Console Remote Config URL")
	updateCmd.Flags().String("auth", "", "Sync projects for one auth id")

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print projects config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := config.GetProjectsFilePath()
			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(map[string]string{"path": path}); err != nil {
					return fmt.Errorf("encode projects path json: %w", err)
				}
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	pathCmd.Flags().Bool("json", false, "Print path as JSON")

	purgeCmd := &cobra.Command{
		Use:   "purge",
		Short: "Delete cached projects config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Delete cached projects config file %s?", config.GetProjectsFilePath()),
					confirmation.Yes,
					shared.ConfirmationOptions{Destructive: true},
				)
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			if err := svc.PurgeProjects(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged: %s\n", config.GetProjectsFilePath())
			return nil
		},
	}
	purgeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation dialog")

	projectsCmd.AddCommand(listCmd, updateCmd, pathCmd, purgeCmd)
	return projectsCmd
}

// printProjects handles print projects and returns the resulting value or error.
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
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(projectsJSON(projects, withURL)); err != nil {
			return fmt.Errorf("encode projects json: %w", err)
		}
		logProjectsTotal(projects)
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderProjectsTable(projects, highlightFilters, withURL))
	logProjectsTotal(projects)
	return nil
}

// logProjectsTotal handles log projects total and returns the resulting value or error.
func logProjectsTotal(projects []core.Project) {
	corelog.For("projects").Info("total", "projects", len(projects))
}

// renderProjectsTable renders render projects table and returns the resulting value or error.
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

// projectJSON holds project json state used by the projects package.
type projectJSON struct {
	// Project stores project for projectJSON.
	Project string `json:"project"`
	// ProjectID stores project id for projectJSON.
	ProjectID string `json:"project_id"`
	// Number stores number for projectJSON.
	Number string `json:"number,omitempty"`
	// State stores state for projectJSON.
	State string `json:"state,omitempty"`
	// ETag stores etag for projectJSON.
	ETag string `json:"etag,omitempty"`
	// AuthID stores auth id for projectJSON.
	AuthID string `json:"auth_id"`
	// DiscoveredBy stores discovering auth ids for projectJSON.
	DiscoveredBy []string `json:"discovered_by,omitempty"`
	// UpdatedAt stores updated at for projectJSON.
	UpdatedAt string `json:"updated_at,omitempty"`
	// SyncedAt stores synced at for projectJSON.
	SyncedAt string `json:"synced_at,omitempty"`
	// URL stores url for projectJSON.
	URL string `json:"url,omitempty"`
}

// projectsJSON handles projects json and returns the resulting value or error.
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

// renderHighlightedText renders render highlighted text and returns the resulting value or error.
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

// indicesSet handles indices set and returns the resulting value or error.
func indicesSet(indices []int) map[int]bool {
	set := make(map[int]bool, len(indices))
	for _, index := range indices {
		set[index] = true
	}
	return set
}

// applyBackground handles apply background and returns the resulting value or error.
func applyBackground(style lipgloss.Style, bg color.Color) lipgloss.Style {
	if bg == nil {
		return style
	}
	return style.Background(bg)
}

// humanDateTime handles human date time and returns the resulting value or error.
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
