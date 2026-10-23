package config

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	charmlog "charm.land/log/v2"

	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func TestProjectVisibilityDefaultsAndRoundTrip(t *testing.T) {
	setupTestDirs(t)
	policy, err := LoadProjectVisibility()
	if err != nil || !reflect.DeepEqual(policy, DefaultProjectVisibility()) {
		t.Fatalf("default = %#v, %v", policy, err)
	}
	if _, err := os.Stat(GetProjectVisibilityFilePath()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read created preferences: %v", err)
	}
	policy.Blacklist.ProjectIDs = []string{"zeta", "alpha", "alpha"}
	policy.Whitelist.Enabled = true
	policy.Whitelist.ProjectIDs = []string{"alpha", "beta"}
	if err := SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadProjectVisibility()
	if err != nil || !reflect.DeepEqual(loaded.Blacklist.ProjectIDs, []string{"alpha", "zeta"}) {
		t.Fatalf("round trip = %#v, %v", loaded, err)
	}
	assertFileMode(t, GetProjectVisibilityFilePath(), PrivateFileMode)
	for id, want := range map[string]string{"alpha": "blacklisted", "server@alpha": "blacklisted", "beta": "", "server@beta": "", "gamma": "not_whitelisted"} {
		if got := loaded.HiddenReason(id); got != want {
			t.Errorf("HiddenReason(%q) = %q, want %q", id, got, want)
		}
	}
	loaded.Whitelist.ProjectIDs = []string{}
	if loaded.HiddenReason("beta") != "not_whitelisted" {
		t.Fatal("empty enabled whitelist must hide every project")
	}
	loaded.Whitelist.Enabled = false
	loaded.Blacklist.Enabled = false
	if loaded.HiddenReason("alpha") != "" {
		t.Fatal("disabled lists must retain membership without hiding projects")
	}
}

func TestProjectVisibilitySurvivesResetAndIsProfileScoped(t *testing.T) {
	setupTestDirs(t)
	policy := DefaultProjectVisibility()
	policy.Blacklist.ProjectIDs = []string{"alpha"}
	if err := SaveProjectVisibility(policy); err != nil {
		t.Fatal(err)
	}
	if _, err := ResetProjects(); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadProjectVisibility()
	if err != nil || loaded.HiddenReason("alpha") != "blacklisted" {
		t.Fatalf("reset erased visibility: %#v, %v", loaded, err)
	}
	if err := SwitchProfile("other"); err != nil {
		t.Fatal(err)
	}
	loaded, err = LoadProjectVisibility()
	if err != nil || loaded.HiddenReason("alpha") != "" {
		t.Fatalf("preferences leaked across profiles: %#v, %v", loaded, err)
	}
}

func TestProjectVisibilityRejectsInvalidFilesAndIDs(t *testing.T) {
	setupTestDirs(t)
	for _, raw := range []string{"", "null", "{", `{"version":2}`, `{"version":1,"blacklist":{"project_ids":["server@alpha"]}}`} {
		writeFile(t, GetProjectVisibilityFilePath(), raw, PrivateFileMode)
		if _, err := LoadProjectVisibility(); err == nil {
			t.Errorf("accepted malformed preferences %q", raw)
		}
	}
	for _, id := range []string{"", " alpha", "alpha beta", "=alpha", "client@alpha"} {
		policy := DefaultProjectVisibility()
		policy.Blacklist.ProjectIDs = []string{id}
		if err := SaveProjectVisibility(policy); err == nil {
			t.Errorf("accepted ID %q", id)
		}
	}
}

func TestProjectVisibilityLogs(t *testing.T) {
	t.Setenv(env.LogLevel, "debug")
	t.Setenv(env.LogNoTimestamp, "1")
	t.Setenv(env.LogPlain, "1")
	var logs bytes.Buffer
	previousLevel := corelog.CurrentLevel()
	corelog.ConfigureCLIOutput(&logs, &logs)
	corelog.InitWithDefault(corelog.ModeCLI, charmlog.DebugLevel)
	t.Cleanup(func() {
		corelog.ConfigureCLIOutput(os.Stderr, os.Stderr)
		corelog.SetLevel(previousLevel)
	})
	load := func() error {
		_, err := LoadProjectVisibility()
		return err
	}
	policy := DefaultProjectVisibility()
	policy.Blacklist.ProjectIDs = []string{"alpha", "alpha"}
	policy.Whitelist.Enabled = true
	policy.Whitelist.ProjectIDs = []string{"alpha", "beta"}
	invalidPolicy := DefaultProjectVisibility()
	invalidPolicy.Version = 2
	for _, tc := range []struct {
		name    string
		prepare func(*testing.T)
		run     func() error
		wantErr bool
		want    []string
	}{
		{
			name: "missing file uses defaults", run: load,
			want: []string{"read project visibility config", "project visibility config missing; using defaults"},
		},
		{
			name: "loaded settings", run: load,
			prepare: func(t *testing.T) {
				if err := SaveProjectVisibility(policy); err != nil {
					t.Fatal(err)
				}
			},
			want: []string{"read project visibility config", "INFO loaded project visibility config", "blacklist_enabled=true", "blacklist_count=1", "whitelist_enabled=true", "whitelist_count=2"},
		},
		{
			name: "saved settings", run: func() error { return SaveProjectVisibility(policy) },
			want: []string{"write project visibility config", "saved project visibility config", "blacklist_enabled=true", "blacklist_count=1", "whitelist_enabled=true", "whitelist_count=2"},
		},
		{
			name: "decode failure", run: load, wantErr: true,
			prepare: func(t *testing.T) { writeFile(t, GetProjectVisibilityFilePath(), "{", PrivateFileMode) },
			want:    []string{"read project visibility config", "decode project visibility config failed", "err="},
		},
		{
			name: "load validation failure", run: load, wantErr: true,
			prepare: func(t *testing.T) { writeFile(t, GetProjectVisibilityFilePath(), `{"version":2}`, PrivateFileMode) },
			want:    []string{"validate project visibility config failed", "err="},
		},
		{
			name: "save validation failure", run: func() error { return SaveProjectVisibility(invalidPolicy) }, wantErr: true,
			want: []string{"validate project visibility config failed", "err="},
		},
		{
			name: "read failure", run: load, wantErr: true,
			prepare: visibilityPathDirectory,
			want:    []string{"read project visibility config failed", "err="},
		},
		{
			name: "write failure", run: func() error { return SaveProjectVisibility(policy) }, wantErr: true,
			prepare: visibilityPathDirectory,
			want:    []string{"write project visibility config", "write project visibility config failed", "err="},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setupTestDirs(t)
			if tc.prepare != nil {
				tc.prepare(t)
			}
			logs.Reset()
			if err := tc.run(); (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tc.wantErr)
			}
			for _, want := range append(tc.want, "component=config", "path="+GetProjectVisibilityFilePath()) {
				if !strings.Contains(logs.String(), want) {
					t.Errorf("missing log %q in:\n%s", want, logs.String())
				}
			}
		})
	}
}

func visibilityPathDirectory(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(GetProjectVisibilityFilePath(), PrivateDirMode); err != nil {
		t.Fatal(err)
	}
}
