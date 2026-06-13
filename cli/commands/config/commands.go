package config

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	coreconfig "github.com/yumauri/fbrcm/core/config"
)

func New() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage global config file",
	}

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print global config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := coreconfig.GetGlobalConfigFilePath()
			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(map[string]string{"path": path})
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	pathCmd.Flags().Bool("json", false, "Print path as JSON")

	configCmd.AddCommand(pathCmd)
	return configCmd
}
