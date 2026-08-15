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
	scenarios, err := LoadScenarios(root, Suite{ProjectID: "fixture-project"})
	if err != nil {
		t.Fatal(err)
	}
	if len(scenarios) != 2 || len(scenarios[0].ExpectedHTTP) != 0 || len(scenarios[0].ExpectedFiles) != 1 || scenarios[0].ExpectedFiles[0] != "export.json" || !scenarios[0].Offline {
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
	if remote.Args[2] != "fixture-project" || remote.ExpectedHTTP[0].Method != "GET" || remote.ExpectedHTTP[0].Path != "/v1/projects/fixture-project/versions" || !remote.HTTPReplayOnly {
		t.Fatalf("expanded remote scenario = %#v", remote)
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
