package mcpserver

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
