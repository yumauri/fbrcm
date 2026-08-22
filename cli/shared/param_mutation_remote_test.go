package shared

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core"
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
	if len(opts.ParamFilters) != 0 || opts.ParamArgument == nil || *opts.ParamArgument != "flag" {
		t.Fatalf("parameter selection = filters %v, argument %#v", opts.ParamFilters, opts.ParamArgument)
	}
	if opts.Search.Raw != "login" || !opts.Yes {
		t.Fatalf("opts = %+v", opts)
	}
}

func TestRunParameterMutationRemoteRejectsDraftInStatelessMode(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(core.WithExecutionPolicy(context.Background(), core.StatelessExecutionPolicy()))
	_, err := RunParameterMutationRemote(cmd, nil, ParameterMutationOpts{Draft: true}, "update", "", nil)
	var argument *ArgumentError
	if !errors.As(err, &argument) || !strings.Contains(err.Error(), "--draft cannot be used with --stateless") {
		t.Fatalf("RunParameterMutationRemote error = %T %v, want typed stateless draft error", err, err)
	}
}
