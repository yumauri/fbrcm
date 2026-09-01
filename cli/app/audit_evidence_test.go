package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/schemas"
)

var auditTestClasses = []string{
	"artifact", "batch", "boundary", "determinism", "discovery", "effectiveness", "effects", "failure",
	"interaction", "invocation", "no_op", "selection", "stdin", "success", "warning",
}

type auditEvidenceMatrix struct {
	AuditStandardVersion string                 `json:"audit_standard_version"`
	ContractVersion      string                 `json:"contract_version"`
	Catalog              map[string]string      `json:"catalog"`
	Commands             []auditCommandEvidence `json:"commands"`
}

type auditCommandEvidence struct {
	ID      string                       `json:"id"`
	Classes map[string]auditEvidenceCell `json:"classes"`
}

type auditEvidenceCell struct {
	Evidence      []string `json:"evidence,omitempty"`
	NotApplicable string   `json:"not_applicable,omitempty"`
}

func TestEveryExecutableCommandHasCompleteAuditEvidenceMatrix(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "contract_v1_audit_evidence.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var matrix auditEvidenceMatrix
	if err := json.Unmarshal(raw, &matrix); err != nil {
		t.Fatal(err)
	}
	if matrix.AuditStandardVersion != "1.0.0" || matrix.ContractVersion != contract.Version {
		t.Fatalf("evidence versions = audit %q, contract %q", matrix.AuditStandardVersion, matrix.ContractVersion)
	}
	validateAuditEvidenceCatalog(t, matrix.Catalog)

	root := NewRootForContract("audit-evidence")
	capabilities := contract.DetailedCapabilities(root)
	if len(matrix.Commands) != len(capabilities) {
		t.Fatalf("evidence command count = %d, want %d", len(matrix.Commands), len(capabilities))
	}
	for index, capability := range capabilities {
		entry := matrix.Commands[index]
		if entry.ID != capability.ID {
			t.Fatalf("evidence command %d = %q, want %q", index, entry.ID, capability.ID)
		}
		if len(entry.Classes) != len(auditTestClasses) {
			t.Fatalf("%s evidence has %d classes, want %d", entry.ID, len(entry.Classes), len(auditTestClasses))
		}
		responseRaw, err := schemas.ReadByID(capability.ResponseSchema)
		if err != nil {
			t.Fatal(err)
		}
		inputRaw, err := schemas.ReadByID(capability.InvocationSchema)
		if err != nil {
			t.Fatal(err)
		}
		for _, class := range auditTestClasses {
			cell, ok := entry.Classes[class]
			if !ok {
				t.Errorf("%s omits %s evidence", entry.ID, class)
				continue
			}
			applicable := auditClassApplies(class, capability, inputRaw, responseRaw)
			if applicable {
				if cell.NotApplicable != "" || len(cell.Evidence) == 0 {
					t.Errorf("%s %s must contain evidence, got %#v", entry.ID, class, cell)
				}
			} else if strings.TrimSpace(cell.NotApplicable) == "" || len(cell.Evidence) != 0 {
				t.Errorf("%s %s must contain only a justified N/A, got %#v", entry.ID, class, cell)
			}
			for _, evidenceID := range cell.Evidence {
				if scenarioName, ok := strings.CutPrefix(evidenceID, "e2e:"); ok {
					validateAuditScenarioEvidence(t, entry.ID, scenarioName)
					continue
				}
				if _, ok := matrix.Catalog[evidenceID]; !ok {
					t.Errorf("%s %s references unknown evidence %q", entry.ID, class, evidenceID)
				}
			}
		}
	}
}

func validateAuditEvidenceCatalog(t *testing.T, catalog map[string]string) {
	t.Helper()
	for id, reference := range catalog {
		path, symbol, ok := strings.Cut(reference, "#")
		if id == "" || !ok || path == "" || symbol == "" {
			t.Errorf("invalid audit evidence catalog entry %q=%q", id, reference)
			continue
		}
		raw, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
		if err != nil {
			t.Errorf("audit evidence %q: %v", id, err)
			continue
		}
		if !bytes.Contains(raw, []byte("func "+symbol+"(")) {
			t.Errorf("audit evidence %q references missing test %s", id, reference)
		}
	}
}

func validateAuditScenarioEvidence(t *testing.T, commandID, scenarioName string) {
	t.Helper()
	path := filepath.Join("..", "..", "e2e", "testdata", "scenarios", scenarioName, "scenario.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("%s references missing E2E scenario %q: %v", commandID, scenarioName, err)
		return
	}
	var scenario struct {
		CommandID string `json:"command_id"`
	}
	if err := json.Unmarshal(raw, &scenario); err != nil {
		t.Errorf("decode %s: %v", path, err)
	} else if scenario.CommandID != commandID {
		t.Errorf("scenario %s covers %q, not %q", scenarioName, scenario.CommandID, commandID)
	}
}

func auditClassApplies(class string, capability contract.Capability, inputRaw, responseRaw []byte) bool {
	switch class {
	case "discovery", "invocation", "boundary", "effectiveness", "failure", "determinism":
		return true
	case "selection":
		return bytes.Contains(inputRaw, []byte(`"x-fbrcm-matching"`))
	case "stdin":
		return capability.Supports.Stdin
	case "success":
		return slices.Contains(responseReachableOutcomes(responseRaw), "success")
	case "no_op":
		return isAuditMutation(capability.ID)
	case "interaction":
		return capability.Interaction.Mode != "none"
	case "warning":
		return slices.ContainsFunc(contract.KnownWarningCodes(), func(code string) bool {
			return bytes.Contains(responseRaw, []byte(fmt.Sprintf("%q", code)))
		})
	case "batch":
		return slices.Contains(responseReachableOutcomes(responseRaw), "partial_success")
	case "effects":
		return capability.SideEffectLevel > 0 || capability.NetworkAccess != "none"
	case "artifact":
		return bytes.Contains(responseRaw, []byte(`"sha256"`)) && bytes.Contains(responseRaw, []byte(`"size_bytes"`))
	default:
		panic("unknown audit test class " + class)
	}
}

func responseReachableOutcomes(raw []byte) []string {
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil
	}
	allOf, _ := schema["allOf"].([]any)
	if len(allOf) != 2 {
		return nil
	}
	commandSchema, _ := allOf[1].(map[string]any)
	constraints, _ := commandSchema["allOf"].([]any)
	for _, rawConstraint := range constraints {
		constraint, _ := rawConstraint.(map[string]any)
		if _, conditional := constraint["if"]; conditional {
			continue
		}
		properties, _ := constraint["properties"].(map[string]any)
		outcomeSchema, _ := properties["outcome"].(map[string]any)
		if enum, ok := outcomeSchema["enum"].([]any); ok {
			result := make([]string, 0, len(enum))
			for _, value := range enum {
				result = append(result, value.(string))
			}
			return result
		}
	}
	return nil
}

func isAuditMutation(commandID string) bool {
	return slices.Contains([]string{
		"add", "apply", "auth.add.gcloud", "auth.add.google", "auth.add.oauth", "auth.add.service-account", "auth.bind", "auth.delete", "auth.login", "auth.quota-project.set", "auth.quota-project.unset",
		"cache.clear", "conditions.add", "conditions.delete", "conditions.edit", "conditions.move", "conditions.rename", "config.edit", "config.reset", "config.set", "delete", "draft.change-note", "draft.discard", "draft.publish", "duplicate",
		"experiments.delete", "groups.add", "groups.delete", "groups.edit", "groups.rename", "hooks.trust", "hooks.untrust", "profile", "profile.delete", "profile.rename", "profile.switch", "project.import", "project.quota-project.set", "project.quota-project.unset",
		"projects.aliases.import", "projects.aliases.remove", "projects.aliases.set", "projects.forget", "projects.promote", "projects.reset", "projects.update", "rollouts.delete", "theme.delete", "theme.import", "theme.rename", "theme.reset", "theme.switch", "update", "versions.restore", "versions.rollback",
	}, commandID)
}
