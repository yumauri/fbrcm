package contract

// Explain publishes structural evidence rather than formatted terminal text.
// Keep its discriminated states closed and internally consistent.
func applyExplainTypeSemantics(name string, schema map[string]any) {
	props, _ := schema["properties"].(map[string]any)
	for _, v := range props {
		if p, ok := v.(map[string]any); ok {
			if _, array := p["items"]; array {
				p["type"] = "array"
			}
		}
	}
	null := map[string]any{"type": "null"}
	text := map[string]any{"type": "string"}
	nonempty := map[string]any{"type": "string", "minLength": 1}
	empty := map[string]any{"type": "array", "maxItems": 0}
	constraints := []any{}
	switch name {
	case "valueJSON":
		constraints = append(constraints, fieldValueConstraint("kind", "literal", map[string]any{"value": text, "managed_kind": null}), fieldValueConstraint("kind", "managed", map[string]any{"value": null, "managed_kind": nonempty}))
		for _, kind := range []string{"unknown", "in_app_default"} {
			constraints = append(constraints, fieldValueConstraint("kind", kind, map[string]any{"value": null, "managed_kind": null}))
		}
	case "selectionJSON":
		constraints = append(constraints, fieldValueConstraint("state", "unresolved", map[string]any{"source": null, "condition": null}), fieldValueConstraint("state", "resolved", map[string]any{"source": nonempty}), fieldValueConstraint("source", "condition", map[string]any{"condition": nonempty}), fieldValueConstraint("source", "default", map[string]any{"condition": null}))
	case "inputJSON":
		constraints = append(constraints, fieldValueConstraint("state", "known", map[string]any{"value": text, "observed_value": null, "time_min": null, "time_max": null, "reason_code": map[string]any{"const": "none"}}), fieldValueConstraint("state", "absent", map[string]any{"value": null, "observed_value": null}), fieldValueConstraint("state", "unknown", map[string]any{"value": null}), fieldValueConstraint("state", "bounded", map[string]any{"value": null, "observed_value": text, "reason_code": map[string]any{"const": "bounded_observation"}}))
	case "RuleEvaluation":
		constraints = append(constraints, fieldValueConstraint("status", "skipped", map[string]any{"evaluation": null, "reason_code": map[string]any{"const": "short_circuit"}, "inputs": empty}), fieldValueConstraint("status", "unsupported", map[string]any{"evaluation": map[string]any{"const": "unknown"}, "reason_code": map[string]any{"const": "unsupported_expression"}, "predicate": map[string]any{"properties": map[string]any{"kind": map[string]any{"const": "unsupported"}}}}), fieldValueConstraint("status", "evaluated", map[string]any{"evaluation": nonempty}))
	case "Predicate":
		constraints = append(constraints, fieldValueConstraint("kind", "literal", map[string]any{"literal": map[string]any{"type": "boolean"}, "field": null, "operator": null, "comparison": null, "expected": empty, "targets": empty}), fieldValueConstraint("kind", "comparison", map[string]any{"literal": null, "field": nonempty, "operator": nonempty, "comparison": nonempty, "expected": map[string]any{"minItems": 1}, "targets": empty}), fieldValueConstraint("kind", "client_targets", map[string]any{"literal": null, "field": map[string]any{"enum": []string{"app.browserAndVersion", "app.operatingSystemAndVersion"}}, "operator": map[string]any{"const": "inOne"}, "comparison": map[string]any{"const": "client"}, "expected": empty, "targets": map[string]any{"minItems": 1}}), fieldValueConstraint("kind", "unsupported", map[string]any{"literal": null, "field": null, "operator": null, "comparison": null, "expected": empty, "targets": empty}))
	case "conditionJSON":
		props["priority"].(map[string]any)["minimum"] = 1
		constraints = append(constraints, fieldValueConstraint("selection", "selected", map[string]any{"evaluation": map[string]any{"const": "true"}, "exclusion_reason": map[string]any{"const": "none"}}), fieldValueConstraint("exclusion_reason", "not_matched", map[string]any{"evaluation": map[string]any{"const": "false"}, "selection": map[string]any{"const": "not_selected"}}))
		constraints = append(constraints, fieldValueConstraint("exclusion_reason", "none", map[string]any{"blocked_by": empty}), fieldValueConstraint("evaluation", "false", map[string]any{"selection": map[string]any{"const": "not_selected"}, "exclusion_reason": map[string]any{"const": "not_matched"}, "blocked_by": empty}))
		for _, reason := range []string{"subsumed", "higher_priority_match"} {
			constraints = append(constraints, fieldValueConstraint("exclusion_reason", reason, map[string]any{"blocked_by": map[string]any{"minItems": 1}, "selection": map[string]any{"const": "not_selected"}}))
		}
	case "outcomeJSON":
		constraints = append(constraints, fieldValueConstraint("source", "condition", map[string]any{"condition": nonempty}), fieldValueConstraint("source", "default", map[string]any{"condition": null}))
	case "SourceSpan":
		for _, field := range []string{"start", "end"} {
			props[field].(map[string]any)["minimum"] = 0
		}
	case "captureJSON":
		props["entry"].(map[string]any)["minimum"] = 1
	case "capturedResponseJSON":
		props["http_status"].(map[string]any)["minimum"] = 0
		props["http_status"].(map[string]any)["maximum"] = 599
		constraints = append(constraints, fieldValueConstraint("parameter_state", "present", map[string]any{"value": text}))
		for _, state := range []string{"absent", "unavailable"} {
			constraints = append(constraints, fieldValueConstraint("parameter_state", state, map[string]any{"value": null}))
		}
		for _, comparison := range []string{"match", "mismatch"} {
			constraints = append(constraints, fieldValueConstraint("comparison", comparison, map[string]any{"parameter_state": map[string]any{"const": "present"}, "value": text, "comparison_reason": map[string]any{"const": "none"}}))
		}
	}
	if len(constraints) > 0 {
		schema["allOf"] = constraints
	}
}
