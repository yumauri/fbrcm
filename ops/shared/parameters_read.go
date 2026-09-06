package shared

import (
	"fmt"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops/invocation"
)

// AddParametersReadFlags exposes the shared Remote Config freshness policy.
func AddParametersReadFlags(cmd invocation.FlagGroups, updateHelp string) {
	cmd.Flags().Bool("update", false, updateHelp)
	cmd.Flags().Bool("cached", false, "Use cached Remote Config even when stale; fetch from Firebase if absent")
	cmd.MarkFlagsMutuallyExclusive("update", "cached")
}

// ReadParametersReadOptions validates freshness options before project resolution.
func ReadParametersReadOptions(cmd invocation.Call) (core.ParametersReadOptions, error) {
	update, _ := cmd.Flags().GetBool("update")
	cached, _ := cmd.Flags().GetBool("cached")
	opts := core.ParametersReadOptions{Update: update, Cached: cached}
	if update && cached {
		return opts, InvalidArgument(fmt.Errorf("--update and --cached are mutually exclusive"))
	}
	if (update || cached) && !core.ExecutionPolicyFromContext(CommandContext(cmd)).ReadLocalState {
		flag := "--update"
		if cached {
			flag = "--cached"
		}
		return opts, InvalidArgument(fmt.Errorf("%s cannot be used with --stateless; Remote Config reads are already live", flag))
	}
	return opts, nil
}
