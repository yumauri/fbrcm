package schemas

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEmbeddedSchemasHaveUniqueVersionedIDs(t *testing.T) {
	ids, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Fatal("no schemas embedded")
	}
	seen := make(map[string]bool, len(ids))
	currentPrefix := "urn:fbrcm:schema:cli:1.0.0:"
	foundCurrent := false
	for _, id := range ids {
		if seen[id] || !strings.HasPrefix(id, "urn:fbrcm:schema:cli:") {
			t.Fatalf("invalid or duplicate schema id %q", id)
		}
		foundCurrent = foundCurrent || strings.HasPrefix(id, currentPrefix)
		seen[id] = true
		raw, readErr := ReadByID(id)
		if readErr != nil {
			t.Fatal(readErr)
		}
		var document struct {
			Draft string `json:"$schema"`
			ID    string `json:"$id"`
		}
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatalf("decode %s: %v", id, err)
		}
		if document.ID != id || document.Draft != "https://json-schema.org/draft/2020-12/schema" {
			t.Fatalf("schema header for %s = %#v", id, document)
		}
	}
	if !foundCurrent {
		t.Fatal("no schemas embedded for current contract")
	}
}
