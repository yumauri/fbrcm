package hooks

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	corehooks "github.com/yumauri/fbrcm/core/hooks"
)

type statusResult struct {
	LocalConfig string   `json:"local_config,omitempty"`
	LocalHooks  bool     `json:"local_hooks"`
	Trusted     bool     `json:"trusted"`
	Fingerprint string   `json:"fingerprint,omitempty"`
	Timeout     string   `json:"timeout"`
	PrePublish  []string `json:"pre_publish"`
	PostPublish []string `json:"post_publish"`
}

func New() *cobra.Command {
	cmd := &cobra.Command{Use: "hooks", Short: "Inspect and trust publication hooks"}
	cmd.AddCommand(newStatusCommand(), newFingerprintCommand(), newTrustCommand(), newUntrustCommand())
	contract.MustRegisterResponsePath(cmd, "status", statusResult{})
	contract.MustRegisterResponsePath(cmd, "fingerprint", fingerprintResult{})
	contract.MustRegisterResponsePath(cmd, "trust", statusResult{})
	contract.MustRegisterResponsePath(cmd, "untrust", untrustResult{})
	return cmd
}

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show effective hooks and local trust status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolution, err := corehooks.Resolve()
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			result := hookStatusResult(resolution)
			if jsonOut {
				return shared.WriteJSON(cmd, result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Local config: %s\n", displayPath(result.LocalConfig))
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Local hooks: %t\n", result.LocalHooks)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Trusted: %t\n", result.Trusted)
			if result.Fingerprint != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Fingerprint: %s\n", result.Fingerprint)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Timeout: %s\n", result.Timeout)
			writeCommands(cmd, "pre_publish", result.PrePublish)
			writeCommands(cmd, "post_publish", result.PostPublish)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print hook status as JSON")
	return cmd
}

func newFingerprintCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fingerprint",
		Short: "Print the effective local hooks fingerprint",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolution, err := corehooks.Resolve()
			if err != nil {
				return err
			}
			if !resolution.LocalHooks {
				return hooksNotConfiguredError()
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, fingerprintResult{Fingerprint: resolution.Fingerprint})
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), resolution.Fingerprint)
			return err
		},
	}
	return cmd
}

type fingerprintResult struct {
	Fingerprint string `json:"fingerprint"`
}

func newTrustCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trust",
		Short: "Trust the current local hook definition",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolution, err := corehooks.Resolve()
			if err != nil {
				return err
			}
			if !resolution.LocalHooks {
				return hooksNotConfiguredError()
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				if err := shared.RequireYesInMachineMode(cmd, yes, "trusting local hooks", false); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Local config: %s\nWorking directory: %s\nFingerprint: %s\n", resolution.LocalPath, filepath.Dir(resolution.LocalPath), resolution.Fingerprint)
				writeCommandsTo(cmd.ErrOrStderr(), "pre_publish", resolution.Hooks.PrePublish)
				writeCommandsTo(cmd.ErrOrStderr(), "post_publish", resolution.Hooks.PostPublish)
				confirm := shared.NewConfirmation("Trust and allow these local hooks to execute?", shared.ConfirmationOptions{})
				confirm.Input = cmd.InOrStdin()
				confirm.Output = cmd.ErrOrStderr()
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}
			trusted, err := corehooks.TrustCurrent()
			if err != nil {
				return err
			}
			if trusted.Fingerprint != resolution.Fingerprint {
				_, _, _ = corehooks.UntrustCurrent()
				return &shared.ConflictError{Code: "hooks.changed", Resource: "hooks", Retryable: true, Remediation: []shared.Remediation{{Description: "review the current hook definition", Strategy: shared.RemediationRunCommand, Argv: []string{"hooks", "status"}}}, Err: fmt.Errorf("local hooks changed while trust was being granted; review them and try again")}
			}
			result := hookStatusResult(trusted)
			if jsonOut {
				return shared.WriteJSON(cmd, result)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Trusted local hooks in %s (%s)\n", trusted.LocalPath, trusted.Fingerprint)
			return err
		},
	}
	shared.AddYesFlag(cmd, "Trust without confirmation")
	cmd.Flags().Bool("json", false, "Print trust result as JSON")
	return cmd
}

func hooksNotConfiguredError() error {
	return &shared.ValidationError{Code: "hooks.not_configured", Source: "configuration", Stage: "selection", Err: fmt.Errorf("the effective configuration does not define local hooks")}
}

func newUntrustCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untrust",
		Short: "Remove trust for the current local config",
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolution, changed, err := corehooks.UntrustCurrent()
			if err != nil {
				return err
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, untrustResult{LocalConfig: resolution.LocalPath, Changed: changed, Trusted: false})
			}
			if changed {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Removed trust for %s\n", resolution.LocalPath)
			} else {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "No stored trust for %s\n", displayPath(resolution.LocalPath))
			}
			return err
		},
	}
	cmd.Flags().Bool("json", false, "Print untrust result as JSON")
	return cmd
}

type untrustResult struct {
	LocalConfig string `json:"local_config"`
	Changed     bool   `json:"changed"`
	Trusted     bool   `json:"trusted"`
}

func hookStatusResult(resolution corehooks.Resolution) statusResult {
	result := statusResult{
		LocalHooks: resolution.LocalHooks,
		Trusted:    resolution.Trusted, Fingerprint: resolution.Fingerprint,
		Timeout:     resolution.Timeout.String(),
		PrePublish:  append([]string(nil), resolution.Hooks.PrePublish...),
		PostPublish: append([]string(nil), resolution.Hooks.PostPublish...),
	}
	if resolution.LocalExists {
		result.LocalConfig = resolution.LocalPath
	}
	return result
}

func displayPath(path string) string {
	if path == "" {
		return "none"
	}
	return path
}

func writeCommands(cmd *cobra.Command, event string, commands []string) {
	writeCommandsTo(cmd.OutOrStdout(), event, commands)
}

func writeCommandsTo(out interface{ Write([]byte) (int, error) }, event string, commands []string) {
	_, _ = fmt.Fprintf(out, "%s hooks (%d):\n", event, len(commands))
	for index, command := range commands {
		_, _ = fmt.Fprintf(out, "  %d. %s\n", index+1, command)
	}
}
