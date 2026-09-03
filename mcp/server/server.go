package mcpserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/schemas"
)

// Execute runs one isolated machine invocation. The server serializes calls;
// the callback may suspend that exact operation for browser authorization.
type Execute func(context.Context, contract.Capability, Invocation, bool, bool, func(core.OAuthAuthorizationEvent)) contract.Envelope

type Server struct {
	Protocol *mcp.Server
	options  Options
	execute  Execute
	version  string
	ctx      context.Context
	cancel   context.CancelFunc
	gate     chan struct{}
	mu       sync.Mutex
	closed   bool
	jobs     map[string]*operation
	wg       sync.WaitGroup
}

type tool struct {
	capability contract.Capability
	schema     *jsonschema.Schema
	defaults   inputDefaults
}

type operation struct {
	ctx             context.Context
	cancel          context.CancelFunc
	key             string
	tool            *tool
	input           Invocation
	canonical       []byte
	session         *mcp.ServerSession
	urlAuth         bool
	form            bool
	done            chan struct{}
	events          chan *inputRequest
	interactionGate chan struct{}
	mu              sync.Mutex
	pending         *inputRequest
	interacted      bool
	result          contract.Envelope
}

type inputRequest struct {
	id       string
	params   *mcp.ElicitParams
	response chan *mcp.ElicitResult
}

func New(ctx context.Context, capabilities []contract.Capability, version string, options Options, execute Execute) (*Server, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	s := &Server{options: options, execute: execute, version: version, ctx: ctx, cancel: cancel, gate: make(chan struct{}, 1), jobs: make(map[string]*operation)}
	s.Protocol = mcp.NewServer(&mcp.Implementation{Name: "fbrcm", Version: version}, &mcp.ServerOptions{
		Capabilities: &mcp.ServerCapabilities{Tools: &mcp.ToolCapabilities{}},
		Instructions: "Firebase Remote Config tools use the existing fbrcm machine contract. Supply required arguments and options. Optional arguments/options wrappers default to empty objects; optional stdin defaults to null. Required fields and conditional constraints still apply. Launch configuration fixes credentials and permissions. Inspect and validate before publishing; authentication is not mutation approval. Never ask users to paste tokens into chat.",
	})
	known := make(map[string]bool, len(capabilities))
	for _, capability := range capabilities {
		known[capability.ID] = true
	}
	for id := range catalog {
		if !known[id] {
			cancel()
			return nil, fmt.Errorf("MCP catalog references unknown command %q", id)
		}
	}
	for _, capability := range capabilities {
		if !options.allows(capability) {
			continue
		}
		input, t, err := toolInputSchema(capability, options)
		if err != nil {
			cancel()
			return nil, err
		}
		properties := input["properties"].(map[string]any)
		optionProperties := properties["options"].(map[string]any)["properties"].(map[string]any)
		output, err := schemas.Bundle(capability.ResponseSchema)
		if err != nil {
			cancel()
			return nil, err
		}
		// Every result is a CLI envelope object. The bundled schema implies
		// this through allOf, but legacy MCP tool definitions require an
		// explicit root type. Preserve all existing envelope constraints.
		output["type"] = "object"
		schemas.MakePortable(output)
		destructive, open := capability.Destructive, capability.NetworkAccess != "none" || options.AllowHooks
		s.Protocol.AddTool(&mcp.Tool{Name: catalog[capability.ID].name, Description: capability.Summary,
			InputSchema: input, OutputSchema: output, Annotations: &mcp.ToolAnnotations{
				DestructiveHint: &destructive, OpenWorldHint: &open, IdempotentHint: capability.Idempotency == "yes",
				// Stateful inspection can update caches and credentials. Do not label
				// it strictly read-only merely because it cannot publish.
				ReadOnlyHint: options.Stateless && !catalog[capability.ID].mutation && optionProperties["to"] == nil && optionProperties["plan-out"] == nil,
			}}, func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return s.call(ctx, req, t)
		})
	}
	return s, nil
}

// Close cancels live operations, including temporary OAuth listeners, and waits
// for their cleanup. Call after the stdio connection ends.
func (s *Server) Close() {
	s.mu.Lock()
	s.closed = true
	s.cancel()
	s.mu.Unlock()
	s.wg.Wait()
	s.mu.Lock()
	clear(s.jobs)
	s.mu.Unlock()
}

func (s *Server) failure(t *tool, err error) contract.Envelope {
	return ops.Failure(s.ctx, t.capability.ID, s.version, s.options.Stateless, err)
}

func completed(envelope contract.Envelope) (*mcp.CallToolResult, error) {
	raw, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{StructuredContent: envelope, Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}, IsError: envelope.Outcome != "success"}, nil
}

func (s *Server) call(ctx context.Context, req *mcp.CallToolRequest, t *tool) (*mcp.CallToolResult, error) {
	if !s.options.allows(t.capability) {
		return completed(s.failure(t, shared.InvalidArgument(fmt.Errorf("tool is disabled by launch policy"))))
	}
	if len(req.Params.Arguments) > 16<<20 {
		return completed(s.failure(t, shared.InvalidArgument(fmt.Errorf("tool input exceeds 16 MiB"))))
	}
	var value any
	arguments := req.Params.Arguments
	if len(arguments) == 0 {
		arguments = json.RawMessage(`{}`)
	}
	decoder := json.NewDecoder(bytes.NewReader(arguments))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return completed(s.failure(t, shared.InvalidArgument(err)))
	}
	t.defaults.normalize(value)
	if err := t.schema.Validate(value); err != nil {
		return completed(s.failure(t, shared.InvalidArgument(fmt.Errorf("invalid tool arguments: %w", err))))
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var invocation Invocation
	if err := json.Unmarshal(canonical, &invocation); err != nil {
		return completed(s.failure(t, shared.InvalidArgument(err)))
	}
	if err := invocation.Validate(t.capability); err != nil {
		return completed(s.failure(t, shared.InvalidArgument(err)))
	}

	s.mu.Lock()
	var op *operation
	if s.closed {
		s.mu.Unlock()
		return completed(s.failure(t, context.Canceled))
	}
	if req.Params.RequestState != "" {
		op = s.jobs[req.Params.RequestState]
		if op == nil || op.tool != t || op.session != req.Session || !bytes.Equal(op.canonical, canonical) {
			s.mu.Unlock()
			return completed(s.failure(t, shared.InvalidArgument(fmt.Errorf("unknown, expired, or mismatched continuation; do not automatically replay a mutation"))))
		}
	} else {
		if len(req.Params.InputResponses) != 0 || len(s.jobs) >= 64 {
			s.mu.Unlock()
			return completed(s.failure(t, shared.InvalidArgument(fmt.Errorf("unsolicited input responses or too many pending operations"))))
		}
		opCtx, cancel := context.WithTimeout(s.ctx, s.options.RequestTimeout)
		op = &operation{ctx: opCtx, cancel: cancel, key: rand.Text(), tool: t, input: invocation, canonical: canonical, session: req.Session, done: make(chan struct{}), events: make(chan *inputRequest, 1), interactionGate: make(chan struct{}, 1)}
		if caps := req.ClientCapabilities(); caps != nil && caps.Elicitation != nil {
			op.urlAuth = caps.Elicitation.URL != nil
			op.form = caps.Elicitation.Form != nil || caps.Elicitation.URL == nil
		}
		op.urlAuth = op.urlAuth && !s.options.Stateless && s.options.BrowserAuth == "auto"
		s.jobs[op.key] = op
		s.wg.Add(1)
		go s.run(op)
	}
	s.mu.Unlock()
	select {
	case <-op.done:
		return completed(op.result)
	default:
	}
	if req.Params.RequestState != "" && len(req.Params.InputResponses) == 0 {
		op.mu.Lock()
		pending := op.pending
		op.mu.Unlock()
		if pending != nil {
			return &mcp.CallToolResult{InputRequests: mcp.InputRequestMap{pending.id: pending.params}, RequestState: op.key}, nil
		}
	}
	if len(req.Params.InputResponses) > 0 {
		op.mu.Lock()
		pending := op.pending
		response, ok := req.Params.InputResponses[inputID(pending)].(*mcp.ElicitResult)
		if pending == nil || !ok || response == nil || len(req.Params.InputResponses) != 1 || (response.Action != "accept" && response.Action != "decline" && response.Action != "cancel") {
			op.mu.Unlock()
			return completed(s.failure(t, shared.InvalidArgument(fmt.Errorf("unexpected interaction response"))))
		}
		op.pending = nil
		op.mu.Unlock()
		pending.response <- response
	}
	select {
	case <-ctx.Done():
		op.cancel()
		return completed(s.failure(t, ctx.Err()))
	case <-op.done:
		return completed(op.result)
	case event := <-op.events:
		return &mcp.CallToolResult{InputRequests: mcp.InputRequestMap{event.id: event.params}, RequestState: op.key}, nil
	}
}

func inputID(in *inputRequest) string {
	if in == nil {
		return ""
	}
	return in.id
}

func (op *operation) ask(ctx context.Context, params *mcp.ElicitParams) bool {
	select {
	case op.interactionGate <- struct{}{}:
		defer func() { <-op.interactionGate }()
	case <-ctx.Done():
		return false
	}
	event := &inputRequest{id: rand.Text(), params: params, response: make(chan *mcp.ElicitResult, 1)}
	op.mu.Lock()
	op.pending = event
	op.interacted = true
	op.mu.Unlock()
	select {
	case op.events <- event:
	case <-ctx.Done():
		return false
	}
	select {
	case response := <-event.response:
		if response.Action != "accept" {
			return false
		}
		if params.Mode == "form" {
			confirmed, _ := response.Content["confirm"].(bool)
			return confirmed
		}
		return true
	case <-ctx.Done():
		return false
	}
}

func (s *Server) run(op *operation) {
	defer s.wg.Done()
	defer op.cancel()
	defer close(op.done)
	defer func() {
		// Retain a completed result briefly so a continuation never reruns work.
		if !op.interacted {
			s.mu.Lock()
			delete(s.jobs, op.key)
			s.mu.Unlock()
			return
		}
		time.AfterFunc(time.Minute, func() { s.mu.Lock(); delete(s.jobs, op.key); s.mu.Unlock() })
	}()
	mutation := catalog[op.tool.capability.ID].mutation || len(op.input.Options["to"]) != 0 || len(op.input.Options["plan-out"]) != 0
	if mutation && s.options.Confirmation == "host" {
		if !op.form {
			op.result = s.failure(op.tool, shared.InteractionRequired("this mutation requires host form confirmation; enable elicitation or explicitly configure --confirmation=none", true, ""))
			return
		}
		message := "Allow fbrcm " + catalog[op.tool.capability.ID].name + " with these exact inputs? " + string(op.canonical)
		if !op.ask(op.ctx, &mcp.ElicitParams{Mode: "form", Message: message, RequestedSchema: confirmationSchema()}) {
			if op.ctx.Err() != nil {
				op.result = s.failure(op.tool, op.ctx.Err())
				return
			}
			op.result = s.failure(op.tool, shared.InteractionRequired("mutation confirmation was declined, canceled, or timed out; no operation was started", true, ""))
			return
		}
	}
	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	case <-op.ctx.Done():
		op.result = s.failure(op.tool, op.ctx.Err())
		return
	}
	if err := op.ctx.Err(); err != nil {
		op.result = s.failure(op.tool, err)
		return
	}
	observer := func(event core.OAuthAuthorizationEvent) {
		if event.Done {
			return
		}
		ctx, cancel := context.WithTimeout(op.ctx, s.options.AuthTimeout)
		defer cancel()
		if !op.ask(ctx, &mcp.ElicitParams{Mode: "url", Message: "Sign in to Google to continue this Firebase operation.", URL: event.URL, ElicitationID: rand.Text()}) {
			if event.Cancel != nil {
				event.Cancel()
			}
		}
	}
	op.result = s.execute(op.ctx, op.tool.capability, op.input, mutation, op.urlAuth, observer)
}

func confirmationSchema() any {
	return json.RawMessage(`{"type":"object","properties":{"confirm":{"type":"boolean","description":"Approve this exact operation"}},"required":["confirm"]}`)
}
