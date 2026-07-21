package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
)

type configShowResult struct {
	Path   string                `json:"path"`
	Exists bool                  `json:"exists"`
	Config *coreconfig.AppConfig `json:"config"`
}

type configValueResult struct {
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Source string `json:"source"`
}

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [key]",
		Short: "Show effective global configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			state, err := loadConfigState()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				value, source, err := configValue(state, args[0])
				if err != nil {
					return err
				}
				if jsonOut {
					return shared.WriteJSON(cmd, configValueResult{Key: args[0], Value: value, Source: source})
				}
				return writeConfigValue(cmd, args[0], value)
			}
			if jsonOut {
				return shared.WriteJSON(cmd, configShowResult{Path: state.Path, Exists: state.Exists, Config: state.Effective})
			}
			raw, err := coreconfig.MarshalAppConfig(state.Effective)
			if err != nil {
				return fmt.Errorf("encode effective config: %w", err)
			}
			_, err = cmd.OutOrStdout().Write(raw)
			return err
		},
	}
	cmd.Flags().Bool("json", false, "Print configuration as JSON")
	return cmd
}

func writeConfigValue(cmd *cobra.Command, key string, value any) error {
	switch value := value.(type) {
	case string:
		_, err := fmt.Fprintln(cmd.OutOrStdout(), value)
		return err
	case bool:
		_, err := fmt.Fprintln(cmd.OutOrStdout(), value)
		return err
	default:
		nested := value
		parts := strings.Split(key, ".")
		for _, part := range slices.Backward(parts) {
			nested = map[string]any{part: nested}
		}
		raw, err := coreconfig.MarshalTOML(nested)
		if err != nil {
			return fmt.Errorf("encode selected config: %w", err)
		}
		_, err = cmd.OutOrStdout().Write(raw)
		return err
	}
}
