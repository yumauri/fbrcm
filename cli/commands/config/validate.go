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
		Short: "Validate layered configuration and keybindings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			scope, err := readConfigScope(cmd, scopeAll, scopeAll, scopeEffective, scopeGlobal, scopeLocal)
			if err != nil {
				return err
			}
			report, err := validationReportForScope(scope)
			if err != nil {
				return err
			}
			if err := writeValidationResult(cmd, jsonOut, report); err != nil {
				return err
			}
			if !report.Valid {
				cmd.Root().SilenceUsage = true
				return shared.WithExitCode(nil, 1)
			}
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print validation result as JSON")
	addScopeFlag(cmd, scopeAll)
	return cmd
}

func validationReportForScope(scope string) (configValidationResult, error) {
	globalPath := coreconfig.GetGlobalConfigFilePath()
	if scope == scopeGlobal {
		state, err := decodeConfigForValidation(globalPath)
		return state.Report, err
	}
	localPath, _, err := coreconfig.GetLocalConfigFilePath()
	if err != nil {
		return configValidationResult{}, err
	}
	if scope == scopeLocal {
		state, err := decodeConfigForValidation(localPath)
		return state.Report, err
	}
	if scope == scopeEffective {
		state, err := loadConfigState()
		if err != nil {
			return configValidationResult{}, err
		}
		return state.Report, nil
	}

	report := configValidationResult{Path: "global and local configuration", Valid: true, Errors: []configDiagnostic{}, Warnings: []configDiagnostic{}}
	for _, source := range []struct {
		name string
		path string
	}{{"global", globalPath}, {"local", localPath}} {
		state, stateErr := decodeConfigForValidation(source.path)
		if stateErr != nil {
			return configValidationResult{}, stateErr
		}
		report.Exists = report.Exists || state.Exists
		for _, diagnostic := range state.Report.Errors {
			diagnostic.Key = scopedDiagnosticKey(source.name, diagnostic.Key)
			report.Errors = append(report.Errors, diagnostic)
		}
		for _, diagnostic := range state.Report.Warnings {
			diagnostic.Key = scopedDiagnosticKey(source.name, diagnostic.Key)
			report.Warnings = append(report.Warnings, diagnostic)
		}
	}
	if len(report.Errors) == 0 {
		state, stateErr := loadConfigState()
		if stateErr != nil {
			return configValidationResult{}, stateErr
		}
		for _, diagnostic := range state.Report.Errors {
			diagnostic.Key = scopedDiagnosticKey(scopeEffective, diagnostic.Key)
			report.Errors = append(report.Errors, diagnostic)
		}
		for _, diagnostic := range state.Report.Warnings {
			diagnostic.Key = scopedDiagnosticKey(scopeEffective, diagnostic.Key)
			report.Warnings = append(report.Warnings, diagnostic)
		}
	}
	sortDiagnostics(report.Errors)
	sortDiagnostics(report.Warnings)
	report.Valid = len(report.Errors) == 0
	return report, nil
}

func scopedDiagnosticKey(scope, key string) string {
	if key == "" {
		return scope
	}
	return scope + "." + key
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
		if report.Path == "global and local configuration" {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "✓ config valid: defaults (global and local files do not exist)")
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
