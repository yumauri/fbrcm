package draft

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
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
	for _, flag := range []string{"against", "cached", "filter", "search", "group", "expr", "parameters", "conditions", "json", "exit-code"} {
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

func TestDiffExitCodeAndCorruptDraftRecovery(t *testing.T) {
	setupCommandTest(t)
	base := commandRemoteConfig("1", "old")
	draftRaw := commandRemoteConfig("1", "new")
	now := time.Now().UTC()
	if err := config.SaveDraft(&config.Draft{FormatVersion: config.DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1", CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: base, RemoteConfig: draftRaw}); err != nil {
		t.Fatal(err)
	}
	cmd := New(nil)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"diff", "demo", "--exit-code"})
	err := cmd.Execute()
	var exitErr *shared.ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("diff exit error = %#v", err)
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
	diffOut := executeCommand(t, "diff", "demo")
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
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatalf("SwitchProfile returned error: %v", err)
	}
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
