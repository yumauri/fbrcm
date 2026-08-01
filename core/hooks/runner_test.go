package hooks

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func setupHookTest(t *testing.T, local string) string {
	t.Helper()
	root := t.TempDir()
	configRoot := filepath.Join(root, "config")
	t.Setenv(env.ConfigDir, configRoot)
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	t.Setenv(env.NoLocalConfig, "")
	t.Setenv(TrustEnvironment, "")
	config.SetLocalConfigDisabled(false)
	t.Cleanup(func() { config.SetLocalConfigDisabled(false) })
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	if local != "" {
		if err := os.WriteFile(filepath.Join(repo, config.LocalConfigFileName), []byte(local), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return repo
}

func TestLocalHooksRequireTrustAndChangesInvalidateIt(t *testing.T) {
	repo := setupHookTest(t, `[hooks]
timeout = "3s"
pre_publish = ["true"]
`)

	resolution, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Trusted || !resolution.LocalHooks || resolution.Fingerprint == "" {
		t.Fatalf("resolution = %+v", resolution)
	}
	_, err = Prepare(Metadata{Target: "demo", Candidate: []byte(`{}`)}, nil)
	if err == nil || !strings.Contains(err.Error(), "hooks trust") {
		t.Fatalf("Prepare untrusted error = %v", err)
	}

	trusted, err := TrustCurrent()
	if err != nil || !trusted.Trusted {
		t.Fatalf("TrustCurrent = %+v, %v", trusted, err)
	}
	if info, err := os.Stat(trustStorePath()); err != nil || info.Mode().Perm() != config.PrivateFileMode {
		t.Fatalf("trust store mode = %v, %v", info, err)
	}
	if err := os.WriteFile(filepath.Join(repo, config.LocalConfigFileName), []byte(`[hooks]
timeout = "3s"
pre_publish = ["printf changed"]
`), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if changed.Trusted || changed.Fingerprint == trusted.Fingerprint {
		t.Fatalf("changed resolution = %+v, prior = %+v", changed, trusted)
	}
}

func TestRunnerProvidesContextEnvironmentAndOrderedEvents(t *testing.T) {
	repo := setupHookTest(t, `[hooks]
timeout = "3s"
pre_publish = ["test \"$FBRCM_HOOK_EVENT\" = pre_publish && test \"$FBRCM_PROJECT_ID\" = demo && grep -q candidate \"$FBRCM_CANDIDATE_FILE\" && printf pre > pre.marker"]
post_publish = ["test \"$FBRCM_HOOK_EVENT\" = post_publish && grep -q published \"$FBRCM_PUBLISHED_FILE\" && printf post > post.marker"]
`)
	t.Setenv("FBRCM_PROJECT_ID", "wrong")
	if _, err := TrustCurrent(); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	session, err := Prepare(Metadata{
		Operation: "update", Target: "demo", Current: []byte(`{"old":true}`),
		Candidate: []byte(`{"candidate":true}`), ChangeNote: "policy test",
	}, &output)
	if err != nil {
		t.Fatal(err)
	}
	tempDir := session.tempDir
	defer session.Close()
	if err := session.Run(context.Background(), PrePublish, nil); err != nil {
		t.Fatal(err)
	}
	if err := session.Run(context.Background(), PostPublish, []byte(`{"published":true}`)); err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"pre.marker", "post.marker"} {
		if _, err := os.Stat(filepath.Join(repo, marker)); err != nil {
			t.Fatalf("missing %s: %v", marker, err)
		}
	}
	if !strings.Contains(output.String(), "Running pre_publish hook 1/1") || !strings.Contains(output.String(), "Running post_publish hook 1/1") {
		t.Fatalf("output = %q", output.String())
	}
	session.Close()
	if _, err := os.Stat(tempDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary hook directory still exists: %v", err)
	}
}

func TestGlobalHooksAreTrustedAndUseGlobalConfigDirectory(t *testing.T) {
	setupHookTest(t, "")
	if err := config.SaveAppConfig(&config.AppConfig{Hooks: &config.HooksConfig{PrePublish: []string{"true"}}}); err != nil {
		t.Fatal(err)
	}
	resolution, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	path, local := resolution.Source(PrePublish)
	if !resolution.Trusted || local || path != config.GetGlobalConfigFilePath() {
		t.Fatalf("resolution = %+v, source = %s local=%t", resolution, path, local)
	}
}

func TestHookTimeoutIsReported(t *testing.T) {
	setupHookTest(t, `[hooks]
timeout = "1ms"
pre_publish = ["sleep 1"]
`)
	if _, err := TrustCurrent(); err != nil {
		t.Fatal(err)
	}
	session, err := Prepare(Metadata{Target: "demo", Candidate: []byte(`{}`)}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	err = session.Run(context.Background(), PrePublish, nil)
	var hookErr *Error
	if !errors.As(err, &hookErr) || !hookErr.TimedOut {
		t.Fatalf("Run error = %v", err)
	}
}

func TestHookErrorIncludesBoundedOutputTail(t *testing.T) {
	setupHookTest(t, `[hooks]
pre_publish = ["printf policy-error >&2; exit 9"]
`)
	if _, err := TrustCurrent(); err != nil {
		t.Fatal(err)
	}
	session, err := Prepare(Metadata{Target: "demo", Candidate: []byte(`{}`)}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	err = session.Run(context.Background(), PrePublish, nil)
	var hookErr *Error
	if !errors.As(err, &hookErr) || hookErr.Output != "policy-error" || !strings.Contains(err.Error(), "Hook output:\npolicy-error") {
		t.Fatalf("Run error = %#v / %v", hookErr, err)
	}
}
