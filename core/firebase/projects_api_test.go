package firebase

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestListProjectsAndGetProject(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.Method == http.MethodGet && req.URL.Path == "/v1/projects":
				body := `{"projects":[{"projectId":"demo","name":"Demo","lifecycleState":"ACTIVE"}]}`
				return jsonHTTPResponse(http.StatusOK, body, ""), nil
			case req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/v3/projects/"):
				body := `{"projectId":"demo","displayName":"Demo Display","state":"ACTIVE","etag":"e1","updateTime":"2026-01-01T00:00:00Z"}`
				return jsonHTTPResponse(http.StatusOK, body, ""), nil
			default:
				return nil, io.EOF
			}
		}),
	})

	projects, err := svc.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects = %v", err)
	}
	if len(projects) != 1 || projects[0].ProjectID != "demo" {
		t.Fatalf("projects = %+v", projects)
	}
	if projects[0].Name != "Demo Display" {
		t.Fatalf("enriched name = %q, want Demo Display", projects[0].Name)
	}

	project, err := svc.GetProject(context.Background(), "demo")
	if err != nil || project.Name != "Demo Display" {
		t.Fatalf("GetProject = %+v err=%v", project, err)
	}
}

func TestListProjectsSkipsDeleteRequested(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/v1/projects" {
				body := `{"projects":[{"projectId":"gone","lifecycleState":"DELETE_REQUESTED"},{"projectId":"keep","name":"Keep","lifecycleState":"ACTIVE"}]}`
				return jsonHTTPResponse(http.StatusOK, body, ""), nil
			}
			if strings.HasPrefix(req.URL.Path, "/v3/projects/keep") {
				return jsonHTTPResponse(http.StatusOK, `{"projectId":"keep","displayName":"Keep","state":"ACTIVE"}`, ""), nil
			}
			return nil, io.EOF
		}),
	})

	projects, err := svc.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects = %v", err)
	}
	if len(projects) != 1 || projects[0].ProjectID != "keep" {
		t.Fatalf("projects = %+v, want keep only", projects)
	}
}

func TestProjectMergeInto(t *testing.T) {
	base := Project{Name: "Old", ProjectID: "demo", State: "ACTIVE"}
	details := Project{Name: "New", ETag: "e2", UpdateTime: "2026-01-02T00:00:00Z"}
	merged := details.mergeInto(base)
	if merged.Name != "New" || merged.ETag != "e2" || merged.ProjectID != "demo" {
		t.Fatalf("mergeInto = %+v", merged)
	}
}
