package applycmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/publication"
	"github.com/yumauri/fbrcm/ops/machine"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestApplyNoChangePlanSucceedsWithoutFirebase(t *testing.T) {
	rawConfig := json.RawMessage(`{"parameters":{},"version":{"versionNumber":"1"}}`)
	plan := publication.New("test", "update", "stateless", nil)
	plan.Targets = append(plan.Targets, publication.Target{
		Target: "demo", ProjectID: "demo", Template: "client", Action: publication.ActionNone,
		Base: publication.Snapshot{Version: "1", RemoteConfig: rawConfig}, Candidate: publication.Snapshot{RemoteConfig: rawConfig},
		Validation: publication.Validation{Source: "local", ValidatedAt: plan.CreatedAt}, Source: publication.Source{Kind: "direct"},
	})
	if err := publication.Seal(plan); err != nil {
		t.Fatal(err)
	}
	raw, err := publication.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "test.fbrcm-plan.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := New(nil)
	cmd.SetContext(core.WithExecutionPolicy(t.Context(), core.StatelessExecutionPolicy()))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{path, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"status": "unchanged"`)) {
		t.Fatalf("output = %s", out.String())
	}
}

func TestClassifyPublishResultCoversEveryStatusAndWarning(t *testing.T) {
	published := json.RawMessage(`{"parameters":{},"version":{"versionNumber":"2"}}`)
	longFailure := errors.New(strings.Repeat("failure ", 900))
	tests := []struct {
		name        string
		err         error
		status      Status
		stage       string
		accepted    bool
		warning     string
		wantVersion bool
	}{
		{name: "published", status: statusPublished, accepted: true, wantVersion: true},
		{name: "conflict", err: &firebase.APIError{StatusCode: 412}, status: statusConflict, stage: "publication"},
		{name: "failure", err: longFailure, status: statusPublishFailed, stage: "publication"},
		{name: "hook", err: &core.RemoteConfigPublishedHookError{ProjectID: "demo", RemoteConfig: published, HookErr: errors.New("hook failed")}, status: statusPublishedHookFailed, stage: "post_publish_hook", accepted: true, warning: "publication.post_publish_hook_failed", wantVersion: true},
		{name: "cache", err: &core.RemoteConfigPublishedCacheError{ProjectID: "demo", RemoteConfig: published, Err: errors.New("cache failed")}, status: statusPublishedCacheFailed, stage: "cache", accepted: true, warning: "publication.cache_stale", wantVersion: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.SetContext(machine.WithState(context.Background()))
			result, accepted := classifyPublishResult(cmd, "demo", targetResult{Target: "demo"}, published, test.err)
			if result.Status != test.status || accepted != test.accepted || result.Validated != true || result.ValidationSource != core.ValidationSourceFirebase {
				t.Fatalf("result = %#v, accepted=%t", result, accepted)
			}
			if test.wantVersion && result.PublishedVersion != "2" {
				t.Errorf("published version = %q", result.PublishedVersion)
			}
			if test.err == nil && result.Error != nil || test.err != nil && (result.Error == nil || result.Error.Stage != test.stage || len([]rune(result.Error.Message)) > machine.MaxSafeTextRunes+1) {
				t.Errorf("error = %#v", result.Error)
			}
			warnings := shared.MachineWarnings(cmd)
			if test.warning == "" && len(warnings) != 0 {
				t.Errorf("unexpected warnings = %#v", warnings)
			}
			if test.warning != "" && (len(warnings) != 1 || warnings[0].Code != test.warning || warnings[0].Target != "demo" || len(warnings[0].Remediation) != 1 || len(warnings[0].Remediation[0].Argv) == 0) {
				t.Errorf("warnings = %#v", warnings)
			}
		})
	}
}

func TestClassifyPreflightTargetCoversPreparedAlreadyAppliedAndStale(t *testing.T) {
	base := json.RawMessage(`{"parameters":{"flag":{"defaultValue":{"value":"off"}}},"version":{"versionNumber":"1"}}`)
	candidate := json.RawMessage(`{"parameters":{"flag":{"defaultValue":{"value":"on"}}},"version":{"versionNumber":"2"}}`)
	baseDigest, err := publication.RemoteConfigDigest(base)
	if err != nil {
		t.Fatal(err)
	}
	candidateDigest, err := publication.RemoteConfigDigest(candidate)
	if err != nil {
		t.Fatal(err)
	}
	target := publication.Target{
		Target: "demo", Base: publication.Snapshot{ETag: "etag-1", SHA256: baseDigest}, Candidate: publication.Snapshot{SHA256: candidateDigest},
	}
	prepared, result, stale, err := classifyPreflightTarget(target, base, "etag-1")
	if err != nil || prepared == nil || result != nil || stale {
		t.Fatalf("prepared classification = %#v, %#v, %t, %v", prepared, result, stale, err)
	}
	prepared, result, stale, err = classifyPreflightTarget(target, candidate, "etag-2")
	if err != nil || prepared != nil || result == nil || result.Status != statusAlreadyApplied || !result.Validated || result.ValidationSource != core.ValidationSourceLocal || stale {
		t.Fatalf("already-applied classification = %#v, %#v, %t, %v", prepared, result, stale, err)
	}
	third := json.RawMessage(`{"parameters":{"flag":{"defaultValue":{"value":"other"}}},"version":{"versionNumber":"3"}}`)
	prepared, result, stale, err = classifyPreflightTarget(target, third, "etag-3")
	if err != nil || prepared != nil || result != nil || !stale {
		t.Fatalf("stale classification = %#v, %#v, %t, %v", prepared, result, stale, err)
	}
}

func TestNonAtomicWarningHasTypedDetailsAndSkipsDryRun(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", true, "")
	if err := cmd.Flags().Set("json", "true"); err != nil {
		t.Fatal(err)
	}
	cmd.SetContext(machine.WithState(context.Background()))
	addNonAtomicWarning(cmd, 1, false)
	addNonAtomicWarning(cmd, 2, true)
	if warnings := shared.MachineWarnings(cmd); len(warnings) != 0 {
		t.Fatalf("suppressed warnings = %#v", warnings)
	}
	addNonAtomicWarning(cmd, 2, false)
	warnings := shared.MachineWarnings(cmd)
	if len(warnings) != 1 || warnings[0].Code != "publication.non_atomic" {
		t.Fatalf("warnings = %#v", warnings)
	}
	details, ok := warnings[0].Details.(struct {
		TargetCount int `json:"target_count"`
	})
	if !ok || details.TargetCount != 2 {
		t.Fatalf("warning details = %#v", warnings[0].Details)
	}
}

func TestCleanupMatchingDraftDeletesOnlyExactSourceAndWarnsOnDriftOrFailure(t *testing.T) {
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := config.SwitchProfile(config.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	raw := json.RawMessage(`{"parameters":{}}`)
	now := time.Now().UTC()
	saveDraft := func() {
		t.Helper()
		if err := config.SaveDraft(&config.Draft{FormatVersion: config.DraftFormatVersion, ProjectID: "demo", BaseVersion: "1", BaseETag: "etag-1", CreatedAt: now, UpdatedAt: now, BaseRemoteConfig: raw, RemoteConfig: raw}); err != nil {
			t.Fatal(err)
		}
	}
	saveDraft()
	record, exists, err := svc.LoadDraftRecord("demo")
	if err != nil || !exists {
		t.Fatalf("load draft = %#v, %t, %v", record, exists, err)
	}
	cmd := &cobra.Command{}
	cmd.SetContext(machine.WithState(context.Background()))
	cleanupMatchingDraft(cmd, svc, publication.Target{Target: "demo", Source: publication.Source{Kind: "draft", Fingerprint: record.UpdatedAt.UTC().Format(time.RFC3339Nano)}})
	if _, exists, err := svc.LoadDraftRecord("demo"); err != nil || exists {
		t.Fatalf("matching draft remains: exists=%t err=%v", exists, err)
	}

	saveDraft()
	cleanupMatchingDraft(cmd, svc, publication.Target{Target: "demo", Source: publication.Source{Kind: "draft", Fingerprint: "changed"}})
	if _, exists, err := svc.LoadDraftRecord("demo"); err != nil || !exists {
		t.Fatalf("changed draft was removed: exists=%t err=%v", exists, err)
	}
	if err := os.WriteFile(config.GetDraftPath("demo"), []byte(`{`), 0o600); err != nil {
		t.Fatal(err)
	}
	cleanupMatchingDraft(cmd, svc, publication.Target{Target: "demo", Source: publication.Source{Kind: "draft", Fingerprint: "changed"}})
	warnings := shared.MachineWarnings(cmd)
	if len(warnings) != 2 || warnings[0].Code != "plan.source_draft_changed" || warnings[1].Code != "publication.draft_cleanup_failed" {
		t.Fatalf("cleanup warnings = %#v", warnings)
	}
}
