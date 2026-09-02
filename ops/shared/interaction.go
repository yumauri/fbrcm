package shared

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/yumauri/fbrcm/ops/invocation"
)

var machineMode atomic.Bool

// InteractionError reports that a command needs a choice machine mode cannot make.
type InteractionError struct {
	Message        string
	Type           string
	Destructive    bool
	RequiredOption string
	SuggestedArgv  []string
}

func (e *InteractionError) Error() string { return e.Message }

func SetMachineMode(enabled bool) { machineMode.Store(enabled) }

func MachineMode(cmd invocation.Call) bool {
	if machineMode.Load() {
		return true
	}
	if cmd == nil {
		return false
	}
	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		flag = cmd.InheritedFlags().Lookup("json")
	}
	return flag != nil && flag.Value.String() == "true"
}

func InteractionRequired(message string, destructive bool, flag string) error {
	interactionType := "input_required"
	if destructive || flag == "--yes" {
		interactionType = "confirmation"
	}
	var suggested []string
	if flag == "--yes" {
		suggested = []string{flag}
	}
	return &InteractionError{Message: message, Type: interactionType, Destructive: destructive, RequiredOption: flag, SuggestedArgv: suggested}
}

// InteractionRequiredWithArguments reports an interaction whose complete,
// directly reusable non-interactive alternative is known.
func InteractionRequiredWithArguments(message, interactionType string, destructive bool, requiredOption string, argv ...string) error {
	return &InteractionError{Message: message, Type: interactionType, Destructive: destructive, RequiredOption: requiredOption, SuggestedArgv: append([]string(nil), argv...)}
}

type interactionReader struct{}

func (interactionReader) Read([]byte) (int, error) {
	return 0, InteractionRequired("interactive input is unavailable in JSON mode", false, "")
}

func NonInteractiveInput() io.Reader { return interactionReader{} }

func RequireYesInMachineMode(cmd invocation.Call, yes bool, action string, destructive bool) error {
	if !MachineMode(cmd) || yes {
		return nil
	}
	return InteractionRequired(fmt.Sprintf("confirmation is required before %s", action), destructive, "--yes")
}
