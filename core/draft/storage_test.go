package draft

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestLoadSaveDeleteAndList(t *testing.T) {
	setupDraftTestEnv(t)
	saveParametersCache(t, "demo", "etag-1", remoteConfigRaw("1", map[string]string{"flag": "base"}))

	if raw, hasDraft, err := Load("demo"); err != nil || hasDraft || raw != nil {
		t.Fatalf("Load missing = (%q, %v, %v), want (nil, false, nil)", raw, hasDraft, err)
	}
	if err := Save("demo", json.RawMessage(`{"parameters":`)); err == nil {
		t.Fatal("Save invalid JSON returned nil error")
	}

	raw := remoteConfigRaw("1", map[string]string{"flag": "draft"})
	if err := Save("demo", raw); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	loaded, hasDraft, err := Load("demo")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !hasDraft {
		t.Fatal("hasDraft = false, want true")
	}
	assertParamValue(t, loaded, "flag", "draft")

	ids, err := ListProjectIDs()
	if err != nil {
		t.Fatalf("ListProjectIDs returned error: %v", err)
	}
	if !slices.Contains(ids, "demo") {
		t.Fatalf("ListProjectIDs = %v, want demo", ids)
	}

	if err := Delete("demo"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, hasDraft, err := Load("demo"); err != nil || hasDraft {
		t.Fatalf("Load after delete hasDraft = %v, err = %v; want false, nil", hasDraft, err)
	}
}
