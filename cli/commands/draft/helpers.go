package draft

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

type listItem struct {
	ProjectID     string         `json:"project_id"`
	Project       string         `json:"project"`
	BaseVersion   string         `json:"base_version,omitempty"`
	CreatedAt     *time.Time     `json:"created_at,omitempty"`
	UpdatedAt     *time.Time     `json:"updated_at,omitempty"`
	Size          int64          `json:"size"`
	Status        string         `json:"status"`
	Valid         bool           `json:"valid"`
	BaseAvailable bool           `json:"base_available"`
	Path          string         `json:"path"`
	Changes       map[string]int `json:"changes,omitempty"`
}

func loadItems(rawFilters []string) ([]listItem, error) {
	ids, err := config.ListDraftProjectIDs()
	if err != nil {
		return nil, err
	}
	names := projectNames()
	filters := shared.ParseFilters(rawFilters)
	items := make([]listItem, 0, len(ids))
	for _, projectID := range ids {
		name := names[projectID]
		if len(filters) > 0 && !shared.MatchAnyFilter(projectID, filters) && !shared.MatchAnyFilter(name, filters) {
			continue
		}
		path := config.GetDraftPath(projectID)
		item := listItem{ProjectID: projectID, Project: name, Path: path, Status: "invalid"}
		if info, statErr := os.Stat(path); statErr == nil {
			item.Size = info.Size()
		}
		stored, loadErr := config.LoadDraft(projectID)
		if loadErr == nil {
			item.BaseVersion = stored.BaseVersion
			created, updated := stored.CreatedAt, stored.UpdatedAt
			item.CreatedAt, item.UpdatedAt = &created, &updated
			baseCfg, baseErr := firebase.ParseRemoteConfig(stored.BaseRemoteConfig)
			draftCfg, draftErr := firebase.ParseRemoteConfig(stored.RemoteConfig)
			if baseErr == nil && draftErr == nil {
				result := rcdiff.CompareRemoteConfigs(baseCfg, draftCfg)
				item.Valid, item.BaseAvailable = true, true
				item.Status = "ready"
				if !result.HasChanges() {
					item.Status = "unchanged"
				}
				p, g, c := result.ParameterSummary(), result.GroupDescriptionSummary(), result.ConditionSummary()
				item.Changes = map[string]int{"parameters": p.Added + p.Removed + p.Changed, "group_descriptions": g.Added + g.Removed + g.Changed, "conditions": c.Added + c.Removed + c.Changed}
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func renderList(items []listItem) string {
	noColor := clistyles.NoColorEnabled()
	headers := []string{"Project ID", "Project", "Base", "Updated", "Changes", "Status"}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = lipgloss.Width(header)
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		updated := ""
		if item.UpdatedAt != nil {
			updated = item.UpdatedAt.Local().Format("2006-01-02 15:04")
		}
		changes := "unknown"
		if item.Changes != nil {
			changes = fmt.Sprintf("%d params, %d conditions", item.Changes["parameters"], item.Changes["conditions"])
		}
		row := []string{item.ProjectID, item.Project, item.BaseVersion, updated, changes, item.Status}
		for i, cell := range row {
			widths[i] = max(widths[i], lipgloss.Width(cell))
		}
		rows = append(rows, row)
	}

	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if col == 2 {
			style = style.AlignHorizontal(lipgloss.Right)
		}
		if noColor {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if row >= 0 && row%2 == 1 {
			style = style.Background(clistyles.ColorRowStripe)
		}
		if col == 1 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateDim)
	}

	width := 3*len(headers) + 1
	for _, cellWidth := range widths {
		width += cellWidth
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

func projectNames() map[string]string {
	projects, err := config.LoadProjects()
	if err != nil {
		return nil
	}
	out := make(map[string]string, len(projects))
	for _, project := range projects {
		out[project.ProjectID] = project.Name
	}
	return out
}

func resolveDraft(query string) (string, core.Project, error) {
	ids, err := config.ListDraftProjectIDs()
	if err != nil {
		return "", core.Project{}, err
	}
	for _, id := range ids {
		if strings.EqualFold(id, strings.TrimSpace(query)) {
			return id, projectForID(id), nil
		}
	}
	mode, value := filter.ParseModePrefixedQuery(query)
	matches := make([]string, 0)
	for _, id := range ids {
		project := projectForID(id)
		nameMatch, _ := filter.Match(project.Name, value, mode)
		idMatch, _ := filter.Match(id, value, mode)
		if nameMatch || idMatch {
			matches = append(matches, id)
		}
	}
	if len(matches) == 1 {
		return matches[0], projectForID(matches[0]), nil
	}
	if len(matches) == 0 {
		return "", core.Project{}, fmt.Errorf("draft not found for %q", query)
	}
	return "", core.Project{}, fmt.Errorf("several drafts match %q: %s", query, strings.Join(matches, ", "))
}

func projectForID(projectID string) core.Project {
	projects, _ := config.LoadProjects()
	for _, project := range projects {
		if project.ProjectID == projectID {
			return project
		}
	}
	return core.Project{ProjectID: projectID}
}

func selectedDraftIDs(cmd *cobra.Command, args []string) ([]string, error) {
	all, _ := cmd.Flags().GetBool("all")
	if all && len(args) > 0 {
		return nil, fmt.Errorf("project arguments cannot be used with --all")
	}
	if !all && len(args) == 0 {
		return nil, fmt.Errorf("provide at least one project or use --all")
	}
	if all {
		return config.ListDraftProjectIDs()
	}
	ids := make([]string, 0, len(args))
	seen := make(map[string]bool)
	for _, query := range args {
		id, _, err := resolveDraft(query)
		if err != nil {
			return nil, err
		}
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	slices.Sort(ids)
	return ids, nil
}

func readDiffOptions(cmd *cobra.Command) diffOptions {
	o := diffOptions{}
	o.against, _ = cmd.Flags().GetString("against")
	o.cached, _ = cmd.Flags().GetBool("cached")
	o.json, _ = cmd.Flags().GetBool("json")
	o.parameters, _ = cmd.Flags().GetBool("parameters")
	o.conditions, _ = cmd.Flags().GetBool("conditions")
	o.filters, _ = cmd.Flags().GetStringArray("filter")
	o.groups, _ = cmd.Flags().GetStringArray("group")
	o.search, _ = cmd.Flags().GetString("search")
	o.expr, _ = cmd.Flags().GetString("expr")
	return o
}

func filterDiff(project core.Project, result rcdiff.Result, from, to *firebase.RemoteConfig, opts diffOptions) rcdiff.Result {
	if opts.parameters {
		result.Conditions = nil
	}
	if opts.conditions {
		result.Parameters = nil
		result.GroupDescriptions = nil
		return result
	}
	filters := shared.ParseFilters(opts.filters)
	groupSet := make(map[string]bool)
	for _, group := range opts.groups {
		groupSet[group] = true
	}
	compiled, ok := shared.CompileExpr(strings.TrimSpace(opts.expr), project.ProjectID)
	if !ok {
		result.Parameters = nil
		result.GroupDescriptions = nil
		return result
	}
	search := shared.NewParameterSearch(opts.search)
	params := result.Parameters[:0]
	for _, change := range result.Parameters {
		param, cfg := change.Final, to
		if param == nil {
			param, cfg = change.Current, from
		}
		if param == nil || len(groupSet) > 0 && !groupSet[change.Group] || !shared.MatchAnyFilter(change.Key, filters) || !shared.MatchParameterSearch(change.Key, *param, cfg, search) {
			continue
		}
		group := change.Group
		if group == "" {
			group = "default"
		}
		match, valid := shared.MatchParameterByCompiledExpr(compiled, project, cfg, change.Key, group)
		if valid && match {
			params = append(params, change)
		}
	}
	result.Parameters = params
	if len(groupSet) > 0 {
		groups := result.GroupDescriptions[:0]
		for _, change := range result.GroupDescriptions {
			if groupSet[change.Group] {
				groups = append(groups, change)
			}
		}
		result.GroupDescriptions = groups
	}
	return result
}
