package core

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestGetRemoteConfigPreviousCachedVersion(t *testing.T) {
	svc := setupCoreTestEnv(t)
	for _, version := range []string{"7", "9", "10"} {
		cache := &config.ParametersCache{ETag: "etag-" + version, CachedAt: time.Now().UTC(), RemoteConfig: []byte(`{"version":{"versionNumber":"` + version + `"}}`)}
		if version == "10" {
			if err := config.SaveParametersCache("demo", cache); err != nil {
				t.Fatal(err)
			}
		} else if err := config.SaveParametersCacheSnapshot("demo", cache); err != nil {
			t.Fatal(err)
		}
	}
	resolved, err := svc.GetRemoteConfigVersion(context.Background(), "demo", "previous", true)
	if err != nil {
		t.Fatalf("GetRemoteConfigVersion previous cached = %v", err)
	}
	if resolved.Version.VersionNumber != "9" {
		t.Fatalf("previous cached version = %q, want 9", resolved.Version.VersionNumber)
	}
}

func TestGetRemoteConfigRelativeCachedVersion(t *testing.T) {
	svc := setupCoreTestEnv(t)
	for _, version := range []string{"7", "9", "10"} {
		cache := &config.ParametersCache{ETag: "etag-" + version, CachedAt: time.Now().UTC(), RemoteConfig: []byte(`{"version":{"versionNumber":"` + version + `"}}`)}
		if version == "10" {
			if err := config.SaveParametersCache("demo", cache); err != nil {
				t.Fatal(err)
			}
		} else if err := config.SaveParametersCacheSnapshot("demo", cache); err != nil {
			t.Fatal(err)
		}
	}
	resolved, err := svc.GetRemoteConfigVersion(context.Background(), "demo", "current~2", true)
	if err != nil {
		t.Fatalf("GetRemoteConfigVersion current~2 cached = %v", err)
	}
	if resolved.Version.VersionNumber != "7" {
		t.Fatalf("relative cached version = %q, want 7", resolved.Version.VersionNumber)
	}
}

func TestGetRemoteConfigPreviousFirebaseVersion(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions") {
			if req.URL.Query().Get("pageSize") != "2" {
				t.Fatalf("pageSize = %q, want 2", req.URL.Query().Get("pageSize"))
			}
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"10"},{"versionNumber":"9"}]}`, ""), nil
		}
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/remoteConfig") {
			return jsonResponse(http.StatusOK, `{"version":{"versionNumber":"9"}}`, `"etag-9"`), nil
		}
		return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
	})})
	injectFirebaseService(t, svc, "main", client)
	resolved, err := svc.GetRemoteConfigVersion(context.Background(), "demo", "previous", false)
	if err != nil {
		t.Fatalf("GetRemoteConfigVersion previous = %v", err)
	}
	if resolved.Version.VersionNumber != "9" {
		t.Fatalf("previous Firebase version = %q, want 9", resolved.Version.VersionNumber)
	}
}

func TestGetRemoteConfigRelativeFirebaseVersion(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions") {
			if req.URL.Query().Get("pageSize") != "3" {
				t.Fatalf("pageSize = %q, want 3", req.URL.Query().Get("pageSize"))
			}
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"10"},{"versionNumber":"8"},{"versionNumber":"5"}]}`, ""), nil
		}
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/remoteConfig") {
			return jsonResponse(http.StatusOK, `{"version":{"versionNumber":"5"}}`, `"etag-5"`), nil
		}
		return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
	})})
	injectFirebaseService(t, svc, "main", client)
	resolved, err := svc.GetRemoteConfigVersion(context.Background(), "demo", "latest~2", false)
	if err != nil {
		t.Fatalf("GetRemoteConfigVersion latest~2 = %v", err)
	}
	if resolved.Version.VersionNumber != "5" {
		t.Fatalf("relative Firebase version = %q, want 5", resolved.Version.VersionNumber)
	}
}

func TestGetRemoteConfigVersionHonorsStatelessExecutionPolicy(t *testing.T) {
	svc := setupCoreTestEnv(t)
	local := &config.ParametersCache{ETag: "etag-local", CachedAt: time.Now().UTC(), RemoteConfig: []byte(`{"version":{"versionNumber":"9"},"parameters":{"flag":{"defaultValue":{"value":"local"}}}}`)}
	if err := config.SaveParametersCacheSnapshot("demo", local); err != nil {
		t.Fatal(err)
	}
	snapshotPath := config.GetParametersCacheVersionPath("demo", "9")
	snapshotBefore, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}

	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/remoteConfig") {
			return jsonResponse(http.StatusOK, `{"version":{"versionNumber":"9"},"parameters":{"flag":{"defaultValue":{"value":"remote"}}}}`, `"etag-remote"`), nil
		}
		return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
	})})
	ctx := WithExecutionPolicy(context.Background(), StatelessExecutionPolicy())
	ctx, err = WithDirectFirebaseService(ctx, "demo", client)
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := svc.GetRemoteConfigVersion(ctx, "demo", "9", false)
	if err != nil {
		t.Fatalf("GetRemoteConfigVersion = %v", err)
	}
	if resolved.Cached || resolved.Cache == nil || resolved.Cache.ETag != `"etag-remote"` {
		t.Fatalf("resolved version = %#v, want uncached remote result", resolved)
	}
	snapshotAfter, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(snapshotAfter) != string(snapshotBefore) {
		t.Fatal("version snapshot changed under stateless policy")
	}

	_, err = svc.GetRemoteConfigVersion(ctx, "demo", "9", true)
	var policyErr *ExecutionPolicyError
	if !errors.As(err, &policyErr) || policyErr.Capability != "local-state reads" {
		t.Fatalf("cached-only error = %v, want local-read ExecutionPolicyError", err)
	}
}

func TestGetRemoteConfigVersionPairSharesFirebaseHistoryRequest(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	current := &config.ParametersCache{ETag: "etag-10", CachedAt: time.Now().UTC(), RemoteConfig: []byte(`{"version":{"versionNumber":"10"}}`)}
	if err := config.SaveParametersCache("demo", current); err != nil {
		t.Fatal(err)
	}

	listRequests := 0
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions") {
			listRequests++
			if req.URL.Query().Get("pageSize") != "2" {
				t.Fatalf("pageSize = %q, want 2", req.URL.Query().Get("pageSize"))
			}
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"10"},{"versionNumber":"9"}]}`, ""), nil
		}
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/remoteConfig") {
			if got := req.URL.Query().Get("versionNumber"); got != "9" {
				t.Fatalf("versionNumber = %q, want 9", got)
			}
			return jsonResponse(http.StatusOK, `{"version":{"versionNumber":"9"}}`, `"etag-9"`), nil
		}
		return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
	})})
	injectFirebaseService(t, svc, "main", client)

	from, to, err := svc.GetRemoteConfigVersionPair(context.Background(), "demo", "previous", "current", false)
	if err != nil {
		t.Fatalf("GetRemoteConfigVersionPair = %v", err)
	}
	if from.Version.VersionNumber != "9" || to.Version.VersionNumber != "10" {
		t.Fatalf("pair = %s -> %s, want 9 -> 10", from.Version.VersionNumber, to.Version.VersionNumber)
	}
	if listRequests != 1 {
		t.Fatalf("list requests = %d, want 1", listRequests)
	}
}

func TestGetRemoteConfigVersionPairClassifiesMissingRelativeVersion(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions") {
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"10"}]}`, ""), nil
		}
		return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
	})})
	injectFirebaseService(t, svc, "main", client)

	_, _, err := svc.GetRemoteConfigVersionPair(context.Background(), "demo", "current~2", "previous", false)
	var lookup *RemoteConfigVersionLookupError
	if !errors.As(err, &lookup) || lookup.Kind != "not_found" || lookup.ProjectID != "demo" || lookup.Selector != "current~2" {
		t.Fatalf("error = %#v, lookup = %#v", err, lookup)
	}
}

func TestListRemoteConfigVersionsDoesNotMarkFilteredFirstVersionCurrent(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	if err := config.SaveParametersCache("demo", &config.ParametersCache{
		ETag: "etag-10", CachedAt: time.Now().UTC(), RemoteConfig: []byte(`{"version":{"versionNumber":"10"}}`),
	}); err != nil {
		t.Fatal(err)
	}
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet || !strings.Contains(req.URL.Path, "listVersions") {
			return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
		}
		return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"9"},{"versionNumber":"8"}]}`, ""), nil
	})})
	injectFirebaseService(t, svc, "main", client)

	result, err := svc.ListRemoteConfigVersions(context.Background(), "demo", VersionListOptions{Limit: 20, Before: "9"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Versions) != 2 {
		t.Fatalf("versions = %#v", result.Versions)
	}
	for _, version := range result.Versions {
		if version.Current {
			t.Fatalf("filtered version %s incorrectly marked current", version.VersionNumber)
		}
	}
}

func TestRemoteConfigVersionLookupErrorsAreTyped(t *testing.T) {
	svc := setupCoreTestEnv(t)
	for _, test := range []struct {
		selector string
		kind     string
	}{
		{selector: "not-a-version", kind: "invalid_argument"},
		{selector: "7", kind: "not_found"},
		{selector: "previous", kind: "not_found"},
		{selector: " current", kind: "invalid_argument"},
		{selector: "CURRENT", kind: "invalid_argument"},
	} {
		_, err := svc.GetRemoteConfigVersion(context.Background(), "demo", test.selector, true)
		var lookup *RemoteConfigVersionLookupError
		if !errors.As(err, &lookup) || lookup.Kind != test.kind || lookup.ProjectID != "demo" || lookup.Selector != test.selector {
			t.Errorf("selector %q error = %#v, want typed %s lookup error", test.selector, err, test.kind)
		}
	}
}

func TestParseRelativeVersionSelectorRejectsInvalidDistance(t *testing.T) {
	for _, selector := range []string{"current~0", "latest~-1", "current~x", "current~300", "42~1"} {
		if _, _, _, err := parseRelativeVersionSelector(selector); err == nil {
			t.Fatalf("parseRelativeVersionSelector(%q) accepted invalid selector", selector)
		}
	}
}
