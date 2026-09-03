package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareEnvironmentIsolatesCLIState(t *testing.T) {
	t.Setenv("FBRCM_LOG_PLAIN", "parent-plain-logs")
	t.Setenv("GOOGLE_CLOUD_QUOTA_PROJECT", "parent-quota-project")
	t.Setenv("FBRCM_E2E_ACCESS_TOKEN", "parent-e2e-token")
	t.Setenv("FBRCM_GOOGLE_ACCESS_TOKEN", "parent-stateless-token")
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	if err := os.MkdirAll(fixtures, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtures, "auth.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixtures, "oauth-client.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	environment, err := PrepareEnvironment(
		filepath.Join(root, "run"),
		fixtures,
		Suite{ProjectID: "fixture-project", ProjectName: "Fixture Project", QuotaProjectID: "suite-quota-project"},
		"http://127.0.0.1:1234",
		filepath.Join(root, "cert.pem"),
		"fixture-token",
		240,
		"warn",
		false,
		map[string]string{"FBRCM_GOOGLE_ACCESS_TOKEN": e2eAccessTokenVariable, "FBRCM_LOG_PLAIN": "1"},
	)
	if err != nil {
		t.Fatal(err)
	}
	values := environmentMap(environment.Variables)
	if values["FBRCM_LOG_PLAIN"] != "1" {
		t.Fatal("scenario plain-log setting was not applied")
	}
	if values["FBRCM_CONFIG_DIR"] == "" || values["FBRCM_CACHE_DIR"] == "" || values["HOME"] == "" {
		t.Fatalf("isolated roots are missing: %#v", values)
	}
	if values["HTTPS_PROXY"] != "http://127.0.0.1:1234" || values["SSL_CERT_FILE"] != filepath.Join(root, "cert.pem") {
		t.Fatalf("proxy environment = %#v", values)
	}
	if values["GOOGLE_CLOUD_QUOTA_PROJECT"] != "suite-quota-project" {
		t.Fatalf("quota project = %q", values["GOOGLE_CLOUD_QUOTA_PROJECT"])
	}
	if values["FBRCM_GOOGLE_ACCESS_TOKEN"] != "fixture-token" {
		t.Fatalf("stateless access token = %q", values["FBRCM_GOOGLE_ACCESS_TOKEN"])
	}
	if _, exists := values["FBRCM_E2E_ACCESS_TOKEN"]; exists {
		t.Fatal("private E2E access token variable leaked into child environment")
	}
	if values["COLUMNS"] != "240" {
		t.Fatalf("terminal width = %q", values["COLUMNS"])
	}
	if values["FBRCM_LOG_LEVEL"] != "warn" || values["FBRCM_LOG_NO_TIMESTAMP"] != "1" {
		t.Fatalf("deterministic logging environment = %#v", values)
	}
	tokenPath := filepath.Join(values["FBRCM_CACHE_DIR"], "profiles", "default", "auth", authID, "token.json")
	raw, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatal(err)
	}
	var token map[string]any
	if err := json.Unmarshal(raw, &token); err != nil {
		t.Fatal(err)
	}
	if token["access_token"] != "fixture-token" {
		t.Fatalf("access token = %v", token["access_token"])
	}
	projectsPath := filepath.Join(values["FBRCM_CONFIG_DIR"], "profiles", "default", "projects-config.json")
	if _, err := os.Stat(projectsPath); err != nil {
		t.Fatal(err)
	}

	withoutQuota, err := PrepareEnvironment(
		filepath.Join(root, "run-without-quota"),
		fixtures,
		Suite{ProjectID: "fixture-project", ProjectName: "Fixture Project"},
		"http://127.0.0.1:1234",
		filepath.Join(root, "cert.pem"),
		"",
		0,
		"",
		false,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := environmentMap(withoutQuota.Variables)["GOOGLE_CLOUD_QUOTA_PROJECT"]; exists {
		t.Fatal("parent GOOGLE_CLOUD_QUOTA_PROJECT leaked into isolated environment")
	}
	if _, exists := environmentMap(withoutQuota.Variables)["FBRCM_LOG_PLAIN"]; exists {
		t.Fatal("parent plain-log setting leaked into isolated environment")
	}
	if _, exists := environmentMap(withoutQuota.Variables)["FBRCM_GOOGLE_ACCESS_TOKEN"]; exists {
		t.Fatal("parent stateless access token leaked into ordinary scenario environment")
	}
	if got := environmentMap(withoutQuota.Variables)["COLUMNS"]; got != "200" {
		t.Fatalf("default terminal width = %q", got)
	}
	if got := environmentMap(withoutQuota.Variables)["FBRCM_LOG_LEVEL"]; got != "debug" {
		t.Fatalf("default log level = %q", got)
	}

	replayStateless, err := PrepareEnvironment(
		filepath.Join(root, "run-replay-stateless"),
		fixtures,
		Suite{ProjectID: "fixture-project", ProjectName: "Fixture Project"},
		"http://127.0.0.1:1234",
		filepath.Join(root, "cert.pem"),
		"",
		0,
		"",
		false,
		map[string]string{"FBRCM_GOOGLE_ACCESS_TOKEN": e2eAccessTokenVariable},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := environmentMap(replayStateless.Variables)["FBRCM_GOOGLE_ACCESS_TOKEN"]; got != replayToken {
		t.Fatalf("replay stateless access token = %q, want replay token", got)
	}

	literalStateless, err := PrepareEnvironment(
		filepath.Join(root, "run-literal-stateless"),
		fixtures,
		Suite{ProjectID: "fixture-project", ProjectName: "Fixture Project"},
		"http://127.0.0.1:1234",
		filepath.Join(root, "cert.pem"),
		"fixture-token",
		0,
		"",
		false,
		map[string]string{"FBRCM_GOOGLE_ACCESS_TOKEN": "incorrect"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := environmentMap(literalStateless.Variables)["FBRCM_GOOGLE_ACCESS_TOKEN"]; got != "incorrect" {
		t.Fatalf("literal stateless access token = %q, want incorrect", got)
	}
}

func TestPrepareEnvironmentCanEnableLocalConfig(t *testing.T) {
	root := t.TempDir()
	fixtures := filepath.Join(root, "fixtures")
	if err := os.MkdirAll(fixtures, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"auth.json", "oauth-client.json"} {
		if err := os.WriteFile(filepath.Join(fixtures, name), []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	environment, err := PrepareEnvironment(
		filepath.Join(root, "run"), fixtures, Suite{ProjectID: "fixture-project"},
		"http://127.0.0.1:1234", filepath.Join(root, "cert.pem"), "token", 200, "debug", true, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := environmentMap(environment.Variables)["FBRCM_NO_LOCAL_CONFIG"]; exists {
		t.Fatal("FBRCM_NO_LOCAL_CONFIG is set for a local-config scenario")
	}
}

func TestApplyStateFixtureOverlaysEnvironmentRoots(t *testing.T) {
	root := t.TempDir()
	fixture := filepath.Join(root, "fixture")
	for directory, filename := range map[string]string{
		"config/profiles/default": "config.txt",
		"cache/profiles/default":  "cache.txt",
		"home":                    "home.txt",
		"work":                    "work.txt",
	} {
		path := filepath.Join(fixture, directory)
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, filename), []byte(directory), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	environment := Environment{
		Variables: []string{
			"FBRCM_CONFIG_DIR=" + filepath.Join(root, "config"),
			"FBRCM_CACHE_DIR=" + filepath.Join(root, "cache"),
			"HOME=" + filepath.Join(root, "home"),
		},
		WorkDir: filepath.Join(root, "work"),
	}
	if err := ApplyStateFixture(environment, fixture); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(root, "config", "profiles", "default", "config.txt"),
		filepath.Join(root, "cache", "profiles", "default", "cache.txt"),
		filepath.Join(root, "home", "home.txt"),
		filepath.Join(root, "work", "work.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("overlay file %s: %v", path, err)
		}
	}
}

func TestApplyStateFixtureExpandsEnvironmentPathPlaceholders(t *testing.T) {
	root := t.TempDir()
	fixture := filepath.Join(root, "fixture")
	fixtureConfig := filepath.Join(fixture, "config")
	if err := os.MkdirAll(fixtureConfig, 0o700); err != nil {
		t.Fatal(err)
	}
	contents := "<E2E_CONFIG_DIR>\n<E2E_CACHE_DIR>\n<E2E_HOME_DIR>\n<E2E_WORK_DIR>\n<E2E_CANONICAL_WORK_DIR>\n"
	if err := os.WriteFile(filepath.Join(fixtureConfig, "paths.txt"), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	environment := Environment{
		Variables: []string{
			"FBRCM_CONFIG_DIR=" + filepath.Join(root, "config"),
			"FBRCM_CACHE_DIR=" + filepath.Join(root, "cache"),
			"HOME=" + filepath.Join(root, "home"),
		},
		WorkDir: filepath.Join(root, "work"),
	}
	if err := ApplyStateFixture(environment, fixture); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "config", "paths.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Join([]string{
		filepath.Join(root, "config"),
		filepath.Join(root, "cache"),
		filepath.Join(root, "home"),
		filepath.Join(root, "work"),
		canonicalFixturePath(filepath.Join(root, "work")),
		"",
	}, "\n")
	if string(raw) != want {
		t.Fatalf("expanded fixture = %q, want %q", raw, want)
	}
}
