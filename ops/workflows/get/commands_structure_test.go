package get

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	cmdtest "github.com/yumauri/fbrcm/cli/commands/testutil"
	"github.com/yumauri/fbrcm/core"
)

func TestNewCommandStructure(t *testing.T) {
	cmdtest.AssertCommandStructure(t, New(nil), "get [parameter]",
		"json", "project", "filter", "expr", "search", "all", "update")
}

func TestStatelessExactProjectTargetsBypassDiscoveryAndDeduplicate(t *testing.T) {
	ctx := core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy())
	projects, gotCtx, err := resolveGetProjectsForExecution(ctx, &cobra.Command{}, nil, []string{
		"=one-project",
		"server@=two-project",
		"client@=one-project",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotCtx != ctx {
		t.Fatal("exact-only selection unexpectedly replaced the execution context")
	}
	if len(projects) != 2 || projects[0].ProjectID != "one-project" || projects[1].ProjectID != "server@two-project" {
		t.Fatalf("projects = %+v", projects)
	}
}

func TestUpdateIsRejectedForStdin(t *testing.T) {
	input, err := os.CreateTemp(t.TempDir(), "remote-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = input.Close() }()
	if _, err := input.WriteString(`{"parameters":{}}`); err != nil {
		t.Fatal(err)
	}
	if _, err := input.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	cmd := New(nil)
	cmd.SetIn(input)
	cmd.SetArgs([]string{"--update"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "--update cannot be used with stdin") {
		t.Fatalf("Execute() error = %v, want stdin update rejection", err)
	}
}
