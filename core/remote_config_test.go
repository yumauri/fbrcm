package core

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corehooks "github.com/yumauri/fbrcm/core/hooks"
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

func TestExportRemoteConfigWithDirectFirebaseService(t *testing.T) {
	svc := setupCoreTestEnv(t)

	const body = `{"version":{"versionNumber":"8"},"parameters":{"flag":{"defaultValue":{"value":"direct"}}}}`
	var requestPath string
	requestCount := 0
	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requestCount++
			requestPath = req.URL.EscapedPath()
			return jsonResponse(http.StatusOK, body, `"etag-direct"`), nil
		}),
	})
	ctx, err := WithDirectFirebaseService(context.Background(), "demo", client)
	if err != nil {
		t.Fatal(err)
	}

	raw, etag, err := svc.ExportRemoteConfig(ctx, "server@demo")
	if err != nil {
		t.Fatalf("ExportRemoteConfig = %v", err)
	}
	if etag != `"etag-direct"` {
		t.Fatalf("etag = %q, want %q", etag, `"etag-direct"`)
	}
	if !strings.Contains(requestPath, "/projects/demo/") || !strings.Contains(requestPath, "/namespaces/firebase-server/remoteConfig") {
		t.Fatalf("request path = %q, want server template for physical project demo", requestPath)
	}
	assertRemoteConfigVersion(t, raw, "8")

	if _, _, err := svc.ExportRemoteConfig(context.Background(), "demo"); err == nil || !strings.Contains(err.Error(), "read projects config") {
		t.Fatalf("ExportRemoteConfig without direct context = %v, want stored-project lookup error", err)
	}
	if requestCount != 1 {
		t.Fatalf("direct firebase service handled %d requests after its context was discarded, want 1", requestCount)
	}
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

func TestValidateRemoteConfigWithETagReportsFirebaseValidationSource(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusBadRequest, `{"error":"invalid template"}`, ""), nil
	})})
	injectFirebaseService(t, svc, "main", client)

	err := svc.ValidateRemoteConfigWithETag(firebase.WithDryRun(context.Background()), "demo", remoteConfigRaw("2", nil), "etag-1")
	if source, ok := RemoteConfigValidationSource(err); !ok || source != ValidationSourceFirebase {
		t.Fatalf("validation source = %q/%t for error %v", source, ok, err)
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

func TestPublishRemoteConfigWithETagHonorsStatelessExecutionPolicy(t *testing.T) {
	svc := setupCoreTestEnv(t)
	saveDefaultParametersCache(t, map[string]string{"flag": "cached"})
	cachePath := config.GetParametersCachePath("demo")
	cacheBefore, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	markerDir := t.TempDir()
	preMarker := filepath.Join(markerDir, "pre")
	postMarker := filepath.Join(markerDir, "post")
	if err := config.SaveAppConfig(&config.AppConfig{Hooks: &config.HooksConfig{
		PrePublish:  []string{"printf pre > " + shellQuote(preMarker)},
		PostPublish: []string{"printf post > " + shellQuote(postMarker)},
	}}); err != nil {
		t.Fatal(err)
	}

	payload := remoteConfigRaw("2", map[string]string{"flag": "published"})
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodPut && !strings.Contains(req.URL.RawQuery, "validateOnly") {
			return jsonResponse(http.StatusOK, string(payload), `"etag-2"`), nil
		}
		return nil, io.EOF
	})})
	ctx := WithExecutionPolicy(context.Background(), StatelessExecutionPolicy())
	ctx, err = WithDirectFirebaseService(ctx, "demo", client)
	if err != nil {
		t.Fatal(err)
	}

	raw, etag, err := svc.PublishRemoteConfigWithETag(ctx, "demo", payload, "etag-1")
	if err != nil {
		t.Fatalf("PublishRemoteConfigWithETag = %v", err)
	}
	if string(raw) != string(payload) || etag != `"etag-2"` {
		t.Fatalf("published response = %s/%q, want payload/etag-2", raw, etag)
	}
	for _, marker := range []string{preMarker, postMarker} {
		if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("hook marker %s stat = %v, want not found", marker, err)
		}
	}
	cacheAfter, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(cacheAfter) != string(cacheBefore) {
		t.Fatalf("cache file changed under stateless policy")
	}
}

func TestPublishRemoteConfigWithETagReportsRemoteSuccessWhenCacheWriteFails(t *testing.T) {
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

	blocked := config.GetParametersCacheDirPath()
	if err := os.MkdirAll(filepath.Dir(blocked), 0o755); err != nil {
		t.Fatalf("MkdirAll = %v", err)
	}
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("WriteFile = %v", err)
	}

	raw, etag, err := svc.PublishRemoteConfigWithETag(context.Background(), "demo", payload, "etag-1")
	var cacheErr *RemoteConfigPublishedCacheError
	if !errors.As(err, &cacheErr) {
		t.Fatalf("error = %v, want RemoteConfigPublishedCacheError", err)
	}
	if string(raw) != string(payload) || etag != `"etag-2"` {
		t.Fatalf("published response = %s/%q, want payload/etag-2", raw, etag)
	}
	if string(cacheErr.RemoteConfig) != string(payload) || cacheErr.ETag != `"etag-2"` {
		t.Fatalf("typed outcome = %s/%q, want payload/etag-2", cacheErr.RemoteConfig, cacheErr.ETag)
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

func TestPublishRemoteConfigWithETagDryRunRunsOnlyPrePublishHooks(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	markerDir := t.TempDir()
	preMarker := filepath.Join(markerDir, "pre")
	postMarker := filepath.Join(markerDir, "post")
	if err := config.SaveAppConfig(&config.AppConfig{Hooks: &config.HooksConfig{
		PrePublish:  []string{"printf pre > " + shellQuote(preMarker)},
		PostPublish: []string{"printf post > " + shellQuote(postMarker)},
	}}); err != nil {
		t.Fatal(err)
	}
	payload := remoteConfigRaw("2", map[string]string{"flag": "published"})
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodPut {
			return jsonResponse(http.StatusOK, string(payload), `"etag-2"`), nil
		}
		return nil, io.EOF
	})})
	injectFirebaseService(t, svc, "main", client)

	if _, _, err := svc.PublishRemoteConfigWithETag(firebase.WithDryRun(context.Background()), "demo", payload, "etag-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(preMarker); err != nil {
		t.Fatalf("pre hook did not run: %v", err)
	}
	if _, err := os.Stat(postMarker); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("post hook ran during dry-run: %v", err)
	}
}

func TestPublishRemoteConfigWithETagPreHookFailurePreventsWrite(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	if err := config.SaveAppConfig(&config.AppConfig{Hooks: &config.HooksConfig{PrePublish: []string{"exit 17"}}}); err != nil {
		t.Fatal(err)
	}
	writes := 0
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		writes++
		return jsonResponse(http.StatusOK, `{}`, `"etag-2"`), nil
	})})
	injectFirebaseService(t, svc, "main", client)

	_, _, err := svc.PublishRemoteConfigWithETag(context.Background(), "demo", remoteConfigRaw("2", nil), "etag-1")
	var hookErr *corehooks.Error
	if !errors.As(err, &hookErr) || hookErr.Event != corehooks.PrePublish || hookErr.ExitCode != 17 {
		t.Fatalf("publish error = %v", err)
	}
	if writes != 0 {
		t.Fatalf("Firebase writes = %d, want zero", writes)
	}
}

func TestPublishRemoteConfigWithETagPostHookFailureReportsPublishedState(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	if err := config.SaveAppConfig(&config.AppConfig{Hooks: &config.HooksConfig{PostPublish: []string{"exit 23"}}}); err != nil {
		t.Fatal(err)
	}
	payload := remoteConfigRaw("2", map[string]string{"flag": "published"})
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodPut {
			return jsonResponse(http.StatusOK, string(payload), `"etag-2"`), nil
		}
		return nil, io.EOF
	})})
	injectFirebaseService(t, svc, "main", client)

	raw, etag, err := svc.PublishRemoteConfigWithETag(context.Background(), "demo", payload, "etag-1")
	var publishedErr *RemoteConfigPublishedHookError
	var hookErr *corehooks.Error
	if !errors.As(err, &publishedErr) || !errors.As(err, &hookErr) || hookErr.Event != corehooks.PostPublish {
		t.Fatalf("publish error = %v", err)
	}
	if len(raw) == 0 || etag != `"etag-2"` || string(publishedErr.RemoteConfig) != string(payload) {
		t.Fatalf("published state = %s/%s, typed=%+v", raw, etag, publishedErr)
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
