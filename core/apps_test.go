package core

import (
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
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
