package draft

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
)

func setupDraftTestEnv(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
}

func saveParametersCache(t *testing.T, projectID, etag string, raw json.RawMessage) *config.ParametersCache {
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

type fakeDeps struct {
	cache        *config.ParametersCache
	publishErr   error
	publishCalls int
}

func (f *fakeDeps) deps() Deps {
	return Deps{
		GetParameters: func(ctx context.Context, projectID string, force bool) (*config.ParametersCache, string, error) {
			return f.cache, "cache", nil
		},
		InspectParametersCache: func(projectID string) (*config.ParametersCache, error) {
			return f.cache, nil
		},
		PublishRemoteConfigWithETag: func(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error) {
			f.publishCalls++
			if f.publishErr != nil {
				return nil, "", f.publishErr
			}
			return raw, "etag-2", nil
		},
	}
}

func remoteConfigRaw(version string, params map[string]string) json.RawMessage {
	cfg := firebase.RemoteConfig{
		Parameters: make(map[string]firebase.RemoteConfigParam, len(params)),
		Version:    firebase.RemoteConfigVersion{VersionNumber: version},
	}
	for key, value := range params {
		cfg.Parameters[key] = remoteConfigParam(value, "STRING")
	}
	return marshalRemoteConfig(cfg)
}

func typedRemoteConfigRaw(version, key, value, valueType string) json.RawMessage {
	cfg := firebase.RemoteConfig{
		Parameters: map[string]firebase.RemoteConfigParam{
			key: remoteConfigParam(value, valueType),
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: version},
	}
	return marshalRemoteConfig(cfg)
}

func groupedRemoteConfigRaw(version, group string, params map[string]string) json.RawMessage {
	groupParams := make(map[string]firebase.RemoteConfigParam, len(params))
	for key, value := range params {
		groupParams[key] = remoteConfigParam(value, "STRING")
	}
	cfg := firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			group: {Parameters: groupParams},
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: version},
	}
	return marshalRemoteConfig(cfg)
}

func conditionalRemoteConfigRaw(version, key, condition, conditionalValue string) json.RawMessage {
	param := remoteConfigParam("default", "STRING")
	param.ConditionalValues = map[string]firebase.RemoteConfigValue{
		condition: {Value: conditionalValue},
	}
	cfg := firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: condition, Expression: "true"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			key: param,
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: version},
	}
	return marshalRemoteConfig(cfg)
}

func conditionOnlyRemoteConfigRaw(version, key, condition string) json.RawMessage {
	cfg := firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: condition, Expression: "true"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			key: remoteConfigParam("default", "STRING"),
		},
		Version: firebase.RemoteConfigVersion{VersionNumber: version},
	}
	return marshalRemoteConfig(cfg)
}

func remoteConfigParam(value, valueType string) firebase.RemoteConfigParam {
	v := firebase.RemoteConfigValue{Value: value}
	return firebase.RemoteConfigParam{
		DefaultValue: &v,
		ValueType:    valueType,
	}
}

func marshalRemoteConfig(cfg firebase.RemoteConfig) json.RawMessage {
	raw, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return raw
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

func assertGroupParamValue(t *testing.T, raw json.RawMessage, groupName, key, want string) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	group, ok := cfg.ParameterGroups[groupName]
	if !ok {
		t.Fatalf("group %q missing", groupName)
	}
	param, ok := group.Parameters[key]
	if !ok {
		t.Fatalf("parameter %q missing in group %q", key, groupName)
	}
	if param.DefaultValue == nil {
		t.Fatalf("parameter %q default value missing", key)
	}
	if got := param.DefaultValue.Value; got != want {
		t.Fatalf("parameter %q value = %q, want %q", key, got, want)
	}
}

func assertGroupMissing(t *testing.T, raw json.RawMessage, groupName string) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	if _, ok := cfg.ParameterGroups[groupName]; ok {
		t.Fatalf("group %q exists, want missing", groupName)
	}
}

func assertConditionalMissing(t *testing.T, raw json.RawMessage, key, condition string) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	param, ok := cfg.Parameters[key]
	if !ok {
		t.Fatalf("parameter %q missing", key)
	}
	if _, ok := param.ConditionalValues[condition]; ok {
		t.Fatalf("conditional value %q exists, want missing", condition)
	}
}

func assertConditionalValue(t *testing.T, raw json.RawMessage, key, condition, want string) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	param, ok := cfg.Parameters[key]
	if !ok {
		t.Fatalf("parameter %q missing", key)
	}
	value, ok := param.ConditionalValues[condition]
	if !ok {
		t.Fatalf("conditional value %q missing", condition)
	}
	if value.Value != want {
		t.Fatalf("conditional value %q = %q, want %q", condition, value.Value, want)
	}
}

func assertParamDescription(t *testing.T, raw json.RawMessage, key, want string) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	param, ok := cfg.Parameters[key]
	if !ok {
		t.Fatalf("parameter %q missing", key)
	}
	if got := param.Description; got != want {
		t.Fatalf("parameter %q description = %q, want %q", key, got, want)
	}
}

func assertFlagMissing(t *testing.T, raw json.RawMessage) {
	t.Helper()
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig returned error: %v", err)
	}
	if _, ok := cfg.Parameters["flag"]; ok {
		t.Fatalf("parameter %q exists, want missing", "flag")
	}
}

func loadDraft(t *testing.T, projectID string) (json.RawMessage, bool) {
	t.Helper()
	raw, hasDraft, err := Load(projectID)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	return raw, hasDraft
}
