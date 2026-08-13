package shared

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ConfirmFileOverwrite determines whether a destination may be written and
// whether the write must explicitly replace an existing file.
func ConfirmFileOverwrite(cmd *cobra.Command, path string, yes bool) (overwrite, proceed bool, err error) {
	return confirmFileOverwrite(cmd, path, yes, true)
}

// ConfirmFileOverwriteWithoutBypass protects an existing destination for a
// command whose public interface intentionally has no confirmation-bypass
// option. Machine mode reports the required interaction instead of prompting.
func ConfirmFileOverwriteWithoutBypass(cmd *cobra.Command, path string) (overwrite, proceed bool, err error) {
	return confirmFileOverwrite(cmd, path, false, false)
}

func confirmFileOverwrite(cmd *cobra.Command, path string, yes, allowBypass bool) (overwrite, proceed bool, err error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, true, nil
		}
		return false, false, fmt.Errorf("inspect destination file: %w", err)
	}
	if info.IsDir() {
		return false, false, fmt.Errorf("destination path is a directory: %s", path)
	}
	if allowBypass && yes {
		return true, true, nil
	}
	if MachineMode(cmd) {
		flag := ""
		if allowBypass {
			flag = "--yes"
		}
		if flag != "" {
			return false, false, InteractionRequiredWithArguments("confirmation is required before overwriting existing file "+path, "destination_conflict", true, flag, flag)
		}
		return false, false, InteractionRequiredWithArguments("confirmation is required before overwriting existing file "+path, "destination_conflict", true, "")
	}

	confirm := NewConfirmation(
		fmt.Sprintf("Overwrite existing file %s?", path),
		ConfirmationOptions{Destructive: true},
	)
	confirm.Input = cmd.InOrStdin()
	confirm.Output = cmd.ErrOrStderr()
	ok, err := confirm.RunPrompt()
	if err != nil {
		return false, false, err
	}
	return ok, ok, nil
}
