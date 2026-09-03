package project

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
)

func TestProjectQuotaProjectSetShowUnset(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{
		Name: "Demo", ProjectID: "demo", AuthID: "main",
	}})
	if err := config.SaveAuth(&config.AuthFile{
		Version:       config.AuthConfigVersion,
		DefaultAuthID: "main",
		Auth: []config.AuthEntry{{
			ID: "main", Type: config.AuthTypeGCloud, Label: "Main", QuotaProjectID: "auth-billing-project",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	set := newProjectQuotaProjectSetCommand(svc)
	if err := set.RunE(set, []string{"demo", "project-billing-project"}); err != nil {
		t.Fatal(err)
	}
	show := newProjectQuotaProjectShowCommand(svc)
	var out bytes.Buffer
	show.SetOut(&out)
	if err := show.RunE(show, []string{"demo"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Effective quota project: project-billing-project") || !strings.Contains(out.String(), "Source: project") {
		t.Fatalf("show output = %q", out.String())
	}

	unset := newProjectQuotaProjectUnsetCommand(svc)
	if err := unset.RunE(unset, []string{"demo"}); err != nil {
		t.Fatal(err)
	}
	project, err := svc.ProjectByID("demo")
	if err != nil {
		t.Fatal(err)
	}
	if project.QuotaProjectID != "" {
		t.Fatalf("quota project after unset = %q", project.QuotaProjectID)
	}

	projects, err := config.LoadProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].QuotaProjectID != "" {
		t.Fatalf("persisted projects = %+v", projects)
	}
}
