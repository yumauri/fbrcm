package remoteconfig_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestFixturesParse(t *testing.T) {
	fixtures := []string{
		"root_params.json",
		"grouped_params.json",
		"with_conditions.json",
		"empty_values.json",
		"personalization.json",
	}

	dir := fixtureDir(t)
	for _, name := range fixtures {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			cfg, err := firebase.ParseRemoteConfig(raw)
			if err != nil {
				t.Fatalf("ParseRemoteConfig: %v", err)
			}
			if cfg.Version.VersionNumber == "" {
				t.Fatalf("VersionNumber is empty")
			}
			out, err := firebase.MarshalRemoteConfig(cfg)
			if err != nil {
				t.Fatalf("MarshalRemoteConfig: %v", err)
			}
			roundTrip, err := firebase.ParseRemoteConfig(out)
			if err != nil {
				t.Fatalf("round-trip parse: %v", err)
			}
			if roundTrip.Version.VersionNumber != cfg.Version.VersionNumber {
				t.Fatalf("version = %q, want %q", roundTrip.Version.VersionNumber, cfg.Version.VersionNumber)
			}
		})
	}
}

func fixtureDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}
