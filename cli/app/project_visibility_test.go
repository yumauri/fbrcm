package app

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/ops/contract"
)

func visibilityCommand(t *testing.T, args ...string) contract.Envelope {
	t.Helper()
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	root := newRootCommandWithOfflineInit(svc, "test", "test", "test", func(context.Context, bool) {})
	args = append(args, "--json")
	envelope, raw := executeJSONContractRoot(t, root, args...)
	validateContractDocument(t, envelope.Schema, raw)
	return envelope
}

func visibilityData(t *testing.T, envelope contract.Envelope) map[string]any {
	t.Helper()
	if envelope.ExitCode != 0 {
		t.Fatalf("command failed: %#v", envelope)
	}
	return structToContractValue(t, envelope.Data).(map[string]any)
}

// Exercises every executable policy command through the public CLI envelope.
func TestProjectVisibilityCommandsConform(t *testing.T) {
	useTemporaryContractState(t)
	for _, name := range []string{"blacklist", "whitelist"} {
		for _, test := range []struct {
			args    []string
			changed bool
		}{
			{[]string{"add", "alpha", "beta", "alpha"}, true},
			{[]string{"add", "alpha"}, false},
			{[]string{"enable"}, name == "whitelist"},
			{[]string{"disable"}, true},
			{[]string{"disable"}, false},
			{[]string{"remove", "alpha", "missing"}, true},
			{[]string{"clear"}, true},
			{[]string{"clear"}, false},
		} {
			args := append([]string{"projects", name}, test.args...)
			data := visibilityData(t, visibilityCommand(t, args...))
			if data["changed"] != test.changed || data["list"] != name || (data["no_op_reason"] == nil) != test.changed {
				t.Fatalf("%v result = %#v", args, data)
			}
		}
		data := visibilityData(t, visibilityCommand(t, "projects", name, "list"))
		if data["count"] != float64(0) || data["enabled"] != false {
			t.Fatalf("empty list = %#v", data)
		}
	}
	for _, args := range [][]string{
		{"projects", "blacklist", "add"},
		{"projects", "blacklist", "remove", "server@alpha"},
		{"projects", "whitelist", "add", "alpha", "--filter", "beta"},
		{"projects", "whitelist", "add", "--filter", "server@alpha"},
		{"projects", "blacklist", "enable", "extra"},
		{"projects", "whitelist", "list", "--stateless"},
	} {
		envelope := visibilityCommand(t, args...)
		if envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" {
			t.Fatalf("invalid %v result = %#v", args, envelope)
		}
		if strings.Contains(strings.Join(args, " "), "--filter server@") && !strings.Contains(envelope.Errors[0].Message, "--filter") {
			t.Fatalf("filter error names the wrong option: %#v", envelope.Errors[0])
		}
	}
}

func TestProjectVisibilitySelectionAndPersistenceConform(t *testing.T) {
	useTemporaryContractState(t)
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	projects := []config.Project{{Name: "Alpha", ProjectID: "alpha", AuthID: "main", ProjectNumber: "100"}, {Name: "Beta", ProjectID: "beta", AuthID: "main"}}
	if err := config.SaveProjects(projects, time.Now()); err != nil {
		t.Fatal(err)
	}
	visibilityData(t, visibilityCommand(t, "projects", "blacklist", "add", "--filter", "=alpha"))
	data := visibilityData(t, visibilityCommand(t, "projects", "list"))
	if data["count"] != float64(1) {
		t.Fatalf("list = %#v", data)
	}
	for _, args := range [][]string{{"project", "show", "alpha"}, {"project", "templates", "show", "alpha"}, {"get", "--project", "server@=alpha"}} {
		envelope := visibilityCommand(t, args...)
		if envelope.ExitCode != 6 || envelope.Errors[0].Code != "project.hidden" || envelope.Errors[0].Retryable {
			t.Fatalf("hidden %v result = %#v", args, envelope)
		}
	}
	data = visibilityData(t, visibilityCommand(t, "projects", "list", "--include-hidden"))
	if data["count"] != float64(2) || data["items"].([]any)[0].(map[string]any)["hidden_reason"] != "blacklisted" {
		t.Fatalf("include hidden = %#v", data)
	}
	visibilityData(t, visibilityCommand(t, "projects", "whitelist", "add", "alpha"))
	envelope := visibilityCommand(t, "projects", "whitelist", "add", "alpha")
	if len(envelope.Warnings) != 1 || envelope.Warnings[0].Code != "project.remains_hidden" {
		t.Fatalf("overlap warning = %#v", envelope)
	}
	visibilityData(t, visibilityCommand(t, "projects", "whitelist", "enable"))
	data = visibilityData(t, visibilityCommand(t, "projects", "list"))
	if data["count"] != float64(0) {
		t.Fatalf("blacklist must win: %#v", data)
	}
	visibilityData(t, visibilityCommand(t, "projects", "blacklist", "remove", "alpha"))
	visibilityData(t, visibilityCommand(t, "projects", "whitelist", "clear"))
	data = visibilityData(t, visibilityCommand(t, "projects", "list"))
	if data["count"] != float64(0) {
		t.Fatalf("empty enabled whitelist = %#v", data)
	}
	before, err := os.ReadFile(config.GetProjectVisibilityFilePath())
	if err != nil {
		t.Fatal(err)
	}
	// Inventory administration still selects hidden projects.
	visibilityCommand(t, "projects", "forget", "--filter", "=alpha", "--yes")
	visibilityCommand(t, "projects", "reset", "--yes")
	after, err := os.ReadFile(config.GetProjectVisibilityFilePath())
	if err != nil || string(after) != string(before) {
		t.Fatalf("inventory administration changed policy: %q, %v", after, err)
	}
	data = visibilityData(t, visibilityCommand(t, "projects", "blacklist", "add", "--filter", "=missing"))
	if data["no_op_reason"] != "no_match" {
		t.Fatalf("empty inventory filter = %#v", data)
	}
}

func TestProjectVisibilityInvocationSchemas(t *testing.T) {
	for _, name := range []string{"blacklist", "whitelist"} {
		for _, action := range []string{"add", "remove"} {
			id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:projects." + name + "." + action + ":input"
			for _, test := range []struct {
				arguments string
				options   string
				valid     bool
			}{
				{`{"project_id":["alpha","future-project"]}`, `{}`, true},
				{`{}`, `{"filter":["^team-"]}`, true},
				{`{}`, `{}`, false},
				{`{"project_id":[]}`, `{}`, false},
				{`{"project_id":["alpha"]}`, `{"filter":["beta"]}`, false},
				{`{"project_id":["server@alpha"]}`, `{}`, false},
				{`{"project_id":["=alpha"]}`, `{}`, false},
				{`{}`, `{"filter":["server@alpha"]}`, false},
				{`{}`, `{"filter":[" "]}`, false},
			} {
				var value any
				if err := json.Unmarshal([]byte(`{"arguments":`+test.arguments+`,"options":`+test.options+`,"stdin":null}`), &value); err != nil {
					t.Fatal(err)
				}
				validateContractValue(t, id, value, test.valid)
			}
		}
	}
	for _, capability := range contract.DetailedCapabilities(NewRootForContract("test")) {
		if !contract.IsProjectVisibilityCommand(capability.ID) {
			continue
		}
		if capability.NetworkAccess != "none" || capability.Interaction.Mode != "none" || capability.Destructive || capability.Idempotency != "yes" || capability.Supports.Stateless || capability.Supports.Stdin {
			t.Fatalf("policy capabilities = %#v", capability)
		}
		if strings.HasSuffix(capability.ID, ".list") && capability.SideEffectLevel > 1 {
			t.Fatalf("listing has mutation effects: %#v", capability)
		}
	}
}

func TestProjectVisibilityDraftSelectionConform(t *testing.T) {
	useTemporaryContractState(t)
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	for _, id := range []string{"alpha", "server@alpha", "beta"} {
		raw := json.RawMessage(`{"parameters":{}}`)
		if err := config.SaveDraft(&config.Draft{FormatVersion: config.DraftFormatVersion, ProjectID: id, CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: raw, RemoteConfig: raw}); err != nil {
			t.Fatal(err)
		}
	}
	visibilityData(t, visibilityCommand(t, "projects", "blacklist", "add", "alpha"))
	data := visibilityData(t, visibilityCommand(t, "draft", "list"))
	if data["count"] != float64(1) {
		t.Fatalf("draft list = %#v", data)
	}
	for _, args := range [][]string{{"draft", "show", "server@alpha"}, {"draft", "list", "--filter", "server@=alpha"}} {
		envelope := visibilityCommand(t, args...)
		if envelope.ExitCode != 6 || envelope.Errors[0].Code != "project.hidden" {
			t.Fatalf("hidden draft %v = %#v", args, envelope)
		}
	}
	data = visibilityData(t, visibilityCommand(t, "draft", "list", "--include-hidden"))
	if data["count"] != float64(3) {
		t.Fatalf("override draft list = %#v", data)
	}
	visibilityData(t, visibilityCommand(t, "draft", "discard", "--all", "--yes"))
	ids, err := config.ListDraftProjectIDs()
	if err != nil || len(ids) != 2 {
		t.Fatalf("discard --all removed hidden drafts: %v, %v", ids, err)
	}
}

func TestProjectVisibilityUnknownEntriesAndFilterSnapshots(t *testing.T) {
	useTemporaryContractState(t)
	visibilityData(t, visibilityCommand(t, "projects", "blacklist", "add", "future-project"))
	data := visibilityData(t, visibilityCommand(t, "projects", "blacklist", "list"))
	if data["items"].([]any)[0].(map[string]any)["known"] != false {
		t.Fatalf("undiscovered entry = %#v", data)
	}
	visibilityData(t, visibilityCommand(t, "projects", "blacklist", "remove", "--filter", "=future-project"))
	policy, err := config.LoadProjectVisibility()
	if err != nil || len(policy.Blacklist.ProjectIDs) != 0 {
		t.Fatalf("unknown entry was not removed: %#v, %v", policy, err)
	}
	if err := config.SaveProjects([]config.Project{{ProjectID: "team-one", Name: "One", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	visibilityData(t, visibilityCommand(t, "projects", "whitelist", "add", "--filter", "^team-"))
	visibilityData(t, visibilityCommand(t, "projects", "whitelist", "enable"))
	if err := config.SaveProjects([]config.Project{{ProjectID: "team-one", AuthID: "main"}, {ProjectID: "team-two", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	data = visibilityData(t, visibilityCommand(t, "projects", "list"))
	if data["count"] != float64(1) || data["items"].([]any)[0].(map[string]any)["project_id"] != "team-one" {
		t.Fatalf("filters were persisted instead of resolved IDs: %#v", data)
	}
}

func TestProjectVisibilityResponseSchemaBoundaries(t *testing.T) {
	useTemporaryContractState(t)
	envelope := visibilityCommand(t, "projects", "blacklist", "add", "alpha")
	for _, mutate := range []func(map[string]any){
		func(data map[string]any) { data["operation"] = "remove" },
		func(data map[string]any) { data["count"] = -1 },
		func(data map[string]any) { data["no_op_reason"] = "already_applied" },
		func(data map[string]any) { data["changed"] = false },
		func(data map[string]any) { data["selection"].(map[string]any)["kind"] = "filters" },
		func(data map[string]any) { data["selection"].(map[string]any)["filters"] = []any{"=alpha"} },
		func(data map[string]any) { data["affected_project_ids"] = []any{"server@alpha"} },
	} {
		document := structToContractValue(t, envelope).(map[string]any)
		mutate(document["data"].(map[string]any))
		validateContractValue(t, envelope.Schema, document, false)
	}
}
