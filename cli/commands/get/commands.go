package get

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"fbrcm/cli/shared"
	clistyles "fbrcm/cli/styles"
	"fbrcm/core"
	"fbrcm/core/filter"
	"fbrcm/core/firebase"
	corelog "fbrcm/core/log"
)

const defaultGroupLabel = "(default)"

type parameterConditionJSON struct {
	Name  string  `json:"name"`
	Value *string `json:"value"`
}

type parameterRowJSON struct {
	Project      string                   `json:"project"`
	ProjectID    string                   `json:"project_id"`
	Group        string                   `json:"group"`
	Key          string                   `json:"key"`
	DefaultValue *string                  `json:"default_value"`
	Conditional  bool                     `json:"conditional"`
	Conditions   []parameterConditionJSON `json:"conditions"`
	Type         string                   `json:"type"`
	Version      *string                  `json:"version"`
	CachedAt     *time.Time               `json:"cached_at"`
	Status       *string                  `json:"status"`
}

type parameterRow struct {
	Project      string
	ProjectID    string
	Group        string
	Key          string
	DefaultValue *string
	Conditional  bool
	Conditions   []parameterConditionJSON
	Type         string
	Version      string
	CachedAt     time.Time
	Status       string
	ValueLines   []valueLine
}

type tableLayout struct {
	includeProject bool
	includeGroup   bool
	includeKey     bool
	includeType    bool
	showNames      bool
	valueWidth     int
}

func New(svc *core.Core) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get parameters from all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectFilter, err := cmd.Flags().GetString("project")
			if err != nil {
				return err
			}
			projectExpr, err := cmd.Flags().GetString("expr")
			if err != nil {
				return err
			}
			paramFilter, err := cmd.Flags().GetString("filter")
			if err != nil {
				return err
			}
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			update, err := cmd.Flags().GetBool("update")
			if err != nil {
				return err
			}

			if stdinAvailable(cmd.InOrStdin()) {
				corelog.For("get").Info("stdin mode enabled; using remote config from stdin")
				compiledExpr, ok := shared.CompileExpr(projectExpr, "<stdin>")
				if !ok {
					return nil
				}
				cfg, rows, err := loadStdinParameterRows(cmd)
				if err != nil {
					return err
				}
				rows = filterParameterRows(rows, paramFilter)
				rows = filterParameterRowsByExpr(rows, core.Project{Name: "<stdin>", ProjectID: "<stdin>"}, cfg, compiledExpr)
				sortParameterRows(rows)

				if jsonOut {
					return writeRowsJSON(cmd, rows)
				}

				mode, query := parseFilter(paramFilter)
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderParametersTable(rows, mode, query, false))
				logGetTotals("table-stdin", rows)
				return nil
			}

			projects, _, err := svc.ListProjects(context.Background())
			if err != nil {
				return err
			}
			projects = filterProjects(projects, projectFilter)
			sortProjects(projects)

			loaded, err := loadProjectsParameters(context.Background(), svc, projects, update)
			if err != nil {
				return err
			}
			compiledExpr, ok := shared.CompileExpr(projectExpr, "")
			if !ok {
				return nil
			}

			rows := make([]parameterRow, 0)
			for _, item := range loaded {
				if item.cfg == nil || item.cache == nil {
					continue
				}
				rows = append(rows, flattenParameters(item.project, item.cfg, item.cache.CachedAt, item.status, "", compiledExpr)...)
			}

			rows = filterParameterRows(rows, paramFilter)
			sortParameterRows(rows)

			if jsonOut {
				if err := writeRowsJSON(cmd, rows); err != nil {
					return err
				}
				logGetTotals("json", rows)
				return nil
			}

			mode, query := parseFilter(paramFilter)
			projectMode, _ := parseFilter(projectFilter)
			tableRows := buildTableRows(loaded, rows)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderParametersTable(tableRows, mode, query, projectMode != filter.ModeExact))
			logGetTotals("table", tableRows)
			return nil
		},
	}

	getCmd.Flags().Bool("json", false, "Print parameters as JSON")
	getCmd.Flags().StringP("project", "p", "", "Filter projects by mode-prefixed query (^, /, ~, =)")
	getCmd.Flags().StringP("filter", "f", "", "Filter parameters by mode-prefixed query (^, /, ~, =)")
	getCmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	getCmd.Flags().Bool("update", false, "Revalidate cached parameters before printing")
	return getCmd
}

func writeRowsJSON(cmd *cobra.Command, rows []parameterRow) error {
	out := make([]parameterRowJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, parameterRowJSON{
			Project:      row.Project,
			ProjectID:    row.ProjectID,
			Group:        row.Group,
			Key:          row.Key,
			DefaultValue: row.DefaultValue,
			Conditional:  row.Conditional,
			Conditions:   row.Conditions,
			Type:         row.Type,
			Version:      stringPtrOrNil(row.Version),
			CachedAt:     timePtrOrNil(row.CachedAt),
			Status:       stringPtrOrNil(row.Status),
		})
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(out); err != nil {
		return fmt.Errorf("encode parameters json: %w", err)
	}
	return nil
}

func loadStdinParameterRows(cmd *cobra.Command) (*firebase.RemoteConfig, []parameterRow, error) {
	raw, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return nil, nil, fmt.Errorf("read stdin: %w", err)
	}
	corelog.For("get").Info("loaded remote config from stdin", "bytes", len(raw))
	if !json.Valid(raw) {
		return nil, nil, fmt.Errorf("stdin remote config is not valid json")
	}

	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode stdin remote config: %w", err)
	}

	version := stdinVersion(raw)
	project := core.Project{
		Name:      "<stdin>",
		ProjectID: "<stdin>",
	}
	rows := flattenParameters(project, cfg, time.Time{}, "", version, nil)
	corelog.For("get").Info("parsed stdin remote config", "parameters", len(rows), "version", version)
	return cfg, rows, nil
}

func stdinVersion(raw []byte) string {
	var payload struct {
		Version *struct {
			VersionNumber string `json:"versionNumber"`
		} `json:"version"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if payload.Version == nil {
		return ""
	}
	return strings.TrimSpace(payload.Version.VersionNumber)
}

func loadProjectParameters(ctx context.Context, svc *core.Core, projectID string, update bool) (*core.ParametersCache, string, error) {
	if update {
		return svc.RevalidateParameters(ctx, projectID)
	}
	return svc.GetParameters(ctx, projectID, false)
}

type loadedProjectParameters struct {
	project core.Project
	cache   *core.ParametersCache
	cfg     *firebase.RemoteConfig
	source  string
	status  string
}

func loadProjectsParameters(ctx context.Context, svc *core.Core, projects []core.Project, update bool) ([]loadedProjectParameters, error) {
	if len(projects) == 0 {
		return nil, nil
	}

	type job struct {
		index   int
		project core.Project
	}
	type result struct {
		index  int
		loaded loadedProjectParameters
		err    error
	}

	jobs := make(chan job)
	results := make(chan result, len(projects))

	workerCount := min(firebase.MaxConcurrentRequests(), len(projects))

	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for work := range jobs {
				loaded, err := loadProjectParametersWithFallback(ctx, svc, work.project, update)
				select {
				case results <- result{index: work.index, loaded: loaded, err: err}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for i, project := range projects {
			select {
			case jobs <- job{index: i, project: project}:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	loaded := make([]loadedProjectParameters, len(projects))
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		loaded[res.index] = res.loaded
	}

	return loaded, nil
}

func loadProjectParametersWithFallback(ctx context.Context, svc *core.Core, project core.Project, update bool) (loadedProjectParameters, error) {
	cache, source, err := loadProjectParameters(ctx, svc, project.ProjectID, update)
	if err == nil {
		cfg, parseErr := firebase.ParseRemoteConfig(cache.RemoteConfig)
		if parseErr != nil {
			return loadedProjectParameters{}, fmt.Errorf("decode remote config for %s: %w", project.ProjectID, parseErr)
		}
		return loadedProjectParameters{
			project: project,
			cache:   cache,
			cfg:     cfg,
			source:  source,
			status:  core.ParametersStatusLabel(source, cache.CachedAt, true, nil),
		}, nil
	}

	cache, state, inspectErr := svc.InspectParametersCache(project.ProjectID)
	if inspectErr != nil {
		return loadedProjectParameters{}, err
	}
	if state != core.ParametersCacheMissing && cache != nil {
		cfg, parseErr := firebase.ParseRemoteConfig(cache.RemoteConfig)
		if parseErr != nil {
			return loadedProjectParameters{}, fmt.Errorf("decode cached remote config for %s: %w", project.ProjectID, parseErr)
		}
		return loadedProjectParameters{
			project: project,
			cache:   cache,
			cfg:     cfg,
			source:  "cache-stale",
			status:  "staled",
		}, nil
	}

	return loadedProjectParameters{
		project: project,
		status:  "missing",
	}, nil
}

func flattenParameters(project core.Project, cfg *firebase.RemoteConfig, cachedAt time.Time, status, version string, compiledExpr *filter.Expression) []parameterRow {
	if cfg == nil {
		return nil
	}

	if version == "" {
		version = strings.TrimSpace(cfg.Version.VersionNumber)
	}

	conditionOrder := make(map[string]int, len(cfg.Conditions))
	conditionColors := make(map[string]string, len(cfg.Conditions))
	for i, condition := range cfg.Conditions {
		conditionOrder[condition.Name] = i
		conditionColors[condition.Name] = condition.TagColor
	}

	rows := make([]parameterRow, 0)
	seen := make(map[string]struct{})
	groupKeys := sortedStringKeys(cfg.ParameterGroups)
	for _, groupKey := range groupKeys {
		group := cfg.ParameterGroups[groupKey]
		paramKeys := sortedStringKeys(group.Parameters)
		for _, key := range paramKeys {
			match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, key, groupKey)
			if !ok || !match {
				continue
			}
			seen[key] = struct{}{}
			rows = append(rows, buildParameterRow(project, groupKey, key, group.Parameters[key], version, cachedAt, status, conditionOrder, conditionColors))
		}
	}

	rootParams := make(map[string]firebase.RemoteConfigParam)
	for key, param := range cfg.Parameters {
		if _, ok := seen[key]; ok {
			continue
		}
		rootParams[key] = param
	}
	for _, key := range sortedStringKeys(rootParams) {
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, key, defaultGroupLabel)
		if !ok || !match {
			continue
		}
		rows = append(rows, buildParameterRow(project, defaultGroupLabel, key, rootParams[key], version, cachedAt, status, conditionOrder, conditionColors))
	}

	return rows
}

func buildParameterRow(project core.Project, group, key string, param firebase.RemoteConfigParam, version string, cachedAt time.Time, status string, conditionOrder map[string]int, conditionColors map[string]string) parameterRow {
	conditions := make([]parameterConditionJSON, 0, len(param.ConditionalValues))
	valueLines := make([]valueLine, 0, len(param.ConditionalValues)+1)

	for _, name := range sortedConditionalKeys(param.ConditionalValues, conditionOrder) {
		value := formatRemoteConfigValue(param.ConditionalValues[name], param.ValueType)
		conditions = append(conditions, parameterConditionJSON{
			Name:  name,
			Value: valueForJSON(value),
		})
		valueLines = append(valueLines, valueLine{
			Label:     name,
			Value:     value,
			Color:     clistyles.ConditionLipglossColor(conditionColors[name]),
			IsDefault: false,
			ValueType: valueTypeKey(param.ValueType),
		})
	}

	var defaultValue *string
	if param.DefaultValue != nil {
		formatted := formatRemoteConfigValue(*param.DefaultValue, param.ValueType)
		defaultValue = valueForJSON(formatted)
		valueLines = append(valueLines, valueLine{
			Label:     "Default value",
			Value:     formatted,
			IsDefault: true,
			ValueType: valueTypeKey(param.ValueType),
		})
	}

	valueType := strings.TrimSpace(param.ValueType)
	if valueType == "" {
		valueType = "string"
	}

	return parameterRow{
		Project:      project.Name,
		ProjectID:    project.ProjectID,
		Group:        group,
		Key:          key,
		DefaultValue: defaultValue,
		Conditional:  len(conditions) > 0,
		Conditions:   conditions,
		Type:         valueType,
		Version:      version,
		CachedAt:     cachedAt,
		Status:       status,
		ValueLines:   valueLines,
	}
}

func filterProjects(projects []core.Project, raw string) []core.Project {
	mode, query := parseFilter(raw)
	if query == "" {
		return projects
	}

	filtered := make([]core.Project, 0, len(projects))
	for _, project := range projects {
		nameMatch, _ := filter.Match(project.Name, query, mode)
		idMatch, _ := filter.Match(project.ProjectID, query, mode)
		if nameMatch || idMatch {
			filtered = append(filtered, project)
		}
	}
	return filtered
}

func filterParameterRows(rows []parameterRow, raw string) []parameterRow {
	mode, query := parseFilter(raw)
	if query == "" {
		return rows
	}

	filtered := make([]parameterRow, 0, len(rows))
	for _, row := range rows {
		match, _ := filter.Match(row.Key, query, mode)
		if match {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func filterParameterRowsByExpr(rows []parameterRow, project core.Project, cfg *firebase.RemoteConfig, compiledExpr *filter.Expression) []parameterRow {
	if compiledExpr == nil {
		return rows
	}

	filtered := make([]parameterRow, 0, len(rows))
	for _, row := range rows {
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, row.Key, row.Group)
		if ok && match {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func parseFilter(raw string) (filter.Mode, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return filter.ModeFuzzy, ""
	}

	mode, ok := filter.ModeFromLabel(string([]rune(raw)[0]))
	if !ok {
		return filter.ModeFuzzy, raw
	}

	return mode, string([]rune(raw)[1:])
}

func sortProjects(projects []core.Project) {
	sort.Slice(projects, func(i, j int) bool {
		leftName := strings.ToLower(strings.TrimSpace(projects[i].Name))
		rightName := strings.ToLower(strings.TrimSpace(projects[j].Name))
		if leftName == "" {
			leftName = strings.ToLower(projects[i].ProjectID)
		}
		if rightName == "" {
			rightName = strings.ToLower(projects[j].ProjectID)
		}
		if leftName == rightName {
			return strings.ToLower(projects[i].ProjectID) < strings.ToLower(projects[j].ProjectID)
		}
		return leftName < rightName
	})
}

func sortParameterRows(rows []parameterRow) {
	sort.Slice(rows, func(i, j int) bool {
		leftProject := strings.ToLower(strings.TrimSpace(rows[i].Project))
		rightProject := strings.ToLower(strings.TrimSpace(rows[j].Project))
		if leftProject == "" {
			leftProject = strings.ToLower(rows[i].ProjectID)
		}
		if rightProject == "" {
			rightProject = strings.ToLower(rows[j].ProjectID)
		}
		switch {
		case leftProject != rightProject:
			return leftProject < rightProject
		case !strings.EqualFold(rows[i].ProjectID, rows[j].ProjectID):
			return strings.ToLower(rows[i].ProjectID) < strings.ToLower(rows[j].ProjectID)
		case !strings.EqualFold(rows[i].Group, rows[j].Group):
			return strings.ToLower(rows[i].Group) < strings.ToLower(rows[j].Group)
		default:
			return strings.ToLower(rows[i].Key) < strings.ToLower(rows[j].Key)
		}
	})
}

func buildTableRows(projects []loadedProjectParameters, rows []parameterRow) []parameterRow {
	rowsByProject := make(map[string][]parameterRow, len(projects))
	for _, row := range rows {
		rowsByProject[row.ProjectID] = append(rowsByProject[row.ProjectID], row)
	}

	out := make([]parameterRow, 0, len(rows)+len(projects))
	for _, project := range projects {
		projectRows := rowsByProject[project.project.ProjectID]
		if len(projectRows) == 0 {
			out = append(out, parameterRow{
				Project:    project.project.Name,
				ProjectID:  project.project.ProjectID,
				Status:     project.status,
				ValueLines: []valueLine{{Label: "Missing values", Missing: true}},
			})
			continue
		}
		out = append(out, projectRows...)
	}

	return out
}

func renderParametersTable(rows []parameterRow, mode filter.Mode, query string, includeProject bool) string {
	noColor := clistyles.NoColorEnabled()
	projectWidth := lipgloss.Width("Project")
	groupWidth := lipgloss.Width("Group")
	keyWidth := lipgloss.Width("Key")
	typeWidth := lipgloss.Width("Type")
	globalLabelWidth := longestConditionWidth(rows)
	layout := chooseTableLayout(rows, globalLabelWidth, includeProject, mode == filter.ModeExact && strings.TrimSpace(query) != "")
	valuesWidth := max(lipgloss.Width("Values"), layout.valueWidth)
	tableRows := make([][]string, 0, len(rows))

	for _, row := range rows {
		rowIndex := len(tableRows)
		var rowBG color.Color
		if !noColor && isStripedDataRow(rowIndex) {
			rowBG = clistyles.ColorRowStripe
		}
		match, highlights := filter.Match(row.Key, query, mode)
		if !match {
			highlights = nil
		}

		projectCell := row.Project
		if strings.TrimSpace(projectCell) == "" {
			projectCell = row.ProjectID
		}
		keyCell := renderHighlightedText(row.Key, clistyles.PanelText, highlights, rowBG)

		rowCells := make([]string, 0, 5)
		if layout.includeProject {
			rowCells = append(rowCells, projectCell)
		}
		if layout.includeGroup {
			rowCells = append(rowCells, row.Group)
		}
		if layout.includeKey {
			rowCells = append(rowCells, keyCell)
		}
		if layout.includeType {
			rowCells = append(rowCells, row.Type)
		}
		renderedValues := renderValueTree(row.ValueLines, row.Status, globalLabelWidth, layout.showNames, layout.valueWidth, rowBG)
		rowCells = append(rowCells, renderedValues)
		tableRows = append(tableRows, rowCells)

		if layout.includeProject {
			projectWidth = max(projectWidth, lipgloss.Width(projectCell))
		}
		if layout.includeGroup {
			groupWidth = max(groupWidth, lipgloss.Width(row.Group))
		}
		if layout.includeKey {
			keyWidth = max(keyWidth, lipgloss.Width(row.Key))
		}
		if layout.includeType {
			typeWidth = max(typeWidth, lipgloss.Width(row.Type))
		}
		valuesWidth = max(valuesWidth, lipgloss.Width(renderedValues))
	}

	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if noColor {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if isStripedDataRow(row) {
			style = style.Background(clistyles.ColorRowStripe)
		}
		keyCol := 0
		if layout.includeProject {
			keyCol = 1
		}
		if layout.includeGroup {
			keyCol++
		}
		switch col {
		case 0:
			if layout.includeProject {
				if isErrorStatus(rowStatus(rows, row)) {
					return style.Foreground(clistyles.PaletteError)
				}
				return style.Foreground(clistyles.PaletteSlateBright)
			}
			return style.Foreground(clistyles.PaletteBlueBright)
		case keyCol:
			return style.Foreground(clistyles.PaletteBlueBright)
		default:
			return style.Foreground(clistyles.PaletteSlateDim)
		}
	}

	headers := make([]string, 0, 5)
	if layout.includeProject {
		headers = append(headers, "Project")
	}
	if layout.includeGroup {
		headers = append(headers, "Group")
	}
	if layout.includeKey {
		headers = append(headers, "Key")
	}
	if layout.includeType {
		headers = append(headers, "Type")
	}
	headers = append(headers, "Values")

	width := valuesWidth + tableOverhead(len(headers))
	if layout.includeKey {
		width += keyWidth
	}
	if layout.includeProject {
		width += projectWidth
	}
	if layout.includeGroup {
		width += groupWidth
	}
	if layout.includeType {
		width += typeWidth
	}

	tbl := table.New().
		Headers(headers...).
		Rows(tableRows...).
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

func isStripedDataRow(row int) bool {
	return row >= 0 && row%2 == 1
}

func rowStatus(rows []parameterRow, row int) string {
	if row <= 0 {
		return ""
	}
	index := row - 1
	if index < 0 || index >= len(rows) {
		return ""
	}
	return rows[index].Status
}

func chooseTableLayout(rows []parameterRow, labelWidth int, includeProject bool, allowHideKey bool) tableLayout {
	terminalWidth := detectTerminalWidth()
	layout := tableLayout{
		includeProject: includeProject,
		includeGroup:   true,
		includeKey:     true,
		includeType:    true,
		showNames:      true,
		valueWidth:     max(lipgloss.Width("Values"), maxValueWidth(rows, labelWidth, true)),
	}
	if terminalWidth <= 0 {
		return layout
	}

	projectWidth := lipgloss.Width("Project")
	groupWidth := lipgloss.Width("Group")
	keyWidth := lipgloss.Width("Key")
	typeWidth := lipgloss.Width("Type")
	for _, row := range rows {
		projectCell := row.Project
		if strings.TrimSpace(projectCell) == "" {
			projectCell = row.ProjectID
		}
		projectWidth = max(projectWidth, lipgloss.Width(projectCell))
		groupWidth = max(groupWidth, lipgloss.Width(row.Group))
		keyWidth = max(keyWidth, lipgloss.Width(row.Key))
		typeWidth = max(typeWidth, lipgloss.Width(row.Type))
	}

	available := func(includeGroup, includeKey, includeType bool) int {
		cols := 0
		width := 0
		if includeProject {
			cols++
			width += projectWidth
		}
		if includeGroup {
			cols++
			width += groupWidth
		}
		if includeKey {
			cols++
			width += keyWidth
		}
		if includeType {
			cols++
			width += typeWidth
		}
		cols++ // values
		return terminalWidth - width - tableOverhead(cols)
	}

	natural := maxValueWidth(rows, labelWidth, true)
	valueWidth := available(true, true, true)
	clippingNeeded := natural > valueWidth
	valueRoom := minValueRoom(rows, labelWidth, true, valueWidth)
	if clippingNeeded && valueRoom < 10 {
		layout.includeType = false
		valueWidth = available(true, true, false)
		clippingNeeded = natural > valueWidth
		valueRoom = minValueRoom(rows, labelWidth, true, valueWidth)
	}
	if clippingNeeded && valueRoom < 10 {
		layout.includeGroup = false
		valueWidth = available(false, true, false)
		clippingNeeded = natural > valueWidth
		valueRoom = minValueRoom(rows, labelWidth, true, valueWidth)
	}
	if clippingNeeded && valueRoom < 10 {
		layout.showNames = false
		natural = maxValueWidth(rows, labelWidth, false)
		valueWidth = available(layout.includeGroup, true, layout.includeType)
		clippingNeeded = natural > valueWidth
		valueRoom = minValueRoom(rows, labelWidth, false, valueWidth)
	}
	if allowHideKey && clippingNeeded && valueRoom < 10 {
		layout.includeKey = false
	}

	natural = maxValueWidth(rows, labelWidth, layout.showNames)
	valueWidth = available(layout.includeGroup, layout.includeKey, layout.includeType)
	if valueWidth <= 0 {
		valueWidth = 1
	}
	layout.valueWidth = max(1, min(natural, valueWidth))
	return layout
}

func detectTerminalWidth() int {
	if columns := strings.TrimSpace(os.Getenv("COLUMNS")); columns != "" {
		if width, err := strconv.Atoi(columns); err == nil && width > 0 {
			return width
		}
	}

	info, err := os.Stdout.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && width > 0 {
			return width
		}
	}

	return 80
}

func tableOverhead(cols int) int {
	return cols*3 + 1
}

func renderHighlightedText(value string, base lipgloss.Style, highlights []int, rowBG color.Color) string {
	if clistyles.NoColorEnabled() || len(highlights) == 0 {
		return value
	}

	highlightSet := make(map[int]struct{}, len(highlights))
	for _, idx := range highlights {
		highlightSet[idx] = struct{}{}
	}

	runes := []rune(value)
	parts := make([]string, 0, len(runes))
	for i, r := range runes {
		style := applyBackground(base, rowBG)
		if _, ok := highlightSet[i]; ok {
			style = applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteYellow), rowBG)
			parts = append(parts, style.Render(string(r)))
			continue
		}
		parts = append(parts, style.Render(string(r)))
	}
	return strings.Join(parts, "")
}

func renderConditionLabel(label string, conditionColor, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return label
	}
	return applyBackground(lipgloss.NewStyle().Foreground(conditionColor), rowBG).Render(label)
}

type valueLine struct {
	Label     string
	Value     string
	Color     color.Color
	IsDefault bool
	Missing   bool
	ValueType string
}

func renderValueTree(lines []valueLine, status string, labelWidth int, showNames bool, maxWidth int, rowBG color.Color) string {
	if len(lines) == 0 {
		return ""
	}

	rendered := make([]string, 0, len(lines))
	for i, line := range lines {
		prefix := valueTreePrefix(i, len(lines))
		label := line.Label
		if line.Missing {
			label = renderMissingLabel(status, rowBG)
			rendered = append(rendered, clipStyledLine(renderTreeChrome(prefix, rowBG)+renderTreeChrome(" ", rowBG)+label, maxWidth))
			continue
		} else if line.IsDefault {
			label = renderDefaultLabel(label, rowBG)
		} else {
			label = renderConditionLabel(label, line.Color, rowBG)
		}

		if !showNames {
			head := renderTreeChrome(prefix+" ", rowBG)
			value := renderValueText(clipPlainText(line.Value, max(maxWidth-lipgloss.Width(head), 1)), line.ValueType, rowBG)
			rendered = append(rendered, head+value)
			continue
		}
		fillWidth := max(labelWidth-lipgloss.Width(line.Label)+1, 1)
		filler := renderTreeChrome(strings.Repeat("╌", fillWidth), rowBG)
		head := renderTreeChrome(prefix+" ", rowBG) + label + renderTreeChrome(" ", rowBG) + filler + renderTreeChrome(" ", rowBG)
		value := renderValueText(clipPlainText(line.Value, max(maxWidth-lipgloss.Width(head), 1)), line.ValueType, rowBG)
		rendered = append(rendered, head+value)
	}

	return strings.Join(rendered, "\n")
}

func longestConditionWidth(rows []parameterRow) int {
	width := lipgloss.Width("Default value")
	for _, row := range rows {
		for _, line := range row.ValueLines {
			width = max(width, lipgloss.Width(line.Label))
		}
	}
	return width
}

func maxValueWidth(rows []parameterRow, labelWidth int, showNames bool) int {
	width := lipgloss.Width("Values")
	for _, row := range rows {
		width = max(width, lipgloss.Width(renderValueTree(row.ValueLines, row.Status, labelWidth, showNames, 1<<30, nil)))
	}
	return width
}

func minValueRoom(rows []parameterRow, labelWidth int, showNames bool, cellWidth int) int {
	room := 1 << 30
	found := false
	for _, row := range rows {
		for i, line := range row.ValueLines {
			if line.Missing {
				continue
			}
			headWidth := valueLineHeadWidth(line, i, len(row.ValueLines), labelWidth, showNames)
			valueRoom := cellWidth - headWidth
			if !found || valueRoom < room {
				room = valueRoom
				found = true
			}
		}
	}
	if !found {
		return cellWidth
	}
	return room
}

func valueTreePrefix(index, total int) string {
	if total <= 1 {
		return "╌╌╌"
	}
	switch index {
	case 0:
		return "╌┬╌"
	case total - 1:
		return " ╰╌"
	default:
		return " ├╌"
	}
}

func renderTreeChrome(value string, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return value
	}
	return applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDim), rowBG).Render(value)
}

func renderDefaultLabel(label string, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return label
	}
	return applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDim).Italic(true), rowBG).Render(label)
}

func renderMissingLabel(status string, rowBG color.Color) string {
	if clistyles.NoColorEnabled() {
		return "Missing values"
	}
	style := lipgloss.NewStyle().Italic(true).Strikethrough(true)
	if isErrorStatus(status) {
		style = style.Foreground(clistyles.PaletteError)
	} else {
		style = style.Foreground(clistyles.PaletteSlateDim)
	}
	return applyBackground(style, rowBG).Render("Missing values")
}

func isErrorStatus(status string) bool {
	return status == "staled" || status == "missing"
}

func logGetTotals(output string, rows []parameterRow) {
	logger := corelog.For("get")
	logger.Info("total", "output", output, "projects", countOutputProjects(rows), "values", countOutputValues(rows))
}

func countOutputProjects(rows []parameterRow) int {
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.ProjectID) == "" {
			continue
		}
		seen[row.ProjectID] = struct{}{}
	}
	return len(seen)
}

func countOutputValues(rows []parameterRow) int {
	total := 0
	for _, row := range rows {
		if row.DefaultValue != nil {
			total++
		}
		total += len(row.Conditions)
	}
	return total
}

func stdinAvailable(in io.Reader) bool {
	file, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

func stringPtrOrNil(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	v := value
	return &v
}

func timePtrOrNil(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	v := value
	return &v
}

func renderValueText(value, valueType string, rowBG color.Color) string {
	if value == "" || clistyles.NoColorEnabled() {
		return value
	}
	if strings.HasPrefix(value, "(empty ") && strings.HasSuffix(value, ")") {
		return applyBackground(lipgloss.NewStyle().Foreground(clistyles.PaletteSlateDim).Italic(true), rowBG).Render(value)
	}
	style := valueTextStyle(value, valueType)
	return applyBackground(style, rowBG).Render(value)
}

func valueTextStyle(value, valueType string) lipgloss.Style {
	switch valueType {
	case "boolean":
		if strings.EqualFold(value, "true") {
			return lipgloss.NewStyle().Foreground(clistyles.ConditionLipglossColor("GREEN"))
		}
		if strings.EqualFold(value, "false") {
			return lipgloss.NewStyle().Foreground(clistyles.PaletteError)
		}
	case "number":
		return lipgloss.NewStyle().Foreground(clistyles.PaletteBlueBright)
	case "json":
		return lipgloss.NewStyle().Foreground(clistyles.ConditionLipglossColor("CYAN"))
	}
	return lipgloss.NewStyle().Foreground(clistyles.PaletteSlateBright)
}

func valueTypeKey(valueType string) string {
	valueType = strings.TrimSpace(strings.ToLower(valueType))
	if valueType == "" {
		return "string"
	}
	return valueType
}

func clipStyledLine(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= maxWidth {
		return value
	}
	return clipPlainText(value, maxWidth)
}

func clipPlainText(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxWidth {
		return value
	}
	if maxWidth == 1 {
		return "…"
	}
	return string(runes[:maxWidth-1]) + "…"
}

func valueLineHeadWidth(line valueLine, index, total, labelWidth int, showNames bool) int {
	prefixWidth := lipgloss.Width(valueTreePrefix(index, total)) + 1
	if line.Missing {
		return prefixWidth
	}
	if !showNames {
		return prefixWidth
	}
	return prefixWidth + lipgloss.Width(line.Label) + 1 + max(labelWidth-lipgloss.Width(line.Label)+1, 1) + 1
}

func applyBackground(style lipgloss.Style, bg color.Color) lipgloss.Style {
	if bg == nil {
		return style
	}
	return style.Background(bg)
}

func sortedStringKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := strings.ToLower(keys[i])
		right := strings.ToLower(keys[j])
		if left == right {
			return keys[i] < keys[j]
		}
		return left < right
	})
	return keys
}

func sortedConditionalKeys(items map[string]firebase.RemoteConfigValue, order map[string]int) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		left, leftOK := order[keys[i]]
		right, rightOK := order[keys[j]]
		switch {
		case leftOK && rightOK && left != right:
			return left < right
		case leftOK != rightOK:
			return leftOK
		default:
			leftKey := strings.ToLower(keys[i])
			rightKey := strings.ToLower(keys[j])
			if leftKey == rightKey {
				return keys[i] < keys[j]
			}
			return leftKey < rightKey
		}
	})

	return keys
}

func formatRemoteConfigValue(value firebase.RemoteConfigValue, valueType string) string {
	switch {
	case value.UseInAppDefault:
		return "<in-app default>"
	case len(value.PersonalizationValue) > 0:
		return "<personalization>"
	case len(value.RolloutValue) > 0:
		return "<rollout>"
	case value.Value == "":
		return "(empty " + emptyValueType(valueType) + ")"
	default:
		return strings.ReplaceAll(value.Value, "\n", "\\n")
	}
}

func valueForJSON(value string) *string {
	if strings.HasPrefix(value, "(empty ") && strings.HasSuffix(value, ")") {
		return nil
	}
	v := value
	return &v
}

func emptyValueType(valueType string) string {
	valueType = strings.TrimSpace(strings.ToLower(valueType))
	if valueType == "" {
		return "string"
	}
	return valueType
}
