package project

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"

	"fbrcm/cli/shared"
	clistyles "fbrcm/cli/styles"
	"fbrcm/core"
	"fbrcm/core/filter"
	"fbrcm/core/firebase"
)

type importOptions struct {
	groups                          []string
	paramFilter                     string
	expr                            string
	removeAllConditions             bool
	removeProjectSpecificConditions bool
	merge                           bool
	override                        bool
	mergeResolve                    string
}

type importStrategy string

const (
	importStrategyMerge    importStrategy = "merge"
	importStrategyOverride importStrategy = "override"
)

type conflictResolution string

const (
	conflictResolutionCurrent conflictResolution = "current"
	conflictResolutionImport  conflictResolution = "import"
)

type mergeChoice struct {
	label string
	value string
}

func (c mergeChoice) String() string {
	return c.label
}

type paramSlot struct {
	group string
	param firebase.RemoteConfigParam
}

type missingImportGroupsError struct {
	missing   []string
	available []groupSummary
}

func (e *missingImportGroupsError) Error() string {
	if len(e.available) > 0 {
		available := make([]string, 0, len(e.available))
		for _, group := range e.available {
			available = append(available, group.Name)
		}
		return fmt.Sprintf("requested groups not found in import: %s; available groups: %s", strings.Join(e.missing, ", "), strings.Join(available, ", "))
	}
	return fmt.Sprintf("requested groups not found in import: %s", strings.Join(e.missing, ", "))
}

func runImportCommand(cmd *cobra.Command, svc *core.Core, project core.Project) error {
	opts, err := readImportOptions(cmd)
	if err != nil {
		return err
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}
	ctx := context.Background()
	if dryRun {
		ctx = firebase.WithDryRun(ctx)
	}

	raw, err := readImportRemoteConfig(cmd)
	if err != nil {
		return err
	}
	if raw == nil {
		return nil
	}
	if !json.Valid(raw) {
		return fmt.Errorf("remote config input is not valid json")
	}

	importCfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return fmt.Errorf("decode remote config: %w", err)
	}
	importCfg = cloneRemoteConfig(importCfg)
	importCfg.Version = firebase.RemoteConfigVersion{}

	if err := transformImportConfig(project, importCfg, opts); err != nil {
		var missingErr *missingImportGroupsError
		if errors.As(err, &missingErr) && len(missingErr.available) > 0 {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), renderGroupsTable(missingErr.available))
		}
		return err
	}

	currentRaw, currentETag, err := svc.ExportRemoteConfig(ctx, project.ProjectID)
	if err != nil {
		return err
	}
	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return fmt.Errorf("decode current remote config: %w", err)
	}
	currentCfg = cloneRemoteConfig(currentCfg)
	currentCfg.Version = firebase.RemoteConfigVersion{}

	finalCfg, err := buildFinalImportConfig(cmd, currentCfg, importCfg, opts)
	if err != nil {
		return err
	}
	finalCfg.Version = firebase.RemoteConfigVersion{}
	pruneUnusedConditions(finalCfg)
	dropUnknownConditionReferences(finalCfg)
	removeEmptyGroups(finalCfg)

	finalRaw, err := marshalRemoteConfigForUpload(finalCfg)
	if err != nil {
		return err
	}

	diffText, hasChanges := shared.RenderRemoteConfigDiff(currentCfg, finalCfg)
	if !hasChanges {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "NO CHANGES")
		return nil
	}

	if err := svc.ValidateRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, currentETag); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)

	confirm := confirmation.New(
		fmt.Sprintf("Publish Remote Config changes to %s?", project.ProjectID),
		confirmation.Yes,
	)
	ok, err := confirm.RunPrompt()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if _, _, err := svc.PublishRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, currentETag); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "published: %s\n", project.ProjectID)
	return nil
}

func readImportOptions(cmd *cobra.Command) (importOptions, error) {
	var opts importOptions
	var err error

	opts.groups, err = cmd.Flags().GetStringArray("group")
	if err != nil {
		return opts, err
	}
	opts.paramFilter, err = cmd.Flags().GetString("filter")
	if err != nil {
		return opts, err
	}
	opts.expr, err = cmd.Flags().GetString("expr")
	if err != nil {
		return opts, err
	}
	opts.removeAllConditions, err = cmd.Flags().GetBool("remove-all-conditions")
	if err != nil {
		return opts, err
	}
	opts.removeProjectSpecificConditions, err = cmd.Flags().GetBool("remove-project-specific-conditions")
	if err != nil {
		return opts, err
	}
	opts.merge, err = cmd.Flags().GetBool("merge")
	if err != nil {
		return opts, err
	}
	opts.override, err = cmd.Flags().GetBool("override")
	if err != nil {
		return opts, err
	}
	opts.mergeResolve, err = cmd.Flags().GetString("merge-resolve")
	if err != nil {
		return opts, err
	}
	opts.mergeResolve = strings.TrimSpace(strings.ToLower(opts.mergeResolve))
	if opts.mergeResolve != "" && opts.mergeResolve != string(conflictResolutionCurrent) && opts.mergeResolve != string(conflictResolutionImport) {
		return opts, fmt.Errorf("invalid --merge-resolve value %q; expected current or import", opts.mergeResolve)
	}
	if opts.mergeResolve != "" && !opts.merge {
		return opts, fmt.Errorf("--merge-resolve requires --merge")
	}

	opts.groups = normalizeGroups(opts.groups)
	opts.expr = strings.TrimSpace(opts.expr)
	opts.paramFilter = strings.TrimSpace(opts.paramFilter)
	return opts, nil
}

func normalizeGroups(groups []string) []string {
	seen := make(map[string]struct{}, len(groups))
	out := make([]string, 0, len(groups))
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if _, ok := seen[group]; ok {
			continue
		}
		seen[group] = struct{}{}
		out = append(out, group)
	}
	return out
}

func transformImportConfig(project core.Project, cfg *firebase.RemoteConfig, opts importOptions) error {
	if len(opts.groups) > 0 {
		if err := filterImportGroups(cfg, opts.groups); err != nil {
			return err
		}
		pruneUnusedConditions(cfg)
	}
	if opts.paramFilter != "" {
		filterImportParameters(cfg, opts.paramFilter)
		pruneUnusedConditions(cfg)
	}
	if opts.expr != "" {
		compiledExpr, ok := shared.CompileExpr(opts.expr, project.ProjectID)
		if !ok {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
			cfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
			cfg.Conditions = nil
			return nil
		}
		filterImportParametersByExpr(project, cfg, compiledExpr)
		pruneUnusedConditions(cfg)
	}

	switch {
	case opts.removeAllConditions:
		removeAllConditions(cfg)
	case opts.removeProjectSpecificConditions:
		removeProjectSpecificConditions(cfg)
	}

	pruneUnusedConditions(cfg)
	dropUnknownConditionReferences(cfg)
	removeEmptyGroups(cfg)
	return nil
}

func filterImportGroups(cfg *firebase.RemoteConfig, groups []string) error {
	selected := make(map[string]firebase.RemoteConfigGroup, len(groups))
	missing := make([]string, 0)
	for _, group := range groups {
		value, ok := cfg.ParameterGroups[group]
		if !ok {
			missing = append(missing, group)
			continue
		}
		selected[group] = value
	}
	if len(missing) > 0 {
		return &missingImportGroupsError{
			missing:   append([]string(nil), missing...),
			available: summarizeGroups(cfg.ParameterGroups),
		}
	}
	cfg.Parameters = nil
	cfg.ParameterGroups = selected
	return nil
}

func filterImportParameters(cfg *firebase.RemoteConfig, raw string) {
	mode, query := parseImportFilter(raw)
	if query == "" {
		return
	}

	cfg.Parameters = filterImportParamMap(cfg.Parameters, mode, query)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = filterImportParamMap(group.Parameters, mode, query)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func filterImportParametersByExpr(project core.Project, cfg *firebase.RemoteConfig, compiledExpr *filter.Expression) {
	if compiledExpr == nil {
		return
	}

	cfg.Parameters = filterImportParamMapByExpr(project, cfg, cfg.Parameters, "(default)", compiledExpr)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = filterImportParamMapByExpr(project, cfg, group.Parameters, groupName, compiledExpr)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func filterImportParamMap(params map[string]firebase.RemoteConfigParam, mode filter.Mode, query string) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}

	filtered := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		match, _ := filter.Match(key, query, mode)
		if match {
			filtered[key] = param
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func filterImportParamMapByExpr(project core.Project, cfg *firebase.RemoteConfig, params map[string]firebase.RemoteConfigParam, groupName string, compiledExpr *filter.Expression) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}

	filtered := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, key, groupName)
		if ok && match {
			filtered[key] = param
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func parseImportFilter(raw string) (filter.Mode, string) {
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

func removeAllConditions(cfg *firebase.RemoteConfig) {
	cfg.Conditions = nil
	cfg.Parameters = stripAllConditionalValues(cfg.Parameters, nil)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = stripAllConditionalValues(group.Parameters, nil)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func removeProjectSpecificConditions(cfg *firebase.RemoteConfig) {
	deleted := make(map[string]struct{})
	kept := make([]firebase.RemoteConfigCondition, 0, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		if isProjectSpecificCondition(condition.Expression) {
			deleted[condition.Name] = struct{}{}
			continue
		}
		kept = append(kept, condition)
	}
	cfg.Conditions = kept
	cfg.Parameters = stripAllConditionalValues(cfg.Parameters, deleted)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = stripAllConditionalValues(group.Parameters, deleted)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func stripAllConditionalValues(params map[string]firebase.RemoteConfigParam, deleted map[string]struct{}) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		if len(param.ConditionalValues) > 0 {
			filtered := make(map[string]firebase.RemoteConfigValue, len(param.ConditionalValues))
			for cond, value := range param.ConditionalValues {
				if deleted == nil {
					continue
				}
				if _, ok := deleted[cond]; ok {
					continue
				}
				filtered[cond] = value
			}
			if len(filtered) > 0 {
				param.ConditionalValues = filtered
			} else {
				param.ConditionalValues = nil
			}
		}
		if param.DefaultValue == nil && len(param.ConditionalValues) == 0 {
			continue
		}
		out[key] = param
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isProjectSpecificCondition(expr string) bool {
	for _, needle := range []string{
		"inExperiment",
		"inUserAudience",
		"app.id",
		"app.userProperty[",
		"app.firebaseInstallationId",
		"app.instanceId",
		"app.instance_id",
	} {
		if strings.Contains(expr, needle) {
			return true
		}
	}
	return false
}

func buildFinalImportConfig(cmd *cobra.Command, currentCfg, importCfg *firebase.RemoteConfig, opts importOptions) (*firebase.RemoteConfig, error) {
	if !configHasContent(currentCfg) {
		return cloneRemoteConfig(importCfg), nil
	}

	strategy, err := chooseImportStrategy(cmd, opts)
	if err != nil {
		return nil, err
	}
	if strategy == importStrategyOverride {
		return cloneRemoteConfig(importCfg), nil
	}

	return mergeRemoteConfigs(cmd, currentCfg, importCfg, opts)
}

func chooseImportStrategy(cmd *cobra.Command, opts importOptions) (importStrategy, error) {
	switch {
	case opts.override:
		return importStrategyOverride, nil
	case opts.merge:
		return importStrategyMerge, nil
	default:
		prompt := selection.New("Current config exists. How to apply import?", []mergeChoice{
			{label: "Merge imported config into current config", value: string(importStrategyMerge)},
			{label: "Override current config with imported config", value: string(importStrategyOverride)},
		})
		choice, err := prompt.RunPrompt()
		if err != nil {
			return "", err
		}
		return importStrategy(choice.value), nil
	}
}

func mergeRemoteConfigs(cmd *cobra.Command, currentCfg, importCfg *firebase.RemoteConfig, opts importOptions) (*firebase.RemoteConfig, error) {
	finalCfg := cloneRemoteConfig(currentCfg)
	if finalCfg.Parameters == nil {
		finalCfg.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	if finalCfg.ParameterGroups == nil {
		finalCfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}

	conditionIndex := make(map[string]int, len(finalCfg.Conditions))
	for i, condition := range finalCfg.Conditions {
		conditionIndex[condition.Name] = i
	}
	for _, condition := range importCfg.Conditions {
		index, ok := conditionIndex[condition.Name]
		if !ok {
			finalCfg.Conditions = append(finalCfg.Conditions, condition)
			conditionIndex[condition.Name] = len(finalCfg.Conditions) - 1
			continue
		}
		if reflect.DeepEqual(finalCfg.Conditions[index], condition) {
			continue
		}
		resolution, err := resolveConflict(cmd, opts, "condition "+condition.Name, finalCfg.Conditions[index], condition)
		if err != nil {
			return nil, err
		}
		if resolution == conflictResolutionImport {
			finalCfg.Conditions[index] = condition
		}
	}

	for _, groupName := range sortedGroupNames(importCfg.ParameterGroups) {
		importGroup := importCfg.ParameterGroups[groupName]
		currentGroup, ok := finalCfg.ParameterGroups[groupName]
		if !ok {
			finalCfg.ParameterGroups[groupName] = firebase.RemoteConfigGroup{
				Description: importGroup.Description,
				Parameters:  map[string]firebase.RemoteConfigParam{},
			}
			continue
		}
		if currentGroup.Description != importGroup.Description {
			resolution, err := resolveConflict(cmd, opts, "group description "+groupName, currentGroup.Description, importGroup.Description)
			if err != nil {
				return nil, err
			}
			if resolution == conflictResolutionImport {
				currentGroup.Description = importGroup.Description
				finalCfg.ParameterGroups[groupName] = currentGroup
			}
		}
	}

	currentSlots := collectParamSlots(finalCfg)
	importSlots := collectParamSlots(importCfg)
	for _, key := range sortedParamKeys(importSlots) {
		importSlot := importSlots[key]
		currentSlot, ok := currentSlots[key]
		if !ok {
			setParamSlot(finalCfg, key, importSlot)
			currentSlots[key] = importSlot
			continue
		}
		if currentSlot.group == importSlot.group && reflect.DeepEqual(currentSlot.param, importSlot.param) {
			continue
		}

		resolution, err := resolveConflict(cmd, opts, "parameter "+key, currentSlot, importSlot)
		if err != nil {
			return nil, err
		}
		if resolution == conflictResolutionImport {
			removeParamSlot(finalCfg, key, currentSlot.group)
			setParamSlot(finalCfg, key, importSlot)
			currentSlots[key] = importSlot
		}
	}

	return finalCfg, nil
}

func resolveConflict(cmd *cobra.Command, opts importOptions, label string, currentValue, importValue any) (conflictResolution, error) {
	if opts.mergeResolve != "" {
		return conflictResolution(opts.mergeResolve), nil
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nConflict: %s\n", label)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), shared.RenderConflictPreview(label, toSharedConflictValue(currentValue), toSharedConflictValue(importValue)))
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())

	prompt := selection.New("Choose value", []mergeChoice{
		{label: "Use import value", value: string(conflictResolutionImport)},
		{label: "Keep current value", value: string(conflictResolutionCurrent)},
	})
	choice, err := prompt.RunPrompt()
	if err != nil {
		return "", err
	}
	return conflictResolution(choice.value), nil
}

func toSharedConflictValue(value any) any {
	slot, ok := value.(paramSlot)
	if !ok {
		return value
	}
	return shared.ParamSlotPreview{
		Group: slot.group,
		Param: slot.param,
	}
}

func collectParamSlots(cfg *firebase.RemoteConfig) map[string]paramSlot {
	out := make(map[string]paramSlot)
	for key, param := range cfg.Parameters {
		out[key] = paramSlot{param: param}
	}
	for groupName, group := range cfg.ParameterGroups {
		for key, param := range group.Parameters {
			out[key] = paramSlot{group: groupName, param: param}
		}
	}
	return out
}

func sortedParamKeys(slots map[string]paramSlot) []string {
	keys := make([]string, 0, len(slots))
	for key := range slots {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedGroupNames(groups map[string]firebase.RemoteConfigGroup) []string {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func setParamSlot(cfg *firebase.RemoteConfig, key string, slot paramSlot) {
	if slot.group == "" {
		if cfg.Parameters == nil {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
		}
		cfg.Parameters[key] = slot.param
		return
	}

	group := cfg.ParameterGroups[slot.group]
	if group.Parameters == nil {
		group.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	group.Parameters[key] = slot.param
	cfg.ParameterGroups[slot.group] = group
}

func removeParamSlot(cfg *firebase.RemoteConfig, key, groupName string) {
	if groupName == "" {
		delete(cfg.Parameters, key)
		return
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return
	}
	delete(group.Parameters, key)
	if len(group.Parameters) == 0 {
		delete(cfg.ParameterGroups, groupName)
		return
	}
	cfg.ParameterGroups[groupName] = group
}

func configHasContent(cfg *firebase.RemoteConfig) bool {
	return cfg != nil && (len(cfg.Conditions) > 0 || len(cfg.Parameters) > 0 || len(cfg.ParameterGroups) > 0)
}

func pruneUnusedConditions(cfg *firebase.RemoteConfig) {
	if cfg == nil || len(cfg.Conditions) == 0 {
		return
	}

	used := make(map[string]struct{})
	collectUsedConditions(used, cfg.Parameters)
	for _, group := range cfg.ParameterGroups {
		collectUsedConditions(used, group.Parameters)
	}

	kept := make([]firebase.RemoteConfigCondition, 0, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		if _, ok := used[condition.Name]; ok {
			kept = append(kept, condition)
		}
	}
	cfg.Conditions = kept
}

func collectUsedConditions(used map[string]struct{}, params map[string]firebase.RemoteConfigParam) {
	for _, param := range params {
		for condition := range param.ConditionalValues {
			used[condition] = struct{}{}
		}
	}
}

func dropUnknownConditionReferences(cfg *firebase.RemoteConfig) {
	allowed := make(map[string]struct{}, len(cfg.Conditions))
	for _, condition := range cfg.Conditions {
		allowed[condition.Name] = struct{}{}
	}
	cfg.Parameters = stripUnknownConditionRefs(cfg.Parameters, allowed)
	for groupName, group := range cfg.ParameterGroups {
		group.Parameters = stripUnknownConditionRefs(group.Parameters, allowed)
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
			continue
		}
		cfg.ParameterGroups[groupName] = group
	}
}

func stripUnknownConditionRefs(params map[string]firebase.RemoteConfigParam, allowed map[string]struct{}) map[string]firebase.RemoteConfigParam {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, param := range params {
		if len(param.ConditionalValues) > 0 {
			filtered := make(map[string]firebase.RemoteConfigValue, len(param.ConditionalValues))
			for cond, value := range param.ConditionalValues {
				if _, ok := allowed[cond]; !ok {
					continue
				}
				filtered[cond] = value
			}
			if len(filtered) > 0 {
				param.ConditionalValues = filtered
			} else {
				param.ConditionalValues = nil
			}
		}
		if param.DefaultValue == nil && len(param.ConditionalValues) == 0 {
			continue
		}
		out[key] = param
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

type groupSummary struct {
	Name       string
	Parameters int
}

func summarizeGroups(groups map[string]firebase.RemoteConfigGroup) []groupSummary {
	names := sortedGroupNames(groups)
	out := make([]groupSummary, 0, len(names))
	for _, name := range names {
		out = append(out, groupSummary{
			Name:       name,
			Parameters: len(groups[name].Parameters),
		})
	}
	return out
}

func renderGroupsTable(groups []groupSummary) string {
	rows := make([][]string, 0, len(groups))
	groupWidth := lipgloss.Width("Group")
	parametersWidth := lipgloss.Width("Parameters")
	for _, group := range groups {
		count := fmt.Sprintf("%d", group.Parameters)
		rows = append(rows, []string{group.Name, count})
		groupWidth = max(groupWidth, lipgloss.Width(group.Name))
		parametersWidth = max(parametersWidth, lipgloss.Width(count))
	}

	styleFunc := func(row, col int) lipgloss.Style {
		_ = col
		style := lipgloss.NewStyle().Padding(0, 1)
		if clistyles.NoColorEnabled() {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateBright)
	}

	tbl := table.New().
		Headers("Group", "Parameters").
		Rows(rows...).
		Width(groupWidth + parametersWidth + 7).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func removeEmptyGroups(cfg *firebase.RemoteConfig) {
	for groupName, group := range cfg.ParameterGroups {
		if len(group.Parameters) == 0 {
			delete(cfg.ParameterGroups, groupName)
		}
	}
	if len(cfg.ParameterGroups) == 0 {
		cfg.ParameterGroups = nil
	}
	if len(cfg.Parameters) == 0 {
		cfg.Parameters = nil
	}
}

func cloneRemoteConfig(cfg *firebase.RemoteConfig) *firebase.RemoteConfig {
	if cfg == nil {
		return &firebase.RemoteConfig{}
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return &firebase.RemoteConfig{}
	}
	var out firebase.RemoteConfig
	if err := json.Unmarshal(data, &out); err != nil {
		return &firebase.RemoteConfig{}
	}
	return &out
}

func marshalRemoteConfigForUpload(cfg *firebase.RemoteConfig) ([]byte, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode remote config: %w", err)
	}
	return data, nil
}
