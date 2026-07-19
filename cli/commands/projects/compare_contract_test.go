package projects

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

func TestCompareJSONUsesChangedContract(t *testing.T) {
	result := rcdiff.Result{Parameters: []rcdiff.ParameterChange{{Key: "flag", Kind: rcdiff.ChangeChanged}}}
	payload := compareJSON(core.Project{ProjectID: "source"}, core.Project{ProjectID: "target"}, result).(map[string]any)
	if changed, ok := payload["changed"].(bool); !ok || !changed {
		t.Fatalf("changed = %#v, want true", payload["changed"])
	}
}

func TestPromoteJSONIncludesChanged(t *testing.T) {
	result := rcdiff.Result{Conditions: []rcdiff.ConditionChange{{Name: "mobile", Kind: rcdiff.ChangeAdded}}}
	payload := promoteJSON(core.Project{ProjectID: "source"}, core.Project{ProjectID: "target"}, compareOptions{DryRun: true}, false, nil, result).(map[string]any)
	if changed, ok := payload["changed"].(bool); !ok || !changed {
		t.Fatalf("changed = %#v, want true", payload["changed"])
	}
}

func TestLoadProjectConfigCachedRequiresLocalCache(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}

	svc := &core.Core{}
	if _, err := loadProjectConfig(context.Background(), svc, "missing", true); err == nil || !strings.Contains(err.Error(), "parameters cache not found") {
		t.Fatalf("missing cached config error = %v", err)
	}

	raw := json.RawMessage(`{"parameters":{"flag":{"defaultValue":{"value":"on"}}},"version":{"versionNumber":"7"}}`)
	if err := config.SaveParametersCache("demo", &config.ParametersCache{ETag: "etag-7", CachedAt: time.Now().Add(-time.Hour), RemoteConfig: raw}); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadProjectConfig(context.Background(), svc, "demo", true)
	if err != nil {
		t.Fatalf("load stale cached config = %v", err)
	}
	if cfg.Version.VersionNumber != "7" || cfg.Parameters["flag"].DefaultValue.Value != "on" {
		t.Fatalf("cached config = %#v", cfg)
	}
}
