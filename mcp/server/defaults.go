package mcpserver

import (
	"fmt"
	"maps"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// inputDefaults only supplies the normalized invocation's empty wrappers, not
// individual flags, selectors, file paths, or document contents.
type inputDefaults map[string]any

var wrapperNames = []string{"arguments", "options", "stdin"}

// addInputDefaults adapts the published MCP schema, keeping the compiled schema
// unchanged for validation after normalization. Defaults also initialize forms.
func addInputDefaults(input map[string]any, compiled *jsonschema.Schema) (inputDefaults, error) {
	defaults := make(inputDefaults)
	properties := input["properties"].(map[string]any)
	for _, name := range wrapperNames {
		var value any
		if name != "stdin" {
			value = map[string]any{}
		}
		if acceptsWrapperDefault(compiled, name, value) {
			defaults[name] = value
			properties[name].(map[string]any)["default"] = value
		}
	}
	if err := optionalWrappers(input, compiled, defaults); err != nil {
		return nil, err
	}
	return defaults, nil
}

// Unconditional constraints can require fields outside the root properties
// declaration (for example, add requires value/type through allOf).
func acceptsWrapperDefault(schema *jsonschema.Schema, name string, value any) bool {
	if property := schema.Properties[name]; property != nil && property.Validate(value) != nil {
		return false
	}
	for _, child := range schema.AllOf {
		if !acceptsWrapperDefault(child, name, value) {
			return false
		}
	}
	return true
}

// Only traverse schemas evaluating the whole invocation, never the contents of
// arguments/options/stdin or schema libraries. A property predicate which rejects
// its default must still require that property: otherwise JSON Schema's
// properties keyword would silently pass on omission, changing if/not/oneOf.
func optionalWrappers(value any, compiled *jsonschema.Schema, defaults inputDefaults) error {
	node, ok := value.(map[string]any)
	if !ok || compiled == nil {
		return nil
	}
	for key := range node {
		switch key {
		case "$schema", "$id", "$defs", "title", "description", "$comment",
			"type", "properties", "additionalProperties", "required",
			"allOf", "anyOf", "oneOf", "if", "then", "else", "not":
		default:
			// Fail closed if future invocation schemas introduce root-level
			// references, property counts, or other presence-sensitive rules.
			if !strings.HasPrefix(key, "x-") {
				return fmt.Errorf("unsupported invocation keyword %q for MCP wrapper defaults", key)
			}
		}
	}
	required := make([]any, 0, len(compiled.Required))
	for _, name := range compiled.Required {
		if _, optional := defaults[name]; !optional {
			required = append(required, name)
		}
	}
	for _, name := range wrapperNames {
		defaultValue, optional := defaults[name]
		if property := compiled.Properties[name]; optional && property != nil && property.Validate(defaultValue) != nil {
			required = append(required, name)
		}
	}
	if _, existed := node["required"]; existed || len(required) != 0 {
		node["required"] = required
	}
	for key, children := range map[string][]*jsonschema.Schema{
		"allOf": compiled.AllOf, "anyOf": compiled.AnyOf, "oneOf": compiled.OneOf,
	} {
		items, _ := node[key].([]any)
		for i, child := range children {
			if err := optionalWrappers(items[i], child, defaults); err != nil {
				return err
			}
		}
	}
	for key, child := range map[string]*jsonschema.Schema{
		"if": compiled.If, "then": compiled.Then, "else": compiled.Else, "not": compiled.Not,
	} {
		if err := optionalWrappers(node[key], child, defaults); err != nil {
			return err
		}
	}
	return nil
}

func (defaults inputDefaults) normalize(value any) {
	object, ok := value.(map[string]any)
	if !ok {
		return // Let the unchanged schema reject non-object inputs.
	}
	for name, defaultValue := range defaults {
		if _, supplied := object[name]; supplied {
			continue // Explicit nulls and invalid values must not be replaced.
		}
		if empty, ok := defaultValue.(map[string]any); ok {
			defaultValue = maps.Clone(empty)
		}
		object[name] = defaultValue
	}
}
