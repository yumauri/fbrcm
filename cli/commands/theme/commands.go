package theme

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	corestyles "github.com/yumauri/fbrcm/core/styles"
)

const (
	builtInThemeLabel       = "built-in"
	paletteSwatchCount      = 8
	paletteSwatchWidth      = 2
	paletteSwatchGlyph      = "▇"
	themeTablePaletteChrome = 11
)

func New() *cobra.Command {
	themeCmd := &cobra.Command{
		Use:   "theme",
		Short: "Manage themes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			current, err := currentTheme()
			if err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, current)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), renderCurrentTheme(current.Theme, shared.TerminalWidth()))
			return err
		},
	}
	themeCmd.AddCommand(
		newListCommand(),
		newSwitchCommand(),
		newResetCommand(),
		newDeleteCommand(),
		newPathCommand(),
		newRenameCommand(),
		newImportCommand(http.DefaultClient),
	)
	contract.RegisterResponse(themeCmd, themeCurrentResult{})
	contract.MustRegisterResponsePath(themeCmd, "list", []themeListItem{})
	contract.MustRegisterResponsePath(themeCmd, "switch", themeSwitchResult{})
	contract.MustRegisterResponsePath(themeCmd, "reset", themeResetResult{})
	contract.MustRegisterResponsePath(themeCmd, "delete", themeDeleteResult{})
	contract.MustRegisterResponsePath(themeCmd, "path", themePathResult{})
	contract.MustRegisterResponsePath(themeCmd, "rename", themeRenameResult{})
	contract.MustRegisterResponsePath(themeCmd, "import", themeImportResult{}, themeBatchImportResult{})
	return themeCmd
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed themes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			current, err := currentTheme()
			if err != nil {
				return err
			}
			themes, err := coreconfig.ListThemes()
			if err != nil {
				return err
			}
			items := newThemeListItems(themes, current.Theme)
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, items)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), renderThemesTable(items))
			return err
		},
	}
}

func newSwitchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <name>",
		Short: "Select an installed theme",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := readScope(cmd)
			if err != nil {
				return err
			}
			before, err := currentTheme()
			if err != nil {
				return err
			}
			previousSelection, err := scopedSelection(scope)
			if err != nil {
				return err
			}
			requested := args[0]
			desiredSelection := requested
			if requested == coreconfig.BuiltInThemeName {
				desiredSelection = ""
			}
			if err := coreconfig.SetConfiguredTheme(requested, scope); err != nil {
				return err
			}
			after, err := currentTheme()
			if err != nil {
				return err
			}
			result := themeSwitchResult{
				Scope: scope, RequestedTheme: requested, PreviousTheme: displayedTheme(previousSelection),
				PreviousEffectiveTheme: displayedTheme(before.Theme), EffectiveTheme: displayedTheme(after.Theme), Changed: previousSelection != desiredSelection,
				Overridden: after.Theme != desiredSelection,
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, result)
			}
			if result.Overridden {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "switched %s theme: %s\neffective theme remains %s because another configuration layer overrides it\n", scope, requested, displayedTheme(after.Theme))
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "switched: %s\n", requested)
			return err
		},
	}
	cmd.Flags().String("scope", coreconfig.ThemeScopeGlobal, "Configuration scope (global or local)")
	return cmd
}

func newResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset a theme selection to the built-in palette",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			scope, err := readScope(cmd)
			if err != nil {
				return err
			}
			previous, err := scopedSelection(scope)
			if err != nil {
				return err
			}
			if err := coreconfig.ResetConfiguredTheme(scope); err != nil {
				return err
			}
			current, err := currentTheme()
			if err != nil {
				return err
			}
			changed := previous != ""
			status := "unchanged"
			if changed {
				status = "reset"
			}
			result := themeResetResult{
				Scope: scope, PreviousTheme: displayedTheme(previous),
				EffectiveTheme: displayedTheme(current.Theme), Status: status, Changed: changed,
			}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, result)
			}
			if !changed {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s theme already uses its default\n", scope)
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "reset %s theme; effective theme: %s\n", scope, displayedTheme(current.Theme))
			return err
		},
	}
	cmd.Flags().String("scope", coreconfig.ThemeScopeGlobal, "Configuration scope (global or local)")
	return cmd
}

func newPathCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "path <theme>",
		Short: "Print a theme file path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := coreconfig.GetThemeFilePath(args[0])
			if err != nil {
				return err
			}
			exists := false
			if info, statErr := os.Lstat(path); statErr == nil {
				exists = info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
			} else if !errors.Is(statErr, os.ErrNotExist) {
				return fmt.Errorf("inspect theme path %s: %w", path, statErr)
			}
			result := themePathResult{Theme: args[0], Path: path, Exists: exists}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, result)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	}
}

func newRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename an installed theme",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			changed := args[0] != args[1]
			if err := coreconfig.RenameTheme(args[0], args[1]); err != nil {
				return err
			}
			current, err := currentTheme()
			if err != nil {
				return err
			}
			result := themeRenameResult{OldTheme: args[0], NewTheme: args[1], EffectiveTheme: current.Theme, Changed: changed}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, result)
			}
			if !changed {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "theme unchanged: %s\n", args[0])
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "renamed: %s -> %s\n", args[0], args[1])
			return err
		},
	}
}

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <theme>",
		Short: "Delete an installed theme",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := coreconfig.GetThemeFilePath(args[0])
			if err != nil {
				return err
			}
			if err := coreconfig.EnsureThemeCanDelete(args[0]); err != nil {
				return err
			}
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				if err := shared.RequireYesInMachineMode(cmd, yes, "deleting theme "+args[0], true); err != nil {
					return err
				}
				confirm := shared.NewConfirmation(fmt.Sprintf("Delete theme %s?\n%s", args[0], path), shared.ConfirmationOptions{Destructive: true})
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
			if err := coreconfig.DeleteTheme(args[0]); err != nil {
				return err
			}
			result := themeDeleteResult{Theme: args[0], Status: "deleted", Path: path}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, result)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "deleted theme: %s\n", args[0])
			return err
		},
	}
	shared.AddYesFlag(cmd, "Delete without confirmation")
	return cmd
}

func currentTheme() (themeCurrentResult, error) {
	resolved, err := coreconfig.ResolveAppConfig()
	if err != nil {
		return themeCurrentResult{}, err
	}
	result := themeCurrentResult{Theme: resolved.Effective.Theme, Source: "default", BuiltIn: resolved.Effective.Theme == ""}
	if resolved.Local.Config.Theme != "" {
		result.Source = coreconfig.ThemeScopeLocal
	} else if resolved.Global.Config.Theme != "" {
		result.Source = coreconfig.ThemeScopeGlobal
	}
	return result, nil
}

func scopedSelection(scope string) (string, error) {
	resolved, err := coreconfig.ResolveAppConfig()
	if err != nil {
		return "", err
	}
	if scope == coreconfig.ThemeScopeLocal {
		return resolved.Local.Config.Theme, nil
	}
	return resolved.Global.Config.Theme, nil
}

func readScope(cmd *cobra.Command) (string, error) {
	scope, err := cmd.Flags().GetString("scope")
	if err != nil {
		return "", err
	}
	scope = strings.ToLower(strings.TrimSpace(scope))
	if slices.Contains([]string{coreconfig.ThemeScopeGlobal, coreconfig.ThemeScopeLocal}, scope) {
		return scope, nil
	}
	return "", shared.InvalidArgument(fmt.Errorf("unsupported theme scope %q; use global or local", scope))
}

func displayedTheme(name string) string {
	if name == "" {
		return builtInThemeLabel
	}
	return name
}

func newThemeListItems(themes []string, active string) []themeListItem {
	items := make([]themeListItem, 0, len(themes)+1)
	items = append(items, themeListItem{Theme: coreconfig.BuiltInThemeName, Active: active == "", BuiltIn: true})
	for _, name := range themes {
		items = append(items, themeListItem{Theme: name, Active: name == active, BuiltIn: false})
	}
	return items
}

func renderThemesTable(items []themeListItem) string {
	return renderThemesTableAtWidth(items, shared.TerminalWidth())
}

func renderThemesTableAtWidth(items []themeListItem, terminalWidth int) string {
	noColor := clistyles.NoColorEnabled()
	palettes := loadThemePalettePreviews(items, noColor)
	rows := make([][]string, 0, len(items))
	themeWidth := lipgloss.Width("Theme")
	activeWidth := lipgloss.Width("Active")
	paletteWidth := paletteSwatchCount * paletteSwatchWidth
	for _, item := range items {
		active := ""
		if item.Active {
			active = "✓"
		}
		row := []string{item.Theme, active}
		if !noColor {
			row = []string{item.Theme, active, renderPalettePreview(palettes[item.Theme], paletteWidth)}
		}
		rows = append(rows, row)
		themeWidth = max(themeWidth, lipgloss.Width(item.Theme))
		activeWidth = max(activeWidth, lipgloss.Width(active))
	}
	overhead := 8
	paletteContentWidth := 0
	if !noColor {
		overhead = themeTablePaletteChrome
		paletteContentWidth = paletteWidth
	}
	if naturalWidth := themeWidth + activeWidth + overhead + paletteContentWidth; terminalWidth > 0 && naturalWidth > terminalWidth {
		if !noColor && terminalWidth-activeWidth-paletteWidth-overhead < 1 {
			paletteWidth = max(paletteSwatchWidth, terminalWidth-activeWidth-1-overhead)
			paletteWidth -= paletteWidth % paletteSwatchWidth
			paletteContentWidth = paletteWidth
			for index, item := range items {
				rows[index][2] = renderPalettePreview(palettes[item.Theme], paletteWidth)
			}
		}
		themeWidth = max(1, terminalWidth-activeWidth-overhead-paletteContentWidth)
		for index := range rows {
			rows[index][0] = ansi.Truncate(rows[index][0], themeWidth, "…")
		}
	}
	themeHeader := ansi.Truncate("Theme", themeWidth, "…")
	headers := []string{themeHeader, "Active"}
	width := themeWidth + activeWidth + overhead
	if !noColor {
		headers = []string{themeHeader, "Active", ansi.Truncate("Palette", paletteWidth, "…")}
		width += paletteWidth
	}
	styleFunc := func(row, _ int) lipgloss.Style {
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
	tbl := table.New().Headers(headers...).Rows(rows...).Width(width).
		Border(lipgloss.NormalBorder()).BorderHeader(true).BorderRow(false).StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func loadThemePalettePreviews(items []themeListItem, disabled bool) map[string]corestyles.Palette {
	palettes := make(map[string]corestyles.Palette, len(items))
	if disabled {
		return palettes
	}
	for _, item := range items {
		if item.BuiltIn {
			palettes[item.Theme] = corestyles.DefaultPalette()
			continue
		}
		resolved, err := coreconfig.LoadTheme(item.Theme)
		if err == nil {
			palettes[item.Theme] = resolved.Palette
		}
	}
	return palettes
}

func renderPalettePreview(palette corestyles.Palette, width int) string {
	if width < paletteSwatchWidth {
		return ""
	}
	if palette == nil {
		return ansi.Truncate("unavailable", width, "…")
	}
	previewTokens := corestyles.PreviewTokens()
	count := min(len(previewTokens), width/paletteSwatchWidth)
	var preview strings.Builder
	for _, token := range previewTokens[:count] {
		preview.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(palette[token])).Render(strings.Repeat(paletteSwatchGlyph, paletteSwatchWidth)))
	}
	return preview.String()
}

func renderCurrentTheme(name string, terminalWidth int) string {
	displayed := displayedTheme(name)
	if clistyles.NoColorEnabled() {
		return displayed
	}
	palette := corestyles.DefaultPalette()
	if name != "" {
		resolved, err := coreconfig.LoadTheme(name)
		if err != nil {
			return displayed + "  unavailable"
		}
		palette = resolved.Palette
	}
	paletteWidth := paletteSwatchCount * paletteSwatchWidth
	if terminalWidth > 0 {
		paletteWidth = min(paletteWidth, max(0, terminalWidth-lipgloss.Width(displayed)-2))
		paletteWidth -= paletteWidth % paletteSwatchWidth
	}
	preview := renderPalettePreview(palette, paletteWidth)
	if preview == "" {
		return displayed
	}
	return displayed + "  " + preview
}

type themeCurrentResult struct {
	Theme   string `json:"theme"`
	Source  string `json:"source" contract:"enum=default|global|local"`
	BuiltIn bool   `json:"built_in"`
}

type themeListItem struct {
	Theme   string `json:"theme"`
	Active  bool   `json:"active"`
	BuiltIn bool   `json:"built_in"`
}

type themeSwitchResult struct {
	Scope                  string `json:"scope" contract:"enum=global|local"`
	RequestedTheme         string `json:"requested_theme"`
	PreviousTheme          string `json:"previous_theme"`
	PreviousEffectiveTheme string `json:"previous_effective_theme"`
	EffectiveTheme         string `json:"effective_theme"`
	Changed                bool   `json:"changed"`
	Overridden             bool   `json:"overridden"`
}

type themeResetResult struct {
	Scope          string `json:"scope" contract:"enum=global|local"`
	PreviousTheme  string `json:"previous_theme"`
	EffectiveTheme string `json:"effective_theme"`
	Status         string `json:"status" contract:"enum=reset|unchanged"`
	Changed        bool   `json:"changed"`
}

type themePathResult struct {
	Theme  string `json:"theme"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

type themeRenameResult struct {
	OldTheme       string `json:"old_theme"`
	NewTheme       string `json:"new_theme"`
	EffectiveTheme string `json:"effective_theme"`
	Changed        bool   `json:"changed"`
}

type themeDeleteResult struct {
	Theme  string `json:"theme"`
	Status string `json:"status" contract:"enum=deleted"`
	Path   string `json:"path"`
}

type themeImportResult struct {
	Theme  string `json:"theme"`
	Status string `json:"status" contract:"enum=imported"`
	Path   string `json:"path"`
	Source string `json:"source"`
}

type themeBatchImportItem struct {
	Theme  string `json:"theme"`
	Status string `json:"status" contract:"enum=imported|skipped"`
	Path   string `json:"path"`
	Reason string `json:"reason,omitempty" contract:"enum=already_exists"`
}

type themeBatchImportResult struct {
	Source        string                 `json:"source"`
	Count         int                    `json:"count"`
	ImportedCount int                    `json:"imported_count"`
	SkippedCount  int                    `json:"skipped_count"`
	Items         []themeBatchImportItem `json:"items"`
}
