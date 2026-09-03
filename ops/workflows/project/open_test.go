package project

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
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

func TestOpenCommandStatelessUsesLiteralProjectIDWithoutService(t *testing.T) {
	var openedURL string
	cmd := newOpenCommand(nil, func(url string) error {
		openedURL = url
		return nil
	})
	cmd.SetContext(core.WithExecutionPolicy(cmd.Context(), core.StatelessExecutionPolicy()))
	cmd.SetArgs([]string{"demo-project"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute stateless open command: %v", err)
	}
	if want := firebase.RemoteConfigConsoleURL("demo-project"); openedURL != want {
		t.Fatalf("opened URL = %q, want %q", openedURL, want)
	}

	for _, project := range []string{"=demo-project", "client@demo-project", "server@demo-project"} {
		cmd := newOpenCommand(nil, func(string) error { return nil })
		cmd.SetContext(core.WithExecutionPolicy(cmd.Context(), core.StatelessExecutionPolicy()))
		cmd.SetArgs([]string{project})
		if err := cmd.Execute(); err == nil {
			t.Errorf("stateless project open accepted %q", project)
		}
	}
}

func TestOpenCommandOfflinePrintsURLWithoutOpeningBrowser(t *testing.T) {
	firebase.SetOfflineMode(true)
	t.Cleanup(func() { firebase.SetOfflineMode(false) })
	statefulService := saveProjectsForTest(t, []config.Project{{Name: "Demo", ProjectID: "demo-project", AuthID: "main"}})

	tests := []struct {
		name      string
		service   *core.Core
		configure func(*cobra.Command)
	}{
		{name: "stateful", service: statefulService, configure: func(*cobra.Command) {}},
		{name: "stateless", configure: func(cmd *cobra.Command) {
			cmd.SetContext(core.WithExecutionPolicy(cmd.Context(), core.StatelessExecutionPolicy()))
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opened := false
			cmd := newOpenCommand(test.service, func(string) error {
				opened = true
				return nil
			})
			test.configure(cmd)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetArgs([]string{"demo-project"})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute offline project open: %v", err)
			}
			if opened {
				t.Fatal("offline project open launched a browser")
			}
			if want := firebase.RemoteConfigConsoleURL("demo-project") + "\n"; output.String() != want {
				t.Fatalf("offline output = %q, want %q", output.String(), want)
			}
		})
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
