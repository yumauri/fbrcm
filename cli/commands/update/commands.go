package updatecmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	"fbrcm/cli/shared"
	"fbrcm/core"
	"fbrcm/core/filter"
	"fbrcm/core/firebase"
	corelog "fbrcm/core/log"
)

// valueSpec holds value spec state used by the updatecmd package.
type valueSpec struct {
	// value stores value for valueSpec.
	value string
	// valueType stores value type for valueSpec.
	valueType string
}

// updateSpec holds update spec state used by the updatecmd package.
type updateSpec struct {
	// value stores value for updateSpec.
	value *valueSpec
	// name stores name for updateSpec.
	name string
	// group stores group for updateSpec.
	group string
	// description stores description for updateSpec.
	description string
	// nameChanged stores name changed for updateSpec.
	nameChanged bool
	// groupChanged stores group changed for updateSpec.
	groupChanged bool
	// descriptionChanged stores description changed for updateSpec.
	descriptionChanged bool
}

// projectConfig holds project config state used by the updatecmd package.
type projectConfig struct {
	// project stores project for projectConfig.
	project core.Project
	// cache stores cache for projectConfig.
	cache *core.ParametersCache
	// cfg stores cfg for projectConfig.
	cfg *firebase.RemoteConfig
}

// paramTarget holds param target state used by the updatecmd package.
type paramTarget struct {
	// key stores key for paramTarget.
	key string
	// group stores group for paramTarget.
	group string
	// param stores param for paramTarget.
	param firebase.RemoteConfigParam
}

// updateTotals holds update totals state used by the updatecmd package.
type updateTotals struct {
	// modifiedProjects stores modified projects for updateTotals.
	modifiedProjects int
	// updatedParams stores updated params for updateTotals.
	updatedParams int
}

const defaultGroupLabel = "(root)"

// New constructs new and returns the resulting value or error.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [parameter]",
		Short: "Update Remote Config parameters",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectFilter, err := cmd.Flags().GetString("project")
			if err != nil {
				return err
			}
			paramExpr, err := cmd.Flags().GetString("expr")
			if err != nil {
				return err
			}
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			paramFilter, err := cmd.Flags().GetString("filter")
			if err != nil {
				return err
			}
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			groupName, err := cmd.Flags().GetString("group")
			if err != nil {
				return err
			}
			noGroup, err := cmd.Flags().GetBool("no-group")
			if err != nil {
				return err
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}
			value, err := readValueSpec(cmd)
			if err != nil {
				return err
			}
			if len(args) > 0 {
				if strings.TrimSpace(paramFilter) != "" {
					return fmt.Errorf("parameter argument cannot be used together with --filter")
				}
				paramFilter = "=" + args[0]
			}

			groupChanged := cmd.Flags().Changed("group")
			if noGroup {
				groupChanged = true
				groupName = ""
			}
			descriptionChanged := cmd.Flags().Changed("description")
			nameChanged := cmd.Flags().Changed("name")
			groupName = strings.TrimSpace(groupName)
			name = strings.TrimSpace(name)
			if nameChanged && name == "" {
				return fmt.Errorf("--name cannot be empty")
			}

			spec := updateSpec{
				value:              value,
				name:               name,
				group:              groupName,
				description:        description,
				nameChanged:        nameChanged,
				groupChanged:       groupChanged,
				descriptionChanged: descriptionChanged,
			}

			if stdinAvailable(cmd.InOrStdin()) {
				corelog.For("update").Info("stdin mode enabled; using remote config from stdin")
				return runUpdateStdin(cmd, paramFilter, paramExpr, spec)
			}
			return runUpdateRemote(cmd, svc, projectFilter, paramExpr, paramFilter, spec, yes, dryRun)
		},
	}

	cmd.Flags().StringP("project", "p", "", "Filter projects by mode-prefixed query (^, /, ~, =)")
	cmd.Flags().StringP("filter", "f", "", "Filter parameters by mode-prefixed query (^, /, ~, =)")
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	cmd.Flags().Bool("dry-run", false, "Log Firebase write requests without sending them")
	cmd.Flags().BoolP("yes", "y", false, "Print diff and update without confirmation")
	cmd.Flags().String("description", "", "Parameter description")
	cmd.Flags().String("group", "", "Target parameter group")
	cmd.Flags().Bool("no-group", false, "Move parameter out of its group")
	cmd.Flags().String("name", "", "New parameter name")
	cmd.Flags().String("boolean", "", "Boolean parameter value: true or false")
	cmd.Flags().String("number", "", "Number parameter value")
	cmd.Flags().String("string", "", "String parameter value")
	cmd.Flags().String("json", "", "JSON parameter value")
	cmd.MarkFlagsMutuallyExclusive("boolean", "number", "string", "json")
	cmd.MarkFlagsMutuallyExclusive("group", "no-group")
	return cmd
}

// readValueSpec reads read value spec and returns the resulting value or error.
func readValueSpec(cmd *cobra.Command) (*valueSpec, error) {
	// flagSpec holds flag spec state used by the updatecmd package.
	type flagSpec struct {
		// name stores name for flagSpec.
		name string
		// valueType stores value type for flagSpec.
		valueType string
		// validate stores validate for flagSpec.
		validate func(string) error
	}

	specs := []flagSpec{
		{name: "boolean", valueType: "BOOLEAN", validate: func(value string) error {
			if value == "true" || value == "false" {
				return nil
			}
			return fmt.Errorf("--boolean must be true or false")
		}},
		{name: "number", valueType: "NUMBER", validate: func(value string) error {
			if _, err := strconv.ParseFloat(value, 64); err != nil {
				return fmt.Errorf("--number must be valid number")
			}
			return nil
		}},
		{name: "string", valueType: "STRING", validate: func(string) error { return nil }},
		{name: "json", valueType: "JSON", validate: func(value string) error {
			if !json.Valid([]byte(value)) {
				return fmt.Errorf("--json must be valid json")
			}
			return nil
		}},
	}

	var selected []valueSpec
	for _, spec := range specs {
		value, err := cmd.Flags().GetString(spec.name)
		if err != nil {
			return nil, err
		}
		if !cmd.Flags().Changed(spec.name) {
			continue
		}
		if err := spec.validate(value); err != nil {
			return nil, err
		}
		selected = append(selected, valueSpec{value: value, valueType: spec.valueType})
	}
	if len(selected) == 0 {
		return nil, nil
	}
	if len(selected) > 1 {
		return nil, fmt.Errorf("only one of --boolean, --number, --string, or --json may be used")
	}
	return &selected[0], nil
}

// runUpdateRemote runs run update remote and returns the resulting value or error.
func runUpdateRemote(cmd *cobra.Command, svc *core.Core, projectFilter, paramExpr, paramFilter string, spec updateSpec, yes bool, dryRun bool) error {
	ctx := context.Background()
	if dryRun {
		ctx = firebase.WithDryRun(ctx)
	}

	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return err
	}
	projects = filterProjects(projects, projectFilter)
	sortProjects(projects)
	compiledExpr, ok := shared.CompileExpr(paramExpr, "")
	if !ok {
		return nil
	}

	var totals updateTotals
	for _, project := range projects {
		for {
			cfg, err := revalidateProjectConfig(ctx, svc, project)
			if err != nil {
				return err
			}
			matched := collectMatchingParams(project, cfg.cfg, paramFilter, compiledExpr)
			if len(matched) == 0 {
				break
			}
			updated, finalCfg, err := confirmAndUpdateProject(cmd, project.ProjectID, cfg.cfg, matched, spec, yes, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if len(updated) == 0 {
				break
			}

			finalRaw, err := marshalRemoteConfig(finalCfg)
			if err != nil {
				return err
			}
			if err := svc.ValidateRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, cfg.cache.ETag); err != nil {
				if isRemoteConfigConflict(err) {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "remote config changed during update; restarting project %s\n", project.ProjectID)
					continue
				}
				return err
			}
			if _, _, err := svc.PublishRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, cfg.cache.ETag); err != nil {
				if isRemoteConfigConflict(err) {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "remote config changed during update; restarting project %s\n", project.ProjectID)
					continue
				}
				return err
			}

			totals.modifiedProjects++
			totals.updatedParams += len(updated)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✏️ published: %s\n", project.ProjectID)
			break
		}
	}
	logUpdateTotals("remote", totals)
	return nil
}

// runUpdateStdin runs run update stdin and returns the resulting value or error.
func runUpdateStdin(cmd *cobra.Command, paramFilter, paramExpr string, spec updateSpec) error {
	raw, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if !json.Valid(raw) {
		return fmt.Errorf("stdin remote config is not valid json")
	}
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return fmt.Errorf("decode stdin remote config: %w", err)
	}
	cfg = cloneRemoteConfig(cfg)
	compiledExpr, ok := shared.CompileExpr(paramExpr, "<stdin>")
	if !ok {
		return nil
	}

	project := core.Project{Name: "<stdin>", ProjectID: "<stdin>"}
	matched := collectMatchingParams(project, cfg, paramFilter, compiledExpr)
	updated, finalCfg, err := confirmAndUpdateProject(cmd, "<stdin>", cfg, matched, spec, true, cmd.ErrOrStderr())
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(finalCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode remote config: %w", err)
	}
	if _, err := cmd.OutOrStdout().Write(out); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	totals := updateTotals{updatedParams: len(updated)}
	if len(updated) > 0 {
		totals.modifiedProjects = 1
	}
	logUpdateTotals("stdin", totals)
	return nil
}

// confirmAndUpdateProject handles confirm and update project and returns the resulting value or error.
func confirmAndUpdateProject(cmd *cobra.Command, label string, cfg *firebase.RemoteConfig, matched []paramTarget, spec updateSpec, yes bool, diffOut io.Writer) ([]paramTarget, *firebase.RemoteConfig, error) {
	finalCfg := cloneRemoteConfig(cfg)
	updated := make([]paramTarget, 0, len(matched))

	for _, target := range matched {
		nextCfg := cloneRemoteConfig(finalCfg)
		if err := updateParamSlot(nextCfg, target, spec); err != nil {
			return nil, nil, err
		}
		diffText, hasChanges := shared.RenderRemoteConfigDiff(finalCfg, nextCfg)
		if !hasChanges {
			continue
		}
		_, _ = fmt.Fprintln(diffOut, diffText)
		if !yes {
			ok, err := runConfirmationPrompt(
				fmt.Sprintf("Update %s in %s?", formatParameterHeader(target.key, target.group), label),
				cmd.OutOrStdout(),
			)
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				continue
			}
		}
		finalCfg = nextCfg
		updated = append(updated, target)
	}
	if len(updated) == 0 {
		return nil, finalCfg, nil
	}
	return updated, finalCfg, nil
}

// updateParamSlot updates update param slot and returns the resulting value or error.
func updateParamSlot(cfg *firebase.RemoteConfig, target paramTarget, spec updateSpec) error {
	param := target.param
	if spec.value != nil {
		param.DefaultValue = &firebase.RemoteConfigValue{Value: spec.value.value}
		param.ValueType = spec.value.valueType
	}
	if spec.descriptionChanged {
		param.Description = spec.description
	}

	nextGroup := target.group
	if spec.groupChanged {
		nextGroup = spec.group
	}
	nextKey := target.key
	if spec.nameChanged {
		nextKey = spec.name
	}
	if (target.key != nextKey || target.group != nextGroup) && paramSlotExists(cfg, nextKey, nextGroup) {
		return fmt.Errorf("parameter %s already exists", formatParameterHeader(nextKey, nextGroup))
	}
	removeParamSlot(cfg, target.key, target.group)
	setParamSlot(cfg, nextKey, nextGroup, param)
	return nil
}

// setParamSlot sets set param slot and returns the resulting value or error.
func setParamSlot(cfg *firebase.RemoteConfig, key, groupName string, param firebase.RemoteConfigParam) bool {
	if groupName == "" {
		if cfg.Parameters == nil {
			cfg.Parameters = map[string]firebase.RemoteConfigParam{}
		}
		cfg.Parameters[key] = param
		return true
	}
	if cfg.ParameterGroups == nil {
		cfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}
	group := cfg.ParameterGroups[groupName]
	if group.Parameters == nil {
		group.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	group.Parameters[key] = param
	cfg.ParameterGroups[groupName] = group
	return true
}

// paramSlotExists handles param slot exists and returns the resulting value or error.
func paramSlotExists(cfg *firebase.RemoteConfig, key, groupName string) bool {
	if cfg == nil {
		return false
	}
	if groupName == "" {
		_, ok := cfg.Parameters[key]
		return ok
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		return false
	}
	_, ok = group.Parameters[key]
	return ok
}

// removeParamSlot removes remove param slot and returns the resulting value or error.
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

// runConfirmationPrompt runs run confirmation prompt and returns the resulting value or error.
func runConfirmationPrompt(prompt string, fallbackOut io.Writer) (bool, error) {
	confirm := shared.NewConfirmation(prompt, confirmation.Yes, shared.ConfirmationOptions{})
	if fallbackOut != nil {
		confirm.Output = fallbackOut
	}
	return confirm.RunPrompt()
}

// collectMatchingParams handles collect matching params and returns the resulting value or error.
func collectMatchingParams(project core.Project, cfg *firebase.RemoteConfig, rawFilter string, compiledExpr *filter.Expression) []paramTarget {
	all := collectParamTargets(cfg)
	mode, query := parseFilter(rawFilter)

	filtered := make([]paramTarget, 0, len(all))
	for _, target := range all {
		if query != "" {
			match, _ := filter.Match(target.key, query, mode)
			if !match {
				continue
			}
		}
		match, ok := shared.MatchParameterByCompiledExpr(compiledExpr, project, cfg, target.key, target.groupOrDefault())
		if !ok || !match {
			continue
		}
		filtered = append(filtered, target)
	}
	return filtered
}

// collectParamTargets handles collect param targets and returns the resulting value or error.
func collectParamTargets(cfg *firebase.RemoteConfig) []paramTarget {
	if cfg == nil {
		return nil
	}

	out := make([]paramTarget, 0, len(cfg.Parameters)+len(cfg.ParameterGroups))
	for _, key := range sortedStringKeys(cfg.Parameters) {
		out = append(out, paramTarget{key: key, param: cfg.Parameters[key]})
	}
	for _, groupName := range sortedStringKeys(cfg.ParameterGroups) {
		group := cfg.ParameterGroups[groupName]
		for _, key := range sortedStringKeys(group.Parameters) {
			out = append(out, paramTarget{key: key, group: groupName, param: group.Parameters[key]})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !strings.EqualFold(out[i].key, out[j].key) {
			return strings.ToLower(out[i].key) < strings.ToLower(out[j].key)
		}
		return strings.ToLower(out[i].group) < strings.ToLower(out[j].group)
	})
	return out
}

// groupOrDefault handles group or default for paramTarget and returns the resulting state or error.
func (t paramTarget) groupOrDefault() string {
	if strings.TrimSpace(t.group) == "" {
		return defaultGroupLabel
	}
	return t.group
}

// revalidateProjectConfig handles revalidate project config and returns the resulting value or error.
func revalidateProjectConfig(ctx context.Context, svc *core.Core, project core.Project) (*projectConfig, error) {
	cache, _, err := svc.RevalidateParameters(ctx, project.ProjectID)
	if err != nil {
		return nil, err
	}
	cfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, fmt.Errorf("decode remote config for %s: %w", project.ProjectID, err)
	}
	return &projectConfig{project: project, cache: cache, cfg: cloneRemoteConfig(cfg)}, nil
}

// filterProjects filters filter projects and returns the resulting value or error.
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

// parseFilter parses parse filter and returns the resulting value or error.
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

// sortProjects handles sort projects and returns the resulting value or error.
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

// sortedStringKeys handles sorted string keys and returns the resulting value or error.
func sortedStringKeys[T any](items map[string]T) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// formatParameterHeader formats format parameter header and returns the resulting value or error.
func formatParameterHeader(key, group string) string {
	if group == "" {
		return key
	}
	return fmt.Sprintf("%s [%s]", key, group)
}

// cloneRemoteConfig handles clone remote config and returns the resulting value or error.
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

// marshalRemoteConfig handles marshal remote config and returns the resulting value or error.
func marshalRemoteConfig(cfg *firebase.RemoteConfig) ([]byte, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode remote config: %w", err)
	}
	return data, nil
}

// stdinAvailable handles stdin available and returns the resulting value or error.
func stdinAvailable(in io.Reader) bool {
	info, ok := stdinFileInfo(in)
	if !ok {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

// stdinFileInfo handles stdin file info and returns the resulting value or error.
func stdinFileInfo(in io.Reader) (os.FileInfo, bool) {
	file, ok := in.(*os.File)
	if !ok {
		return nil, false
	}
	info, err := file.Stat()
	if err != nil {
		return nil, false
	}
	return info, true
}

// logUpdateTotals handles log update totals and returns the resulting value or error.
func logUpdateTotals(mode string, totals updateTotals) {
	corelog.For("update").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.updatedParams)
}

// isRemoteConfigConflict reports is remote config conflict and returns the resulting value or error.
func isRemoteConfigConflict(err error) bool {
	if err == nil {
		return false
	}

	target := err
	for target != nil {
		msg := strings.ToLower(target.Error())
		if strings.Contains(msg, "returned 412") ||
			strings.Contains(msg, "precondition failed") ||
			strings.Contains(msg, "conditionnotmet") ||
			strings.Contains(msg, "etag") ||
			strings.Contains(msg, "if-match") {
			return true
		}
		target = errors.Unwrap(target)
	}
	return false
}
