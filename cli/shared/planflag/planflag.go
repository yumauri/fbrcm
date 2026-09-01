// Package planflag owns the shared --plan-out value contract without creating
// an import cycle between cli/shared and cli/shared/rc.
package planflag

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/machine"
)

// OutputPath returns the validated exclusive plan destination when plan mode
// was explicitly selected.
func OutputPath(cmd *cobra.Command) (string, bool, error) {
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
