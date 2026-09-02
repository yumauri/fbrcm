package contract

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	corehooks "github.com/yumauri/fbrcm/core/hooks"
	"github.com/yumauri/fbrcm/core/rc/importer"
	"github.com/yumauri/fbrcm/ops/machine"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestClassifyStableProblemCategories(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		code     string
		category string
		exitCode int
	}{
		{"argument", shared.InvalidArgument(errors.New("unknown flag: --wat")), "argument.invalid", "argument", 2},
		{"change note", &firebase.InvalidChangeNoteError{Err: errors.New("change note must be one line")}, "argument.invalid", "argument", 2},
		{"invalid version", &core.RemoteConfigVersionLookupError{Kind: "invalid_argument", ProjectID: "demo", Selector: "bad", Err: errors.New("invalid version")}, "argument.invalid", "argument", 2},
		{"invalid profile", &config.ProfileError{Kind: config.ProfileErrorInvalidArgument, Profile: "bad/profile", Err: errors.New("invalid profile")}, "profile.invalid", "profile", 3},
		{"profile", &shared.ValidationError{Code: "profile.invalid", Source: "profile", Err: errors.New("profile dev does not exist")}, "profile.invalid", "profile", 3},
		{"configuration", &config.InvalidConfigurationError{Path: "projects.json", Stage: "decoding", Err: errors.New("invalid JSON")}, "configuration.invalid", "configuration", 3},
		{"invalid theme", &config.ThemeError{Kind: config.ThemeErrorInvalidArgument, Theme: "broken", Err: errors.New("invalid theme")}, "theme.invalid", "configuration", 3},
		{"auth", &shared.ValidationError{Code: "auth.failed", Source: "auth", Err: errors.New("OAuth credentials are invalid")}, "auth.failed", "auth", 4},
		{"permission", &firebase.APIError{Service: "remote-config", Operation: "get", StatusCode: 403}, "firebase.permission_denied", "permission", 5},
		{"not found", &firebase.APIError{Service: "remote-config", Operation: "get", StatusCode: 404}, "resource.not_found", "not_found", 6},
		{"version not found", &core.RemoteConfigVersionLookupError{Kind: "not_found", ProjectID: "demo", Selector: "7", Err: errors.New("version not found")}, "version.not_found", "not_found", 6},
		{"malformed managed resource", &firebase.ManagedFeatureResourceError{Collection: "experiments", ItemID: "bad/resource", Err: errors.New("invalid resource")}, "argument.invalid", "argument", 2},
		{"personalization not found", &core.ManagedFeatureLookupError{Feature: "personalization", ProjectID: "demo", ID: "missing", Err: errors.New("personalization not found")}, "personalization.not_found", "not_found", 6},
		{"import group not found", &importer.MissingGroupsError{Missing: []string{"missing"}}, "group.not_found", "not_found", 6},
		{"profile not found", &config.ProfileError{Kind: config.ProfileErrorNotFound, Profile: "missing", Err: errors.New("profile not found")}, "profile.not_found", "not_found", 6},
		{"theme not found", &config.ThemeError{Kind: config.ThemeErrorNotFound, Theme: "missing", Err: errors.New("theme not found")}, "theme.not_found", "not_found", 6},
		{"conflict", &firebase.APIError{Service: "remote-config", Operation: "update", StatusCode: 412}, "remote_config.conflict", "conflict", 7},
		{"profile conflict", &config.ProfileError{Kind: config.ProfileErrorConflict, Profile: "active", Err: errors.New("profile conflict")}, "profile.conflict", "conflict", 7},
		{"theme conflict", &config.ThemeError{Kind: config.ThemeErrorConflict, Theme: "active", Err: errors.New("theme conflict")}, "theme.conflict", "conflict", 7},
		{"expression", &shared.ExpressionError{Expression: "project_id ==", Context: "project", Err: errors.New("unexpected end")}, "expression.invalid", "validation", 8},
		{"validation", shared.InvalidInput("validation.failed", "stdin", errors.New("invalid JSON input")), "validation.failed", "validation", 8},
		{"timeout", context.DeadlineExceeded, "command.timeout", "timeout", 9},
		{"interaction", shared.InteractionRequired("choose a value", false, "--value"), "interaction.required", "interaction", 10},
		{"unavailable", firebase.ErrOffline, "network.offline", "unavailable", 11},
		{"partial", &shared.BatchError{FailedTargets: []string{"failed"}, SuccessfulTargetCount: 1}, "batch.partial_success", "partial_success", 12},
		{"io", &os.PathError{Op: "open", Path: "/missing", Err: os.ErrNotExist}, "file.io_failed", "io", 13},
		{"hook", &corehooks.Error{Event: corehooks.PrePublish, Err: errors.New("failed")}, "hook.failed", "hook", 14},
		{"canceled", context.Canceled, "command.canceled", "canceled", 130},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			problem := Classify(test.err)
			if problem.Code != test.code || problem.Category != test.category {
				t.Fatalf("problem = %#v", problem)
			}
			if got := ExitCode(nil, test.err); got != test.exitCode {
				t.Fatalf("exit code = %d, want %d", got, test.exitCode)
			}
			envelope := BuildEnvelope(nil, "test", nil, test.err)
			if envelope.ExitCode != test.exitCode {
				t.Fatalf("envelope exit code = %d, want %d", envelope.ExitCode, test.exitCode)
			}
		})
	}
}

func TestMissingImportGroupsPreserveSelectionDetails(t *testing.T) {
	err := &importer.MissingGroupsError{
		Missing:   []string{"missing-a", "missing-b"},
		Available: []importer.GroupSummary{{Name: "available", Parameters: 2}},
	}
	problem := Classify(err)
	details, ok := problem.Details.(SelectionDetails)
	if !ok || details.Kind != "selection" || details.Resource != "group" || details.Query != "missing-a, missing-b" || len(details.Candidates) != 1 || details.Candidates[0].Name != "available" || details.Candidates[0].ID != "available" {
		t.Fatalf("problem details = %#v", problem.Details)
	}
}

func TestSelectionProblemKeepsResourceNeutralCandidates(t *testing.T) {
	err := &shared.ProjectResolutionError{Resource: "project", Kind: "ambiguous", Query: "prod", Candidates: []shared.SelectionCandidate{{Name: "Prod EU", ID: "prod-eu"}}}
	problem := Classify(err)
	details, ok := problem.Details.(SelectionDetails)
	if !ok || details.Kind != "selection" || details.Resource != "project" || details.Query != "prod" || len(details.Candidates) != 1 || details.Candidates[0].Name != "Prod EU" || details.Candidates[0].ID != "prod-eu" {
		t.Fatalf("problem details = %#v", problem.Details)
	}
	if len(problem.Remediation) != 1 || problem.Remediation[0].Strategy != shared.RemediationReplaceSelector || len(problem.Remediation[0].Argv) != 1 || problem.Remediation[0].Argv[0] != "prod-eu" {
		t.Fatalf("problem remediation = %#v", problem.Remediation)
	}
}

func TestParameterAmbiguityIsAnArgumentFailureWithoutDuplicateRemediation(t *testing.T) {
	err := &shared.SelectionError{Resource: "parameter", Kind: "ambiguous", Query: "flag", Candidates: []shared.SelectionCandidate{{Name: "flag (group-a)", ID: "flag"}, {Name: "flag (group-b)", ID: "flag"}}}
	problem := Classify(err)
	if problem.Code != "parameter.ambiguous" || problem.Category != "argument" || ExitCode(nil, err) != 2 {
		t.Fatalf("problem = %#v", problem)
	}
	if len(problem.Remediation) != 0 {
		t.Fatalf("duplicate candidate IDs produced unusable remediation: %#v", problem.Remediation)
	}
}

func TestAllFailedBatchPreservesTypedTargetProblems(t *testing.T) {
	err := &shared.BatchError{
		Operation:     "update",
		FailedTargets: []string{"first", "second"},
		Failures: []shared.BatchFailure{
			{Target: "first", Err: &firebase.APIError{Service: "remote-config", Operation: "update", StatusCode: 403}},
			{Target: "second", Err: &firebase.APIError{Service: "remote-config", Operation: "update", StatusCode: 403}},
		},
	}
	problem := Classify(err)
	if problem.Code != "batch.failed" || problem.Category != "permission" || problem.Retryable || ExitCode(nil, err) != 5 {
		t.Fatalf("problem = %#v", problem)
	}
	details, ok := problem.Details.(BatchDetails)
	if !ok || len(details.Failures) != 2 || details.Failures[0].Target != "first" || details.Failures[0].Code != "firebase.permission_denied" {
		t.Fatalf("batch details = %#v", problem.Details)
	}
}

func TestBatchIsRetryableOnlyWhenEveryFailureIsRetryable(t *testing.T) {
	err := &shared.BatchError{Failures: []shared.BatchFailure{
		{Target: "retryable", Err: &firebase.APIError{Service: "remote-config", Operation: "update", StatusCode: 503}},
		{Target: "invalid", Err: shared.InvalidArgument(errors.New("invalid selector"))},
	}}
	problem := Classify(err)
	if problem.Retryable {
		t.Fatalf("mixed batch is retryable: %#v", problem)
	}
}

func TestValidationFailuresPreserveLocalAndFirebaseSource(t *testing.T) {
	for _, source := range []string{core.ValidationSourceLocal, core.ValidationSourceFirebase} {
		t.Run(source, func(t *testing.T) {
			problem := Classify(&core.RemoteConfigValidationError{Source: source, Err: errors.New("invalid remote config")})
			raw, err := json.Marshal(problem.Details)
			if err != nil {
				t.Fatal(err)
			}
			if problem.Code != "remote_config.validation_failed" || problem.Category != "validation" || !strings.Contains(string(raw), `"source":"`+source+`"`) {
				t.Fatalf("problem = %#v details=%s", problem, raw)
			}
		})
	}
}

func TestRemoteValidationWrapperPreservesTypedAPIFailures(t *testing.T) {
	for _, test := range []struct {
		name      string
		status    int
		code      string
		category  string
		exitCode  int
		retryable bool
	}{
		{name: "permission", status: 403, code: "firebase.permission_denied", category: "permission", exitCode: 5},
		{name: "rate limit", status: 429, code: "firebase.rate_limited", category: "unavailable", exitCode: 11, retryable: true},
		{name: "service unavailable", status: 503, code: "firebase.service_unavailable", category: "unavailable", exitCode: 11, retryable: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := &core.RemoteConfigValidationError{Source: core.ValidationSourceFirebase, Err: &firebase.APIError{Service: "remote-config", Operation: "validate", StatusCode: test.status}}
			problem := Classify(err)
			if problem.Code != test.code || problem.Category != test.category || problem.Retryable != test.retryable || ExitCode(nil, err) != test.exitCode {
				t.Fatalf("problem = %#v exit=%d", problem, ExitCode(nil, err))
			}
		})
	}
}

func TestRemoteValidationCandidateRejectionRetainsValidationClassification(t *testing.T) {
	err := &core.RemoteConfigValidationError{Source: core.ValidationSourceFirebase, Err: &firebase.APIError{Service: "remote-config", Operation: "validate", StatusCode: 400}}
	problem := Classify(err)
	details, ok := problem.Details.(struct {
		Kind   string `json:"kind"`
		Source string `json:"source"`
	})
	if problem.Code != "remote_config.validation_failed" || problem.Category != "validation" || ExitCode(nil, err) != 8 || !ok || details.Kind != "validation" || details.Source != core.ValidationSourceFirebase {
		t.Fatalf("problem = %#v exit=%d", problem, ExitCode(nil, err))
	}
}

func TestProjectLookupErrorUsesTypedSelectionDetails(t *testing.T) {
	err := &core.ProjectLookupError{Query: "=missing", Err: errors.New("no projects matched")}
	problem := Classify(err)
	details, ok := problem.Details.(SelectionDetails)
	if problem.Code != "project.not_found" || problem.Category != "not_found" || ExitCode(nil, err) != 6 || !ok || details.Kind != "selection" || details.Resource != "project" || details.Query != "=missing" || len(details.Candidates) != 0 {
		t.Fatalf("problem = %#v exit=%d", problem, ExitCode(nil, err))
	}
}

func TestOAuthInteractionHasSafeHumanRemediation(t *testing.T) {
	err := &core.OAuthInteractionError{AuthID: "personal", Err: firebase.ErrOAuthInteractionRequired}
	problem := Classify(err)
	if problem.Code != "interaction.required" || problem.Category != "interaction" || problem.Target == nil || *problem.Target != "personal" || ExitCode(nil, err) != 10 {
		t.Fatalf("problem = %#v", problem)
	}
	if len(problem.Remediation) != 1 || problem.Remediation[0].Strategy != shared.RemediationRunCommand || !slices.Equal(problem.Remediation[0].Argv, []string{"auth", "login", "personal"}) {
		t.Fatalf("remediation = %#v", problem.Remediation)
	}
}

func TestAuthenticationFailuresRemainTypedAndTargeted(t *testing.T) {
	tests := []struct {
		name      string
		failure   *firebase.AuthenticationError
		wantCode  string
		wantCat   string
		wantRetry bool
		wantExit  int
	}{
		{
			name: "invalid stored credentials", failure: &firebase.AuthenticationError{
				Kind: firebase.AuthenticationCredentialsInvalid, AuthType: "service-account", Operation: "load_credentials", Err: errors.New("invalid private key"),
			}, wantCode: "auth.credentials_invalid", wantCat: "auth", wantExit: 4,
		},
		{
			name: "identity provider unavailable", failure: &firebase.AuthenticationError{
				Kind: firebase.AuthenticationRequestFailed, AuthType: "oauth", Operation: "refresh_token", HTTPStatus: 503, RemoteCode: "temporarily_unavailable", Retryable: true, Err: errors.New("token endpoint unavailable"),
			}, wantCode: "network.unavailable", wantCat: "unavailable", wantRetry: true, wantExit: 11,
		},
		{
			name: "gcloud setup required", failure: &firebase.AuthenticationError{
				Kind: firebase.AuthenticationSetupRequired, AuthType: "gcloud", Operation: "discover_credentials", Err: errors.New("default credentials not found"),
			}, wantCode: "auth.setup_required", wantCat: "auth", wantExit: 4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := &core.AuthError{Kind: test.failure.Kind, AuthID: "personal", Err: test.failure}
			problem := Classify(err)
			if problem.Code != test.wantCode || problem.Category != test.wantCat || problem.Retryable != test.wantRetry || problem.Target == nil || *problem.Target != "personal" || ExitCode(nil, err) != test.wantExit {
				t.Fatalf("problem = %#v exit=%d", problem, ExitCode(nil, err))
			}
			if test.failure.HTTPStatus > 0 {
				raw, marshalErr := json.Marshal(problem.Details)
				if marshalErr != nil || !strings.Contains(string(raw), `"kind":"remote_api"`) || !strings.Contains(string(raw), `"service":"google_auth"`) {
					t.Fatalf("details = %s, %v", raw, marshalErr)
				}
			}
		})
	}
}

func TestQuotaProjectFailuresUseExistingTypedAuthProblems(t *testing.T) {
	for _, test := range []struct {
		name     string
		source   firebase.QuotaProjectSource
		authKind string
		wantCode string
	}{
		{
			name:     "environment configuration",
			source:   firebase.QuotaProjectSourceEnvironment,
			authKind: "configuration",
			wantCode: "auth.configuration_invalid",
		},
		{
			name:     "ADC credential configuration",
			source:   firebase.QuotaProjectSourceCredentials,
			authKind: firebase.AuthenticationCredentialsInvalid,
			wantCode: "auth.credentials_invalid",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cause := &firebase.QuotaProjectError{Source: test.source, Err: errors.New("invalid quota project")}
			var inner error = cause
			if test.source == firebase.QuotaProjectSourceCredentials {
				inner = &firebase.AuthenticationError{Kind: test.authKind, AuthType: "gcloud", Operation: "load_credentials", Err: cause}
			}
			err := &core.AuthError{Kind: test.authKind, AuthID: "personal", Err: inner}
			problem := Classify(err)
			if problem.Code != test.wantCode || problem.Category != "auth" || problem.Retryable || problem.Target == nil || *problem.Target != "personal" || ExitCode(nil, err) != 4 {
				t.Fatalf("problem = %#v exit=%d", problem, ExitCode(nil, err))
			}
		})
	}
}

func TestInteractionProblemSuggestsNonInteractiveConfirmation(t *testing.T) {
	problem := Classify(shared.InteractionRequired("confirmation is required", true, "--yes"))
	if problem.Category != "interaction" || problem.Retryable || len(problem.Remediation) != 1 || problem.Remediation[0].Strategy != shared.RemediationRetryWithArguments || len(problem.Remediation[0].Argv) != 1 || problem.Remediation[0].Argv[0] != "--yes" {
		t.Fatalf("problem = %#v", problem)
	}
}

func TestInteractionProblemDoesNotPublishIncompleteInputArguments(t *testing.T) {
	problem := Classify(shared.InteractionRequiredWithArguments("input requires --from or stdin", "external_input", false, "--from"))
	if len(problem.Remediation) != 0 {
		t.Fatalf("incomplete input remediation = %#v", problem.Remediation)
	}
	raw, err := json.Marshal(problem.Details)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"interaction_type":"external_input"`) || !strings.Contains(string(raw), `"required_option":"--from"`) {
		t.Fatalf("interaction details = %s", raw)
	}
}

func TestDraftConflictPreservesPublishAndDiscardRemediation(t *testing.T) {
	problem := Classify(&shared.ConflictError{
		Code: "draft.exists", Resource: "draft", Target: "demo", Err: errors.New("draft exists"),
		Remediation: []shared.Remediation{
			{Description: "publish draft", Strategy: shared.RemediationRunCommand, Argv: []string{"draft", "publish", "demo"}},
			{Description: "discard draft", Strategy: shared.RemediationRunCommand, Argv: []string{"draft", "discard", "demo"}},
		},
	})
	if problem.Category != "conflict" || problem.Code != "draft.exists" || len(problem.Remediation) != 2 || problem.Remediation[0].Argv[1] != "publish" || problem.Remediation[1].Argv[1] != "discard" {
		t.Fatalf("problem = %#v", problem)
	}
}

func TestBuildEnvelopeFailureHasNullData(t *testing.T) {
	cmd := commandForTest("show")
	envelope := BuildEnvelope(cmd, "test", nil, &shared.SelectionError{Resource: "resource", Kind: "not_found", Query: "missing"})
	if envelope.Outcome != "failure" || envelope.Data != nil || envelope.ExitCode != 6 || len(envelope.Errors) != 1 {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestBuildEnvelopePreservesProfilelessMachineContext(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.Profile, "")
	if err := config.SetProfileOverride(""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.SetProfileOverride("") })

	cmd := commandForTest("show")
	cmd.SetContext(machine.WithProfileless(shared.WithMachineState(context.Background())))
	envelope := BuildEnvelope(cmd, "test", nil, &shared.SelectionError{Resource: "resource", Kind: "not_found", Query: "missing"})
	if envelope.Context.Profile != nil {
		t.Fatalf("context profile = %q, want null", *envelope.Context.Profile)
	}
	if _, err := os.Stat(filepath.Join(root, "config")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config root stat = %v, want envelope construction without profile filesystem access", err)
	}
}

func TestBuildEnvelopeUsesClassifiedExitForPreExecutionErrors(t *testing.T) {
	cmd := commandForTest("import")
	envelope := BuildEnvelope(cmd, "test", nil, errors.New("required flag(s) from not set"))
	if envelope.Outcome != "failure" || envelope.ExitCode != 2 || len(envelope.Errors) != 1 || envelope.Errors[0].Category != "argument" {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestProblemCodeCatalogIsSortedUniqueAndSyntacticallyValid(t *testing.T) {
	for name, codes := range map[string][]string{"problem": KnownProblemCodes(), "warning": KnownWarningCodes()} {
		if !slices.IsSorted(codes) {
			t.Fatalf("%s codes are not sorted: %v", name, codes)
		}
		if len(slices.Compact(append([]string(nil), codes...))) != len(codes) {
			t.Fatalf("%s codes contain duplicates: %v", name, codes)
		}
		for _, code := range codes {
			parts := strings.Split(code, ".")
			if len(parts) < 2 || strings.ContainsAny(code, " -/") {
				t.Errorf("invalid %s code %q", name, code)
			}
		}
	}
}

func TestBuildEnvelopePartialSuccessKeepsTypedData(t *testing.T) {
	cmd := commandForTest("publish")
	envelope := BuildEnvelope(cmd, "test", []byte(`{"succeeded":1,"failed":1}`), &shared.BatchError{FailedTargets: []string{"failed"}, SuccessfulTargetCount: 1})
	if envelope.Outcome != "partial_success" || envelope.ExitCode != 12 || envelope.Data == nil {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestClassifierNeverInfersSemanticsFromErrorWording(t *testing.T) {
	err := errors.New("one or more template targets failed because the argument is invalid")
	problem := Classify(err)
	if problem.Code != "internal.unclassified" || problem.Category != "internal" || ExitCode(nil, err) != 15 {
		t.Fatalf("problem = %#v", problem)
	}
}

func TestBuildEnvelopeCollectsStructuredWarningsAndRemediation(t *testing.T) {
	cmd := commandForTest("publish")
	cmd.SetContext(shared.WithMachineState(context.Background()))
	shared.AddMachineWarning(cmd, shared.MachineWarning{
		Code:    "publication.cache_stale",
		Message: "published, but cache refresh failed",
		Target:  "demo",
		Details: map[string]any{"stage": "cache"},
		Remediation: []shared.Remediation{{
			Description: "refresh the stale target",
			Strategy:    shared.RemediationRunCommand,
			Argv:        []string{"get", "--update", "--project", "=demo"},
		}},
	})
	envelope := BuildEnvelope(cmd, "test", []byte(`{"published":true}`), nil)
	if len(envelope.Warnings) != 1 || envelope.Warnings[0].Code != "publication.cache_stale" || envelope.Warnings[0].Target == nil || *envelope.Warnings[0].Target != "demo" || len(envelope.Warnings[0].Remediation) != 1 {
		t.Fatalf("warnings = %#v", envelope.Warnings)
	}
	if got := envelope.Warnings[0].Remediation[0]; got.Strategy != shared.RemediationRunCommand || len(got.Argv) != 4 || got.Argv[0] != "get" || got.Argv[3] != "=demo" {
		t.Fatalf("remediation = %#v", got)
	}
}

func TestPublishedPostProcessingFailuresArePartialSuccess(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{"cache", &core.RemoteConfigPublishedCacheError{ProjectID: "demo", Err: errors.New("disk full")}, "publication.cache_failed"},
		{"hook", &core.RemoteConfigPublishedHookError{ProjectID: "demo", HookErr: &corehooks.Error{Event: corehooks.PostPublish, Output: "failed", Err: errors.New("exit 1")}}, "publication.hook_failed"},
		{"hook and cache", &core.RemoteConfigPublishedHookError{ProjectID: "demo", HookErr: &corehooks.Error{Event: corehooks.PostPublish, Output: "failed", Err: errors.New("exit 1")}, CacheErr: errors.New("disk full")}, "publication.hook_failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := BuildEnvelope(commandForTest("publish"), "test", []byte(`{"published":true}`), test.err)
			if envelope.Outcome != "partial_success" || envelope.ExitCode != 12 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != test.code || envelope.Errors[0].Retryable {
				t.Fatalf("envelope = %#v", envelope)
			}
		})
	}
}

func TestMachineErrorsRedactSecretsAndBoundMessagesAndHookOutput(t *testing.T) {
	secret := "super-secret-value"
	long := strings.Repeat("x", 5000)
	problem := Classify(errors.New(`request failed: {"access_token":"` + secret + `"} bearer ` + secret + ` token=` + secret + ` ` + long))
	if strings.Contains(problem.Message, secret) || len([]rune(problem.Message)) != 4097 || !strings.HasSuffix(problem.Message, "…") {
		t.Fatalf("unsafe bounded message: length=%d message=%q", len([]rune(problem.Message)), problem.Message)
	}

	hookProblem := Classify(&corehooks.Error{Event: corehooks.PrePublish, Output: `password=` + secret + ` ` + long, Err: errors.New("hook failed")})
	raw, err := json.Marshal(hookProblem.Details)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), secret) || !strings.Contains(string(raw), "[REDACTED]") || len([]rune(string(raw))) > 4300 {
		t.Fatalf("unsafe hook details: %s", raw)
	}
}

func TestBuildEnvelopeDiffIsSuccessfulSemanticResult(t *testing.T) {
	root := &cobra.Command{Use: "fbrcm"}
	projects := &cobra.Command{Use: "projects"}
	diff := &cobra.Command{Use: "diff"}
	root.AddCommand(projects)
	projects.AddCommand(diff)
	envelope := BuildEnvelope(diff, "test", []byte(`{"changed":true}`), shared.WithExitCode(nil, 1))
	if envelope.Outcome != "success" || envelope.ExitCode != 1 || len(envelope.Errors) != 0 {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestBuildEnvelopeValidationNegativeIsFailure(t *testing.T) {
	root := &cobra.Command{Use: "fbrcm"}
	configCmd := &cobra.Command{Use: "config"}
	validate := &cobra.Command{Use: "validate"}
	root.AddCommand(configCmd)
	configCmd.AddCommand(validate)
	envelope := BuildEnvelope(validate, "test", []byte(`{"valid":false}`), shared.WithExitCode(nil, 1))
	if envelope.Outcome != "failure" || envelope.ExitCode != 1 || len(envelope.Errors) != 1 || envelope.Errors[0].Code != "validation.failed" {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestBuildEnvelopeWrapsTopLevelArrays(t *testing.T) {
	envelope := BuildEnvelope(commandForTest("list"), "test", []byte(`[1,2]`), nil)
	raw, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"count":2,"items":[1,2]}` {
		t.Fatalf("data = %s", raw)
	}
}

func TestArtifactEncodesBinaryContent(t *testing.T) {
	artifact := NewArtifact(nil, "application/octet-stream", []byte{0xff, 0x00}, nil, false)
	if artifact.Encoding != "base64" || artifact.Base64 == nil || *artifact.Base64 != "/wA=" || artifact.SizeBytes != 2 {
		t.Fatalf("artifact = %#v", artifact)
	}
}

func TestJSONArtifactDigestDescribesEmbeddedCompactToken(t *testing.T) {
	artifact := NewArtifact(nil, "application/json", []byte("{\n  \"html\": \"<tag>\"\n}"), nil, false)
	want := []byte(`{"html":"\u003ctag\u003e"}`)
	if artifact.Encoding != "json" || string(artifact.JSONContent) != string(want) {
		t.Fatalf("artifact JSON = %q", artifact.JSONContent)
	}
	sum := sha256.Sum256(want)
	if artifact.SizeBytes != int64(len(want)) || artifact.SHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("artifact digest = %d/%s", artifact.SizeBytes, artifact.SHA256)
	}
}

func TestJSONRequested(t *testing.T) {
	if !JSONRequested([]string{"show", "--json"}) || !JSONRequested([]string{"--json=true"}) || !JSONRequested([]string{"--json=false", "--json"}) || JSONRequested([]string{"--json", "--json=false"}) || JSONRequested([]string{"--", "--json"}) {
		t.Fatal("JSONRequested did not honor global flag syntax")
	}
}

func commandForTest(use string) *cobra.Command {
	root := &cobra.Command{Use: "fbrcm"}
	cmd := &cobra.Command{Use: use}
	root.AddCommand(cmd)
	return cmd
}
