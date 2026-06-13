package addcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// addValueSpec holds add value spec state used by the addcmd package.
type addValueSpec struct {
	// value stores value for addValueSpec.
	value string
	// valueType stores value type for addValueSpec.
	valueType string
}

// addTotals holds add totals state used by the addcmd package.
type addTotals struct {
	// modifiedProjects stores modified projects for addTotals.
	modifiedProjects int
	// addedParams stores added params for addTotals.
	addedParams int
}

// projectConfig holds project config state used by the addcmd package.
type projectConfig struct {
	// project stores project for projectConfig.
	project core.Project
	// cache stores cache for projectConfig.
	cache *core.ParametersCache
	// cfg stores cfg for projectConfig.
	cfg *firebase.RemoteConfig
}

// remoteConfigOrder holds remote config order state used by the addcmd package.
type remoteConfigOrder struct {
	// topLevel stores top level for remoteConfigOrder.
	topLevel []string
	// parameters stores parameters for remoteConfigOrder.
	parameters []string
	// groups stores groups for remoteConfigOrder.
	groups []string
	// groupParameters stores group parameters for remoteConfigOrder.
	groupParameters map[string][]string
	// conditionalValues stores conditional values for remoteConfigOrder.
	conditionalValues map[string][]string
	// versionRaw stores version raw for remoteConfigOrder.
	versionRaw []byte
}

// objectEntry holds object entry state used by the addcmd package.
type objectEntry struct {
	// key stores key for objectEntry.
	key string
	// writeValue stores write value for objectEntry.
	writeValue func()
}

// orderedJSONNode holds ordered jsonnode state used by the addcmd package.
type orderedJSONNode struct {
	// kind stores kind for orderedJSONNode.
	kind byte
	// members stores members for orderedJSONNode.
	members []orderedJSONMember
	// items stores items for orderedJSONNode.
	items []*orderedJSONNode
	// raw stores raw for orderedJSONNode.
	raw []byte
}

// orderedJSONMember holds ordered jsonmember state used by the addcmd package.
type orderedJSONMember struct {
	// key stores key for orderedJSONMember.
	key string
	// value stores value for orderedJSONMember.
	value *orderedJSONNode
}

// New constructs new and returns the resulting value or error.
func New(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <parameter>",
		Short: "Add Remote Config parameter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectFilters, err := cmd.Flags().GetStringArray("project")
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
			groupName, err := cmd.Flags().GetString("group")
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}
			spec, err := readAddValueSpec(cmd)
			if err != nil {
				return err
			}

			key := strings.TrimSpace(args[0])
			if key == "" {
				return fmt.Errorf("parameter key cannot be empty")
			}
			groupName = strings.TrimSpace(groupName)

			if stdinAvailable(cmd.InOrStdin()) {
				corelog.For("add").Info("stdin mode enabled; using remote config from stdin")
				return runAddStdin(cmd, key, groupName, description, spec, projectExpr)
			}
			return runAddRemote(cmd, svc, key, projectFilters, projectExpr, groupName, description, spec, dryRun)
		},
	}

	cmd.Flags().StringArrayP("project", "p", nil, "Filter projects by mode-prefixed query (^, /, ~, =); may be repeated")
	cmd.Flags().String("expr", "", "Filter projects by expr-lang expression")
	cmd.Flags().Bool("dry-run", false, "Log Firebase write requests without sending them")
	cmd.Flags().String("description", "", "Parameter description")
	cmd.Flags().String("group", "", "Target parameter group")
	cmd.Flags().String("boolean", "", "Boolean parameter value: true or false")
	cmd.Flags().String("number", "", "Number parameter value")
	cmd.Flags().String("string", "", "String parameter value")
	cmd.Flags().String("json", "", "JSON parameter value")
	cmd.MarkFlagsMutuallyExclusive("boolean", "number", "string", "json")
	return cmd
}

// readAddValueSpec reads read add value spec and returns the resulting value or error.
func readAddValueSpec(cmd *cobra.Command) (addValueSpec, error) {
	// flagSpec holds flag spec state used by the addcmd package.
	type flagSpec struct {
		// name stores name for flagSpec.
		name string
		// valueType stores value type for flagSpec.
		valueType string
		// validate stores validate for flagSpec.
		validate func(string) error
	}

	specs := []flagSpec{
		{
			name:      "boolean",
			valueType: "BOOLEAN",
			validate: func(value string) error {
				switch value {
				case "true", "false":
					return nil
				default:
					return fmt.Errorf("--boolean must be true or false")
				}
			},
		},
		{
			name:      "number",
			valueType: "NUMBER",
			validate: func(value string) error {
				if _, err := strconv.ParseFloat(value, 64); err != nil {
					return fmt.Errorf("--number must be valid number")
				}
				return nil
			},
		},
		{
			name:      "string",
			valueType: "STRING",
			validate: func(string) error {
				return nil
			},
		},
		{
			name:      "json",
			valueType: "JSON",
			validate: func(value string) error {
				if !json.Valid([]byte(value)) {
					return fmt.Errorf("--json must be valid json")
				}
				return nil
			},
		},
	}

	var selected []addValueSpec
	for _, spec := range specs {
		value, err := cmd.Flags().GetString(spec.name)
		if err != nil {
			return addValueSpec{}, err
		}
		if !cmd.Flags().Changed(spec.name) {
			continue
		}
		if err := spec.validate(value); err != nil {
			return addValueSpec{}, err
		}
		selected = append(selected, addValueSpec{value: value, valueType: spec.valueType})
	}

	if len(selected) == 0 {
		return addValueSpec{}, fmt.Errorf("exactly one of --boolean, --number, --string, or --json is required")
	}
	if len(selected) > 1 {
		return addValueSpec{}, fmt.Errorf("only one of --boolean, --number, --string, or --json may be used")
	}
	return selected[0], nil
}

// runAddRemote runs run add remote and returns the resulting value or error.
func runAddRemote(cmd *cobra.Command, svc *core.Core, key string, projectFilters []string, projectExpr, groupName, description string, spec addValueSpec, dryRun bool) error {
	ctx := context.Background()
	if dryRun {
		ctx = firebase.WithDryRun(ctx)
	}

	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return err
	}
	projects = shared.FilterProjects(projects, projectFilters)
	projects, err = shared.FilterProjectsByExpr(ctx, svc, projects, projectExpr)
	if err != nil {
		return err
	}
	sortProjects(projects)

	var totals addTotals
	for _, project := range projects {
		for {
			cfg, err := revalidateProjectConfig(ctx, svc, project)
			if err != nil {
				return err
			}

			changed, finalCfg := addParameter(cfg.cfg, key, groupName, description, spec)
			if !changed {
				corelog.For("add").Error("parameter already exists; skipping", "project_id", project.ProjectID, "parameter", key)
				break
			}
			diffText, hasChanges := shared.RenderRemoteConfigDiff(cfg.cfg, finalCfg)
			if !hasChanges {
				break
			}
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)

			finalRaw, err := marshalRemoteConfig(finalCfg)
			if err != nil {
				return err
			}
			if err := svc.ValidateRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, cfg.cache.ETag); err != nil {
				if isRemoteConfigConflict(err) {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "remote config changed during add; restarting project %s\n", project.ProjectID)
					continue
				}
				return err
			}
			if _, _, err := svc.PublishRemoteConfigWithETag(ctx, project.ProjectID, finalRaw, cfg.cache.ETag); err != nil {
				if isRemoteConfigConflict(err) {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "remote config changed during add; restarting project %s\n", project.ProjectID)
					continue
				}
				return err
			}

			totals.modifiedProjects++
			totals.addedParams++
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "➕ published: %s\n", project.ProjectID)
			break
		}
	}

	logAddTotals("remote", totals)
	return nil
}

// runAddStdin runs run add stdin and returns the resulting value or error.
func runAddStdin(cmd *cobra.Command, key, groupName, description string, spec addValueSpec, projectExpr string) error {
	raw, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if !json.Valid(raw) {
		return fmt.Errorf("stdin remote config is not valid json")
	}

	remoteConfigRaw, err := shared.ExtractRemoteConfigJSON(raw)
	if err != nil {
		return err
	}

	cfg, err := firebase.ParseRemoteConfig(remoteConfigRaw)
	if err != nil {
		return fmt.Errorf("decode stdin remote config: %w", err)
	}
	cfg = cloneRemoteConfig(cfg)

	if !shared.MatchProjectByExpr(core.Project{Name: "<stdin>", ProjectID: "<stdin>"}, cfg, projectExpr) {
		return nil
	}

	order, err := parseRemoteConfigOrder(remoteConfigRaw)
	if err != nil {
		return fmt.Errorf("parse stdin remote config order: %w", err)
	}

	changed, finalCfg := addParameter(cfg, key, groupName, description, spec)
	if !changed {
		corelog.For("add").Error("parameter already exists; skipping", "project_id", "<stdin>", "parameter", key)
	} else {
		diffText, hasChanges := shared.RenderRemoteConfigDiff(cfg, finalCfg)
		if hasChanges {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), diffText)
		}
		if groupName == "" {
			order.parameters = append(order.parameters, key)
		} else {
			if !slices.Contains(order.groups, groupName) {
				order.groups = append(order.groups, groupName)
			}
			order.groupParameters[groupName] = append(order.groupParameters[groupName], key)
		}
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

	totals := addTotals{}
	if changed {
		totals.modifiedProjects = 1
		totals.addedParams = 1
	}
	logAddTotals("stdin", totals)
	return nil
}

// addParameter handles add parameter and returns the resulting value or error.
func addParameter(cfg *firebase.RemoteConfig, key, groupName, description string, spec addValueSpec) (bool, *firebase.RemoteConfig) {
	finalCfg := cloneRemoteConfig(cfg)
	if parameterExists(finalCfg, key) {
		return false, finalCfg
	}

	param := firebase.RemoteConfigParam{
		DefaultValue: &firebase.RemoteConfigValue{Value: spec.value},
		Description:  description,
		ValueType:    spec.valueType,
	}

	if groupName == "" {
		if finalCfg.Parameters == nil {
			finalCfg.Parameters = map[string]firebase.RemoteConfigParam{}
		}
		finalCfg.Parameters[key] = param
		return true, finalCfg
	}

	if finalCfg.ParameterGroups == nil {
		finalCfg.ParameterGroups = map[string]firebase.RemoteConfigGroup{}
	}
	group := finalCfg.ParameterGroups[groupName]
	if group.Parameters == nil {
		group.Parameters = map[string]firebase.RemoteConfigParam{}
	}
	group.Parameters[key] = param
	finalCfg.ParameterGroups[groupName] = group
	return true, finalCfg
}

// parameterExists handles parameter exists and returns the resulting value or error.
func parameterExists(cfg *firebase.RemoteConfig, key string) bool {
	if cfg == nil {
		return false
	}
	if _, ok := cfg.Parameters[key]; ok {
		return true
	}
	for _, group := range cfg.ParameterGroups {
		if _, ok := group.Parameters[key]; ok {
			return true
		}
	}
	return false
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
	return &projectConfig{
		project: project,
		cache:   cache,
		cfg:     cloneRemoteConfig(cfg),
	}, nil
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

// logAddTotals handles log add totals and returns the resulting value or error.
func logAddTotals(mode string, totals addTotals) {
	corelog.For("add").Info("total", "mode", mode, "projects", totals.modifiedProjects, "parameters", totals.addedParams)
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

// marshalPrettyRemoteConfigWithOrder handles marshal pretty remote config with order and returns the resulting value or error.
func marshalPrettyRemoteConfigWithOrder(cfg *firebase.RemoteConfig, order remoteConfigOrder) ([]byte, error) {
	if cfg == nil {
		return []byte("{}\n"), nil
	}

	var buf bytes.Buffer
	writeRemoteConfigObject(&buf, cfg, order, 0)
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// writeRemoteConfigObject writes write remote config object and returns the resulting value or error.
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

// remoteConfigFieldPresent handles remote config field present and returns the resulting value or error.
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

// writeObject writes write object and returns the resulting value or error.
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

// writeConditions writes write conditions and returns the resulting value or error.
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

// writeCondition writes write condition and returns the resulting value or error.
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

// writeGroups writes write groups and returns the resulting value or error.
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

// writeGroup writes write group and returns the resulting value or error.
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

// writeParametersMap writes write parameters map and returns the resulting value or error.
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

// writeParam writes write param and returns the resulting value or error.
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

// writeConditionalValues writes write conditional values and returns the resulting value or error.
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

// writeRemoteConfigValue writes write remote config value and returns the resulting value or error.
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

// writeVersion writes write version and returns the resulting value or error.
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

// writeIndent writes write indent and returns the resulting value or error.
func writeIndent(buf *bytes.Buffer, indent int) {
	for range indent {
		buf.WriteString("  ")
	}
}

// writeJSONString writes write jsonstring and returns the resulting value or error.
func writeJSONString(buf *bytes.Buffer, value string) {
	encoded, _ := json.Marshal(value)
	buf.Write(normalizeExportJSON(encoded))
}

// normalizeExportJSON handles normalize export json and returns the resulting value or error.
func normalizeExportJSON(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte(`\u003c`), []byte("<"))
	body = bytes.ReplaceAll(body, []byte(`\u003e`), []byte(">"))
	body = bytes.ReplaceAll(body, []byte(`\u0026`), []byte("&"))
	return body
}

// orderedKeys handles ordered keys and returns the resulting value or error.
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

// parseRemoteConfigOrder parses parse remote config order and returns the resulting value or error.
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

// objectMemberOrder handles object member order and returns the resulting value or error.
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

// collectConditionalValueOrders handles collect conditional value orders and returns the resulting value or error.
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

// orderPath handles order path and returns the resulting value or error.
func orderPath(groupName, paramKey string) string {
	if groupName == "" {
		return paramKey
	}
	return groupName + "\x00" + paramKey
}

// parseOrderedJSONValue parses parse ordered jsonvalue and returns the resulting value or error.
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

// parseOrderedJSONObject parses parse ordered jsonobject and returns the resulting value or error.
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

// parseOrderedJSONArray parses parse ordered jsonarray and returns the resulting value or error.
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

// skipJSONWhitespace handles skip jsonwhitespace and returns the resulting value or error.
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

// scanJSONStringEnd handles scan jsonstring end and returns the resulting value or error.
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

// scanPrimitiveEnd handles scan primitive end and returns the resulting value or error.
func scanPrimitiveEnd(body []byte, start int) (int, bool) {
	for i := start; i < len(body); i++ {
		switch body[i] {
		case ',', '}', ']', ' ', '\n', '\r', '\t':
			return i, true
		}
	}
	return len(body), true
}

// unquoteJSONString handles unquote jsonstring and returns the resulting value or error.
func unquoteJSONString(raw []byte) (string, error) {
	var out string
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	return out, nil
}
