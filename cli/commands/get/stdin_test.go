package get

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestWriteRowsJSONFormatsRows(t *testing.T) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	cachedAt := time.Date(2026, 6, 13, 10, 11, 12, 0, time.UTC)
	defaultValue := "on"
	err := writeRowsJSON(cmd, []parameterRow{
		{
			Project:      "Project A",
			ProjectID:    "project-a",
			Group:        defaultGroupLabel,
			Key:          "welcome_flag",
			Description:  "Welcome <flag>",
			DefaultValue: &defaultValue,
			Conditional:  true,
			Conditions: []parameterConditionJSON{
				{Name: "beta", Value: &defaultValue},
			},
			Type:     "STRING",
			Version:  "42",
			CachedAt: cachedAt,
			Status:   "cache",
		},
	})
	if err != nil {
		t.Fatalf("writeRowsJSON returned error: %v", err)
	}

	if !strings.Contains(out.String(), `"description": "Welcome <flag>"`) {
		t.Fatalf("JSON output escaped HTML or omitted description: %s", out.String())
	}

	var rows []parameterRowJSON
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatalf("decode JSON output: %v\n%s", err, out.String())
	}
	if len(rows) != 1 {
		t.Fatalf("row count = %d, want 1", len(rows))
	}
	if rows[0].ProjectID != "project-a" || rows[0].Key != "welcome_flag" {
		t.Fatalf("row = %#v, want project-a/welcome_flag", rows[0])
	}
	if rows[0].Version == nil || *rows[0].Version != "42" {
		t.Fatalf("version = %#v, want 42", rows[0].Version)
	}
}

func TestFilterJSONFileNames(t *testing.T) {
	got := filterJSONFileNames([]string{
		"alpha.json",
		"beta.JSON",
		"notes.txt",
		"nested/config.json",
		"gamma",
	})
	want := []string{"alpha.json", "beta.JSON", "nested/config.json"}
	if len(got) != len(want) {
		t.Fatalf("filterJSONFileNames length = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("filterJSONFileNames = %#v, want %#v", got, want)
		}
	}
}

func TestStdinDirectoryProjectName(t *testing.T) {
	got := stdinDirectoryProjectName("my-firebase_project")
	if got != "My Firebase Project" {
		t.Fatalf("stdinDirectoryProjectName = %q, want My Firebase Project", got)
	}
}

func TestStdinVersion(t *testing.T) {
	got := stdinVersion([]byte(`{"version":{"versionNumber":" 123 "}}`))
	if got != "123" {
		t.Fatalf("stdinVersion = %q, want 123", got)
	}
	if got := stdinVersion([]byte(`{"parameters":{}}`)); got != "" {
		t.Fatalf("stdinVersion without version = %q, want empty", got)
	}
}
