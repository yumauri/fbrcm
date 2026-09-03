package schemas

import "strings"

// MakePortable rewrites a bundled Draft 2020-12 schema in place for consumers
// that require object-valued subschemas and scalar type keywords. It preserves
// validation semantics, including unrestricted JSON values and forbidden inputs.
// Only schema positions are visited: defaults, enums, constants, examples, and
// extension metadata are instance data and must never be rewritten.
func MakePortable(schema map[string]any) {
	portableSchema(schema, "")
}

func portableSchema(value any, keyword string) any {
	if allowed, ok := value.(bool); ok {
		// These keywords accept booleans even in restricted client schemas.
		switch keyword {
		case "additionalProperties", "unevaluatedProperties", "additionalItems", "unevaluatedItems":
			return value
		}
		if !allowed {
			return map[string]any{"not": map[string]any{}}
		}
		if keyword == "not" {
			return map[string]any{}
		}
		return map[string]any{"anyOf": jsonValueTypes()}
	}
	node, ok := value.(map[string]any)
	if !ok {
		return value
	}
	for _, key := range []string{"properties", "patternProperties", "dependentSchemas", "$defs", "definitions", "dependencies"} {
		if children, ok := node[key].(map[string]any); ok {
			for name, child := range children {
				// Draft 7 dependencies may be property-name arrays, not schemas.
				children[name] = portableSchema(child, key)
			}
		}
	}
	for _, key := range []string{"allOf", "anyOf", "oneOf", "prefixItems"} {
		if children, ok := node[key].([]any); ok {
			for i, child := range children {
				children[i] = portableSchema(child, key)
			}
		}
	}
	for _, key := range []string{"items", "contains", "not", "propertyNames", "if", "then", "else", "additionalProperties", "unevaluatedProperties", "additionalItems", "unevaluatedItems", "contentSchema"} {
		if child, exists := node[key]; exists {
			if children, ok := child.([]any); ok && key == "items" {
				for i, item := range children {
					children[i] = portableSchema(item, key)
				}
			} else {
				node[key] = portableSchema(child, key)
			}
		}
	}
	if types, ok := node["type"].([]any); ok {
		branches := make([]any, len(types))
		for i, typ := range types {
			branches[i] = map[string]any{"type": typ}
		}
		delete(node, "type")
		if _, exists := node["anyOf"]; !exists {
			node["anyOf"] = branches
		} else {
			// Type and an existing anyOf both constrain the value. Do not
			// overwrite the existing union or merge it into this one.
			allOf, _ := node["allOf"].([]any)
			node["allOf"] = append(allOf, map[string]any{"anyOf": branches})
		}
	}
	if keyword != "not" && annotationOnly(node) {
		// A definition library or free-form slot accepts every JSON value,
		// not just objects. State the entire JSON value domain explicitly.
		node["anyOf"] = jsonValueTypes()
	}
	return node
}

func annotationOnly(node map[string]any) bool {
	for key := range node {
		switch key {
		case "$schema", "$id", "$anchor", "$dynamicAnchor", "$comment", "$defs", "definitions", "title", "description", "default", "examples", "deprecated", "readOnly", "writeOnly":
			continue
		default:
			if !strings.HasPrefix(key, "x-") {
				return false
			}
		}
	}
	return true
}

func jsonValueTypes() []any {
	// JSON Schema's number type includes integers.
	return []any{
		map[string]any{"type": "null"},
		map[string]any{"type": "boolean"},
		map[string]any{"type": "object"},
		map[string]any{"type": "array"},
		map[string]any{"type": "number"},
		map[string]any{"type": "string"},
	}
}
