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
	if len(rows) != 1 || rows[0].Alias != "prod" || rows[0].ProjectID != "acme-production-42" {
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
	rows := []projectAliasRow{{Alias: strings.Repeat("production", 8), ProjectID: "acme-production-project-42"}}
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
