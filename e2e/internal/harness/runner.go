package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RunOptions struct {
	Mode              Mode
	CLI               Command
	HoverflyBinary    string
	GuardPath         string
	CertificatePath   string
	KeyPath           string
	FixturesRoot      string
	StateFixturesRoot string
	SchemasRoot       string
	AccessToken       string
	ScenarioRoot      string
	ToolsRoot         string
	SkipUpstreamTLS   bool
}

type RunReport struct {
	Result  Result
	Changes []SnapshotChange
}

func RunScenario(ctx context.Context, scenario Scenario, suite Suite, options RunOptions) (RunReport, error) {
	cassettePath := filepath.Join(scenario.Directory, "http.json")
	_, cassetteErr := os.Stat(cassettePath)
	cassetteExists := cassetteErr == nil
	if cassetteErr != nil && !os.IsNotExist(cassetteErr) {
		return RunReport{}, fmt.Errorf("stat cassette: %w", cassetteErr)
	}
	hasHTTP := len(scenario.ExpectedHTTP) > 0
	capture := hasHTTP && !scenario.HTTPReplayOnly && options.Mode.Capture(cassetteExists)
	if hasHTTP && scenario.HTTPReplayOnly && !cassetteExists {
		return RunReport{}, fmt.Errorf("replay-only cassette %s is missing and must be provided as a synthetic fixture", cassettePath)
	}
	if hasHTTP && !capture && !cassetteExists {
		return RunReport{}, fmt.Errorf("cassette %s is missing; use -mode=record-missing", cassettePath)
	}
	if capture && strings.TrimSpace(options.AccessToken) == "" {
		return RunReport{}, fmt.Errorf("FBRCM_E2E_ACCESS_TOKEN is required to capture HTTP traffic")
	}

	proxyDirectory := filepath.Join(options.ScenarioRoot, "proxy")
	if err := os.MkdirAll(proxyDirectory, 0o700); err != nil {
		return RunReport{}, fmt.Errorf("create proxy directory: %w", err)
	}
	proxyCassettePath := cassettePath
	if !hasHTTP {
		proxyCassettePath = filepath.Join(proxyDirectory, "empty-http.json")
		if err := atomicWrite(proxyCassettePath, emptySimulation(), 0o600); err != nil {
			return RunReport{}, err
		}
	}
	proxy, err := StartHoverfly(ctx, HoverflyOptions{
		Binary:          options.HoverflyBinary,
		Directory:       proxyDirectory,
		CertificatePath: options.CertificatePath,
		KeyPath:         options.KeyPath,
		GuardPath:       options.GuardPath,
		AllowedRequests: scenario.ExpectedHTTP,
		CassettePath:    proxyCassettePath,
		Capture:         capture,
		SkipUpstreamTLS: options.SkipUpstreamTLS,
	})
	if err != nil {
		return RunReport{}, err
	}
	defer func() { _ = proxy.Stop() }()

	terminalWidth := scenario.TerminalWidth
	if terminalWidth == 0 {
		terminalWidth = suite.DefaultTerminalWidth
	}
	logLevel := scenario.LogLevel
	if logLevel == "" {
		logLevel = suite.DefaultLogLevel
	}
	environment, err := PrepareEnvironment(
		filepath.Join(options.ScenarioRoot, "environment"),
		options.FixturesRoot,
		suite,
		proxy.ProxyURL(),
		options.CertificatePath,
		options.AccessToken,
		terminalWidth,
		logLevel,
		scenario.LocalConfig,
		scenario.Environment,
	)
	if err != nil {
		return RunReport{}, err
	}
	if scenario.Fixture != "" {
		if err := ApplyStateFixture(environment, filepath.Join(options.StateFixturesRoot, scenario.Fixture)); err != nil {
			return RunReport{}, err
		}
	}
	if scenario.Offline {
		values := environmentMap(environment.Variables)
		values["FBRCM_OFFLINE"] = "1"
		environment.Variables = flattenEnvironment(values)
	}
	result := Run(ctx, options.CLI, scenario.Args, environment.Variables, environment.WorkDir, nil)
	journal, err := proxy.Journal(ctx)
	if err != nil {
		return RunReport{}, err
	}
	if err := validateJournal(journal, scenario.ExpectedHTTP, capture, scenario.HTTPUnordered); err != nil {
		return RunReport{Result: result}, fmt.Errorf(
			"%w\nexit code: %d\nstdout:\n%s\nstderr:\n%s",
			err,
			result.ExitCode,
			redactDiagnosticOutput(result.Stdout, options.AccessToken),
			redactDiagnosticOutput(result.Stderr, options.AccessToken),
		)
	}
	if result.ExitCode != scenario.ExpectedExitCode {
		return RunReport{Result: result}, fmt.Errorf(
			"exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s",
			result.ExitCode,
			scenario.ExpectedExitCode,
			redactDiagnosticOutput(result.Stdout, options.AccessToken),
			redactDiagnosticOutput(result.Stderr, options.AccessToken),
		)
	}

	report := RunReport{Result: result}
	if capture {
		simulation, err := proxy.ExportSimulation(ctx, options.AccessToken)
		if err != nil {
			return report, err
		}
		if err := atomicWrite(cassettePath, simulation, 0o644); err != nil {
			return report, err
		}
		report.Changes = append(report.Changes, SnapshotChange{Path: cassettePath, Created: !cassetteExists, Updated: cassetteExists})
	}
	if err := proxy.Stop(); err != nil {
		return report, err
	}

	if scenario.JSONOutput {
		if !json.Valid(result.Stdout) {
			return report, fmt.Errorf("stdout is not valid JSON:\n%s", result.Stdout)
		}
		if err := ValidateCommandOutput(options.SchemasRoot, scenario.CommandID, result.Stdout); err != nil {
			return report, err
		}
	}
	replacements := runtimeSnapshotReplacements(options.ScenarioRoot, options.ToolsRoot, proxy.ProxyURL())
	stateReplacements, err := expectedStateFileSnapshotReplacements(scenario, environment)
	if err != nil {
		return report, err
	}
	replacements = append(replacements, stateReplacements...)
	if strings.HasPrefix(scenario.CommandID, "hooks.") {
		replacements = append(replacements, hookFingerprintReplacements(result.Stdout, result.Stderr)...)
	}
	for _, output := range []struct {
		name string
		raw  []byte
	}{
		{"stdout.golden", CanonicalizeSnapshot(result.Stdout, replacements...)},
		{"stderr.golden", CanonicalizeSnapshot(result.Stderr, replacements...)},
	} {
		path := filepath.Join(scenario.Directory, output.name)
		_, snapshotErr := os.Stat(path)
		exists := snapshotErr == nil
		if snapshotErr != nil && !os.IsNotExist(snapshotErr) {
			return report, fmt.Errorf("stat snapshot %s: %w", path, snapshotErr)
		}
		change, err := CheckSnapshot(path, output.raw, options.Mode.UpdateOutput(exists))
		if err != nil {
			return report, err
		}
		if change.Created || change.Updated {
			report.Changes = append(report.Changes, change)
		}
	}
	fileChanges, err := checkExpectedFileSnapshots(scenario, environment.WorkDir, options.Mode)
	if err != nil {
		return report, err
	}
	report.Changes = append(report.Changes, fileChanges...)
	stateChanges, err := checkExpectedStateFileSnapshots(scenario, environment, options.Mode, replacements)
	if err != nil {
		return report, err
	}
	report.Changes = append(report.Changes, stateChanges...)
	if err := checkExpectedAbsentStatePaths(scenario, environment); err != nil {
		return report, err
	}
	return report, nil
}

func checkExpectedFileSnapshots(scenario Scenario, workDir string, mode Mode) ([]SnapshotChange, error) {
	changes := make([]SnapshotChange, 0, len(scenario.ExpectedFiles))
	for _, relative := range scenario.ExpectedFiles {
		actualPath := filepath.Join(workDir, relative)
		info, err := os.Lstat(actualPath)
		if err != nil {
			return nil, fmt.Errorf("inspect expected file %s: %w", relative, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("expected file %s is not a regular file", relative)
		}
		raw, err := os.ReadFile(actualPath)
		if err != nil {
			return nil, fmt.Errorf("read expected file %s: %w", relative, err)
		}
		snapshotPath := filepath.Join(scenario.Directory, "files", relative+".golden")
		_, snapshotErr := os.Stat(snapshotPath)
		exists := snapshotErr == nil
		if snapshotErr != nil && !os.IsNotExist(snapshotErr) {
			return nil, fmt.Errorf("stat file snapshot %s: %w", snapshotPath, snapshotErr)
		}
		change, err := CheckSnapshot(snapshotPath, raw, mode.UpdateOutput(exists))
		if err != nil {
			return nil, err
		}
		if change.Created || change.Updated {
			changes = append(changes, change)
		}
	}
	return changes, nil
}

var hookFingerprintPattern = regexp.MustCompile(`\b[0-9a-f]{64}\b`)
var bearerCredentialPattern = regexp.MustCompile(`(?i)Bearer\s+[^\s"\]]+`)

func redactDiagnosticOutput(raw []byte, secrets ...string) []byte {
	redacted := append([]byte(nil), raw...)
	for _, secret := range secrets {
		if strings.TrimSpace(secret) != "" {
			redacted = bytes.ReplaceAll(redacted, []byte(secret), []byte("<REDACTED>"))
		}
	}
	return bearerCredentialPattern.ReplaceAll(redacted, []byte("Bearer <REDACTED>"))
}

func hookFingerprintReplacements(outputs ...[]byte) []SnapshotReplacement {
	seen := make(map[string]bool)
	var replacements []SnapshotReplacement
	for _, output := range outputs {
		for _, match := range hookFingerprintPattern.FindAll(output, -1) {
			value := string(match)
			if seen[value] {
				continue
			}
			seen[value] = true
			replacements = append(replacements, SnapshotReplacement{Old: value, New: "<E2E_HOOK_FINGERPRINT>"})
		}
	}
	return replacements
}

func runtimeSnapshotReplacements(scenarioRoot, toolsRoot, proxyURL string) []SnapshotReplacement {
	replacements := make([]SnapshotReplacement, 0, 6)
	replacements = appendPathReplacements(replacements, scenarioRoot, "<E2E_RUN_ROOT>")
	replacements = appendPathReplacements(replacements, toolsRoot, "<E2E_TOOLS_ROOT>")
	proxyAddress := strings.TrimPrefix(proxyURL, "http://")
	if proxyURL != "" {
		replacements = append(replacements, SnapshotReplacement{Old: proxyURL, New: "<E2E_PROXY_URL>"})
	}
	if proxyAddress != "" && proxyAddress != proxyURL {
		replacements = append(replacements, SnapshotReplacement{Old: proxyAddress, New: "<E2E_PROXY_ADDRESS>"})
	}
	return replacements
}

func appendPathReplacements(replacements []SnapshotReplacement, path, placeholder string) []SnapshotReplacement {
	if strings.TrimSpace(path) == "" {
		return replacements
	}
	cleaned := filepath.Clean(path)
	replacements = append(replacements, SnapshotReplacement{Old: cleaned, New: placeholder})
	evaluated, err := filepath.EvalSymlinks(cleaned)
	if err == nil && evaluated != cleaned {
		replacements = append(replacements, SnapshotReplacement{Old: evaluated, New: placeholder})
	}
	return replacements
}

func validateJournal(journal Journal, expected []HTTPExpectation, capture, unordered bool) error {
	if journal.Total != len(expected) || len(journal.Entries) != len(expected) {
		return fmt.Errorf("Hoverfly journal contains %d requests (total %d), want %d", len(journal.Entries), journal.Total, len(expected))
	}
	wantMode := "simulate"
	if capture {
		wantMode = "capture"
	}
	if unordered {
		remaining := append([]HTTPExpectation(nil), expected...)
		for index, entry := range journal.Entries {
			matched := -1
			matchedQuery := false
			for candidateIndex, candidate := range remaining {
				if !journalEntryMatches(entry, candidate, wantMode) {
					continue
				}
				exactQuery := candidate.Query != ""
				if matched == -1 || exactQuery && !matchedQuery {
					matched = candidateIndex
					matchedQuery = exactQuery
				}
			}
			if matched == -1 {
				return fmt.Errorf("request %d (%s %s%s, status %d, mode %s) has no matching unordered expectation", index+1, entry.Request.Method, entry.Request.Destination, entry.Request.Path, entry.Response.Status, entry.Mode)
			}
			remaining = append(remaining[:matched], remaining[matched+1:]...)
		}
		return nil
	}
	for index, entry := range journal.Entries {
		want := expected[index]
		if entry.Mode != wantMode {
			return fmt.Errorf("request %d mode = %q, want %q", index+1, entry.Mode, wantMode)
		}
		if entry.Request.Method != want.Method {
			return fmt.Errorf("request %d method = %q, want %q", index+1, entry.Request.Method, want.Method)
		}
		if entry.Request.Destination != want.Host && entry.Request.Destination != want.Host+":443" {
			return fmt.Errorf("request %d destination = %q, want %q", index+1, entry.Request.Destination, want.Host)
		}
		if entry.Request.Path != want.Path {
			return fmt.Errorf("request %d path = %q, want %q", index+1, entry.Request.Path, want.Path)
		}
		if want.Query != "" && entry.Request.Query != want.Query {
			return fmt.Errorf("request %d query = %q, want %q", index+1, entry.Request.Query, want.Query)
		}
		if got := headerValue(entry.Request.Headers, "X-Goog-User-Project"); got != want.QuotaProjectID {
			return fmt.Errorf("request %d X-Goog-User-Project = %q, want %q", index+1, got, want.QuotaProjectID)
		}
		if entry.Response.Status != want.Status {
			return fmt.Errorf("request %d status = %d, want %d", index+1, entry.Response.Status, want.Status)
		}
	}
	return nil
}

func journalEntryMatches(entry JournalEntry, want HTTPExpectation, wantMode string) bool {
	if entry.Mode != wantMode || entry.Request.Method != want.Method || entry.Request.Path != want.Path || entry.Response.Status != want.Status {
		return false
	}
	if entry.Request.Destination != want.Host && entry.Request.Destination != want.Host+":443" {
		return false
	}
	if want.Query != "" && entry.Request.Query != want.Query {
		return false
	}
	return headerValue(entry.Request.Headers, "X-Goog-User-Project") == want.QuotaProjectID
}

func headerValue(headers map[string][]string, name string) string {
	for key, values := range headers {
		if !strings.EqualFold(key, name) || len(values) == 0 {
			continue
		}
		return strings.TrimSpace(values[0])
	}
	return ""
}

func emptySimulation() []byte {
	return []byte(`{
  "data": {
    "globalActions": {"delays": [], "delaysLogNormal": []},
    "pairs": []
  },
  "meta": {
    "hoverflyVersion": "v1.12.10",
    "schemaVersion": "v5.3",
    "timeExported": "2024-01-01T00:00:00Z"
  }
}
`)
}
