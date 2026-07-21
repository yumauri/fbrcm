package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
)

func newValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate global configuration and keybindings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			state, err := decodeConfigForValidation(coreconfig.GetGlobalConfigFilePath())
			if err != nil {
				return err
			}
			if err := writeValidationResult(cmd, jsonOut, state.Report); err != nil {
				return err
			}
			if !state.Report.Valid {
				cmd.Root().SilenceUsage = true
				return shared.WithExitCode(nil, 1)
			}
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print validation result as JSON")
	return cmd
}

func writeValidationResult(cmd *cobra.Command, jsonOut bool, report configValidationResult) error {
	if jsonOut {
		return shared.WriteJSON(cmd, report)
	}
	for _, warning := range report.Warnings {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "warning: %s: %s\n", diagnosticKey(warning), warning.Message); err != nil {
			return err
		}
	}
	if report.Valid {
		if report.Exists {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "✓ config valid: %s\n", report.Path)
			return err
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "✓ config valid: defaults (file does not exist at %s)\n", report.Path)
		return err
	}
	for _, diagnostic := range report.Errors {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", diagnosticKey(diagnostic), diagnostic.Message); err != nil {
			return err
		}
	}
	return nil
}

func diagnosticKey(diagnostic configDiagnostic) string {
	if diagnostic.Key == "" {
		return diagnostic.Code
	}
	return diagnostic.Key
}
