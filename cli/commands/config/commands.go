package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
)

func New() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage global config file",
	}
	configCmd.AddCommand(newPathCommand())
	return configCmd
}

func newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print global config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := coreconfig.GetGlobalConfigFilePath()
			if jsonOut {
				return shared.WriteJSON(cmd, map[string]string{"path": path})
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print path as JSON")
	return cmd
}
