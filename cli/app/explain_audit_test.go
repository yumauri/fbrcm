package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/schemas"
)

func TestContractAuditMetadataHasNoInapplicableBranches(t *testing.T) {
	for _, c := range contract.DetailedCapabilities(NewRootForContract("audit")) {
		conditions := append([]contract.BehaviorConditionClause{}, c.NetworkWhen...)
		conditions = append(conditions, c.InteractionWhen...)
		conditions = append(conditions, c.DestructiveWhen...)
		for _, effect := range c.SideEffectWhen {
			conditions = append(conditions, effect.When...)
		}
		for _, flag := range c.Flags {
			conditions = append(conditions, flag.EffectiveWhen...)
		}
		for _, retry := range c.IdempotencyWhen {
			conditions = append(conditions, retry.When...)
		}
		if !c.Supports.Stateless {
			for _, clause := range conditions {
				for _, p := range clause.AllOf {
					if p.Source == "option" && p.Name == "stateless" && p.Operator == "equals" && p.Value == true {
						t.Errorf("%s advertises an unreachable stateless branch", c.ID)
					}
				}
			}
		}
		raw, err := schemas.ReadByID(c.InvocationSchema)
		if err != nil {
			t.Fatal(err)
		}
		var document any
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatal(err)
		}
		var inspect func(any)
		inspect = func(value any) {
			switch node := value.(type) {
			case map[string]any:
				if _, hasItems := node["items"]; hasItems {
					switch typ := node["type"].(type) {
					case string:
						if typ != "array" {
							t.Errorf("%s hides item constraints on a %s", c.ID, typ)
						}
					case []any:
						if !slices.Contains(typ, any("array")) {
							t.Errorf("%s hides item constraints on %v", c.ID, typ)
						}
					}
				}
				for _, child := range node {
					inspect(child)
				}
			case []any:
				for _, child := range node {
					inspect(child)
				}
			}
		}
		inspect(document)
	}
}

func TestExplainAuditInputBoundaries(t *testing.T) {
	inputID := "urn:fbrcm:schema:cli:" + contract.Version + ":command:explain:input"
	for _, tc := range []struct {
		name, capture, assignment string
	}{
		{name: "quoted curl executable", capture: `'curl' '` + explainFetchURL + `' --data '{}'`},
		{name: "case insensitive HAR decoder fields", capture: `{"LOG":{"ENTRIES":[{"REQUEST":{"METHOD":"POST","URL":"` + explainFetchURL + `","POSTDATA":{"TEXT":"{}"}}}]}}`},
		{name: "Go timestamp decoder boundary", capture: `{"log":{"entries":[{"startedDateTime":"2026-11-04T1:00:00Z","request":{"method":"POST","url":"` + explainFetchURL + `","postData":{"text":"{}"}}}]}}`},
		{name: "unrelated HAR entries and nullable fields", capture: `{"log":{"entries":[null,{}, {"request":null}, {"request":{"url":"not a URI"},"response":{"status":999,"content":{"encoding":"unused"}}}, {"startedDateTime":null,"time":null,"request":{"method":"POST","url":"` + explainFetchURL + `","headers":[null,{}, {"name":null,"value":null}],"postData":{"text":"{}","encoding":null}}}]}}`},
		{name: "escaped indexed field", assignment: `app.customSignal['can\'t']=value`},
		{name: "indexed field whitespace", assignment: `  app.customSignal [ 'key' ] =value`},
		{name: "indexed field vertical tab", assignment: "\vapp.customSignal\v[\v'key'\v]\v=value"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := explainContractService(t, nil, nil, &firebase.RemoteConfigValue{Value: "false"})
			root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
			options := map[string]any{"project": "demo", "cached": true}
			var stdin any
			args := []string{"explain", "flag", "--project", "demo", "--cached", "--json"}
			if tc.capture != "" {
				stdin = tc.capture
				if strings.HasPrefix(tc.capture, "{") {
					if err := json.Unmarshal([]byte(tc.capture), &stdin); err != nil {
						t.Fatal(err)
					}
				}
				root.SetIn(shared.DocumentInput(strings.NewReader(tc.capture)))
			} else {
				root.SetIn(shared.NonInteractiveInput())
				options["set"] = []any{tc.assignment}
				args = append(args, "--set", tc.assignment)
			}
			envelope, raw := executeJSONContractRoot(t, root, args...)
			validateContractDocument(t, envelope.Schema, raw)
			if envelope.ExitCode != 0 {
				t.Fatalf("supported runtime input failed: %s", raw)
			}
			validateContractValue(t, inputID, map[string]any{"arguments": map[string]any{"parameter": "flag"}, "options": options, "stdin": stdin}, true)
		})
	}
	for _, query := range []string{"=", "server@=", "client@ "} {
		validateContractValue(t, inputID, map[string]any{"arguments": map[string]any{"parameter": "flag"}, "options": map[string]any{"project": query}, "stdin": nil}, false)
	}
	raw, err := schemas.ReadByID(inputID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"explain_project_resolution"`) || !strings.Contains(string(raw), `"explain_capture_source"`) {
		t.Fatal("explain must publish its project-resolution and capture-source algorithms")
	}
}

func TestExplainAuditProjectSelectionAndSourcePrecedence(t *testing.T) {
	for _, tc := range []struct {
		query, key, code string
		exit             int
	}{
		{query: "123456789", key: "flag"},
		{query: "demo", key: "flag"},
		{query: "Demo", key: "flag"},
		{query: "^demo", key: "flag"},
		{query: "missing", key: "flag", code: "project.not_found", exit: 6},
		{query: "999", key: "flag", code: "project.not_found", exit: 6},
		{query: "^d", key: "flag", code: "project.ambiguous", exit: 2},
		{query: "demo", key: "FLAG", code: "parameter.not_found", exit: 6},
	} {
		t.Run(tc.query+"/"+tc.key, func(t *testing.T) {
			svc := explainContractService(t, nil, nil, &firebase.RemoteConfigValue{Value: "false"})
			if err := config.SaveProjects([]config.Project{{ProjectID: "demo", ProjectNumber: "123456789", Name: "Demo", AuthID: "main"}, {ProjectID: "different", ProjectNumber: "987654321", Name: "Different", AuthID: "main"}}, time.Now()); err != nil {
				t.Fatal(err)
			}
			root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
			root.SetIn(shared.NonInteractiveInput())
			envelope, raw := executeJSONContractRoot(t, root, "explain", tc.key, "--project", tc.query, "--cached", "--json")
			validateContractDocument(t, envelope.Schema, raw)
			if envelope.ExitCode != tc.exit || tc.code != "" && (len(envelope.Errors) != 1 || envelope.Errors[0].Code != tc.code) {
				t.Fatalf("selection result: %s", raw)
			}
		})
	}
	t.Run("file takes precedence without consuming stdin", func(t *testing.T) {
		svc := explainContractService(t, nil, nil, &firebase.RemoteConfigValue{Value: "false"})
		path := filepath.Join(t.TempDir(), "capture.txt")
		if err := os.WriteFile(path, []byte(`curl '`+explainFetchURL+`' --data '{}'`), 0o600); err != nil {
			t.Fatal(err)
		}
		root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
		reader := strings.NewReader("unused input")
		root.SetIn(shared.DocumentInput(reader))
		envelope, raw := executeJSONContractRoot(t, root, "explain", "flag", "--from", path, "--cached", "--json")
		validateContractDocument(t, envelope.Schema, raw)
		if envelope.ExitCode != 0 || reader.Len() != len("unused input") {
			t.Fatalf("file/stdin precedence: %s", raw)
		}
	})
	t.Run("capture mismatch", func(t *testing.T) {
		svc := explainContractService(t, nil, nil, &firebase.RemoteConfigValue{Value: "false"})
		root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
		root.SetIn(shared.DocumentInput(strings.NewReader(`curl '` + strings.Replace(explainFetchURL, "123456789", "999", 1) + `' --data '{}'`)))
		envelope, raw := executeJSONContractRoot(t, root, "explain", "flag", "--project", "demo", "--cached", "--json")
		validateContractDocument(t, envelope.Schema, raw)
		if envelope.ExitCode != 2 || envelope.Errors[0].Code != "argument.invalid" {
			t.Fatalf("capture/explicit project mismatch: %s", raw)
		}
	})
}

func TestExplainAuditCaptureFailuresAndEntryBounds(t *testing.T) {
	for _, tc := range []struct {
		name, input, entry, from, code string
		exit                           int
	}{
		{name: "malformed JSON", input: "{", code: "capture.invalid", exit: 8},
		{name: "missing HAR log", input: "{}", code: "capture.invalid", exit: 8},
		{name: "null capture", input: "null", code: "capture.invalid", exit: 8},
		{name: "array capture", input: "[]", code: "capture.invalid", exit: 8},
		{name: "no Firebase candidates", input: `{"log":{"entries":[{}]}}`, code: "capture.invalid", exit: 8},
		{name: "unsupported selected body encoding", input: `{"log":{"entries":[{"request":{"method":"POST","url":"` + explainFetchURL + `","postData":{"text":"{}","encoding":"unsupported"}}}]}}`, code: "capture.invalid", exit: 8},
		{name: "candidate timestamp cannot be serialized", input: `{"log":{"entries":[{"startedDateTime":"2026-11-04T00:00:00+24:00","request":{"method":"POST","url":"` + explainFetchURL + `","postData":{"text":"{}"}}},{"request":{"method":"POST","url":"` + explainFetchURL + `","postData":{"text":"{}"}}}]}}`, code: "capture.invalid", exit: 8},
		{name: "timestamp bound exceeds the RFC3339 year range", input: `{"log":{"entries":[{"startedDateTime":"9999-12-31T23:59:59Z","time":1000,"request":{"method":"POST","url":"` + explainFetchURL + `","postData":{"text":"{}"}}}]}}`, code: "capture.invalid", exit: 8},
		{name: "explicit stdin without a redirected stream", from: "-", code: "capture.invalid", exit: 8},
		{name: "explicit stdin with an empty document", from: "-", input: " ", code: "capture.invalid", exit: 8},
		{name: "entry without capture", entry: "1", code: "argument.invalid", exit: 2},
		{name: "largest entry is a lookup failure", input: `curl '` + explainFetchURL + `' --data '{}'`, entry: "9223372036854775807", code: "argument.invalid", exit: 2},
		{name: "entry overflow", entry: "9223372036854775808", code: "argument.invalid", exit: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := explainContractService(t, nil, nil, &firebase.RemoteConfigValue{Value: "false"})
			root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
			root.SetIn(shared.NonInteractiveInput())
			if tc.input != "" {
				root.SetIn(shared.DocumentInput(strings.NewReader(tc.input)))
			}
			args := []string{"explain", "flag", "--project", "demo", "--cached", "--json"}
			if tc.entry != "" {
				args = append(args, "--entry", tc.entry)
			}
			if tc.from != "" {
				args = append(args, "--from", tc.from)
			}
			envelope, raw := executeJSONContractRoot(t, root, args...)
			validateContractDocument(t, envelope.Schema, raw)
			if envelope.ExitCode != tc.exit || len(envelope.Errors) != 1 || envelope.Errors[0].Code != tc.code {
				t.Fatalf("capture boundary: %s", raw)
			}
		})
	}
}
