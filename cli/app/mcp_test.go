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

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/mcpserver"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core/config"
)

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
	options := mcpserver.Options{Stateless: true, RequestTimeout: time.Second, AuthTimeout: time.Second}
	envelope := runHostedMachine(context.Background(), nil, "test", "", "", capability, input, options, false, false, nil, io.Discard)
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
		if !IsMCPInvocation(args) {
			t.Fatalf("missed %v", args)
		}
	}
	for _, args := range [][]string{{"get", "mcp"}, {"--profile", "mcp", "get"}} {
		if IsMCPInvocation(args) {
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
	root := NewRootForContract("test")
	defer contract.UnregisterResponses(root)
	root.SetContext(ctx)
	root.SetArgs([]string{"mcp", "--stateless", "--toolsets", "inspect"})
	root.SetIn(toServerR)
	root.SetOut(fromServerW)
	root.SetErr(io.Discard)
	done := make(chan error, 1)
	go func() { _, err := root.ExecuteC(); done <- err }()
	client := mcp.NewClient(&mcp.Implementation{Name: "stdio-test", Version: "1"}, nil)
	cs, err := client.Connect(ctx, &mcp.IOTransport{Reader: fromServerR, Writer: toServerW}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cs.Close() }()
	for range 2 {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "get", Arguments: json.RawMessage(`{"arguments":{},"options":{},"stdin":{"parameters":{"feature":{"defaultValue":{"value":"hello"}}}}}`)})
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
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown: %v", err)
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
	o := mcpserver.Options{Stateless: true, AllowWrites: true, AuthTimeout: time.Second}
	envelope := runHostedMachine(t.Context(), nil, "test", "", "", capability, input, o, true, false, nil, io.Discard)
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
