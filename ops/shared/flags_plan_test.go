package shared

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddPlanOutFlagOnlyConflictsWithPlanMode(t *testing.T) {
	cmd := &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error { return nil }}
	AddDryRunFlag(cmd)
	cmd.Flags().Bool("draft", false, "")
	AddYesFlag(cmd, "")
	AddPlanOutFlag(cmd)

	for _, args := range [][]string{
		{"--dry-run", "--yes"},
		{"--draft", "--yes"},
	} {
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%v unexpectedly conflicts: %v", args, err)
		}
	}

	for _, args := range [][]string{
		{"--plan-out", "plan.json", "--dry-run"},
		{"--plan-out", "plan.json", "--draft"},
		{"--plan-out", "plan.json", "--yes"},
	} {
		fresh := &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error { return nil }}
		AddDryRunFlag(fresh)
		fresh.Flags().Bool("draft", false, "")
		AddYesFlag(fresh, "")
		AddPlanOutFlag(fresh)
		fresh.SetArgs(args)
		if err := fresh.Execute(); err == nil {
			t.Fatalf("%v did not conflict", args)
		}
	}
}
