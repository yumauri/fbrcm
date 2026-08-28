package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidLogLevel(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error", "fatal", "silent"} {
		if !validLogLevel(level) {
			t.Errorf("validLogLevel(%q) = false", level)
		}
	}
	for _, level := range []string{"", "warning", "trace"} {
		if validLogLevel(level) {
			t.Errorf("validLogLevel(%q) = true", level)
		}
	}
}

func TestLoadSuiteAppliesOutputDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suite.json")
	if err := os.WriteFile(path, []byte(`{"project_id":"fixture"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	suite, err := LoadSuite(path)
	if err != nil {
		t.Fatal(err)
	}
	if suite.DefaultTerminalWidth != 200 || suite.DefaultLogLevel != "debug" {
		t.Fatalf("suite defaults = width %d, log level %q", suite.DefaultTerminalWidth, suite.DefaultLogLevel)
	}
}

func TestLoadSuiteReadsOutputDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suite.json")
	if err := os.WriteFile(path, []byte(`{
  "project_id": "fixture",
  "default_terminal_width": 88,
  "default_log_level": "WARN"
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	suite, err := LoadSuite(path)
	if err != nil {
		t.Fatal(err)
	}
	if suite.DefaultTerminalWidth != 88 || suite.DefaultLogLevel != "warn" {
		t.Fatalf("suite defaults = width %d, log level %q", suite.DefaultTerminalWidth, suite.DefaultLogLevel)
	}
}

func TestLoadSuiteReadsRecordingSequences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suite.json")
	if err := os.WriteFile(path, []byte(`{
  "project_id": "fixture",
  "recording_sequences": [
    {"name": " parameter-lifecycle ", "scenarios": ["parameter-add", " parameter-update ", "parameter-delete"]}
  ]
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	suite, err := LoadSuite(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(suite.RecordingSequences) != 1 {
		t.Fatalf("recording sequences = %#v", suite.RecordingSequences)
	}
	sequence := suite.RecordingSequences[0]
	if sequence.Name != "parameter-lifecycle" || len(sequence.Scenarios) != 3 || sequence.Scenarios[1] != "parameter-update" {
		t.Fatalf("recording sequence = %#v", sequence)
	}
}

func TestLoadSuiteRejectsDuplicateScenarioInRecordingSequence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suite.json")
	if err := os.WriteFile(path, []byte(`{
  "project_id": "fixture",
  "recording_sequences": [
    {"name": "parameter-lifecycle", "scenarios": ["parameter-add", "parameter-add"]}
  ]
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSuite(path); err == nil {
		t.Fatal("LoadSuite accepted a duplicate scenario in a recording sequence")
	}
}

func TestLoadScenariosAcceptsZeroHTTPAndExpandsProject(t *testing.T) {
	root := t.TempDir()
	for name, raw := range map[string]string{
		"local": `{
  "name":"local",
  "command_id":"auth.list",
  "args":["auth","list","--json"],
  "expected_exit_code":0,
  "expected_http":[],
  "expected_files":["export.json"],
  "expected_state_files":[{"root":"CONFIG","path":"default/projects.json","json_replacements":{"/synced_at":"<SYNCED_AT>"}}],
  "expected_absent_state_paths":[{"root":"cache","path":"default/remote-config"}],
  "json_output":true,
  "envs":{"FBRCM_GOOGLE_ACCESS_TOKEN":"${FBRCM_E2E_ACCESS_TOKEN}"},
  "offline":true
}`,
		"remote": `{
  "name":"remote",
  "command_id":"versions.list",
  "args":["versions","list","${PROJECT_ID}"],
  "expected_exit_code":0,
  "expected_http":[{"method":"get","host":"firebase.example","path":"/v1/projects/${PROJECT_ID}/versions","status":200}],
  "http_replay_only":true
}`,
	} {
		directory := filepath.Join(root, name)
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	scenarios, err := LoadScenarios(root, Suite{ProjectID: "fixture-project", QuotaProjectID: "billing-project"})
	if err != nil {
		t.Fatal(err)
	}
	if len(scenarios) != 2 || len(scenarios[0].ExpectedHTTP) != 0 || len(scenarios[0].ExpectedFiles) != 1 || scenarios[0].ExpectedFiles[0] != "export.json" || !scenarios[0].Offline || scenarios[0].Environment["FBRCM_GOOGLE_ACCESS_TOKEN"] != e2eAccessTokenVariable {
		t.Fatalf("scenarios = %#v", scenarios)
	}
	stateFile := scenarios[0].ExpectedStateFiles[0]
	if stateFile.Root != "config" || stateFile.Path != filepath.Join("default", "projects.json") || stateFile.JSONReplacements["/synced_at"] != "<SYNCED_AT>" {
		t.Fatalf("state file = %#v", stateFile)
	}
	absent := scenarios[0].ExpectedAbsentStatePaths[0]
	if absent.Root != "cache" || absent.Path != filepath.Join("default", "remote-config") {
		t.Fatalf("absent state path = %#v", absent)
	}
	remote := scenarios[1]
	if remote.Args[2] != "fixture-project" || remote.ExpectedHTTP[0].Method != "GET" || remote.ExpectedHTTP[0].Path != "/v1/projects/fixture-project/versions" || remote.ExpectedHTTP[0].QuotaProjectID != "" || !remote.HTTPReplayOnly {
		t.Fatalf("expanded remote scenario = %#v", remote)
	}
}

func TestLoadScenariosRequiresQuotaProjectForGoogleAPIRequests(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "remote")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"remote",
  "command_id":"projects.list",
  "args":["--stateless","projects","list","--json"],
  "expected_exit_code":0,
  "expected_http":[{"method":"get","host":"cloudresourcemanager.googleapis.com","path":"/v1/projects","status":200}]
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"}); err == nil {
		t.Fatal("LoadScenarios accepted a Google API request without a quota project")
	}
	scenarios, err := LoadScenarios(root, Suite{ProjectID: "fixture-project", QuotaProjectID: "billing-project"})
	if err != nil {
		t.Fatal(err)
	}
	if got := scenarios[0].ExpectedHTTP[0].QuotaProjectID; got != "billing-project" {
		t.Fatalf("quota project = %q, want billing-project", got)
	}
}

func TestLoadScenariosRejectsReplayOnlyWithoutHTTP(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "invalid")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"invalid",
  "command_id":"auth.list",
  "args":["auth","list","--json"],
  "expected_exit_code":0,
  "expected_http":[],
  "http_replay_only":true
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"}); err == nil {
		t.Fatal("LoadScenarios accepted http_replay_only without HTTP expectations")
	}
}

func TestLoadScenariosRejectsHarnessOwnedEnvironmentOverride(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "invalid")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"invalid",
  "command_id":"project.export",
  "args":["--stateless","project","export","fixture-project","--json"],
  "expected_exit_code":0,
  "expected_http":[],
  "envs":{"HTTPS_PROXY":"https://uncontrolled.example"}
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"}); err == nil {
		t.Fatal("LoadScenarios accepted a harness-owned environment override")
	}
}

func TestLoadScenariosRejectsPresentAndAbsentStateConflict(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "conflict")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"conflict",
  "command_id":"cache.clear",
  "args":["cache","clear","--yes","--json"],
  "expected_exit_code":0,
  "expected_http":[],
  "expected_state_files":[{"root":"cache","path":"default/remote-config"}],
  "expected_absent_state_paths":[{"root":"cache","path":"default/remote-config"}]
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"}); err == nil {
		t.Fatal("LoadScenarios accepted one state path as both present and absent")
	}
}

func TestLoadScenariosRejectsUnsafeExpectedStateFile(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "unsafe")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"unsafe",
  "command_id":"projects.update",
  "args":["projects","update","--json"],
  "expected_exit_code":0,
  "expected_http":[],
  "expected_state_files":[{"root":"home","path":"../outside.json"}]
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"}); err == nil {
		t.Fatal("LoadScenarios accepted an unsafe expected state file")
	}
}

func TestLoadScenariosRejectsInvalidStateJSONReplacement(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "invalid")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"invalid",
  "command_id":"projects.update",
  "args":["projects","update","--json"],
  "expected_exit_code":0,
  "expected_http":[],
  "expected_state_files":[{"root":"config","path":"projects.json","json_replacements":{"synced_at":"<SYNCED_AT>"}}]
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"}); err == nil {
		t.Fatal("LoadScenarios accepted an invalid JSON pointer")
	}
}

func TestLoadScenariosRejectsUnsafeExpectedFilePath(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "unsafe")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"unsafe",
  "command_id":"versions.export",
  "args":["versions","export","demo","1"],
  "expected_exit_code":0,
  "expected_http":[],
  "expected_files":["../outside.json"]
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"}); err == nil {
		t.Fatal("LoadScenarios accepted an expected file outside the scenario work directory")
	}
}

func TestLoadScenariosRejectsUnknownRecordingSequenceScenario(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "parameter-add")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"parameter-add",
  "command_id":"add",
  "args":["add","test_parameter","--value","initial"],
  "expected_exit_code":0,
  "expected_http":[]
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	suite := Suite{
		ProjectID: "fixture-project",
		RecordingSequences: []RecordingSequence{{
			Name:      "parameter-lifecycle",
			Scenarios: []string{"parameter-add", "parameter-delete"},
		}},
	}
	if _, err := LoadScenarios(root, suite); err == nil {
		t.Fatal("LoadScenarios accepted an unknown recording-sequence scenario")
	}
}

func TestLoadScenariosRequiresSequenceForMutatingLiveHTTP(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "parameter-add")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "name":"parameter-add",
  "command_id":"add",
  "args":["add","test_parameter","--value","initial"],
  "expected_exit_code":0,
  "expected_http":[
    {"method":"PUT","host":"firebase.example","path":"/v1/projects/fixture/remoteConfig","status":200}
  ]
}`
	if err := os.WriteFile(filepath.Join(directory, "scenario.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"}); err == nil {
		t.Fatal("LoadScenarios accepted mutating live HTTP outside a recording sequence")
	}
}

func TestRequiresRecordingSequenceRecognizesNonMutatingWrites(t *testing.T) {
	for name, scenario := range map[string]Scenario{
		"validate only": {
			ExpectedHTTP: []HTTPExpectation{{Method: "PUT", Query: "validateOnly=true"}},
		},
		"permission check": {
			ExpectedHTTP: []HTTPExpectation{{Method: "POST", Path: "/v1/projects/demo:testIamPermissions"}},
		},
		"synthetic replay": {
			HTTPReplayOnly: true,
			ExpectedHTTP:   []HTTPExpectation{{Method: "DELETE"}},
		},
	} {
		if requiresRecordingSequence(scenario) {
			t.Errorf("%s scenario requires a recording sequence", name)
		}
	}
}

func TestOrderScenariosForModeMakesRecordingSequenceContiguous(t *testing.T) {
	scenarios := []Scenario{
		{Name: "parameter-add"},
		{Name: "parameter-delete"},
		{Name: "project-show"},
		{Name: "parameter-update"},
	}
	suite := Suite{RecordingSequences: []RecordingSequence{{
		Name:      "parameter-lifecycle",
		Scenarios: []string{"parameter-add", "parameter-update", "parameter-delete"},
	}}}

	ordered, err := OrderScenariosForMode(scenarios, suite, ModeRefreshAll)
	if err != nil {
		t.Fatal(err)
	}
	if len(ordered) != 4 || ordered[0].Name != "parameter-add" || ordered[1].Name != "parameter-update" || ordered[2].Name != "parameter-delete" || ordered[3].Name != "project-show" {
		t.Fatalf("ordered scenarios = %#v", ordered)
	}
}

func TestOrderScenariosForModeLeavesReplayOrderIndependent(t *testing.T) {
	scenarios := []Scenario{{Name: "parameter-add"}, {Name: "parameter-delete"}, {Name: "parameter-update"}}
	suite := Suite{RecordingSequences: []RecordingSequence{{
		Name:      "parameter-lifecycle",
		Scenarios: []string{"parameter-add", "parameter-update", "parameter-delete"},
	}}}
	for _, mode := range []Mode{ModeReplay, ModeUpdateOutput} {
		ordered, err := OrderScenariosForMode(scenarios, suite, mode)
		if err != nil {
			t.Fatal(err)
		}
		if ordered[1].Name != "parameter-delete" {
			t.Errorf("mode %q reordered independent scenarios: %#v", mode, ordered)
		}
	}
}

func TestOrderScenariosForModeRejectsPartialRecordMissingSequence(t *testing.T) {
	root := t.TempDir()
	addDirectory := filepath.Join(root, "parameter-add")
	if err := os.MkdirAll(addDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(addDirectory, "http.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	scenarios := []Scenario{
		{Name: "parameter-add", Directory: addDirectory},
		{Name: "parameter-delete", Directory: filepath.Join(root, "parameter-delete")},
	}
	suite := Suite{RecordingSequences: []RecordingSequence{{
		Name:      "parameter-lifecycle",
		Scenarios: []string{"parameter-add", "parameter-delete"},
	}}}
	if _, err := OrderScenariosForMode(scenarios, suite, ModeRecordMissing); err == nil {
		t.Fatal("OrderScenariosForMode accepted a partially recorded sequence")
	}
}

func TestValidateRecordingRunFilterRejectsPartialSequence(t *testing.T) {
	suite := Suite{RecordingSequences: []RecordingSequence{{
		Name:      "parameter-lifecycle",
		Scenarios: []string{"parameter-add", "parameter-update", "parameter-delete"},
	}}}
	if err := ValidateRecordingRunFilter(suite, ModeRefreshAll, `^TestCLI$/^parameter-add$`); err == nil {
		t.Fatal("ValidateRecordingRunFilter accepted one member of a recording sequence")
	}
	if err := ValidateRecordingRunFilter(suite, ModeRefreshAll, `^TestCLI$/^parameter-(add|update|delete)$`); err != nil {
		t.Fatalf("ValidateRecordingRunFilter rejected the complete recording sequence: %v", err)
	}
	if err := ValidateRecordingRunFilter(suite, ModeRefreshAll, `^TestCLI$/^project-show$`); err != nil {
		t.Fatalf("ValidateRecordingRunFilter rejected an unrelated standalone scenario: %v", err)
	}
}

func TestValidateRecordingRunFilterDoesNotRestrictReplay(t *testing.T) {
	suite := Suite{RecordingSequences: []RecordingSequence{{
		Name:      "parameter-lifecycle",
		Scenarios: []string{"parameter-add", "parameter-update", "parameter-delete"},
	}}}
	if err := ValidateRecordingRunFilter(suite, ModeReplay, `^TestCLI$/^parameter-add$`); err != nil {
		t.Fatalf("ValidateRecordingRunFilter restricted independent replay: %v", err)
	}
}
