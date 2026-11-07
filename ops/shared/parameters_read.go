package shared

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops/invocation"
)

// ReadParametersWithFallback applies the human inspection convention used by
// get: a failed Firebase read can still show an existing stale template, with
// an explicit cache-stale source. Stateless reads never fall back to disk.
func ReadParametersWithFallback(ctx context.Context, svc *core.Core, projectID string, opts core.ParametersReadOptions) (*core.ParametersCache, string, error) {
	cache, source, err := svc.ReadParameters(ctx, projectID, opts)
	if err == nil || !core.ExecutionPolicyFromContext(ctx).ReadLocalState {
		return cache, source, err
	}
	stale, state, inspectErr := svc.InspectParametersCache(projectID)
	if inspectErr == nil && state != core.ParametersCacheMissing && stale != nil {
		return stale, "cache-stale", nil
	}
	return nil, "", err
}

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
