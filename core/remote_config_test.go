package core

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestExportRemoteConfig(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")

	const body = `{"version":{"versionNumber":"7"},"parameters":{"flag":{"defaultValue":{"value":"x"}}}}`
	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/remoteConfig") {
				return jsonResponse(http.StatusOK, body, `"etag-export"`), nil
			}
			return nil, io.EOF
		}),
	})
	injectFirebaseService(t, svc, "main", client)

	raw, etag, err := svc.ExportRemoteConfig(context.Background(), "demo")
	if err != nil {
		t.Fatalf("ExportRemoteConfig = %v", err)
	}
	if etag != `"etag-export"` {
		t.Fatalf("etag = %q, want %q", etag, `"etag-export"`)
	}
	assertRemoteConfigVersion(t, raw, "7")
}

func TestDownloadRemoteConfigDefaults(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")

	const body = `{"flag":"on"}`
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Query().Get("format") != "JSON" {
			t.Fatalf("format = %q", req.URL.Query().Get("format"))
		}
		return jsonResponse(http.StatusOK, body, ""), nil
	})})
	injectFirebaseService(t, svc, "main", client)

	defaults, err := svc.DownloadRemoteConfigDefaults(context.Background(), "demo", firebase.DefaultsFormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	if string(defaults) != body {
		t.Fatalf("defaults = %s, want %s", defaults, body)
	}
}

func TestValidateRemoteConfigWithETag(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")

	payload := remoteConfigRaw("1", map[string]string{"flag": "on"})
	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPut && strings.Contains(req.URL.RawQuery, "validateOnly=true") {
				return jsonResponse(http.StatusOK, string(payload), `"etag-1"`), nil
			}
			return nil, io.EOF
		}),
	})
	injectFirebaseService(t, svc, "main", client)

	if err := svc.ValidateRemoteConfigWithETag(context.Background(), "demo", payload, "etag-1"); err != nil {
		t.Fatalf("ValidateRemoteConfigWithETag = %v", err)
	}
}

func TestValidateRemoteConfigWithETagNormalizesUpdatePayload(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")

	payload := []byte(`{"conditions":[{"name":"staff","expression":"true","tagColor":"deep_orange"}],"version":{"versionNumber":"7"}}`)
	var uploaded []byte
	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			var err error
			uploaded, err = io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			return jsonResponse(http.StatusOK, `{}`, `"etag-1"`), nil
		}),
	})
	injectFirebaseService(t, svc, "main", client)

	if err := svc.ValidateRemoteConfigWithETag(context.Background(), "demo", payload, "etag-1"); err != nil {
		t.Fatalf("ValidateRemoteConfigWithETag = %v", err)
	}
	if strings.Contains(string(uploaded), "version") {
		t.Fatalf("uploaded payload retains read-only version metadata: %s", uploaded)
	}
	if !strings.Contains(string(uploaded), `"tagColor":"DEEP_ORANGE"`) {
		t.Fatalf("uploaded payload does not normalize tagColor: %s", uploaded)
	}
}

func TestPublishRemoteConfigWithETagDryRunSkipsCache(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")

	payload := remoteConfigRaw("2", map[string]string{"flag": "published"})
	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPut && !strings.Contains(req.URL.RawQuery, "validateOnly") {
				return jsonResponse(http.StatusOK, string(payload), `"etag-2"`), nil
			}
			return nil, io.EOF
		}),
	})
	injectFirebaseService(t, svc, "main", client)

	ctx := firebase.WithDryRun(context.Background())
	raw, etag, err := svc.PublishRemoteConfigWithETag(ctx, "demo", payload, "etag-1")
	if err != nil {
		t.Fatalf("PublishRemoteConfigWithETag dry-run = %v", err)
	}
	if etag != `"etag-2"` {
		t.Fatalf("etag = %q, want %q", etag, `"etag-2"`)
	}
	if string(raw) != string(payload) {
		t.Fatalf("raw = %s, want %s", raw, payload)
	}

	_, state, err := svc.InspectParametersCache("demo")
	if err != nil || state != ParametersCacheMissing {
		t.Fatalf("InspectParametersCache after dry-run publish = state %v err %v, want missing", state, err)
	}
}

func TestPublishRemoteConfigWithETagWritesVersionedCache(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")

	payload := remoteConfigRaw("2", map[string]string{"flag": "published"})
	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPut && !strings.Contains(req.URL.RawQuery, "validateOnly") {
				return jsonResponse(http.StatusOK, string(payload), `"etag-2"`), nil
			}
			return nil, io.EOF
		}),
	})
	injectFirebaseService(t, svc, "main", client)

	if _, _, err := svc.PublishRemoteConfigWithETag(context.Background(), "demo", payload, "etag-1"); err != nil {
		t.Fatalf("PublishRemoteConfigWithETag = %v", err)
	}
	cache, err := config.LoadParametersCacheVersion("demo", "2")
	if err != nil {
		t.Fatalf("LoadParametersCacheVersion = %v", err)
	}
	if cache.ETag != `"etag-2"` {
		t.Fatalf("etag = %q, want etag-2", cache.ETag)
	}
}

func TestPublishRemoteConfigWithETagRejectsInvalidJSON(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	injectFirebaseService(t, svc, "main", firebase.NewServiceWithHTTPClient(http.DefaultClient))

	_, _, err := svc.PublishRemoteConfigWithETag(context.Background(), "demo", json.RawMessage("{"), "etag-1")
	if err == nil || !strings.Contains(err.Error(), "decode remote config") {
		t.Fatalf("PublishRemoteConfigWithETag invalid = %v, want decode error", err)
	}
}
