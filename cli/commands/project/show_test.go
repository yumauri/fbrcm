package project

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestShowCommandPrintsProjectDetailsAndAuthIdentities(t *testing.T) {
	project := config.Project{
		Name:          "Alpha",
		ProjectID:     "alpha-project",
		ProjectNumber: "123",
		State:         "ACTIVE",
		ETag:          "etag-value",
		AuthID:        "main",
		Disabled:      true,
		DiscoveredBy:  []string{"main", "work"},
		UpdatedAt:     "2026-07-19T10:11:12Z",
		SyncedAt:      "2026-07-20T11:12:13Z",
	}
	svc := saveProjectsForTest(t, []config.Project{project})
	cmd := newShowCommand(svc)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"alpha"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute show command: %v", err)
	}
	for _, want := range []string{
		"Project: Alpha",
		"Project ID: alpha-project",
		"Status: disabled",
		"Number: 123",
		"State: ACTIVE",
		"Selected auth: main",
		"Auth identities: main, work",
		"Updated at: " + shared.FormatDateTime(project.UpdatedAt),
		"Synced at: " + shared.FormatDateTime(project.SyncedAt),
		"ETag: etag-value",
		"URL: " + firebase.RemoteConfigConsoleURL(project.ProjectID),
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("show output = %q, want substring %q", out.String(), want)
		}
	}
}

func TestShowCommandJSONUsesProjectListContract(t *testing.T) {
	project := config.Project{
		Name:         "Alpha",
		ProjectID:    "alpha-project",
		AuthID:       "main",
		DiscoveredBy: []string{"main", "work"},
	}
	svc := saveProjectsForTest(t, []config.Project{project})
	cmd := newShowCommand(svc)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"alpha", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute show command: %v", err)
	}
	var got shared.ProjectJSON
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode show JSON: %v", err)
	}
	if got.ProjectID != project.ProjectID || got.AuthID != "main" || strings.Join(got.DiscoveredBy, ",") != "main,work" {
		t.Fatalf("show JSON = %#v", got)
	}
	if got.URL != firebase.RemoteConfigConsoleURL(project.ProjectID) {
		t.Fatalf("show URL = %q", got.URL)
	}
}
