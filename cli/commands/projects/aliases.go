package projects

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	coreconfig "github.com/yumauri/fbrcm/core/config"
)

type projectAliasRow struct {
	Alias     string                        `json:"alias"`
	ProjectID string                        `json:"project_id"`
	Source    coreconfig.ProjectAliasSource `json:"source"`
}

type projectAliasSetResult struct {
	Alias             string                        `json:"alias"`
	PreviousProjectID string                        `json:"previous_project_id,omitempty"`
	ProjectID         string                        `json:"project_id"`
	Changed           bool                          `json:"changed"`
	Source            coreconfig.ProjectAliasSource `json:"source"`
}

type projectAliasRemoveResult struct {
	Alias             string                        `json:"alias"`
	PreviousProjectID string                        `json:"previous_project_id,omitempty"`
	Status            string                        `json:"status"`
	Changed           bool                          `json:"changed"`
	Source            coreconfig.ProjectAliasSource `json:"source,omitempty"`
	RemainingSource   coreconfig.ProjectAliasSource `json:"remaining_source,omitempty"`
}

func newAliasesCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "aliases", Short: "Manage repository project aliases"}
	cmd.AddCommand(newAliasesListCommand(), newAliasesSetCommand(), newAliasesRemoveCommand(), newAliasesImportCommand())
	return cmd
}

func newAliasesListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repository project aliases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			registry, err := loadAliasRegistryForCommand()
			if err != nil {
				return err
			}
			entries := registry.SortedEntries()
			rows := make([]projectAliasRow, 0, len(entries))
			for _, entry := range entries {
				rows = append(rows, projectAliasRow{Alias: entry.Alias, ProjectID: entry.ProjectID, Source: entry.Source})
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, rows)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), renderProjectAliasesTable(rows, shared.TerminalWidth()))
			return err
		},
	}
	cmd.Flags().Bool("json", false, "Print project aliases as JSON")
	return cmd
}

func newAliasesSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <alias> <project-id>",
		Short: "Set a repository project alias",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := coreconfig.ValidateProjectAliasName(args[0]); err != nil {
				return err
			}
			if err := coreconfig.ValidateProjectAliasProjectID(args[1]); err != nil {
				return fmt.Errorf("project alias %q: %w", args[0], err)
			}
			registry, err := loadAliasRegistryForCommand()
			if err != nil {
				return err
			}
			entry, exists := registry.Entries[args[0]]
			if exists && entry.ProjectID == args[1] {
				result := projectAliasSetResult{Alias: args[0], PreviousProjectID: entry.ProjectID, ProjectID: args[1], Changed: false, Source: entry.Source}
				jsonOut, _ := cmd.Flags().GetBool("json")
				if jsonOut {
					return shared.WriteJSON(cmd, result)
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "unchanged project alias: %s -> %s\n", args[0], args[1])
				return err
			}
			if exists && entry.Source != coreconfig.ProjectAliasSourceFBRCM {
				return firebaseAliasMutationError("change", entry, registry.FirebasePath)
			}
			cfg, path, err := loadLocalAliasConfig()
			if err != nil {
				return err
			}
			previous := coreconfig.CloneProjectAliases(cfg)[args[0]]
			if previous != "" && previous != args[1] {
				yes, _ := cmd.Flags().GetBool("yes")
				if !yes {
					confirm := shared.NewConfirmation(
						fmt.Sprintf("Change project alias %s from %s to %s?", args[0], previous, args[1]),
						shared.ConfirmationOptions{},
					)
					confirm.Input = cmd.InOrStdin()
					confirm.Output = cmd.ErrOrStderr()
					ok, promptErr := confirm.RunPrompt()
					if promptErr != nil || !ok {
						return promptErr
					}
				}
			}
			previous, changed, err := coreconfig.SetProjectAlias(cfg, args[0], args[1])
			if err != nil {
				return err
			}
			if changed {
				if err := saveLocalAliasConfig(path, cfg); err != nil {
					return err
				}
			}
			result := projectAliasSetResult{Alias: args[0], PreviousProjectID: previous, ProjectID: args[1], Changed: changed, Source: coreconfig.ProjectAliasSourceFBRCM}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, result)
			}
			verb := "unchanged"
			if changed && previous == "" {
				verb = "added"
			} else if changed {
				verb = "updated"
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s project alias: %s -> %s\n", verb, args[0], args[1])
			return err
		},
	}
	shared.AddYesFlag(cmd, "Replace an existing alias without confirmation")
	cmd.Flags().Bool("json", false, "Print update result as JSON")
	return cmd
}

func newAliasesRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <alias>",
		Short: "Remove a repository project alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := coreconfig.ValidateProjectAliasName(args[0]); err != nil {
				return err
			}
			registry, err := loadAliasRegistryForCommand()
			if err != nil {
				return err
			}
			entry, exists := registry.Entries[args[0]]
			if exists && entry.Source == coreconfig.ProjectAliasSourceFirebase {
				return firebaseAliasMutationError("remove", entry, registry.FirebasePath)
			}
			cfg, path, err := loadLocalAliasConfig()
			if err != nil {
				return err
			}
			previous := coreconfig.CloneProjectAliases(cfg)[args[0]]
			if previous != "" {
				yes, _ := cmd.Flags().GetBool("yes")
				if !yes {
					message := fmt.Sprintf("Remove project alias %s for %s?", args[0], previous)
					if exists && entry.Source == coreconfig.ProjectAliasSourceBoth {
						message = fmt.Sprintf("Remove the .fbrcm.toml definition of project alias %s? It will remain available from .firebaserc.", args[0])
					}
					confirm := shared.NewConfirmation(
						message,
						shared.ConfirmationOptions{Destructive: true},
					)
					confirm.Input = cmd.InOrStdin()
					confirm.Output = cmd.ErrOrStderr()
					ok, promptErr := confirm.RunPrompt()
					if promptErr != nil || !ok {
						return promptErr
					}
				}
			}
			previous, changed, err := coreconfig.RemoveProjectAlias(cfg, args[0])
			if err != nil {
				return err
			}
			if changed {
				if err := saveLocalAliasConfig(path, cfg); err != nil {
					return err
				}
			}
			status := "not_found"
			remainingSource := coreconfig.ProjectAliasSource("")
			if changed {
				status = "removed"
				if exists && entry.Source == coreconfig.ProjectAliasSourceBoth {
					status = "removed_native"
					remainingSource = coreconfig.ProjectAliasSourceFirebase
				}
			}
			result := projectAliasRemoveResult{Alias: args[0], PreviousProjectID: previous, Status: status, Changed: changed, Source: entry.Source, RemainingSource: remainingSource}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, result)
			}
			if status == "removed_native" {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "removed native project alias: %s -> %s; alias remains available from .firebaserc\n", args[0], previous)
			} else if changed {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "removed project alias: %s -> %s\n", args[0], previous)
			} else {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "project alias not configured: %s\n", args[0])
			}
			return err
		},
	}
	shared.AddYesFlag(cmd, "Remove the alias without confirmation")
	cmd.Flags().Bool("json", false, "Print removal result as JSON")
	return cmd
}

func loadAliasRegistryForCommand() (coreconfig.ProjectAliasRegistry, error) {
	if coreconfig.LocalConfigDisabled() {
		return coreconfig.ProjectAliasRegistry{}, fmt.Errorf("project aliases are disabled by --no-local-config or FBRCM_NO_LOCAL_CONFIG")
	}
	return coreconfig.LoadProjectAliasRegistry()
}

func firebaseAliasMutationError(action string, entry coreconfig.ProjectAliasEntry, path string) error {
	return fmt.Errorf(
		"cannot %s project alias %q because it is defined in %s; use Firebase CLI or edit that file",
		action,
		entry.Alias,
		path,
	)
}

func loadLocalAliasConfig() (*coreconfig.AppConfig, string, error) {
	if coreconfig.LocalConfigDisabled() {
		return nil, "", fmt.Errorf("project aliases are disabled by --no-local-config or FBRCM_NO_LOCAL_CONFIG")
	}
	resolved, err := coreconfig.ResolveAppConfig()
	if err != nil {
		return nil, "", err
	}
	return coreconfig.CloneAppConfig(resolved.Local.Config), resolved.Local.Path, nil
}

func saveLocalAliasConfig(path string, cfg *coreconfig.AppConfig) error {
	raw, err := coreconfig.MarshalAppConfig(cfg)
	if err != nil {
		return fmt.Errorf("encode local config: %w", err)
	}
	return coreconfig.SaveLocalAppConfigRaw(path, raw)
}

func renderProjectAliasesTable(rows []projectAliasRow, terminalWidth int) string {
	headers := []string{"Alias", "Project ID", "Source"}
	widths := []int{lipgloss.Width(headers[0]), lipgloss.Width(headers[1]), lipgloss.Width(headers[2])}
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{row.Alias, row.ProjectID, string(row.Source)})
		widths[0] = max(widths[0], lipgloss.Width(row.Alias))
		widths[1] = max(widths[1], lipgloss.Width(row.ProjectID))
		widths[2] = max(widths[2], lipgloss.Width(string(row.Source)))
	}
	tableWidth := func() int { return widths[0] + widths[1] + widths[2] + 10 }
	if terminalWidth > 0 {
		for _, preserveHeaders := range []bool{true, false} {
			for _, column := range []int{0, 1} {
				minimum := 1
				if preserveHeaders {
					minimum = lipgloss.Width(headers[column])
				}
				for widths[column] > minimum && tableWidth() > terminalWidth {
					widths[column]--
				}
			}
		}
		for column := range headers {
			headers[column] = ansi.Truncate(headers[column], widths[column], "…")
		}
		for row := range tableRows {
			for column := range tableRows[row] {
				tableRows[row][column] = ansi.Truncate(tableRows[row][column], widths[column], "…")
			}
		}
	}
	noColor := clistyles.NoColorEnabled()
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
		if col == 0 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateDim)
	}
	tbl := table.New().Headers(headers...).Rows(tableRows...).Width(tableWidth()).
		Border(lipgloss.NormalBorder()).BorderHeader(true).BorderRow(false).StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}
