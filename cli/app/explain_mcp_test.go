package app

import (
	"encoding/json"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/contract"
)

func TestMCPExplainContract(t *testing.T) {
	svc := explainContractService(t, []string{`device.os == 'web'`, `device.os == 'ios'`}, nil, &firebase.RemoteConfigValue{Value: "false"})
	client := connectMCPStdio(t, svc, "--profile", "default", "--no-local-config", "--toolsets", "inspect", "--browser-auth", "never")
	listed, err := client.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var tool *mcp.Tool
	for _, entry := range listed.Tools {
		if entry.Name == "parameters.explain" {
			tool = entry
		}
	}
	if tool == nil {
		t.Fatal("parameters.explain missing from stateful inspect discovery without write permission")
	}
	if tool.Annotations == nil || tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint {
		t.Fatal("explain must advertise a non-destructive operation")
	}
	compiler := jsonschema.NewCompiler()
	const schemaID = "https://fbrcm.invalid/test/mcp-explain-output"
	if err := compiler.AddResource(schemaID, structToContractValue(t, tool.OutputSchema)); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(schemaID)
	if err != nil {
		t.Fatal(err)
	}
	har := explainHAR("true", "142")
	var multi struct {
		Log struct {
			Entries []json.RawMessage `json:"entries"`
		} `json:"log"`
	}
	if err := json.Unmarshal([]byte(har), &multi); err != nil {
		t.Fatal(err)
	}
	multi.Log.Entries = append(multi.Log.Entries, multi.Log.Entries[0])
	multiple, err := json.Marshal(multi)
	if err != nil {
		t.Fatal(err)
	}
	curl := `curl '` + explainFetchURL + `' -H 'Authorization: Bearer secret-auth' --data '{"app_id":"1:123456789:web:abc"}'`
	for _, test := range []struct {
		name      string
		parameter string
		stdin     any
		options   map[string]any
		state     string
		format    string
		code      string
		exit      int
	}{
		{name: "curl selects condition", stdin: curl, state: "resolved", format: "curl"},
		{name: "HAR preserves unknown outcomes", stdin: json.RawMessage(har), state: "unresolved", format: "har"},
		{name: "overrides resolve missing inputs", options: map[string]any{"set": []string{"device.os=web"}}, state: "resolved"},
		{name: "no capture remains noninteractive", state: "unresolved"},
		{name: "ambiguous HAR requires explicit entry", stdin: json.RawMessage(multiple), code: "interaction.required", exit: 10},
		{name: "selected HAR entry", stdin: json.RawMessage(multiple), options: map[string]any{"entry": 2}, state: "unresolved", format: "har"},
		{name: "malformed capture", stdin: "curl", code: "capture.invalid", exit: 8},
		{name: "missing parameter", parameter: "missing", code: "parameter.not_found", exit: 6},
		{name: "invalid schema input", options: map[string]any{"entry": -1}, code: "argument.invalid", exit: 2},
		{name: "success after errors", stdin: curl, state: "resolved", format: "curl"},
	} {
		t.Run(test.name, func(t *testing.T) {
			parameter := test.parameter
			if parameter == "" {
				parameter = "flag"
			}
			options := map[string]any{"project": "demo", "cached": true}
			maps.Copy(options, test.options)
			input, err := json.Marshal(map[string]any{"arguments": map[string]any{"parameter": parameter}, "options": options, "stdin": test.stdin})
			if err != nil {
				t.Fatal(err)
			}
			res, err := client.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.explain", Arguments: json.RawMessage(input)})
			if err != nil || res.NeedsInput() || res.IsError != (test.exit != 0) {
				t.Fatalf("unexpected MCP result: %v %v", res, err)
			}
			if len(res.Content) != 1 {
				t.Fatalf("expected one JSON content block: %v", res.Content)
			}
			raw := []byte(res.Content[0].(*mcp.TextContent).Text)
			var document any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(document, res.StructuredContent) {
				t.Fatal("text and structured result differ")
			}
			if err := schema.Validate(document); err != nil {
				t.Fatalf("result violates advertised MCP schema: %v\n%s", err, raw)
			}
			var envelope contract.Envelope
			if err := json.Unmarshal(raw, &envelope); err != nil {
				t.Fatal(err)
			}
			validateContractDocument(t, envelope.Schema, raw)
			if envelope.Command != "explain" || envelope.RequestedCommand != "explain" || envelope.ExitCode != test.exit {
				t.Fatalf("unexpected shared envelope: %s", raw)
			}
			if strings.Contains(string(raw), "secret-key") || strings.Contains(string(raw), "secret-auth") {
				t.Fatal("capture credentials leaked into MCP result")
			}
			if test.code != "" {
				if len(envelope.Errors) != 1 || envelope.Errors[0].Code != test.code {
					t.Fatalf("unexpected typed error: %s", raw)
				}
				if test.exit == 10 && (!strings.Contains(string(raw), `"interaction_type":"capture_selection"`) || !strings.Contains(string(raw), `"entry":2`)) {
					t.Fatalf("missing HAR selection details: %s", raw)
				}
				return
			}
			dataRaw, err := json.Marshal(envelope.Data)
			if err != nil {
				t.Fatal(err)
			}
			var data explainDataView
			if err := json.Unmarshal(dataRaw, &data); err != nil {
				t.Fatal(err)
			}
			if data.Result.Selection.State != test.state || data.Parameter.Key != "flag" || len(data.Conditions) != 2 {
				t.Fatalf("explanation lost evaluation details: %s", raw)
			}
			if test.state == "resolved" && (data.Result.Selection.Condition == nil || *data.Result.Selection.Condition != "c1" || data.Result.Value.Value == nil || *data.Result.Value.Value != "true") {
				t.Fatalf("known inputs did not select the expected value: %s", raw)
			}
			if test.state == "unresolved" && (len(data.Outcomes) != 3 || len(data.NeededInputs) != 1 || data.NeededInputs[0] != "device.os") {
				t.Fatalf("unknown result lost alternatives or needed inputs: %s", raw)
			}
			if test.format != "" && (data.Capture == nil || data.Capture.Format != test.format) {
				t.Fatalf("capture not preserved: %s", raw)
			}
		})
	}
}
