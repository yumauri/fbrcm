package projects

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
)

const (
	aliasImportConflictError     = "error"
	aliasImportConflictKeep      = "keep"
	aliasImportConflictOverwrite = "overwrite"
)

type projectAliasImportItem struct {
	Alias             string `json:"alias"`
	PreviousProjectID string `json:"previous_project_id,omitempty"`
	ProjectID         string `json:"project_id"`
	Action            string `json:"action"`
}

type projectAliasImportResult struct {
	From           string                   `json:"from"`
	Destination    string                   `json:"destination"`
	ConflictPolicy string                   `json:"conflict_policy"`
	DryRun         bool                     `json:"dry_run"`
	Changed        bool                     `json:"changed"`
	Items          []projectAliasImportItem `json:"items"`
}

func newAliasesImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import --from <path>",
		Short: "Import Firebase CLI project aliases into .fbrcm.toml",
		Args:  cobra.NoArgs,
		RunE:  runAliasesImport,
	}
	cmd.Flags().String("from", "", "Firebase RC file to import")
	cmd.Flags().String("conflict", aliasImportConflictError, "Conflict policy: error, keep, or overwrite")
	shared.AddDryRunFlag(cmd)
	shared.AddYesFlag(cmd, "Import without confirmation")
	cmd.Flags().Bool("json", false, "Print import result as JSON")
	_ = cmd.MarkFlagRequired("from")
	return cmd
}

func runAliasesImport(cmd *cobra.Command, _ []string) error {
	if coreconfig.LocalConfigDisabled() {
		return fmt.Errorf("project aliases are disabled by --no-local-config or FBRCM_NO_LOCAL_CONFIG")
	}
	from, _ := cmd.Flags().GetString("from")
	from, err := filepath.Abs(from)
	if err != nil {
		return fmt.Errorf("resolve import path: %w", err)
	}
	policy, _ := cmd.Flags().GetString("conflict")
	if !slices.Contains([]string{aliasImportConflictError, aliasImportConflictKeep, aliasImportConflictOverwrite}, policy) {
		return fmt.Errorf("invalid --conflict value %q: use error, keep, or overwrite", policy)
	}
	imported, err := coreconfig.LoadFirebaseProjectAliasesFile(from)
	if err != nil {
		return err
	}
	cfg, destination, err := loadLocalAliasConfig()
	if err != nil {
		return err
	}
	items, conflicts := planProjectAliasImport(coreconfig.CloneProjectAliases(cfg), imported, policy)
	if len(conflicts) > 0 {
		return fmt.Errorf(
			"project alias import has conflicts for %s; use --conflict keep or --conflict overwrite",
			strings.Join(conflicts, ", "),
		)
	}

	changedCount := 0
	for _, item := range items {
		if item.Action == "add" || item.Action == "overwrite" {
			changedCount++
		}
	}
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	jsonOut, _ := cmd.Flags().GetBool("json")
	result := projectAliasImportResult{
		From:           from,
		Destination:    destination,
		ConflictPolicy: policy,
		DryRun:         dryRun,
		Changed:        changedCount > 0,
		Items:          items,
	}

	if !jsonOut {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), renderProjectAliasImportTable(items, shared.TerminalWidth())); err != nil {
			return err
		}
	}
	if changedCount == 0 {
		if jsonOut {
			return shared.WriteJSON(cmd, result)
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "No project aliases to import.")
		return err
	}
	if dryRun {
		if jsonOut {
			return shared.WriteJSON(cmd, result)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Dry run: would import %s.\n", rcdisplay.FormatCount(changedCount, "project alias", "project aliases"))
		return err
	}

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		confirm := shared.NewConfirmation(
			fmt.Sprintf("Import %s into %s?", rcdisplay.FormatCount(changedCount, "project alias", "project aliases"), destination),
			shared.ConfirmationOptions{},
		)
		confirm.Input = cmd.InOrStdin()
		confirm.Output = cmd.ErrOrStderr()
		ok, promptErr := confirm.RunPrompt()
		if promptErr != nil || !ok {
			return promptErr
		}
	}
	for _, item := range items {
		if item.Action != "add" && item.Action != "overwrite" {
			continue
		}
		if _, _, err := coreconfig.SetProjectAlias(cfg, item.Alias, item.ProjectID); err != nil {
			return err
		}
	}
	if err := saveLocalAliasConfig(destination, cfg); err != nil {
		return err
	}
	if jsonOut {
		return shared.WriteJSON(cmd, result)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Imported %s into %s.\n", rcdisplay.FormatCount(changedCount, "project alias", "project aliases"), destination)
	return err
}

func planProjectAliasImport(current, imported map[string]string, policy string) ([]projectAliasImportItem, []string) {
	aliases := make([]string, 0, len(imported))
	for alias := range imported {
		aliases = append(aliases, alias)
	}
	slices.Sort(aliases)
	items := make([]projectAliasImportItem, 0, len(aliases))
	conflicts := make([]string, 0)
	for _, alias := range aliases {
		projectID := imported[alias]
		previous, exists := current[alias]
		action := "add"
		switch {
		case !exists:
		case previous == projectID:
			action = "unchanged"
		case policy == aliasImportConflictKeep:
			action = "keep"
		case policy == aliasImportConflictOverwrite:
			action = "overwrite"
		default:
			action = "conflict"
			conflicts = append(conflicts, alias)
		}
		items = append(items, projectAliasImportItem{
			Alias:             alias,
			PreviousProjectID: previous,
			ProjectID:         projectID,
			Action:            action,
		})
	}
	return items, conflicts
}

func renderProjectAliasImportTable(items []projectAliasImportItem, terminalWidth int) string {
	headers := []string{"Alias", "Current Project ID", "Imported Project ID", "Action"}
	widths := shared.HeaderWidths(headers)
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		previous := item.PreviousProjectID
		if previous == "" {
			previous = "—"
		}
		row := []string{item.Alias, previous, item.ProjectID, item.Action}
		rows = append(rows, row)
		shared.UpdateTableWidths(widths, row)
	}
	if terminalWidth > 0 && shared.TableWidth(widths) > terminalWidth {
		for _, preserveHeaders := range []bool{true, false} {
			for _, column := range []int{1, 2, 0} {
				minimum := 1
				if preserveHeaders {
					minimum = lipgloss.Width(headers[column])
				}
				for widths[column] > minimum && shared.TableWidth(widths) > terminalWidth {
					widths[column]--
				}
			}
		}
		for column := range headers {
			headers[column] = ansi.Truncate(headers[column], widths[column], "…")
		}
		for row := range rows {
			for column := range rows[row] {
				rows[row][column] = ansi.Truncate(rows[row][column], widths[column], "…")
			}
		}
	}
	return shared.StyledTable(headers, rows, widths, nil, nil)
}
