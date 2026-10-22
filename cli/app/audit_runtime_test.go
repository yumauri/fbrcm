package app

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/publication"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestJSONUnconfiguredAndInvalidInputFailuresConform(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		code          string
		exit          int
		emptyProjects bool
	}{
		{"auth paths without registry", []string{"auth", "path", "missing"}, "auth.setup_required", 4, false},
		{"auth delete without registry", []string{"auth", "delete", "missing", "--yes"}, "auth.setup_required", 4, false},
		{"auth quota set without registry", []string{"auth", "quota-project", "set", "missing", "billing-project"}, "auth.setup_required", 4, false},
		{"auth quota unset without registry", []string{"auth", "quota-project", "unset", "missing"}, "auth.setup_required", 4, false},
		{"get invalid stdin", []string{"get"}, "remote_config.invalid", 8, false},
		{"stateless get invalid stdin", []string{"get", "--stateless"}, "remote_config.invalid", 8, false},
		{"missing plan show", []string{"plan", "show", "missing.json"}, "file.io_failed", 13, false},
		{"missing plan validate", []string{"plan", "validate", "missing.json"}, "file.io_failed", 13, false},
		{"project quota without registry", []string{"project", "quota-project", "set", "missing", "billing-project"}, "file.io_failed", 13, false},
		{"template preferences without registry", []string{"project", "templates", "show", "missing"}, "file.io_failed", 13, false},
		{"forget without registry", []string{"projects", "forget"}, "file.io_failed", 13, false},
		{"forget empty registry", []string{"projects", "forget"}, "project.not_found", 6, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stateRoot := useTemporaryContractState(t)
			t.Setenv(env.Profile, "")
			t.Setenv(env.Offline, "1")
			t.Setenv(env.GoogleAccessToken, "")
			if test.emptyProjects {
				if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
					t.Fatal(err)
				}
				if err := config.SaveProjects(nil, time.Now().UTC()); err != nil {
					t.Fatal(err)
				}
			}
			svc, err := core.NewService(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			root := newRootCommandWithOfflineInit(svc, "test", "test", "test", func(context.Context, bool) {})
			root.SetIn(shared.NonInteractiveInput())
			if test.name == "get invalid stdin" || test.name == "stateless get invalid stdin" {
				root.SetIn(shared.DocumentInput(strings.NewReader("invalid")))
			}
			args := append([]string(nil), test.args...)
			if args[0] == "plan" {
				args[2] = filepath.Join(stateRoot, "missing.json")
			}
			args = append(args, "--json", "--no-local-config")
			envelope, raw := executeJSONContractRoot(t, root, args...)
			if envelope.Outcome != "failure" || envelope.ExitCode != test.exit || len(envelope.Errors) != 1 || envelope.Errors[0].Code != test.code {
				t.Fatalf("envelope = %#v", envelope)
			}
			validateContractDocument(t, envelope.Schema, raw)
		})
	}
}

func TestJSONInvalidPublicationPlanInputsConform(t *testing.T) {
	for _, path := range [][]string{{"plan", "show", "-"}, {"plan", "validate", "-"}, {"apply", "-"}} {
		for _, input := range []string{"", "null", "[]", "{}", "invalid"} {
			t.Run(strings.Join(path, ".")+"/"+input, func(t *testing.T) {
				useTemporaryContractState(t)
				root := NewRootForContract("test")
				root.SetIn(shared.DocumentInput(strings.NewReader(input)))
				envelope, raw := executeJSONContractRoot(t, root, append(append([]string(nil), path...), "--json", "--no-local-config")...)
				if envelope.Outcome != "failure" || envelope.ExitCode != 8 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "plan.invalid" {
					t.Fatalf("envelope = %#v", envelope)
				}
				validateContractDocument(t, envelope.Schema, raw)
			})
		}
	}
}

func TestTypedPlanQuotaAndAppProblemsConform(t *testing.T) {
	useTemporaryContractState(t)
	root := NewRootForContract("test")
	tests := []struct {
		command string
		cause   error
	}{
		{"apply", &publication.Error{Kind: publication.ErrorInvalid, Err: errors.New("invalid plan")}},
		{"apply", &publication.Error{Kind: publication.ErrorIntegrity, Err: errors.New("digest mismatch")}},
		{"apply", &publication.Error{Kind: publication.ErrorUnsupportedVersion, Err: errors.New("unsupported version")}},
		{"apply", &shared.ConflictError{Code: "plan.stale", Resource: "publication_environment", Err: errors.New("environment changed")}},
		{"update", &shared.ConflictError{Code: "plan.exists", Resource: "publication_plan", Err: errors.New("destination exists")}},
		{"projects.list", &firebase.QuotaProjectRequiredError{}},
		{"get", &firebase.QuotaProjectRequiredError{TargetProjectID: "demo"}},
		{"get", &firebase.QuotaProjectInvariantError{Method: "GET", Host: "example.invalid"}},
		{"apps.show", &shared.SelectionError{Resource: "app", Kind: "not_found", Query: "missing", Err: errors.New("missing app")}},
		{"apps.show", &shared.SelectionError{Resource: "app", Kind: "ambiguous", Query: "demo", Err: errors.New("several apps")}},
	}
	for _, test := range tests {
		problem := contract.Classify(test.cause)
		t.Run(test.command+"/"+problem.Code, func(t *testing.T) {
			cmd, _, err := root.Find(strings.Split(test.command, "."))
			if err != nil {
				t.Fatal(err)
			}
			envelope := contract.BuildEnvelope(cmd, "test", nil, test.cause)
			if len(envelope.Errors) != 1 || envelope.Errors[0].Code != problem.Code || envelope.Errors[0].Message != problem.Message {
				t.Fatalf("typed problem replaced by contract enforcement: %#v", envelope)
			}
			validateContractDocument(t, envelope.Schema, marshalEnvelope(t, envelope))
			value := structToContractValue(t, problem).(map[string]any)
			value["category"] = "canceled"
			validateContractValue(t, contract.ErrorSchemaID(), value, false)
			value["category"] = problem.Category
			value["retryable"] = true
			validateContractValue(t, contract.ErrorSchemaID(), value, false)
		})
	}
}
