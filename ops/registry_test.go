package ops

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestRepeatedInvocationsHaveIndependentOptionsAndWarnings(t *testing.T) {
	t.Setenv(env.GoogleAccessToken, "")
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	registry, err := NewRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	var capability contract.Capability
	for _, c := range registry.Capabilities() {
		if c.ID == "get" {
			capability = c
		}
	}
	base := registry.factories["get"]
	registry.factories["get"] = func() *invocation.Definition {
		d := base()
		run := d.RunE
		d.RunE = func(call invocation.Call, args []string) error {
			if call.Flags().Changed("search") {
				shared.AddMachineWarning(call, shared.MachineWarning{Code: "cache.stale", Message: "fixture warning", Target: "test", Details: struct {
					Source string `json:"source"`
				}{"fixture"}})
			}
			return run(call, args)
		}
		return d
	}
	input := Input{Arguments: map[string]json.RawMessage{}, Options: map[string]json.RawMessage{"search": json.RawMessage(`"one"`)}, Stdin: json.RawMessage(`{"parameters":{"one":{"defaultValue":{"value":"1"}},"two":{"defaultValue":{"value":"2"}}}}`)}
	execution := Execution{Version: "test", Stateless: true, AuthTimeout: time.Second}
	first := registry.Execute(t.Context(), capability, input, execution)
	input.Options = map[string]json.RawMessage{}
	second := registry.Execute(t.Context(), capability, input, execution)
	for i, envelope := range []contract.Envelope{first, second} {
		if envelope.Outcome != "success" {
			t.Fatalf("call %d failed: %#v", i, envelope)
		}
		raw, err := json.Marshal(envelope.Data)
		if err != nil {
			t.Fatal(err)
		}
		var collection struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(raw, &collection); err != nil {
			t.Fatal(err)
		}
		if collection.Count != i+1 {
			t.Fatalf("call %d count=%d", i, collection.Count)
		}
	}
	if len(first.Warnings) != 1 || len(second.Warnings) != 0 {
		t.Fatal("warning state leaked between invocations")
	}
}

func TestCancellationDoesNotReplaceOperationError(t *testing.T) {
	registry, err := NewRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	base := registry.factories["get"]
	registry.factories["get"] = func() *invocation.Definition {
		d := base()
		d.RunE = func(invocation.Call, []string) error {
			cancel()
			return shared.InvalidArgument(errors.New("operation failure"))
		}
		return d
	}
	var capability contract.Capability
	for _, c := range registry.Capabilities() {
		if c.ID == "get" {
			capability = c
		}
	}
	t.Setenv(env.GoogleAccessToken, "")
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	envelope := registry.Execute(ctx, capability, Input{Stdin: json.RawMessage(`{"parameters":{}}`)}, Execution{Version: "test", Stateless: true})
	if len(envelope.Errors) != 1 || envelope.Errors[0].Code != "argument.invalid" || envelope.ExitCode != 2 {
		t.Fatalf("operation failure replaced by cancellation: %#v", envelope)
	}
}
