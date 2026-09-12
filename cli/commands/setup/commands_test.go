package setup

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/ops/shared"
)

type fakePrompter struct {
	selections []string
	texts      []string
	files      []string
	selectAt   int
	textAt     int
	fileAt     int
}

func (p *fakePrompter) selectOne(_ string, choices []promptChoice) (string, error) {
	if p.selectAt >= len(p.selections) {
		return "", errors.New("unexpected selection prompt")
	}
	value := p.selections[p.selectAt]
	p.selectAt++
	for _, choice := range choices {
		if choice.Value == value {
			return value, nil
		}
	}
	return "", errors.New("queued selection is not available")
}

func (p *fakePrompter) text(_ string, _ string, validate func(string) error) (string, error) {
	if p.textAt >= len(p.texts) {
		return "", errors.New("unexpected text prompt")
	}
	value := p.texts[p.textAt]
	p.textAt++
	if validate != nil {
		if err := validate(value); err != nil {
			return "", err
		}
	}
	return value, nil
}

func (p *fakePrompter) pickJSON(_ string) (string, error) {
	if p.fileAt >= len(p.files) {
		return "", errors.New("unexpected file prompt")
	}
	value := p.files[p.fileAt]
	p.fileAt++
	return value, nil
}

type fakeSetupService struct {
	state          core.StartupState
	google         bool
	loginIDs       []string
	loginNoOpen    []bool
	syncIDs        []string
	syncProjects   []core.Project
	quotaSetID     string
	quotaSetValue  string
	addedMethod    string
	addedID        string
	addedQuota     string
	inspectCalls   int
	loginErr       error
	syncErr        error
	setQuotaErr    error
	addIdentityErr error
}

func (s *fakeSetupService) GoogleAuthAvailable() bool { return s.google }

func (s *fakeSetupService) InspectStartupState() (core.StartupState, error) {
	s.inspectCalls++
	return s.state, nil
}

func (s *fakeSetupService) add(method, authID, quota string) (config.AuthEntry, error) {
	if s.addIdentityErr != nil {
		return config.AuthEntry{}, s.addIdentityErr
	}
	s.addedMethod, s.addedID, s.addedQuota = method, authID, quota
	entry := config.AuthEntry{ID: authID, Type: method, QuotaProjectID: quota}
	s.state.Auth = append(s.state.Auth, entry)
	if s.state.DefaultAuthID == "" {
		s.state.DefaultAuthID = authID
	}
	return entry, nil
}

func (s *fakeSetupService) AddGoogleAuthWithQuotaProject(authID, _ string, quota string) (config.AuthEntry, error) {
	return s.add(config.AuthTypeGoogle, authID, quota)
}

func (s *fakeSetupService) AddOAuthAuthWithQuotaProject(authID, _ string, _ []byte, quota string) (config.AuthEntry, error) {
	return s.add(config.AuthTypeOAuth, authID, quota)
}

func (s *fakeSetupService) AddServiceAccountAuthWithQuotaProject(authID, _ string, _ []byte, quota string) (config.AuthEntry, error) {
	return s.add(config.AuthTypeServiceAccount, authID, quota)
}

func (s *fakeSetupService) AddGCloudAuthWithQuotaProject(authID, _ string, quota string) (config.AuthEntry, error) {
	return s.add(config.AuthTypeGCloud, authID, quota)
}

func (s *fakeSetupService) SetAuthQuotaProject(authID, quota string) (config.AuthEntry, string, bool, error) {
	if s.setQuotaErr != nil {
		return config.AuthEntry{}, "", false, s.setQuotaErr
	}
	s.quotaSetID, s.quotaSetValue = authID, quota
	for index := range s.state.Auth {
		if s.state.Auth[index].ID == authID {
			previous := s.state.Auth[index].QuotaProjectID
			s.state.Auth[index].QuotaProjectID = quota
			return s.state.Auth[index], previous, previous != quota, nil
		}
	}
	return config.AuthEntry{}, "", false, errors.New("auth not found")
}

func (s *fakeSetupService) EnsureAuthLogin(_ context.Context, authID string, noOpen bool) error {
	s.loginIDs = append(s.loginIDs, authID)
	s.loginNoOpen = append(s.loginNoOpen, noOpen)
	return s.loginErr
}

func (s *fakeSetupService) SyncProjectsForAuth(_ context.Context, authID string) ([]core.Project, string, error) {
	s.syncIDs = append(s.syncIDs, authID)
	return append([]core.Project(nil), s.syncProjects...), "firebase", s.syncErr
}

func executeSetup(t *testing.T, svc setupService, prompts prompter, args ...string) (string, string, error) {
	t.Helper()
	cmd := newCommand(svc, prompts)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestCompleteProfileReturnsWithoutPromptingOrNetwork(t *testing.T) {
	svc := &fakeSetupService{state: core.StartupState{
		Profile:  "work",
		Auth:     []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud}},
		Projects: []core.Project{{ProjectID: "demo"}},
	}}
	out, _, err := executeSetup(t, svc, &fakePrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "setup already complete") || !strings.Contains(out, "1 identity") || !strings.Contains(out, "1 project") {
		t.Fatalf("output = %q", out)
	}
	if len(svc.loginIDs) != 0 || len(svc.syncIDs) != 0 {
		t.Fatalf("unexpected network calls: login=%v sync=%v", svc.loginIDs, svc.syncIDs)
	}
}

func TestEmptyProfileAddsGCloudAuthenticatesAndDiscovers(t *testing.T) {
	svc := &fakeSetupService{
		state:        core.StartupState{Profile: "default"},
		syncProjects: []core.Project{{ProjectID: "one", DiscoveredBy: []string{"default"}}, {ProjectID: "two", DiscoveredBy: []string{"default"}}},
	}
	prompts := &fakePrompter{
		selections: []string{config.AuthTypeGCloud},
		texts:      []string{"default", "billing-project"},
	}
	out, _, err := executeSetup(t, svc, prompts, "--noopen")
	if err != nil {
		t.Fatal(err)
	}
	if svc.addedMethod != config.AuthTypeGCloud || svc.addedID != "default" || svc.addedQuota != "billing-project" {
		t.Fatalf("added = method=%q id=%q quota=%q", svc.addedMethod, svc.addedID, svc.addedQuota)
	}
	if len(svc.loginIDs) != 1 || svc.loginIDs[0] != "default" || len(svc.loginNoOpen) != 1 || !svc.loginNoOpen[0] {
		t.Fatalf("login = ids=%v noopen=%v", svc.loginIDs, svc.loginNoOpen)
	}
	if len(svc.syncIDs) != 1 || svc.syncIDs[0] != "default" {
		t.Fatalf("sync IDs = %v", svc.syncIDs)
	}
	if !strings.Contains(out, "projects discovered: 2 projects") || !strings.Contains(out, "next: fbrcm projects list") {
		t.Fatalf("output = %q", out)
	}
}

func TestCredentialMethodsImportSelectedJSON(t *testing.T) {
	tests := []struct {
		name   string
		method string
		raw    string
	}{
		{
			name:   "OAuth",
			method: config.AuthTypeOAuth,
			raw:    `{"installed":{"client_id":"client","client_secret":"secret","auth_uri":"https://accounts.example/authorize","token_uri":"https://accounts.example/token","redirect_uris":["http://localhost"]}}`,
		},
		{
			name:   "service account",
			method: config.AuthTypeServiceAccount,
			raw:    `{"type":"service_account","project_id":"demo","private_key":"key","client_email":"service@example.com","token_uri":"https://accounts.example/token"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "credentials.json")
			if err := os.WriteFile(path, []byte(test.raw), 0o600); err != nil {
				t.Fatal(err)
			}
			svc := &fakeSetupService{
				state:        core.StartupState{Profile: "default"},
				syncProjects: []core.Project{{ProjectID: "one", DiscoveredBy: []string{"imported"}}},
			}
			prompts := &fakePrompter{
				selections: []string{test.method},
				texts:      []string{"imported", "billing-project"},
				files:      []string{path},
			}
			if _, _, err := executeSetup(t, svc, prompts); err != nil {
				t.Fatal(err)
			}
			if svc.addedMethod != test.method || svc.addedID != "imported" || svc.addedQuota != "billing-project" {
				t.Fatalf("added = method=%q id=%q quota=%q", svc.addedMethod, svc.addedID, svc.addedQuota)
			}
			if prompts.fileAt != 1 {
				t.Fatalf("file prompts = %d", prompts.fileAt)
			}
		})
	}
}

func TestInvalidCredentialJSONIsNotPersisted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.json")
	if err := os.WriteFile(path, []byte(`{"installed":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	svc := &fakeSetupService{state: core.StartupState{Profile: "default"}}
	prompts := &fakePrompter{
		selections: []string{config.AuthTypeOAuth},
		texts:      []string{"imported"},
		files:      []string{path},
	}
	_, _, err := executeSetup(t, svc, prompts)
	var validation *shared.ValidationError
	if !errors.As(err, &validation) || validation.Code != "auth.credentials_invalid" {
		t.Fatalf("error = %#v", err)
	}
	if len(svc.state.Auth) != 0 {
		t.Fatalf("persisted auth = %+v", svc.state.Auth)
	}
}

func TestExistingIdentityWithoutQuotaResumesSetup(t *testing.T) {
	svc := &fakeSetupService{
		state:        core.StartupState{Profile: "default", Auth: []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud}}, DefaultAuthID: "main"},
		syncProjects: []core.Project{{ProjectID: "one", DiscoveredBy: []string{"main"}}},
	}
	prompts := &fakePrompter{texts: []string{"billing-project"}}
	out, _, err := executeSetup(t, svc, prompts)
	if err != nil {
		t.Fatal(err)
	}
	if svc.quotaSetID != "main" || svc.quotaSetValue != "billing-project" {
		t.Fatalf("quota set = %q %q", svc.quotaSetID, svc.quotaSetValue)
	}
	if !strings.Contains(out, "setup complete") {
		t.Fatalf("output = %q", out)
	}
}

func TestNoProjectsCanFinishWithValidAuthentication(t *testing.T) {
	svc := &fakeSetupService{state: core.StartupState{
		Profile:       "default",
		Auth:          []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud, QuotaProjectID: "billing-project"}},
		DefaultAuthID: "main",
	}}
	prompts := &fakePrompter{selections: []string{choiceFinish}}
	out, _, err := executeSetup(t, svc, prompts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "authentication setup complete") || !strings.Contains(out, "projects discovered: 0 projects") {
		t.Fatalf("output = %q", out)
	}
}

func TestJSONReturnsInteractionBeforeInspectingState(t *testing.T) {
	shared.SetMachineMode(true)
	defer shared.SetMachineMode(false)
	svc := &fakeSetupService{}
	_, _, err := executeSetup(t, svc, &fakePrompter{})
	var interaction *shared.InteractionError
	if !errors.As(err, &interaction) || interaction.Type != "guided_setup" {
		t.Fatalf("error = %#v", err)
	}
	if svc.inspectCalls != 0 {
		t.Fatalf("inspect calls = %d, want 0", svc.inspectCalls)
	}
}

func TestDiscoveryCountsOnlyProjectsFoundBySelectedIdentity(t *testing.T) {
	svc := &fakeSetupService{
		state: core.StartupState{Profile: "default", Auth: []config.AuthEntry{{ID: "main", Type: config.AuthTypeGCloud, QuotaProjectID: "billing-project"}}},
		syncProjects: []core.Project{
			{ProjectID: "found", DiscoveredBy: []string{"main"}},
			{ProjectID: "retained", DiscoveredBy: []string{"other"}, Disabled: true},
		},
	}
	out, _, err := executeSetup(t, svc, &fakePrompter{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "projects discovered: 1 project") {
		t.Fatalf("output = %q", out)
	}
}

func TestLoginFailurePreservesAddedIdentityForResume(t *testing.T) {
	svc := &fakeSetupService{state: core.StartupState{Profile: "default"}, loginErr: errors.New("login failed")}
	prompts := &fakePrompter{selections: []string{config.AuthTypeGCloud}, texts: []string{"default", "billing-project"}}
	_, _, err := executeSetup(t, svc, prompts)
	if err == nil || !strings.Contains(err.Error(), "authenticate default") {
		t.Fatalf("error = %v", err)
	}
	if len(svc.state.Auth) != 1 || svc.state.Auth[0].ID != "default" {
		t.Fatalf("persisted auth = %+v", svc.state.Auth)
	}
}

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	if cmd.Use != "setup" {
		t.Fatalf("Use = %q", cmd.Use)
	}
	if cmd.Flags().Lookup("noopen") == nil {
		t.Fatal("missing --noopen")
	}
	if cmd.Args == nil {
		t.Fatal("missing argument validator")
	}
	if err := cmd.Args(&cobra.Command{}, []string{"extra"}); err == nil {
		t.Fatal("extra argument accepted")
	}
}
