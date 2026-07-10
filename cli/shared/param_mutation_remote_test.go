package shared

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestReadParameterMutationOpts(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringArray("project", nil, "")
	cmd.Flags().String("expr", "", "")
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().StringArray("filter", nil, "")
	cmd.Flags().String("search", "", "")
	cmd.Flags().Bool("yes", false, "")

	if err := cmd.Flags().Set("project", "demo"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("search", "login"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}

	opts, err := ReadParameterMutationOpts(cmd, []string{"flag"})
	if err != nil {
		t.Fatalf("ReadParameterMutationOpts = %v", err)
	}
	if len(opts.ProjectFilters) != 1 || opts.ProjectFilters[0] != "demo" {
		t.Fatalf("ProjectFilters = %v", opts.ProjectFilters)
	}
	if len(opts.ParamFilters) != 1 || opts.ParamFilters[0] != "=flag" {
		t.Fatalf("ParamFilters = %v", opts.ParamFilters)
	}
	if opts.Search.Raw != "login" || !opts.Yes {
		t.Fatalf("opts = %+v", opts)
	}
}
