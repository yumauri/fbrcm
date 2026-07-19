package shared

import (
	"errors"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const diffExitCodeFlag = "exit-code"

// AddDiffExitCodeFlag adds the opt-in conventional diff exit status contract.
func AddDiffExitCodeFlag(cmd *cobra.Command) {
	cmd.Flags().Bool(diffExitCodeFlag, false, "Return 1 when differences exist and 2 on errors")
}

// DiffExitCodeEnabled reports whether the command requested conventional diff
// exit statuses.
func DiffExitCodeEnabled(cmd *cobra.Command) bool {
	if cmd == nil || cmd.Flags().Lookup(diffExitCodeFlag) == nil {
		return false
	}
	enabled, err := cmd.Flags().GetBool(diffExitCodeFlag)
	return err == nil && enabled
}

// DiffExitCodeRequested reports whether the command enabled the flag or the
// raw invocation requested it. Inspecting raw arguments keeps the contract
// intact when flag parsing stops at an earlier unknown flag.
func DiffExitCodeRequested(cmd *cobra.Command, args []string) bool {
	if DiffExitCodeEnabled(cmd) {
		return true
	}
	if cmd == nil || cmd.Flags().Lookup(diffExitCodeFlag) == nil {
		return false
	}
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == "--"+diffExitCodeFlag {
			return true
		}
		value, found := strings.CutPrefix(arg, "--"+diffExitCodeFlag+"=")
		if !found {
			continue
		}
		enabled, err := strconv.ParseBool(value)
		if err != nil {
			return true
		}
		return enabled
	}
	return false
}

// DiffFoundError returns exit status 1 when a successful comparison found
// differences and conventional diff exit statuses were requested.
func DiffFoundError(cmd *cobra.Command) error {
	if DiffExitCodeEnabled(cmd) {
		return WithExitCode(nil, 1)
	}
	return nil
}

// DiffCommandError maps a comparison failure to exit status 2 when conventional
// diff exit statuses were requested. Existing explicit exit statuses are kept.
func DiffCommandError(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return err
	}
	if DiffExitCodeEnabled(cmd) {
		return WithExitCode(err, 2)
	}
	return err
}
