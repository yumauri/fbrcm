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

func TestBindProjectsAuthMatchesAndPersists(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	projects := []config.Project{
		{Name: "Alpha App", ProjectID: "alpha", AuthID: "old", DiscoveredBy: []string{"old"}},
		{Name: "Beta App", ProjectID: "beta", AuthID: "old"},
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	matched, err := svc.BindProjectsAuth([]string{"=alpha"}, "main")
	if err != nil {
		t.Fatalf("BindProjectsAuth = %v", err)
	}
	if len(matched) != 1 || matched[0].ProjectID != "alpha" {
		t.Fatalf("matched = %+v, want alpha only", matched)
	}
	if matched[0].AuthID != "main" {
		t.Fatalf("AuthID = %q, want main", matched[0].AuthID)
	}
	if got := strings.Join(matched[0].DiscoveredBy, ","); got != "old" {
		t.Fatalf("DiscoveredBy = %q, want observed identities unchanged", got)
	}

	project, err := svc.ProjectByID("alpha")
	if err != nil {
		t.Fatalf("ProjectByID alpha = %v", err)
	}
	if project.AuthID != "main" {
		t.Fatalf("persisted AuthID = %q, want main", project.AuthID)
	}
	if got := strings.Join(project.DiscoveredBy, ","); got != "old" {
		t.Fatalf("persisted DiscoveredBy = %q, want observed identities unchanged", got)
	}
}

func TestBindProjectsAuthRequiresConfiguredAuth(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveAuth(&config.AuthFile{Version: config.AuthConfigVersion}); err != nil {
		t.Fatalf("SaveAuth empty = %v", err)
	}
	if err := config.SaveProjects([]config.Project{{ProjectID: "demo", AuthID: "main"}}, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	_, err := svc.BindProjectsAuth(nil, "missing")
	if err == nil || !strings.Contains(err.Error(), `auth "missing" is not configured`) {
		t.Fatalf("BindProjectsAuth = %v, want missing auth error", err)
	}
}

func TestBindProjectIDsAuthUsesExactProjectIDs(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	projects := []config.Project{
		{Name: "Other", ProjectID: "alpha", AuthID: "old"},
		{Name: "alpha", ProjectID: "beta", AuthID: "old"},
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	matched, err := svc.BindProjectIDsAuth([]string{"alpha"}, "main")
	if err != nil {
		t.Fatalf("BindProjectIDsAuth = %v", err)
	}
	if len(matched) != 1 || matched[0].ProjectID != "alpha" {
		t.Fatalf("matched = %+v, want exact project ID alpha", matched)
	}
	beta, err := svc.ProjectByID("beta")
	if err != nil {
		t.Fatalf("ProjectByID beta = %v", err)
	}
	if beta.AuthID != "old" {
		t.Fatalf("project named alpha was rebound to %q, want old", beta.AuthID)
	}
}

func TestBindProjectsAuthNoMatches(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	if err := config.SaveProjects([]config.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	_, err := svc.BindProjectsAuth([]string{"=missing"}, "main")
	if err == nil || !strings.Contains(err.Error(), "no projects matched") {
		t.Fatalf("BindProjectsAuth = %v, want no projects matched", err)
	}
}

func TestProjectByIDMissing(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects([]config.Project{{ProjectID: "demo", AuthID: "main"}}, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	_, err := svc.ProjectByID("other")
	if err == nil || !strings.Contains(err.Error(), `project "other" is not in projects config`) {
		t.Fatalf("ProjectByID = %v, want not found error", err)
	}
}

func TestSyncProjectsForAuthUsesFirebaseStub(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "existing")

	client := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodGet && req.URL.Host == "cloudresourcemanager.googleapis.com" && req.URL.Path == "/v1/projects" {
				body := `{"projects":[{"projectId":"demo-sync","name":"Demo Sync","lifecycleState":"ACTIVE"}]}`
				return jsonResponse(http.StatusOK, body, ""), nil
			}
			if req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/v3/projects/") {
				body := `{"projectId":"demo-sync","displayName":"Demo Sync","state":"ACTIVE","etag":"e1","updateTime":"2026-01-01T00:00:00Z"}`
				return jsonResponse(http.StatusOK, body, ""), nil
			}
			return nil, errors.New("unexpected request: " + req.URL.String())
		}),
	})
	injectFirebaseService(t, svc, "main", client)

	projects, source, err := svc.SyncProjectsForAuth(context.Background(), "main")
	if err != nil {
		t.Fatalf("SyncProjectsForAuth = %v", err)
	}
	if source != "firebase" {
		t.Fatalf("source = %q, want firebase", source)
	}
	found := false
	for _, project := range projects {
		if project.ProjectID == "demo-sync" {
			found = true
			if project.AuthID != "main" {
				t.Fatalf("AuthID = %q, want main", project.AuthID)
			}
		}
	}
	if !found {
		t.Fatalf("projects = %+v, want demo-sync", projects)
	}
}

func TestListProjectsUsesCacheWhenPresent(t *testing.T) {
	svc := setupCoreTestEnv(t)
	projects := []config.Project{{Name: "Cached", ProjectID: "cached", AuthID: "main"}}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	got, source, err := svc.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects = %v", err)
	}
	if source != "cache" {
		t.Fatalf("source = %q, want cache", source)
	}
	if len(got) != 1 || got[0].ProjectID != "cached" {
		t.Fatalf("projects = %+v, want cached", got)
	}
}

func TestPurgeProjectsClearsCache(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects([]config.Project{{ProjectID: "demo"}}, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	if err := svc.PurgeProjects(); err != nil {
		t.Fatalf("PurgeProjects = %v", err)
	}
	_, err := config.LoadProjects()
	if err == nil {
		t.Fatal("LoadProjects after purge = nil, want error")
	}
}
