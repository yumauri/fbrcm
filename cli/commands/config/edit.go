package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/progress"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
)

type editorRunner func(cmd *cobra.Command, editor, path string) error

func newEditCommand(run editorRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit global or repository configuration in a text editor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit, err := cmd.Flags().GetString("editor")
			if err != nil {
				return err
			}
			editor := resolveEditor(explicit)
			scope, err := readConfigScope(cmd, scopeGlobal, scopeGlobal, scopeLocal)
			if err != nil {
				return err
			}
			full, err := cmd.Flags().GetBool("full")
			if err != nil {
				return err
			}
			configState, err := loadConfigEditState()
			if err != nil {
				return err
			}
			_, path, _ := scopedConfig(configState, scope)
			state, err := loadConfigStateForEdit(configState, scope, full)
			if err != nil {
				return err
			}
			tempDir := coreconfig.GetConfigRootDirPath()
			if scope == scopeLocal {
				tempDir = filepath.Dir(path)
			}
			if err := os.MkdirAll(tempDir, 0o755); err != nil {
				return err
			}
			if scope == scopeGlobal {
				if err := coreconfig.EnsurePrivateDir(tempDir); err != nil {
					return err
				}
			}
			temp, err := os.CreateTemp(tempDir, ".config.toml.edit-*")
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
			if err := validateScopedCandidate(configState, validated.Stored, scope); err != nil {
				return fmt.Errorf("edited config is invalid; original was not changed; staged config preserved at %s: %w", tempPath, err)
			}
			if scope == scopeLocal {
				err = coreconfig.SaveLocalAppConfigRaw(path, raw)
			} else {
				err = coreconfig.SaveAppConfigRaw(raw)
			}
			if err != nil {
				return err
			}
			if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("remove staged config: %w", err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "updated: %s\n", path)
			return err
		},
	}
	cmd.Flags().String("editor", "", "Editor command; overrides FBRCM_EDITOR, VISUAL, and EDITOR")
	cmd.Flags().Bool("full", false, "Start with a complete generated keybinding template")
	addScopeFlag(cmd, scopeGlobal)
	return cmd
}

func loadConfigEditState() (configState, error) {
	globalPath := coreconfig.GetGlobalConfigFilePath()
	localPath, localFound, err := coreconfig.GetLocalConfigFilePath()
	if err != nil {
		return configState{}, err
	}
	globalState, err := decodeConfigForValidation(globalPath)
	if err != nil {
		return configState{}, err
	}
	localState, err := decodeConfigForValidation(localPath)
	if err != nil {
		return configState{}, err
	}
	global := cloneAppConfig(globalState.Stored)
	local := cloneAppConfig(localState.Stored)
	merged, mergeErr := coreconfig.MergeAppConfigs(global, local)
	if mergeErr != nil {
		merged = &coreconfig.AppConfig{}
	}
	state := stateFromConfig("effective", globalState.Exists || localState.Exists, merged)
	state.GlobalPath = globalPath
	state.GlobalExists = globalState.Exists
	state.Global = global
	state.LocalPath = localPath
	state.LocalExists = localFound
	state.Local = local
	state.Merged = merged
	state.Stored = global
	return state, nil
}

func loadConfigStateForEdit(state configState, scope string, full bool) ([]byte, error) {
	stored, path, exists := scopedConfig(state, scope)
	if full {
		template := cloneAppConfig(stored)
		tuiconfig.MigrateAdminShortcuts(template.Keys)
		if template.PowerlineGlyphs == nil {
			enabled := true
			template.PowerlineGlyphs = &enabled
		}
		template.Keys = tuiconfig.ToConfigMap(tuiconfig.Merge(tuiconfig.DefaultKeyMap(), template.Keys))
		raw, err := coreconfig.MarshalAppConfig(template)
		if err != nil {
			return nil, fmt.Errorf("encode full config template: %w", err)
		}
		return append([]byte("# Complete generated template. Remove entries you do not want to override.\n# View effective bindings with: fbrcm config show keys\n\n"), raw...), nil
	}
	if !exists {
		return []byte("# Add only values you want to override.\n# View valid key names with: fbrcm config show keys\n"), nil
	}
	raw, err := os.ReadFile(path)
	if err == nil {
		return raw, nil
	}
	return nil, fmt.Errorf("read %s config: %w", scope, err)
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

func runEditor(_ *cobra.Command, editor, path string) error {
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
	// Interactive editors must inherit the real process streams. Cobra's
	// coordinated writers expose an Fd for in-process prompts, but os/exec sees
	// them as generic writers and inserts a pipe, which makes Vim report that
	// its output is not connected to a terminal.
	attachEditorTerminal(process)
	progress.Stop()
	return process.Run()
}

func attachEditorTerminal(process *exec.Cmd) {
	process.Stdin = os.Stdin
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr
}

func validationSummary(report configValidationResult) string {
	parts := make([]string, 0, len(report.Errors))
	for _, diagnostic := range report.Errors {
		parts = append(parts, diagnosticKey(diagnostic)+": "+diagnostic.Message)
	}
	return strings.Join(parts, "; ")
}
