package projects

import (
	"testing"

	"github.com/yumauri/fbrcm/tui/testutil"
)

// parityTestModel builds a representative projects panel used to lock in
// rendered output before view_render splits.
func parityTestModel() Model {
	return loadedProjectsModel()
}

func TestProjectsViewSnapshot(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := testutil.NormalizeViewSnapshot(parityTestModel().View(true))
	if got != projectsViewSnapshot {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, projectsViewSnapshot)
	}
}

const projectsViewSnapshot = `── [1] Projects ─────────── 3 ─╮
 Alpha Project                 │
  alpha                        │
                               │
 Beta Project                  │
  beta                         │
                               │
 Gamma Project                 │
  gamma                        │
                               │
                               │
───────────────────────────────╯`
