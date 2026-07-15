package profile

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core/config"
)

func New() *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), config.GetActiveProfileName())
			return nil
		},
	}
	profileCmd.AddCommand(newListCommand(), newSwitchCommand(), newRenameCommand(), newPathCommand(), newPurgeCommand())
	return profileCmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			activeProfile := config.GetActiveProfileName()
			profiles, err := config.ListProfiles()
			if err != nil {
				return err
			}
			if jsonOut {
				return shared.WriteJSON(cmd, newProfileListItems(profiles, activeProfile))
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderProfilesTable(profiles, activeProfile))
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Print profiles as JSON")
	return cmd
}

func newSwitchCommand() *cobra.Command {
	return &cobra.Command{
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
}

func newRenameCommand() *cobra.Command {
	return &cobra.Command{
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
}

func newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path <profile>",
		Short: "Print profile config and cache directory paths",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printProfilePaths(cmd, args[0])
		},
	}
	cmd.Flags().Bool("json", false, "Print paths as JSON")
	return cmd
}

func newPurgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge <profile>",
		Short: "Delete profile config and cache directories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return purgeProfile(cmd, args[0])
		},
	}
	shared.AddYesFlag(cmd, "Skip confirmation dialog")
	return cmd
}

type profilePathItem struct {
	Path string `json:"path"`
}

func printProfilePaths(cmd *cobra.Command, profileName string) error {
	configPath, cachePath, err := profilePaths(profileName)
	if err != nil {
		return err
	}
	jsonOut, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	if jsonOut {
		return shared.WriteJSON(cmd, []profilePathItem{
			{Path: configPath},
			{Path: cachePath},
		})
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), configPath)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), cachePath)
	return nil
}

func purgeProfile(cmd *cobra.Command, profileName string) error {
	configPath, cachePath, err := profilePaths(profileName)
	if err != nil {
		return err
	}
	if err := config.EnsureProfileCanPurge(profileName); err != nil {
		return err
	}
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}
	if !yes {
		confirm := shared.NewConfirmation(
			fmt.Sprintf("Delete profile %s folders?\n%s\n%s", profileName, configPath, cachePath),
			shared.ConfirmationOptions{Destructive: true},
		)
		ok, err := confirm.RunPrompt()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	if err := config.PurgeProfile(profileName); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged profile: %s\n", profileName)
	return nil
}

func profilePaths(profileName string) (string, string, error) {
	configPath, err := config.GetProfileConfigDirPath(profileName)
	if err != nil {
		return "", "", err
	}
	cachePath, err := config.GetProfileCacheDirPath(profileName)
	if err != nil {
		return "", "", err
	}
	return configPath, cachePath, nil
}

type profileListItem struct {
	Profile string `json:"profile"`
	Active  bool   `json:"active"`
}

// newProfileListItems prepares profiles for JSON output.
func newProfileListItems(profiles []string, activeProfile string) []profileListItem {
	items := make([]profileListItem, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, profileListItem{
			Profile: profile,
			Active:  profile == activeProfile,
		})
	}
	return items
}

func renderProfilesTable(profiles []string, activeProfile string) string {
	noColor := clistyles.NoColorEnabled()
	rows := make([][]string, 0, len(profiles))
	profileWidth := lipgloss.Width("Profile")
	activeWidth := lipgloss.Width("Active")
	for _, profile := range profiles {
		activeMarker := ""
		if profile == activeProfile {
			activeMarker = "✓"
		}
		rows = append(rows, []string{profile, activeMarker})
		profileWidth = max(profileWidth, lipgloss.Width(profile))
		activeWidth = max(activeWidth, lipgloss.Width(activeMarker))
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
		Headers("Profile", "Active").
		Rows(rows...).
		Width(profileWidth + activeWidth + 8).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}
