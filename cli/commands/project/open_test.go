package project

import (
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
		{name: "exact project id wins", query: "PRODUCTION-PROJECT", projectID: "production-project"},
		{name: "exact project name wins", query: "production", projectID: "production-us"},
		{name: "substring match", query: "stag", projectID: "staging-project"},
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
}
