package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	charmlog "charm.land/log/v2"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type appRoundTripFunc func(*http.Request) (*http.Response, error)

func (f appRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestListFirebaseAppsPaginatesAndSorts(t *testing.T) {
	requests := 0
	svc := newAppsTestCore(t, appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if requests == 1 {
			return appJSONResponse(req, `{"apps":[{"name":"projects/demo/webApps/w","displayName":"Zeta","platform":"WEB","appId":"web"}],"nextPageToken":"next"}`), nil
		}
		if req.URL.Query().Get("pageToken") != "next" {
			t.Fatalf("page token = %q", req.URL.Query().Get("pageToken"))
		}
		return appJSONResponse(req, `{"apps":[{"name":"projects/demo/androidApps/a","displayName":"Alpha","platform":"ANDROID","appId":"android","namespace":"com.example"}]}`), nil
	}))
	apps, err := svc.ListFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 2 || apps[0].AppID != "android" || apps[1].AppID != "web" {
		t.Fatalf("apps = %#v", apps)
	}
}

func TestGetFirebaseAppResolvesPrecedenceAndAmbiguity(t *testing.T) {
	svc := newAppsTestCore(t, appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/v1beta1/projects/demo:searchApps" {
			return appJSONResponse(req, `{"apps":[{"name":"projects/demo/androidApps/a","displayName":"Shared","platform":"ANDROID","appId":"android","namespace":"com.example"},{"name":"projects/demo/webApps/w","displayName":"Shared","platform":"WEB","appId":"web","namespace":"example"}]}`), nil
		}
		return appJSONResponse(req, `{"name":"projects/demo/androidApps/a","displayName":"Shared","appId":"android","projectId":"demo","packageName":"com.example","state":"ACTIVE"}`), nil
	}))
	details, err := svc.GetFirebaseApp(context.Background(), "demo", "android")
	if err != nil || details.ProjectID != "demo" || details.PackageName == nil || *details.PackageName != "com.example" {
		t.Fatalf("details = %#v, %v", details, err)
	}
	_, err = svc.GetFirebaseApp(context.Background(), "demo", "Shared")
	var lookup *AppLookupError
	if err == nil || !errors.As(err, &lookup) || lookup.Kind != "ambiguous" || len(lookup.Candidates) != 2 {
		t.Fatalf("ambiguous error = %#v", err)
	}
}

func TestReadFirebaseAppsCachesInventoryForOneHour(t *testing.T) {
	requests := 0
	svc := newAppsTestCore(t, appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		return appJSONResponse(req, `{"apps":[{"name":"projects/demo/webApps/w","displayName":"Demo","platform":"WEB","appId":"1:123:web:abc","state":"ACTIVE"}]}`), nil
	}))
	first, err := svc.ReadFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.ReadFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if requests != 1 || first.Source != AppCacheSourceFirebase || second.Source != AppCacheSourceCache {
		t.Fatalf("requests = %d, sources = %q, %q", requests, first.Source, second.Source)
	}
	if config.AppsCacheTTL != time.Hour {
		t.Fatalf("AppsCacheTTL = %s", config.AppsCacheTTL)
	}
}

func TestReadFirebaseAppsLogsFreshCacheLifecycle(t *testing.T) {
	svc := newAppsTestCore(t, appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(req.URL.Path, ":searchApps"):
			return appJSONResponse(req, `{"apps":[{"name":"projects/demo/webApps/w","displayName":"Demo","platform":"WEB","appId":"1:123:web:abc","state":"ACTIVE"}]}`), nil
		case strings.HasSuffix(req.URL.Path, ":getConfig"):
			return appJSONResponse(req, `{"projectId":"demo","appId":"1:123:web:abc","apiKey":"key"}`), nil
		default:
			return appJSONResponse(req, `{"name":"projects/demo/webApps/w","displayName":"Demo","appId":"1:123:web:abc","projectId":"demo","webId":"w","state":"ACTIVE"}`), nil
		}
	}))

	if _, err := svc.ReadFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReadFirebaseApp(context.Background(), "demo", "1:123:web:abc", ListFirebaseAppsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReadFirebaseAppConfig(context.Background(), "demo", "1:123:web:abc", ListFirebaseAppsOptions{}); err != nil {
		t.Fatal(err)
	}

	t.Setenv(env.LogNoTimestamp, "1")
	t.Setenv(env.NoColor, "1")
	var logs bytes.Buffer
	previousLevel := corelog.CurrentLevel()
	corelog.ConfigureCLIOutput(&logs, &logs)
	corelog.InitWithDefault(corelog.ModeCLI, charmlog.InfoLevel)
	t.Cleanup(func() {
		corelog.ConfigureCLIOutput(os.Stderr, os.Stderr)
		corelog.SetLevel(previousLevel)
	})

	if _, err := svc.ReadFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReadFirebaseApp(context.Background(), "demo", "1:123:web:abc", ListFirebaseAppsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReadFirebaseAppConfig(context.Background(), "demo", "1:123:web:abc", ListFirebaseAppsOptions{}); err != nil {
		t.Fatal(err)
	}

	for _, message := range []string{
		"loaded applications cache",
		"applications cache fresh",
		"serving applications from fresh cache",
		"loaded application details cache",
		"application details cache fresh",
		"serving application details from fresh cache",
		"loaded application configuration cache",
		"application configuration cache fresh",
		"serving application configuration from fresh cache",
	} {
		if !strings.Contains(logs.String(), message) {
			t.Errorf("logs do not contain %q:\n%s", message, logs.String())
		}
	}
	if !strings.Contains(logs.String(), "demo") || !strings.Contains(logs.String(), "1:123:web:abc") {
		t.Errorf("logs do not contain cache identity fields:\n%s", logs.String())
	}
}

func TestReadFirebaseAppsUpdateAndCachedOnly(t *testing.T) {
	requests := 0
	svc := newAppsTestCore(t, appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		return appJSONResponse(req, `{"apps":[{"name":"projects/demo/androidApps/a","platform":"ANDROID","appId":"1:123:android:abc","state":"ACTIVE"}]}`), nil
	}))
	if _, err := svc.ReadFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReadFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{Update: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReadFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{CachedOnly: true}); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestReadFirebaseAppsFallsBackToStaleIndex(t *testing.T) {
	payload := []byte(`[{"resource_name":"projects/demo/webApps/w","display_name":"Demo","platform":"web","app_id":"1:123:web:abc","namespace":"","api_key_id":"","state":"ACTIVE"}]`)
	svc := newAppsTestCore(t, appRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("offline")
	}))
	if err := config.SaveAppsIndexCache("demo", time.Now().Add(-2*time.Hour), payload); err != nil {
		t.Fatal(err)
	}
	result, err := svc.ReadFirebaseApps(context.Background(), "demo", ListFirebaseAppsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != AppCacheSourceCacheStale || result.RefreshError == nil || len(result.Apps) != 1 {
		t.Fatalf("result = %#v", result)
	}
}

func TestReadFirebaseAppConfigDoesNotImplicitlyFallbackToStaleArtifact(t *testing.T) {
	failConfig := false
	svc := newAppsTestCore(t, appRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, ":searchApps") {
			return appJSONResponse(req, `{"apps":[{"name":"projects/demo/androidApps/a","platform":"ANDROID","appId":"1:123:android:abc","state":"ACTIVE"}]}`), nil
		}
		if failConfig {
			return nil, errors.New("offline")
		}
		return appJSONResponse(req, `{"configFilename":"google-services.json","configFileContents":"e30K"}`), nil
	}))
	first, err := svc.ReadFirebaseAppConfig(context.Background(), "demo", "1:123:android:abc", ListFirebaseAppsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(first.Config.App)
	if err != nil {
		t.Fatal(err)
	}
	if err := config.SaveAppConfigCache("demo", first.Config.App.AppID, time.Now().Add(-2*time.Hour), first.Config.SuggestedFilename, first.Config.MediaType, payload, first.Config.Contents); err != nil {
		t.Fatal(err)
	}
	failConfig = true
	if _, err := svc.ReadFirebaseAppConfig(context.Background(), "demo", "1:123:android:abc", ListFirebaseAppsOptions{}); err == nil {
		t.Fatal("expected stale app configuration refresh failure")
	}
	cached, err := svc.ReadFirebaseAppConfig(context.Background(), "demo", "1:123:android:abc", ListFirebaseAppsOptions{CachedOnly: true})
	if err != nil || cached.Source != AppCacheSourceCacheStale || string(cached.Config.Contents) != "{}\n" {
		t.Fatalf("cached = %#v, err = %v", cached, err)
	}
}

func newAppsTestCore(t *testing.T, transport http.RoundTripper) *Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	svc, err := NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	svc.InjectFirebaseService("main", firebase.NewServiceWithHTTPClient(&http.Client{Transport: transport}))
	return svc
}

func appJSONResponse(req *http.Request, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}
