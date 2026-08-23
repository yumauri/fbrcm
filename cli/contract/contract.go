// Package contract implements the versioned machine-readable CLI protocol.
package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/machine"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corehooks "github.com/yumauri/fbrcm/core/hooks"
	"github.com/yumauri/fbrcm/core/rc/importer"
	"github.com/yumauri/fbrcm/schemas"
)

const Version = "1.0.0"

type Producer struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Context struct {
	Profile *string `json:"profile"`
	Offline bool    `json:"offline"`
	DryRun  bool    `json:"dry_run"`
	Draft   bool    `json:"draft"`
}

type Remediation struct {
	Description string   `json:"description"`
	Strategy    string   `json:"strategy"`
	Argv        []string `json:"argv"`
}

type Problem struct {
	Code        string        `json:"code"`
	Category    string        `json:"category"`
	Message     string        `json:"message"`
	Retryable   bool          `json:"retryable"`
	Target      *string       `json:"target"`
	Stage       *string       `json:"stage"`
	Details     any           `json:"details"`
	Remediation []Remediation `json:"remediation"`
}

type Warning struct {
	Code        string        `json:"code"`
	Message     string        `json:"message"`
	Target      *string       `json:"target"`
	Details     any           `json:"details"`
	Remediation []Remediation `json:"remediation"`
}

type SelectionCandidate struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type SelectionDetails struct {
	Kind       string               `json:"kind"`
	Resource   string               `json:"resource"`
	Query      string               `json:"query"`
	Candidates []SelectionCandidate `json:"candidates"`
}

type BatchTargetProblem struct {
	Target      string        `json:"target"`
	Code        string        `json:"code"`
	Category    string        `json:"category"`
	Message     string        `json:"message"`
	Retryable   bool          `json:"retryable"`
	Stage       *string       `json:"stage"`
	Details     any           `json:"details"`
	Remediation []Remediation `json:"remediation"`
}

type BatchDetails struct {
	Kind                  string               `json:"kind"`
	Operation             string               `json:"operation"`
	FailedTargets         []string             `json:"failed_targets"`
	Failures              []BatchTargetProblem `json:"failures"`
	SuccessfulTargetCount int                  `json:"successful_target_count"`
	PublishedTargetCount  int                  `json:"published_target_count"`
}

type Envelope struct {
	Schema           string    `json:"schema"`
	ContractVersion  string    `json:"contract_version"`
	Command          string    `json:"command"`
	RequestedCommand string    `json:"requested_command"`
	Outcome          string    `json:"outcome"`
	ExitCode         int       `json:"exit_code"`
	Producer         Producer  `json:"producer"`
	Context          Context   `json:"context"`
	Data             any       `json:"data"`
	Errors           []Problem `json:"errors"`
	Warnings         []Warning `json:"warnings"`
}

// JSONRequested recognizes the persistent flag even when Cobra parsing fails.
func JSONRequested(args []string) bool {
	requested := false
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "--json" {
			requested = true
			continue
		}
		if value, ok := strings.CutPrefix(arg, "--json="); ok {
			enabled, err := strconv.ParseBool(value)
			requested = err != nil || enabled
		}
	}
	return requested
}

func Enabled(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		flag = cmd.InheritedFlags().Lookup("json")
	}
	return flag != nil && flag.Value.String() == "true"
}

func CommandID(cmd *cobra.Command) string {
	if cmd == nil {
		return "root"
	}
	path := strings.TrimSpace(strings.TrimPrefix(cmd.CommandPath(), "fbrcm"))
	if path == "" {
		return "root"
	}
	return strings.ReplaceAll(path, " ", ".")
}

func SchemaID(command string) string {
	return "urn:fbrcm:schema:cli:" + Version + ":command:" + command + ":response"
}

// EnvelopeSchemaID identifies the shared response-envelope schema referenced
// by every command response schema.
func EnvelopeSchemaID() string {
	return "urn:fbrcm:schema:cli:" + Version + ":envelope"
}

// ErrorSchemaID identifies the reusable structured-problem schema.
func ErrorSchemaID() string {
	return "urn:fbrcm:schema:cli:" + Version + ":error"
}

func BuildEnvelope(cmd *cobra.Command, version string, captured []byte, err error) Envelope {
	command := envelopeCommandID(cmd)
	code := 0
	problems := make([]Problem, 0, 1)
	outcome := "success"
	if err != nil {
		problem, semanticSuccess := classifyCommandError(cmd, command, err)
		code = ExitCode(cmd, err)
		if semanticSuccess {
			problem = Problem{}
		} else {
			problems = append(problems, problem)
			outcome = "failure"
			if problem.Category == "partial_success" && len(bytes.TrimSpace(captured)) > 0 {
				outcome = "partial_success"
			}
		}
	}
	data, dataErr := decodeData(captured)
	if dataErr != nil {
		problem := Problem{Code: "internal.contract_violation", Category: "internal", Message: shared.SafeErrorText(dataErr), Retryable: false, Details: nil, Remediation: []Remediation{}}
		problems = append(problems, problem)
		outcome = "failure"
		code = 15
		data = nil
	}
	if outcome == "failure" && len(bytes.TrimSpace(captured)) == 0 {
		data = nil
	}
	var profilePtr *string
	if !machine.Profileless(shared.CommandContext(cmd)) {
		profile := config.GetActiveProfileName()
		if strings.TrimSpace(profile) != "" {
			profilePtr = &profile
		}
	}
	envelope := Envelope{
		Schema:           SchemaID(command),
		ContractVersion:  Version,
		Command:          command,
		RequestedCommand: requestedCommandID(cmd),
		Outcome:          outcome,
		ExitCode:         code,
		Producer:         Producer{Name: "fbrcm", Version: version},
		Context:          Context{Profile: profilePtr, Offline: firebase.IsOffline(), DryRun: boolFlag(cmd, "dry-run"), Draft: boolFlag(cmd, "draft")},
		Data:             data,
		Errors:           problems,
		Warnings:         collectedWarnings(cmd),
	}
	if hasResponseModel(cmd) || IsCommandGroup(cmd) {
		if validationErr := validatePublishedEnvelope(envelope); validationErr != nil {
			return contractViolationEnvelope(envelope, validationErr)
		}
	}
	return envelope
}

var compiledEnvelopeSchemas sync.Map // map[string]*jsonschema.Schema

func validatePublishedEnvelope(envelope Envelope) error {
	compiledValue, ok := compiledEnvelopeSchemas.Load(envelope.Schema)
	if !ok {
		compiled, err := compilePublishedSchema(envelope.Schema)
		if err != nil {
			return err
		}
		compiledValue, _ = compiledEnvelopeSchemas.LoadOrStore(envelope.Schema, compiled)
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode response envelope: %w", err)
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("decode response envelope: %w", err)
	}
	if document, ok := value.(map[string]any); ok {
		if err := validateCountCorrelations(document["data"], "data"); err != nil {
			return err
		}
	}
	if err := compiledValue.(*jsonschema.Schema).Validate(value); err != nil {
		return fmt.Errorf("response does not conform to %s: %w", envelope.Schema, err)
	}
	return nil
}

func validateCountCorrelations(value any, path string) error {
	typed, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	rawCount, exists := typed["count"]
	if !exists {
		return nil
	}
	count, valid := rawCount.(float64)
	if !valid {
		return nil
	}
	for _, collectionName := range []string{"items", "commands"} {
		if collection, ok := typed[collectionName].([]any); ok && count != float64(len(collection)) {
			return fmt.Errorf("response count correlation failed at %s: count=%v but %s has %d entries", path, count, collectionName, len(collection))
		}
	}
	return nil
}

func compilePublishedSchema(id string) (*jsonschema.Schema, error) {
	document, err := readSchemaDocument(id)
	if err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	for _, dependencyID := range []string{SemanticSchemaID(), EnvelopeSchemaID(), CapabilitySchemaID()} {
		if dependencyID == id {
			continue
		}
		dependency, readErr := readSchemaDocument(dependencyID)
		if readErr != nil {
			return nil, readErr
		}
		if addErr := compiler.AddResource(dependencyID, dependency); addErr != nil {
			return nil, fmt.Errorf("register schema %s: %w", dependencyID, addErr)
		}
	}
	if err := compiler.AddResource(id, document); err != nil {
		return nil, fmt.Errorf("register schema %s: %w", id, err)
	}
	compiled, err := compiler.Compile(id)
	if err != nil {
		return nil, fmt.Errorf("compile schema %s: %w", id, err)
	}
	return compiled, nil
}

func readSchemaDocument(id string) (any, error) {
	raw, err := schemas.ReadByID(id)
	if err != nil {
		return nil, err
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, fmt.Errorf("decode schema %s: %w", id, err)
	}
	return document, nil
}

func contractViolationEnvelope(envelope Envelope, err error) Envelope {
	envelope.Outcome = "failure"
	envelope.ExitCode = 15
	envelope.Data = nil
	envelope.Errors = []Problem{{
		Code:        "internal.contract_violation",
		Category:    "internal",
		Message:     shared.SafeErrorText(err),
		Retryable:   false,
		Details:     nil,
		Remediation: []Remediation{},
	}}
	envelope.Warnings = []Warning{}
	return envelope
}

func envelopeCommandID(cmd *cobra.Command) string {
	if cmd != nil {
		if help := cmd.Flags().Lookup("help"); help != nil && help.Changed && help.Value.String() == "true" {
			return "help"
		}
		if IsCommandGroup(cmd) {
			if shared.CommandRunStarted(cmd) {
				return "help"
			}
			return "root"
		}
	}
	return CommandID(cmd)
}

func requestedCommandID(cmd *cobra.Command) string {
	if cmd == nil {
		return "root"
	}
	resolved := CommandID(cmd)
	if !IsCommandGroup(cmd) || shared.CommandRunStarted(cmd) {
		return resolved
	}
	parts := []string{}
	if resolved != "root" {
		parts = append(parts, strings.Split(resolved, ".")...)
	}
	for _, argument := range cmd.Flags().Args() {
		if strings.HasPrefix(argument, "-") {
			continue
		}
		parts = append(parts, argument)
	}
	if len(parts) == 0 {
		return "root"
	}
	return strings.Join(parts, ".")
}

func Write(w io.Writer, envelope Envelope) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(envelope)
}

func decodeData(raw []byte) (any, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return struct{}{}, nil
	}
	if json.Valid(raw) {
		switch raw[0] {
		case '[':
			var items []json.RawMessage
			if err := json.Unmarshal(raw, &items); err != nil {
				return nil, err
			}
			return struct {
				Count int               `json:"count"`
				Items []json.RawMessage `json:"items"`
			}{Count: len(items), Items: items}, nil
		case '{':
			return json.RawMessage(append([]byte(nil), raw...)), nil
		default:
			return struct {
				Value json.RawMessage `json:"value"`
			}{Value: append(json.RawMessage(nil), raw...)}, nil
		}
	}
	return struct {
		Text string `json:"text"`
	}{Text: string(raw)}, nil
}

func boolFlag(cmd *cobra.Command, name string) bool {
	if cmd == nil {
		return false
	}
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		flag = cmd.InheritedFlags().Lookup(name)
	}
	return flag != nil && flag.Value.String() == "true"
}

func ExitCode(cmd *cobra.Command, err error) int {
	if err == nil {
		return 0
	}
	var exitErr *shared.ExitError
	if errors.As(err, &exitErr) && exitErr.Code > 0 {
		return exitErr.Code
	}
	problem := Classify(err)
	if cmd != nil && problem.Category == "internal" && !shared.CommandRunStarted(cmd) {
		problem.Category = "argument"
	}
	return exitCodeForCategory(problem.Category)
}

func exitCodeForCategory(category string) int {
	switch category {
	case "argument":
		return 2
	case "configuration", "profile":
		return 3
	case "auth":
		return 4
	case "permission":
		return 5
	case "not_found":
		return 6
	case "conflict":
		return 7
	case "validation":
		return 8
	case "timeout":
		return 9
	case "interaction":
		return 10
	case "unavailable":
		return 11
	case "partial_success":
		return 12
	case "io":
		return 13
	case "hook":
		return 14
	case "canceled":
		return 130
	default:
		return 15
	}
}

func isSemanticResult(err error) bool {
	var exitErr *shared.ExitError
	return errors.As(err, &exitErr) && exitErr.Code == 1 && exitErr.Err == nil
}

func classifyCommandError(cmd *cobra.Command, command string, err error) (Problem, bool) {
	if !isSemanticResult(err) {
		problem := Classify(err)
		if cmd != nil && IsCommandGroup(cmd) && !shared.CommandRunStarted(cmd) && len(cmd.Flags().Args()) > 0 {
			problem.Code, problem.Category = "argument.unknown_command", "argument"
			problem.Details = struct {
				Kind             string `json:"kind"`
				RequestedCommand string `json:"requested_command"`
				ResolvedCommand  string `json:"resolved_command"`
			}{Kind: "invocation", RequestedCommand: requestedCommandID(cmd), ResolvedCommand: CommandID(cmd)}
		} else if cmd != nil && problem.Category == "internal" && !shared.CommandRunStarted(cmd) {
			problem.Code, problem.Category = "argument.invalid", "argument"
		}
		return problem, false
	}
	if strings.HasSuffix(command, ".diff") {
		return Problem{}, true
	}
	message, code := "the command reported an unsuccessful result", "result.unsuccessful"
	switch command {
	case "config.validate":
		message, code = "configuration validation failed", "validation.failed"
	case "doctor":
		message, code = "one or more diagnostic checks failed", "diagnostic.failed"
	}
	return Problem{Code: code, Category: "validation", Message: message, Retryable: false, Details: nil, Remediation: []Remediation{}}, false
}

func Classify(err error) Problem {
	problem := Problem{Code: "internal.unclassified", Category: "internal", Message: shared.SafeErrorText(err), Retryable: false, Details: nil, Remediation: []Remediation{}}
	if _, ok := errors.AsType[*firebase.InvalidChangeNoteError](err); ok {
		problem.Code, problem.Category = "argument.invalid", "argument"
		return problem
	}
	if invalidConfig, ok := errors.AsType[*config.InvalidConfigurationError](err); ok {
		problem.Code, problem.Category = "configuration.invalid", "configuration"
		problem.Target = optionalString(invalidConfig.Path)
		problem.Stage = optionalString(invalidConfig.Stage)
		problem.Details = struct {
			Kind   string `json:"kind"`
			Source string `json:"source"`
		}{Kind: "validation", Source: "configuration"}
		return problem
	}
	if argument, ok := errors.AsType[*shared.ArgumentError](err); ok {
		problem.Code, problem.Category = defaultString(argument.Code, "argument.invalid"), "argument"
		return problem
	}
	if profile, ok := errors.AsType[*config.ProfileError](err); ok {
		problem.Target = optionalString(profile.Profile)
		switch profile.Kind {
		case config.ProfileErrorInvalidArgument:
			problem.Code, problem.Category = "profile.invalid", "profile"
			problem.Details = struct {
				Kind   string `json:"kind"`
				Source string `json:"source"`
			}{Kind: "validation", Source: "profile"}
		case config.ProfileErrorNotFound:
			problem.Code, problem.Category = "profile.not_found", "not_found"
			problem.Details = SelectionDetails{Kind: "selection", Resource: "profile", Query: profile.Profile, Candidates: []SelectionCandidate{}}
		case config.ProfileErrorConflict:
			problem.Code, problem.Category = "profile.conflict", "conflict"
			problem.Details = struct {
				Kind     string `json:"kind"`
				Resource string `json:"resource"`
			}{Kind: "conflict", Resource: "profile"}
		}
		return problem
	}
	if versionLookup, ok := errors.AsType[*core.RemoteConfigVersionLookupError](err); ok {
		problem.Target = optionalString(versionLookup.ProjectID)
		if versionLookup.Kind == "invalid_argument" {
			problem.Code, problem.Category = "argument.invalid", "argument"
			return problem
		}
		problem.Code, problem.Category = "version.not_found", "not_found"
		problem.Details = SelectionDetails{Kind: "selection", Resource: "version", Query: versionLookup.Selector, Candidates: []SelectionCandidate{}}
		return problem
	}
	if managedResource, ok := errors.AsType[*firebase.ManagedFeatureResourceError](err); ok {
		problem.Code, problem.Category = "argument.invalid", "argument"
		problem.Target = optionalString(managedResource.ItemID)
		return problem
	}
	if managedLookup, ok := errors.AsType[*core.ManagedFeatureLookupError](err); ok {
		if managedLookup.Feature == "personalization" {
			problem.Code, problem.Category = "personalization.not_found", "not_found"
			problem.Target = optionalString(managedLookup.ProjectID)
			problem.Details = SelectionDetails{Kind: "selection", Resource: "personalization", Query: managedLookup.ID, Candidates: []SelectionCandidate{}}
			return problem
		}
	}
	if projectLookup, ok := errors.AsType[*core.ProjectLookupError](err); ok {
		problem.Code, problem.Category = "project.not_found", "not_found"
		problem.Details = SelectionDetails{Kind: "selection", Resource: "project", Query: projectLookup.Query, Candidates: []SelectionCandidate{}}
		return problem
	}
	if interaction, ok := errors.AsType[*shared.InteractionError](err); ok {
		problem.Code, problem.Category = "interaction.required", "interaction"
		problem.Retryable = false
		problem.Details = struct {
			Kind           string  `json:"kind"`
			Type           string  `json:"interaction_type"`
			RequiredOption *string `json:"required_option"`
			Destructive    bool    `json:"destructive"`
		}{Kind: "interaction", Type: defaultString(interaction.Type, "input_required"), RequiredOption: optionalString(interaction.RequiredOption), Destructive: interaction.Destructive}
		if len(interaction.SuggestedArgv) > 0 {
			problem.Remediation = []Remediation{{Description: "rerun with the complete non-interactive arguments", Strategy: shared.RemediationRetryWithArguments, Argv: append([]string(nil), interaction.SuggestedArgv...)}}
		}
		return problem
	}
	if expression, ok := errors.AsType[*shared.ExpressionError](err); ok {
		problem.Code, problem.Category = "expression.invalid", "validation"
		problem.Target = optionalString(expression.Target)
		problem.Details = struct {
			Kind       string `json:"kind"`
			Expression string `json:"expression"`
			Context    string `json:"context"`
		}{Kind: "expression", Expression: expression.Expression, Context: expression.Context}
		return problem
	}
	if validation, ok := errors.AsType[*shared.ValidationError](err); ok {
		problem.Code, problem.Category = defaultString(validation.Code, "validation.failed"), "validation"
		switch validation.Source {
		case "profile":
			problem.Category = "profile"
		case "configuration":
			problem.Category = "configuration"
		case "auth":
			problem.Category = "auth"
		}
		problem.Target = optionalString(validation.Target)
		problem.Stage = optionalString(validation.Stage)
		problem.Details = struct {
			Kind   string `json:"kind"`
			Source string `json:"source"`
		}{Kind: "validation", Source: validation.Source}
		return problem
	}
	if authentication, ok := errors.AsType[*firebase.AuthenticationError](err); ok {
		if authErr, ok := errors.AsType[*core.AuthError](err); ok {
			problem.Target = optionalString(authErr.AuthID)
		}
		if authentication.HTTPStatus > 0 {
			problem.Details = struct {
				Kind         string `json:"kind"`
				Service      string `json:"service"`
				Operation    string `json:"operation"`
				HTTPStatus   int    `json:"http_status"`
				RemoteStatus string `json:"remote_status"`
				RemoteCode   string `json:"remote_code"`
				RetryAfterMS int64  `json:"retry_after_ms"`
			}{Kind: "remote_api", Service: "google_auth", Operation: authentication.Operation, HTTPStatus: authentication.HTTPStatus, RemoteCode: authentication.RemoteCode, RetryAfterMS: authentication.RetryAfter.Milliseconds()}
		}
		switch authentication.Kind {
		case firebase.AuthenticationSetupRequired:
			problem.Code, problem.Category = "auth.setup_required", "auth"
			problem.Details = struct {
				Kind string `json:"kind"`
			}{Kind: "auth"}
		case firebase.AuthenticationCredentialsInvalid:
			problem.Code, problem.Category = "auth.credentials_invalid", "auth"
			if authentication.HTTPStatus == 0 {
				problem.Details = struct {
					Kind   string `json:"kind"`
					Source string `json:"source"`
				}{Kind: "validation", Source: "auth"}
			}
		case firebase.AuthenticationRequestFailed:
			problem.Code, problem.Category, problem.Retryable = "network.unavailable", "unavailable", true
			var netErr net.Error
			if authentication.HTTPStatus == http.StatusRequestTimeout || authentication.HTTPStatus == http.StatusGatewayTimeout || errors.As(authentication.Err, &netErr) && netErr.Timeout() {
				problem.Code, problem.Category = "network.timeout", "timeout"
			}
		}
		return problem
	}
	if authErr, ok := errors.AsType[*core.AuthError](err); ok {
		problem.Target = optionalString(authErr.AuthID)
		problem.Details = struct {
			Kind string `json:"kind"`
		}{Kind: "auth"}
		switch authErr.Kind {
		case "invalid_argument":
			problem.Code, problem.Category = "auth.id_invalid", "argument"
		case "not_found":
			problem.Code, problem.Category = "auth.not_found", "not_found"
		case "configuration":
			problem.Code, problem.Category = "auth.configuration_invalid", "auth"
		default:
			problem.Code, problem.Category = "auth.setup_required", "auth"
		}
		return problem
	}
	if oauthInteraction, ok := errors.AsType[*core.OAuthInteractionError](err); ok {
		problem.Code, problem.Category = "interaction.required", "interaction"
		problem.Target = optionalString(oauthInteraction.AuthID)
		problem.Details = struct {
			Kind   string `json:"kind"`
			AuthID string `json:"auth_id"`
		}{Kind: "oauth_authorization", AuthID: oauthInteraction.AuthID}
		problem.Remediation = []Remediation{{Description: "authorize the auth identity in a human session", Strategy: shared.RemediationRunCommand, Argv: []string{"auth", "login", oauthInteraction.AuthID}}}
		return problem
	}
	if conflict, ok := errors.AsType[*shared.ConflictError](err); ok {
		problem.Code, problem.Category = defaultString(conflict.Code, "resource.conflict"), "conflict"
		problem.Retryable = conflict.Retryable
		problem.Target = optionalString(conflict.Target)
		problem.Details = struct {
			Kind     string `json:"kind"`
			Resource string `json:"resource"`
		}{Kind: "conflict", Resource: conflict.Resource}
		problem.Remediation = convertRemediation(conflict.Remediation)
		return problem
	}
	if resolution, ok := errors.AsType[*shared.SelectionError](err); ok {
		resource := defaultString(resolution.Resource, "resource")
		category, code := "not_found", resource+".not_found"
		if resolution.Kind == "ambiguous" {
			category, code = "argument", resource+".ambiguous"
		}
		problem.Category, problem.Code = category, code
		candidates := make([]SelectionCandidate, 0, len(resolution.Candidates))
		candidateIDCounts := make(map[string]int, len(resolution.Candidates))
		for _, candidate := range resolution.Candidates {
			candidates = append(candidates, SelectionCandidate{Name: candidate.Name, ID: candidate.ID})
			if candidate.ID != "" {
				candidateIDCounts[candidate.ID]++
			}
		}
		problem.Details = SelectionDetails{Kind: "selection", Resource: resource, Query: resolution.Query, Candidates: candidates}
		if resolution.Kind == "ambiguous" {
			for _, candidate := range resolution.Candidates {
				if candidate.ID != "" && candidateIDCounts[candidate.ID] == 1 {
					problem.Remediation = append(problem.Remediation, Remediation{Description: "replace the ambiguous " + resource + " selector with this exact ID", Strategy: shared.RemediationReplaceSelector, Argv: []string{candidate.ID}})
				}
			}
		}
		return problem
	}
	if missingGroups, ok := errors.AsType[*importer.MissingGroupsError](err); ok {
		candidates := make([]SelectionCandidate, 0, len(missingGroups.Available))
		for _, group := range missingGroups.Available {
			candidates = append(candidates, SelectionCandidate{Name: group.Name, ID: group.Name})
		}
		problem.Code, problem.Category = "group.not_found", "not_found"
		problem.Details = SelectionDetails{
			Kind: "selection", Resource: "group", Query: strings.Join(missingGroups.Missing, ", "), Candidates: candidates,
		}
		return problem
	}
	if batch, ok := errors.AsType[*shared.BatchError](err); ok {
		problem.Code = "batch.failed"
		problem.Category = "internal"
		failures := classifyBatchFailures(batch.Failures)
		if batch.SuccessfulTargetCount > 0 || batch.PublishedTargetCount > 0 {
			problem.Code, problem.Category = "batch.partial_success", "partial_success"
		} else if len(failures) > 0 {
			problem.Category = failures[0].Category
		}
		problem.Retryable = batchFailuresRetryable(failures)
		problem.Details = BatchDetails{Kind: "batch", Operation: batch.Operation, FailedTargets: batch.FailedTargets, Failures: failures, SuccessfulTargetCount: batch.SuccessfulTargetCount, PublishedTargetCount: batch.PublishedTargetCount}
		problem.Remediation = convertRemediation(batch.Remediation)
		return problem
	}
	if errors.Is(err, context.DeadlineExceeded) {
		problem.Code, problem.Category, problem.Retryable = "command.timeout", "timeout", true
		return problem
	}
	if errors.Is(err, context.Canceled) {
		problem.Code, problem.Category = "command.canceled", "canceled"
		return problem
	}
	if errors.Is(err, firebase.ErrOffline) {
		problem.Code, problem.Category, problem.Retryable = "network.offline", "unavailable", true
		return problem
	}
	if publishedHook, ok := errors.AsType[*core.RemoteConfigPublishedHookError](err); ok {
		problem.Code, problem.Category = "publication.hook_failed", "partial_success"
		problem.Target = optionalString(publishedHook.ProjectID)
		problem.Retryable = false
		return problem
	}
	if publishedCache, ok := errors.AsType[*core.RemoteConfigPublishedCacheError](err); ok {
		problem.Code, problem.Category = "publication.cache_failed", "partial_success"
		problem.Target = optionalString(publishedCache.ProjectID)
		problem.Retryable = false
		return problem
	}
	if hookErr, ok := errors.AsType[*corehooks.Error](err); ok {
		problem.Code, problem.Category = "hook.failed", "hook"
		problem.Target = optionalString(hookErr.Command)
		problem.Details = struct {
			Kind     string `json:"kind"`
			Event    string `json:"event"`
			Index    int    `json:"index"`
			ExitCode int    `json:"exit_code"`
			TimedOut bool   `json:"timed_out"`
			Output   string `json:"output"`
		}{Kind: "hook", Event: string(hookErr.Event), Index: hookErr.Index, ExitCode: hookErr.ExitCode, TimedOut: hookErr.TimedOut, Output: shared.SafeText(hookErr.Output)}
		if hookErr.TimedOut {
			problem.Retryable = true
		}
		return problem
	}
	if netErr, ok := errors.AsType[net.Error](err); ok {
		problem.Code, problem.Category, problem.Retryable = "network.unavailable", "unavailable", true
		if netErr.Timeout() {
			problem.Code, problem.Category = "network.timeout", "timeout"
		}
		return problem
	}
	var remoteValidation *core.RemoteConfigValidationError
	hasRemoteValidation := errors.As(err, &remoteValidation)
	var apiErr *firebase.APIError
	if errors.As(err, &apiErr) && (!hasRemoteValidation || remoteValidation.Source == core.ValidationSourceFirebase && apiErr.StatusCode != http.StatusBadRequest) {
		problem.Retryable = apiErr.Retryable()
		problem.Details = struct {
			Kind         string `json:"kind"`
			Service      string `json:"service"`
			Operation    string `json:"operation"`
			HTTPStatus   int    `json:"http_status"`
			RemoteStatus string `json:"remote_status"`
			RemoteCode   string `json:"remote_code"`
			RetryAfterMS int64  `json:"retry_after_ms"`
		}{Kind: "remote_api", Service: apiErr.Service, Operation: apiErr.Operation, HTTPStatus: apiErr.StatusCode, RemoteStatus: apiErr.RemoteStatus, RemoteCode: apiErr.RemoteCode, RetryAfterMS: apiErr.RetryAfter.Milliseconds()}
		switch apiErr.StatusCode {
		case 401:
			problem.Code, problem.Category = "auth.credentials_invalid", "auth"
		case 403:
			problem.Code, problem.Category = "firebase.permission_denied", "permission"
		case 404:
			problem.Code, problem.Category = "resource.not_found", "not_found"
		case 409, 412:
			problem.Code, problem.Category = "remote_config.conflict", "conflict"
		case 408, 504:
			problem.Code, problem.Category = "firebase.timeout", "timeout"
		case 429:
			problem.Code, problem.Category = "firebase.rate_limited", "unavailable"
		default:
			if apiErr.StatusCode >= 500 {
				problem.Code, problem.Category = "firebase.service_unavailable", "unavailable"
			} else {
				problem.Code, problem.Category = "firebase.request_failed", "validation"
			}
		}
		return problem
	}
	if hasRemoteValidation {
		problem.Code, problem.Category = "remote_config.validation_failed", "validation"
		problem.Details = struct {
			Kind   string `json:"kind"`
			Source string `json:"source"`
		}{Kind: "validation", Source: remoteValidation.Source}
		return problem
	}
	if pathErr, ok := errors.AsType[*os.PathError](err); ok {
		problem.Code, problem.Category = "file.io_failed", "io"
		if errors.Is(err, os.ErrPermission) {
			problem.Code, problem.Category = "filesystem.permission_denied", "permission"
		}
		problem.Details = struct {
			Kind      string `json:"kind"`
			Operation string `json:"operation"`
			Path      string `json:"path"`
		}{Kind: "file", Operation: pathErr.Op, Path: pathErr.Path}
		return problem
	}
	return problem
}

func classifyBatchFailures(items []shared.BatchFailure) []BatchTargetProblem {
	result := make([]BatchTargetProblem, 0, len(items))
	for _, item := range items {
		if item.Err == nil {
			continue
		}
		classified := Classify(item.Err)
		result = append(result, BatchTargetProblem{
			Target:      item.Target,
			Code:        classified.Code,
			Category:    classified.Category,
			Message:     classified.Message,
			Retryable:   classified.Retryable,
			Stage:       classified.Stage,
			Details:     classified.Details,
			Remediation: classified.Remediation,
		})
	}
	return result
}

func batchFailuresRetryable(failures []BatchTargetProblem) bool {
	if len(failures) == 0 {
		return false
	}
	for _, failure := range failures {
		if !failure.Retryable {
			return false
		}
	}
	return true
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	copy := value
	return &copy
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func convertRemediation(items []shared.Remediation) []Remediation {
	result := make([]Remediation, 0, len(items))
	for _, item := range items {
		result = append(result, Remediation{Description: item.Description, Strategy: item.Strategy, Argv: append([]string(nil), item.Argv...)})
	}
	return result
}

func collectedWarnings(cmd *cobra.Command) []Warning {
	items := shared.MachineWarnings(cmd)
	result := make([]Warning, 0, len(items))
	for _, item := range items {
		result = append(result, Warning{Code: item.Code, Message: shared.SafeText(item.Message), Target: optionalString(item.Target), Details: item.Details, Remediation: convertRemediation(item.Remediation)})
	}
	return result
}

func NewBuffer() *bytes.Buffer { return &bytes.Buffer{} }

func FormatHumanError(err error) string {
	if err == nil || err.Error() == "" {
		return ""
	}
	return fmt.Sprintf("error: %v\n", err)
}
