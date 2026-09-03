package mcpserver

import (
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/schemas"
)

func toolInputSchema(capability contract.Capability, options Options) (map[string]any, *tool, error) {
	input, err := schemas.Bundle(capability.InvocationSchema)
	if err != nil {
		return nil, nil, err
	}
	specializeStateless(input, "", options.Stateless)
	properties := input["properties"].(map[string]any)
	optionProperties := properties["options"].(map[string]any)["properties"].(map[string]any)
	for name := range optionProperties {
		if boundOption(name) || (!options.AllowWrites && (name == "to" || name == "plan-out")) {
			delete(optionProperties, name)
		}
	}
	schemas.MakePortable(input)
	compiler := jsonschema.NewCompiler()
	id := "https://fbrcm.invalid/mcp/" + capability.ID
	if err := compiler.AddResource(id, input); err != nil {
		return nil, nil, err
	}
	compiled, err := compiler.Compile(id)
	if err != nil {
		return nil, nil, err
	}
	defaults, err := addInputDefaults(input, compiled)
	if err != nil {
		return nil, nil, err
	}
	return input, &tool{capability: capability, schema: compiled, defaults: defaults}, nil
}

// specializeStateless partially evaluates the existing invocation schema with
// the server's fixed execution mode. In particular, conditional selector and
// draft/cache constraints must still apply after removing the stateless flag
// from model-controlled options. No selector grammar is reimplemented here.
func specializeStateless(value any, path string, stateless bool) any {
	switch node := value.(type) {
	case map[string]any:
		if path == "options" {
			if properties, ok := node["properties"].(map[string]any); ok {
				if flag, ok := properties["stateless"].(map[string]any); ok {
					if expected, ok := flag["const"].(bool); ok && expected != stateless {
						return false
					}
					delete(properties, "stateless")
				}
			}
			if required, ok := node["required"].([]any); ok {
				filtered := make([]any, 0, len(required))
				for _, name := range required {
					if name != "stateless" {
						filtered = append(filtered, name)
					}
				}
				node["required"] = filtered
			}
		}
		for key, child := range node {
			switch key {
			case "properties":
				if properties, ok := child.(map[string]any); ok {
					for name, property := range properties {
						properties[name] = specializeStateless(property, name, stateless)
					}
				}
			case "if", "then", "else", "not", "allOf", "anyOf", "oneOf":
				node[key] = specializeStateless(child, path, stateless)
			}
		}
	case []any:
		for i, child := range node {
			node[i] = specializeStateless(child, path, stateless)
		}
	}
	return value
}
