// Package operation adapts shared application workflows to Cobra. Application
// and MCP execution must not depend on this adapter.
package operation

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
)

func Command(definition *invocation.Definition) *cobra.Command {
	cmd := &cobra.Command{Use: definition.Use, Short: definition.Short, Long: definition.Long}
	switch b := definition.Args; b.Kind {
	case "none":
		cmd.Args = cobra.NoArgs
	case "any":
		cmd.Args = cobra.ArbitraryArgs
	case "exact":
		cmd.Args = cobra.ExactArgs(b.Min)
	case "maximum":
		cmd.Args = cobra.MaximumNArgs(b.Max)
	case "range":
		cmd.Args = cobra.RangeArgs(b.Min, b.Max)
	}
	cmd.Flags().AddFlagSet(definition.Flags())
	for _, name := range definition.Required {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(err)
		}
	}
	for _, names := range definition.Exclusive {
		cmd.MarkFlagsMutuallyExclusive(names...)
	}
	if definition.RunE != nil {
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			cmd.SetContext(invocation.WithVersion(ctx, cmd.Root().Version))
			err := definition.RunE(cmd, args)
			if _, ok := errors.AsType[*shared.ExitError](err); ok {
				cmd.Root().SilenceErrors = true
				cmd.Root().SilenceUsage = true
			}
			return err
		}
	}
	for _, child := range definition.Children {
		cmd.AddCommand(Command(child))
	}
	if len(definition.Responses) != 0 {
		contract.RegisterResponse(cmd, definition.Responses...)
	}
	if definition.NoData {
		contract.RegisterNoData(cmd)
	}
	return cmd
}
