package mcpserver_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/mcp/server"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/schemas"
)

func TestPublishedToolSchemasHaveExplicitObjectRoots(t *testing.T) {
	for _, mode := range []struct {
		name      string
		stateless bool
		writes    bool
		count     int
	}{
		{"stateless_readonly", true, false, 18},
		{"stateless_writable", true, true, 39},
		{"stateful_readonly", false, false, 24},
		{"stateful_writable", false, true, 49},
	} {
		t.Run(mode.name, func(t *testing.T) {
			o := options()
			o.Stateless, o.AllowWrites = mode.stateless, mode.writes
			o.Toolsets = append(o.Toolsets, "diagnostics")
			client, _ := connect(t, o, func(context.Context, contract.Capability, mcpserver.Invocation, bool, bool, func(core.OAuthAuthorizationEvent)) contract.Envelope {
				t.Error("discovery must not execute a tool")
				return contract.Envelope{}
			}, nil)
			listed, err := client.ListTools(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(listed.Tools) != mode.count {
				t.Fatalf("listed %d tools, want %d", len(listed.Tools), mode.count)
			}
			for _, tool := range listed.Tools {
				t.Run(tool.Name, func(t *testing.T) {
					// Validate the JSON sent to clients, not just the Go SDK's
					// permissive 'any' fields. Legacy MCP clients require an
					// explicit root type even when allOf already implies it.
					raw, err := json.Marshal(tool)
					if err != nil {
						t.Fatal(err)
					}
					var published struct {
						Input  map[string]any `json:"inputSchema"`
						Output map[string]any `json:"outputSchema"`
					}
					if err := json.Unmarshal(raw, &published); err != nil {
						t.Fatal(err)
					}
					for name, schema := range map[string]map[string]any{"inputSchema": published.Input, "outputSchema": published.Output} {
						if schema["type"] != "object" {
							t.Errorf("%s.type = %v, want explicit object", name, schema["type"])
						}
						before, _ := json.Marshal(schema)
						schemas.MakePortable(schema)
						after, _ := json.Marshal(schema)
						if string(before) != string(after) {
							t.Errorf("%s contains non-portable schema forms", name)
						}
						compiler := jsonschema.NewCompiler()
						const id = "https://fbrcm.invalid/test"
						if err := compiler.AddResource(id, schema); err != nil {
							t.Fatal(err)
						}
						if _, err := compiler.Compile(id); err != nil {
							t.Fatalf("%s is not self-contained valid JSON Schema: %v", name, err)
						}
					}
					original, err := schemas.Bundle("urn:fbrcm:schema:cli:" + contract.Version + ":command:" + operationID(tool.Name) + ":response")
					if err != nil {
						t.Fatal(err)
					}
					original["type"] = "object"
					schemas.MakePortable(original)
					if !reflect.DeepEqual(published.Output, original) {
						t.Fatal("MCP output schema changed the CLI envelope constraints")
					}
				})
			}
		})
	}
}
