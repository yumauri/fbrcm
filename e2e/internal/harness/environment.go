package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	authID          = "e2e"
	replayToken     = "e2e-replay-token"
	fixedTimestamp  = "2024-01-01T00:00:00Z"
	farFutureExpiry = "2099-01-01T00:00:00Z"
)

type Environment struct {
	Variables []string
	WorkDir   string
}

// ApplyStateFixture overlays deterministic config, cache, home, and work files.
func ApplyStateFixture(environment Environment, fixtureRoot string) error {
	values := environmentMap(environment.Variables)
	targets := map[string]string{
		"config": values["FBRCM_CONFIG_DIR"],
		"cache":  values["FBRCM_CACHE_DIR"],
		"home":   values["HOME"],
		"work":   environment.WorkDir,
	}
	replacements := []SnapshotReplacement{
		{Old: "<E2E_CONFIG_DIR>", New: targets["config"]},
		{Old: "<E2E_CACHE_DIR>", New: targets["cache"]},
		{Old: "<E2E_HOME_DIR>", New: targets["home"]},
		{Old: "<E2E_WORK_DIR>", New: targets["work"]},
		{Old: "<E2E_CANONICAL_WORK_DIR>", New: canonicalFixturePath(targets["work"])},
	}
	if _, err := os.Stat(fixtureRoot); err != nil {
		return fmt.Errorf("read state fixture %s: %w", fixtureRoot, err)
	}
	for name, destination := range targets {
		source := filepath.Join(fixtureRoot, name)
		if _, err := os.Stat(source); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return fmt.Errorf("read state fixture directory %s: %w", source, err)
		}
		if err := copyFixtureTree(source, destination, replacements); err != nil {
			return err
		}
	}
	return nil
}

func canonicalFixturePath(path string) string {
	canonical, err := filepath.EvalSymlinks(path)
	if err == nil {
		return canonical
	}
	return filepath.Clean(path)
}

func copyFixtureTree(source, destination string, replacements []SnapshotReplacement) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("read state fixture directory %s: %w", source, err)
	}
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return fmt.Errorf("create state fixture directory %s: %w", destination, err)
	}
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("state fixture %s must not contain symbolic links", filepath.Join(source, entry.Name()))
		}
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())
		if entry.IsDir() {
			if err := copyFixtureTree(sourcePath, destinationPath, replacements); err != nil {
				return err
			}
			continue
		}
		raw, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("read state fixture %s: %w", sourcePath, err)
		}
		raw = CanonicalizeSnapshot(raw, replacements...)
		if err := os.WriteFile(destinationPath, raw, 0o600); err != nil {
			return fmt.Errorf("stage state fixture %s: %w", destinationPath, err)
		}
	}
	return nil
}

func PrepareEnvironment(root, fixturesRoot string, suite Suite, proxyURL, certificatePath, accessToken string, terminalWidth int, logLevel string, localConfig bool) (Environment, error) {
	configDir := filepath.Join(root, "config")
	cacheDir := filepath.Join(root, "cache")
	profileConfigDir := filepath.Join(configDir, "default")
	profileCacheDir := filepath.Join(cacheDir, "default")
	workDir := filepath.Join(root, "work")
	homeDir := filepath.Join(root, "home")
	for _, directory := range []string{configDir, cacheDir, profileConfigDir, profileCacheDir, workDir, homeDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return Environment{}, fmt.Errorf("create E2E directory %s: %w", directory, err)
		}
	}

	authConfigDir := filepath.Join(profileConfigDir, "auth", authID)
	authCacheDir := filepath.Join(profileCacheDir, "auth", authID)
	for _, directory := range []string{authConfigDir, authCacheDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return Environment{}, fmt.Errorf("create E2E auth directory: %w", err)
		}
	}
	if err := copyPrivateFile(filepath.Join(fixturesRoot, "auth.json"), filepath.Join(profileConfigDir, "auth-config.json")); err != nil {
		return Environment{}, err
	}
	if err := copyPrivateFile(filepath.Join(fixturesRoot, "oauth-client.json"), filepath.Join(authConfigDir, "client-secret.json")); err != nil {
		return Environment{}, err
	}
	if err := writeProjectsFixture(filepath.Join(profileConfigDir, "projects-config.json"), suite); err != nil {
		return Environment{}, err
	}
	if strings.TrimSpace(accessToken) == "" {
		accessToken = replayToken
	}
	if err := writeJSONFile(filepath.Join(authCacheDir, "token.json"), map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expiry":       farFutureExpiry,
	}, 0o600); err != nil {
		return Environment{}, err
	}

	values := environmentMap(os.Environ())
	for _, key := range []string{
		"ALL_PROXY", "all_proxy", "HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy",
		"NO_PROXY", "no_proxy", "FBRCM_OFFLINE", "FBRCM_PROFILE", "FBRCM_EDITOR",
		"FBRCM_HOOK_TRUST", "GOOGLE_CLOUD_QUOTA_PROJECT", "XDG_CONFIG_HOME",
	} {
		delete(values, key)
	}
	values["HOME"] = homeDir
	values["FBRCM_CONFIG_DIR"] = configDir
	values["FBRCM_CACHE_DIR"] = cacheDir
	if !localConfig {
		values["FBRCM_NO_LOCAL_CONFIG"] = "1"
	}
	if logLevel == "" {
		logLevel = defaultScenarioLogLevel
	}
	values["FBRCM_LOG_LEVEL"] = logLevel
	values["FBRCM_LOG_NO_TIMESTAMP"] = "1"
	if suite.QuotaProjectID != "" {
		values["GOOGLE_CLOUD_QUOTA_PROJECT"] = suite.QuotaProjectID
	}
	values["NO_COLOR"] = "1"
	if terminalWidth == 0 {
		terminalWidth = defaultTerminalWidth
	}
	values["COLUMNS"] = strconv.Itoa(terminalWidth)
	values["TERM"] = "dumb"
	values["TZ"] = "UTC"
	values["LANG"] = "C.UTF-8"
	values["LC_ALL"] = "C.UTF-8"
	values["HTTP_PROXY"] = proxyURL
	values["HTTPS_PROXY"] = proxyURL
	values["SSL_CERT_FILE"] = certificatePath
	values["NO_PROXY"] = "127.0.0.1,localhost,proxy.golang.org,sum.golang.org"
	return Environment{Variables: flattenEnvironment(values), WorkDir: workDir}, nil
}

func writeProjectsFixture(path string, suite Suite) error {
	return writeJSONFile(path, map[string]any{
		"version": 2,
		"projects": []map[string]any{{
			"name":             suite.ProjectName,
			"project_id":       suite.ProjectID,
			"auth_id":          authID,
			"templates":        []string{"client"},
			"primary_template": "client",
			"synced_at":        fixedTimestamp,
		}},
		"synced_at": fixedTimestamp,
	}, 0o600)
}

func writeJSONFile(path string, value any, mode os.FileMode) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, mode); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func copyPrivateFile(source, destination string) error {
	raw, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read fixture %s: %w", source, err)
	}
	if err := os.WriteFile(destination, raw, 0o600); err != nil {
		return fmt.Errorf("stage fixture %s: %w", destination, err)
	}
	return nil
}

func environmentMap(environment []string) map[string]string {
	values := make(map[string]string, len(environment))
	for _, entry := range environment {
		key, value, found := strings.Cut(entry, "=")
		if found {
			values[key] = value
		}
	}
	return values
}

func flattenEnvironment(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	environment := make([]string, 0, len(keys))
	for _, key := range keys {
		environment = append(environment, key+"="+values[key])
	}
	return environment
}
