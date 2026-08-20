package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	projectVariable         = "${PROJECT_ID}"
	defaultTerminalWidth    = 200
	defaultScenarioLogLevel = "debug"
)

// Suite identifies the stable Firebase project represented by committed cassettes.
type Suite struct {
	ProjectID            string `json:"project_id"`
	ProjectName          string `json:"project_name"`
	QuotaProjectID       string `json:"quota_project_id,omitempty"`
	DefaultTerminalWidth int    `json:"default_terminal_width,omitempty"`
	DefaultLogLevel      string `json:"default_log_level,omitempty"`
}

// Scenario describes one black-box CLI execution and its expected traffic.
type Scenario struct {
	Name                     string                 `json:"name"`
	CommandID                string                 `json:"command_id"`
	Args                     []string               `json:"args"`
	ExpectedExitCode         int                    `json:"expected_exit_code"`
	ExpectedHTTP             []HTTPExpectation      `json:"expected_http"`
	HTTPReplayOnly           bool                   `json:"http_replay_only,omitempty"`
	ExpectedFiles            []string               `json:"expected_files,omitempty"`
	ExpectedStateFiles       []StateFileExpectation `json:"expected_state_files,omitempty"`
	ExpectedAbsentStatePaths []StatePathExpectation `json:"expected_absent_state_paths,omitempty"`
	JSONOutput               bool                   `json:"json_output,omitempty"`
	Fixture                  string                 `json:"fixture,omitempty"`
	LocalConfig              bool                   `json:"local_config,omitempty"`
	Offline                  bool                   `json:"offline,omitempty"`
	TerminalWidth            int                    `json:"terminal_width,omitempty"`
	LogLevel                 string                 `json:"log_level,omitempty"`
	Directory                string                 `json:"-"`
}

// StateFileExpectation snapshots one file rooted in the isolated config or
// cache directory. JSON replacements canonicalize explicitly declared
// volatile values before comparison while preserving all other bytes.
type StateFileExpectation struct {
	Root             string            `json:"root"`
	Path             string            `json:"path"`
	JSONReplacements map[string]string `json:"json_replacements,omitempty"`
}

// StatePathExpectation requires a file or directory not to exist after the
// command completes.
type StatePathExpectation struct {
	Root string `json:"root"`
	Path string `json:"path"`
}

// HTTPExpectation is the exact observable HTTP exchange expected from a scenario.
type HTTPExpectation struct {
	Method string `json:"method"`
	Host   string `json:"host"`
	Path   string `json:"path"`
	Query  string `json:"query,omitempty"`
	Status int    `json:"status"`
}

func LoadSuite(path string) (Suite, error) {
	var suite Suite
	if err := readJSON(path, &suite); err != nil {
		return Suite{}, err
	}
	if strings.TrimSpace(suite.ProjectID) == "" {
		return Suite{}, fmt.Errorf("suite project_id is required")
	}
	if strings.TrimSpace(suite.ProjectName) == "" {
		suite.ProjectName = suite.ProjectID
	}
	suite.QuotaProjectID = strings.TrimSpace(suite.QuotaProjectID)
	if suite.DefaultTerminalWidth < 0 {
		return Suite{}, fmt.Errorf("suite default_terminal_width must be positive")
	}
	if suite.DefaultTerminalWidth == 0 {
		suite.DefaultTerminalWidth = defaultTerminalWidth
	}
	suite.DefaultLogLevel = strings.ToLower(strings.TrimSpace(suite.DefaultLogLevel))
	if suite.DefaultLogLevel == "" {
		suite.DefaultLogLevel = defaultScenarioLogLevel
	}
	if !validLogLevel(suite.DefaultLogLevel) {
		return Suite{}, fmt.Errorf("suite has unsupported default_log_level %q", suite.DefaultLogLevel)
	}
	return suite, nil
}

func LoadScenarios(root string, suite Suite) ([]Scenario, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read scenarios directory: %w", err)
	}
	scenarios := make([]Scenario, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		directory := filepath.Join(root, entry.Name())
		var scenario Scenario
		if err := readJSON(filepath.Join(directory, "scenario.json"), &scenario); err != nil {
			return nil, err
		}
		if scenario.Name == "" {
			scenario.Name = entry.Name()
		}
		if scenario.Name != entry.Name() {
			return nil, fmt.Errorf("scenario %s name %q must match its directory", entry.Name(), scenario.Name)
		}
		if len(scenario.Args) == 0 {
			return nil, fmt.Errorf("scenario %s has no arguments", scenario.Name)
		}
		if strings.TrimSpace(scenario.CommandID) == "" {
			return nil, fmt.Errorf("scenario %s command_id is required", scenario.Name)
		}
		if scenario.TerminalWidth < 0 {
			return nil, fmt.Errorf("scenario %s terminal_width must be positive", scenario.Name)
		}
		scenario.LogLevel = strings.ToLower(strings.TrimSpace(scenario.LogLevel))
		if scenario.LogLevel != "" && !validLogLevel(scenario.LogLevel) {
			return nil, fmt.Errorf("scenario %s has unsupported log_level %q", scenario.Name, scenario.LogLevel)
		}
		for index := range scenario.Args {
			scenario.Args[index] = strings.ReplaceAll(scenario.Args[index], projectVariable, suite.ProjectID)
		}
		for index := range scenario.ExpectedHTTP {
			expectation := &scenario.ExpectedHTTP[index]
			expectation.Method = strings.ToUpper(strings.TrimSpace(expectation.Method))
			expectation.Host = strings.TrimSpace(expectation.Host)
			expectation.Path = strings.ReplaceAll(strings.TrimSpace(expectation.Path), projectVariable, suite.ProjectID)
			expectation.Query = strings.TrimSpace(expectation.Query)
			if expectation.Method == "" || expectation.Host == "" || expectation.Path == "" {
				return nil, fmt.Errorf("scenario %s expected_http entry %d requires method, host, and path", scenario.Name, index+1)
			}
			if expectation.Status < 100 || expectation.Status > 599 {
				return nil, fmt.Errorf("scenario %s expected_http entry %d has invalid status %d", scenario.Name, index+1, expectation.Status)
			}
		}
		if scenario.HTTPReplayOnly && len(scenario.ExpectedHTTP) == 0 {
			return nil, fmt.Errorf("scenario %s http_replay_only requires expected_http entries", scenario.Name)
		}
		seenFiles := make(map[string]bool, len(scenario.ExpectedFiles))
		for index, path := range scenario.ExpectedFiles {
			cleaned, ok := cleanRelativeSnapshotPath(path)
			if !ok {
				return nil, fmt.Errorf("scenario %s expected_files entry %d must be a relative file path", scenario.Name, index+1)
			}
			if seenFiles[cleaned] {
				return nil, fmt.Errorf("scenario %s expected_files contains duplicate path %q", scenario.Name, cleaned)
			}
			seenFiles[cleaned] = true
			scenario.ExpectedFiles[index] = cleaned
		}
		seenStateFiles := make(map[string]bool, len(scenario.ExpectedStateFiles))
		for index := range scenario.ExpectedStateFiles {
			expectation := &scenario.ExpectedStateFiles[index]
			expectation.Root = strings.ToLower(strings.TrimSpace(expectation.Root))
			if expectation.Root != "config" && expectation.Root != "cache" {
				return nil, fmt.Errorf("scenario %s expected_state_files entry %d root must be config or cache", scenario.Name, index+1)
			}
			cleaned, ok := cleanRelativeSnapshotPath(expectation.Path)
			if !ok {
				return nil, fmt.Errorf("scenario %s expected_state_files entry %d path must be relative", scenario.Name, index+1)
			}
			expectation.Path = cleaned
			key := expectation.Root + ":" + cleaned
			if seenStateFiles[key] {
				return nil, fmt.Errorf("scenario %s expected_state_files contains duplicate %s path %q", scenario.Name, expectation.Root, cleaned)
			}
			seenStateFiles[key] = true
			for pointer, placeholder := range expectation.JSONReplacements {
				if !validJSONPointer(pointer) || strings.TrimSpace(placeholder) == "" {
					return nil, fmt.Errorf("scenario %s expected_state_files entry %d has invalid JSON replacement %q", scenario.Name, index+1, pointer)
				}
			}
		}
		seenAbsentStatePaths := make(map[string]bool, len(scenario.ExpectedAbsentStatePaths))
		for index := range scenario.ExpectedAbsentStatePaths {
			expectation := &scenario.ExpectedAbsentStatePaths[index]
			expectation.Root = strings.ToLower(strings.TrimSpace(expectation.Root))
			if expectation.Root != "config" && expectation.Root != "cache" {
				return nil, fmt.Errorf("scenario %s expected_absent_state_paths entry %d root must be config or cache", scenario.Name, index+1)
			}
			cleaned, ok := cleanRelativeSnapshotPath(expectation.Path)
			if !ok {
				return nil, fmt.Errorf("scenario %s expected_absent_state_paths entry %d path must be relative", scenario.Name, index+1)
			}
			expectation.Path = cleaned
			key := expectation.Root + ":" + cleaned
			if seenAbsentStatePaths[key] {
				return nil, fmt.Errorf("scenario %s expected_absent_state_paths contains duplicate %s path %q", scenario.Name, expectation.Root, cleaned)
			}
			if seenStateFiles[key] {
				return nil, fmt.Errorf("scenario %s expects %s state path %q to be both present and absent", scenario.Name, expectation.Root, cleaned)
			}
			seenAbsentStatePaths[key] = true
		}
		if scenario.Fixture != "" {
			cleaned := filepath.Clean(scenario.Fixture)
			if cleaned == "." || filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
				return nil, fmt.Errorf("scenario %s fixture must be a relative fixture name", scenario.Name)
			}
			scenario.Fixture = cleaned
		}
		scenario.Directory = directory
		scenarios = append(scenarios, scenario)
	}
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].Name < scenarios[j].Name })
	if len(scenarios) == 0 {
		return nil, fmt.Errorf("no scenarios found in %s", root)
	}
	return scenarios, nil
}

func cleanRelativeSnapshotPath(path string) (string, bool) {
	cleaned := filepath.Clean(strings.TrimSpace(path))
	if cleaned == "." || filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", false
	}
	return cleaned, true
}

func validJSONPointer(pointer string) bool {
	if pointer == "" {
		return true
	}
	if !strings.HasPrefix(pointer, "/") {
		return false
	}
	for index := 0; index < len(pointer); index++ {
		if pointer[index] != '~' {
			continue
		}
		if index+1 >= len(pointer) || (pointer[index+1] != '0' && pointer[index+1] != '1') {
			return false
		}
		index++
	}
	return true
}

func validLogLevel(value string) bool {
	switch value {
	case "debug", "info", "warn", "error", "fatal", "silent":
		return true
	default:
		return false
	}
}

func readJSON(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
