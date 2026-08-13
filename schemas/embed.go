// Package schemas embeds the published fbrcm CLI JSON Schemas.
package schemas

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed cli/1.0.0/*.json
var files embed.FS

type documentID struct {
	ID string `json:"$id"`
}

func List() ([]string, error) {
	entries, err := fs.Glob(files, "cli/1.0.0/*.json")
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(entries))
	for _, path := range entries {
		raw, readErr := files.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		var header documentID
		if unmarshalErr := json.Unmarshal(raw, &header); unmarshalErr != nil {
			return nil, fmt.Errorf("decode schema %s: %w", path, unmarshalErr)
		}
		ids = append(ids, header.ID)
	}
	sort.Strings(ids)
	return ids, nil
}

func ReadByID(id string) (json.RawMessage, error) {
	entries, err := fs.Glob(files, "cli/1.0.0/*.json")
	if err != nil {
		return nil, err
	}
	for _, path := range entries {
		raw, readErr := files.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		var header documentID
		if json.Unmarshal(raw, &header) == nil && header.ID == id {
			return raw, nil
		}
	}
	return nil, fmt.Errorf("schema %q not found", id)
}
