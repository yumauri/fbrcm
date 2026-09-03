package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"github.com/yumauri/fbrcm/e2e/internal/harness"
)

type capabilityIndex struct {
	Commands []struct {
		ID              string `json:"id"`
		SideEffectLevel int    `json:"side_effect_level"`
		Destructive     bool   `json:"destructive"`
		Supports        struct {
			Stateless bool `json:"stateless"`
		} `json:"supports"`
	} `json:"commands"`
}

type readCoverageConfig struct {
	AdditionalReadCommands []string          `json:"additional_read_commands"`
	Excluded               map[string]string `json:"excluded"`
}

func TestReadCommandCoverage(t *testing.T) {
	e2eRoot, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	suite, err := harness.LoadSuite(filepath.Join(e2eRoot, "testdata", "suite.json"))
	if err != nil {
		t.Fatal(err)
	}
	scenarios, err := harness.LoadScenarios(filepath.Join(e2eRoot, "testdata", "scenarios"), suite)
	if err != nil {
		t.Fatal(err)
	}
	var capabilities capabilityIndex
	readJSONTest(t, filepath.Join(e2eRoot, "..", "cli", "app", "testdata", "contract_v1_capabilities.golden.json"), &capabilities)
	var config readCoverageConfig
	readJSONTest(t, filepath.Join(e2eRoot, "testdata", "read_coverage.json"), &config)

	known := make(map[string]bool, len(capabilities.Commands))
	targets := make(map[string]bool)
	for _, capability := range capabilities.Commands {
		known[capability.ID] = true
		if capability.SideEffectLevel <= 2 && !capability.Destructive {
			targets[capability.ID] = true
		}
	}
	for _, id := range config.AdditionalReadCommands {
		if !known[id] {
			t.Errorf("additional read command %q is absent from capability metadata", id)
		}
		targets[id] = true
	}
	covered := make(map[string]bool)
	for _, scenario := range scenarios {
		if !known[scenario.CommandID] {
			t.Errorf("scenario %s has unknown command_id %q", scenario.Name, scenario.CommandID)
		}
		covered[scenario.CommandID] = true
	}
	for id, reason := range config.Excluded {
		if reason == "" {
			t.Errorf("read coverage exclusion %q has no reason", id)
		}
		if !targets[id] {
			t.Errorf("read coverage exclusion %q is stale or not a read target", id)
		}
		if covered[id] {
			t.Errorf("read coverage exclusion %q is stale because a scenario now covers it", id)
		}
	}
	var missing []string
	for id := range targets {
		if !covered[id] {
			if _, excluded := config.Excluded[id]; !excluded {
				missing = append(missing, id)
			}
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("read commands need an E2E scenario or a reasoned exclusion: %v", missing)
	}
}

func TestCoreParameterMutationDryRunCoverage(t *testing.T) {
	e2eRoot, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	suite, err := harness.LoadSuite(filepath.Join(e2eRoot, "testdata", "suite.json"))
	if err != nil {
		t.Fatal(err)
	}
	scenarios, err := harness.LoadScenarios(filepath.Join(e2eRoot, "testdata", "scenarios"), suite)
	if err != nil {
		t.Fatal(err)
	}

	required := map[string]bool{
		"add":       false,
		"update":    false,
		"duplicate": false,
		"delete":    false,
	}
	for _, scenario := range scenarios {
		if _, ok := required[scenario.CommandID]; !ok {
			continue
		}
		if !slices.Contains(scenario.Args, "--dry-run") {
			continue
		}
		if !scenario.JSONOutput || scenario.ExpectedExitCode != 0 {
			t.Errorf("scenario %s must assert successful JSON output", scenario.Name)
		}
		validationRequests := 0
		for _, request := range scenario.ExpectedHTTP {
			switch request.Method {
			case "GET":
			case "PUT":
				if request.Query != "validateOnly=true" {
					t.Errorf("scenario %s permits non-validation PUT query %q", scenario.Name, request.Query)
				}
				validationRequests++
			default:
				t.Errorf("scenario %s permits unsafe HTTP method %s", scenario.Name, request.Method)
			}
		}
		if validationRequests != 1 {
			t.Errorf("scenario %s has %d validation requests, want 1", scenario.Name, validationRequests)
		}
		required[scenario.CommandID] = true
	}

	var missing []string
	for id, covered := range required {
		if !covered {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("core parameter commands need safe dry-run E2E coverage: %v", missing)
	}
}

func TestStatelessCommandCoverage(t *testing.T) {
	e2eRoot, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	suite, err := harness.LoadSuite(filepath.Join(e2eRoot, "testdata", "suite.json"))
	if err != nil {
		t.Fatal(err)
	}
	scenarios, err := harness.LoadScenarios(filepath.Join(e2eRoot, "testdata", "scenarios"), suite)
	if err != nil {
		t.Fatal(err)
	}
	var capabilities capabilityIndex
	readJSONTest(t, filepath.Join(e2eRoot, "..", "cli", "app", "testdata", "contract_v1_capabilities.golden.json"), &capabilities)

	supported := make(map[string]bool)
	for _, capability := range capabilities.Commands {
		if capability.Supports.Stateless {
			supported[capability.ID] = true
		}
	}
	covered := make(map[string]bool)
	for _, scenario := range scenarios {
		if !slices.Contains(scenario.Args, "--stateless") {
			continue
		}
		if scenario.CommandID == "mcp" && !slices.Contains(scenario.Args, "--json") {
			// Streaming MCP launch coverage is outside the one-shot CLI JSON
			// supports.stateless inventory.
			continue
		}
		if !supported[scenario.CommandID] {
			t.Errorf("scenario %s uses --stateless for unsupported command %q", scenario.Name, scenario.CommandID)
			continue
		}
		covered[scenario.CommandID] = true
	}

	var missing []string
	for id := range supported {
		if !covered[id] {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("commands advertising supports.stateless need a stateless E2E scenario: %v", missing)
	}
}

func readJSONTest(t *testing.T, path string, target any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatal(fmt.Errorf("decode %s: %w", path, err))
	}
}
