package profile

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/core/config"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
)

func New() *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			active := config.GetActiveProfileName()
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, profileCurrentResult{Profile: active})
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), active)
			return nil
		},
	}
	profileCmd.AddCommand(newListCommand(), newSwitchCommand(), newRenameCommand(), newPathCommand(), newDeleteCommand())
	contract.RegisterResponse(profileCmd, profileCurrentResult{})
	contract.MustRegisterResponsePath(profileCmd, "list", []profileListItem{})
	contract.MustRegisterResponsePath(profileCmd, "switch", profileSwitchResult{})
	contract.MustRegisterResponsePath(profileCmd, "rename", profileRenameResult{})
	contract.MustRegisterResponsePath(profileCmd, "path", []profilePathItem{})
	contract.MustRegisterResponsePath(profileCmd, "delete", profileDeleteResult{})
	return profileCmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List profiles",
		Args:  cobra.NoArgs,
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
			before := config.GetActiveProfileName()
			if err := config.SwitchProfile(args[0]); err != nil {
				return err
			}
			effective := config.GetActiveProfileName()
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, profileSwitchResult{RequestedProfile: args[0], PreviousProfile: before, EffectiveProfile: effective, Changed: before != args[0], Overridden: effective != args[0]})
			}
			if effective != args[0] {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "✅ switched global profile: %s\neffective profile remains %s because repository configuration overrides it\n", args[0], effective)
				return nil
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
			changed := args[0] != args[1]
			if err := config.RenameProfile(args[0], args[1]); err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, profileRenameResult{OldProfile: args[0], NewProfile: args[1], EffectiveProfile: config.GetActiveProfileName(), Changed: changed})
			}
			if !changed {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🤷 profile unchanged: %s\n", args[0])
				return nil
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

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <profile>",
		Short: "Delete profile config and cache directories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteProfile(cmd, args[0])
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

func deleteProfile(cmd *cobra.Command, profileName string) error {
	configPath, cachePath, err := profilePaths(profileName)
	if err != nil {
		return err
	}
	if err := config.EnsureProfileCanDelete(profileName); err != nil {
		return err
	}
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}
	if !yes {
		if err := shared.RequireYesInMachineMode(cmd, yes, "deleting profile "+profileName, true); err != nil {
			return err
		}
		confirm := shared.NewConfirmation(
			fmt.Sprintf("Delete profile %s folders?\n%s\n%s", profileName, configPath, cachePath),
			shared.ConfirmationOptions{Destructive: true},
		)
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
	if err := config.DeleteProfile(profileName); err != nil {
		return err
	}
	if contract.Enabled(cmd) {
		return shared.WriteJSON(cmd, profileDeleteResult{Profile: profileName, Status: "deleted", DeletedPaths: []string{configPath, cachePath}})
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 deleted profile: %s\n", profileName)
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

type profileCurrentResult struct {
	Profile string `json:"profile"`
}

type profileSwitchResult struct {
	RequestedProfile string `json:"requested_profile"`
	PreviousProfile  string `json:"previous_profile"`
	EffectiveProfile string `json:"effective_profile"`
	Changed          bool   `json:"changed"`
	Overridden       bool   `json:"overridden"`
}

type profileRenameResult struct {
	OldProfile       string `json:"old_profile"`
	NewProfile       string `json:"new_profile"`
	EffectiveProfile string `json:"effective_profile"`
	Changed          bool   `json:"changed"`
}

type profileDeleteResult struct {
	Profile      string   `json:"profile"`
	Status       string   `json:"status" contract:"enum=deleted"`
	DeletedPaths []string `json:"deleted_paths"`
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
