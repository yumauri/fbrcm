package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/ops"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
)

const explainFetchURL = "https://firebaseremoteconfig.googleapis.com/v1/projects/123456789/namespaces/firebase:fetch?key=secret-key"

type explainValueView struct {
	Kind        string  `json:"kind"`
	Value       *string `json:"value"`
	ManagedKind *string `json:"managed_kind"`
}
type explainDataView struct {
	Parameter struct {
		Key, Type    string
		DefaultValue explainValueView `json:"default_value"`
	} `json:"parameter"`
	Inputs []struct {
		Field, State string
		Value        *string
		ReasonCode   string `json:"reason_code"`
		Source       struct {
			Kind        string
			DerivedFrom []string `json:"derived_from"`
		}
		Required bool `json:"required_for_resolution"`
	} `json:"inputs"`
	NeededInputs []string `json:"needed_inputs"`
	Conditions   []struct {
		Name, Selection string
		Evaluation      evaluate.State
		Exclusion       string `json:"exclusion_reason"`
		Rules           []evaluate.RuleEvaluation
	} `json:"conditions"`
	Result struct {
		Selection struct {
			State             string
			Source, Condition *string
		}
		Value explainValueView
	} `json:"result"`
	Outcomes []struct {
		Source    string
		Condition *string
		When      []struct {
			Condition string
			Matches   bool
		}
		Value explainValueView
	} `json:"possible_outcomes"`
	Capture *struct {
		Format   string
		Entry    int
		Response struct {
			Comparison     string
			Reason         string `json:"comparison_reason"`
			ParameterState string `json:"parameter_state"`
			Assignments    []evaluate.Assignment
		}
	} `json:"capture"`
}

func explainContractService(t *testing.T, expressions []string, values []firebase.RemoteConfigValue, defaultValue *firebase.RemoteConfigValue) *core.Core {
	t.Helper()
	useTemporaryContractState(t)
	t.Chdir(t.TempDir())
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveAuth(&config.AuthFile{Version: config.AuthConfigVersion, DefaultAuthID: "main", Auth: []config.AuthEntry{config.DefaultOAuthAuthEntry("main", "Main")}}); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveProjects([]config.Project{{ProjectID: "demo", ProjectNumber: "123456789", Name: "Demo", AuthID: "main"}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	template := firebase.RemoteConfig{Version: firebase.RemoteConfigVersion{VersionNumber: "142"}, Parameters: map[string]firebase.RemoteConfigParam{}}
	parameter := firebase.RemoteConfigParam{ValueType: "BOOLEAN", DefaultValue: defaultValue, ConditionalValues: map[string]firebase.RemoteConfigValue{}}
	for i, expression := range expressions {
		name := fmt.Sprintf("c%d", i+1)
		template.Conditions = append(template.Conditions, firebase.RemoteConfigCondition{Name: name, Expression: expression, TagColor: "GREEN"})
		value := firebase.RemoteConfigValue{Value: "true"}
		if len(values) > i {
			value = values[i]
		}
		parameter.ConditionalValues[name] = value
	}
	template.Parameters["flag"] = parameter
	raw, err := json.Marshal(template)
	if err != nil {
		t.Fatal(err)
	}
	if err := config.SaveParametersCache("demo", &config.ParametersCache{CachedAt: time.Now(), ETag: "etag-142", RemoteConfig: raw}); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func explainHAR(response string, version string) string {
	payload := `{"entries":{"flag":"` + response + `"},"templateVersion":"` + version + `"}`
	encoded, _ := json.Marshal(payload)
	return `{"log":{"entries":[{"startedDateTime":"2026-11-04T12:00:00Z","time":1000,"request":{"method":"POST","url":"` + explainFetchURL + `","headers":[{"name":"Authorization","value":"Bearer secret-auth"}],"postData":{"text":"{}"}},"response":{"status":200,"headers":[{"name":"Date","value":"Wed, 04 Nov 2026 12:00:00 GMT"}],"content":{"text":` + string(encoded) + `}}}]}}`
}

func TestExplainMachineContractRuntime(t *testing.T) {
	for _, tc := range []struct {
		name            string
		expressions     []string
		values          []firebase.RemoteConfigValue
		input           string
		args            []string
		selection, kind string
		wantValue       *string
		code            string
		exit            int
	}{
		{name: "default only", selection: "resolved", kind: "literal", wantValue: new("false")},
		{name: "known winner", expressions: []string{"true", "true"}, selection: "resolved", kind: "literal", wantValue: new("true")},
		{name: "unknown is successful", expressions: []string{`device.os == 'web'`}, selection: "unresolved", kind: "unknown"},
		{name: "override", expressions: []string{`device.os == 'web'`}, args: []string{"--set", "device.os=web"}, selection: "resolved", kind: "literal", wantValue: new("true")},
		{name: "shadowed unknown and constant value", expressions: []string{`device.os == 'web'`, `device.os == 'web'`}, values: []firebase.RemoteConfigValue{{Value: "false"}, {Value: "true"}}, selection: "unresolved", kind: "literal", wantValue: new("false")},
		{name: "short circuit", expressions: []string{`false && app.audiences.inAtLeastOne(['Purchasers'])`}, selection: "resolved", kind: "literal", wantValue: new("false")},
		{name: "unsupported expression", expressions: []string{`future.rule == 'x'`}, selection: "unresolved", kind: "unknown"},
		{name: "bounded captured time", expressions: []string{`dateTime <= dateTime('2026-11-04T12:00:00')`}, input: explainHAR("false", "142"), selection: "unresolved", kind: "unknown"},
		{name: "captured match", expressions: []string{"true"}, input: explainHAR("true", "142"), selection: "resolved", kind: "literal", wantValue: new("true")},
		{name: "captured mismatch", expressions: []string{"true"}, input: explainHAR("false", "142"), selection: "resolved", kind: "literal", wantValue: new("true")},
		{name: "version mismatch warning", expressions: []string{"true"}, input: explainHAR("true", "141"), selection: "resolved", kind: "literal", wantValue: new("true")},
		{name: "managed selected value", expressions: []string{"true"}, values: []firebase.RemoteConfigValue{{PersonalizationValue: json.RawMessage(`{"personalizationId":"p1"}`)}}, selection: "resolved", kind: "managed"},
		{name: "derived platform", expressions: []string{`device.os == 'web'`}, input: `curl '` + explainFetchURL + `' --data '{"app_id":"1:123456789:web:abc"}'`, selection: "resolved", kind: "literal", wantValue: new("true")},
		{name: "typed browser targets", expressions: []string{`app.browserAndVersion.inOne([browserName('Chrome').version.>=('77'), browserName('Firefox').anyVersion])`}, args: []string{"--set", "app.browserAndVersion.name=Chrome", "--set", "app.browserAndVersion.version=78"}, selection: "resolved", kind: "literal", wantValue: new("true")},
		{name: "unicode source spans", expressions: []string{`app.customSignal['表示名'].exactlyMatches(['值']) && device.os == 'web'`}, args: []string{"--set", "app.customSignal['表示名']=值", "--set", "device.os=web"}, selection: "resolved", kind: "literal", wantValue: new("true")},
		{name: "missing parameter", args: []string{"--set", "device.os=/unknown"}, code: "parameter.not_found", exit: 6},
		{name: "invalid capture", input: "curl not-a-firebase-request", code: "capture.invalid", exit: 8},
		{name: "invalid percentile", args: []string{"--set", "percent=101"}, code: "argument.invalid", exit: 2},
		{name: "invalid entry", args: []string{"--entry", "-1"}, code: "argument.invalid", exit: 2},
		{name: "invalid capture status", input: strings.Replace(explainHAR("true", "142"), `"status":200`, `"status":999`, 1), code: "capture.invalid", exit: 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := explainContractService(t, tc.expressions, tc.values, &firebase.RemoteConfigValue{Value: "false"})
			root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
			if tc.input != "" {
				root.SetIn(shared.DocumentInput(strings.NewReader(tc.input)))
			} else {
				root.SetIn(shared.NonInteractiveInput())
			}
			key := "flag"
			if tc.name == "missing parameter" {
				key = "missing"
			}
			args := append([]string{"explain", key, "--project", "demo", "--cached", "--json"}, tc.args...)
			envelope, raw := executeJSONContractRoot(t, root, args...)
			if envelope.ExitCode != tc.exit {
				t.Fatalf("exit=%d want %d: %s", envelope.ExitCode, tc.exit, raw)
			}
			validateContractDocument(t, envelope.Schema, raw)
			if strings.Contains(string(raw), "secret-") || strings.Contains(string(raw), "\x1b") {
				t.Fatalf("capture credentials or terminal styling leaked: %s", raw)
			}
			if tc.code != "" {
				if len(envelope.Errors) != 1 || envelope.Errors[0].Code != tc.code {
					t.Fatalf("failure: %s", raw)
				}
				return
			}
			if envelope.Outcome != "success" {
				t.Fatalf("not a successful explanation: %s", raw)
			}
			encoded, _ := json.Marshal(envelope.Data)
			var data explainDataView
			if err := json.Unmarshal(encoded, &data); err != nil {
				t.Fatal(err)
			}
			if data.Result.Selection.State != tc.selection || data.Result.Value.Kind != tc.kind || (tc.wantValue != nil && (data.Result.Value.Value == nil || *data.Result.Value.Value != *tc.wantValue)) {
				t.Fatalf("wrong decision: %s", raw)
			}
			for i, c := range data.Conditions {
				for _, rule := range c.Rules {
					if rule.Expression != tc.expressions[i][rule.Span.Start:rule.Span.End] {
						t.Fatalf("rule span does not refer to original expression: %#v", rule)
					}
				}
			}
			switch tc.name {
			case "unknown is successful":
				if len(data.NeededInputs) != 1 || data.NeededInputs[0] != "device.os" || len(data.Outcomes) != 2 {
					t.Fatalf("missing outcomes or input requirements: %s", raw)
				}
			case "shadowed unknown and constant value":
				if data.Conditions[1].Exclusion != "subsumed" || len(data.Outcomes) != 2 {
					t.Fatalf("unreachable outcome included: %s", raw)
				}
			case "short circuit":
				if data.Conditions[0].Rules[1].Status != "skipped" || data.Conditions[0].Rules[1].Evaluation != nil || len(data.Inputs) != 0 {
					t.Fatalf("skipped rule evaluated: %s", raw)
				}
			case "unsupported expression":
				if data.Conditions[0].Rules[0].ReasonCode != "unsupported_expression" || len(data.NeededInputs) != 0 {
					t.Fatalf("unsupported expression mislabeled missing input: %s", raw)
				}
			case "bounded captured time":
				if len(data.Inputs) != 1 || data.Inputs[0].State != "bounded" || !data.Inputs[0].Required {
					t.Fatalf("time uncertainty lost: %s", raw)
				}
			case "captured match":
				if data.Capture.Response.Comparison != "match" {
					t.Fatalf("comparison: %s", raw)
				}
			case "captured mismatch":
				if data.Capture.Response.Comparison != "mismatch" {
					t.Fatalf("comparison: %s", raw)
				}
			case "version mismatch warning":
				if data.Capture.Response.Comparison != "inconclusive" || len(envelope.Warnings) != 1 || envelope.Warnings[0].Code != "explain.capture_version_mismatch" {
					t.Fatalf("version mismatch treated as verification: %s", raw)
				}
			case "derived platform":
				if len(data.Inputs) != 2 {
					t.Fatalf("derivation input missing: %s", raw)
				}
			case "typed browser targets":
				if p := data.Conditions[0].Rules[0].Predicate; p.Kind != "client_targets" || len(p.Targets) != 2 || p.Targets[1].Version != nil {
					t.Fatalf("client target structure lost: %s", raw)
				}
			}
		})
	}
}

func TestExplainCaptureSelectionAndAssignmentsConform(t *testing.T) {
	for _, selected := range []bool{false, true} {
		t.Run(fmt.Sprint(selected), func(t *testing.T) {
			svc := explainContractService(t, []string{"true"}, nil, &firebase.RemoteConfigValue{Value: "false"})
			var single struct {
				Log struct{ Entries []json.RawMessage }
			}
			if err := json.Unmarshal([]byte(explainHAR("true", "142")), &single); err != nil {
				t.Fatal(err)
			}
			entry := string(single.Log.Entries[0])
			har := `{"log":{"entries":[{"request":{"method":"GET","url":"https://example.com/"}},` + entry + `,` + entry + `]}}`
			root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
			root.SetIn(shared.DocumentInput(strings.NewReader(har)))
			args := []string{"explain", "flag", "--cached", "--json"}
			if selected {
				args = append(args, "--entry", "3")
			}
			envelope, raw := executeJSONContractRoot(t, root, args...)
			validateContractDocument(t, envelope.Schema, raw)
			if !selected {
				if envelope.ExitCode != 10 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "interaction.required" {
					t.Fatalf("ambiguous capture was not a selection problem: %s", raw)
				}
				detail, _ := json.Marshal(envelope.Errors[0].Details)
				var selection struct {
					Option     string `json:"required_option"`
					Candidates []shared.CaptureCandidate
				}
				if err := json.Unmarshal(detail, &selection); err != nil {
					t.Fatal(err)
				}
				if selection.Option != "--entry" || len(selection.Candidates) != 2 || selection.Candidates[0].Entry != 2 || selection.Candidates[1].Entry != 3 || selection.Candidates[0].Project != "123456789" || strings.Contains(string(detail), "secret-") {
					t.Fatalf("unsafe or incorrect candidates: %s", detail)
				}
			} else {
				encoded, _ := json.Marshal(envelope.Data)
				var data explainDataView
				_ = json.Unmarshal(encoded, &data)
				if envelope.ExitCode != 0 || data.Capture == nil || data.Capture.Entry != 3 {
					t.Fatalf("original entry not selected: %s", raw)
				}
			}
		})
	}
	for _, managed := range []bool{false, true} {
		t.Run("app_default_and_capture_"+fmt.Sprint(managed), func(t *testing.T) {
			svc := explainContractService(t, nil, nil, &firebase.RemoteConfigValue{UseInAppDefault: true})
			input := explainHAR("true", "142")
			if managed {
				input = strings.Replace(input, `\"templateVersion\":\"142\"`, `\"templateVersion\":\"142\",\"personalizationMetadata\":{\"flag\":{}}`, 1)
			}
			root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
			root.SetIn(shared.DocumentInput(strings.NewReader(input)))
			envelope, raw := executeJSONContractRoot(t, root, "explain", "flag", "--project", "demo", "--cached", "--json")
			validateContractDocument(t, envelope.Schema, raw)
			if envelope.ExitCode != 0 {
				t.Fatalf("app default: %s", raw)
			}
			encoded, _ := json.Marshal(envelope.Data)
			var data explainDataView
			_ = json.Unmarshal(encoded, &data)
			if data.Result.Value.Kind != "in_app_default" || data.Result.Value.Value != nil {
				t.Fatalf("in-app default value invented: %s", raw)
			}
			if managed && (len(data.Capture.Response.Assignments) != 1 || data.Capture.Response.Assignments[0].Kind != "personalization" || data.Capture.Response.Comparison != "inconclusive" || data.Capture.Response.Reason != "firebase_managed") {
				t.Fatalf("assignment lost: %s", raw)
			}
		})
	}
}

func TestExplainInvocationSchemasAndStructuredStdin(t *testing.T) {
	svc := explainContractService(t, []string{`device.os == 'web'`}, nil, &firebase.RemoteConfigValue{Value: "false"})
	id := "urn:fbrcm:schema:cli:" + contract.Version + ":command:explain:input"
	for _, tc := range []struct {
		options map[string]any
		stdin   any
		valid   bool
	}{
		{map[string]any{"project": "demo", "cached": true}, nil, true},
		{map[string]any{}, nil, false},
		{map[string]any{"project": "demo", "entry": -1}, nil, false},
		{map[string]any{"project": "demo", "cached": true, "update": true}, nil, false},
		{map[string]any{"project": "demo", "set": []any{"device.category=Mobile", "dateTime=2026-11-04T12:00:00Z"}}, nil, true},
		{map[string]any{"project": "demo", "set": []any{"not_an_input=x"}}, nil, false},
		{map[string]any{}, `curl '` + explainFetchURL + `' --data '{}'`, true},
		// String syntax is checked by the published capture_document runtime
		// validator, which returns capture.invalid for this structural input.
		{map[string]any{}, "not a capture", true},
	} {
		validateContractValue(t, id, map[string]any{"arguments": map[string]any{"parameter": "flag"}, "options": tc.options, "stdin": tc.stdin}, tc.valid)
	}
	registry, err := ops.NewRegistry(svc)
	if err != nil {
		t.Fatal(err)
	}
	var capability contract.Capability
	for _, c := range registry.Capabilities() {
		if c.ID == "explain" {
			capability = c
			break
		}
	}
	if capability.StdinSchema == nil || *capability.StdinSchema != "urn:fbrcm:schema:cli:"+contract.Version+":stdin:capture" {
		t.Fatal("capture stdin schema not advertised")
	}
	curl, _ := json.Marshal(`curl '` + explainFetchURL + `' --data '{"app_id":"1:123456789:web:abc"}'`)
	for _, stdin := range []json.RawMessage{curl, json.RawMessage(explainHAR("true", "142"))} {
		input := ops.Input{Arguments: map[string]json.RawMessage{"parameter": json.RawMessage(`"flag"`)}, Options: map[string]json.RawMessage{"project": json.RawMessage(`"demo"`), "cached": json.RawMessage(`true`)}, Stdin: stdin}
		envelope := registry.Execute(context.Background(), capability, input, ops.Execution{Version: "test", Profile: "default", AuthTimeout: time.Second})
		raw := marshalEnvelope(t, envelope)
		validateContractDocument(t, envelope.Schema, raw)
		if envelope.ExitCode != 0 {
			t.Fatalf("structured capture stdin failed: %s", raw)
		}
	}
}

func TestExplainResponseSchemaRejectsContradictoryStates(t *testing.T) {
	svc := explainContractService(t, []string{`false && device.os == 'web'`}, nil, &firebase.RemoteConfigValue{Value: "false"})
	root := newRootCommandWithOfflineInit(svc, "test", "", "", func(context.Context, bool) {})
	root.SetIn(shared.NonInteractiveInput())
	envelope, raw := executeJSONContractRoot(t, root, "explain", "flag", "--project", "demo", "--cached", "--json")
	if envelope.ExitCode != 0 {
		t.Fatalf("baseline failed: %s", raw)
	}
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"literal cannot be null", func(d map[string]any) { d["result"].(map[string]any)["value"].(map[string]any)["value"] = nil }},
		{"parameter value is a string", func(d map[string]any) { d["result"].(map[string]any)["value"].(map[string]any)["value"] = false }},
		{"unresolved selection has no source", func(d map[string]any) {
			d["result"].(map[string]any)["selection"].(map[string]any)["state"] = "unresolved"
		}},
		{"false cannot be selected", func(d map[string]any) { d["conditions"].([]any)[0].(map[string]any)["selection"] = "selected" }},
		{"priority is positive", func(d map[string]any) { d["conditions"].([]any)[0].(map[string]any)["priority"] = 0 }},
		{"skipped has no evaluation", func(d map[string]any) {
			d["conditions"].([]any)[0].(map[string]any)["rules"].([]any)[1].(map[string]any)["evaluation"] = "false"
		}},
		{"span cannot be negative", func(d map[string]any) {
			d["conditions"].([]any)[0].(map[string]any)["rules"].([]any)[0].(map[string]any)["span"].(map[string]any)["start"] = -1
		}},
		{"literal predicate has no input", func(d map[string]any) {
			d["conditions"].([]any)[0].(map[string]any)["rules"].([]any)[0].(map[string]any)["predicate"].(map[string]any)["field"] = "device.os"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			tc.mutate(doc["data"].(map[string]any))
			validateContractValue(t, envelope.Schema, doc, false)
		})
	}
}
