package core

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func TestProjectRegistrySnapshotIsolationAndFreshReads(t *testing.T) {
	svc := setupCoreTestEnv(t)
	projects := []Project{{ProjectID: "alpha", AuthID: "main", Templates: []rctarget.Kind{rctarget.Client}, DiscoveredBy: []string{"main"}}}
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	ctx := WithProjectRegistrySnapshot(context.Background(), projects)
	projects[0].AuthID = "other"
	projects[0].Templates[0] = rctarget.Server
	projects[0].DiscoveredBy[0] = "other"
	if err := os.WriteFile(config.GetProjectsFilePath(), []byte("{"), config.PrivateFileMode); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		project, err := svc.projectByIDForContext(ctx, "alpha")
		if err != nil || project.AuthID != "main" || project.Templates[0] != rctarget.Client || project.DiscoveredBy[0] != "main" {
			t.Fatalf("snapshot changed: %#v, %v", project, err)
		}
		project.Templates[0] = rctarget.Server
		project.DiscoveredBy[0] = "other"
	}
	if _, err := svc.projectByIDForContext(ctx, "missing"); err == nil || !strings.Contains(err.Error(), "not in projects config") {
		t.Fatalf("snapshot miss = %v", err)
	}
	if _, err := svc.ProjectByID("alpha"); err == nil {
		t.Fatal("later operations must read the current file and fail closed")
	}
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	if project, err := svc.ProjectByID("alpha"); err != nil || project.AuthID != "other" {
		t.Fatalf("later lookup did not observe registry update: %#v, %v", project, err)
	}
	// An operation's snapshot must never select an identity from another profile.
	if err := config.SwitchProfile("second"); err != nil {
		t.Fatal(err)
	}
	projects[0].AuthID = "second-profile"
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	if project, err := svc.projectByIDForContext(ctx, "alpha"); err != nil || project.AuthID != "second-profile" {
		t.Fatalf("cross-profile snapshot = %#v, %v", project, err)
	}
}

func TestProjectRegistrySnapshotPreservesClientGuards(t *testing.T) {
	svc := setupCoreTestEnv(t)
	ctx := WithProjectRegistrySnapshot(context.Background(), []Project{{ProjectID: "alpha", AuthID: "main", Disabled: true}})
	if _, err := svc.firebaseServiceForProject(ctx, "server@alpha"); err == nil || !strings.Contains(err.Error(), "is disabled") {
		t.Fatalf("disabled snapshot project = %v", err)
	}
	stateless := WithExecutionPolicy(ctx, StatelessExecutionPolicy())
	if WithProjectRegistrySnapshot(stateless, nil) != stateless {
		t.Fatal("stateless execution must not capture a registry")
	}
	var executionErr *ExecutionPolicyError
	if _, err := svc.firebaseServiceForProject(stateless, "alpha"); !errors.As(err, &executionErr) {
		t.Fatalf("stateless snapshot lookup = %v", err)
	}
	direct := firebase.NewServiceWithHTTPClient(nil)
	stateless, err := WithDirectFirebaseService(stateless, "alpha", direct)
	if err != nil {
		t.Fatal(err)
	}
	if fb, err := svc.firebaseServiceForProject(stateless, "alpha"); err != nil || fb != direct {
		t.Fatalf("direct service = %p, %v", fb, err)
	}
}
