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
	if out.Len() != 0 {
		t.Fatalf("stdout before final rendering = %q, want buffered results", out.String())
	}
	WriteRemoteMutationResults(cmd, totals, "publish", "✅")
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

func TestRunRemotePublishLoopContinuesAfterProjectFailure(t *testing.T) {
	svc := setupLoopTestCore(t)
	projects := []core.Project{
		{ProjectID: "project-a", AuthID: "main"},
		{ProjectID: "project-b", AuthID: "main"},
		{ProjectID: "project-c", AuthID: "main"},
		{ProjectID: "project-d", AuthID: "main"},
	}
	stored := make([]config.Project, 0, len(projects))
	raw := []byte(`{"version":{"versionNumber":"1"},"parameters":{"flag":{"defaultValue":{"value":"old"}}}}`)
	for _, project := range projects {
		stored = append(stored, config.Project{ProjectID: project.ProjectID, AuthID: project.AuthID})
		saveLoopParametersCacheForProject(t, project.ProjectID, raw, "etag-1")
	}
	if err := config.SaveProjects(stored, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}
	published := make(map[string]int)
	validated := make(map[string]int)
	svc.InjectFirebaseService("main", firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		projectID := ""
		for _, project := range projects {
			if strings.Contains(req.URL.Path, project.ProjectID) {
				projectID = project.ProjectID
				break
			}
		}
		switch {
		case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions"):
			return jsonHTTPResponse(http.StatusOK, `{"versions":[{"versionNumber":"1"}]}`, ""), nil
		case req.Method == http.MethodPut && strings.Contains(req.URL.RawQuery, "validateOnly") && projectID == "project-b":
			validated[projectID]++
			return jsonHTTPResponse(http.StatusBadRequest, `{"error":{"message":"invalid candidate"}}`, ""), nil
		case req.Method == http.MethodPut && strings.Contains(req.URL.RawQuery, "validateOnly") && projectID == "project-d":
			validated[projectID]++
			return jsonHTTPResponse(http.StatusPreconditionFailed, `{"error":{"message":"etag mismatch"}}`, ""), nil
		case req.Method == http.MethodPut && strings.Contains(req.URL.RawQuery, "validateOnly"):
			validated[projectID]++
			return jsonHTTPResponse(http.StatusOK, `{}`, ""), nil
		case req.Method == http.MethodPut:
			published[projectID]++
			return jsonHTTPResponse(http.StatusOK, string(raw), `"etag-2"`), nil
		default:
			return nil, errors.New("unexpected: " + req.Method + " " + req.URL.String())
		}
	})}))

	cmd := &cobra.Command{}
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	totals, err := RunRemotePublishLoop(context.Background(), cmd, svc, projects, "update", "✅", func(_ core.Project, cfg *ProjectConfig) (RemoteConfigMutation, error) {
		return func(current *firebase.RemoteConfig) (int, *firebase.RemoteConfig, error) {
			next, cloneErr := firebase.CloneRemoteConfig(current)
			if cloneErr != nil {
				return 0, nil, cloneErr
			}
			next.Parameters["flag"] = firebase.RemoteConfigParam{DefaultValue: &firebase.RemoteConfigValue{Value: "new"}}
			return 1, next, nil
		}, nil
	})
	if err == nil {
		t.Fatal("RunRemotePublishLoop returned nil error")
	}
	if published["project-a"] != 1 || published["project-b"] != 0 || published["project-c"] != 1 || published["project-d"] != 0 {
		t.Fatalf("published = %#v, want a/c only", published)
	}
	if len(totals.Results) != 4 || totals.Results[1].Status != RemoteMutationValidationFailed || totals.Results[3].Status != RemoteMutationConflict {
		t.Fatalf("results = %+v, want ordered validation failure and conflict", totals.Results)
	}
	if validated["project-d"] != 1 {
		t.Fatalf("project-d validation attempts = %d, want one without automatic replanning", validated["project-d"])
	}
	if strings.Contains(errOut.String(), "-p '=project-b'") || strings.Contains(errOut.String(), "-p '=project-d'") {
		t.Fatalf("stderr before final rendering = %q, want buffered recovery hints", errOut.String())
	}
	_, _ = out.WriteString("INFO total\n")
	WriteRemoteMutationResults(cmd, totals, "publish", "✅")
	if !strings.Contains(errOut.String(), "-p '=project-b'") || !strings.Contains(errOut.String(), "-p '=project-d'") {
		t.Fatalf("stderr = %q, want exact retry filters", errOut.String())
	}
	if strings.LastIndex(out.String(), "Results:") < strings.LastIndex(out.String(), "INFO total") {
		t.Fatalf("stdout = %q, want results after final log", out.String())
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
	saveLoopParametersCacheForProject(t, "demo", raw, etag)
}

func saveLoopParametersCacheForProject(t *testing.T, projectID string, raw []byte, etag string) {
	t.Helper()
	cache := &config.ParametersCache{
		ETag:         etag,
		CachedAt:     time.Now().UTC().Add(-15 * time.Minute),
		RemoteConfig: raw,
	}
	if err := config.SaveParametersCache(projectID, cache); err != nil {
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
