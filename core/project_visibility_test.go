package core

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
)

func TestVisibleProjectsPreservesInventoryAndSkipsStatelessReads(t *testing.T) {
	svc := setupCoreTestEnv(t)
	projects := []Project{{ProjectID: "alpha", AuthID: "main"}, {ProjectID: "beta", AuthID: "main", Disabled: true}}
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	policy := config.DefaultProjectVisibility()
	policy.Blacklist.ProjectIDs = []string{"alpha"}
	if err := config.SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	state, err := svc.InspectStartupState()
	if err != nil || len(state.Projects) != 2 {
		t.Fatalf("startup must retain inventory for setup detection: %#v, %v", state, err)
	}
	visible, err := VisibleProjects(ctx, projects)
	if err != nil || len(visible) != 1 || visible[0].ProjectID != "beta" || !visible[0].Disabled {
		t.Fatalf("visible = %#v, %v", visible, err)
	}
	var hidden *ProjectHiddenError
	if err := RequireProjectVisible(ctx, "server@alpha"); !errors.As(err, &hidden) || hidden.Reason != "blacklisted" {
		t.Fatalf("exact hidden target = %v", err)
	}
	if err := RequireProjectVisible(WithHiddenProjects(ctx, true), "alpha"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeleteProjectIDs([]string{"alpha"}); err != nil {
		t.Fatal(err)
	}
	// Rediscovery can rebuild a forgotten record without touching policy.
	merged := mergeProjects(nil, projects, "main", []string{"main"}, "", time.Now())
	if err := config.SaveProjects(merged, time.Now()); err != nil {
		t.Fatal(err)
	}
	visible, err = VisibleProjects(ctx, merged)
	if err != nil || len(visible) != 1 || visible[0].ProjectID != "beta" {
		t.Fatalf("rediscovered visible = %#v, %v", visible, err)
	}
	// A malformed local file must never be read by stateless execution.
	if err := os.WriteFile(config.GetProjectVisibilityFilePath(), []byte("{"), config.PrivateFileMode); err != nil {
		t.Fatal(err)
	}
	stateless := WithExecutionPolicy(ctx, StatelessExecutionPolicy())
	visible, err = VisibleProjects(stateless, projects)
	if err != nil || len(visible) != 2 {
		t.Fatalf("stateless = %#v, %v", visible, err)
	}
	if _, err := VisibleProjects(ctx, projects); err == nil {
		t.Fatal("malformed policy must fail closed")
	}
}
