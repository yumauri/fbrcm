package project

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestOpenCommandResolvesProjectAndOpensRemoteConfigConsole(t *testing.T) {
	projects := []config.Project{
		{Name: "Alpha", ProjectID: "production-project", AuthID: "main"},
		{Name: "production-project", ProjectID: "name-collision", AuthID: "main"},
		{Name: "Production", ProjectID: "production-us", AuthID: "main"},
		{Name: "Production EU", ProjectID: "production-eu", AuthID: "main"},
		{Name: "Staging", ProjectID: "staging-project", AuthID: "main"},
	}
	svc := saveProjectsForTest(t, projects)

	tests := []struct {
		name      string
		query     string
		projectID string
	}{
		{name: "exact project id wins", query: "production-project", projectID: "production-project"},
		{name: "exact project name wins", query: "Production", projectID: "production-us"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var openedURL string
			cmd := newOpenCommand(svc, func(url string) error {
				openedURL = url
				return nil
			})
			cmd.SetArgs([]string{tt.query})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute open command: %v", err)
			}
			want := firebase.RemoteConfigConsoleURL(tt.projectID)
			if openedURL != want {
				t.Fatalf("opened URL = %q, want %q", openedURL, want)
			}
		})
	}
	for _, query := range []string{"PRODUCTION-PROJECT", "production", "stag", " staging-project"} {
		cmd := newOpenCommand(svc, func(string) error { return nil })
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetArgs([]string{query})
		if err := cmd.Execute(); err == nil {
			t.Fatalf("non-exact project selector %q unexpectedly resolved", query)
		}
	}
}

func TestOpenCommandJSONReturnsURLWithoutOpeningBrowser(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{Name: "Production", ProjectID: "production-project", AuthID: "main"}})
	opened := false
	cmd := newOpenCommand(svc, func(string) error {
		opened = true
		return nil
	})
	cmd.Flags().Bool("json", false, "")
	cmd.SetArgs([]string{"production-project", "--json"})
	var output bytes.Buffer
	cmd.SetOut(&output)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute open command: %v", err)
	}
	if opened {
		t.Fatal("JSON mode opened a browser")
	}
	var result struct {
		ProjectID string `json:"project_id"`
		URL       string `json:"url"`
		Opened    bool   `json:"opened"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.ProjectID != "production-project" || result.URL != firebase.RemoteConfigConsoleURL("production-project") || result.Opened {
		t.Fatalf("result = %#v", result)
	}
}
