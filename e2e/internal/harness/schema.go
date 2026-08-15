package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ValidateCommandOutput validates a CLI JSON envelope against its published schema.
func ValidateCommandOutput(schemasRoot, commandID string, raw []byte) error {
	if strings.TrimSpace(schemasRoot) == "" {
		return fmt.Errorf("schemas root is required for JSON scenario %s", commandID)
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	entries, err := os.ReadDir(schemasRoot)
	if err != nil {
		return fmt.Errorf("read CLI schemas: %w", err)
	}
	wantedPath := filepath.Join(schemasRoot, strings.ReplaceAll(commandID, ".", "_")+".response.schema.json")
	var wantedID string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(schemasRoot, entry.Name())
		document, id, readErr := readSchema(path)
		if readErr != nil {
			return readErr
		}
		if id == "" {
			continue
		}
		if err := compiler.AddResource(id, document); err != nil {
			return fmt.Errorf("register schema %s: %w", path, err)
		}
		if path == wantedPath {
			wantedID = id
		}
	}
	if wantedID == "" {
		return fmt.Errorf("response schema for command %s was not found at %s", commandID, wantedPath)
	}
	compiled, err := compiler.Compile(wantedID)
	if err != nil {
		return fmt.Errorf("compile response schema for %s: %w", commandID, err)
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("decode JSON stdout for %s: %w", commandID, err)
	}
	if err := compiled.Validate(value); err != nil {
		return fmt.Errorf("stdout does not conform to the published %s response schema: %w", commandID, err)
	}
	return nil
}

func readSchema(path string) (any, string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read schema %s: %w", path, err)
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, "", fmt.Errorf("decode schema %s: %w", path, err)
	}
	root, _ := document.(map[string]any)
	id, _ := root["$id"].(string)
	return document, id, nil
}
