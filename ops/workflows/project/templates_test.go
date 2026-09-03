package project

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func TestTemplatesShowPrintsNormalizedClientDefault(t *testing.T) {
	saveProjectsForTest(t, []config.Project{{
		Name:      "Northstar Wallet",
		ProjectID: "northstar-wallet",
		AuthID:    "main",
	}})
	cmd := newTemplatesShowCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"northstar-wallet"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute templates show: %v", err)
	}
	for _, want := range []string{
		"Project: Northstar Wallet",
		"Project ID: northstar-wallet",
		"Enabled templates: client",
		"Primary template: client",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("templates show output = %q, want %q", out.String(), want)
		}
	}
}

func TestTemplatesShowJSON(t *testing.T) {
	saveProjectsForTest(t, []config.Project{{
		Name:            "Northstar Wallet",
		ProjectID:       "northstar-wallet",
		AuthID:          "main",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Server,
	}})
	cmd := newTemplatesShowCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"northstar-wallet", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute templates show JSON: %v", err)
	}
	var got projectTemplatesJSON
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode templates JSON: %v", err)
	}
	if got.Project != "Northstar Wallet" ||
		got.ProjectID != "northstar-wallet" ||
		len(got.Templates) != 2 ||
		got.Templates[0] != rctarget.Client ||
		got.Templates[1] != rctarget.Server ||
		got.PrimaryTemplate != rctarget.Server {
		t.Fatalf("templates JSON = %#v", got)
	}
}

func TestTemplatesSetUpdatesBothTemplatesAndPrimary(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{
		Name:      "Northstar Wallet",
		ProjectID: "northstar-wallet",
		AuthID:    "main",
	}})
	cmd := newTemplatesSetCommand(svc)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"northstar-wallet",
		"--templates", "client,server",
		"--primary", "server",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute templates set: %v", err)
	}
	project, err := svc.ProjectByID("northstar-wallet")
	if err != nil {
		t.Fatal(err)
	}
	if len(project.Templates) != 2 ||
		project.Templates[0] != rctarget.Client ||
		project.Templates[1] != rctarget.Server ||
		project.PrimaryTemplate != rctarget.Server {
		t.Fatalf("persisted templates = %v/%q", project.Templates, project.PrimaryTemplate)
	}
	if !strings.Contains(out.String(), "Enabled templates: client, server") ||
		!strings.Contains(out.String(), "Primary template: server") {
		t.Fatalf("templates set output = %q", out.String())
	}
}

func TestTemplatesSetSingleTemplateMakesItPrimary(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{
		Name:            "Northstar Wallet",
		ProjectID:       "northstar-wallet",
		AuthID:          "main",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Client,
	}})
	cmd := newTemplatesSetCommand(svc)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"northstar-wallet", "--templates", "server"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute templates set server-only: %v", err)
	}
	project, err := svc.ProjectByID("northstar-wallet")
	if err != nil {
		t.Fatal(err)
	}
	if len(project.Templates) != 1 ||
		project.Templates[0] != rctarget.Server ||
		project.PrimaryTemplate != rctarget.Server {
		t.Fatalf("persisted templates = %v/%q, want server/server", project.Templates, project.PrimaryTemplate)
	}
}

func TestTemplatesSetJSONAcceptsRepeatedCaseInsensitiveTemplates(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{
		Name:      "Northstar Wallet",
		ProjectID: "northstar-wallet",
		AuthID:    "main",
	}})
	cmd := newTemplatesSetCommand(svc)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{
		"northstar-wallet",
		"--templates", "Server",
		"--templates", "CLIENT",
		"--primary", "SERVER",
		"--json",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute templates set JSON: %v", err)
	}
	var got projectTemplatesJSON
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode templates set JSON: %v", err)
	}
	if len(got.Templates) != 2 ||
		got.Templates[0] != rctarget.Client ||
		got.Templates[1] != rctarget.Server ||
		got.PrimaryTemplate != rctarget.Server {
		t.Fatalf("templates set JSON = %#v", got)
	}
}

func TestTemplatesSetPrimaryOnlyPreservesEnabledTemplates(t *testing.T) {
	svc := saveProjectsForTest(t, []config.Project{{
		Name:            "Northstar Wallet",
		ProjectID:       "northstar-wallet",
		AuthID:          "main",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Client,
	}})
	cmd := newTemplatesSetCommand(svc)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"northstar-wallet", "--primary", "server"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute templates set primary-only: %v", err)
	}
	project, err := svc.ProjectByID("northstar-wallet")
	if err != nil {
		t.Fatal(err)
	}
	if len(project.Templates) != 2 || project.PrimaryTemplate != rctarget.Server {
		t.Fatalf("persisted templates = %v/%q, want both/server", project.Templates, project.PrimaryTemplate)
	}
}

func TestTemplatesSetValidatesMutation(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "requires mutation",
			args: []string{"northstar-wallet"},
			want: "at least one of --templates or --primary is required",
		},
		{
			name: "rejects invalid template",
			args: []string{"northstar-wallet", "--templates", "mobile"},
			want: `template must be client or server, got "mobile"`,
		},
		{
			name: "rejects disabled primary",
			args: []string{"northstar-wallet", "--templates", "client", "--primary", "server"},
			want: `primary template "server" is not enabled`,
		},
		{
			name: "rejects target prefix",
			args: []string{"server@northstar-wallet", "--primary", "server"},
			want: "omit the server@ prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := saveProjectsForTest(t, []config.Project{{
				Name:      "Northstar Wallet",
				ProjectID: "northstar-wallet",
				AuthID:    "main",
			}})
			cmd := newTemplatesSetCommand(svc)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("templates set error = %v, want %q", err, tt.want)
			}
			project, loadErr := svc.ProjectByID("northstar-wallet")
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			if len(project.Templates) != 1 ||
				project.Templates[0] != rctarget.Client ||
				project.PrimaryTemplate != rctarget.Client {
				t.Fatalf("invalid mutation changed project to %v/%q", project.Templates, project.PrimaryTemplate)
			}
		})
	}
}
