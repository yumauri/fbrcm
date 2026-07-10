package core

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
)

// Facade integration smoke tests. Draft pipeline behavior is covered in core/draft/*_test.go.

func TestDraftFacadeEditBuildsTree(t *testing.T) {
	svc := newDraftTestService(t)
	saveDefaultParametersCache(t, map[string]string{"flag": "old"})

	cache, tree, hasDraft, err := svc.SetStringParameterValue(context.Background(), "demo", "", "flag", "default", "new", false)
	if err != nil {
		t.Fatalf("SetStringParameterValue returned error: %v", err)
	}
	if !hasDraft {
		t.Fatal("hasDraft = false, want true")
	}
	if cache.ETag != "etag-1" {
		t.Fatalf("cache ETag = %q, want etag-1", cache.ETag)
	}
	if got := treeValue(tree, "flag"); got != "new" {
		t.Fatalf("tree flag value = %q, want new", got)
	}

	raw, hasDraft, err := svc.LoadDraft("demo")
	if err != nil {
		t.Fatalf("LoadDraft returned error: %v", err)
	}
	if !hasDraft {
		t.Fatal("hasDraft after edit = false, want true")
	}
	assertParamValue(t, raw, "flag", "new")
}

func TestDraftFacadeBuildDraftAwareTree(t *testing.T) {
	svc := newDraftTestService(t)
	cache := saveDefaultParametersCache(t, map[string]string{"flag": "cached"})
	if err := svc.SaveDraft("demo", remoteConfigRaw("2", map[string]string{"flag": "draft"})); err != nil {
		t.Fatalf("SaveDraft returned error: %v", err)
	}

	tree, hasDraft, err := svc.BuildDraftAwareParametersTree("demo", cache)
	if err != nil {
		t.Fatalf("BuildDraftAwareParametersTree returned error: %v", err)
	}
	if !hasDraft {
		t.Fatal("hasDraft = false, want true")
	}
	if got := treeValue(tree, "flag"); got != "draft" {
		t.Fatalf("tree flag value = %q, want draft", got)
	}
	if tree.ETag != "etag-1" {
		t.Fatalf("tree ETag = %q, want etag-1", tree.ETag)
	}
}

func TestDraftFacadeDiscardRestoresCache(t *testing.T) {
	svc := newDraftTestService(t)
	saveDefaultParametersCache(t, map[string]string{"flag": "cached"})
	if err := svc.SaveDraft("demo", remoteConfigRaw("2", map[string]string{"flag": "draft"})); err != nil {
		t.Fatalf("SaveDraft returned error: %v", err)
	}

	cache, tree, err := svc.DiscardDraft(context.Background(), "demo")
	if err != nil {
		t.Fatalf("DiscardDraft returned error: %v", err)
	}
	if cache.ETag != "etag-1" {
		t.Fatalf("cache ETag = %q, want etag-1", cache.ETag)
	}
	if got := treeValue(tree, "flag"); got != "cached" {
		t.Fatalf("tree flag value = %q, want cached", got)
	}
	if _, hasDraft, err := svc.LoadDraft("demo"); err != nil || hasDraft {
		t.Fatalf("LoadDraft after discard hasDraft = %v, err = %v; want false, nil", hasDraft, err)
	}
}

func newDraftTestService(t *testing.T) *Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
	svc, err := NewService(context.Background())
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	return svc
}

func saveDefaultParametersCache(t *testing.T, params map[string]string) *ParametersCache {
	t.Helper()
	return saveParametersCacheRaw(t, "demo", "etag-1", remoteConfigRaw("1", params))
}

func saveParametersCacheRaw(t *testing.T, projectID, etag string, raw json.RawMessage) *ParametersCache {
	t.Helper()
	cache := &config.ParametersCache{
		ETag:         etag,
		CachedAt:     time.Now().UTC(),
		RemoteConfig: raw,
	}
	if err := config.SaveParametersCache(projectID, cache); err != nil {
		t.Fatalf("SaveParametersCache returned error: %v", err)
	}
	return cache
}

func remoteConfigRaw(version string, params map[string]string) json.RawMessage {
	cfg := firebase.RemoteConfig{
		Parameters: make(map[string]firebase.RemoteConfigParam, len(params)),
		Version:    firebase.RemoteConfigVersion{VersionNumber: version},
	}
	for key, value := range params {
		cfg.Parameters[key] = remoteConfigParam(value, "STRING")
	}
	return marshalRemoteConfigForTest(cfg)
}

func remoteConfigParam(value, valueType string) firebase.RemoteConfigParam {
	v := firebase.RemoteConfigValue{Value: value}
	return firebase.RemoteConfigParam{
		DefaultValue: &v,
		ValueType:    valueType,
	}
}

func marshalRemoteConfigForTest(cfg firebase.RemoteConfig) json.RawMessage {
	raw, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return raw
}

func treeValue(tree *ParametersTree, key string) string {
	for _, group := range tree.Groups {
		for _, param := range group.Parameters {
			if param.Key == key && len(param.Values) > 0 {
				return param.Values[len(param.Values)-1].RawValue
			}
		}
	}
	return ""
}

func assertParamValue(t *testing.T, raw json.RawMessage, key, want string) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	param, ok := cfg.Parameters[key]
	if !ok {
		t.Fatalf("parameter %q missing", key)
	}
	if param.DefaultValue == nil {
		t.Fatalf("parameter %q default value missing", key)
	}
	if got := param.DefaultValue.Value; got != want {
		t.Fatalf("parameter %q value = %q, want %q", key, got, want)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
