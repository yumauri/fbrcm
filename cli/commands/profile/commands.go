package profile

import (
	"encoding/json"
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	clistyles "fbrcm/cli/styles"
	"fbrcm/core/config"
)

// New constructs new and returns the resulting value or error.
func New() *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			profiles, err := config.ListProfiles()
			if err != nil {
				return err
			}
			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(profiles)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderProfilesTable(profiles))
			return nil
		},
	}
	listCmd.Flags().Bool("json", false, "Print profiles as JSON")

	switchCmd := &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch to a profile, creating it if needed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.SwitchProfile(args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✅ switched: %s\n", args[0])
			return nil
		},
	}

	renameCmd := &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename an existing profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.RenameProfile(args[0], args[1]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "➡️ renamed: %s -> %s\n", args[0], args[1])
			return nil
		},
	}

	profileCmd.AddCommand(listCmd, switchCmd, renameCmd)
	return profileCmd
}

// renderProfilesTable renders render profiles table and returns the resulting value or error.
func renderProfilesTable(profiles []string) string {
	noColor := clistyles.NoColorEnabled()
	rows := make([][]string, 0, len(profiles))
	profileWidth := lipgloss.Width("Profile")
	for _, profile := range profiles {
		rows = append(rows, []string{profile})
		profileWidth = max(profileWidth, lipgloss.Width(profile))
	}

	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if noColor {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if row >= 0 && row%2 == 1 {
			style = style.Background(clistyles.ColorRowStripe)
		}
		return style.Foreground(clistyles.PaletteSlateBright)
	}

	tbl := table.New().
		Headers("Profile").
		Rows(rows...).
		Width(profileWidth + 4).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}
