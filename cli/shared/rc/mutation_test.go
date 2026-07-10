package rc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestPublishProjectConfigMutationSuccess(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{}
	projectCfg := &ProjectConfig{
		Project: core.Project{ProjectID: "project-a"},
		Cache:   &core.ParametersCache{ETag: "etag"},
		Config:  &firebase.RemoteConfig{},
	}

	count, retry, err := PublishProjectConfigMutation(context.Background(), publisher, projectCfg, "update", nil, func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
		next, err := firebase.CloneRemoteConfig(current)
		if err != nil {
			return 0, nil, err
		}
		next.Parameters = map[string]firebase.RemoteConfigParam{
			"flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "on"}},
		}
		return 1, next, nil
	})
	if err != nil {
		t.Fatalf("PublishProjectConfigMutation returned error: %v", err)
	}
	if retry {
		t.Fatalf("retry = true, want false")
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
	if publisher.validateCalls != 1 || publisher.publishCalls != 1 {
		t.Fatalf("validate/publish calls = %d/%d, want 1/1", publisher.validateCalls, publisher.publishCalls)
	}
	if publisher.projectID != "project-a" || publisher.etag != "etag" {
		t.Fatalf("publish target = %s/%s, want project-a/etag", publisher.projectID, publisher.etag)
	}
	var cfg firebase.RemoteConfig
	if err := json.Unmarshal(publisher.raw, &cfg); err != nil {
		t.Fatalf("decode published raw: %v\n%s", err, string(publisher.raw))
	}
	if _, ok := cfg.Parameters["flag"]; !ok {
		t.Fatalf("published raw missing flag: %s", string(publisher.raw))
	}
}

func TestPublishProjectConfigMutationNoChangeSkipsPublish(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{}
	projectCfg := &ProjectConfig{
		Project: core.Project{ProjectID: "project-a"},
		Cache:   &core.ParametersCache{ETag: "etag"},
		Config:  &firebase.RemoteConfig{},
	}

	count, retry, err := PublishProjectConfigMutation(context.Background(), publisher, projectCfg, "delete", nil, func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
		return 0, current, nil
	})
	if err != nil {
		t.Fatalf("PublishProjectConfigMutation returned error: %v", err)
	}
	if count != 0 || retry {
		t.Fatalf("count/retry = %d/%v, want 0/false", count, retry)
	}
	if publisher.validateCalls != 0 || publisher.publishCalls != 0 {
		t.Fatalf("validate/publish calls = %d/%d, want 0/0", publisher.validateCalls, publisher.publishCalls)
	}
}

func TestPublishProjectConfigMutationConflictRetries(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{validateErr: errors.New("returned 412")}
	projectCfg := &ProjectConfig{
		Project: core.Project{ProjectID: "project-a"},
		Cache:   &core.ParametersCache{ETag: "etag"},
		Config:  &firebase.RemoteConfig{},
	}
	var errOut strings.Builder

	count, retry, err := PublishProjectConfigMutation(context.Background(), publisher, projectCfg, "add", &errOut, func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
		cloned, err := firebase.CloneRemoteConfig(current)
		if err != nil {
			return 0, nil, err
		}
		return 2, cloned, nil
	})
	if err != nil {
		t.Fatalf("PublishProjectConfigMutation returned error: %v", err)
	}
	if count != 2 || !retry {
		t.Fatalf("count/retry = %d/%v, want 2/true", count, retry)
	}
	if !strings.Contains(errOut.String(), "remote config changed during add; restarting project project-a") {
		t.Fatalf("retry output = %q, want add conflict message", errOut.String())
	}
}
