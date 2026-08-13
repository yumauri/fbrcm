package shared

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/machine"
)

type MachineWarning = machine.Warning

func WithMachineState(ctx context.Context) context.Context { return machine.WithState(ctx) }

func MarkCommandRun(cmd *cobra.Command) {
	machine.FromContext(CommandContext(cmd)).MarkRun()
}

func CommandRunStarted(cmd *cobra.Command) bool {
	return machine.FromContext(CommandContext(cmd)).RunStarted()
}

func AddMachineWarning(cmd *cobra.Command, warning MachineWarning) {
	machine.FromContext(CommandContext(cmd)).AddWarning(warning)
}

func MachineWarnings(cmd *cobra.Command) []MachineWarning {
	return machine.FromContext(CommandContext(cmd)).Warnings()
}
