package rc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

type fakeRemoteConfigPublisher struct {
	validateErr   error
	publishErr    error
	publishRaw    json.RawMessage
	validated     bool
	published     bool
	validateCalls int
	publishCalls  int
	projectID     string
	raw           json.RawMessage
	etag          string
}

func (f *fakeRemoteConfigPublisher) ValidateRemoteConfigWithETag(_ context.Context, projectID string, raw json.RawMessage, etag string) error {
	f.validated = true
	f.validateCalls++
	f.projectID = projectID
	f.raw = raw
	f.etag = etag
	return f.validateErr
}

func (f *fakeRemoteConfigPublisher) PublishRemoteConfigWithETag(_ context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error) {
	f.published = true
	f.publishCalls++
	f.projectID = projectID
	f.raw = raw
	f.etag = etag
	return f.publishRaw, "", f.publishErr
}

func TestValidateAndPublishRemoteConfigSuccess(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{}

	retry, err := ValidateAndPublishRemoteConfig(context.Background(), publisher, "demo", []byte(`{}`), "etag", "update", nil)
	if err != nil {
		t.Fatalf("ValidateAndPublishRemoteConfig returned error: %v", err)
	}
	if retry {
		t.Fatalf("retry = true, want false")
	}
	if !publisher.validated || !publisher.published {
		t.Fatalf("validated=%v published=%v, want both true", publisher.validated, publisher.published)
	}
}

func TestPublishProjectConfigMutationResultCapturesPublishedVersion(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{
		publishRaw: json.RawMessage(`{"version":{"versionNumber":"42"}}`),
	}
	projectCfg := &ProjectConfig{
		Project: core.Project{ProjectID: "demo"},
		Cache:   &core.ParametersCache{ETag: "etag-41"},
		Config:  &firebase.RemoteConfig{},
	}
	result, err := PublishProjectConfigMutationResult(context.Background(), publisher, projectCfg, "update", nil, func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
		return 1, current, nil
	})
	if err != nil {
		t.Fatalf("PublishProjectConfigMutationResult = %v", err)
	}
	if result.ChangedCount != 1 || result.Retry || result.FailureStage != "" || result.PublishedVersion != "42" || !result.Validated || result.ValidationSource != core.ValidationSourceFirebase {
		t.Fatalf("result = %+v, want changed count 1 and published version 42", result)
	}
}

func TestPublishProjectConfigMutationResultIdentifiesValidationConflictStage(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{validateErr: &firebase.APIError{StatusCode: 412}}
	projectCfg := &ProjectConfig{
		Project: core.Project{ProjectID: "demo"},
		Cache:   &core.ParametersCache{ETag: "etag-41"},
		Config:  &firebase.RemoteConfig{},
	}
	result, err := PublishProjectConfigMutationResult(context.Background(), publisher, projectCfg, "update", nil, func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
		return 1, current, nil
	})
	if err != nil {
		t.Fatalf("PublishProjectConfigMutationResult = %v", err)
	}
	if !result.Retry || result.FailureStage != "validation" || result.Validated || result.ValidationSource != core.ValidationSourceFirebase {
		t.Fatalf("result = %+v, want retryable validation conflict", result)
	}
}

func TestValidateAndPublishRemoteConfigValidateConflict(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{validateErr: &firebase.APIError{StatusCode: 412}}
	var errOut bytes.Buffer

	retry, err := ValidateAndPublishRemoteConfig(context.Background(), publisher, "demo", []byte(`{}`), "etag", "delete", &errOut)
	if err != nil {
		t.Fatalf("ValidateAndPublishRemoteConfig returned error: %v", err)
	}
	if !retry {
		t.Fatalf("retry = false, want true")
	}
	if publisher.published {
		t.Fatalf("published after validate conflict")
	}
	if got := errOut.String(); !strings.Contains(got, "remote config changed during delete; restarting project demo") {
		t.Fatalf("retry output = %q", got)
	}
}

func TestValidateAndPublishRemoteConfigPublishConflict(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{publishErr: &firebase.APIError{StatusCode: 412}}

	retry, err := ValidateAndPublishRemoteConfig(context.Background(), publisher, "demo", []byte(`{}`), "etag", "add", nil)
	if err != nil {
		t.Fatalf("ValidateAndPublishRemoteConfig returned error: %v", err)
	}
	if !retry {
		t.Fatalf("retry = false, want true")
	}
}

func TestValidateAndPublishRemoteConfigNonConflict(t *testing.T) {
	publisher := &fakeRemoteConfigPublisher{validateErr: errors.New("permission denied")}

	retry, err := ValidateAndPublishRemoteConfig(context.Background(), publisher, "demo", []byte(`{}`), "etag", "add", nil)
	if err == nil {
		t.Fatalf("ValidateAndPublishRemoteConfig returned nil error")
	}
	if retry {
		t.Fatalf("retry = true, want false")
	}
}
