package core

import (
	"context"
	"errors"
	"net/http"
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

func TestParseRelativeVersionSelectorRejectsInvalidDistance(t *testing.T) {
	for _, selector := range []string{"current~0", "latest~-1", "current~x", "current~300", "42~1"} {
		if _, _, _, err := parseRelativeVersionSelector(selector); err == nil {
			t.Fatalf("parseRelativeVersionSelector(%q) accepted invalid selector", selector)
		}
	}
}
