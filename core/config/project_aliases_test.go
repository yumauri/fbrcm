package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestProjectAliasesDecodeAndRoundTrip(t *testing.T) {
	raw := []byte("[projects.aliases]\ndev = \"acme-development-42\"\nprod = \"acme-production-42\"\n")
	cfg, err := DecodeAppConfig(raw, true)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"dev": "acme-development-42", "prod": "acme-production-42"}
	if got := CloneProjectAliases(cfg); !reflect.DeepEqual(got, want) {
		t.Fatalf("aliases = %#v, want %#v", got, want)
	}
	encoded, err := MarshalAppConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeAppConfig(encoded, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := CloneProjectAliases(decoded); !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip aliases = %#v, want %#v", got, want)
	}
}

func TestProjectAliasValidation(t *testing.T) {
	for _, alias := range []string{"Prod", "prod.eu", "-prod", "prod eu", ""} {
		if err := ValidateProjectAliasName(alias); err == nil {
			t.Fatalf("ValidateProjectAliasName(%q) succeeded", alias)
		}
	}
	for _, projectID := range []string{"", " server@demo-project", "server@demo-project", "=demo-project", "demo project"} {
		if err := ValidateProjectAliasProjectID(projectID); err == nil {
			t.Fatalf("ValidateProjectAliasProjectID(%q) succeeded", projectID)
		}
	}
	if err := ValidatePhysicalProjectID("acme-production-42"); err != nil {
		t.Fatalf("ValidatePhysicalProjectID valid = %v", err)
	}
	if err := ValidateProjectAliases(map[string]string{"prod-eu_2": "acme-production-42"}); err != nil {
		t.Fatalf("valid aliases = %v", err)
	}
}

func TestResolveProjectAliasRequiresExactCase(t *testing.T) {
	aliases := map[string]string{"prod": "acme-production-42"}
	if alias, projectID, ok := ResolveProjectAlias(aliases, "prod"); !ok || alias != "prod" || projectID != "acme-production-42" {
		t.Fatalf("exact alias = %q, %q, %t", alias, projectID, ok)
	}
	if _, _, ok := ResolveProjectAlias(aliases, "PROD"); ok {
		t.Fatal("case-mismatched alias unexpectedly resolved")
	}
}

func TestResolveAppConfigRejectsGlobalProjectAliases(t *testing.T) {
	setupTestDirs(t)
	if err := EnsurePrivateDir(GetConfigRootDirPath()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(GetGlobalConfigFilePath(), []byte("[projects.aliases]\nprod = \"acme-production-42\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveAppConfig()
	if err == nil || !strings.Contains(err.Error(), "repository-scoped") || !strings.Contains(err.Error(), GetGlobalConfigFilePath()) {
		t.Fatalf("ResolveAppConfig error = %v", err)
	}
}

func TestLoadProjectAliasesUsesNearestLocalConfig(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	nested := filepath.Join(root, "one", "two")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, LocalConfigFileName), "[projects.aliases]\nprod = \"acme-production-42\"\n", 0o644)
	withWorkingDirectory(t, nested)

	aliases, err := LoadProjectAliases()
	if err != nil {
		t.Fatal(err)
	}
	if aliases["prod"] != "acme-production-42" {
		t.Fatalf("aliases = %#v", aliases)
	}

	SetLocalConfigDisabled(true)
	t.Cleanup(func() { SetLocalConfigDisabled(false) })
	aliases, err = LoadProjectAliases()
	if err != nil {
		t.Fatal(err)
	}
	if len(aliases) != 0 {
		t.Fatalf("disabled aliases = %#v, want empty", aliases)
	}
}

func TestSetAndRemoveProjectAlias(t *testing.T) {
	cfg := &AppConfig{}
	previous, changed, err := SetProjectAlias(cfg, "prod", "acme-production-42")
	if err != nil || previous != "" || !changed {
		t.Fatalf("set new = %q, %t, %v", previous, changed, err)
	}
	previous, changed, err = SetProjectAlias(cfg, "prod", "acme-production-42")
	if err != nil || previous != "acme-production-42" || changed {
		t.Fatalf("set unchanged = %q, %t, %v", previous, changed, err)
	}
	previous, changed, err = RemoveProjectAlias(cfg, "prod")
	if err != nil || previous != "acme-production-42" || !changed || cfg.Projects != nil {
		t.Fatalf("remove = %q, %t, %v, projects=%#v", previous, changed, err, cfg.Projects)
	}
}
