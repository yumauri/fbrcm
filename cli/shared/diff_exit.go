package shared

import "github.com/spf13/cobra"

// DiffFoundError returns status 1 when a successful comparison found
// differences. Both human and JSON output use this semantic result status.
func DiffFoundError(cmd *cobra.Command) error {
	if cmd != nil {
		root := cmd.Root()
		root.SilenceErrors = true
		root.SilenceUsage = true
	}
	return WithExitCode(nil, 1)
}
