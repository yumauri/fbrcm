package operation

import (
	"io"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
)

func TestAdapterOwnsSemanticExitPresentationAndBuildMetadata(t *testing.T) {
	root := &cobra.Command{Use: "fbrcm", Version: "test-build"}
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.AddCommand(Command(&invocation.Definition{Use: "diff", Args: invocation.NoArgs, RunE: func(call invocation.Call, _ []string) error {
		if invocation.Version(call) != "test-build" {
			t.Fatal("build metadata not forwarded")
		}
		return shared.DiffFoundError(call)
	}}))
	root.SetArgs([]string{"diff"})
	if err := root.Execute(); err == nil {
		t.Fatal("semantic exit lost")
	}
	if !root.SilenceErrors || !root.SilenceUsage {
		t.Fatal("semantic diff printed a failure")
	}
}
