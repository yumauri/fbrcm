package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	projectVariable         = "${PROJECT_ID}"
	quotaProjectVariable    = "${QUOTA_PROJECT_ID}"
	e2eAccessTokenVariable  = "${FBRCM_E2E_ACCESS_TOKEN}"
	defaultTerminalWidth    = 200
	defaultScenarioLogLevel = "debug"
)

var protectedScenarioEnvironment = map[string]bool{
	"ALL_PROXY":              true,
	"FBRCM_CACHE_DIR":        true,
	"FBRCM_CONFIG_DIR":       true,
	"FBRCM_E2E_ACCESS_TOKEN": true,
	"HOME":                   true,
	"HTTPS_PROXY":            true,
	"HTTP_PROXY":             true,
	"NO_PROXY":               true,
	"SSL_CERT_FILE":          true,
}

// Suite identifies the stable Firebase project represented by committed cassettes.
type Suite struct {
	ProjectID            string              `json:"project_id"`
	ProjectName          string              `json:"project_name"`
	QuotaProjectID       string              `json:"quota_project_id,omitempty"`
	DefaultTerminalWidth int                 `json:"default_terminal_width,omitempty"`
	DefaultLogLevel      string              `json:"default_log_level,omitempty"`
	RecordingSequences   []RecordingSequence `json:"recording_sequences,omitempty"`
}

// RecordingSequence orders independent scenarios whose live Firebase effects
// establish the preconditions for the next recording in the sequence.
type RecordingSequence struct {
	Name      string   `json:"name"`
	Scenarios []string `json:"scenarios"`
}

// Scenario describes one black-box CLI execution and its expected traffic.
type Scenario struct {
	Name                     string                 `json:"name"`
	CommandID                string                 `json:"command_id"`
	Args                     []string               `json:"args"`
	ExpectedExitCode         int                    `json:"expected_exit_code"`
	ExpectedHTTP             []HTTPExpectation      `json:"expected_http"`
	HTTPUnordered            bool                   `json:"http_unordered,omitempty"`
	HTTPReplayOnly           bool                   `json:"http_replay_only,omitempty"`
	ExpectedFiles            []string               `json:"expected_files,omitempty"`
	ExpectedStateFiles       []StateFileExpectation `json:"expected_state_files,omitempty"`
	ExpectedAbsentStatePaths []StatePathExpectation `json:"expected_absent_state_paths,omitempty"`
	JSONOutput               bool                   `json:"json_output,omitempty"`
	Environment              map[string]string      `json:"envs,omitempty"`
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
	Method         string `json:"method"`
	Host           string `json:"host"`
	Path           string `json:"path"`
	Query          string `json:"query,omitempty"`
	Status         int    `json:"status"`
	QuotaProjectID string `json:"quota_project_id,omitempty"`
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
	seenSequenceNames := make(map[string]bool, len(suite.RecordingSequences))
	for sequenceIndex := range suite.RecordingSequences {
		sequence := &suite.RecordingSequences[sequenceIndex]
		sequence.Name = strings.TrimSpace(sequence.Name)
		if sequence.Name == "" {
			return Suite{}, fmt.Errorf("suite recording_sequences entry %d requires a name", sequenceIndex+1)
		}
		if seenSequenceNames[sequence.Name] {
			return Suite{}, fmt.Errorf("suite recording_sequences contains duplicate name %q", sequence.Name)
		}
		seenSequenceNames[sequence.Name] = true
		if len(sequence.Scenarios) == 0 {
			return Suite{}, fmt.Errorf("suite recording sequence %q has no scenarios", sequence.Name)
		}
		seenScenarios := make(map[string]bool, len(sequence.Scenarios))
		for scenarioIndex := range sequence.Scenarios {
			scenarioName := strings.TrimSpace(sequence.Scenarios[scenarioIndex])
			if scenarioName == "" {
				return Suite{}, fmt.Errorf("suite recording sequence %q has an empty scenario at position %d", sequence.Name, scenarioIndex+1)
			}
			if seenScenarios[scenarioName] {
				return Suite{}, fmt.Errorf("suite recording sequence %q contains duplicate scenario %q", sequence.Name, scenarioName)
			}
			seenScenarios[scenarioName] = true
			sequence.Scenarios[scenarioIndex] = scenarioName
		}
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
		for name, value := range scenario.Environment {
			if name == "" || strings.ContainsAny(name, "=\x00") {
				return nil, fmt.Errorf("scenario %s envs contains invalid variable name %q", scenario.Name, name)
			}
			if strings.ContainsRune(value, '\x00') {
				return nil, fmt.Errorf("scenario %s envs variable %q contains a null byte", scenario.Name, name)
			}
			if protectedScenarioEnvironment[strings.ToUpper(name)] {
				return nil, fmt.Errorf("scenario %s envs cannot override harness-owned variable %s", scenario.Name, name)
			}
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
			expectation.QuotaProjectID = strings.ReplaceAll(strings.TrimSpace(expectation.QuotaProjectID), quotaProjectVariable, suite.QuotaProjectID)
			if expectation.Method == "" || expectation.Host == "" || expectation.Path == "" {
				return nil, fmt.Errorf("scenario %s expected_http entry %d requires method, host, and path", scenario.Name, index+1)
			}
			if expectation.Status < 100 || expectation.Status > 599 {
				return nil, fmt.Errorf("scenario %s expected_http entry %d has invalid status %d", scenario.Name, index+1, expectation.Status)
			}
			if requiresQuotaProjectHeader(expectation.Host) {
				if expectation.QuotaProjectID == "" {
					expectation.QuotaProjectID = suite.QuotaProjectID
				}
				if expectation.QuotaProjectID == "" {
					return nil, fmt.Errorf("scenario %s expected_http entry %d requires a quota project", scenario.Name, index+1)
				}
			} else if expectation.QuotaProjectID != "" {
				return nil, fmt.Errorf("scenario %s expected_http entry %d declares a quota project for unsupported host %s", scenario.Name, index+1, expectation.Host)
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
	knownScenarios := make(map[string]Scenario, len(scenarios))
	for _, scenario := range scenarios {
		knownScenarios[scenario.Name] = scenario
	}
	for _, sequence := range suite.RecordingSequences {
		for _, scenarioName := range sequence.Scenarios {
			if _, exists := knownScenarios[scenarioName]; !exists {
				return nil, fmt.Errorf("suite recording sequence %q references unknown scenario %q", sequence.Name, scenarioName)
			}
		}
	}
	for _, sequence := range suite.RecordingSequences {
		for _, scenarioName := range sequence.Scenarios {
			scenario := knownScenarios[scenarioName]
			if len(scenario.ExpectedHTTP) == 0 || scenario.HTTPReplayOnly {
				return nil, fmt.Errorf("suite recording sequence %q scenario %q must declare live HTTP traffic", sequence.Name, scenarioName)
			}
		}
	}
	sequencedScenarios := make(map[string]bool)
	for _, sequence := range suite.RecordingSequences {
		for _, scenarioName := range sequence.Scenarios {
			sequencedScenarios[scenarioName] = true
		}
	}
	for _, scenario := range scenarios {
		if requiresRecordingSequence(scenario) && !sequencedScenarios[scenario.Name] {
			return nil, fmt.Errorf("scenario %q declares mutating live HTTP traffic but does not belong to a recording sequence", scenario.Name)
		}
	}
	return scenarios, nil
}

func requiresQuotaProjectHeader(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ":443")
	return host == "cloudresourcemanager.googleapis.com" || host == "firebaseremoteconfig.googleapis.com"
}

// OrderScenariosForMode makes every declared recording sequence a mandatory,
// contiguous block in HTTP capture modes. Replay and output-only modes retain
// the normal independent alphabetical order.
func OrderScenariosForMode(scenarios []Scenario, suite Suite, mode Mode) ([]Scenario, error) {
	if mode == ModeReplay || mode == ModeUpdateOutput {
		return scenarios, nil
	}
	byName := make(map[string]Scenario, len(scenarios))
	for _, scenario := range scenarios {
		byName[scenario.Name] = scenario
	}
	ordered := make([]Scenario, 0, len(scenarios))
	sequenced := make(map[string]bool)
	for _, sequence := range suite.RecordingSequences {
		existingCassettes := 0
		for _, scenarioName := range sequence.Scenarios {
			scenario, exists := byName[scenarioName]
			if !exists {
				return nil, fmt.Errorf("recording sequence %q references unavailable scenario %q", sequence.Name, scenarioName)
			}
			if mode == ModeRecordMissing {
				_, err := os.Stat(filepath.Join(scenario.Directory, "http.json"))
				switch {
				case err == nil:
					existingCassettes++
				case !os.IsNotExist(err):
					return nil, fmt.Errorf("inspect recording sequence %q scenario %q cassette: %w", sequence.Name, scenarioName, err)
				}
			}
			ordered = append(ordered, scenario)
			sequenced[scenarioName] = true
		}
		if mode == ModeRecordMissing && existingCassettes != 0 && existingCassettes != len(sequence.Scenarios) {
			return nil, fmt.Errorf("recording sequence %q has both existing and missing HTTP cassettes; use -mode=refresh-all to record the complete sequence", sequence.Name)
		}
	}
	for _, scenario := range scenarios {
		if !sequenced[scenario.Name] {
			ordered = append(ordered, scenario)
		}
	}
	return ordered, nil
}

// ValidateRecordingRunFilter prevents Go's subtest filter from selecting only
// part of a mandatory recording sequence. Selecting every member or no member
// of each sequence remains valid.
func ValidateRecordingRunFilter(suite Suite, mode Mode, runPattern string) error {
	if mode == ModeReplay || mode == ModeUpdateOutput || len(suite.RecordingSequences) == 0 {
		return nil
	}
	parts := strings.Split(runPattern, "/")
	if len(parts) < 2 {
		return nil
	}
	childPattern, err := regexp.Compile(parts[1])
	if err != nil {
		return fmt.Errorf("parse scenario part of -run filter: %w", err)
	}
	for _, sequence := range suite.RecordingSequences {
		matched := 0
		for _, scenarioName := range sequence.Scenarios {
			if childPattern.MatchString(scenarioName) {
				matched++
			}
		}
		if matched != 0 && matched != len(sequence.Scenarios) {
			return fmt.Errorf("-run filter selects only part of recording sequence %q; select every member or none", sequence.Name)
		}
	}
	return nil
}

func requiresRecordingSequence(scenario Scenario) bool {
	if scenario.HTTPReplayOnly {
		return false
	}
	for _, expectation := range scenario.ExpectedHTTP {
		switch expectation.Method {
		case "GET", "HEAD", "OPTIONS":
			continue
		case "PUT":
			if expectation.Query == "validateOnly=true" {
				continue
			}
		case "POST":
			if strings.HasSuffix(expectation.Path, ":testIamPermissions") {
				continue
			}
		}
		return true
	}
	return false
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
