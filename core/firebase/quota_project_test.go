package firebase

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	coreenv "github.com/yumauri/fbrcm/core/env"
)

func TestQuotaProjectPolicyPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		policy quotaProjectPolicy
		target string
		want   string
	}{
		{
			name: "environment overrides credentials and target",
			policy: quotaProjectPolicy{
				environmentQuotaProjectID: "environment-project",
				credentialQuotaProjectID:  "credential-project",
				useTargetProjectQuota:     true,
			},
			target: "target-project",
			want:   "environment-project",
		},
		{
			name: "credentials override target",
			policy: quotaProjectPolicy{
				credentialQuotaProjectID: "credential-project",
				useTargetProjectQuota:    true,
			},
			target: "target-project",
			want:   "credential-project",
		},
		{name: "gcloud target fallback", policy: quotaProjectPolicy{useTargetProjectQuota: true}, target: "target-project", want: "target-project"},
		{name: "project listing has no target fallback", policy: quotaProjectPolicy{useTargetProjectQuota: true}},
		{name: "oauth and service account have no implicit target fallback", target: "target-project"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.policy.projectID(test.target); got != test.want {
				t.Fatalf("projectID(%q) = %q, want %q", test.target, got, test.want)
			}
		})
	}
}

func TestEnvironmentQuotaProjectID(t *testing.T) {
	for name, value := range map[string]string{"empty": "", "spaces": "   ", "tab": "\t"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(coreenv.GoogleCloudQuotaProject, value)
			got, err := environmentQuotaProjectID()
			if err != nil || got != "" {
				t.Fatalf("environmentQuotaProjectID() = %q, %v; want absent", got, err)
			}
		})
	}

	t.Setenv(coreenv.GoogleCloudQuotaProject, "  billing-project  ")
	got, err := environmentQuotaProjectID()
	if err != nil || got != "billing-project" {
		t.Fatalf("environmentQuotaProjectID() = %q, %v", got, err)
	}
}

func TestEnvironmentQuotaProjectRejectsUnsafeValueBeforeAuthLoading(t *testing.T) {
	t.Setenv(coreenv.GoogleCloudQuotaProject, "billing project")
	for _, construct := range []struct {
		name string
		load func() error
	}{
		{name: "normal", load: func() error {
			_, err := NewServiceForAuth(context.Background(), config.AuthEntry{Type: "unsupported"}, false)
			return err
		}},
		{name: "diagnostic", load: func() error {
			_, err := NewDiagnosticServiceForAuth(context.Background(), config.AuthEntry{Type: "unsupported"})
			return err
		}},
		{name: "access token", load: func() error {
			_, err := NewServiceWithAccessToken(context.Background(), "access-token")
			return err
		}},
	} {
		t.Run(construct.name, func(t *testing.T) {
			err := construct.load()
			var quotaErr *QuotaProjectError
			if !errors.As(err, &quotaErr) || quotaErr.Source != QuotaProjectSourceEnvironment || quotaErr.Variable != coreenv.GoogleCloudQuotaProject {
				t.Fatalf("error = %#v, want typed environment quota-project error", err)
			}
		})
	}
}

func TestCredentialQuotaProjectID(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/that/must/not/be-read.json")
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{name: "metadata credentials"},
		{name: "missing quota project", data: []byte(`{"type":"service_account"}`)},
		{name: "selected credential JSON", data: []byte(`{"quota_project_id":"  credential-project  "}`), want: "credential-project"},
		{name: "malformed JSON", data: []byte(`{`), wantErr: true},
		{name: "unsafe quota project", data: []byte(`{"quota_project_id":"billing project"}`), wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := credentialQuotaProjectID(test.data)
			if test.wantErr {
				var quotaErr *QuotaProjectError
				if !errors.As(err, &quotaErr) || quotaErr.Source != QuotaProjectSourceCredentials {
					t.Fatalf("error = %#v, want typed credential quota-project error", err)
				}
				return
			}
			if err != nil || got != test.want {
				t.Fatalf("credentialQuotaProjectID() = %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestServiceFromAuthResultAppliesEnvironmentAcrossIdentityPolicies(t *testing.T) {
	for _, test := range []struct {
		name   string
		result authHTTPClientResult
	}{
		{name: "oauth", result: authHTTPClientResult{client: http.DefaultClient}},
		{name: "service account", result: authHTTPClientResult{client: http.DefaultClient}},
		{name: "gcloud", result: authHTTPClientResult{client: http.DefaultClient, credentialQuotaProjectID: "adc-project", useTargetProjectQuota: true}},
	} {
		t.Run(test.name, func(t *testing.T) {
			svc := serviceFromAuthHTTPClientResult(test.result, "environment-project")
			req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
			if err != nil {
				t.Fatal(err)
			}
			svc.setQuotaProject(req, "target-project")
			if got := req.Header.Get("X-Goog-User-Project"); got != "environment-project" {
				t.Fatalf("quota project header = %q", got)
			}
		})
	}
}

func TestQuotaProjectHeaderCoversFirebaseAndResourceManagerRequests(t *testing.T) {
	const wantQuotaProject = "billing-project"
	requestCount := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		if got := req.Header.Get("X-Goog-User-Project"); got != wantQuotaProject {
			t.Fatalf("%s %s quota project = %q, want %q", req.Method, req.URL.RequestURI(), got, wantQuotaProject)
		}
		switch {
		case req.URL.Path == "/v1/projects":
			return jsonHTTPResponse(http.StatusOK, `{"projects":[]}`, ""), nil
		case strings.HasPrefix(req.URL.Path, "/v3/projects/"):
			return jsonHTTPResponse(http.StatusOK, `{"projectId":"target-project","displayName":"Target","state":"ACTIVE"}`, ""), nil
		case strings.HasSuffix(req.URL.Path, ":testIamPermissions"):
			return jsonHTTPResponse(http.StatusOK, `{"permissions":[]}`, ""), nil
		case strings.HasSuffix(req.URL.Path, ":listVersions"):
			return jsonHTTPResponse(http.StatusOK, `{"versions":[]}`, ""), nil
		case strings.HasSuffix(req.URL.Path, ":rollback"):
			return jsonHTTPResponse(http.StatusOK, `{}`, "etag"), nil
		case strings.HasSuffix(req.URL.Path, ":downloadDefaults"):
			return jsonHTTPResponse(http.StatusOK, `{}`, ""), nil
		case strings.Contains(req.URL.Path, "/experiments"):
			return jsonHTTPResponse(http.StatusOK, `{"experiments":[]}`, ""), nil
		case strings.Contains(req.URL.Path, "/rollouts/") && req.Method == http.MethodDelete:
			return jsonHTTPResponse(http.StatusNoContent, "", ""), nil
		case strings.HasSuffix(req.URL.Path, "/remoteConfig"):
			return jsonHTTPResponse(http.StatusOK, `{}`, "etag"), nil
		default:
			return nil, io.EOF
		}
	})}
	svc := &Service{
		httpClient: client,
		quotaProjectPolicy: quotaProjectPolicy{
			environmentQuotaProjectID: wantQuotaProject,
			credentialQuotaProjectID:  "credential-project",
			useTargetProjectQuota:     true,
		},
	}
	ctx := context.Background()
	if _, err := svc.ListProjects(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetProject(ctx, "target-project"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.TestProjectPermissions(ctx, "target-project", []string{"cloudconfig.configs.get"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.GetRemoteConfig(ctx, "target-project"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ValidateRemoteConfig(ctx, "target-project", []byte(`{}`), "etag"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.UpdateRemoteConfig(ctx, "target-project", []byte(`{}`), "etag"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ListRemoteConfigVersions(ctx, "target-project", ListVersionsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RollbackRemoteConfig(ctx, "target-project", "1"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DownloadRemoteConfigDefaults(ctx, "target-project", DefaultsFormatJSON); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ListExperiments(ctx, "123", "target-project", ListManagedFeaturesOptions{}); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteRollout(ctx, "123", "target-project", "rollout-1"); err != nil {
		t.Fatal(err)
	}
	if requestCount != 11 {
		t.Fatalf("request count = %d, want 11", requestCount)
	}
}
