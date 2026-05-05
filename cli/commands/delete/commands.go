package deletecmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	"fbrcm/cli/shared"
	"fbrcm/core"
	"fbrcm/core/filter"
	"fbrcm/core/firebase"
	corelog "fbrcm/core/log"
)

type projectConfig struct {
	project core.Project
	cache   *core.ParametersCache
	cfg     *firebase.RemoteConfig
}

type paramTarget struct {
	key   string
	group string
	param firebase.RemoteConfigParam
}

type deleteTotals struct {
	modifiedProjects int
	deletedParams    int
}

type remoteConfigOrder struct {
	topLevel          []string
	parameters        []string
	groups            []string
	groupParameters   map[string][]string
	conditionalValues map[string][]string
	versionRaw        []byte
}

const defaultDeleteGroupLabel = "(default)"

func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [parameter]",
		Short: "Delete Remote Config parameters",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectFilter, err := cmd.Flags().GetString("project")
			if err != nil {
				return err
			}
			projectExpr, err := cmd.Flags().GetString("expr")
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
			if len(args) > 0 {
				if strings.TrimSpace(paramFilter) != "" {
					return fmt.Errorf("parameter argument cannot be used together with --filter")
				}
				paramFilter = "=" + args[0]
			}

			if stdinAvailable(cmd.InOrStdin()) {
				corelog.For("delete").Info("stdin mode enabled; using remote config from stdin")
				return runDeleteStdin(cmd, paramFilter, projectExpr)
			}
			return runDeleteRemote(cmd, svc, projectFilter, projectExpr, paramFilter, yes, dryRun)
		},
	}

	cmd.Flags().StringP("project", "p", "", "Filter projects by mode-prefixed query (^, /, ~, =)")
	cmd.Flags().StringP("filter", "f", "", "Filter parameters by mode-prefixed query (^, /, ~, =)")
	cmd.Flags().String("expr", "", "Filter parameters by expr-lang expression")
	cmd.Flags().Bool("dry-run", false, "Log Firebase write requests without sending them")
	cmd.Flags().BoolP("yes", "y", false, "Print diff and delete without confirmation")
	return cmd
}

func runDeleteRemote(cmd *cobra.Command, svc *core.Core, projectFilter, projectExpr, paramFilter string, yes bool, dryRun bool) error {
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
	compiledExpr, ok := shared.CompileExpr(projectExpr, "")
	if !ok {
		return nil
	}

	var totals deleteTotals
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

			deleted, finalCfg, err := confirmAndDeleteProject(cmd, project.ProjectID, cfg.cfg, matched, yes, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if len(deleted) == 0 {
				break
			}

			finalRaw, err := marshalRemoteConfig(finalCfg)
			if err != nil {
				return err
			}
			if err := svc.ValidateRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, cfg.cache.ETag); err != nil {
				if isRemoteConfigConflict(err) {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "remote config changed during delete; restarting project %s\n", project.ProjectID)
					continue
				}
				return err
			}
			if _, _, err := svc.PublishRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, cfg.cache.ETag); err != nil {
				if isRemoteConfigConflict(err) {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "remote config changed during delete; restarting project %s\n", project.ProjectID)
					continue
				}
				return err
			}

			totals.modifiedProjects++
			totals.deletedParams += len(deleted)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "published: %s\n", project.ProjectID)
			break
		}
	}

	logDeleteTotals("remote", totals)
	return nil
}

func runDeleteStdin(cmd *cobra.Command, paramFilter, projectExpr string) error {
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
	compiledExpr, ok := shared.CompileExpr(projectExpr, "<stdin>")
	if !ok {
		return nil
	}

	order, err := parseRemoteConfigOrder(raw)
	if err != nil {
		return fmt.Errorf("parse stdin remote config order: %w", err)
	}

	project := core.Project{Name: "<stdin>", ProjectID: "<stdin>"}
	matched := collectMatchingParams(project, cfg, paramFilter, compiledExpr)
	deleted, finalCfg, err := confirmAndDeleteProject(cmd, "<stdin>", cfg, matched, true, cmd.ErrOrStderr())
	if err != nil {
		return err
	}

	out, err := marshalPrettyRemoteConfigWithOrder(finalCfg, order)
	if err != nil {
		return err
	}
	if _, err := cmd.OutOrStdout().Write(out); err != nil {
		return err
	}
	if len(out) == 0 || out[len(out)-1] != '\n' {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	totals := deleteTotals{deletedParams: len(deleted)}
	if len(deleted) > 0 {
		totals.modifiedProjects = 1
	}
	logDeleteTotals("stdin", totals)
	return nil
}

func revalidateProjectConfig(ctx context.Context, svc *core.Core, project core.Project) (*projectConfig, error) {
	cache, _, err := svc.RevalidateParameters(ctx, project.ProjectID)
	if err != nil {
		return nil, err
	}
	cfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, fmt.Errorf("decode remote config for %s: %w", project.ProjectID, err)
	}
	return &projectConfig{
		project: project,
		cache:   cache,
		cfg:     cloneRemoteConfig(cfg),
	}, nil
}

func confirmAndDeleteProject(cmd *cobra.Command, label string, cfg *firebase.RemoteConfig, matched []paramTarget, yes bool, diffOut io.Writer) ([]paramTarget, *firebase.RemoteConfig, error) {
	finalCfg := cloneRemoteConfig(cfg)
	deleted := make([]paramTarget, 0, len(matched))

	for _, target := range matched {
		diffText := renderDeletedParameter(target)
		_, _ = fmt.Fprintln(diffOut, diffText)

		if !yes {
			ok, err := runConfirmationPrompt(
				fmt.Sprintf("Delete %s from %s?", formatParameterHeader(target.key, target.group), label),
				cmd.OutOrStdout(),
			)
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				continue
			}
		}

		removeParamSlot(finalCfg, target.key, target.group)
		deleted = append(deleted, target)
	}

	if len(deleted) == 0 {
		return nil, finalCfg, nil
	}

	return deleted, finalCfg, nil
}

func runConfirmationPrompt(prompt string, fallbackOut io.Writer) (bool, error) {
	confirm := confirmation.New(prompt, confirmation.Yes)
	if fallbackOut != nil {
		confirm.Output = fallbackOut
	}
	return confirm.RunPrompt()
}

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

func (t paramTarget) groupOrDefault() string {
	if strings.TrimSpace(t.group) == "" {
		return defaultDeleteGroupLabel
	}
	return t.group
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

func marshalRemoteConfig(cfg *firebase.RemoteConfig) ([]byte, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode remote config: %w", err)
	}
	return data, nil
}

func marshalPrettyRemoteConfigWithOrder(cfg *firebase.RemoteConfig, order remoteConfigOrder) ([]byte, error) {
	if cfg == nil {
		return []byte("{}\n"), nil
	}

	var buf bytes.Buffer
	writeRemoteConfigObject(&buf, cfg, order, 0)
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func renderDeletedParameter(target paramTarget) string {
	lines := []string{fmt.Sprintf("  - %s", colorRemoved(formatParameterHeader(target.key, target.group)))}
	if strings.TrimSpace(target.param.ValueType) != "" {
		lines = append(lines, fmt.Sprintf("      - type:                %s", colorRemoved(target.param.ValueType)))
	}
	if strings.TrimSpace(target.param.Description) != "" {
		lines = append(lines, fmt.Sprintf("      - description:         %s", colorRemoved(formatPlainValue(target.param.Description))))
	}
	if target.param.DefaultValue != nil {
		lines = append(lines, fmt.Sprintf("      - default:             %s", colorRemoved(formatRemoteValue(*target.param.DefaultValue))))
	}
	for _, condition := range sortedConditionalNames(target.param.ConditionalValues) {
		lines = append(lines, fmt.Sprintf("      - cond %-15s %s", condition+":", colorRemoved(formatRemoteValue(target.param.ConditionalValues[condition]))))
	}
	return strings.Join(lines, "\n")
}

func sortedConditionalNames(values map[string]firebase.RemoteConfigValue) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringKeys[T any](items map[string]T) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func formatParameterHeader(key, group string) string {
	if group == "" {
		return key
	}
	return fmt.Sprintf("%s [%s]", key, group)
}

func formatRemoteValue(value firebase.RemoteConfigValue) string {
	switch {
	case len(value.PersonalizationValue) > 0:
		return string(normalizeExportJSON(bytes.TrimSpace(value.PersonalizationValue)))
	case len(value.RolloutValue) > 0:
		return string(normalizeExportJSON(bytes.TrimSpace(value.RolloutValue)))
	case value.UseInAppDefault:
		return "useInAppDefault"
	default:
		return formatPlainValue(value.Value)
	}
}

func formatPlainValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "(empty)"
	}
	quoted, err := json.Marshal(value)
	if err != nil {
		return value
	}
	if isSimpleToken(value) {
		return value
	}
	return string(quoted)
}

func isSimpleToken(value string) bool {
	for _, r := range value {
		if r == ' ' || r == '\t' || r == '\n' || r == '"' {
			return false
		}
	}
	return true
}

func colorRemoved(value string) string {
	if value == "" {
		return value
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(value)
}

func stdinAvailable(in io.Reader) bool {
	info, ok := stdinFileInfo(in)
	if !ok {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

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

func normalizeExportJSON(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte(`\u003c`), []byte("<"))
	body = bytes.ReplaceAll(body, []byte(`\u003e`), []byte(">"))
	body = bytes.ReplaceAll(body, []byte(`\u0026`), []byte("&"))
	return body
}

func logDeleteTotals(mode string, totals deleteTotals) {
	corelog.For("delete").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.deletedParams)
}

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

type objectEntry struct {
	key        string
	writeValue func()
}

type orderedJSONNode struct {
	kind    byte
	members []orderedJSONMember
	items   []*orderedJSONNode
	raw     []byte
}

type orderedJSONMember struct {
	key   string
	value *orderedJSONNode
}

func writeRemoteConfigObject(buf *bytes.Buffer, cfg *firebase.RemoteConfig, order remoteConfigOrder, indent int) {
	fields := map[string]objectEntry{
		"conditions": {
			key: "conditions",
			writeValue: func() {
				writeConditions(buf, cfg.Conditions, indent+1)
			},
		},
		"parameters": {
			key: "parameters",
			writeValue: func() {
				writeParametersMap(buf, cfg.Parameters, order.parameters, order.conditionalValues, "", indent+1)
			},
		},
		"parameterGroups": {
			key: "parameterGroups",
			writeValue: func() {
				writeGroups(buf, cfg.ParameterGroups, order, indent+1)
			},
		},
		"version": {
			key: "version",
			writeValue: func() {
				if len(order.versionRaw) > 0 {
					buf.Write(order.versionRaw)
					return
				}
				writeVersion(buf, cfg.Version, indent+1)
			},
		},
	}

	entries := make([]objectEntry, 0, 4)
	seen := make(map[string]struct{}, 4)
	for _, key := range order.topLevel {
		if !remoteConfigFieldPresent(cfg, key) {
			continue
		}
		entry, ok := fields[key]
		if !ok {
			continue
		}
		entries = append(entries, entry)
		seen[key] = struct{}{}
	}
	for _, key := range []string{"conditions", "parameters", "parameterGroups", "version"} {
		if _, ok := seen[key]; ok || !remoteConfigFieldPresent(cfg, key) {
			continue
		}
		entries = append(entries, fields[key])
	}
	writeObject(buf, indent, entries)
}

func remoteConfigFieldPresent(cfg *firebase.RemoteConfig, key string) bool {
	switch key {
	case "conditions":
		return len(cfg.Conditions) > 0
	case "parameters":
		return len(cfg.Parameters) > 0
	case "parameterGroups":
		return len(cfg.ParameterGroups) > 0
	case "version":
		return strings.TrimSpace(cfg.Version.VersionNumber) != "" ||
			strings.TrimSpace(cfg.Version.UpdateTime) != "" ||
			strings.TrimSpace(cfg.Version.Description) != ""
	default:
		return false
	}
}

func writeObject(buf *bytes.Buffer, indent int, entries []objectEntry) {
	buf.WriteByte('{')
	if len(entries) == 0 {
		buf.WriteByte('}')
		return
	}
	for i, entry := range entries {
		buf.WriteByte('\n')
		writeIndent(buf, indent+1)
		writeJSONString(buf, entry.key)
		buf.WriteString(": ")
		entry.writeValue()
		if i < len(entries)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('\n')
	writeIndent(buf, indent)
	buf.WriteByte('}')
}

func writeConditions(buf *bytes.Buffer, conditions []firebase.RemoteConfigCondition, indent int) {
	buf.WriteByte('[')
	if len(conditions) == 0 {
		buf.WriteByte(']')
		return
	}
	for i, condition := range conditions {
		buf.WriteByte('\n')
		writeIndent(buf, indent+1)
		writeCondition(buf, condition, indent+1)
		if i < len(conditions)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('\n')
	writeIndent(buf, indent)
	buf.WriteByte(']')
}

func writeCondition(buf *bytes.Buffer, condition firebase.RemoteConfigCondition, indent int) {
	entries := make([]objectEntry, 0, 4)
	if condition.Name != "" {
		value := condition.Name
		entries = append(entries, objectEntry{key: "name", writeValue: func() { writeJSONString(buf, value) }})
	}
	if condition.Expression != "" {
		value := condition.Expression
		entries = append(entries, objectEntry{key: "expression", writeValue: func() { writeJSONString(buf, value) }})
	}
	if condition.Description != "" {
		value := condition.Description
		entries = append(entries, objectEntry{key: "description", writeValue: func() { writeJSONString(buf, value) }})
	}
	if condition.TagColor != "" {
		value := condition.TagColor
		entries = append(entries, objectEntry{key: "tagColor", writeValue: func() { writeJSONString(buf, value) }})
	}
	writeObject(buf, indent, entries)
}

func writeGroups(buf *bytes.Buffer, groups map[string]firebase.RemoteConfigGroup, order remoteConfigOrder, indent int) {
	keys := orderedKeys(groups, order.groups)
	entries := make([]objectEntry, 0, len(keys))
	for _, key := range keys {
		groupName := key
		group := groups[key]
		entries = append(entries, objectEntry{
			key: groupName,
			writeValue: func() {
				writeGroup(buf, groupName, group, order, indent+1)
			},
		})
	}
	writeObject(buf, indent, entries)
}

func writeGroup(buf *bytes.Buffer, groupName string, group firebase.RemoteConfigGroup, order remoteConfigOrder, indent int) {
	entries := make([]objectEntry, 0, 2)
	if group.Description != "" {
		value := group.Description
		entries = append(entries, objectEntry{key: "description", writeValue: func() { writeJSONString(buf, value) }})
	}
	if len(group.Parameters) > 0 {
		params := group.Parameters
		paramOrder := order.groupParameters[groupName]
		entries = append(entries, objectEntry{
			key: "parameters",
			writeValue: func() {
				writeParametersMap(buf, params, paramOrder, order.conditionalValues, groupName, indent+1)
			},
		})
	}
	writeObject(buf, indent, entries)
}

func writeParametersMap(buf *bytes.Buffer, params map[string]firebase.RemoteConfigParam, order []string, conditionalOrders map[string][]string, groupName string, indent int) {
	keys := orderedKeys(params, order)
	entries := make([]objectEntry, 0, len(keys))
	for _, key := range keys {
		paramKey := key
		param := params[key]
		condOrder := conditionalOrders[orderPath(groupName, paramKey)]
		entries = append(entries, objectEntry{
			key: paramKey,
			writeValue: func() {
				writeParam(buf, param, condOrder, indent+1)
			},
		})
	}
	writeObject(buf, indent, entries)
}

func writeParam(buf *bytes.Buffer, param firebase.RemoteConfigParam, conditionalOrder []string, indent int) {
	entries := make([]objectEntry, 0, 4)
	if param.DefaultValue != nil {
		value := *param.DefaultValue
		entries = append(entries, objectEntry{
			key: "defaultValue",
			writeValue: func() {
				writeRemoteConfigValue(buf, value, indent+1)
			},
		})
	}
	if len(param.ConditionalValues) > 0 {
		values := param.ConditionalValues
		entries = append(entries, objectEntry{
			key: "conditionalValues",
			writeValue: func() {
				writeConditionalValues(buf, values, conditionalOrder, indent+1)
			},
		})
	}
	if param.Description != "" {
		value := param.Description
		entries = append(entries, objectEntry{key: "description", writeValue: func() { writeJSONString(buf, value) }})
	}
	if param.ValueType != "" {
		value := param.ValueType
		entries = append(entries, objectEntry{key: "valueType", writeValue: func() { writeJSONString(buf, value) }})
	}
	writeObject(buf, indent, entries)
}

func writeConditionalValues(buf *bytes.Buffer, values map[string]firebase.RemoteConfigValue, order []string, indent int) {
	keys := orderedKeys(values, order)
	entries := make([]objectEntry, 0, len(keys))
	for _, key := range keys {
		condition := key
		value := values[key]
		entries = append(entries, objectEntry{
			key: condition,
			writeValue: func() {
				writeRemoteConfigValue(buf, value, indent+1)
			},
		})
	}
	writeObject(buf, indent, entries)
}

func writeRemoteConfigValue(buf *bytes.Buffer, value firebase.RemoteConfigValue, indent int) {
	entries := make([]objectEntry, 0, 4)
	if value.Value != "" || (!value.UseInAppDefault && len(value.PersonalizationValue) == 0 && len(value.RolloutValue) == 0) {
		raw := value.Value
		entries = append(entries, objectEntry{key: "value", writeValue: func() { writeJSONString(buf, raw) }})
	}
	if value.UseInAppDefault {
		entries = append(entries, objectEntry{key: "useInAppDefault", writeValue: func() { buf.WriteString("true") }})
	}
	if len(value.PersonalizationValue) > 0 {
		raw := append([]byte(nil), value.PersonalizationValue...)
		entries = append(entries, objectEntry{key: "personalizationValue", writeValue: func() { buf.Write(normalizeExportJSON(bytes.TrimSpace(raw))) }})
	}
	if len(value.RolloutValue) > 0 {
		raw := append([]byte(nil), value.RolloutValue...)
		entries = append(entries, objectEntry{key: "rolloutValue", writeValue: func() { buf.Write(normalizeExportJSON(bytes.TrimSpace(raw))) }})
	}
	writeObject(buf, indent, entries)
}

func writeVersion(buf *bytes.Buffer, version firebase.RemoteConfigVersion, indent int) {
	entries := make([]objectEntry, 0, 3)
	if version.VersionNumber != "" {
		value := version.VersionNumber
		entries = append(entries, objectEntry{key: "versionNumber", writeValue: func() { writeJSONString(buf, value) }})
	}
	if version.UpdateTime != "" {
		value := version.UpdateTime
		entries = append(entries, objectEntry{key: "updateTime", writeValue: func() { writeJSONString(buf, value) }})
	}
	if version.Description != "" {
		value := version.Description
		entries = append(entries, objectEntry{key: "description", writeValue: func() { writeJSONString(buf, value) }})
	}
	writeObject(buf, indent, entries)
}

func writeIndent(buf *bytes.Buffer, indent int) {
	for range indent {
		buf.WriteString("  ")
	}
}

func writeJSONString(buf *bytes.Buffer, value string) {
	encoded, _ := json.Marshal(value)
	buf.Write(normalizeExportJSON(encoded))
}

func orderedKeys[T any](items map[string]T, preferred []string) []string {
	keys := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, key := range preferred {
		if _, ok := items[key]; !ok {
			continue
		}
		keys = append(keys, key)
		seen[key] = struct{}{}
	}
	rest := make([]string, 0, len(items)-len(keys))
	for key := range items {
		if _, ok := seen[key]; ok {
			continue
		}
		rest = append(rest, key)
	}
	sort.Strings(rest)
	return append(keys, rest...)
}

func parseRemoteConfigOrder(raw []byte) (remoteConfigOrder, error) {
	body := bytes.TrimSpace(raw)
	root, next, ok := parseOrderedJSONValue(body, 0)
	if !ok || next != len(body) {
		return remoteConfigOrder{}, fmt.Errorf("invalid json")
	}
	if root == nil || root.kind != '{' {
		return remoteConfigOrder{}, fmt.Errorf("root json value is not object")
	}

	order := remoteConfigOrder{
		groupParameters:   make(map[string][]string),
		conditionalValues: make(map[string][]string),
	}
	for _, member := range root.members {
		order.topLevel = append(order.topLevel, member.key)
		switch member.key {
		case "parameters":
			order.parameters = objectMemberOrder(member.value)
			collectConditionalValueOrders(member.value, "", order.conditionalValues)
		case "parameterGroups":
			order.groups = objectMemberOrder(member.value)
			for _, groupMember := range member.value.members {
				for _, field := range groupMember.value.members {
					if field.key != "parameters" {
						continue
					}
					order.groupParameters[groupMember.key] = objectMemberOrder(field.value)
					collectConditionalValueOrders(field.value, groupMember.key, order.conditionalValues)
				}
			}
		case "version":
			order.versionRaw = append([]byte(nil), member.value.raw...)
		}
	}
	return order, nil
}

func objectMemberOrder(node *orderedJSONNode) []string {
	if node == nil || node.kind != '{' {
		return nil
	}
	keys := make([]string, 0, len(node.members))
	for _, member := range node.members {
		keys = append(keys, member.key)
	}
	return keys
}

func collectConditionalValueOrders(node *orderedJSONNode, groupName string, out map[string][]string) {
	if node == nil || node.kind != '{' {
		return
	}
	for _, paramMember := range node.members {
		for _, field := range paramMember.value.members {
			if field.key != "conditionalValues" {
				continue
			}
			out[orderPath(groupName, paramMember.key)] = objectMemberOrder(field.value)
		}
	}
}

func orderPath(groupName, paramKey string) string {
	if groupName == "" {
		return paramKey
	}
	return groupName + "\x00" + paramKey
}

func parseOrderedJSONValue(body []byte, start int) (*orderedJSONNode, int, bool) {
	start = skipJSONWhitespace(body, start)
	if start >= len(body) {
		return nil, 0, false
	}
	switch body[start] {
	case '{':
		return parseOrderedJSONObject(body, start)
	case '[':
		return parseOrderedJSONArray(body, start)
	case '"':
		end, ok := scanJSONStringEnd(body, start)
		if !ok {
			return nil, 0, false
		}
		return &orderedJSONNode{kind: '"', raw: append([]byte(nil), body[start:end+1]...)}, end + 1, true
	default:
		end, ok := scanPrimitiveEnd(body, start)
		if !ok {
			return nil, 0, false
		}
		return &orderedJSONNode{kind: 'v', raw: append([]byte(nil), body[start:end]...)}, end, true
	}
}

func parseOrderedJSONObject(body []byte, start int) (*orderedJSONNode, int, bool) {
	if start >= len(body) || body[start] != '{' {
		return nil, 0, false
	}
	node := &orderedJSONNode{kind: '{'}
	pos := skipJSONWhitespace(body, start+1)
	if pos < len(body) && body[pos] == '}' {
		node.raw = append([]byte(nil), body[start:pos+1]...)
		return node, pos + 1, true
	}
	for {
		keyStart := skipJSONWhitespace(body, pos)
		keyEnd, ok := scanJSONStringEnd(body, keyStart)
		if !ok {
			return nil, 0, false
		}
		key, err := unquoteJSONString(body[keyStart : keyEnd+1])
		if err != nil {
			return nil, 0, false
		}
		colon := skipJSONWhitespace(body, keyEnd+1)
		if colon >= len(body) || body[colon] != ':' {
			return nil, 0, false
		}
		value, next, ok := parseOrderedJSONValue(body, colon+1)
		if !ok {
			return nil, 0, false
		}
		node.members = append(node.members, orderedJSONMember{key: key, value: value})
		pos = skipJSONWhitespace(body, next)
		if pos >= len(body) {
			return nil, 0, false
		}
		if body[pos] == '}' {
			node.raw = append([]byte(nil), body[start:pos+1]...)
			return node, pos + 1, true
		}
		if body[pos] != ',' {
			return nil, 0, false
		}
		pos++
	}
}

func parseOrderedJSONArray(body []byte, start int) (*orderedJSONNode, int, bool) {
	if start >= len(body) || body[start] != '[' {
		return nil, 0, false
	}
	node := &orderedJSONNode{kind: '['}
	pos := skipJSONWhitespace(body, start+1)
	if pos < len(body) && body[pos] == ']' {
		node.raw = append([]byte(nil), body[start:pos+1]...)
		return node, pos + 1, true
	}
	for {
		value, next, ok := parseOrderedJSONValue(body, pos)
		if !ok {
			return nil, 0, false
		}
		node.items = append(node.items, value)
		pos = skipJSONWhitespace(body, next)
		if pos >= len(body) {
			return nil, 0, false
		}
		if body[pos] == ']' {
			node.raw = append([]byte(nil), body[start:pos+1]...)
			return node, pos + 1, true
		}
		if body[pos] != ',' {
			return nil, 0, false
		}
		pos++
	}
}

func skipJSONWhitespace(body []byte, pos int) int {
	for pos < len(body) {
		switch body[pos] {
		case ' ', '\n', '\r', '\t':
			pos++
		default:
			return pos
		}
	}
	return pos
}

func scanJSONStringEnd(body []byte, start int) (int, bool) {
	if start >= len(body) || body[start] != '"' {
		return 0, false
	}
	escaped := false
	for i := start + 1; i < len(body); i++ {
		switch {
		case escaped:
			escaped = false
		case body[i] == '\\':
			escaped = true
		case body[i] == '"':
			return i, true
		}
	}
	return 0, false
}

func scanPrimitiveEnd(body []byte, start int) (int, bool) {
	for i := start; i < len(body); i++ {
		switch body[i] {
		case ',', '}', ']', ' ', '\n', '\r', '\t':
			return i, true
		}
	}
	return len(body), true
}

func unquoteJSONString(raw []byte) (string, error) {
	var out string
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return out, nil
}
