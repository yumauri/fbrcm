package main

import (
	"encoding/json"
	"fmt"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/yumauri/fbrcm/cli/contract"
)

type extensionValidators struct {
	validation     *jsonschema.Schema
	normalization  *jsonschema.Schema
	matching       *jsonschema.Schema
	invariants     *jsonschema.Schema
	inputSelection *jsonschema.Schema
}

func newExtensionValidators(semantic map[string]any) (extensionValidators, error) {
	raw, err := json.Marshal(semantic)
	if err != nil {
		return extensionValidators{}, fmt.Errorf("encode semantic schema: %w", err)
	}
	var normalized any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		return extensionValidators{}, fmt.Errorf("normalize semantic schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(contract.SemanticSchemaID(), normalized); err != nil {
		return extensionValidators{}, fmt.Errorf("register semantic schema: %w", err)
	}
	compile := func(name string) (*jsonschema.Schema, error) {
		compiled, err := compiler.Compile(contract.SemanticRef(name))
		if err != nil {
			return nil, fmt.Errorf("compile %s extension schema: %w", name, err)
		}
		return compiled, nil
	}
	validation, err := compile("validation_rules")
	if err != nil {
		return extensionValidators{}, err
	}
	normalization, err := compile("normalization_rules")
	if err != nil {
		return extensionValidators{}, err
	}
	matching, err := compile("matching_rules")
	if err != nil {
		return extensionValidators{}, err
	}
	invariants, err := compile("invariant_rules")
	if err != nil {
		return extensionValidators{}, err
	}
	inputSelection, err := compile("input_selection_rules")
	if err != nil {
		return extensionValidators{}, err
	}
	return extensionValidators{validation: validation, normalization: normalization, matching: matching, invariants: invariants, inputSelection: inputSelection}, nil
}

func (v extensionValidators) Validate(document any) error {
	return walkExtensionAnnotations(document, "$", v)
}

func walkExtensionAnnotations(value any, path string, validators extensionValidators) error {
	switch typed := value.(type) {
	case map[string]any:
		for name, child := range typed {
			var validator *jsonschema.Schema
			switch name {
			case "x-fbrcm-validation":
				validator = validators.validation
			case "x-fbrcm-normalization":
				validator = validators.normalization
			case "x-fbrcm-matching":
				validator = validators.matching
			case "x-fbrcm-invariants":
				validator = validators.invariants
			case "x-fbrcm-input-selection":
				validator = validators.inputSelection
			}
			if validator != nil {
				if err := validator.Validate(child); err != nil {
					return fmt.Errorf("%s.%s: %w", path, name, err)
				}
			}
			if err := walkExtensionAnnotations(child, path+"."+name, validators); err != nil {
				return err
			}
		}
	case []any:
		for index, child := range typed {
			if err := walkExtensionAnnotations(child, fmt.Sprintf("%s[%d]", path, index), validators); err != nil {
				return err
			}
		}
	case json.RawMessage:
		return nil
	}
	return nil
}

func extensionSchemaDefinitions() map[string]any {
	ref := func(name string) map[string]any { return map[string]any{"$ref": contract.SemanticRef(name)} }
	object := func(required []string, properties map[string]any) map[string]any {
		return map[string]any{"type": "object", "additionalProperties": false, "required": required, "properties": properties}
	}
	stringValue := map[string]any{"type": "string", "minLength": 1}
	validationRule := ref("validation_rule")
	validationCases := map[string]any{"type": "object", "minProperties": 1, "additionalProperties": validationRule}
	validationRules := []any{
		object([]string{"operator"}, map[string]any{"operator": map[string]any{"const": "accept"}}),
		object([]string{"operator", "language", "version", "result_type"}, map[string]any{"operator": map[string]any{"const": "compile_expression"}, "language": stringValue, "version": stringValue, "result_type": stringValue}),
		object([]string{"operator", "field", "normalized_case", "cases"}, map[string]any{"operator": map[string]any{"const": "dispatch_by_field"}, "field": stringValue, "normalized_case": map[string]any{"enum": []string{"lowercase"}}, "cases": validationCases}),
		object([]string{"operator", "fields", "comparison"}, map[string]any{"operator": map[string]any{"const": "fields_differ"}, "fields": map[string]any{"type": "array", "minItems": 2, "maxItems": 2, "items": stringValue}, "comparison": map[string]any{"enum": []string{"exact_codepoint", "unicode_case_fold"}}}),
		object([]string{"operator"}, map[string]any{"operator": map[string]any{"const": "match_schema_pattern"}}),
		object([]string{"operator", "parser", "require_exact"}, map[string]any{"operator": map[string]any{"const": "parse_email"}, "parser": map[string]any{"const": "net/mail.ParseAddress"}, "require_exact": map[string]any{"const": true}}),
		object([]string{"operator", "specification", "consume"}, map[string]any{"operator": map[string]any{"const": "parse_json"}, "specification": stringValue, "consume": map[string]any{"enum": []string{"entire_string"}}}),
		object([]string{"operator", "parser", "require_positive"}, map[string]any{"operator": map[string]any{"const": "parse_duration"}, "parser": map[string]any{"const": "time.ParseDuration"}, "require_positive": map[string]any{"const": true}}),
		object([]string{"operator", "parser", "normalization", "require_absolute"}, map[string]any{"operator": map[string]any{"const": "parse_uri"}, "parser": map[string]any{"const": "net/url.ParseRequestURI"}, "normalization": map[string]any{"const": "trim_unicode_whitespace"}, "require_absolute": map[string]any{"const": true}}),
		object([]string{"operator", "allow_empty"}, map[string]any{"operator": map[string]any{"const": "reject_raw_whitespace_only"}, "allow_empty": map[string]any{"const": true}}),
		object([]string{"operator", "specification"}, map[string]any{"operator": map[string]any{"const": "parse_time"}, "specification": map[string]any{"const": "Go time.RFC3339"}}),
		object([]string{"operator", "maximum_bytes"}, map[string]any{"operator": map[string]any{"const": "theme_import_source"}, "maximum_bytes": map[string]any{"const": 1048576}}),
		object([]string{"operator", "absolute_parser", "relative_parser", "maximum_relative_distance"}, map[string]any{
			"operator":                  map[string]any{"const": "parse_version_selector"},
			"absolute_parser":           map[string]any{"const": "strconv.ParseInt base 10 bitSize 64; require result > 0"},
			"relative_parser":           map[string]any{"const": "strconv.ParseInt base 10 bitSize 32; require result > 0"},
			"maximum_relative_distance": map[string]any{"const": 299},
		}),
		object([]string{"operator", "parser", "minimum"}, map[string]any{
			"operator": map[string]any{"const": "parse_positive_integer"},
			"parser":   map[string]any{"const": "strconv.Atoi"},
			"minimum":  map[string]any{"const": 1},
		}),
		object([]string{"operator", "operation", "project_argument", "maximum", "zero_behavior"}, map[string]any{
			"operator":         map[string]any{"const": "condition_priority"},
			"operation":        map[string]any{"const": "add"},
			"project_argument": map[string]any{"const": "arguments.project"},
			"maximum":          map[string]any{"const": "resolved_condition_count_plus_one"},
			"zero_behavior":    map[string]any{"const": "append"},
		}),
		object([]string{"operator", "operation", "project_argument", "maximum"}, map[string]any{
			"operator":         map[string]any{"const": "condition_priority"},
			"operation":        map[string]any{"const": "move"},
			"project_argument": map[string]any{"const": "arguments.project"},
			"maximum":          map[string]any{"const": "resolved_condition_count"},
		}),
		object([]string{"operator", "validator"}, map[string]any{"operator": map[string]any{"const": "local_validate"}, "validator": map[string]any{"enum": []string{"firebase.NormalizeRemoteConfigForUpdate", "firebase.PrepareRemoteConfigUpdate"}}}),
		object([]string{"operator", "validator", "checks"}, map[string]any{
			"operator":  map[string]any{"const": "publication_plan_integrity"},
			"validator": map[string]any{"const": "publication.Validate"},
			"checks": map[string]any{
				"type": "array", "minItems": 7, "maxItems": 7, "uniqueItems": true,
				"items": map[string]any{"enum": []string{"nonempty_targets", "canonical_target_order", "unique_target_ids", "snapshot_sha256", "none_snapshot_equality", "action_validation_provenance", "content_derived_plan_id"}},
			},
		}),
		object([]string{"operator", "service", "grammar"}, map[string]any{"operator": map[string]any{"const": "remote_validate"}, "service": stringValue, "grammar": stringValue}),
		object([]string{"operator", "collection", "project_argument", "accepted_forms"}, map[string]any{
			"operator":         map[string]any{"const": "managed_feature_id"},
			"collection":       map[string]any{"enum": []string{"experiments", "rollouts"}},
			"project_argument": stringValue,
			"accepted_forms": map[string]any{
				"type": "array", "minItems": 2, "maxItems": 2, "uniqueItems": true,
				"items": map[string]any{"enum": []string{"bare_id", "resolved_project_resource_name"}},
			},
		}),
		object([]string{"operator", "field"}, map[string]any{"operator": map[string]any{"const": "unique_by"}, "field": stringValue}),
		object([]string{"operator", "separator", "range"}, map[string]any{"operator": map[string]any{"const": "unique_tokens"}, "separator": stringValue, "range": map[string]any{"enum": []string{"all_but_last"}}}),
	}

	invariantExpression := ref("invariant_expression")
	expressionArray := map[string]any{"type": "array", "minItems": 1, "items": invariantExpression}
	binary := func(operator, left, right string) map[string]any {
		return object([]string{"op", left, right}, map[string]any{"op": map[string]any{"const": operator}, left: invariantExpression, right: invariantExpression})
	}
	unary := func(operator, operand string) map[string]any {
		return object([]string{"op", operand}, map[string]any{"op": map[string]any{"const": operator}, operand: invariantExpression})
	}
	invariantRules := []any{
		object([]string{"const"}, map[string]any{"const": map[string]any{}}),
		object([]string{"field"}, map[string]any{"field": stringValue}),
		object([]string{"symbol"}, map[string]any{"symbol": map[string]any{"enum": []string{"canonical_artifact_bytes", "existing_destination_replaced"}}}),
		object([]string{"op", "values"}, map[string]any{"op": map[string]any{"const": "and"}, "values": expressionArray}),
		unary("byte_length", "value"),
		object([]string{"op", "collection", "where"}, map[string]any{"op": map[string]any{"const": "count_where"}, "collection": invariantExpression, "where": invariantExpression}),
		binary("eq", "left", "right"),
		binary("gt", "left", "right"),
		binary("iff", "left", "right"),
		binary("implies", "if", "then"),
		object([]string{"op", "value", "set"}, map[string]any{"op": map[string]any{"const": "in"}, "value": invariantExpression, "set": map[string]any{"type": "array"}}),
		unary("is_non_null", "value"),
		unary("length", "value"),
		unary("lowercase_hex", "value"),
		unary("sha256", "value"),
		object([]string{"op", "values"}, map[string]any{"op": map[string]any{"const": "sum"}, "values": expressionArray}),
	}
	firstAvailableInputSelectionRule := object([]string{"operator", "sources", "on_missing", "later_sources"}, map[string]any{
		"operator": map[string]any{"const": "first_available"},
		"sources": map[string]any{
			"oneOf": []any{
				map[string]any{"type": "array", "minItems": 2, "maxItems": 2, "prefixItems": []any{map[string]any{"const": "options.from"}, map[string]any{"const": "stdin.document"}}, "items": false},
				map[string]any{"type": "array", "minItems": 2, "maxItems": 2, "prefixItems": []any{map[string]any{"const": "arguments.source"}, map[string]any{"const": "stdin.document"}}, "items": false},
			},
		},
		"on_missing":    map[string]any{"const": "interaction.required"},
		"later_sources": map[string]any{"const": "ignored_without_consumption"},
	})
	pathOrStdinInputSelectionRule := object([]string{"operator", "path", "stdin", "stdin_path", "non_stdin", "unused_stdin", "missing_stdin"}, map[string]any{
		"operator":      map[string]any{"const": "path_or_stdin_document"},
		"path":          map[string]any{"const": "arguments.plan"},
		"stdin":         map[string]any{"const": "stdin.document"},
		"stdin_path":    map[string]any{"const": "-"},
		"non_stdin":     map[string]any{"const": "read_file"},
		"unused_stdin":  map[string]any{"const": "ignored_without_consumption"},
		"missing_stdin": map[string]any{"const": "plan.invalid"},
	})

	return map[string]any{
		"validation_rule":  map[string]any{"oneOf": validationRules},
		"validation_rules": map[string]any{"type": "array", "minItems": 1, "items": validationRule},
		"normalization_rule": map[string]any{"oneOf": []any{
			object([]string{"operator", "source", "target"}, map[string]any{
				"operator": map[string]any{"enum": []string{"trim_unicode_whitespace", "lowercase", "uppercase", "deduplicate_preserve_first", "canonicalize_target_selector", "canonicalize_positional_target_selector"}},
				"source":   stringValue,
				"target":   stringValue,
			}),
			object([]string{"operator", "source", "target", "prefixes"}, map[string]any{
				"operator": map[string]any{"const": "trim_unicode_whitespace_if_prefix"},
				"source":   stringValue,
				"target":   stringValue,
				"prefixes": map[string]any{"type": "array", "minItems": 1, "uniqueItems": true, "items": stringValue},
			}),
			object([]string{"operator", "source", "target", "order"}, map[string]any{
				"operator": map[string]any{"const": "sort_by_declared_order"},
				"source":   stringValue,
				"target":   stringValue,
				"order":    map[string]any{"type": "array", "minItems": 1, "uniqueItems": true, "items": stringValue},
			}),
		}},
		"normalization_rules": map[string]any{"type": "array", "minItems": 1, "items": ref("normalization_rule")},
		"matching_rule": map[string]any{"oneOf": []any{
			object([]string{"operator", "comparison", "default_template", "lookup"}, map[string]any{
				"operator":         map[string]any{"const": "literal_project_id"},
				"comparison":       map[string]any{"const": "exact_case_sensitive"},
				"default_template": map[string]any{"const": "client"},
				"lookup":           map[string]any{"const": false},
			}),
			object([]string{"operator", "fields", "query_normalization", "default_mode", "mode_prefixes", "comparison"}, map[string]any{
				"operator":            map[string]any{"const": "mode_prefixed_query"},
				"fields":              map[string]any{"type": "array", "minItems": 1, "uniqueItems": true, "items": stringValue},
				"query_normalization": map[string]any{"const": "trim_unicode_whitespace"},
				"default_mode":        map[string]any{"const": "fuzzy"},
				"mode_prefixes": object([]string{"~", "^", "/", "="}, map[string]any{
					"~": map[string]any{"const": "fuzzy"},
					"^": map[string]any{"const": "starts-with"},
					"/": map[string]any{"const": "includes"},
					"=": map[string]any{"const": "exact"},
				}),
				"comparison": map[string]any{"const": "unicode_case_insensitive"},
				"target_prefixes": map[string]any{
					"type": "array", "minItems": 2, "maxItems": 2, "uniqueItems": true,
					"items": map[string]any{"enum": []string{"client", "server"}},
				},
				"unqualified_target_selection": map[string]any{"enum": []string{
					"all_configured_enabled_templates",
					"client_template",
					"existing_drafts_in_configured_enabled_templates_or_client_fallback",
				}},
				"explicit_target_selection":      map[string]any{"const": "single_named_template"},
				"client_target_canonicalization": map[string]any{"const": "unqualified_project_id"},
			}),
			object([]string{"operator", "fields", "query_normalization", "comparison", "precedence"}, map[string]any{
				"operator":            map[string]any{"const": "project_positional_resolution"},
				"fields":              map[string]any{"type": "array", "minItems": 1, "uniqueItems": true, "items": stringValue},
				"query_normalization": map[string]any{"const": "preserve_argv"},
				"comparison":          map[string]any{"const": "exact_case_sensitive"},
				"precedence": map[string]any{
					"type":     "array",
					"minItems": 3,
					"maxItems": 3,
					"prefixItems": []any{
						map[string]any{"const": "exact_project_id"},
						map[string]any{"const": "exact_repository_alias"},
						map[string]any{"const": "exact_display_name"},
					},
					"items": false,
				},
				"target_prefixes": map[string]any{
					"type": "array", "minItems": 2, "maxItems": 2, "uniqueItems": true,
					"items": map[string]any{"enum": []string{"client", "server"}},
				},
				"unqualified_target_selection":   map[string]any{"const": "configured_primary_template"},
				"explicit_target_selection":      map[string]any{"const": "single_named_template"},
				"client_target_canonicalization": map[string]any{"const": "unqualified_project_id"},
			}),
			object([]string{"operator", "candidate_source", "fields", "query_normalization", "comparison", "precedence", "target_prefixes", "unqualified_target_selection", "explicit_target_selection", "client_target_canonicalization", "zero_result", "multiple_result"}, map[string]any{
				"operator":            map[string]any{"const": "draft_resolution"},
				"candidate_source":    map[string]any{"const": "local_draft_target_ids"},
				"fields":              map[string]any{"type": "array", "minItems": 3, "maxItems": 3, "uniqueItems": true, "items": stringValue},
				"query_normalization": map[string]any{"const": "preserve_argv"},
				"comparison":          map[string]any{"const": "exact_case_sensitive"},
				"precedence": map[string]any{
					"type": "array", "minItems": 3, "maxItems": 3,
					"prefixItems": []any{
						map[string]any{"const": "exact_draft_project_id"},
						map[string]any{"const": "exact_repository_alias"},
						map[string]any{"const": "exact_display_name"},
					},
					"items": false,
				},
				"target_prefixes": map[string]any{
					"type": "array", "minItems": 2, "maxItems": 2, "uniqueItems": true,
					"items": map[string]any{"enum": []string{"client", "server"}},
				},
				"unqualified_target_selection":   map[string]any{"const": "configured_primary_template_or_client_fallback"},
				"explicit_target_selection":      map[string]any{"const": "single_named_template"},
				"client_target_canonicalization": map[string]any{"const": "unqualified_project_id"},
				"zero_result":                    map[string]any{"const": "draft.not_found"},
				"multiple_result":                map[string]any{"const": "draft.ambiguous"},
			}),
			object([]string{"operator"}, map[string]any{"operator": map[string]any{"enum": []string{
				"auth_id_resolution",
				"condition_name_resolution",
				"condition_positional_resolution",
				"duplicate_source_resolution",
				"group_name_resolution",
				"help_path_resolution",
				"parameter_argument_resolution",
				"personalization_id_resolution",
				"profile_name_resolution",
				"theme_name_resolution",
				"project_alias_resolution",
				"schema_id_resolution",
				"version_resolution",
			}}}),
			object([]string{"operator", "candidate_source", "comparison", "omitted_result", "reserved_root_token", "unknown_result", "non_executable_result"}, map[string]any{
				"operator":              map[string]any{"const": "command_path_resolution"},
				"candidate_source":      map[string]any{"const": "executable_commands"},
				"comparison":            map[string]any{"const": "exact_case_sensitive_argv_components"},
				"omitted_result":        map[string]any{"const": "capability_index"},
				"reserved_root_token":   map[string]any{"const": "root"},
				"unknown_result":        map[string]any{"const": "command.not_found"},
				"non_executable_result": map[string]any{"const": "command.not_executable"},
			}),
			object([]string{"operator", "project_source", "all_source", "composition", "canonical_order"}, map[string]any{
				"operator":        map[string]any{"const": "draft_batch_selection"},
				"project_source":  map[string]any{"const": "arguments.project"},
				"all_source":      map[string]any{"const": "options.all"},
				"composition":     map[string]any{"const": "exactly_one_source"},
				"canonical_order": map[string]any{"const": "deduplicate_then_sort_target_id"},
			}),
			object([]string{"operator", "sources", "repeated_source_combination", "across_source_combination", "absent_source_behavior", "target_defaults"}, map[string]any{
				"operator":                    map[string]any{"const": "selection_composition"},
				"sources":                     map[string]any{"type": "array", "minItems": 1, "uniqueItems": true, "items": map[string]any{"type": "string", "pattern": `^(?:arguments|options)\.[a-z][a-z0-9_-]*$`}},
				"repeated_source_combination": map[string]any{"const": "or"},
				"across_source_combination":   map[string]any{"const": "and"},
				"absent_source_behavior":      map[string]any{"const": "match_all"},
				"target_defaults": map[string]any{"type": "array", "uniqueItems": true, "items": object([]string{"source", "selection"}, map[string]any{
					"source":    map[string]any{"type": "string", "pattern": `^options\.[a-z][a-z0-9_-]*$`},
					"selection": map[string]any{"const": "all_configured_projects_enabled_templates"},
				})},
			}),
			object([]string{"operator", "fields", "query_normalization", "haystack_normalization", "separator"}, map[string]any{
				"operator":               map[string]any{"const": "case_insensitive_substring"},
				"fields":                 map[string]any{"type": "array", "minItems": 1, "uniqueItems": true, "items": stringValue},
				"query_normalization":    map[string]any{"const": "trim_then_unicode_lowercase"},
				"haystack_normalization": map[string]any{"const": "unicode_lowercase"},
				"separator":              map[string]any{"const": "\n"},
			}),
			object([]string{"operator", "normalized_fields", "raw_fields", "normalized_query", "raw_query", "match", "combination"}, map[string]any{
				"operator":          map[string]any{"const": "parameter_search"},
				"normalized_fields": map[string]any{"type": "array", "minItems": 1, "uniqueItems": true, "items": stringValue},
				"raw_fields":        map[string]any{"type": "array", "minItems": 1, "uniqueItems": true, "items": stringValue},
				"normalized_query":  map[string]any{"const": "lowercase_alphanumeric_words"},
				"raw_query":         map[string]any{"const": "collapse_unicode_whitespace"},
				"match":             map[string]any{"const": "substring"},
				"combination":       map[string]any{"const": "or"},
			}),
		}},
		"matching_rules":        map[string]any{"type": "array", "minItems": 1, "items": ref("matching_rule")},
		"invariant_expression":  map[string]any{"oneOf": invariantRules},
		"invariant_rules":       map[string]any{"type": "array", "minItems": 1, "items": invariantExpression},
		"input_selection_rule":  map[string]any{"oneOf": []any{firstAvailableInputSelectionRule, pathOrStdinInputSelectionRule}},
		"input_selection_rules": map[string]any{"type": "array", "minItems": 1, "items": ref("input_selection_rule")},
	}
}
