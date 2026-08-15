package harness

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateJournalChecksExactHTTPContract(t *testing.T) {
	journal := Journal{Total: 1, Entries: []JournalEntry{{Mode: "simulate"}}}
	journal.Entries[0].Request.Method = http.MethodGet
	journal.Entries[0].Request.Destination = "firebase.example:443"
	journal.Entries[0].Request.Path = "/v1/fixture"
	journal.Entries[0].Request.Query = "validateOnly=true"
	journal.Entries[0].Response.Status = http.StatusInternalServerError
	expected := []HTTPExpectation{{
		Method: http.MethodGet,
		Host:   "firebase.example",
		Path:   "/v1/fixture",
		Query:  "validateOnly=true",
		Status: http.StatusInternalServerError,
	}}
	if err := validateJournal(journal, expected, false); err != nil {
		t.Fatal(err)
	}
	journal.Entries[0].Request.Path = "/v1/unexpected"
	if err := validateJournal(journal, expected, false); err == nil {
		t.Fatal("validateJournal accepted an unexpected path")
	}
}

func TestValidateJournalChecksDeclaredQuery(t *testing.T) {
	journal := Journal{Total: 1, Entries: []JournalEntry{{Mode: "simulate"}}}
	journal.Entries[0].Request.Method = http.MethodPut
	journal.Entries[0].Request.Destination = "firebase.example"
	journal.Entries[0].Request.Path = "/v1/fixture"
	journal.Entries[0].Request.Query = "validateOnly=false"
	journal.Entries[0].Response.Status = http.StatusOK
	expected := []HTTPExpectation{{
		Method: http.MethodPut, Host: "firebase.example", Path: "/v1/fixture",
		Query: "validateOnly=true", Status: http.StatusOK,
	}}
	if err := validateJournal(journal, expected, false); err == nil || !strings.Contains(err.Error(), "query") {
		t.Fatalf("validateJournal query error = %v", err)
	}
}

func TestValidateJournalAcceptsNoRequests(t *testing.T) {
	if err := validateJournal(Journal{}, nil, false); err != nil {
		t.Fatal(err)
	}
}

func TestHookFingerprintReplacements(t *testing.T) {
	fingerprint := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	replacements := hookFingerprintReplacements([]byte("fingerprint="+fingerprint), []byte(fingerprint))
	if len(replacements) != 1 || replacements[0].Old != fingerprint || replacements[0].New != "<E2E_HOOK_FINGERPRINT>" {
		t.Fatalf("replacements = %#v", replacements)
	}
}

func TestRedactDiagnosticOutput(t *testing.T) {
	raw := []byte(`headers={"Authorization":["Bearer opaque-token"]} direct=known-secret`)
	redacted := redactDiagnosticOutput(raw, "known-secret")
	if strings.Contains(string(redacted), "opaque-token") || strings.Contains(string(redacted), "known-secret") {
		t.Fatalf("redacted diagnostics retained a credential: %s", redacted)
	}
	if !strings.Contains(string(redacted), "Bearer <REDACTED>") || !strings.Contains(string(redacted), "direct=<REDACTED>") {
		t.Fatalf("redacted diagnostics = %s", redacted)
	}
}

func TestCheckExpectedFileSnapshots(t *testing.T) {
	root := t.TempDir()
	workDir := filepath.Join(root, "work")
	scenarioDir := filepath.Join(root, "scenario")
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		t.Fatal(err)
	}
	actualPath := filepath.Join(workDir, "version.json")
	if err := os.WriteFile(actualPath, []byte("first\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	scenario := Scenario{Directory: scenarioDir, ExpectedFiles: []string{"version.json"}}
	changes, err := checkExpectedFileSnapshots(scenario, workDir, ModeRecordMissing)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || !changes[0].Created {
		t.Fatalf("create changes = %#v", changes)
	}
	if _, err := checkExpectedFileSnapshots(scenario, workDir, ModeReplay); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(actualPath, []byte("second\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := checkExpectedFileSnapshots(scenario, workDir, ModeReplay); err == nil {
		t.Fatal("file snapshot replay accepted changed bytes")
	}
	changes, err = checkExpectedFileSnapshots(scenario, workDir, ModeUpdateOutput)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || !changes[0].Updated {
		t.Fatalf("update changes = %#v", changes)
	}
}
