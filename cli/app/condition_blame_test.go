package app

import (
	"net/http"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/ops/contract"
)

func TestConditionBlameMachineEnvelopes(t *testing.T) {
	for _, test := range []struct {
		name     string
		selector string
		extra    []string
		present  bool
		status   int
		code     string
		changes  int
	}{
		{name: "creation", selector: "Beta Users", status: 0, changes: 1},
		{name: "retained boundary", selector: "Beta Users", present: true, status: 0},
		{name: "missing exact name", selector: "beta users", status: 6, code: "condition.not_found"},
		{name: "untrimmed name", selector: " Beta Users", status: 6, code: "condition.not_found"},
		{name: "blank name", selector: " ", status: 2, code: "argument.invalid"},
		{name: "conflicting parameter", selector: "Beta Users", extra: []string{"flag"}, status: 2, code: "argument.invalid"},
		{name: "conflicting limits", selector: "Beta Users", extra: []string{"--limit", "2"}, status: 2, code: "argument.invalid"},
		{name: "invalid version", selector: "Beta Users", extra: []string{"--at", "CURRENT"}, status: 2, code: "argument.invalid"},
		{name: "unavailable version", selector: "Beta Users", extra: []string{"--at", "99"}, status: 6, code: "version.not_found"},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			cmd, captured, configRoot, cacheRoot := newStatelessHTTPTestCommand(t, func(req *http.Request) (*http.Response, error) {
				calls++
				if !strings.Contains(req.URL.Path, "/projects/demo/namespaces/firebase-server/remoteConfig") {
					t.Fatalf("unexpected request: %s", req.URL)
				}
				if strings.Contains(req.URL.Path, "listVersions") {
					return statelessHTTPResponse(req, http.StatusOK, `{"versions":[{"versionNumber":"3","updateTime":"2026-11-06T09:00:00Z","updateUser":{"email":"alice@example.com"},"description":"Add beta"},{"versionNumber":"1"}]}`, ""), nil
				}
				version := req.URL.Query().Get("versionNumber")
				conditions := `[]`
				if version == "3" || test.present {
					conditions = `[{"name":"Beta Users","expression":"true","tagColor":"BLUE"}]`
				}
				return statelessHTTPResponse(req, http.StatusOK, `{"version":{"versionNumber":"`+version+`"},"conditions":`+conditions+`}`, `"etag-`+version+`"`), nil
			})
			args := []string{"--stateless", "versions", "blame", "server@demo", "--condition", test.selector, "--all", "--json"}
			cmd.SetArgs(append(args, test.extra...))
			executed, err := cmd.ExecuteC()
			envelope := contract.BuildEnvelope(executed, "test", captured.Bytes(), err)
			if envelope.ExitCode != test.status {
				t.Fatalf("envelope = %#v, execution error = %v, output = %s", envelope, err, captured.String())
			}
			value := structToContractValue(t, envelope)
			validateContractValue(t, contract.SchemaID("versions.blame"), value, true)
			if test.code != "" {
				if len(envelope.Errors) != 1 || envelope.Errors[0].Code != test.code || envelope.Outcome != "failure" {
					t.Fatalf("failure = %#v", envelope)
				}
			} else {
				data := value.(map[string]any)["data"].(map[string]any)
				if data["condition"] != "Beta Users" || len(data["changes"].([]any)) != test.changes || !data["history_exhausted"].(bool) {
					t.Fatalf("data = %#v", data)
				}
				if _, present := data["parameter"]; present {
					t.Fatal("condition result contains parameter identity")
				}
				if test.changes > 0 {
					change := data["changes"].([]any)[0].(map[string]any)
					if change["update_time"] != "2026-11-06T09:00:00Z" || change["update_user"].(map[string]any)["email"] != "alice@example.com" || change["change"].(map[string]any)["kind"] != "added" {
						t.Fatalf("creation = %#v", change)
					}
				}
				if calls != 3 {
					t.Fatalf("requests = %d, want 3", calls)
				}
			}
			assertProfilePathsAbsent(t, configRoot, cacheRoot)
		})
	}
}

func TestConditionBlameInvocationSchema(t *testing.T) {
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:versions.blame:input"
	for _, test := range []struct {
		name      string
		arguments map[string]any
		options   map[string]any
		valid     bool
	}{
		{"condition", map[string]any{"project": "demo"}, map[string]any{"condition": "Beta Users", "all": true}, true},
		{"parameter", map[string]any{"project": "demo", "parameter": "flag"}, map[string]any{}, true},
		{"both", map[string]any{"project": "demo", "parameter": "flag"}, map[string]any{"condition": "Beta Users"}, false},
		{"neither", map[string]any{"project": "demo"}, map[string]any{}, false},
		{"empty", map[string]any{"project": "demo"}, map[string]any{"condition": ""}, false},
		{"blank", map[string]any{"project": "demo"}, map[string]any{"condition": " \t\n"}, false},
		{"at", map[string]any{"project": "demo"}, map[string]any{"condition": "Beta Users", "at": "previous", "limit": 2}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			validateContractValue(t, id, map[string]any{"arguments": test.arguments, "options": test.options, "stdin": nil}, test.valid)
		})
	}
}

func TestConditionBlameResponseSchemaRejectsUnreachableStates(t *testing.T) {
	id := contract.SchemaID("versions.blame")
	base := func() map[string]any {
		return map[string]any{
			"project":   map[string]any{"name": "Demo", "project_id": "demo", "auth_id": "main"},
			"condition": "Beta Users", "at_version": "5", "scanned_version_count": 1,
			"history_exhausted": true, "boundary": map[string]any{"version": "5", "state": "present"}, "changes": []any{},
		}
	}
	validateContractValue(t, id, artifactEnvelopeValue("versions.blame", base()), true)
	for _, test := range []struct {
		name string
		edit func(map[string]any)
	}{
		{"nil changes", func(data map[string]any) { data["changes"] = nil }},
		{"zero scanned", func(data map[string]any) { data["scanned_version_count"] = 0 }},
		{"missing boundary", func(data map[string]any) { data["boundary"] = nil }},
		{"premature boundary", func(data map[string]any) { data["history_exhausted"] = false }},
		{"parameter identity", func(data map[string]any) { data["parameter"] = "flag" }},
		{"parameter boundary", func(data map[string]any) { data["boundary"].(map[string]any)["group"] = "" }},
		{"unchanged entry", func(data map[string]any) {
			data["changes"] = []any{map[string]any{
				"previous_version": "4", "version": "5", "update_time": "2026-11-06T09:00:00Z",
				"update_user": map[string]any{}, "change_note": nil, "update_origin": "CONSOLE",
				"update_type": "INCREMENTAL_UPDATE", "rollback_source": nil,
				"change": map[string]any{"name": "Beta Users", "kind": "unchanged"},
			}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := base()
			test.edit(data)
			validateContractValue(t, id, artifactEnvelopeValue("versions.blame", data), false)
		})
	}
}
