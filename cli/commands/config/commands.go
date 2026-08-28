package config

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
)

func New() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage global and repository configuration",
	}
	configCmd.AddCommand(newPathCommand(), newShowCommand(), newSetCommand(), newResetCommand(), newValidateCommand(), newEditCommand(runEditor))
	contract.MustRegisterResponsePath(configCmd, "path", configPathResult{})
	contract.MustRegisterResponsePath(configCmd, "show", configShowResult{}, configValueResult{})
	contract.MustRegisterResponsePath(configCmd, "set", configSetResult{})
	contract.MustRegisterResponsePath(configCmd, "reset", configResetResult{})
	contract.MustRegisterResponsePath(configCmd, "validate", configValidationResult{})
	contract.MustRegisterNoDataPath(configCmd, "edit")
	return configCmd
}

func newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print a configuration file path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			scope, err := readConfigScope(cmd, scopeGlobal, scopeGlobal, scopeLocal)
			if err != nil {
				return err
			}
			path := coreconfig.GetGlobalConfigFilePath()
			_, statErr := os.Stat(path)
			exists := statErr == nil
			if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
				return statErr
			}
			if scope == scopeLocal {
				path, exists, err = coreconfig.GetLocalConfigFilePath()
				if err != nil {
					return err
				}
				if !exists {
					cmd.Root().SilenceUsage = true
					return &shared.ValidationError{Code: "configuration.local_not_found", Source: "configuration", Stage: "selection", Target: path, Err: fmt.Errorf("no local config found from the current directory to the filesystem root; create one with `fbrcm config edit --scope local` (candidate: %s)", path)}
				}
			}
			if jsonOut {
				return shared.WriteJSON(cmd, configPathResult{Scope: scope, Path: path, Exists: exists})
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print path as JSON")
	addScopeFlag(cmd, scopeGlobal)
	return cmd
}

type configPathResult struct {
	Scope  string `json:"scope" contract:"enum=global|local"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

const (
	scopeAll       = "all"
	scopeEffective = "effective"
	scopeGlobal    = "global"
	scopeLocal     = "local"
)

func addScopeFlag(cmd *cobra.Command, defaultScope string) {
	cmd.Flags().String("scope", defaultScope, "Configuration scope")
}

func readConfigScope(cmd *cobra.Command, defaultScope string, allowed ...string) (string, error) {
	scope, err := cmd.Flags().GetString("scope")
	if err != nil {
		return "", err
	}
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		scope = defaultScope
	}
	if slices.Contains(allowed, scope) {
		return scope, nil
	}
	return "", shared.InvalidArgument(fmt.Errorf("unsupported config scope %q; use %s", scope, strings.Join(allowed, ", ")))
}
