package shared

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	charmlog "charm.land/log/v2"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func TestBulkProjectRequestsShareRegistryRead(t *testing.T) {
	setupProjectAliasResolutionTest(t, "")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	projects := []core.Project{
		{ProjectID: "alpha", AuthID: "first", QuotaProjectID: "billing-alpha"},
		{ProjectID: "beta", AuthID: "second", QuotaProjectID: "billing-beta"},
		{ProjectID: "gamma", AuthID: "first", QuotaProjectID: "billing-gamma"},
		{ProjectID: "delta", AuthID: "second", QuotaProjectID: "billing-delta"},
		{ProjectID: "hidden", AuthID: "first"},
	}
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	policy := config.DefaultProjectVisibility()
	policy.Blacklist.ProjectIDs = []string{"hidden"}
	if err := config.SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	for _, authID := range []string{"first", "second"} {
		svc.InjectFirebaseService(authID, firebase.NewServiceWithHTTPClient(&http.Client{
			Transport: registryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				for _, project := range projects {
					if req.URL.Path != "/v1/projects/"+project.ProjectID+"/remoteConfig" {
						continue
					}
					if project.AuthID != authID || req.Header.Get("X-Goog-User-Project") != project.QuotaProjectID {
						return nil, fmt.Errorf("incorrect auth or quota project for %s", project.ProjectID)
					}
					return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Etag": {"test-etag"}}, Body: io.NopCloser(strings.NewReader(`{"parameters":{},"version":{"versionNumber":"1"}}`))}, nil
				}
				return nil, fmt.Errorf("unexpected request: %s", req.URL)
			}),
		}))
	}
	t.Setenv(env.LogLevel, "debug")
	logFile, err := os.CreateTemp(t.TempDir(), "logs")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = logFile.Close() })
	previousLevel := corelog.CurrentLevel()
	corelog.ConfigureCLIOutput(logFile, logFile)
	corelog.InitWithDefault(corelog.ModeCLI, charmlog.DebugLevel)
	t.Cleanup(func() {
		corelog.ConfigureCLIOutput(os.Stderr, os.Stderr)
		corelog.SetLevel(previousLevel)
	})
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	selected, ctx, err := ResolveProjectTargetsForExecution(cmd.Context(), cmd, svc, nil)
	if err != nil || len(selected) != 4 {
		t.Fatalf("selection = %#v, %v", selected, err)
	}
	// Client resolution must use the selection's snapshot, even concurrently.
	if err := os.WriteFile(config.GetProjectsFilePath(), []byte("{"), config.PrivateFileMode); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for _, project := range selected {
		workers.Go(func() {
			if _, _, err := svc.ExportRemoteConfig(ctx, project.ProjectID); err != nil {
				t.Errorf("request for %s: %v", project.ProjectID, err)
			}
		})
	}
	workers.Wait()
	logs, err := os.ReadFile(logFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []string{"read projects config", "loaded projects config"} {
		if count := strings.Count(string(logs), record); count != 1 {
			t.Errorf("%q count = %d, want 1\n%s", record, count, logs)
		}
	}
	if _, _, err := ResolveProjectTargetsForExecution(cmd.Context(), cmd, svc, nil); err == nil {
		t.Fatal("later selection must reread the current registry")
	}
}

type registryRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn registryRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
