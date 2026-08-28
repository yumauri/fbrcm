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
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func TestBindProjectsAuthMatchesAndPersists(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	projects := []config.Project{
		{Name: "Alpha App", ProjectID: "alpha", AuthID: "old", DiscoveredBy: []string{"old", "main"}},
		{Name: "Beta App", ProjectID: "beta", AuthID: "old"},
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	result, err := svc.BindProjectsAuth([]string{"=alpha"}, "main")
	if err != nil {
		t.Fatalf("BindProjectsAuth = %v", err)
	}
	if len(result.Bound) != 1 || result.Bound[0].ProjectID != "alpha" || len(result.Skipped) != 0 {
		t.Fatalf("result = %+v, want alpha only", result)
	}
	if result.Bound[0].AuthID != "main" {
		t.Fatalf("AuthID = %q, want main", result.Bound[0].AuthID)
	}
	if got := strings.Join(result.Bound[0].DiscoveredBy, ","); got != "main,old" {
		t.Fatalf("DiscoveredBy = %q, want observed identities unchanged", got)
	}

	project, err := svc.ProjectByID("alpha")
	if err != nil {
		t.Fatalf("ProjectByID alpha = %v", err)
	}
	if project.AuthID != "main" {
		t.Fatalf("persisted AuthID = %q, want main", project.AuthID)
	}
	if got := strings.Join(project.DiscoveredBy, ","); got != "main,old" {
		t.Fatalf("persisted DiscoveredBy = %q, want observed identities unchanged", got)
	}
}

func TestSetProjectTemplatePreferencesPersistsViewsAndPrimary(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects([]config.Project{{
		Name:      "Demo",
		ProjectID: "demo",
		AuthID:    "main",
	}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	updated, err := svc.SetProjectTemplatePreferences(
		"server@demo",
		[]rctarget.Kind{rctarget.Client, rctarget.Server},
		rctarget.Server,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := updated.TemplateKinds(); len(got) != 2 || got[0] != rctarget.Server || got[1] != rctarget.Client {
		t.Fatalf("updated template kinds = %v", got)
	}
	persisted, err := svc.ProjectByID("demo")
	if err != nil {
		t.Fatal(err)
	}
	if persisted.PrimaryTemplate != rctarget.Server || len(persisted.Templates) != 2 {
		t.Fatalf("persisted preferences = %#v", persisted)
	}
}

func TestSetProjectQuotaProjectPersistsOnVersionOne(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects([]config.Project{{Name: "Demo", ProjectID: "demo", AuthID: "main"}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	project, previous, changed, err := svc.SetProjectQuotaProject("server@demo", "billing-project")
	if err != nil {
		t.Fatal(err)
	}
	if !changed || previous != "" || project.QuotaProjectID != "billing-project" {
		t.Fatalf("set result = %+v, previous %q, changed %v", project, previous, changed)
	}
	persisted, err := svc.ProjectByID("demo")
	if err != nil {
		t.Fatal(err)
	}
	if persisted.QuotaProjectID != "billing-project" {
		t.Fatalf("persisted quota project = %q", persisted.QuotaProjectID)
	}

	raw, err := os.ReadFile(config.GetProjectsFilePath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"version": 1`) {
		t.Fatalf("projects config = %s, want version 1", raw)
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
	if err == nil || !strings.Contains(err.Error(), `auth "missing" is not configured`) || !strings.Contains(err.Error(), authSetupHint) {
		t.Fatalf("BindProjectsAuth = %v, want missing auth error with setup guidance", err)
	}
}

func TestBindProjectIDsAuthUsesExactProjectIDs(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	projects := []config.Project{
		{Name: "Other", ProjectID: "alpha", AuthID: "old", DiscoveredBy: []string{"main"}},
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

func TestBindProjectsAuthSkipsProjectsNotDiscoveredByAuth(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if _, err := svc.AddGCloudAuth("main", "Main"); err != nil {
		t.Fatalf("AddGCloudAuth = %v", err)
	}
	projects := []config.Project{
		{Name: "Allowed", ProjectID: "allowed", AuthID: "old", DiscoveredBy: []string{"main", "old"}},
		{Name: "Denied", ProjectID: "denied", AuthID: "old", DiscoveredBy: []string{"old"}},
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	result, err := svc.BindProjectsAuth(nil, "main")
	if err != nil {
		t.Fatalf("BindProjectsAuth = %v", err)
	}
	if len(result.Bound) != 1 || result.Bound[0].ProjectID != "allowed" {
		t.Fatalf("bound = %+v, want allowed", result.Bound)
	}
	if len(result.Skipped) != 1 || result.Skipped[0].Project.ProjectID != "denied" {
		t.Fatalf("skipped = %+v, want denied", result.Skipped)
	}
	denied, err := svc.ProjectByID("denied")
	if err != nil {
		t.Fatal(err)
	}
	if denied.AuthID != "old" {
		t.Fatalf("denied auth = %q, want old", denied.AuthID)
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
	var lookupErr *ProjectLookupError
	if err == nil || !errors.As(err, &lookupErr) || lookupErr.Query != "=missing" || !strings.Contains(err.Error(), "no projects matched") {
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

func TestDisabledProjectCannotCreateFirebaseClient(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects([]config.Project{{ProjectID: "demo", AuthID: "main", Disabled: true}}, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	_, err := svc.firebaseServiceForProject(context.Background(), "demo")
	if err == nil || !strings.Contains(err.Error(), `project "demo" is disabled`) {
		t.Fatalf("firebaseServiceForProject = %v, want disabled error", err)
	}

	_, err = svc.firebaseServiceForProject(context.Background(), "server@demo")
	if err == nil || !strings.Contains(err.Error(), `project "server@demo" is disabled`) {
		t.Fatalf("firebaseServiceForProject server target = %v, want disabled error for underlying project", err)
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
	}).WithQuotaProjectOverride("test-quota-project")
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

func TestSyncProjectsForAuthEmptyRegistryKeepsRequestedIdentityInError(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveAuth(&config.AuthFile{Version: config.AuthConfigVersion}); err != nil {
		t.Fatalf("SaveAuth empty = %v", err)
	}

	_, _, err := svc.SyncProjectsForAuth(context.Background(), "main")
	if err == nil || !strings.Contains(err.Error(), `auth "main" is not configured`) || !strings.Contains(err.Error(), authSetupHint) {
		t.Fatalf("SyncProjectsForAuth = %v, want requested auth error with setup guidance", err)
	}
}

func TestSyncProjectsRequiresQuotaProjectBeforeLoadingOAuthOrCallingAPI(t *testing.T) {
	svc := setupCoreTestEnv(t)
	auth := config.DefaultOAuthAuthEntry("main", "Main")
	if err := config.SaveAuth(&config.AuthFile{
		Version: config.AuthConfigVersion, DefaultAuthID: "main", Auth: []config.AuthEntry{auth},
	}); err != nil {
		t.Fatal(err)
	}
	called := false
	svc.InjectFirebaseService("main", firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		called = true
		return nil, errors.New("unexpected API request")
	})}))

	_, _, err := svc.SyncProjectsForAuth(context.Background(), "main")
	var authErr *AuthError
	var quotaErr *firebase.QuotaProjectRequiredError
	if !errors.As(err, &authErr) || authErr.Kind != "quota_project_required" || !errors.As(err, &quotaErr) {
		t.Fatalf("SyncProjectsForAuth error = %#v", err)
	}
	if called {
		t.Fatal("Firebase API was called before quota-project resolution")
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

func TestListProjectsForExecutionUsesDirectDiscoveryWithoutLocalState(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects([]config.Project{{Name: "Cached", ProjectID: "cached", AuthID: "main"}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	direct := firebase.NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.Method == http.MethodGet && req.URL.Host == "cloudresourcemanager.googleapis.com" && req.URL.Path == "/v1/projects":
				return jsonResponse(http.StatusOK, `{"projects":[{"projectId":"remote","projectNumber":"123","name":"projects/123","lifecycleState":"ACTIVE"}]}`, ""), nil
			case req.Method == http.MethodGet && req.URL.Path == "/v3/projects/remote":
				return jsonResponse(http.StatusOK, `{"projectId":"remote","displayName":"Remote","state":"ACTIVE","etag":"etag-1","updateTime":"2026-08-20T10:00:00Z"}`, ""), nil
			default:
				return nil, errors.New("unexpected request: " + req.URL.String())
			}
		}),
	}).WithQuotaProjectOverride("test-quota-project")
	ctx := WithExecutionPolicy(context.Background(), StatelessExecutionPolicy())
	ctx, err := WithDirectFirebaseDiscoveryService(ctx, direct)
	if err != nil {
		t.Fatal(err)
	}

	got, source, err := svc.ListProjectsForExecution(ctx)
	if err != nil {
		t.Fatalf("ListProjectsForExecution = %v", err)
	}
	if source != "firebase" || len(got) != 1 || got[0].ProjectID != "remote" || got[0].Name != "Remote" {
		t.Fatalf("projects/source = %+v/%q, want direct remote project", got, source)
	}
	persisted, err := config.LoadProjects()
	if err != nil || len(persisted) != 1 || persisted[0].ProjectID != "cached" {
		t.Fatalf("persisted projects = %+v, %v; want untouched cache", persisted, err)
	}
}

func TestListProjectsForExecutionRequiresDirectDiscoveryWithoutLocalState(t *testing.T) {
	svc := setupCoreTestEnv(t)
	ctx := WithExecutionPolicy(context.Background(), StatelessExecutionPolicy())

	_, _, err := svc.ListProjectsForExecution(ctx)
	var policyErr *ExecutionPolicyError
	if !errors.As(err, &policyErr) || policyErr.Operation != "configured Firebase project discovery" {
		t.Fatalf("ListProjectsForExecution error = %#v, want discovery ExecutionPolicyError", err)
	}
}

func TestResetProjectsDeletesRegistry(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects([]config.Project{{ProjectID: "demo"}}, time.Now().UTC()); err != nil {
		t.Fatalf("SaveProjects = %v", err)
	}

	changed, err := svc.ResetProjects()
	if err != nil {
		t.Fatalf("ResetProjects = %v", err)
	}
	if !changed {
		t.Fatal("ResetProjects changed = false, want true")
	}
	_, err = config.LoadProjects()
	if err == nil {
		t.Fatal("LoadProjects after reset = nil, want error")
	}

	changed, err = svc.ResetProjects()
	if err != nil {
		t.Fatalf("second ResetProjects = %v", err)
	}
	if changed {
		t.Fatal("second ResetProjects changed = true, want false")
	}
}
