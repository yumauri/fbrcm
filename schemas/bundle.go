package schemas

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"
	"strings"
)

// Bundle makes an embedded schema self-contained. References retain their
// semantics, but no consumer needs a resolver for fbrcm's schema URNs.
func Bundle(id string) (map[string]any, error) {
	b := schemaBundler{paths: map[string]string{id: "#"}, documents: make(map[string]any)}
	root, err := b.read(id, "#")
	if err != nil {
		return nil, err
	}
	defs, _ := root["$defs"].(map[string]any)
	if defs == nil {
		defs = make(map[string]any)
		root["$defs"] = defs
	}
	maps.Copy(defs, b.documents)
	return root, nil
}

type schemaBundler struct {
	paths     map[string]string
	documents map[string]any
}

func (b *schemaBundler) read(id, prefix string) (map[string]any, error) {
	raw, err := ReadByID(id)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	if err := b.rewrite(doc, id, prefix); err != nil {
		return nil, err
	}
	return doc, nil
}

func (b *schemaBundler) rewrite(value any, base, prefix string) error {
	switch node := value.(type) {
	case map[string]any:
		delete(node, "$id")
		if ref, ok := node["$ref"].(string); ok {
			id, fragment, _ := strings.Cut(ref, "#")
			path := prefix
			if id != "" && id != base {
				var exists bool
				path, exists = b.paths[id]
				if !exists {
					name := fmt.Sprintf("bundled_%d", len(b.paths))
					path = "#/$defs/" + name
					b.paths[id] = path
					doc, err := b.read(id, path)
					if err != nil {
						return err
					}
					b.documents[name] = doc
				}
			}
			node["$ref"] = path + fragment
		}
		keys := make([]string, 0, len(node))
		for key := range node {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := b.rewrite(node[key], base, prefix); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range node {
			if err := b.rewrite(child, base, prefix); err != nil {
				return err
			}
		}
	}
	return nil
}
