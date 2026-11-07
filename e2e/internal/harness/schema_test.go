package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCommandOutputUsesActiveContractAndRetainedSchemas(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"contract.lock.json":                 `{"version":"1.1.0"}`,
		"1.0.0/artifact.schema.json":         `{"$id":"urn:test:artifact:1.0.0","type":"string"}`,
		"1.0.0/explain.response.schema.json": `{"$id":"urn:test:explain:1.0.0","type":"object"}`,
		"1.1.0/explain.response.schema.json": `{
  "$id":"urn:test:explain:1.1.0",
  "type":"object",
  "required":["contract_version","command","data"],
  "properties":{
    "contract_version":{"const":"1.1.0"},
    "command":{"const":"explain"},
    "data":{"$ref":"urn:test:artifact:1.0.0"}
  }
}`,
	}
	for name, raw := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		name  string
		raw   string
		valid bool
	}{
		{"current contract with retained artifact", `{"contract_version":"1.1.0","command":"explain","data":"value"}`, true},
		{"old contract is rejected", `{"contract_version":"1.0.0","command":"explain","data":"value"}`, false},
		{"wrong operation is rejected", `{"contract_version":"1.1.0","command":"get","data":"value"}`, false},
		{"retained artifact constraints apply", `{"contract_version":"1.1.0","command":"explain","data":12}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateCommandOutput(root, "explain", []byte(test.raw))
			if (err == nil) != test.valid {
				t.Fatalf("valid=%t: %v", test.valid, err)
			}
		})
	}
	if err := ValidateCommandOutput(root, "missing", []byte(`{}`)); err == nil {
		t.Fatal("missing current response schema was accepted")
	}
}
