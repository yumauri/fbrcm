// Package schemas embeds the published fbrcm CLI JSON Schemas.
package schemas

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"sync"
)

//go:embed cli/1.0.0/*.json
var files embed.FS

type documentID struct {
	ID string `json:"$id"`
}

var (
	indexOnce  sync.Once
	indexPaths map[string]string
	indexIDs   []string
	indexErr   error
)

func loadIndex() {
	entries, err := fs.Glob(files, "cli/1.0.0/*.json")
	if err != nil {
		indexErr = err
		return
	}
	paths := make(map[string]string, len(entries))
	ids := make([]string, 0, len(entries))
	for _, path := range entries {
		raw, readErr := files.ReadFile(path)
		if readErr != nil {
			indexErr = readErr
			return
		}
		var header documentID
		if unmarshalErr := json.Unmarshal(raw, &header); unmarshalErr != nil {
			indexErr = fmt.Errorf("decode schema %s: %w", path, unmarshalErr)
			return
		}
		if previous, exists := paths[header.ID]; exists {
			indexErr = fmt.Errorf("duplicate schema id %q in %s and %s", header.ID, previous, path)
			return
		}
		paths[header.ID] = path
		ids = append(ids, header.ID)
	}
	sort.Strings(ids)
	indexPaths = paths
	indexIDs = ids
}

func List() ([]string, error) {
	indexOnce.Do(loadIndex)
	if indexErr != nil {
		return nil, indexErr
	}
	return append([]string(nil), indexIDs...), nil
}

func ReadByID(id string) (json.RawMessage, error) {
	indexOnce.Do(loadIndex)
	if indexErr != nil {
		return nil, indexErr
	}
	path, exists := indexPaths[id]
	if !exists {
		return nil, fmt.Errorf("schema %q not found", id)
	}
	raw, err := files.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
