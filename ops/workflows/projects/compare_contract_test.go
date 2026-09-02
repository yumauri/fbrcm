package projects

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestCompareJSONUsesChangedContract(t *testing.T) {
	result := rcdiff.Result{Parameters: []rcdiff.ParameterChange{{Key: "flag", Kind: rcdiff.ChangeChanged}}}
	payload := compareJSON(core.Project{ProjectID: "source"}, core.Project{ProjectID: "target"}, result)
	if !payload.Changed {
		t.Fatalf("changed = false, want true")
	}
}

func TestPromoteJSONIncludesChanged(t *testing.T) {
	result := rcdiff.Result{Conditions: []rcdiff.ConditionChange{{Name: "mobile", Kind: rcdiff.ChangeAdded}}}
	payload := promoteJSON(core.Project{ProjectID: "source"}, core.Project{ProjectID: "target"}, compareOptions{DryRun: true}, false, true, core.ValidationSourceFirebase, nil, result)
	if !payload.Changed {
		t.Fatalf("changed = false, want true")
	}
	if !payload.Validated || payload.ValidationSource != core.ValidationSourceFirebase {
		t.Fatalf("validation metadata = %#v/%#v", payload.Validated, payload.ValidationSource)
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
	_, err := loadProjectConfig(context.Background(), svc, "missing", true)
	var selectionErr *shared.SelectionError
	if !errors.As(err, &selectionErr) || selectionErr.Resource != "parameters_cache" || selectionErr.Kind != "not_found" || selectionErr.Query != "missing" || !strings.Contains(err.Error(), "parameters cache not found") {
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
