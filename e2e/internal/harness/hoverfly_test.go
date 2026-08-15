package harness

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSanitizeSimulation(t *testing.T) {
	raw := []byte(`{
  "meta": {"timeExported": "now"},
  "data": {
    "authorization": ["Bearer secret-token"],
    "nested": {"Set-Cookie": "session=secret-token", "safe": "kept"}
  }
}`)
	sanitized, err := SanitizeSimulation(raw, "secret-token")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(sanitized, []byte("secret-token")) || bytes.Contains(bytes.ToLower(sanitized), []byte("authorization")) {
		t.Fatalf("sanitized simulation retained credentials:\n%s", sanitized)
	}
	var value map[string]any
	if err := json.Unmarshal(sanitized, &value); err != nil {
		t.Fatal(err)
	}
	meta := value["meta"].(map[string]any)
	if meta["timeExported"] != fixedTimestamp {
		t.Fatalf("timeExported = %v", meta["timeExported"])
	}
}

func TestSanitizeSimulationRejectsSecretInBody(t *testing.T) {
	raw := []byte(`{"meta":{},"data":{"body":"secret-token"}}`)
	if _, err := SanitizeSimulation(raw, "secret-token"); err == nil {
		t.Fatal("SanitizeSimulation() accepted a secret in a response body")
	}
}
