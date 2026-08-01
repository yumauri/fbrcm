package projects

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

func TestProjectAliasesCommandsRoundTrip(t *testing.T) {
	root := setupAliasCommandTest(t)

	set := newAliasesSetCommand()
	var out bytes.Buffer
	set.SetOut(&out)
	set.SetErr(&out)
	set.SetArgs([]string{"prod", "acme-production-42"})
	if err := set.Execute(); err != nil {
		t.Fatalf("set alias = %v", err)
	}
	if !strings.Contains(out.String(), "added project alias: prod -> acme-production-42") {
		t.Fatalf("set output = %q", out.String())
	}

	localPath := filepath.Join(root, coreconfig.LocalConfigFileName)
	raw, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "prod = 'acme-production-42'") && !strings.Contains(string(raw), `prod = "acme-production-42"`) {
		t.Fatalf("local config = %q", raw)
	}

	list := newAliasesListCommand()
	out.Reset()
	list.SetOut(&out)
	list.SetArgs([]string{"--json"})
	if err := list.Execute(); err != nil {
		t.Fatalf("list aliases = %v", err)
	}
	var rows []projectAliasRow
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Alias != "prod" || rows[0].ProjectID != "acme-production-42" || rows[0].Source != coreconfig.ProjectAliasSourceFBRCM {
		t.Fatalf("alias rows = %#v", rows)
	}

	remove := newAliasesRemoveCommand()
	out.Reset()
	remove.SetOut(&out)
	remove.SetErr(&out)
	remove.SetArgs([]string{"prod", "--yes", "--json"})
	if err := remove.Execute(); err != nil {
		t.Fatalf("remove alias = %v", err)
	}
	var result projectAliasRemoveResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.Status != "removed" || result.PreviousProjectID != "acme-production-42" {
		t.Fatalf("remove result = %#v", result)
	}
}

func TestProjectAliasesListEmptyAndNarrowTable(t *testing.T) {
	setupAliasCommandTest(t)
	if table := renderProjectAliasesTable(nil, 80); !strings.Contains(table, "Alias") || !strings.Contains(table, "Project ID") {
		t.Fatalf("empty table = %q", table)
	}
	rows := []projectAliasRow{{Alias: strings.Repeat("production", 8), ProjectID: "acme-production-project-42", Source: coreconfig.ProjectAliasSourceFirebase}}
	table := renderProjectAliasesTable(rows, 32)
	for line := range strings.SplitSeq(table, "\n") {
		if width := lipgloss.Width(line); width > 32 {
			t.Fatalf("narrow table line width = %d: %q", width, line)
		}
	}
	if !strings.Contains(table, "…") {
		t.Fatalf("narrow table did not crop: %q", table)
	}
}

func TestProjectAliasesListFirebaseAndRejectMutation(t *testing.T) {
	root := setupAliasCommandTest(t)
	if err := os.WriteFile(filepath.Join(root, coreconfig.FirebaseConfigFileName), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, coreconfig.FirebaseRCFileName), []byte(`{"projects":{"prod":"acme-production-42"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	list := newAliasesListCommand()
	var out bytes.Buffer
	list.SetOut(&out)
	list.SetArgs([]string{"--json"})
	if err := list.Execute(); err != nil {
		t.Fatal(err)
	}
	var rows []projectAliasRow
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Source != coreconfig.ProjectAliasSourceFirebase {
		t.Fatalf("Firebase alias rows = %#v", rows)
	}
	unchanged := newAliasesSetCommand()
	out.Reset()
	unchanged.SetOut(&out)
	unchanged.SetArgs([]string{"prod", "acme-production-42", "--json"})
	if err := unchanged.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, coreconfig.LocalConfigFileName)); !os.IsNotExist(err) {
		t.Fatalf("unchanged Firebase alias created native config: %v", err)
	}

	set := newAliasesSetCommand()
	set.SetOut(&bytes.Buffer{})
	set.SetErr(&bytes.Buffer{})
	set.SetArgs([]string{"prod", "other-production-42", "--yes"})
	if err := set.Execute(); err == nil || !strings.Contains(err.Error(), coreconfig.FirebaseRCFileName) {
		t.Fatalf("Firebase alias set error = %v", err)
	}
	remove := newAliasesRemoveCommand()
	remove.SetOut(&bytes.Buffer{})
	remove.SetErr(&bytes.Buffer{})
	remove.SetArgs([]string{"prod", "--yes"})
	if err := remove.Execute(); err == nil || !strings.Contains(err.Error(), coreconfig.FirebaseRCFileName) {
		t.Fatalf("Firebase alias remove error = %v", err)
	}
}

func TestProjectAliasesImportDryRunAndOverwrite(t *testing.T) {
	root := setupAliasCommandTest(t)
	localPath := filepath.Join(root, coreconfig.LocalConfigFileName)
	if err := os.WriteFile(localPath, []byte("[projects.aliases]\nprod = \"old-production-42\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	from := filepath.Join(root, "firebase-aliases.json")
	if err := os.WriteFile(from, []byte(`{"projects":{"prod":"new-production-42","staging":"acme-staging-42"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	reject := newAliasesImportCommand()
	reject.SetOut(&bytes.Buffer{})
	reject.SetErr(&bytes.Buffer{})
	reject.SetArgs([]string{"--from", from, "--yes"})
	if err := reject.Execute(); err == nil || !strings.Contains(err.Error(), "--conflict") {
		t.Fatalf("default import conflict error = %v", err)
	}

	dryRun := newAliasesImportCommand()
	var out bytes.Buffer
	dryRun.SetOut(&out)
	dryRun.SetErr(&out)
	dryRun.SetArgs([]string{"--from", from, "--conflict", "overwrite", "--dry-run", "--json"})
	if err := dryRun.Execute(); err != nil {
		t.Fatal(err)
	}
	var preview projectAliasImportResult
	if err := json.Unmarshal(out.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	if !preview.Changed || !preview.DryRun || len(preview.Items) != 2 {
		t.Fatalf("dry-run result = %#v", preview)
	}
	raw, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "new-production") || strings.Contains(string(raw), "staging") {
		t.Fatalf("dry run changed local config: %s", raw)
	}

	apply := newAliasesImportCommand()
	out.Reset()
	apply.SetOut(&out)
	apply.SetErr(&out)
	apply.SetArgs([]string{"--from", from, "--conflict", "overwrite", "--yes", "--json"})
	if err := apply.Execute(); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(localPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "new-production-42") || !strings.Contains(string(raw), "acme-staging-42") {
		t.Fatalf("imported local config: %s", raw)
	}
}

func TestProjectAliasesRemoveNativeCopyKeepsFirebaseAlias(t *testing.T) {
	root := setupAliasCommandTest(t)
	if err := os.WriteFile(filepath.Join(root, coreconfig.LocalConfigFileName), []byte("[projects.aliases]\nprod = \"acme-production-42\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, coreconfig.FirebaseConfigFileName), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, coreconfig.FirebaseRCFileName), []byte(`{"projects":{"prod":"acme-production-42"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newAliasesRemoveCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"prod", "--yes", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var result projectAliasRemoveResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "removed_native" || result.RemainingSource != coreconfig.ProjectAliasSourceFirebase {
		t.Fatalf("remove shared alias result = %#v", result)
	}
	registry, err := coreconfig.LoadProjectAliasRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if registry.Entries["prod"].Source != coreconfig.ProjectAliasSourceFirebase {
		t.Fatalf("remaining alias registry = %#v", registry)
	}
}

func TestProjectAliasesImportConflictAndNarrowTable(t *testing.T) {
	items, conflicts := planProjectAliasImport(
		map[string]string{"prod": "old-production-42"},
		map[string]string{"prod": "new-production-42"},
		aliasImportConflictError,
	)
	if len(conflicts) != 1 || conflicts[0] != "prod" || len(items) != 1 || items[0].Action != "conflict" {
		t.Fatalf("import plan = %#v, conflicts=%v", items, conflicts)
	}
	table := renderProjectAliasImportTable([]projectAliasImportItem{{
		Alias:             strings.Repeat("production", 6),
		PreviousProjectID: "old-production-project-42",
		ProjectID:         "new-production-project-42",
		Action:            "overwrite",
	}}, 54)
	for line := range strings.SplitSeq(table, "\n") {
		if width := lipgloss.Width(line); width > 54 {
			t.Fatalf("narrow import table line width = %d: %q", width, line)
		}
	}
	if !strings.Contains(table, "…") {
		t.Fatalf("narrow import table did not crop: %q", table)
	}
}

func TestProjectAliasesCommandsRejectDisabledLocalConfig(t *testing.T) {
	setupAliasCommandTest(t)
	coreconfig.SetLocalConfigDisabled(true)
	t.Cleanup(func() { coreconfig.SetLocalConfigDisabled(false) })
	cmd := newAliasesListCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("disabled local config error = %v", err)
	}
}

func setupAliasCommandTest(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	coreconfig.SetLocalConfigDisabled(false)
	t.Cleanup(func() {
		coreconfig.SetLocalConfigDisabled(false)
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	return root
}
