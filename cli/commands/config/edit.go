package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
)

type editorRunner func(cmd *cobra.Command, editor, path string) error

func newEditCommand(run editorRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit global configuration in a text editor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit, err := cmd.Flags().GetString("editor")
			if err != nil {
				return err
			}
			editor := resolveEditor(explicit)
			state, err := loadConfigStateForEdit()
			if err != nil {
				return err
			}
			if err := coreconfig.EnsurePrivateDir(coreconfig.GetConfigRootDirPath()); err != nil {
				return err
			}
			temp, err := os.CreateTemp(coreconfig.GetConfigRootDirPath(), ".config.toml.edit-*")
			if err != nil {
				return fmt.Errorf("create staged config: %w", err)
			}
			tempPath := temp.Name()
			if err := temp.Chmod(coreconfig.PrivateFileMode); err != nil {
				_ = temp.Close()
				return fmt.Errorf("secure staged config: %w", err)
			}
			if _, err := temp.Write(state); err != nil {
				_ = temp.Close()
				return fmt.Errorf("write staged config: %w", err)
			}
			if err := temp.Close(); err != nil {
				return fmt.Errorf("close staged config: %w", err)
			}

			if err := run(cmd, editor, tempPath); err != nil {
				return fmt.Errorf("editor failed; staged config preserved at %s: %w", tempPath, err)
			}
			validated, err := decodeConfigForValidation(tempPath)
			if err != nil {
				return fmt.Errorf("validate edited config; staged config preserved at %s: %w", tempPath, err)
			}
			if !validated.Report.Valid {
				return fmt.Errorf("edited config is invalid; original was not changed; staged config preserved at %s: %s", tempPath, validationSummary(validated.Report))
			}
			for _, warning := range validated.Report.Warnings {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s: %s\n", diagnosticKey(warning), warning.Message)
			}
			raw, err := os.ReadFile(tempPath)
			if err != nil {
				return fmt.Errorf("read edited config: %w", err)
			}
			if err := coreconfig.SaveAppConfigRaw(raw); err != nil {
				return err
			}
			if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("remove staged config: %w", err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "updated: %s\n", coreconfig.GetGlobalConfigFilePath())
			return err
		},
	}
	cmd.Flags().String("editor", "", "Editor command; overrides FBRCM_EDITOR, VISUAL, and EDITOR")
	return cmd
}

func loadConfigStateForEdit() ([]byte, error) {
	path := coreconfig.GetGlobalConfigFilePath()
	raw, err := os.ReadFile(path)
	if err == nil {
		return raw, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read global config: %w", err)
	}
	state := stateFromConfig(path, false, &coreconfig.AppConfig{})
	template := mutableConfig(state)
	raw, err = coreconfig.MarshalAppConfig(template)
	if err != nil {
		return nil, fmt.Errorf("encode default config: %w", err)
	}
	return raw, nil
}

func resolveEditor(explicit string) string {
	for _, value := range []string{explicit, os.Getenv(env.Editor), os.Getenv("VISUAL"), os.Getenv("EDITOR")} {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	if runtime.GOOS == "windows" {
		return "notepad.exe"
	}
	return "vi"
}

func runEditor(cmd *cobra.Command, editor, path string) error {
	var process *exec.Cmd
	if runtime.GOOS == "windows" {
		process = exec.Command("cmd", "/S", "/C", editor+" "+strconv.Quote(path))
	} else {
		shell := strings.TrimSpace(os.Getenv("SHELL"))
		if shell == "" {
			shell = "/bin/sh"
		}
		process = exec.Command(shell, "-c", `exec `+editor+` "$1"`, "fbrcm-editor", path)
	}
	process.Stdin = cmd.InOrStdin()
	process.Stdout = cmd.OutOrStdout()
	process.Stderr = cmd.ErrOrStderr()
	return process.Run()
}

func validationSummary(report configValidationResult) string {
	parts := make([]string, 0, len(report.Errors))
	for _, diagnostic := range report.Errors {
		parts = append(parts, diagnosticKey(diagnostic)+": "+diagnostic.Message)
	}
	return strings.Join(parts, "; ")
}
