package planflag

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/ops/machine"
)

func TestOutputPath(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("plan-out", "", "")

	if path, enabled, err := OutputPath(cmd); err != nil || enabled || path != "" {
		t.Fatalf("unset = %q, %t, %v", path, enabled, err)
	}
	if err := cmd.Flags().Set("plan-out", "  plan.json  "); err != nil {
		t.Fatal(err)
	}
	if path, enabled, err := OutputPath(cmd); err != nil || !enabled || path != "plan.json" {
		t.Fatalf("set = %q, %t, %v", path, enabled, err)
	}

	for _, invalid := range []string{" ", "-"} {
		if err := cmd.Flags().Set("plan-out", invalid); err != nil {
			t.Fatal(err)
		}
		_, _, err := OutputPath(cmd)
		if _, ok := err.(*machine.ArgumentError); !ok {
			t.Fatalf("%q error = %T %v", invalid, err, err)
		}
	}
}
