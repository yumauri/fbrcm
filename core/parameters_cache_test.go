package core

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestInspectParametersCacheStates(t *testing.T) {
	svc := setupCoreTestEnv(t)

	cache, state, err := svc.InspectParametersCache("missing")
	if err != nil {
		t.Fatalf("InspectParametersCache missing = %v", err)
	}
	if cache != nil || state != ParametersCacheMissing {
		t.Fatalf("missing = cache=%v state=%v, want nil/missing", cache, state)
	}

	saveDefaultParametersCache(t, map[string]string{"flag": "on"})
	cache, state, err = svc.InspectParametersCache("demo")
	if err != nil || state != ParametersCacheFresh {
		t.Fatalf("fresh = cache=%v state=%v err=%v", cache, state, err)
	}

	saveStaleParametersCache(t, "demo", "1")
	cache, state, err = svc.InspectParametersCache("demo")
	if err != nil || state != ParametersCacheStale {
		t.Fatalf("stale = cache=%v state=%v err=%v", cache, state, err)
	}

	writeCorruptParametersCache(t, "broken")
	_, state, err = svc.InspectParametersCache("broken")
	if err == nil || !strings.Contains(err.Error(), "decode cached remote config") {
		t.Fatalf("corrupt = err=%v, want decode error", err)
	}
	if state != ParametersCacheMissing {
		t.Fatalf("corrupt state = %v, want missing", state)
	}
}

func TestGetParametersServesFreshCache(t *testing.T) {
	svc := setupCoreTestEnv(t)
	saveDefaultParametersCache(t, map[string]string{"flag": "cached"})

	cache, source, err := svc.GetParameters(context.Background(), "demo", false)
	if err != nil {
		t.Fatalf("GetParameters returned error: %v", err)
	}
	if source != "cache" {
		t.Fatalf("source = %q, want cache", source)
	}
	if cache.ETag != "etag-1" {
		t.Fatalf("etag = %q, want etag-1", cache.ETag)
	}
}

func TestGetParametersForceFetchFromFirebase(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	saveDefaultParametersCache(t, map[string]string{"flag": "old"})

	const fetchedBody = `{"version":{"versionNumber":"99"},"parameters":{"flag":{"defaultValue":{"value":"new"}}}}`
	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/remoteConfig") && !strings.Contains(req.URL.Path, "listVersions") {
				return jsonResponse(http.StatusOK, fetchedBody, `"etag-new"`), nil
			}
			return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
		}),
	})
	injectFirebaseService(t, svc, "main", client)

	cache, source, err := svc.GetParameters(context.Background(), "demo", true)
	if err != nil {
		t.Fatalf("GetParameters force = %v", err)
	}
	if source != "firebase" {
		t.Fatalf("source = %q, want firebase", source)
	}
	assertRemoteConfigVersion(t, cache.RemoteConfig, "99")

	loaded, err := config.LoadParametersCache("demo")
	if err != nil {
		t.Fatalf("LoadParametersCache = %v", err)
	}
	if loaded.ETag != `"etag-new"` {
		t.Fatalf("persisted etag = %q, want %q", loaded.ETag, `"etag-new"`)
	}
}

func TestRevalidateParametersRefreshesVerifiedCache(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	saveStaleParametersCache(t, "demo", "5")

	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions") {
				body := `{"versions":[{"versionNumber":"5"}]}`
				return jsonResponse(http.StatusOK, body, ""), nil
			}
			return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
		}),
	})
	injectFirebaseService(t, svc, "main", client)

	before, _ := config.LoadParametersCache("demo")
	cache, source, err := svc.RevalidateParameters(context.Background(), "demo")
	if err != nil {
		t.Fatalf("RevalidateParameters = %v", err)
	}
	if source != "cache-verified" {
		t.Fatalf("source = %q, want cache-verified", source)
	}
	if !cache.CachedAt.After(before.CachedAt) {
		t.Fatalf("CachedAt not refreshed: before=%v after=%v", before.CachedAt, cache.CachedAt)
	}
}

func TestBuildParametersTreeAndNilCache(t *testing.T) {
	svc := setupCoreTestEnv(t)
	cache := saveDefaultParametersCache(t, map[string]string{"flag": "on"})

	tree, err := svc.BuildParametersTree(cache)
	if err != nil {
		t.Fatalf("BuildParametersTree = %v", err)
	}
	if got := treeValue(tree, "flag"); got != "on" {
		t.Fatalf("flag value = %q, want on", got)
	}

	if _, err := svc.BuildParametersTree(nil); err == nil {
		t.Fatal("BuildParametersTree(nil) = nil, want error")
	}
}

func TestParametersStatusLabel(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		source  string
		cached  time.Time
		hasTree bool
		err     error
		want    string
	}{
		{name: "firebase fetch", source: "firebase", cached: now, want: "fetch"},
		{name: "recent cache", source: "cache", cached: now.Add(-30 * time.Second), want: "fetch"},
		{name: "older cache", source: "cache", cached: now.Add(-5 * time.Minute), want: "cached"},
		{name: "stale cache", source: "cache", cached: now.Add(-11 * time.Minute), want: "staled"},
		{name: "error with tree", source: "cache", cached: now, hasTree: true, err: errors.New("boom"), want: "error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParametersStatusLabel(tc.source, tc.cached, tc.hasTree, tc.err); got != tc.want {
				t.Fatalf("ParametersStatusLabel() = %q, want %q", got, tc.want)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonResponse(status int, body, etag string) *http.Response {
	resp := &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	if etag != "" {
		resp.Header.Set("ETag", etag)
	}
	return resp
}
