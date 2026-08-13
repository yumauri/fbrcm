package draft

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
)

func TestNewCommandStructure(t *testing.T) {
	cmd := New(nil)
	cmdtest.AssertSubcommands(t, cmd, "list", "path", "show", "change-note", "diff", "publish", "discard")
	cmdtest.AssertFlag(t, cmd, "path", "json")
	for _, flag := range []string{"filter", "json"} {
		cmdtest.AssertFlag(t, cmd, "list", flag)
	}
	cmdtest.AssertFlag(t, cmd, "publish", "change-note")
	cmdtest.AssertFlag(t, cmd, "change-note", "clear")
	for _, flag := range []string{"raw", "to"} {
		cmdtest.AssertFlag(t, cmd, "show", flag)
	}
	for _, flag := range []string{"against", "cached", "filter", "search", "group", "expr", "parameters", "conditions", "json"} {
		cmdtest.AssertFlag(t, cmd, "diff", flag)
	}
	for _, subcommand := range []string{"publish", "discard"} {
		for _, flag := range []string{"all", "yes", "json"} {
			cmdtest.AssertFlag(t, cmd, subcommand, flag)
		}
	}
	cmdtest.AssertFlag(t, cmd, "publish", "dry-run")
}

func TestPathCommand(t *testing.T) {
	setupCommandTest(t)

	if got, want := strings.TrimSpace(executeCommand(t, "path")), config.GetDraftsDirPath(); got != want {
		t.Fatalf("draft path = %q, want %q", got, want)
	}

	jsonOut := executeCommand(t, "path", "--json")
	if !strings.Contains(jsonOut, `"path": "`+config.GetDraftsDirPath()+`"`) {
		t.Fatalf("draft path json = %q, want drafts directory", jsonOut)
	}
}

func TestBatchJSONOutputsAreArrays(t *testing.T) {
	setupCommandTest(t)

	for _, args := range [][]string{
		{"publish", "--all", "--yes", "--json"},
		{"discard", "--all", "--yes", "--json"},
	} {
		if got := strings.TrimSpace(executeCommand(t, args...)); got != "[]" {
			t.Fatalf("%v output = %q, want []", args, got)
		}
	}
}

func TestSuccessfulPublishStatusDistinguishesNoOpDryRuns(t *testing.T) {
	for _, test := range []struct {
		name        string
		dryRun      bool
		changed     bool
		wantStatus  string
		wantDeleted bool
	}{
		{name: "dry run unchanged", dryRun: true, wantStatus: "unchanged"},
		{name: "live unchanged", wantStatus: "already-applied", wantDeleted: true},
		{name: "dry run changed", dryRun: true, changed: true, wantStatus: "would-publish"},
		{name: "live changed", changed: true, wantStatus: "published", wantDeleted: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			status, deleted := successfulPublishStatus(test.dryRun, test.changed)
			if status != test.wantStatus || deleted != test.wantDeleted {
				t.Fatalf("successfulPublishStatus(%t, %t) = %q, %t; want %q, %t", test.dryRun, test.changed, status, deleted, test.wantStatus, test.wantDeleted)
			}
		})
	}
}

func TestPostPublicationFailureUsesFirebaseValidationProvenance(t *testing.T) {
	result := publishResult{ValidationSource: core.ValidationSourceLocal}

	markDraftFirebaseValidated(&result)

	if !result.Validated {
		t.Fatal("post-publication result is not marked as validated")
	}
	if result.ValidationSource != core.ValidationSourceFirebase {
		t.Fatalf("validation source = %q, want %q", result.ValidationSource, core.ValidationSourceFirebase)
	}
}

func TestDiffReturnsStatusOneInHumanAndJSONModes(t *testing.T) {
	setupCommandTest(t)
	base := commandRemoteConfig("1", "old")
	draftRaw := commandRemoteConfig("1", "new")
	now := time.Now().UTC()
	if err := config.SaveDraft(&config.Draft{FormatVersion: config.DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1", CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: base, RemoteConfig: draftRaw}); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"diff", "demo"}, {"diff", "demo", "--json"}} {
		cmd := New(nil)
		cmd.SetOut(&bytes.Buffer{})
		var stderr bytes.Buffer
		cmd.SetErr(&stderr)
		cmd.SetArgs(args)
		err := cmd.Execute()
		var exitErr *shared.ExitError
		if !errors.As(err, &exitErr) || exitErr.Code != 1 {
			t.Fatalf("%v diff exit error = %#v", args, err)
		}
		if stderr.Len() != 0 {
			t.Fatalf("%v semantic diff result wrote Cobra diagnostics: %q", args, stderr.String())
		}
	}
}

func TestCorruptDraftRecovery(t *testing.T) {
	setupCommandTest(t)
	raw := commandRemoteConfig("1", "new")
	now := time.Now().UTC()
	if err := config.SaveDraft(&config.Draft{FormatVersion: config.DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1", CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: raw, RemoteConfig: raw}); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(config.GetDraftPath("demo"), []byte(`{"version":{"versionNumber":"1"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	rawOut := executeCommand(t, "show", "demo", "--raw")
	if !strings.Contains(rawOut, `"versionNumber":"1"`) {
		t.Fatalf("raw corrupt draft output = %q", rawOut)
	}
	discardOut := executeCommand(t, "discard", "demo", "--yes", "--json")
	if !strings.Contains(discardOut, `"status": "discarded"`) {
		t.Fatalf("corrupt draft discard output = %q", discardOut)
	}
}

func TestListShowDiffAndDiscardLocalDraft(t *testing.T) {
	setupCommandTest(t)
	base := commandRemoteConfig("1", "old")
	draftRaw := commandRemoteConfig("1", "new")
	now := time.Now().UTC()
	if err := config.SaveDraft(&config.Draft{FormatVersion: config.DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1", CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: base, RemoteConfig: draftRaw}); err != nil {
		t.Fatalf("SaveDraft returned error: %v", err)
	}

	listOut := executeCommand(t, "list", "--json")
	if !strings.Contains(listOut, `"project_id": "demo"`) || !strings.Contains(listOut, `"status": "ready"`) {
		t.Fatalf("draft list output = %s", listOut)
	}
	noteOut := executeCommand(t, "change-note", "demo", "Enable checkout v2", "--json")
	if !strings.Contains(noteOut, `"change_note": "Enable checkout v2"`) {
		t.Fatalf("draft change-note output = %s", noteOut)
	}
	stored, err := config.LoadDraft("demo")
	if err != nil || stored.ChangeNote != "Enable checkout v2" || stored.FormatVersion != 1 {
		t.Fatalf("stored change note = %q, format = %d, err = %v", stored.ChangeNote, stored.FormatVersion, err)
	}
	showOut := executeCommand(t, "show", "demo")
	if !strings.Contains(showOut, `"value":"new"`) || strings.Contains(showOut, "base_remote_config") {
		t.Fatalf("draft show output = %s", showOut)
	}
	cmd := New(nil)
	var diffBuffer bytes.Buffer
	cmd.SetOut(&diffBuffer)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"diff", "demo"})
	err = cmd.Execute()
	var exitErr *shared.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("draft diff exit error = %#v", err)
	}
	diffOut := diffBuffer.String()
	if !strings.Contains(diffOut, "old") || !strings.Contains(diffOut, "new") {
		t.Fatalf("draft diff output = %s", diffOut)
	}
	discardOut := executeCommand(t, "discard", "demo", "--yes", "--json")
	if !strings.Contains(discardOut, `"status": "discarded"`) {
		t.Fatalf("draft discard output = %s", discardOut)
	}
	if _, err := config.LoadDraft("demo"); err == nil {
		t.Fatal("draft still exists after discard")
	}
}

func TestDraftShowDoesNotOverwriteExistingDestinationInMachineMode(t *testing.T) {
	setupCommandTest(t)
	now := time.Now().UTC()
	raw := commandRemoteConfig("1", "new")
	if err := config.SaveDraft(&config.Draft{
		FormatVersion: config.DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1",
		CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: raw, RemoteConfig: raw,
	}); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "draft.json")
	if err := os.WriteFile(destination, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := New(&core.Core{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"show", "demo", "--to", destination})
	shared.SetMachineMode(true)
	t.Cleanup(func() { shared.SetMachineMode(false) })
	err := cmd.Execute()
	var interaction *shared.InteractionError
	if !errors.As(err, &interaction) || interaction.RequiredOption != "" || !interaction.Destructive {
		t.Fatalf("draft show overwrite error = %#v", err)
	}
	got, readErr := os.ReadFile(destination)
	if readErr != nil || string(got) != "original" {
		t.Fatalf("destination = %q, %v", got, readErr)
	}
}

func TestDraftShowReportsNewDestinationAsNotOverwritten(t *testing.T) {
	setupCommandTest(t)
	now := time.Now().UTC()
	raw := commandRemoteConfig("1", "new")
	if err := config.SaveDraft(&config.Draft{
		FormatVersion: config.DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1",
		CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: raw, RemoteConfig: raw,
	}); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "draft.json")
	cmd := New(&core.Core{})
	cmd.PersistentFlags().Bool("json", false, "")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"show", "demo", "--to", destination, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var artifact contract.ArtifactData
	if err := json.Unmarshal(out.Bytes(), &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.Overwritten || artifact.Destination == nil || *artifact.Destination != destination {
		t.Fatalf("artifact = %#v", artifact)
	}
}

func TestResolveDraftUsesRepositoryAliases(t *testing.T) {
	setupCommandTest(t)
	if err := os.WriteFile(filepath.Join(mustGetwd(t), config.LocalConfigFileName), []byte(`[projects.aliases]
prod = "acme-production-42"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	project := config.Project{
		Name: "Production", ProjectID: "acme-production-42", AuthID: "main",
		Templates: []rctarget.Kind{rctarget.Client, rctarget.Server}, PrimaryTemplate: rctarget.Server,
	}
	if err := config.SaveProjects([]config.Project{project}, time.Now()); err != nil {
		t.Fatal(err)
	}
	saveResolutionDraft(t, "acme-production-42")
	saveResolutionDraft(t, "server@acme-production-42")

	for query, want := range map[string]string{
		"prod": "server@acme-production-42", "client@prod": "acme-production-42", "server@prod": "server@acme-production-42",
	} {
		got, _, err := resolveDraft(query)
		if err != nil || got != want {
			t.Fatalf("resolveDraft(%q) = %q, %v; want %q", query, got, err, want)
		}
	}
	for _, query := range []string{"PROD", "production", " prod"} {
		if _, _, err := resolveDraft(query); err == nil {
			t.Fatalf("case-mismatched or padded draft selector %q unexpectedly resolved", query)
		}
	}

	if err := config.SaveProjects(nil, time.Now()); err != nil {
		t.Fatal(err)
	}
	got, _, err := resolveDraft("prod")
	if err != nil || got != "acme-production-42" {
		t.Fatalf("unregistered alias draft = %q, %v", got, err)
	}
}

func TestResolveDraftExactProjectIDWinsOverAlias(t *testing.T) {
	setupCommandTest(t)
	if err := os.WriteFile(filepath.Join(mustGetwd(t), config.LocalConfigFileName), []byte(`[projects.aliases]
prod = "acme-production-42"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	saveResolutionDraft(t, "prod")
	saveResolutionDraft(t, "acme-production-42")
	got, _, err := resolveDraft("prod")
	if err != nil || got != "prod" {
		t.Fatalf("exact draft precedence = %q, %v", got, err)
	}
}

func saveResolutionDraft(t *testing.T, projectID string) {
	t.Helper()
	now := time.Now().UTC()
	raw := commandRemoteConfig("1", "value")
	if err := config.SaveDraft(&config.Draft{
		FormatVersion: config.DraftFormatVersion, ProjectID: projectID, BaseVersion: "1", BaseETag: "etag-1",
		CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: raw, RemoteConfig: raw,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDraftCommandsResolveClientAliasAndServerTargetSeparately(t *testing.T) {
	setupCommandTest(t)
	now := time.Now().UTC()
	for targetID, value := range map[string]string{"demo": "client", "server@demo": "server"} {
		raw := commandRemoteConfig("1", value)
		if err := config.SaveDraft(&config.Draft{
			FormatVersion:    config.DraftFormatVersion,
			ProjectID:        targetID,
			BaseVersion:      "1",
			BaseETag:         "etag-1",
			CreatedAt:        now,
			UpdatedAt:        now,
			BaseRemoteConfig: raw,
			RemoteConfig:     raw,
		}); err != nil {
			t.Fatal(err)
		}
	}

	if out := executeCommand(t, "show", "client@demo"); !strings.Contains(out, `"value":"client"`) {
		t.Fatalf("explicit client draft = %q", out)
	}
	if out := executeCommand(t, "show", "server@demo"); !strings.Contains(out, `"value":"server"`) {
		t.Fatalf("server draft = %q", out)
	}
	clientList := executeCommand(t, "list", "--filter", "client@=demo", "--json")
	if !strings.Contains(clientList, `"project_id": "demo"`) || strings.Contains(clientList, `"server@demo"`) {
		t.Fatalf("client draft list = %q", clientList)
	}
	serverList := executeCommand(t, "list", "--filter", "server@=demo", "--json")
	if !strings.Contains(serverList, `"project_id": "server@demo"`) || strings.Contains(serverList, `"project_id": "demo"`) {
		t.Fatalf("server draft list = %q", serverList)
	}
}

func TestDraftCommandsUseConfiguredTemplatePreferences(t *testing.T) {
	setupCommandTest(t)
	if err := config.SaveProjects([]config.Project{{
		Name:            "Demo",
		ProjectID:       "demo",
		AuthID:          "main",
		Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
		PrimaryTemplate: rctarget.Server,
	}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for targetID, value := range map[string]string{"demo": "client", "server@demo": "server"} {
		raw := commandRemoteConfig("1", value)
		if err := config.SaveDraft(&config.Draft{
			FormatVersion:    config.DraftFormatVersion,
			ProjectID:        targetID,
			BaseVersion:      "1",
			BaseETag:         "etag-1",
			CreatedAt:        now,
			UpdatedAt:        now,
			BaseRemoteConfig: raw,
			RemoteConfig:     raw,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if out := executeCommand(t, "show", "demo"); !strings.Contains(out, `"value":"server"`) {
		t.Fatalf("unqualified primary draft = %q", out)
	}
	out := executeCommand(t, "list", "--filter", "=demo", "--json")
	if !strings.Contains(out, `"project_id": "demo"`) || !strings.Contains(out, `"project_id": "server@demo"`) {
		t.Fatalf("unqualified configured draft list = %q", out)
	}
}

func TestRenderListTablePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	updatedAt := time.Date(2026, 7, 15, 9, 10, 0, 0, time.Local)

	output := renderListAtWidth([]listItem{{
		ProjectID:   "project-a",
		Project:     "Project A",
		BaseVersion: "42",
		UpdatedAt:   &updatedAt,
		Status:      "ready",
		Changes:     map[string]int{"parameters": 3, "conditions": 1},
	}}, 200)

	for _, want := range []string{"┌", "│", "Project ID", "Project", "Base", "Updated", "Changes", "Status", "project-a", "Project A", "42", "3 params, 1 conditions", "ready"} {
		if !strings.Contains(output, want) {
			t.Fatalf("renderList = %q, want substring %q", output, want)
		}
	}
}

func TestRenderListEmptyTablePlainText(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	output := renderList(nil)

	for _, want := range []string{"┌", "Project ID", "Project", "Base", "Updated", "Changes", "Status"} {
		if !strings.Contains(output, want) {
			t.Fatalf("empty renderList = %q, want substring %q", output, want)
		}
	}
	if strings.Contains(output, "No drafts") {
		t.Fatalf("empty renderList uses special empty-state message: %q", output)
	}
}

func TestRenderListCropsChangeNoteOnNarrowTerminal(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	note := "Enable checkout version two for the production environment"
	output := renderListAtWidth([]listItem{{
		ProjectID:  "production-project",
		Project:    "Production Project",
		Status:     "ready",
		Changes:    map[string]int{"parameters": 1, "conditions": 0},
		ChangeNote: &note,
	}}, 80)
	for index, line := range strings.Split(output, "\n") {
		if got := lipgloss.Width(line); got > 80 {
			t.Fatalf("line %d width = %d, want <= 80:\n%s", index, got, output)
		}
	}
	if !strings.Contains(output, "…") {
		t.Fatalf("narrow table did not ellipsize flexible content:\n%s", output)
	}
}

func setupCommandTest(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, root+"/config")
	t.Setenv(env.CacheDir, root+"/cache")
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func executeCommand(t *testing.T, args ...string) string {
	t.Helper()
	cmd := New(&core.Core{})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v\n%s", args, err, out.String())
	}
	return out.String()
}

func commandRemoteConfig(version, value string) json.RawMessage {
	v := firebase.RemoteConfigValue{Value: value}
	raw, err := json.Marshal(firebase.RemoteConfig{Version: firebase.RemoteConfigVersion{VersionNumber: version}, Parameters: map[string]firebase.RemoteConfigParam{"flag": {DefaultValue: &v, ValueType: "STRING"}}})
	if err != nil {
		panic(err)
	}
	return raw
}
