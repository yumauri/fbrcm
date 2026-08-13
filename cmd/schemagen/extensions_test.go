package main

import "testing"

func TestExtensionValidatorsAcceptPublishedRuleShapes(t *testing.T) {
	validators, err := newExtensionValidators(semanticSchema())
	if err != nil {
		t.Fatal(err)
	}
	document := map[string]any{
		"x-fbrcm-validation": []any{
			map[string]any{"operator": "unique_by", "field": "name"},
			map[string]any{"operator": "remote_validate", "service": "firebase_remote_config", "grammar": "template"},
			map[string]any{"operator": "fields_differ", "fields": []any{"arguments.source", "arguments.target"}, "comparison": "unicode_case_fold"},
			map[string]any{"operator": "parse_email", "parser": "net/mail.ParseAddress", "require_exact": true},
			map[string]any{"operator": "parse_uri", "parser": "net/url.ParseRequestURI", "normalization": "trim_unicode_whitespace", "require_absolute": true},
			map[string]any{"operator": "parse_time", "specification": "Go time.RFC3339"},
			map[string]any{
				"operator": "condition_priority", "operation": "add", "project_argument": "arguments.project",
				"maximum": "resolved_condition_count_plus_one", "zero_behavior": "append",
			},
		},
		"x-fbrcm-normalization": []any{
			map[string]any{"operator": "trim_unicode_whitespace", "source": "argv", "target": "normalized_invocation"},
			map[string]any{"operator": "trim_unicode_whitespace_if_prefix", "source": "argv", "target": "normalized_invocation", "prefixes": []any{"keys."}},
			map[string]any{"operator": "canonicalize_positional_target_selector", "source": "argv", "target": "normalized_invocation"},
		},
		"x-fbrcm-matching": []any{
			map[string]any{
				"operator": "mode_prefixed_query", "fields": []any{"project_id", "display_name"},
				"query_normalization": "trim_unicode_whitespace", "default_mode": "fuzzy",
				"mode_prefixes": map[string]any{"~": "fuzzy", "^": "starts-with", "/": "includes", "=": "exact"},
				"comparison":    "unicode_case_insensitive",
			},
			map[string]any{
				"operator": "parameter_search", "normalized_fields": []any{"name"}, "raw_fields": []any{"value"},
				"normalized_query": "lowercase_alphanumeric_words", "raw_query": "collapse_unicode_whitespace",
				"match": "substring", "combination": "or",
			},
			map[string]any{
				"operator": "selection_composition", "sources": []any{"options.project", "options.filter"},
				"repeated_source_combination": "or", "across_source_combination": "and", "absent_source_behavior": "match_all",
				"target_defaults": []any{map[string]any{"source": "options.project", "selection": "all_configured_projects_enabled_templates"}},
			},
			map[string]any{"operator": "group_name_resolution"},
			map[string]any{"operator": "help_path_resolution"},
			map[string]any{
				"operator": "command_path_resolution", "candidate_source": "executable_commands",
				"comparison": "exact_case_sensitive_argv_components", "omitted_result": "capability_index",
				"reserved_root_token": "root", "unknown_result": "command.not_found", "non_executable_result": "command.not_executable",
			},
			map[string]any{
				"operator": "draft_resolution", "candidate_source": "local_draft_target_ids",
				"fields": []any{"project_id", "display_name", "repository_aliases"}, "query_normalization": "preserve_argv",
				"comparison": "exact_case_sensitive", "precedence": []any{"exact_draft_project_id", "exact_repository_alias", "exact_display_name"},
				"target_prefixes":              []any{"client", "server"},
				"unqualified_target_selection": "configured_primary_template_or_client_fallback", "explicit_target_selection": "single_named_template",
				"client_target_canonicalization": "unqualified_project_id", "zero_result": "draft.not_found", "multiple_result": "draft.ambiguous",
			},
			map[string]any{
				"operator": "draft_batch_selection", "project_source": "arguments.project", "all_source": "options.all",
				"composition": "exactly_one_source", "canonical_order": "deduplicate_then_sort_target_id",
			},
		},
		"x-fbrcm-invariants": []any{
			map[string]any{"op": "eq", "left": map[string]any{"field": "count"}, "right": map[string]any{"op": "length", "value": map[string]any{"field": "items"}}},
		},
		"x-fbrcm-input-selection": []any{
			map[string]any{
				"operator": "first_available", "sources": []any{"options.from", "stdin.document"},
				"on_missing": "interaction.required", "later_sources": "ignored_without_consumption",
			},
		},
	}
	if err := validators.Validate(document); err != nil {
		t.Fatalf("valid extension annotations rejected: %v", err)
	}
}

func TestExtensionValidatorsRejectOpaqueAndMalformedRules(t *testing.T) {
	validators, err := newExtensionValidators(semanticSchema())
	if err != nil {
		t.Fatal(err)
	}
	for name, document := range map[string]map[string]any{
		"opaque prose": {"x-fbrcm-validation": "validated by Firebase"},
		"unknown operator": {"x-fbrcm-validation": []any{
			map[string]any{"operator": "firebase_magic", "grammar": "template"},
		}},
		"missing operand": {"x-fbrcm-invariants": []any{
			map[string]any{"op": "eq", "left": map[string]any{"field": "count"}},
		}},
		"malformed matching": {"x-fbrcm-matching": []any{
			map[string]any{"operator": "parameter_search", "normalized_fields": []any{"name"}},
		}},
		"matching without fields": {"x-fbrcm-matching": []any{
			map[string]any{
				"operator": "mode_prefixed_query", "query_normalization": "trim_unicode_whitespace", "default_mode": "fuzzy",
				"mode_prefixes": map[string]any{"~": "fuzzy", "^": "starts-with", "/": "includes", "=": "exact"}, "comparison": "unicode_case_insensitive",
			},
		}},
		"malformed input selection": {"x-fbrcm-input-selection": []any{
			map[string]any{"operator": "first_available", "sources": []any{"stdin.document", "options.from"}},
		}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validators.Validate(document); err == nil {
				t.Fatalf("malformed annotations accepted: %#v", document)
			}
		})
	}
}
