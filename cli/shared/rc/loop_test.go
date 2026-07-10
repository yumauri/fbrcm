package rc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestRunRemotePublishLoopPublishesMatchedProjects(t *testing.T) {
	svc := setupLoopTestCore(t)
	project := core.Project{ProjectID: "demo", AuthID: "main"}
	raw := []byte(`{"version":{"versionNumber":"1"},"parameters":{"flag":{"defaultValue":{"value":"old"}}}}`)

	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: loopTransport(t, "1", string(raw)),
	})
	svc.InjectFirebaseService("main", client)

	saveLoopParametersCache(t, raw, "etag-1")

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	totals, err := RunRemotePublishLoop(context.Background(), cmd, svc, []core.Project{project}, "update", "✅", func(_ core.Project, cfg *ProjectConfig) (RemoteConfigMutation, error) {
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			next, err := firebase.CloneRemoteConfig(current)
			if err != nil {
				return 0, nil, err
			}
			next.Parameters["flag"] = firebase.RemoteConfigParam{
				DefaultValue: &firebase.RemoteConfigValue{Value: "new"},
			}
			return 1, next, nil
		}, nil
	})
	if err != nil {
		t.Fatalf("RunRemotePublishLoop = %v", err)
	}
	if totals.ModifiedProjects != 1 || totals.ChangedParams != 1 {
		t.Fatalf("totals = %+v, want 1 project / 1 param", totals)
	}
	if !strings.Contains(out.String(), "published: demo") {
		t.Fatalf("stdout = %q, want publish line", out.String())
	}
}

func TestRunRemotePublishLoopSkipsNilMutation(t *testing.T) {
	svc := setupLoopTestCore(t)
	project := core.Project{ProjectID: "demo", AuthID: "main"}
	raw := []byte(`{"version":{"versionNumber":"1"},"parameters":{}}`)
	saveLoopParametersCache(t, raw, "etag-1")
	svc.InjectFirebaseService("main", firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: loopTransport(t, "1", string(raw)),
	}))

	cmd := &cobra.Command{}
	totals, err := RunRemotePublishLoop(context.Background(), cmd, svc, []core.Project{project}, "update", "✅", func(_ core.Project, _ *ProjectConfig) (RemoteConfigMutation, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatalf("RunRemotePublishLoop = %v", err)
	}
	if totals.ModifiedProjects != 0 || totals.ChangedParams != 0 {
		t.Fatalf("totals = %+v, want zero", totals)
	}
}

func setupLoopTestCore(t *testing.T) *core.Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile = %v", err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatalf("NewService = %v", err)
	}
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	if err := config.SaveProjects([]config.Project{{ProjectID: "demo", AuthID: "main"}}, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}
	return svc
}

func saveLoopParametersCache(t *testing.T, raw []byte, etag string) {
	t.Helper()
	cache := &config.ParametersCache{
		ETag:         etag,
		CachedAt:     time.Now().UTC().Add(-15 * time.Minute),
		RemoteConfig: raw,
	}
	if err := config.SaveParametersCache("demo", cache); err != nil {
		t.Fatalf("SaveParametersCache = %v", err)
	}
}

func loopTransport(t *testing.T, version, publishBody string) http.RoundTripper {
	t.Helper()
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions"):
			body := `{"versions":[{"versionNumber":"` + version + `"}]}`
			return jsonHTTPResponse(http.StatusOK, body, ""), nil
		case req.Method == http.MethodPut && strings.Contains(req.URL.Path, "/remoteConfig"):
			return jsonHTTPResponse(http.StatusOK, publishBody, `"etag-2"`), nil
		default:
			return nil, errors.New("unexpected: " + req.Method + " " + req.URL.String())
		}
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonHTTPResponse(status int, body, etag string) *http.Response {
	resp := &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	if etag != "" {
		resp.Header.Set("ETag", etag)
	}
	return resp
}
