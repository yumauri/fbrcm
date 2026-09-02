// Package planflag owns the shared --plan-out value contract without creating
// an import cycle between ops/shared and ops/shared/rc.
package planflag

import (
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/machine"
)

// OutputPath returns the validated exclusive plan destination when plan mode
// was explicitly selected.
func OutputPath(cmd invocation.Call) (string, bool, error) {
	if cmd == nil || cmd.Flags().Lookup("plan-out") == nil || !cmd.Flags().Changed("plan-out") {
		return "", false, nil
	}
	path, err := cmd.Flags().GetString("plan-out")
	if err != nil {
		return "", false, err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return "", false, machine.InvalidArgument(fmt.Errorf("--plan-out requires a non-empty path"))
	}
	if path == "-" {
		return "", false, machine.InvalidArgument(fmt.Errorf("--plan-out does not support stdout; choose a private file path"))
	}
	return path, true, nil
}
