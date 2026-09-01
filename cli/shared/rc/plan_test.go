package rc

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/cli/machine"
)

func TestCreatePlanFileIsPrivateAndExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "publication-plan.json")
	original := []byte(`{"plan_id":"plan_original"}`)
	if err := createPlanFile(path, original); err != nil {
		t.Fatalf("createPlanFile = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("plan permissions = %o, want 600", got)
	}

	err = createPlanFile(path, []byte(`{"plan_id":"plan_replacement"}`))
	var conflict *machine.ConflictError
	if !errors.As(err, &conflict) || conflict.Code != "plan.exists" {
		t.Fatalf("second createPlanFile error = %#v, want plan.exists conflict", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("existing plan was replaced: %s", got)
	}
}
