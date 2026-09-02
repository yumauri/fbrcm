package shared

import (
	"fmt"
	"io"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/invocation"
)

// PrintDiffAndConfirm writes diff text and optionally prompts for confirmation.
func PrintDiffAndConfirm(cmd invocation.Call, yes bool, diffOut io.Writer, diffText, prompt string, destructive bool) (bool, error) {
	if diffText != "" && !MachineMode(cmd) {
		_, _ = fmt.Fprintln(diffOut, diffText)
	}
	if yes {
		return true, nil
	}
	if _, planning, err := PlanOutputPath(cmd); err != nil {
		return false, err
	} else if planning {
		return true, nil
	}
	if err := RequireYesInMachineMode(cmd, yes, strings.TrimSuffix(strings.ToLower(prompt), "?"), destructive); err != nil {
		return false, err
	}
	return RunConfirmationPrompt(prompt, destructive, cmd.ErrOrStderr())
}

// RunConfirmedTargetMutations applies a per-target mutation after optional confirmation.
func RunConfirmedTargetMutations(
	cmd invocation.Call, yes bool,
	diffOut io.Writer,
	targets []ParamTarget,
	process func(target ParamTarget) (applied bool, err error),
) ([]ParamTarget, error) {
	applied := make([]ParamTarget, 0, len(targets))
	for _, target := range targets {
		ok, err := process(target)
		if err != nil {
			return nil, err
		}
		if ok {
			applied = append(applied, target)
		}
	}
	return applied, nil
}

// ParamTargetMutationStep describes one target's diff, prompt, and apply action.
type ParamTargetMutationStep struct {
	DiffText    string
	Prompt      string
	Destructive bool
	Skip        bool
	Apply       func(cfg *firebase.RemoteConfig) (*firebase.RemoteConfig, error)
}

// PrepareParamTargetMutation builds the confirm-and-apply step for one matched target.
type PrepareParamTargetMutation func(target ParamTarget, finalCfg *firebase.RemoteConfig) (ParamTargetMutationStep, error)

// ConfirmParamTargets clones cfg and applies prepare/confirm/apply for each matched target.
func ConfirmParamTargets(
	cmd invocation.Call, label string,
	cfg *firebase.RemoteConfig,
	matched []ParamTarget,
	yes bool,
	diffOut io.Writer,
	prepare PrepareParamTargetMutation,
) ([]ParamTarget, *firebase.RemoteConfig, error) {
	finalCfg, err := firebase.CloneRemoteConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	applied, err := RunConfirmedTargetMutations(cmd, yes, diffOut, matched, func(target ParamTarget) (bool, error) {
		step, err := prepare(target, finalCfg)
		if err != nil {
			return false, err
		}
		if step.Skip {
			return false, nil
		}
		prompt := step.Prompt
		if prompt == "" {
			prompt = fmt.Sprintf("Apply change to %s in %s?", FormatParameterHeader(target.Key, target.Group), label)
		}
		ok, err := PrintDiffAndConfirm(cmd, yes, diffOut, step.DiffText, prompt, step.Destructive)
		if err != nil || !ok {
			return false, err
		}
		if step.Apply != nil {
			nextCfg, err := step.Apply(finalCfg)
			if err != nil {
				return false, err
			}
			if nextCfg != nil {
				finalCfg = nextCfg
			}
		}
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if len(applied) == 0 {
		return nil, finalCfg, nil
	}
	return applied, finalCfg, nil
}
