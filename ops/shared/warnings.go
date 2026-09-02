package shared

import (
	"context"

	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/machine"
)

type MachineWarning = machine.Warning

func WithMachineState(ctx context.Context) context.Context { return machine.WithState(ctx) }

func MarkCommandRun(cmd invocation.Call) {
	machine.FromContext(CommandContext(cmd)).MarkRun()
}

func CommandRunStarted(cmd invocation.Call) bool {
	return machine.FromContext(CommandContext(cmd)).RunStarted()
}

func AddMachineWarning(cmd invocation.Call, warning MachineWarning) {
	machine.FromContext(CommandContext(cmd)).AddWarning(warning)
}

func MachineWarnings(cmd invocation.Call) []MachineWarning {
	return machine.FromContext(CommandContext(cmd)).Warnings()
}
