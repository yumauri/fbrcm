package mcpserver_test

import (
	"context"
	"encoding/json"
	"reflect"
	"regexp"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/mcp/server"
	"github.com/yumauri/fbrcm/ops/contract"
)

// Keep expectations independent of the implementation's catalog mapping.
var renamedOperations = map[string]string{
	"parameters.get":       "get",
	"parameters.add":       "add",
	"parameters.update":    "update",
	"parameters.delete":    "delete",
	"parameters.duplicate": "duplicate",
	"plan.apply":           "apply",
	"diagnostics.doctor":   "doctor",
}

func operationID(name string) string {
	if id, ok := renamedOperations[name]; ok {
		return id
	}
	return name
}

func TestAllToolNamesAreNamespacedAndUnique(t *testing.T) {
	o := options()
	o.Stateless = false
	o.AllowWrites = true
	o.Toolsets = append(o.Toolsets, "diagnostics")
	cs, _ := connect(t, o, func(context.Context, contract.Capability, mcpserver.Invocation, bool, bool, func(core.OAuthAuthorizationEvent)) contract.Envelope {
		t.Error("discovery executed an operation")
		return contract.Envelope{}
	}, nil)
	listed, err := cs.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 49 {
		t.Fatalf("full catalog contains %d tools, want 49", len(listed.Tools))
	}
	qualified := regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)+$`)
	names := make(map[string]bool)
	for _, tool := range listed.Tools {
		if !qualified.MatchString(tool.Name) {
			t.Errorf("tool name %q must have non-empty dotted namespace segments", tool.Name)
		}
		if names[tool.Name] {
			t.Errorf("duplicate tool name %q", tool.Name)
		}
		names[tool.Name] = true
	}
	for name, old := range renamedOperations {
		if !names[name] || names[old] {
			t.Errorf("want %q and no alias %q in discovery", name, old)
		}
	}
}

func TestRenamedToolsDispatchOriginalOperationsAndPreserveEnvelopeIDs(t *testing.T) {
	o := options()
	o.Stateless = false
	o.AllowWrites = true
	o.Confirmation = "none"
	o.Toolsets = append(o.Toolsets, "diagnostics")
	calls := make(chan string, 1)
	cs, _ := connect(t, o, func(_ context.Context, c contract.Capability, _ mcpserver.Invocation, _, _ bool, _ func(core.OAuthAuthorizationEvent)) contract.Envelope {
		calls <- c.ID
		return result(c)
	}, nil)
	for _, test := range []struct {
		name  string
		input string
	}{
		{"parameters.get", `{}`},
		{"parameters.add", `{"arguments":{"parameter":"feature"},"options":{"type":"string","value":"hello"}}`},
		{"parameters.update", `{"arguments":{"parameter":"feature"},"options":{"description":"Updated"}}`},
		{"parameters.delete", `{"arguments":{"parameter":"feature"}}`},
		{"parameters.duplicate", `{"arguments":{"source":"feature","target":"copy"}}`},
		{"plan.apply", `{"arguments":{"plan":"example.json"}}`},
		{"diagnostics.doctor", `{}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			id := operationID(test.name)
			for _, invalid := range []bool{false, true} {
				input := test.input
				if invalid {
					input = `{"options":{"yes":true}}`
				}
				res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: test.name, Arguments: json.RawMessage(input)})
				if err != nil || res.IsError != invalid {
					t.Fatalf("invalid=%t: %v %v", invalid, res, err)
				}
				if invalid {
					select {
					case called := <-calls:
						t.Fatalf("invalid input executed %q", called)
					default:
					}
				} else {
					select {
					case called := <-calls:
						if called != id {
							t.Fatalf("dispatched %q, want %q", called, id)
						}
					default:
						t.Fatal("operation was not executed")
					}
				}
				raw, err := json.Marshal(res.StructuredContent)
				if err != nil {
					t.Fatal(err)
				}
				var envelope contract.Envelope
				if err := json.Unmarshal(raw, &envelope); err != nil {
					t.Fatal(err)
				}
				if envelope.Command != id || envelope.RequestedCommand != id || envelope.Schema != "urn:fbrcm:schema:cli:"+contract.Version+":command:"+id+":response" {
					t.Fatalf("original contract identity changed: %s", raw)
				}
				if len(res.Content) != 1 {
					t.Fatalf("expected one JSON text content block: %v", res.Content)
				}
				var textEnvelope any
				if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &textEnvelope); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(textEnvelope, res.StructuredContent) {
					t.Fatal("text content differs from the structured envelope")
				}
			}
			if res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: id, Arguments: json.RawMessage(test.input)}); err == nil {
				t.Fatalf("old name %q should be unknown, got %v", id, res)
			}
			select {
			case called := <-calls:
				t.Fatalf("old name executed %q", called)
			default:
			}
		})
	}
}
