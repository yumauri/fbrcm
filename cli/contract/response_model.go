package contract

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

// TextData is used when a successful machine result contains generated text,
// such as shell completion output.
type TextData struct {
	Text string `json:"text"`
}

// JSONDocument marks a command result that is itself an arbitrary JSON
// document, such as `schema show`.
type JSONDocument struct{}

type responseModel struct {
	types        []reflect.Type
	jsonDocument bool
}

var responseModels sync.Map // map[*cobra.Command]responseModel

func hasResponseModel(cmd *cobra.Command) bool {
	_, ok := responseModels.Load(cmd)
	return ok
}

// RegisterResponse records every successful data DTO a command can emit. A
// top-level slice is modeled as the contract's {count,items} collection.
func RegisterResponse(cmd *cobra.Command, variants ...any) {
	if cmd == nil || len(variants) == 0 {
		panic("contract.RegisterResponse requires a command and at least one DTO")
	}
	model := responseModel{types: make([]reflect.Type, 0, len(variants))}
	for _, variant := range variants {
		if _, ok := variant.(JSONDocument); ok {
			model.jsonDocument = true
			continue
		}
		typeOf := reflect.TypeOf(variant)
		if typeOf == nil {
			panic("contract.RegisterResponse received an untyped nil DTO")
		}
		model.types = append(model.types, typeOf)
	}
	responseModels.Store(cmd, model)
}

// RegisterNoData records that a command can only fail in machine mode and has
// no successful data payload.
func RegisterNoData(cmd *cobra.Command) {
	if cmd == nil {
		panic("contract.RegisterNoData requires a command")
	}
	responseModels.Store(cmd, responseModel{})
}

// MustRegisterResponsePath registers a child command after its command tree is
// assembled. It panics during command construction if the path drifts.
func MustRegisterResponsePath(root *cobra.Command, path string, variants ...any) {
	cmd, remaining, err := root.Find(strings.Fields(path))
	if err != nil || len(remaining) != 0 || cmd == root && strings.TrimSpace(path) != "" {
		panic(fmt.Sprintf("register response model for %q: error=%v remaining=%v", path, err, remaining))
	}
	RegisterResponse(cmd, variants...)
}

// MustRegisterNoDataPath is the no-success-payload counterpart of
// MustRegisterResponsePath.
func MustRegisterNoDataPath(root *cobra.Command, path string) {
	cmd, remaining, err := root.Find(strings.Fields(path))
	if err != nil || len(remaining) != 0 || cmd == root && strings.TrimSpace(path) != "" {
		panic(fmt.Sprintf("register response model for %q: error=%v remaining=%v", path, err, remaining))
	}
	RegisterNoData(cmd)
}

// ResponseDataSchema returns the strict command-specific schema for envelope
// data. Null is accepted for failures without usable data.
func ResponseDataSchema(cmd *cobra.Command) (map[string]any, error) {
	success, err := ResponseSuccessDataSchema(cmd)
	if err != nil {
		return nil, err
	}
	if _, impossible := success["not"]; impossible && len(success) == 1 {
		return map[string]any{"type": "null"}, nil
	}
	return map[string]any{"oneOf": []any{success, map[string]any{"type": "null"}}}, nil
}

// ResponseSuccessDataSchema returns the non-null command-specific data schema
// for successful and partially successful envelopes. Commands registered with
// RegisterNoData receive an unsatisfiable schema because they cannot succeed.
func ResponseSuccessDataSchema(cmd *cobra.Command) (map[string]any, error) {
	value, ok := responseModels.Load(cmd)
	if !ok {
		return nil, fmt.Errorf("command %s has no registered response DTO", CommandID(cmd))
	}
	model := value.(responseModel)
	variants := make([]any, 0, len(model.types)+1)
	seen := make(map[string]bool)
	if model.jsonDocument {
		variants = append(variants, map[string]any{"type": "object", "description": "arbitrary JSON document"})
	}
	for _, typeOf := range model.types {
		schema := schemaForResponseVariant(typeOf)
		applyCommandResponseSemantics(CommandID(cmd), typeOf, schema)
		raw, err := json.Marshal(schema)
		if err != nil {
			return nil, err
		}
		if seen[string(raw)] {
			continue
		}
		seen[string(raw)] = true
		variants = append(variants, schema)
	}
	if len(variants) == 0 {
		return map[string]any{"not": map[string]any{}}, nil
	}
	if len(variants) == 1 {
		return variants[0].(map[string]any), nil
	}
	return map[string]any{"oneOf": variants}, nil
}

// SchemaForDTO returns the non-null schema used for a concrete DTO. It is used
// by the schema publisher for reusable stdin and artifact definitions.
func SchemaForDTO(dto any) map[string]any {
	typeOf := reflect.TypeOf(dto)
	if typeOf == nil {
		panic("contract.SchemaForDTO received an untyped nil DTO")
	}
	return schemaForResponseVariant(typeOf)
}

func schemaForResponseVariant(typeOf reflect.Type) map[string]any {
	if typeOf.Kind() != reflect.Slice && typeOf.Kind() != reflect.Array {
		return schemaForResponseType(typeOf, make(map[reflect.Type]bool))
	}
	item := schemaForResponseType(typeOf.Elem(), make(map[reflect.Type]bool))
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"count", "items"},
		"x-fbrcm-invariants": []any{
			invariant("eq", "left", invariantField("count"), "right", invariant("length", "value", invariantField("items"))),
		},
		"properties": map[string]any{
			"count": map[string]any{"type": "integer", "minimum": 0},
			"items": map[string]any{"type": "array", "items": item},
		},
	}
	applyCollectionSemantics(schema, typeOf.Elem())
	return schema
}

func applyCollectionSemantics(schema map[string]any, itemType reflect.Type) {
	qualifiedName := itemType.PkgPath() + "." + itemType.Name()
	countWhere := func(field string, value any) map[string]any {
		return invariant("count_where", "collection", invariantField("items"), "where", invariant("eq", "left", invariantField("item."+field), "right", invariantConst(value)))
	}
	appendInvariant := func(rule map[string]any) {
		rules, _ := schema["x-fbrcm-invariants"].([]any)
		schema["x-fbrcm-invariants"] = append(rules, rule)
	}
	appendInvariant(invariant("eq", "left", invariantField("count"), "right", invariant("length", "value", invariantField("items"))))
	switch qualifiedName {
	case "github.com/yumauri/fbrcm/cli/commands/auth.authListItem":
		appendInvariant(invariant("implies", "if", invariant("gt", "left", invariant("length", "value", invariantField("items")), "right", invariantConst(0)), "then", invariant("eq", "left", countWhere("default", true), "right", invariantConst(1))))
		appendInvariant(invariant("implies", "if", invariant("eq", "left", invariant("length", "value", invariantField("items")), "right", invariantConst(0)), "then", invariant("eq", "left", countWhere("default", true), "right", invariantConst(0))))
	case "github.com/yumauri/fbrcm/cli/commands/profile.profileListItem":
		appendInvariant(invariant("implies", "if", invariant("gt", "left", invariant("length", "value", invariantField("items")), "right", invariantConst(0)), "then", invariant("eq", "left", countWhere("active", true), "right", invariantConst(1))))
		appendInvariant(invariant("implies", "if", invariant("eq", "left", invariant("length", "value", invariantField("items")), "right", invariantConst(0)), "then", invariant("eq", "left", countWhere("active", true), "right", invariantConst(0))))
	case "github.com/yumauri/fbrcm/cli/commands/versions.versionJSON":
		appendInvariant(invariant("eq", "left", invariant("gt", "left", countWhere("current", true), "right", invariantConst(1)), "right", invariantConst(false)))
	}
}

func schemaForResponseType(typeOf reflect.Type, visiting map[reflect.Type]bool) map[string]any {
	if typeOf.Kind() == reflect.Pointer {
		return nullableSchema(schemaForResponseType(typeOf.Elem(), visiting))
	}
	if typeOf == reflect.TypeFor[json.RawMessage]() {
		return map[string]any{}
	}
	if typeOf == reflect.TypeFor[time.Time]() {
		return map[string]any{"type": "string", "format": "date-time"}
	}
	if typeOf == reflect.TypeFor[ArtifactData]() {
		return artifactDataSchema()
	}
	if typeOf == reflect.TypeFor[Capability]() {
		return map[string]any{"$ref": CapabilitySchemaID()}
	}
	if typeOf == reflect.TypeFor[CapabilityIndex]() {
		return capabilityIndexSchema()
	}
	if typeOf.PkgPath() == "github.com/yumauri/fbrcm/core/firebase" && typeOf.Name() == "RemoteConfigValue" {
		return remoteConfigValueSchema()
	}
	switch typeOf.Kind() {
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.String:
		if ref := semanticStringRef(typeOf); ref != "" {
			return map[string]any{"$ref": ref}
		}
		schema := map[string]any{"type": "string"}
		if values := semanticStringEnum(typeOf); len(values) > 0 {
			schema["enum"] = values
		}
		return schema
	case reflect.Interface:
		return map[string]any{}
	case reflect.Slice:
		return map[string]any{"type": []string{"array", "null"}, "items": schemaForResponseType(typeOf.Elem(), visiting)}
	case reflect.Array:
		return map[string]any{"type": "array", "items": schemaForResponseType(typeOf.Elem(), visiting), "minItems": typeOf.Len(), "maxItems": typeOf.Len()}
	case reflect.Map:
		if typeOf.Key().Kind() != reflect.String {
			return map[string]any{"type": "object"}
		}
		return map[string]any{"type": []string{"object", "null"}, "additionalProperties": schemaForResponseType(typeOf.Elem(), visiting)}
	case reflect.Struct:
		if visiting[typeOf] {
			return map[string]any{"type": "object", "description": "recursive " + typeOf.String()}
		}
		visiting[typeOf] = true
		defer delete(visiting, typeOf)
		properties := make(map[string]any)
		required := make([]string, 0, typeOf.NumField())
		for field := range typeOf.Fields() {
			if field.PkgPath != "" {
				continue
			}
			tag := field.Tag.Get("json")
			name, options := parseJSONTag(tag)
			if name == "-" {
				continue
			}
			if field.Anonymous && name == "" {
				embedded := field.Type
				if embedded.Kind() == reflect.Pointer {
					embedded = embedded.Elem()
				}
				if embedded.Kind() == reflect.Struct {
					embeddedSchema := schemaForResponseType(embedded, visiting)
					mergeObjectSchema(properties, &required, embeddedSchema)
					continue
				}
			}
			if name == "" {
				name = field.Name
			}
			fieldSchema := schemaForResponseType(field.Type, visiting)
			applyFieldSemantics(fieldSchema, field, name)
			properties[name] = fieldSchema
			if !options["omitempty"] && !options["omitzero"] {
				required = append(required, name)
			}
		}
		sort.Strings(required)
		additionalProperties := false
		marshaler := reflect.TypeFor[json.Marshaler]()
		if typeOf.Implements(marshaler) || reflect.PointerTo(typeOf).Implements(marshaler) {
			// A custom marshaler may preserve fields that are not represented by
			// exported Go fields. Keep its known properties useful for discovery,
			// but do not falsely reject preserved wire fields.
			additionalProperties = true
		}
		result := map[string]any{"type": "object", "additionalProperties": additionalProperties, "required": required, "properties": properties}
		applyTypeSemantics(result, typeOf)
		return result
	default:
		return map[string]any{}
	}
}

func applyTypeSemantics(schema map[string]any, typeOf reflect.Type) {
	properties, _ := schema["properties"].(map[string]any)
	qualifiedName := typeOf.PkgPath() + "." + typeOf.Name()
	if _, hasCount := properties["count"]; hasCount {
		for _, collectionName := range []string{"items", "commands"} {
			if collection, ok := properties[collectionName].(map[string]any); ok {
				collection["type"] = "array"
				schema["x-fbrcm-invariants"] = []any{
					invariant("eq", "left", invariantField("count"), "right", invariant("length", "value", invariantField(collectionName))),
				}
				break
			}
		}
	}
	switch qualifiedName {
	case "github.com/yumauri/fbrcm/cli/commands/config.configResetResult":
		schema["allOf"] = []any{
			fieldValueConstraint("status", "unchanged", map[string]any{"changed": map[string]any{"const": false}}),
			fieldValueConstraint("status", "reset", map[string]any{"changed": map[string]any{"const": true}}),
		}
	case "github.com/yumauri/fbrcm/cli/commands/cache.cacheClearResult":
		schema["allOf"] = []any{
			fieldValueConstraint("status", "unchanged", map[string]any{
				"snapshots_deleted": map[string]any{"const": 0},
				"targets_affected":  map[string]any{"const": 0},
				"bytes_deleted":     map[string]any{"const": 0},
			}),
			fieldValueConstraint("status", "cleared", map[string]any{
				"snapshots_deleted": map[string]any{"minimum": 1},
				"targets_affected":  map[string]any{"minimum": 1},
			}),
		}
	case "github.com/yumauri/fbrcm/cli/commands/config.configValidationResult":
		for name, severity := range map[string]string{"errors": "error", "warnings": "warning"} {
			collection := properties[name].(map[string]any)
			collection["type"] = "array"
			item := collection["items"].(map[string]any)
			itemProperties := item["properties"].(map[string]any)
			itemProperties["severity"] = map[string]any{"const": severity}
		}
		schema["allOf"] = []any{
			map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"errors": map[string]any{"maxItems": 0}},
					"required":   []string{"errors"},
				},
				"then": map[string]any{"properties": map[string]any{"valid": map[string]any{"const": true}}},
				"else": map[string]any{"properties": map[string]any{"valid": map[string]any{"const": false}}},
			},
		}
	case "github.com/yumauri/fbrcm/tui/config.Diagnostic":
		properties["severity"] = map[string]any{"type": "string", "enum": []string{"error", "warning"}}
		properties["code"] = map[string]any{"type": "string", "enum": []string{
			"duplicate_binding", "empty_binding", "invalid_binding", "invalid_profile", "keybinding_conflict", "legacy_bindings",
			"missing_profile", "project_alias_source", "repository_scope_required", "toml_decode", "unknown_action", "unknown_block",
		}}
	case "github.com/yumauri/fbrcm/cli/commands/auth.authBindItem":
		schema["allOf"] = []any{
			fieldValueConstraint("status", "bound", map[string]any{"reason": map[string]any{"type": "null"}}),
			fieldValueConstraint("status", "skipped", map[string]any{"reason": map[string]any{"type": "string", "minLength": 1}}),
		}
	case "github.com/yumauri/fbrcm/cli/commands/auth.authBindResult":
		if items, ok := properties["items"].(map[string]any); ok {
			items["type"] = "array"
		}
		schema["x-fbrcm-invariants"] = []any{
			invariant("eq", "left", invariantField("bound"), "right", invariant("count_where", "collection", invariantField("items"), "where", invariant("eq", "left", invariantField("item.status"), "right", invariantConst("bound")))),
			invariant("eq", "left", invariantField("skipped"), "right", invariant("count_where", "collection", invariantField("items"), "where", invariant("eq", "left", invariantField("item.status"), "right", invariantConst("skipped")))),
			invariant("eq", "left", invariant("sum", "values", []any{invariantField("bound"), invariantField("skipped")}), "right", invariant("length", "value", invariantField("items"))),
		}
	case "github.com/yumauri/fbrcm/core/config.AuthEntry",
		"github.com/yumauri/fbrcm/cli/commands/auth.authPathResult":
		applyAuthTypeSemantics(schema)
	case "github.com/yumauri/fbrcm/cli/commands/auth.authMutationResult",
		"github.com/yumauri/fbrcm/cli/commands/auth.authLoginResult":
		schema["x-fbrcm-invariants"] = []any{
			invariant("eq", "left", invariantField("auth_id"), "right", invariantField("paths.id")),
			invariant("eq", "left", invariantField("type"), "right", invariantField("paths.type")),
		}
	case "github.com/yumauri/fbrcm/cli/shared.ProjectJSON":
		applyProjectTemplateSemantics(schema)
		if aliases, ok := properties["aliases"].(map[string]any); ok {
			aliases["type"] = "array"
			aliases["uniqueItems"] = true
		}
	case "github.com/yumauri/fbrcm/cli/commands/project.projectTemplatesJSON":
		applyProjectTemplateSemantics(schema)
	case "github.com/yumauri/fbrcm/cli/commands/project.projectOpenResult":
		properties["opened"] = map[string]any{"const": false}
	case "github.com/yumauri/fbrcm/cli/commands/conditions.conditionValidationResult":
		properties["valid"] = map[string]any{"const": true}
	case "github.com/yumauri/fbrcm/cli/commands/profile.profileRenameResult":
		schema["x-fbrcm-invariants"] = []any{
			invariant("iff", "left", invariant("eq", "left", invariantField("changed"), "right", invariantConst(false)), "right", invariant("eq", "left", invariantField("old_profile"), "right", invariantField("new_profile"))),
		}
	case "github.com/yumauri/fbrcm/cli/commands/profile.profileSwitchResult":
		schema["x-fbrcm-invariants"] = []any{
			invariant("iff", "left", invariant("eq", "left", invariantField("changed"), "right", invariantConst(false)), "right", invariant("eq", "left", invariantField("previous_profile"), "right", invariantField("requested_profile"))),
			invariant("iff", "left", invariant("eq", "left", invariantField("overridden"), "right", invariantConst(false)), "right", invariant("eq", "left", invariantField("effective_profile"), "right", invariantField("requested_profile"))),
		}
	case "github.com/yumauri/fbrcm/cli/commands/hooks.untrustResult":
		properties["trusted"] = map[string]any{"const": false}
	case "github.com/yumauri/fbrcm/cli/commands/hooks.statusResult":
		schema["allOf"] = []any{
			map[string]any{
				"if": map[string]any{"properties": map[string]any{"local_hooks": map[string]any{"const": false}}, "required": []string{"local_hooks"}},
				"then": map[string]any{
					"properties": map[string]any{"trusted": map[string]any{"const": true}},
					"not":        map[string]any{"required": []string{"fingerprint"}},
				},
				"else": map[string]any{
					"properties": map[string]any{"fingerprint": map[string]any{"type": "string", "minLength": 1}},
					"required":   []string{"fingerprint", "local_config"},
				},
			},
		}
	case "github.com/yumauri/fbrcm/cli/commands/projects.projectsResetResult":
		properties["changed"] = map[string]any{"type": "boolean"}
	case "github.com/yumauri/fbrcm/cli/commands/projects.projectAliasRemoveResult":
		schema["allOf"] = []any{
			fieldValueShapeConstraint("status", "not_found", map[string]any{"changed": map[string]any{"const": false}}, nil, []string{"previous_project_id", "source", "remaining_source"}),
			fieldValueShapeConstraint("status", "removed", map[string]any{"changed": map[string]any{"const": true}, "source": map[string]any{"const": "fbrcm"}, "previous_project_id": map[string]any{"type": "string", "minLength": 1}}, []string{"previous_project_id", "source"}, []string{"remaining_source"}),
			fieldValueShapeConstraint("status", "removed_native", map[string]any{"changed": map[string]any{"const": true}, "source": map[string]any{"const": "both"}, "remaining_source": map[string]any{"const": "firebase"}, "previous_project_id": map[string]any{"type": "string", "minLength": 1}}, []string{"previous_project_id", "source", "remaining_source"}, nil),
		}
	case "github.com/yumauri/fbrcm/cli/commands/projects.projectAliasImportItem":
		properties["action"] = map[string]any{"type": "string", "enum": []string{"add", "unchanged", "keep", "overwrite"}}
		schema["allOf"] = []any{
			fieldValueShapeConstraint("action", "add", nil, nil, []string{"previous_project_id"}),
			fieldValueShapeConstraint("action", "unchanged", map[string]any{"previous_project_id": map[string]any{"type": "string", "minLength": 1}}, []string{"previous_project_id"}, nil),
			fieldValueShapeConstraint("action", "keep", map[string]any{"previous_project_id": map[string]any{"type": "string", "minLength": 1}}, []string{"previous_project_id"}, nil),
			fieldValueShapeConstraint("action", "overwrite", map[string]any{"previous_project_id": map[string]any{"type": "string", "minLength": 1}}, []string{"previous_project_id"}, nil),
		}
		schema["x-fbrcm-invariants"] = []any{
			invariant("implies", "if", invariant("eq", "left", invariantField("action"), "right", invariantConst("unchanged")), "then", invariant("eq", "left", invariantField("previous_project_id"), "right", invariantField("project_id"))),
			invariant("implies", "if", invariant("in", "value", invariantField("action"), "set", []any{"keep", "overwrite"}), "then", invariant("eq", "left", invariant("eq", "left", invariantField("previous_project_id"), "right", invariantField("project_id")), "right", invariantConst(false))),
		}
	case "github.com/yumauri/fbrcm/cli/commands/projects.projectAliasImportResult":
		if items, ok := properties["items"].(map[string]any); ok {
			items["type"] = "array"
		}
		schema["x-fbrcm-invariants"] = []any{
			invariant("iff", "left", invariantField("changed"), "right", invariant("gt", "left", invariant("count_where", "collection", invariantField("items"), "where", invariant("in", "value", invariantField("item.action"), "set", []any{"add", "overwrite"})), "right", invariantConst(0))),
		}
	case "github.com/yumauri/fbrcm/cli/commands/project/import.importResult":
		applyImportResultSemantics(schema)
	case "github.com/yumauri/fbrcm/cli/commands/projects.promoteResult":
		applyPromoteResultSemantics(schema)
	case "github.com/yumauri/fbrcm/cli/commands/draft.publishResult":
		applyDraftPublishResultSemantics(schema)
	case "github.com/yumauri/fbrcm/cli/commands/versions.versionPublishResult":
		applyVersionPublishResultSemantics(schema)
	case "github.com/yumauri/fbrcm/cli/commands/versions.versionShowResult":
		schema["x-fbrcm-invariants"] = []any{
			invariant("eq", "left", invariantField("cached"), "right", invariantField("version.cached")),
		}
	case "github.com/yumauri/fbrcm/cli/shared/rc.RemoteMutationJSONResult":
		applyRemoteMutationResultSemantics(schema)
	case "github.com/yumauri/fbrcm/cli/shared/rc.RemoteMutationJSONError",
		"github.com/yumauri/fbrcm/cli/commands/versions.versionOperationError",
		"github.com/yumauri/fbrcm/cli/commands/draft.draftPublishError",
		"github.com/yumauri/fbrcm/cli/commands/projects.promoteError",
		"github.com/yumauri/fbrcm/cli/commands/project/import.importError":
		if message, ok := properties["message"].(map[string]any); ok {
			message["maxLength"] = 4097
			message["x-fbrcm-safe-text"] = "at most 4096 Unicode code points, followed by one ellipsis when truncated"
		}
	}
	if typeOf.PkgPath() != "github.com/yumauri/fbrcm/core/firebase" {
		return
	}
	switch typeOf.Name() {
	case "RemoteConfig":
		if conditions, ok := properties["conditions"].(map[string]any); ok {
			conditions["uniqueItems"] = true
			conditions["x-fbrcm-validation"] = []any{map[string]any{"operator": "unique_by", "field": "name"}}
		}
		for _, name := range []string{"parameters", "parameterGroups"} {
			if entries, ok := properties[name].(map[string]any); ok {
				entries["propertyNames"] = map[string]any{"type": "string", "minLength": 1, "maxLength": 256}
			}
		}
		schema["x-fbrcm-validation"] = []any{map[string]any{
			"operator": "remote_validate", "service": "firebase_remote_config", "grammar": "template",
		}}
	case "RemoteConfigCondition":
		schema["required"] = []string{"expression", "name"}
		properties["name"] = map[string]any{"type": "string", "pattern": `.*\S.*`}
		properties["expression"] = map[string]any{"type": "string", "pattern": `.*\S.*`, "x-fbrcm-validation": []any{map[string]any{
			"operator": "remote_validate", "service": "firebase_remote_config", "grammar": "condition_expression",
		}}}
		colors := []string{"CONDITION_DISPLAY_COLOR_UNSPECIFIED", "BLUE", "BROWN", "CYAN", "DEEP_ORANGE", "GREEN", "INDIGO", "LIME", "ORANGE", "PINK", "PURPLE", "TEAL"}
		properties["tagColor"] = map[string]any{
			"type":                      "string",
			"pattern":                   caseInsensitiveValuesPattern(append([]string{""}, colors...)),
			"x-fbrcm-normalized-values": colors,
		}
	case "RemoteConfigParam":
		properties["valueType"] = map[string]any{"type": "string", "enum": []string{"", "PARAMETER_VALUE_TYPE_UNSPECIFIED", "STRING", "BOOLEAN", "NUMBER", "JSON"}}
		properties["description"] = map[string]any{"type": "string", "maxLength": 256}
	case "RemoteConfigGroup":
		properties["description"] = map[string]any{"type": "string", "maxLength": 256}
		if entries, ok := properties["parameters"].(map[string]any); ok {
			entries["propertyNames"] = map[string]any{"type": "string", "minLength": 1, "maxLength": 256}
		}
	}
}

func applyCommandResponseSemantics(commandID string, typeOf reflect.Type, schema map[string]any) {
	if typeOf == reflect.TypeFor[ArtifactData]() {
		applyArtifactCommandSemantics(commandID, schema)
	}
	properties, _ := schema["properties"].(map[string]any)
	switch commandID {
	case "auth.add.oauth", "auth.add.service-account", "auth.add.gcloud":
		authType := strings.TrimPrefix(commandID, "auth.add.")
		properties["type"] = map[string]any{"const": authType}
		if paths, ok := properties["paths"].(map[string]any); ok {
			pathProperties := paths["properties"].(map[string]any)
			pathProperties["type"] = map[string]any{"const": authType}
		}
	case "versions.restore":
		properties["operation"] = map[string]any{"const": "restore"}
	case "versions.rollback":
		properties["operation"] = map[string]any{"const": "rollback"}
	case "experiments.delete":
		properties["kind"] = map[string]any{"const": "experiment"}
	case "rollouts.delete":
		properties["kind"] = map[string]any{"const": "rollout"}
	case "hooks.trust":
		properties["local_hooks"] = map[string]any{"const": true}
		properties["trusted"] = map[string]any{"const": true}
	}
}

func applyAuthTypeSemantics(schema map[string]any) {
	properties := schema["properties"].(map[string]any)
	properties["type"] = map[string]any{"type": "string", "enum": []string{"oauth", "service-account", "gcloud"}}
	nonempty := map[string]any{"type": "string", "minLength": 1}
	schema["allOf"] = []any{
		fieldValueShapeConstraint("type", "oauth", map[string]any{"client_secret_path": nonempty, "token_path": nonempty}, []string{"client_secret_path", "token_path"}, []string{"service_account_path"}),
		fieldValueShapeConstraint("type", "service-account", map[string]any{"service_account_path": nonempty}, []string{"service_account_path"}, []string{"client_secret_path", "token_path"}),
		fieldValueShapeConstraint("type", "gcloud", map[string]any{}, nil, []string{"client_secret_path", "token_path", "service_account_path"}),
	}
}

func applyProjectTemplateSemantics(schema map[string]any) {
	properties := schema["properties"].(map[string]any)
	templates := properties["templates"].(map[string]any)
	templates["type"] = "array"
	templates["minItems"] = 1
	templates["maxItems"] = 2
	templates["uniqueItems"] = true
	templates["items"] = map[string]any{"type": "string", "enum": []string{"client", "server"}}
	properties["primary_template"] = map[string]any{"type": "string", "enum": []string{"client", "server"}}
	schema["allOf"] = []any{
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"primary_template": map[string]any{"const": "client"}}, "required": []string{"primary_template"}},
			"then": map[string]any{"properties": map[string]any{"templates": map[string]any{"contains": map[string]any{"const": "client"}, "minContains": 1}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"primary_template": map[string]any{"const": "server"}}, "required": []string{"primary_template"}},
			"then": map[string]any{"properties": map[string]any{"templates": map[string]any{"contains": map[string]any{"const": "server"}, "minContains": 1}}},
		},
	}
}

func applyArtifactCommandSemantics(commandID string, schema map[string]any) {
	var content map[string]any
	var mediaType any
	var encodings []string
	switch commandID {
	case "add", "update", "delete":
		content = schemaForResponseType(reflect.TypeFor[firebase.RemoteConfig](), make(map[reflect.Type]bool))
		mediaType = "application/json"
		encodings = []string{"json"}
	case "project.export":
		// Firebase validates only that the export response is JSON before this
		// command wraps it as an artifact; it does not decode the document as a
		// Remote Config object on this path.
		content = map[string]any{}
		mediaType = "application/json"
		encodings = []string{"json", "none"}
	case "versions.export":
		content = schemaForResponseType(reflect.TypeFor[firebase.RemoteConfig](), make(map[reflect.Type]bool))
		mediaType = "application/json"
		encodings = []string{"json", "none"}
	case "draft.show":
		// --raw reads the selected on-disk draft bytes directly so recovery can
		// return a damaged file. A syntactically valid raw file may therefore be
		// any JSON value, not only a decoded Draft or Remote Config object.
		content = map[string]any{}
		mediaType = "application/json"
		encodings = []string{"json", "utf-8", "base64", "none"}
	case "project.defaults":
		// The defaults endpoint is a byte download. The runtime intentionally
		// preserves the response rather than validating it against a format, so
		// every artifact encoding remains reachable.
		content = map[string]any{}
		mediaType = []string{"application/json", "application/xml", "application/x-plist"}
		encodings = []string{"json", "utf-8", "base64", "none"}
	default:
		return
	}
	properties := schema["properties"].(map[string]any)
	properties["target"] = map[string]any{"type": "string", "minLength": 1}
	properties["encoding"] = map[string]any{"type": "string", "enum": encodings}
	if commandID == "draft.show" {
		// JSON mode has no confirmation bypass for an existing destination,
		// so every artifact it can actually return is non-overwriting.
		properties["overwritten"] = map[string]any{"const": false}
	}
	if values, ok := mediaType.([]string); ok {
		properties["media_type"] = map[string]any{"type": "string", "enum": values}
	} else {
		properties["media_type"] = map[string]any{"const": mediaType}
	}
	content["$anchor"] = "artifact_json_content"
	schema["$defs"] = map[string]any{"artifact_json_content": content}
	contentRef := map[string]any{"$ref": "#artifact_json_content"}
	properties["json_content"] = map[string]any{"anyOf": []any{contentRef, map[string]any{"type": "null"}}}
	variants := schema["oneOf"].([]any)
	jsonVariant := variants[0].(map[string]any)["properties"].(map[string]any)
	jsonVariant["json_content"] = contentRef
}

func fieldValueConstraint(field, value string, thenProperties map[string]any) map[string]any {
	return fieldValueShapeConstraint(field, value, thenProperties, nil, nil)
}

func fieldValueShapeConstraint(field, value string, thenProperties map[string]any, required, forbidden []string) map[string]any {
	then := make(map[string]any)
	if thenProperties != nil {
		then["properties"] = thenProperties
	}
	if len(required) > 0 {
		then["required"] = required
	}
	if len(forbidden) > 0 {
		constraints := make([]any, 0, len(forbidden))
		for _, name := range forbidden {
			constraints = append(constraints, map[string]any{"not": map[string]any{"required": []string{name}}})
		}
		then["allOf"] = constraints
	}
	return map[string]any{
		"if": map[string]any{
			"properties": map[string]any{field: map[string]any{"const": value}},
			"required":   []string{field},
		},
		"then": then,
	}
}

func applyImportResultSemantics(schema map[string]any) {
	properties := schema["properties"].(map[string]any)
	properties["status"] = map[string]any{"type": "string", "enum": []string{"unchanged", "validation-failed", "drafted", "would-draft", "imported", "would-import", "imported-hook-failed", "imported-cache-failed"}}
	null := map[string]any{"type": "null"}
	schema["allOf"] = []any{
		fieldValueShapeConstraint("status", "unchanged", map[string]any{"changed": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "local"}, "error": null}, nil, []string{"published"}),
		fieldValueShapeConstraint("status", "validation-failed", map[string]any{"changed": map[string]any{"const": true}, "draft": map[string]any{"const": false}, "validated": map[string]any{"const": false}, "error": resultErrorSchema("validation")}, nil, []string{"published"}),
		fieldValueShapeConstraint("status", "drafted", map[string]any{"changed": map[string]any{"const": true}, "draft": map[string]any{"const": true}, "dry_run": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "local"}, "error": null}, nil, []string{"published"}),
		fieldValueShapeConstraint("status", "would-draft", map[string]any{"changed": map[string]any{"const": true}, "draft": map[string]any{"const": true}, "dry_run": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "local"}, "error": null}, nil, []string{"published"}),
		fieldValueShapeConstraint("status", "imported", map[string]any{"changed": map[string]any{"const": true}, "draft": map[string]any{"const": false}, "dry_run": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published": map[string]any{"const": true}, "error": null}, []string{"published"}, nil),
		fieldValueShapeConstraint("status", "would-import", map[string]any{"changed": map[string]any{"const": true}, "draft": map[string]any{"const": false}, "dry_run": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "error": null}, nil, []string{"published"}),
		fieldValueShapeConstraint("status", "imported-hook-failed", map[string]any{"changed": map[string]any{"const": true}, "draft": map[string]any{"const": false}, "dry_run": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published": map[string]any{"const": true}, "error": resultErrorSchema("post_publish_hook")}, []string{"published"}, nil),
		fieldValueShapeConstraint("status", "imported-cache-failed", map[string]any{"changed": map[string]any{"const": true}, "draft": map[string]any{"const": false}, "dry_run": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published": map[string]any{"const": true}, "error": resultErrorSchema("cache")}, []string{"published"}, nil),
	}
}

func applyPromoteResultSemantics(schema map[string]any) {
	null := map[string]any{"type": "null"}
	schema["allOf"] = []any{
		fieldValueConstraint("status", "unchanged", map[string]any{"changed": map[string]any{"const": false}, "published": map[string]any{"const": false}, "selected": map[string]any{"const": 0}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "local"}, "error": null}),
		fieldValueConstraint("status", "would-publish", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": true}, "published": map[string]any{"const": false}, "selected": map[string]any{"minimum": 1}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "error": null}),
		fieldValueConstraint("status", "published", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": false}, "published": map[string]any{"const": true}, "selected": map[string]any{"minimum": 1}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "error": null}),
		fieldValueConstraint("status", "validation-failed", map[string]any{"changed": map[string]any{"const": true}, "published": map[string]any{"const": false}, "selected": map[string]any{"minimum": 1}, "validated": map[string]any{"const": false}, "error": resultErrorSchema("validation")}),
		fieldValueConstraint("status", "failed", map[string]any{"changed": map[string]any{"const": true}, "published": map[string]any{"const": false}, "selected": map[string]any{"minimum": 1}, "error": resultErrorSchema("publication")}),
		fieldValueConstraint("status", "published-hook-failed", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": false}, "published": map[string]any{"const": true}, "selected": map[string]any{"minimum": 1}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "error": resultErrorSchema("post_publish_hook")}),
		fieldValueConstraint("status", "published-cache-failed", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": false}, "published": map[string]any{"const": true}, "selected": map[string]any{"minimum": 1}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "error": resultErrorSchema("cache")}),
	}
}

func applyDraftPublishResultSemantics(schema map[string]any) {
	properties := schema["properties"].(map[string]any)
	properties["status"] = map[string]any{"type": "string", "enum": []string{"failed", "unchanged", "would-publish", "already-applied", "published", "published-hook-failed", "published-cache-failed", "published-cleanup-failed", "conflict"}}
	null := map[string]any{"type": "null"}
	errorObject := map[string]any{"type": "object"}
	nonempty := map[string]any{"type": "string", "minLength": 1}
	schema["allOf"] = []any{
		fieldValueShapeConstraint("status", "unchanged", map[string]any{"changed": map[string]any{"const": false}, "dry_run": map[string]any{"const": true}, "draft_deleted": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "local"}, "error": null}, nil, []string{"published_version"}),
		fieldValueShapeConstraint("status", "already-applied", map[string]any{"changed": map[string]any{"const": false}, "dry_run": map[string]any{"const": false}, "draft_deleted": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "local"}, "error": null}, nil, []string{"published_version"}),
		fieldValueShapeConstraint("status", "would-publish", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": true}, "draft_deleted": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "error": null}, nil, []string{"published_version"}),
		fieldValueShapeConstraint("status", "published", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": false}, "draft_deleted": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": nonempty, "error": null}, []string{"published_version"}, nil),
		fieldValueShapeConstraint("status", "published-hook-failed", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": false}, "draft_deleted": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": nonempty, "error": resultErrorSchema("post_publish_hook")}, []string{"published_version"}, nil),
		fieldValueShapeConstraint("status", "published-cache-failed", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": false}, "draft_deleted": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": nonempty, "error": resultErrorSchema("cache")}, []string{"published_version"}, nil),
		fieldValueShapeConstraint("status", "published-cleanup-failed", map[string]any{"changed": map[string]any{"const": true}, "dry_run": map[string]any{"const": false}, "draft_deleted": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": nonempty, "error": resultErrorSchema("cleanup")}, []string{"published_version"}, nil),
		fieldValueConstraint("status", "failed", map[string]any{"error": errorObject}),
		fieldValueConstraint("status", "conflict", map[string]any{"error": errorObject}),
	}
}

func resultErrorSchema(stage string) map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{"stage": map[string]any{"const": stage}}, "required": []string{"stage"}}
}

func applyVersionPublishResultSemantics(schema map[string]any) {
	properties := schema["properties"].(map[string]any)
	properties["status"] = map[string]any{"type": "string", "enum": []string{"unchanged", "would-publish", "published", "failed", "validation-failed", "published-local-update-failed", "published-hook-failed", "published-cache-failed"}}
	null := map[string]any{"type": "null"}
	nonempty := map[string]any{"type": "string", "minLength": 1}
	schema["allOf"] = []any{
		fieldValueConstraint("status", "unchanged", map[string]any{"changed": map[string]any{"const": false}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "local"}, "published_version": null, "error": null}),
		fieldValueConstraint("status", "would-publish", map[string]any{"dry_run": map[string]any{"const": true}, "changed": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": null, "error": null}),
		fieldValueConstraint("status", "published", map[string]any{"dry_run": map[string]any{"const": false}, "changed": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": nonempty, "error": null}),
		fieldValueConstraint("status", "failed", map[string]any{"dry_run": map[string]any{"const": true}, "changed": map[string]any{"const": true}, "validated": map[string]any{"const": false}, "validation_source": map[string]any{"const": "local"}, "published_version": null, "error": resultErrorSchema("preparation")}),
		fieldValueConstraint("status", "validation-failed", map[string]any{"dry_run": map[string]any{"const": true}, "changed": map[string]any{"const": true}, "validated": map[string]any{"const": false}, "published_version": null, "error": resultErrorSchema("validation")}),
		fieldValueConstraint("status", "published-local-update-failed", map[string]any{"dry_run": map[string]any{"const": false}, "changed": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": nonempty, "error": resultErrorSchema("cache")}),
		fieldValueConstraint("status", "published-hook-failed", map[string]any{"dry_run": map[string]any{"const": false}, "changed": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": nonempty, "error": resultErrorSchema("post_publish_hook")}),
		fieldValueConstraint("status", "published-cache-failed", map[string]any{"dry_run": map[string]any{"const": false}, "changed": map[string]any{"const": true}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}, "published_version": nonempty, "error": resultErrorSchema("cache")}),
	}
}

func invariant(operator string, entries ...any) map[string]any {
	result := map[string]any{"op": operator}
	for index := 0; index+1 < len(entries); index += 2 {
		result[entries[index].(string)] = entries[index+1]
	}
	return result
}

func invariantField(path string) map[string]any { return map[string]any{"field": path} }

func invariantConst(value any) map[string]any { return map[string]any{"const": value} }

func applyRemoteMutationResultSemantics(schema map[string]any) {
	null := map[string]any{"type": "null"}
	nonNullString := map[string]any{"type": "string", "minLength": 1}
	errorAt := func(stages ...string) map[string]any {
		stage := map[string]any{"type": "string"}
		if len(stages) == 1 {
			stage = map[string]any{"const": stages[0]}
		} else if len(stages) > 1 {
			stage = map[string]any{"enum": stages}
		}
		return map[string]any{"type": "object", "properties": map[string]any{"stage": stage}, "required": []string{"stage"}}
	}
	constraints := []any{
		fieldValueConstraint("status", "unchanged", map[string]any{
			"changed_item_count": map[string]any{"const": 0}, "published_version": null, "error": null, "retry_selector": null,
			"no_op_reason": map[string]any{"$ref": SemanticRef("no_op_reason")}, "validated": map[string]any{"const": true},
			"validation_source": map[string]any{"const": "local"},
		}),
	}
	for _, status := range []string{"published", "drafted", "would-draft", "would-publish"} {
		properties := map[string]any{
			"changed_item_count": map[string]any{"minimum": 1}, "error": null, "retry_selector": null, "no_op_reason": null, "validated": map[string]any{"const": true},
		}
		if status == "published" || status == "would-publish" {
			properties["validation_source"] = map[string]any{"const": "firebase"}
		} else {
			properties["validation_source"] = map[string]any{"const": "local"}
		}
		if status == "published" {
			properties["published_version"] = nonNullString
		} else {
			properties["published_version"] = null
		}
		constraints = append(constraints, fieldValueConstraint("status", status, properties))
	}
	for _, item := range []struct{ status, stage string }{
		{"published-cache-failed", "cache"},
		{"published-hook-failed", "post_publish_hook"},
	} {
		constraints = append(constraints, fieldValueConstraint("status", item.status, map[string]any{
			"changed_item_count": map[string]any{"minimum": 1}, "published_version": nonNullString, "error": errorAt(item.stage),
			"retry_selector": null, "no_op_reason": null, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"},
		}))
	}
	constraints = append(constraints,
		fieldValueConstraint("status", "preparation-failed", map[string]any{"error": errorAt("preparation"), "no_op_reason": null, "published_version": null, "validated": map[string]any{"const": false}, "validation_source": map[string]any{"const": "local"}}),
		fieldValueConstraint("status", "validation-failed", map[string]any{"changed_item_count": map[string]any{"minimum": 1}, "error": errorAt("validation"), "no_op_reason": null, "published_version": null, "validated": map[string]any{"const": false}, "validation_source": map[string]any{"const": "firebase"}}),
		fieldValueConstraint("status", "publish-failed", map[string]any{"changed_item_count": map[string]any{"minimum": 1}, "error": errorAt("publication", "pre_publish_hook"), "no_op_reason": null, "published_version": null, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}}),
		fieldValueConstraint("status", "draft-failed", map[string]any{"error": errorAt("draft"), "no_op_reason": null, "published_version": null, "validation_source": map[string]any{"const": "local"}}),
		fieldValueConstraint("status", "conflict", map[string]any{"error": errorAt("preparation", "validation", "publication"), "no_op_reason": null, "published_version": null}),
	)
	constraints = append(constraints,
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"status": map[string]any{"const": "conflict"}, "error": errorAt("preparation")}, "required": []string{"status", "error"}},
			"then": map[string]any{"properties": map[string]any{"validated": map[string]any{"const": false}, "validation_source": map[string]any{"const": "local"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"status": map[string]any{"const": "conflict"}, "error": errorAt("validation")}, "required": []string{"status", "error"}},
			"then": map[string]any{"properties": map[string]any{"changed_item_count": map[string]any{"minimum": 1}, "validated": map[string]any{"const": false}, "validation_source": map[string]any{"const": "firebase"}}},
		},
		map[string]any{
			"if":   map[string]any{"properties": map[string]any{"status": map[string]any{"const": "conflict"}, "error": errorAt("publication")}, "required": []string{"status", "error"}},
			"then": map[string]any{"properties": map[string]any{"changed_item_count": map[string]any{"minimum": 1}, "validated": map[string]any{"const": true}, "validation_source": map[string]any{"const": "firebase"}}},
		},
	)
	for _, status := range []string{"would-draft", "would-publish"} {
		constraints = append(constraints, fieldValueConstraint("status", status, map[string]any{"dry_run": map[string]any{"const": true}}))
	}
	for _, status := range []string{"published", "published-cache-failed", "published-hook-failed", "drafted"} {
		constraints = append(constraints, fieldValueConstraint("status", status, map[string]any{"dry_run": map[string]any{"const": false}}))
	}
	for _, status := range []string{"drafted", "would-draft", "draft-failed"} {
		constraints = append(constraints, fieldValueConstraint("status", status, map[string]any{"draft": map[string]any{"const": true}}))
	}
	for _, status := range []string{"published", "published-cache-failed", "published-hook-failed", "would-publish", "preparation-failed", "validation-failed", "conflict", "publish-failed"} {
		constraints = append(constraints, fieldValueConstraint("status", status, map[string]any{"draft": map[string]any{"const": false}}))
	}
	schema["allOf"] = constraints
	schema["x-fbrcm-invariants"] = []any{
		invariant("implies", "if", invariant("eq", "left", invariantField("no_op_reason"), "right", invariantConst("no_match")), "then", invariant("eq", "left", invariantField("selection.matched_item_count"), "right", invariantConst(0))),
		invariant("implies", "if", invariant("eq", "left", invariantField("no_op_reason"), "right", invariantConst("already_applied")), "then", invariant("gt", "left", invariantField("selection.matched_item_count"), "right", invariantConst(0))),
		invariant("iff", "left", invariant("and", "values", []any{
			invariant("in", "value", invariantField("status"), "set", []any{"preparation-failed", "validation-failed", "conflict", "publish-failed", "draft-failed"}),
			invariant("is_non_null", "value", invariantField("error")),
		}), "right", invariant("is_non_null", "value", invariantField("retry_selector"))),
	}
}

func caseInsensitiveValuesPattern(values []string) string {
	patterns := make([]string, 0, len(values))
	for _, value := range values {
		var pattern strings.Builder
		for _, char := range value {
			if char >= 'A' && char <= 'Z' {
				lower := char + ('a' - 'A')
				_, _ = fmt.Fprintf(&pattern, "[%c%c]", char, lower)
			} else {
				pattern.WriteRune(char)
			}
		}
		patterns = append(patterns, pattern.String())
	}
	return "^(?:" + strings.Join(patterns, "|") + ")$"
}

func capabilityIndexSchema() map[string]any {
	summary := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"id", "path", "summary", "invocation_schema", "response_schema", "side_effect_level", "destructive"},
		"properties": map[string]any{
			"id":                map[string]any{"type": "string", "minLength": 1},
			"path":              map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"summary":           map[string]any{"type": "string"},
			"invocation_schema": map[string]any{"type": "string", "pattern": `^urn:fbrcm:schema:cli:`},
			"response_schema":   map[string]any{"type": "string", "pattern": `^urn:fbrcm:schema:cli:`},
			"side_effect_level": map[string]any{"type": "integer", "minimum": 0, "maximum": 3},
			"destructive":       map[string]any{"type": "boolean"},
		},
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"contract_version", "count", "commands"},
		"x-fbrcm-invariants": []any{
			invariant("eq", "left", invariantField("count"), "right", invariant("length", "value", invariantField("commands"))),
		},
		"properties": map[string]any{
			"contract_version": map[string]any{"const": Version},
			"count":            map[string]any{"type": "integer", "minimum": 0},
			"commands":         map[string]any{"type": "array", "items": summary},
		},
	}
}

func remoteConfigValueSchema() map[string]any {
	objectValue := func(name string, value any) map[string]any {
		return map[string]any{
			"type": "object", "additionalProperties": false, "required": []string{name},
			"properties": map[string]any{name: value},
		}
	}
	managed := map[string]any{"type": "object"}
	known := []string{"value", "useInAppDefault", "personalizationValue", "experimentValue", "rolloutValue"}
	return map[string]any{
		"description": "Firebase RemoteConfigParameterValue union",
		"oneOf": []any{
			objectValue("value", map[string]any{"type": "string"}),
			objectValue("useInAppDefault", map[string]any{"const": true}),
			objectValue("personalizationValue", managed),
			objectValue("experimentValue", managed),
			objectValue("rolloutValue", managed),
			map[string]any{
				"type": "object", "minProperties": 1, "maxProperties": 1,
				"propertyNames": map[string]any{"not": map[string]any{"enum": known}},
				"description":   "one unknown future value option preserved as opaque JSON",
			},
		},
	}
}

func artifactDataSchema() map[string]any {
	nullValue := map[string]any{"type": "null"}
	properties := map[string]any{
		"target":         map[string]any{"type": []string{"string", "null"}},
		"media_type":     map[string]any{"type": "string", "minLength": 1},
		"encoding":       map[string]any{"$ref": SemanticRef("artifact_encoding")},
		"json_content":   map[string]any{},
		"text_content":   map[string]any{"type": []string{"string", "null"}},
		"base64_content": map[string]any{"type": []string{"string", "null"}},
		"destination":    map[string]any{"type": []string{"string", "null"}},
		"size_bytes":     map[string]any{"type": "integer", "minimum": 0},
		"sha256":         map[string]any{"type": "string", "pattern": `^[0-9a-f]{64}$`},
		"overwritten":    map[string]any{"type": "boolean"},
	}
	variant := func(encoding string, fields map[string]any) map[string]any {
		fields["encoding"] = map[string]any{"const": encoding}
		return map[string]any{"properties": fields}
	}
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"required": []string{"target", "media_type", "encoding", "json_content", "text_content", "base64_content", "destination", "size_bytes", "sha256", "overwritten"},
		"x-fbrcm-invariants": []any{
			invariant("eq", "left", invariantField("size_bytes"), "right", invariant("byte_length", "value", map[string]any{"symbol": "canonical_artifact_bytes"})),
			invariant("eq", "left", invariantField("sha256"), "right", invariant("lowercase_hex", "value", invariant("sha256", "value", map[string]any{"symbol": "canonical_artifact_bytes"}))),
			invariant("iff", "left", invariantField("overwritten"), "right", invariant("and", "values", []any{
				invariant("eq", "left", invariantField("encoding"), "right", invariantConst("none")),
				map[string]any{"symbol": "existing_destination_replaced"},
			})),
		},
		"properties": properties,
		"oneOf": []any{
			variant("json", map[string]any{"text_content": nullValue, "base64_content": nullValue, "destination": nullValue, "overwritten": map[string]any{"const": false}}),
			variant("utf-8", map[string]any{"json_content": nullValue, "text_content": map[string]any{"type": "string"}, "base64_content": nullValue, "destination": nullValue, "overwritten": map[string]any{"const": false}}),
			variant("base64", map[string]any{"json_content": nullValue, "text_content": nullValue, "base64_content": map[string]any{"type": "string", "contentEncoding": "base64"}, "destination": nullValue, "overwritten": map[string]any{"const": false}}),
			variant("none", map[string]any{"json_content": nullValue, "text_content": nullValue, "base64_content": nullValue, "destination": map[string]any{"type": "string", "minLength": 1}}),
		},
	}
}

func semanticStringRef(typeOf reflect.Type) string {
	switch typeOf.PkgPath() + "." + typeOf.Name() {
	case "github.com/yumauri/fbrcm/cli/shared/rc.RemoteMutationStatus":
		return SemanticRef("remote_mutation_status")
	case "github.com/yumauri/fbrcm/cli/shared/rc.NoOpReason":
		return SemanticRef("no_op_reason")
	default:
		return ""
	}
}

func semanticStringEnum(typeOf reflect.Type) []string {
	switch typeOf.PkgPath() + "." + typeOf.Name() {
	case "github.com/yumauri/fbrcm/cli/shared/rc.RemoteMutationStatus":
		return []string{"unchanged", "preparation-failed", "published", "validation-failed", "conflict", "publish-failed", "published-cache-failed", "published-hook-failed", "drafted", "would-draft", "would-publish", "draft-failed"}
	case "github.com/yumauri/fbrcm/cli/shared/rc.NoOpReason":
		return []string{"no_match", "already_applied"}
	case "github.com/yumauri/fbrcm/core/config.ProjectAliasSource":
		return []string{
			string(coreconfig.ProjectAliasSourceFBRCM),
			string(coreconfig.ProjectAliasSourceFirebase),
			string(coreconfig.ProjectAliasSourceBoth),
		}
	case "github.com/yumauri/fbrcm/core/rc/diff.ChangeKind":
		return []string{"added", "removed", "changed", "unchanged"}
	case "github.com/yumauri/fbrcm/core/rc/diff.ItemKind":
		return []string{"condition", "parameter", "group_description"}
	default:
		return nil
	}
}

func applyFieldSemantics(schema map[string]any, field reflect.StructField, name string) {
	if schema["type"] == "integer" && (name == "count" || strings.HasSuffix(name, "_count") || slices.Contains([]string{"added", "bound", "bytes_deleted", "changed", "index", "removed", "selected", "size", "skipped", "snapshots_deleted", "targets_affected", "unchanged"}, name)) {
		schema["minimum"] = 0
	}
	if name == "validation_source" {
		clear(schema)
		schema["$ref"] = SemanticRef("validation_source")
	}
	if name == "retry_selector" {
		schema["pattern"] = `^(?:server@)?=.+$`
		schema["x-fbrcm-grammar"] = "exact target-aware retry selector"
	}
	contractTag := field.Tag.Get("contract")
	for part := range strings.SplitSeq(contractTag, ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "enum":
			schema["enum"] = strings.Split(value, "|")
			appendNullEnum(schema)
		case "format":
			schema["format"] = value
		case "pattern":
			schema["pattern"] = value
		case "ref":
			clear(schema)
			schema["$ref"] = SemanticRef(value)
		}
	}
}

// SemanticSchemaID identifies the reusable semantic definitions shared by
// command input and response schemas.
func SemanticSchemaID() string {
	return "urn:fbrcm:schema:cli:" + Version + ":semantic"
}

// SemanticRef returns an absolute reference that remains valid when a DTO
// schema is nested under an envelope's data property.
func SemanticRef(name string) string {
	return SemanticSchemaID() + "#/$defs/" + name
}

func nullableSchema(schema map[string]any) map[string]any {
	if kind, ok := schema["type"].(string); ok {
		clone := make(map[string]any, len(schema))
		maps.Copy(clone, schema)
		clone["type"] = []string{kind, "null"}
		appendNullEnum(clone)
		return clone
	}
	if kinds, ok := schema["type"].([]string); ok {
		if slices.Contains(kinds, "null") {
			return schema
		}
		clone := make(map[string]any, len(schema))
		maps.Copy(clone, schema)
		clone["type"] = append(append([]string(nil), kinds...), "null")
		appendNullEnum(clone)
		return clone
	}
	return map[string]any{"oneOf": []any{schema, map[string]any{"type": "null"}}}
}

func appendNullEnum(schema map[string]any) {
	kinds, ok := schema["type"].([]string)
	if !ok || !slices.Contains(kinds, "null") {
		return
	}
	switch values := schema["enum"].(type) {
	case []string:
		result := make([]any, 0, len(values)+1)
		for _, value := range values {
			result = append(result, value)
		}
		result = append(result, nil)
		schema["enum"] = result
	case []any:
		if !slices.Contains(values, nil) {
			schema["enum"] = append(values, nil)
		}
	}
}

func parseJSONTag(tag string) (string, map[string]bool) {
	parts := strings.Split(tag, ",")
	options := make(map[string]bool, max(len(parts)-1, 0))
	for _, option := range parts[1:] {
		options[option] = true
	}
	return parts[0], options
}

func mergeObjectSchema(properties map[string]any, required *[]string, schema map[string]any) {
	embeddedProperties, _ := schema["properties"].(map[string]any)
	maps.Copy(properties, embeddedProperties)
	if embeddedRequired, ok := schema["required"].([]string); ok {
		*required = append(*required, embeddedRequired...)
	}
}
