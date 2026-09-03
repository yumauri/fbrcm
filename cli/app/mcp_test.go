package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yumauri/fbrcm/core/config"
	frontend "github.com/yumauri/fbrcm/mcp"
	"github.com/yumauri/fbrcm/mcp/server"
	"github.com/yumauri/fbrcm/ops"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/schemas"
)

func TestMCPEmbeddedCapabilitiesMatchLiveDefinitions(t *testing.T) {
	root := NewRootForContract("test")
	defer contract.UnregisterResponses(root)
	raw, err := json.MarshalIndent(contract.DetailedCapabilities(root), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.TrimSpace(raw), bytes.TrimSpace(schemas.CapabilitiesJSON)) {
		t.Fatal("MCP metadata is stale; run go run ./cmd/schemagen")
	}
}

func TestMCPStandaloneHelpPreservesCLIEnvelope(t *testing.T) {
	state := t.TempDir()
	t.Setenv("FBRCM_CONFIG_DIR", filepath.Join(state, "config"))
	t.Setenv("FBRCM_CACHE_DIR", filepath.Join(state, "cache"))
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	args := []string{"mcp", "--help", "--json"}
	root := NewRootForContract("test")
	defer contract.UnregisterResponses(root)
	var captured bytes.Buffer
	root.SetArgs(args)
	root.SetOut(&captured)
	root.SetErr(io.Discard)
	cmd, err := root.ExecuteC()
	want := contract.BuildEnvelope(cmd, "test", captured.Bytes(), err)
	var output bytes.Buffer
	code := frontend.Run(t.Context(), nil, "test", "", "", args, bytes.NewReader(nil), &output, io.Discard)
	if code != want.ExitCode {
		t.Fatalf("standalone exit %d, CLI exit %d", code, want.ExitCode)
	}
	var got any
	if err := json.Unmarshal(output.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	wantRaw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var wantValue any
	if err := json.Unmarshal(wantRaw, &wantValue); err != nil {
		t.Fatal(err)
	}
	gotRaw, _ := json.Marshal(got)
	normalizedWant, _ := json.Marshal(wantValue)
	if !bytes.Equal(gotRaw, normalizedWant) {
		t.Fatalf("standalone help differs:\n%s\n%s", gotRaw, normalizedWant)
	}
}

func TestMCPJSONFailureDoesNotBootstrapProfile(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(state, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(state, "cache"))
	t.Setenv("FBRCM_PROFILE", "")
	config.SetLocalConfigDisabled(true)
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	root := NewRootForContract("test")
	defer contract.UnregisterResponses(root)
	root.SetArgs([]string{"mcp", "--json"})
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	cmd, err := root.ExecuteC()
	if err == nil {
		t.Fatal("mcp --json accepted")
	}
	envelope := contract.BuildEnvelope(cmd, "test", nil, err)
	if envelope.ExitCode != 2 || envelope.Errors[0].Code != "argument.invalid" || envelope.Context.Profile != nil {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
	entries, err := os.ReadDir(state)
	if err != nil || len(entries) != 0 {
		t.Fatalf("startup persisted state: %v %v", entries, err)
	}
}

func TestMCPJSONContractDescribesOnlyEarlyTransportRejection(t *testing.T) {
	root := NewRootForContract("test")
	defer contract.UnregisterResponses(root)
	capability, err := contract.FindCapability(root, []string{"mcp"})
	if err != nil {
		t.Fatal(err)
	}
	if capability.Supports.Stateless || capability.SideEffectLevel != 0 || capability.NetworkAccess != "none" || capability.Interaction.Mode != "none" {
		t.Fatalf("mcp JSON behavior = %#v", capability)
	}
	if len(capability.ProblemCodes) != 1 || capability.ProblemCodes[0] != "argument.invalid" {
		t.Fatalf("mcp JSON problem codes = %v", capability.ProblemCodes)
	}
	for _, flag := range capability.Flags {
		if flag.Effective || len(flag.EffectiveWhen) != 0 {
			t.Errorf("mcp JSON flag %s is marked effective: %#v", flag.Name, flag)
		}
	}

	validButServerInvalid := map[string]any{
		"arguments": map[string]any{},
		"options": map[string]any{
			"allow-hooks": true, "allow-writes": true,
			"auth-timeout": "-1s", "browser-auth": "unsupported", "confirmation": "unsupported",
			"no-local-config": true, "profile": "work", "request-timeout": "0s",
			"stateless": true, "timeout": "-1s", "toolsets": []any{},
		},
		"stdin": nil,
	}
	validateContractValue(t, capability.InvocationSchema, validButServerInvalid, true)
	for _, invalid := range []map[string]any{
		{"arguments": map[string]any{}, "options": map[string]any{"profile": " "}, "stdin": nil},
		{"arguments": map[string]any{}, "options": map[string]any{"toolsets": []any{""}}, "stdin": nil},
		{"arguments": map[string]any{}, "options": map[string]any{"timeout": "not-a-duration"}, "stdin": nil},
	} {
		validateContractValue(t, capability.InvocationSchema, invalid, false)
	}

	var output bytes.Buffer
	code := frontend.Run(t.Context(), nil, "test", "", "", []string{
		"mcp", "--json", "--profile", "work", "--stateless", "--allow-hooks",
		"--confirmation", "unsupported", "--browser-auth", "unsupported",
		"--request-timeout", "0s", "--auth-timeout", "-1s", "--toolsets=",
	}, bytes.NewReader(nil), &output, io.Discard)
	if code != 2 {
		t.Fatalf("mcp JSON exit = %d, output = %s", code, output.Bytes())
	}
	var envelope contract.Envelope
	if err := json.Unmarshal(output.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Outcome != "failure" || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" || envelope.Context.Profile != nil {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
	validateContractValue(t, capability.ResponseSchema, structToContractValue(t, envelope), true)
}

func TestHostedStatelessStdinUsesExistingContractAndNoLocalState(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(state, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(state, "cache"))
	t.Setenv("FBRCM_GOOGLE_ACCESS_TOKEN", "")
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	root := NewRootForContract("test")
	defer contract.UnregisterResponses(root)
	capability, err := contract.FindCapability(root, []string{"get"})
	if err != nil {
		t.Fatal(err)
	}
	input := mcpserver.Invocation{Arguments: map[string]json.RawMessage{}, Options: map[string]json.RawMessage{}, Stdin: json.RawMessage(`{"parameters":{"flag":{"defaultValue":{"value":"hello"}}}}`)}
	registry, err := ops.NewRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	envelope := registry.Execute(context.Background(), capability, input, ops.Execution{Version: "test", Stateless: true, AuthTimeout: time.Second})
	if envelope.Outcome != "success" {
		t.Fatalf("hosted result: %#v", envelope)
	}
	cli := NewRootForContract("test")
	defer contract.UnregisterResponses(cli)
	cli.SetContext(shared.WithMachineState(t.Context()))
	cli.SetArgs([]string{"get", "--json", "--stateless"})
	cli.SetIn(shared.DocumentInput(bytes.NewReader(input.Stdin)))
	var output bytes.Buffer
	cli.SetOut(&output)
	cli.SetErr(io.Discard)
	command, err := cli.ExecuteC()
	want := contract.BuildEnvelope(command, "test", output.Bytes(), err)
	a, _ := json.Marshal(envelope)
	b, _ := json.Marshal(want)
	if !bytes.Equal(a, b) {
		t.Fatalf("MCP/CLI contract differs:\n%s\n%s", a, b)
	}
	entries, err := os.ReadDir(state)
	if err != nil || len(entries) != 0 {
		t.Fatalf("stateless execution persisted state: %v %v", entries, err)
	}
}

func TestMCPInvocationDetection(t *testing.T) {
	for _, args := range [][]string{{"mcp"}, {"--profile", "work", "mcp"}, {"mcp", "--stateless"}} {
		if !frontend.IsInvocation(args) {
			t.Fatalf("missed %v", args)
		}
	}
	for _, args := range [][]string{{"get", "mcp"}, {"--profile", "mcp", "get"}} {
		if frontend.IsInvocation(args) {
			t.Fatalf("misidentified %v", args)
		}
	}
}

func TestMCPStdioLifecycleAndContract(t *testing.T) {
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	toServerR, toServerW := io.Pipe()
	fromServerR, fromServerW := io.Pipe()
	defer func() { _ = toServerR.Close(); _ = toServerW.Close(); _ = fromServerR.Close(); _ = fromServerW.Close() }()
	done := make(chan int, 1)
	go func() {
		done <- frontend.Run(ctx, nil, "test", "", "", []string{"mcp", "--stateless", "--toolsets", "inspect"}, toServerR, fromServerW, io.Discard)
	}()
	client := mcp.NewClient(&mcp.Implementation{Name: "stdio-test", Version: "1"}, nil)
	cs, err := client.Connect(ctx, &mcp.IOTransport{Reader: fromServerR, Writer: toServerW}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cs.Close() }()
	for range 2 {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "parameters.get", Arguments: json.RawMessage(`{"arguments":{},"options":{},"stdin":{"parameters":{"feature":{"defaultValue":{"value":"hello"}}}}}`)})
		if err != nil || res.IsError {
			t.Fatalf("stdio call: %v %v", res, err)
		}
		var envelope contract.Envelope
		if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Command != "get" || envelope.Outcome != "success" || envelope.Context.Profile != nil {
			t.Fatalf("unexpected envelope: %#v", envelope)
		}
		validateContractValue(t, envelope.Schema, structToContractValue(t, envelope), true)
	}
	_ = cs.Close()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("shutdown: %v", code)
		}
	case <-ctx.Done():
		t.Fatal("stdio server did not stop on EOF")
	}
}

func TestHostedMutationPreservesArtifactContractAndEmptyGroups(t *testing.T) {
	t.Setenv("FBRCM_GOOGLE_ACCESS_TOKEN", "unused-test-token")
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	root := NewRootForContract("test")
	defer contract.UnregisterResponses(root)
	capability, err := contract.FindCapability(root, []string{"add"})
	if err != nil {
		t.Fatal(err)
	}
	input := mcpserver.Invocation{Arguments: map[string]json.RawMessage{"parameter": json.RawMessage(`"feature"`)}, Options: map[string]json.RawMessage{"type": json.RawMessage(`"boolean"`), "value": json.RawMessage(`"true"`)}, Stdin: json.RawMessage(`{"parameters":{},"parameterGroups":{"empty":{"description":"preserve"}}}`)}
	registry, err := ops.NewRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	envelope := registry.Execute(t.Context(), capability, input, ops.Execution{Version: "test", Stateless: true, Confirmed: true, AuthTimeout: time.Second})
	if envelope.Outcome != "success" {
		t.Fatalf("mutation result: %#v", envelope)
	}
	validateContractValue(t, envelope.Schema, structToContractValue(t, envelope), true)
	raw, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatal(err)
	}
	var artifact contract.ArtifactData
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.Encoding != "json" || !bytes.Contains(artifact.JSONContent, []byte(`"empty"`)) || !bytes.Contains(artifact.JSONContent, []byte(`"feature"`)) {
		t.Fatalf("artifact lost mutation or empty group: %s", raw)
	}
}
