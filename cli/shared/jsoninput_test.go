package shared

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestReadJSONInputFromPath(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "secret-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := file.WriteString(`{"client_id":"demo"}`); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	got, err := ReadJSONInput(&cobra.Command{}, file.Name(), "client secret", ErrNoJSONSelection)
	if err != nil {
		t.Fatalf("ReadJSONInput returned error: %v", err)
	}
	if string(got) != `{"client_id":"demo"}` {
		t.Fatalf("ReadJSONInput = %q, want temp file content", string(got))
	}
}
