package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	protocol "github.com/modelcontextprotocol/go-sdk/mcp"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/ops/contract"
)

// Exercise the real entry point using wire JSON, independently of the Go MCP
// client's permissive Tool.OutputSchema decoding and automatic negotiation.
func TestStdioDiscoveryAndOfflineCallsAcrossProtocolVersions(t *testing.T) {
	for _, version := range []string{"2025-11-25", "2026-07-28"} {
		t.Run(version, func(t *testing.T) {
			t.Setenv(env.GoogleAccessToken, "")
			t.Setenv(env.ConfigDir, t.TempDir())
			t.Setenv(env.CacheDir, t.TempDir())
			t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			input, send := io.Pipe()
			receive, output := io.Pipe()
			closePipes := func() {
				_ = send.Close()
				_ = input.Close()
				_ = output.Close()
				_ = receive.Close()
			}
			stop := context.AfterFunc(ctx, closePipes)
			var stderr bytes.Buffer
			done := make(chan int, 1)
			go func() {
				done <- Run(ctx, nil, "test", "", "", []string{"mcp", "--stateless", "--toolsets", "inspect"}, input, output, &stderr)
			}()
			t.Cleanup(func() {
				closePipes()
				cancel()
				stop()
				<-done
			})
			encoder, decoder := json.NewEncoder(send), json.NewDecoder(receive)
			meta := map[string]any{}
			if version == "2026-07-28" {
				meta = map[string]any{
					protocol.MetaKeyProtocolVersion:    version,
					protocol.MetaKeyClientInfo:         map[string]any{"name": "smoke", "version": "1"},
					protocol.MetaKeyClientCapabilities: map[string]any{},
				}
			}
			id := 0
			request := func(method string, params map[string]any) json.RawMessage {
				t.Helper()
				id++
				if len(meta) != 0 {
					params["_meta"] = meta
				}
				if err := encoder.Encode(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
					t.Fatal(err)
				}
				var reply struct {
					ID     int             `json:"id"`
					Result json.RawMessage `json:"result"`
					Error  json.RawMessage `json:"error"`
				}
				if err := decoder.Decode(&reply); err != nil {
					t.Fatal(err)
				}
				if reply.ID != id || len(reply.Error) != 0 {
					t.Fatalf("%s: unexpected reply: %+v", method, reply)
				}
				return reply.Result
			}
			if version == "2025-11-25" {
				raw := request("initialize", map[string]any{"protocolVersion": version, "capabilities": map[string]any{}, "clientInfo": map[string]any{"name": "smoke", "version": "1"}})
				var initialized protocol.InitializeResult
				if err := json.Unmarshal(raw, &initialized); err != nil || initialized.ProtocolVersion != version {
					t.Fatalf("legacy negotiation: %s (%v)", raw, err)
				}
				if err := encoder.Encode(map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"}); err != nil {
					t.Fatal(err)
				}
			} else {
				raw := request("server/discover", map[string]any{})
				var discovered protocol.DiscoverResult
				if err := json.Unmarshal(raw, &discovered); err != nil || !slices.Contains(discovered.SupportedVersions, version) {
					t.Fatalf("modern discovery: %s (%v)", raw, err)
				}
			}
			var listed struct {
				Tools []struct {
					Name         string         `json:"name"`
					InputSchema  map[string]any `json:"inputSchema"`
					OutputSchema map[string]any `json:"outputSchema"`
				} `json:"tools"`
			}
			if err := json.Unmarshal(request("tools/list", map[string]any{}), &listed); err != nil {
				t.Fatal(err)
			}
			var getSchema map[string]any
			for _, tool := range listed.Tools {
				if !strings.Contains(tool.Name, ".") {
					t.Errorf("unqualified tool name %q", tool.Name)
				}
				if tool.InputSchema["type"] != "object" || tool.OutputSchema["type"] != "object" {
					t.Errorf("%s lacks explicit object schemas", tool.Name)
				}
				if tool.Name == "parameters.get" {
					getSchema = tool.OutputSchema
				}
				if tool.Name == "parameters.update" || tool.Name == "auth.login" {
					t.Errorf("unexpected tool %s", tool.Name)
				}
			}
			if getSchema == nil {
				t.Fatal("parameters.get missing from tool list")
			}
			compiler := jsonschema.NewCompiler()
			const schemaID = "https://fbrcm.invalid/smoke/get"
			if err := compiler.AddResource(schemaID, getSchema); err != nil {
				t.Fatal(err)
			}
			responseSchema, err := compiler.Compile(schemaID)
			if err != nil {
				t.Fatal(err)
			}
			for _, call := range []struct {
				options string
				compact bool
				count   int
				isError bool
			}{
				{`{}`, false, 1, false},
				{`{}`, true, 1, false},
				{`{"search":"missing"}`, true, 0, false},
				{`{"yes":true}`, true, 0, true},
				{`{}`, true, 1, false},
			} {
				input := json.RawMessage(`{"arguments":{},"options":` + call.options + `,"stdin":{"parameters":{"smoke_test":{"defaultValue":{"value":"hello"}}}}}`)
				if call.compact {
					var value map[string]any
					if err := json.Unmarshal(input, &value); err != nil {
						t.Fatal(err)
					}
					delete(value, "arguments")
					if call.options == `{}` {
						delete(value, "options")
					}
					input, err = json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
				}
				raw := request("tools/call", map[string]any{"name": "parameters.get", "arguments": input})
				var result protocol.CallToolResult
				if err := json.Unmarshal(raw, &result); err != nil {
					t.Fatal(err)
				}
				if result.IsError != call.isError {
					t.Fatalf("unexpected error status: %s", raw)
				}
				if err := responseSchema.Validate(result.StructuredContent); err != nil {
					t.Fatalf("result violates advertised output schema: %v", err)
				}
				envelopeJSON, err := json.Marshal(result.StructuredContent)
				if err != nil {
					t.Fatal(err)
				}
				var envelope contract.Envelope
				if err := json.Unmarshal(envelopeJSON, &envelope); err != nil {
					t.Fatal(err)
				}
				if envelope.Command != "get" || envelope.RequestedCommand != "get" || envelope.Schema != "urn:fbrcm:schema:cli:"+contract.Version+":command:get:response" {
					t.Fatalf("renamed tool changed the CLI contract identity: %s", envelopeJSON)
				}
				if call.isError {
					if envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" {
						t.Fatalf("unexpected validation envelope: %s", envelopeJSON)
					}
				} else if envelope.Outcome != "success" || envelope.ExitCode != 0 || envelope.Data.(map[string]any)["count"] != float64(call.count) {
					t.Fatalf("unexpected success envelope: %s", envelopeJSON)
				}
			}
		})
	}
}
