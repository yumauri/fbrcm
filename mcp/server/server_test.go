package mcpserver_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/mcp/server"
	"github.com/yumauri/fbrcm/ops"
	"github.com/yumauri/fbrcm/ops/contract"
)

func options() mcpserver.Options {
	return mcpserver.Options{Stateless: true, Toolsets: []string{"inspect", "edit", "drafts", "plans", "publish"}, Confirmation: "host", BrowserAuth: "auto", RequestTimeout: time.Second * 5, AuthTimeout: time.Second}
}

func connect(t *testing.T, o mcpserver.Options, execute mcpserver.Execute, caps *mcp.ClientCapabilities) (*mcp.ClientSession, *mcpserver.Server) {
	t.Helper()
	registry, err := ops.NewRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	server, err := mcpserver.New(t.Context(), registry.Capabilities(), "test", o, execute)
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
	t.Cleanup(func() { _ = cs.Close(); _ = ss.Close(); server.Close() })
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
	if !names["parameters.get"] || names["parameters.update"] || names["draft.show"] || names["auth.login"] || names["config.set"] {
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
		res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.get", Arguments: json.RawMessage(input)})
		if err != nil || !res.IsError {
			t.Fatalf("input %s: %v, %v", input, res, err)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid input reached executor")
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.get", Arguments: json.RawMessage(`{"arguments":{},"options":{},"stdin":null}`)})
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

func TestOptionalWrappersNormalizeBeforeExecution(t *testing.T) {
	var calls atomic.Int32
	o := options()
	o.AllowWrites = true
	o.Confirmation = "none"
	cs, _ := connect(t, o, func(_ context.Context, c contract.Capability, in mcpserver.Invocation, _, _ bool, _ func(core.OAuthAuthorizationEvent)) contract.Envelope {
		calls.Add(1)
		if in.Arguments == nil || in.Options == nil || len(in.Stdin) == 0 {
			t.Errorf("executor received unnormalized invocation: %+v", in)
		}
		return result(c)
	}, nil)
	for _, test := range []struct {
		tool  string
		input string
		valid bool
	}{
		{"parameters.get", "", true}, // The MCP arguments member itself is optional.
		{"parameters.get", `{}`, true},
		{"parameters.get", `{"options":{"project":["=demo"]}}`, true},
		{"parameters.get", `{"stdin":{"parameters":{}}}`, true},
		{"parameters.get", `{"arguments":{"parameter":"feature"}}`, true},
		{"parameters.get", `{"arguments":null}`, false},
		{"parameters.get", `{"options":null}`, false},
		{"parameters.get", `{"options":[]}`, false},
		{"parameters.get", `{"stdin":"invalid"}`, false},
		{"parameters.get", `{"extra":true}`, false},
		{"parameters.get", `{"options":{"yes":true}}`, false},
		{"parameters.get", `{"arguments":{"parameter":"feature"},"options":{"filter":["feature"]}}`, false},
		{"parameters.get", `{"options":{"update":true}}`, false},
		{"parameters.get", `null`, false},
		{"parameters.get", `[]`, false},
		{"conditions.list", `{}`, false},
		{"conditions.list", `{"arguments":{}}`, false},
		{"conditions.list", `{"arguments":{"project":"demo"}}`, true},
		{"conditions.list", `{"arguments":{"project":"^fuzzy"}}`, false},
		{"parameters.add", `{"options":{"type":"string","value":"hello"}}`, false},
		{"parameters.add", `{"arguments":{"parameter":"feature"}}`, false},
		{"parameters.add", `{"arguments":{"parameter":"feature"},"options":{"value":"hello"}}`, false},
		{"parameters.add", `{"arguments":{"parameter":"feature"},"options":{"type":"string","value":"hello"}}`, true},
	} {
		t.Run(test.tool+"/"+test.input, func(t *testing.T) {
			before := calls.Load()
			params := &mcp.CallToolParams{Name: test.tool}
			if test.input != "" {
				params.Arguments = json.RawMessage(test.input)
			}
			res, err := cs.CallTool(t.Context(), params)
			if err != nil || res.IsError == test.valid {
				t.Fatalf("valid=%t: %v %v", test.valid, res, err)
			}
			want := before
			if test.valid {
				want++
			} else if !strings.Contains(res.Content[0].(*mcp.TextContent).Text, `"code":"argument.invalid"`) {
				t.Fatalf("missing typed validation error: %v", res)
			}
			if calls.Load() != want {
				t.Fatalf("executor calls=%d, want %d", calls.Load(), want)
			}
		})
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
	// An omitted wrapper and its explicit default identify the same operation.
	compact := json.RawMessage(strings.ReplaceAll(string(input), `,"stdin":null`, ""))
	initial, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.add", Arguments: compact})
	if err != nil || !initial.NeedsInput() {
		t.Fatalf("expected confirmation: %#v %v", initial, err)
	}
	if calls.Load() != 0 {
		t.Fatal("executed before approval")
	}
	responses := make(mcp.InputResponseMap)
	for id, request := range initial.InputRequests {
		if message := request.(*mcp.ElicitParams).Message; !strings.HasPrefix(message, "Allow fbrcm parameters.add with these exact inputs? ") {
			t.Fatalf("confirmation does not identify the public tool name: %s", message)
		}
		responses[id] = &mcp.ElicitResult{Action: "accept", Content: map[string]any{"confirm": true}}
	}
	wrongTool, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.update", Arguments: input, RequestState: initial.RequestState, InputResponses: responses})
	if err != nil || !wrongTool.IsError || calls.Load() != 0 {
		t.Fatalf("cross-tool continuation accepted: %v %v", wrongTool, err)
	}
	bad, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.add", Arguments: json.RawMessage(strings.ReplaceAll(string(input), "enabled", "different")), RequestState: initial.RequestState, InputResponses: responses})
	if err != nil || !bad.IsError || calls.Load() != 0 {
		t.Fatalf("modified continuation accepted: %v %v", bad, err)
	}
	request := &mcp.CallToolParams{Name: "parameters.add", Arguments: input, RequestState: initial.RequestState, InputResponses: responses}
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
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.add", Arguments: json.RawMessage(`{"arguments":{"parameter":"feature"},"options":{"type":"string","value":"yes","project":["=demo"]},"stdin":null}`)})
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
	initial, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.add", Arguments: input})
	if err != nil || !initial.NeedsInput() {
		t.Fatalf("confirmation: %v %v", initial, err)
	}
	responses := make(mcp.InputResponseMap)
	for id := range initial.InputRequests {
		responses[id] = &mcp.ElicitResult{Action: "decline"}
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.add", Arguments: input, RequestState: initial.RequestState, InputResponses: responses})
	if err != nil || !res.IsError || calls.Load() != 0 {
		t.Fatalf("declined call executed: %v %v", res, err)
	}
	initial, err = cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "parameters.add", Arguments: input})
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
