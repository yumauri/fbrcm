package meta

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/schemas"
)

type schemaListResult struct {
	Count int      `json:"count"`
	Items []string `json:"items"`
}

func NewCapabilities() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capabilities [command...]",
		Short: "Describe CLI commands for machine discovery",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				capability, err := contract.FindCapability(cmd.Root(), args)
				if err != nil {
					return err
				}
				if contract.Enabled(cmd) {
					return shared.WriteJSON(cmd, capability)
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", capability.ID)
				return err
			}
			index := contract.Capabilities(cmd.Root())
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, index)
			}
			for _, capability := range index.Commands {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", capability.ID, capability.Summary)
			}
			return nil
		},
	}
	contract.RegisterResponse(cmd, contract.CapabilityIndex{}, contract.Capability{})
	return cmd
}

func NewSchema() *cobra.Command {
	cmd := &cobra.Command{Use: "schema", Short: "Inspect embedded CLI JSON Schemas"}
	listCmd := &cobra.Command{Use: "list", Short: "List command schema identifiers", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		ids, err := schemas.List()
		if err != nil {
			return err
		}
		if contract.Enabled(cmd) {
			return shared.WriteJSON(cmd, schemaListResult{Count: len(ids), Items: ids})
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), strings.Join(ids, "\n"))
		return err
	}}
	showCmd := &cobra.Command{Use: "show <schema-id>", Short: "Show one generated CLI JSON Schema", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		schema, err := schemas.ReadByID(args[0])
		if err != nil {
			return &shared.SelectionError{Resource: "schema", Kind: "not_found", Query: args[0], Err: err}
		}
		return shared.WriteJSON(cmd, schema)
	}}
	cmd.AddCommand(listCmd, showCmd)
	contract.RegisterResponse(listCmd, schemaListResult{})
	contract.RegisterResponse(showCmd, contract.JSONDocument{})
	return cmd
}
