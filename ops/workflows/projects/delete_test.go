package projects

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func TestForgetCommandFiltersLocallyAndSkipsConfirmation(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{
		{Name: "Alpha App", ProjectID: "alpha", AuthID: "main"},
		{Name: "Beta App", ProjectID: "beta", AuthID: "main"},
	}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	cmd := New(svc)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"forget", "--filter", "/app", "--expr", `project_id == "alpha"`, "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute forget = %v", err)
	}
	if !strings.Contains(out.String(), "Alpha App (alpha)") {
		t.Fatalf("output = %q, want forgotten alpha", out.String())
	}
	projects, err := config.LoadProjects()
	if err != nil || len(projects) != 1 || projects[0].ProjectID != "beta" {
		t.Fatalf("remaining projects = %+v, err=%v; want beta", projects, err)
	}
}

func TestProjectForgetConfirmationPromptPluralizesProject(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{count: 1, want: "Forget 1 project and delete all associated local caches, versions, and drafts?"},
		{count: 2, want: "Forget 2 projects and delete all associated local caches, versions, and drafts?"},
	}
	for _, test := range tests {
		if got := projectForgetConfirmationPrompt(test.count); got != test.want {
			t.Fatalf("projectForgetConfirmationPrompt(%d) = %q, want %q", test.count, got, test.want)
		}
	}
}
