package shared

import (
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
)

func TestNewProjectConfig(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	cache := &core.ParametersCache{
		ETag:         "etag",
		CachedAt:     time.Unix(1, 0).UTC(),
		RemoteConfig: []byte(`{"parameters":{"flag":{"defaultValue":{"value":"on"}}}}`),
	}

	got, err := NewProjectConfig(project, cache)
	if err != nil {
		t.Fatalf("NewProjectConfig returned error: %v", err)
	}
	if got.Project.ProjectID != "demo" {
		t.Fatalf("ProjectID = %q, want demo", got.Project.ProjectID)
	}
	if got.Cache != cache {
		t.Fatalf("Cache was not preserved")
	}
	if got.Config.Parameters["flag"].DefaultValue.Value != "on" {
		t.Fatalf("flag value = %q, want on", got.Config.Parameters["flag"].DefaultValue.Value)
	}

	delete(got.Config.Parameters, "flag")
	if _, ok := got.Config.Parameters["flag"]; ok {
		t.Fatalf("test setup unexpectedly preserved flag")
	}
	if !strings.Contains(string(cache.RemoteConfig), `"flag"`) {
		t.Fatalf("cache raw config was mutated")
	}
}

func TestNewProjectConfigDecodeErrorIncludesProjectID(t *testing.T) {
	_, err := NewProjectConfig(core.Project{ProjectID: "demo"}, &core.ParametersCache{RemoteConfig: []byte(`{`)})
	if err == nil {
		t.Fatalf("NewProjectConfig accepted invalid config")
	}
	if !strings.Contains(err.Error(), "demo") {
		t.Fatalf("error = %q, want project id", err.Error())
	}
}
