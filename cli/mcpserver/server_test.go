package mcpserver_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yumauri/fbrcm/cli/app"
	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/mcpserver"
	"github.com/yumauri/fbrcm/core"
)

func options() mcpserver.Options {
	return mcpserver.Options{Stateless: true, Toolsets: []string{"inspect", "edit", "drafts", "plans", "publish"}, Confirmation: "host", BrowserAuth: "auto", RequestTimeout: time.Second * 5, AuthTimeout: time.Second}
}

func connect(t *testing.T, o mcpserver.Options, execute mcpserver.Execute, caps *mcp.ClientCapabilities) (*mcp.ClientSession, *mcpserver.Server) {
	t.Helper()
	root := app.NewRootForContract("test")
	server, err := mcpserver.New(t.Context(), root, "test", o, execute)
	if err != nil {
		t.Fatal(err)
	}
	a, b := mcp.NewInMemoryTransports()
	ss, err := server.Protocol.Connect(t.Context(), a, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, &mcp.ClientOptions{Capabilities: caps, MultiRoundTrip: &mcp.MultiRoundTripOptions{Disabled: true}})
	cs, err := client.Connect(t.Context(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close(); _ = ss.Close(); server.Close(); contract.UnregisterResponses(root) })
	return cs, server
}

func result(c contract.Capability) contract.Envelope {
	return contract.Envelope{Schema: c.ResponseSchema, ContractVersion: contract.Version, Command: c.ID, RequestedCommand: c.ID, Outcome: "success", ExitCode: 1, Producer: contract.Producer{Name: "fbrcm", Version: "test"}, Data: struct {
		Changed bool `json:"changed"`
	}{true}, Errors: []contract.Problem{}, Warnings: []contract.Warning{}}
}

func TestCatalogPoliciesAndInputValidation(t *testing.T) {
	var calls atomic.Int32
	cs, _ := connect(t, options(), func(_ context.Context, c contract.Capability, _ mcpserver.Invocation, _, _ bool, _ func(core.OAuthAuthorizationEvent)) contract.Envelope {
		calls.Add(1)
		return result(c)
	}, nil)
	listed, err := cs.ListTools(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	names := make(map[string]bool)
	for _, tool := range listed.Tools {
		names[tool.Name] = true
	}
	if !names["get"] || names["update"] || names["draft.show"] || names["auth.login"] || names["config.set"] {
		t.Fatalf("unexpected catalog: %v", names)
	}
	for _, input := range []string{
		`{"arguments":{},"options":{"profile":"other"},"stdin":null}`,
		`{"arguments":{},"options":{"yes":true},"stdin":null}`,
		`{"arguments":{},"options":{"unknown":true},"stdin":null}`,
		`{"arguments":{},"options":{"to":"/tmp/unauthorized"},"stdin":null}`,
		`{"arguments":{},"options":{"filter":[]},"stdin":null}`,
		`{"arguments":{},"options":{"update":true},"stdin":null}`,
	} {
		res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "get", Arguments: json.RawMessage(input)})
		if err != nil || !res.IsError {
			t.Fatalf("input %s: %v, %v", input, res, err)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid input reached executor")
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "get", Arguments: json.RawMessage(`{"arguments":{},"options":{},"stdin":null}`)})
	if err != nil || res.IsError {
		t.Fatalf("successful nonzero exit marked error: %v %v", res, err)
	}
	if calls.Load() != 1 {
		t.Fatal("executor not called exactly once")
	}
	if !strings.Contains(res.Content[0].(*mcp.TextContent).Text, `"exit_code":1`) {
		t.Fatal("missing original envelope")
	}
}

func TestMutationConfirmationIsBoundToExactInvocationAndNeverReplayed(t *testing.T) {
	o := options()
	o.AllowWrites = true
	o.Toolsets = []string{"edit"}
	var calls atomic.Int32
	cs, _ := connect(t, o, func(_ context.Context, c contract.Capability, _ mcpserver.Invocation, confirmed, _ bool, _ func(core.OAuthAuthorizationEvent)) contract.Envelope {
		if !confirmed {
			t.Error("mutation was not confirmed")
		}
		calls.Add(1)
		return result(c)
	}, &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{Form: &mcp.FormElicitationCapabilities{}}})
	input := json.RawMessage(`{"arguments":{"parameter":"feature"},"options":{"type":"string","value":"enabled","project":["=demo"]},"stdin":null}`)
	initial, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "add", Arguments: input})
	if err != nil || !initial.NeedsInput() {
		t.Fatalf("expected confirmation: %#v %v", initial, err)
	}
	if calls.Load() != 0 {
		t.Fatal("executed before approval")
	}
	responses := make(mcp.InputResponseMap)
	for id := range initial.InputRequests {
		responses[id] = &mcp.ElicitResult{Action: "accept", Content: map[string]any{"confirm": true}}
	}
	bad, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "add", Arguments: json.RawMessage(strings.ReplaceAll(string(input), "enabled", "different")), RequestState: initial.RequestState, InputResponses: responses})
	if err != nil || !bad.IsError || calls.Load() != 0 {
		t.Fatalf("modified continuation accepted: %v %v", bad, err)
	}
	request := &mcp.CallToolParams{Name: "add", Arguments: input, RequestState: initial.RequestState, InputResponses: responses}
	for range 2 {
		res, err := cs.CallTool(t.Context(), request)
		if err != nil || res.IsError || res.NeedsInput() {
			t.Fatalf("completion: %v %v", res, err)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("operation replayed %d times", calls.Load())
	}
}

func TestOAuthSuspendsOneOperationAndSurvivesInitialRequest(t *testing.T) {
	o := options()
	o.Stateless = false
	o.Toolsets = []string{"plans"}
	var calls atomic.Int32
	cs, _ := connect(t, o, func(ctx context.Context, c contract.Capability, _ mcpserver.Invocation, _, oauth bool, observer func(core.OAuthAuthorizationEvent)) contract.Envelope {
		if !oauth {
			t.Error("URL capability was not recognized")
		}
		calls.Add(1)
		authCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		observer(core.OAuthAuthorizationEvent{URL: "https://accounts.google.com/authorize", Cancel: cancel})
		if authCtx.Err() != nil {
			t.Error("auth canceled after acceptance")
		}
		return result(c)
	}, &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{URL: &mcp.URLElicitationCapabilities{}}})
	input := json.RawMessage(`{"arguments":{"plan":"example.json"},"options":{},"stdin":null}`)
	initial, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "plan.show", Arguments: input})
	if err != nil || !initial.NeedsInput() {
		t.Fatalf("expected sign-in: %v %v", initial, err)
	}
	responses := make(mcp.InputResponseMap)
	for id, request := range initial.InputRequests {
		if request.(*mcp.ElicitParams).Mode != "url" {
			t.Fatal("wrong interaction mode")
		}
		responses[id] = &mcp.ElicitResult{Action: "accept"}
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "plan.show", Arguments: input, RequestState: initial.RequestState, InputResponses: responses})
	if err != nil || res.IsError || calls.Load() != 1 {
		t.Fatalf("auth continuation replayed: %v %v calls=%d", res, err, calls.Load())
	}
}

func TestMissingConfirmationCapabilityNeverExecutesMutation(t *testing.T) {
	o := options()
	o.AllowWrites = true
	o.Toolsets = []string{"edit"}
	cs, _ := connect(t, o, func(context.Context, contract.Capability, mcpserver.Invocation, bool, bool, func(core.OAuthAuthorizationEvent)) contract.Envelope {
		t.Error("unapproved mutation executed")
		return contract.Envelope{}
	}, nil)
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "add", Arguments: json.RawMessage(`{"arguments":{"parameter":"feature"},"options":{"type":"string","value":"yes","project":["=demo"]},"stdin":null}`)})
	if err != nil || !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "interaction.required") {
		t.Fatalf("missing typed interaction error: %v %v", res, err)
	}
}

func TestDeclinedMutationAndCanceledOperation(t *testing.T) {
	o := options()
	o.AllowWrites = true
	o.Toolsets = []string{"edit"}
	var calls atomic.Int32
	cs, server := connect(t, o, func(ctx context.Context, c contract.Capability, _ mcpserver.Invocation, _, _ bool, _ func(core.OAuthAuthorizationEvent)) contract.Envelope {
		calls.Add(1)
		<-ctx.Done()
		return result(c)
	}, &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{Form: &mcp.FormElicitationCapabilities{}}})
	input := json.RawMessage(`{"arguments":{"parameter":"feature"},"options":{"type":"string","value":"yes","project":["=demo"]},"stdin":null}`)
	initial, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "add", Arguments: input})
	if err != nil || !initial.NeedsInput() {
		t.Fatalf("confirmation: %v %v", initial, err)
	}
	responses := make(mcp.InputResponseMap)
	for id := range initial.InputRequests {
		responses[id] = &mcp.ElicitResult{Action: "decline"}
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "add", Arguments: input, RequestState: initial.RequestState, InputResponses: responses})
	if err != nil || !res.IsError || calls.Load() != 0 {
		t.Fatalf("declined call executed: %v %v", res, err)
	}
	initial, err = cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "add", Arguments: input})
	if err != nil || !initial.NeedsInput() {
		t.Fatalf("confirmation: %v %v", initial, err)
	}
	closed := make(chan struct{})
	go func() { server.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("shutdown left confirmation pending")
	}
	if calls.Load() != 0 {
		t.Fatal("shutdown ran pending mutation")
	}
}

func TestBrowserPolicyAndPartialResultMapping(t *testing.T) {
	o := options()
	o.Stateless = false
	o.BrowserAuth = "never"
	o.Toolsets = []string{"plans"}
	cs, _ := connect(t, o, func(_ context.Context, c contract.Capability, _ mcpserver.Invocation, _, oauth bool, _ func(core.OAuthAuthorizationEvent)) contract.Envelope {
		if oauth {
			t.Error("browser-auth=never was ignored")
		}
		envelope := result(c)
		envelope.Outcome = "partial_success"
		envelope.ExitCode = 12
		return envelope
	}, &mcp.ClientCapabilities{Elicitation: &mcp.ElicitationCapabilities{URL: &mcp.URLElicitationCapabilities{}}})
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "plan.show", Arguments: json.RawMessage(`{"arguments":{"plan":"sample.json"},"options":{},"stdin":null}`)})
	if err != nil || !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "partial_success") {
		t.Fatalf("partial result lost: %v %v", res, err)
	}
}
