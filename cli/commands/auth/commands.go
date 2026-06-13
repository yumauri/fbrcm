package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
)

// New constructs auth command.
func New(svc *core.Core) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage auth identities",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List auth identities",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			entries, defaultAuthID, err := svc.ListAuth()
			if err != nil {
				return err
			}
			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(map[string]any{"default_auth_id": defaultAuthID, "auth": entries})
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderAuthTable(entries, defaultAuthID))
			return nil
		},
	}
	listCmd.Flags().Bool("json", false, "Print auth identities as JSON")

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add auth identity",
	}
	addOAuthCmd := &cobra.Command{
		Use:   "oauth <auth-id>",
		Short: "Add OAuth auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromPath, err := cmd.Flags().GetString("from")
			if err != nil {
				return err
			}
			label, err := cmd.Flags().GetString("label")
			if err != nil {
				return err
			}
			data, err := readOAuthClientSecret(cmd, fromPath)
			if err != nil {
				return err
			}
			entry, err := svc.AddOAuthAuth(args[0], label, data)
			if err != nil {
				return err
			}
			_, paths, err := svc.AuthPaths(entry.ID)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "secret: %s\n", paths.ClientSecretPath)
			return nil
		},
	}
	addOAuthCmd.Flags().String("from", "", "Import OAuth client secret from file path; if omitted, read stdin or open file picker")
	addOAuthCmd.Flags().String("label", "", "Auth identity label")

	addServiceAccountCmd := &cobra.Command{
		Use:   "service-account <auth-id>",
		Short: "Add service account auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromPath, err := cmd.Flags().GetString("from")
			if err != nil {
				return err
			}
			label, err := cmd.Flags().GetString("label")
			if err != nil {
				return err
			}
			data, err := readJSONFileInput(cmd, fromPath, "service account key")
			if err != nil {
				return err
			}
			entry, err := svc.AddServiceAccountAuth(args[0], label, data)
			if err != nil {
				return err
			}
			_, paths, err := svc.AuthPaths(entry.ID)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "service account: %s\n", paths.ServiceAccountPath)
			return nil
		},
	}
	addServiceAccountCmd.Flags().String("from", "", "Import service account key from file path; if omitted, read stdin or open file picker")
	addServiceAccountCmd.Flags().String("label", "", "Auth identity label")

	addGCloudCmd := &cobra.Command{
		Use:   "gcloud <auth-id>",
		Short: "Add gcloud ADC auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label, err := cmd.Flags().GetString("label")
			if err != nil {
				return err
			}
			entry, err := svc.AddGCloudAuth(args[0], label)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔐 added auth: %s\n", entry.ID)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "adc: application default credentials")
			return nil
		},
	}
	addGCloudCmd.Flags().String("label", "", "Auth identity label")
	addCmd.AddCommand(addOAuthCmd, addServiceAccountCmd, addGCloudCmd)

	loginCmd := &cobra.Command{
		Use:   "login <auth-id>",
		Short: "Authenticate auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noOpen, err := cmd.Flags().GetBool("noopen")
			if err != nil {
				return err
			}
			if err := svc.EnsureAuthLogin(context.Background(), args[0], noOpen); err != nil {
				return err
			}
			auth, paths, err := svc.AuthPaths(args[0])
			if err != nil {
				return err
			}
			switch auth.Type {
			case config.AuthTypeOAuth:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔑 authenticated: %s\n", paths.TokenPath)
			case config.AuthTypeServiceAccount:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔑 authenticated: %s\n", paths.ServiceAccountPath)
			case config.AuthTypeGCloud:
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "🔑 authenticated: application default credentials")
			default:
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔑 authenticated: %s\n", auth.ID)
			}
			return nil
		},
	}
	loginCmd.Flags().Bool("noopen", false, "Do not open browser automatically")

	pathCmd := &cobra.Command{
		Use:   "path <auth-id>",
		Short: "Print auth file paths",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			auth, paths, err := svc.AuthPaths(args[0])
			if err != nil {
				return err
			}
			payload := authPathPayload(auth, paths)
			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(payload)
			}
			for _, path := range authPathLines(auth, paths) {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}
	pathCmd.Flags().Bool("json", false, "Print paths as JSON")

	purgeCmd := &cobra.Command{
		Use:   "purge <auth-id>",
		Short: "Delete auth identity files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Delete auth identity %s and its files?", args[0]),
					confirmation.Yes,
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
			auth, paths, err := svc.PurgeAuth(args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged auth: %s\n", auth.ID)
			for _, path := range authPathLines(auth, paths) {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged: %s\n", path)
			}
			return nil
		},
	}
	purgeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation dialog")

	bindCmd := &cobra.Command{
		Use:   "bind <project-query>",
		Short: "Bind projects to auth identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			authID, err := cmd.Flags().GetString("auth")
			if err != nil {
				return err
			}
			if authID == "" {
				return fmt.Errorf("--auth is required")
			}
			projects, err := svc.BindProjectsAuth([]string{args[0]}, authID)
			if err != nil {
				return err
			}
			for _, project := range projects {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🔗 bound: %s -> %s\n", project.ProjectID, authID)
			}
			return nil
		},
	}
	bindCmd.Flags().String("auth", "", "Auth id to bind")

	authCmd.AddCommand(listCmd, addCmd, loginCmd, pathCmd, purgeCmd, bindCmd)
	return authCmd
}

func renderAuthTable(entries []config.AuthEntry, defaultAuthID string) string {
	rows := make([][]string, 0, len(entries))
	idWidth := lipgloss.Width("Auth")
	typeWidth := lipgloss.Width("Type")
	labelWidth := lipgloss.Width("Label")
	defaultWidth := lipgloss.Width("Default")
	for _, entry := range entries {
		marker := ""
		if entry.ID == defaultAuthID {
			marker = "✓"
		}
		rows = append(rows, []string{entry.ID, entry.Type, entry.Label, marker})
		idWidth = max(idWidth, lipgloss.Width(entry.ID))
		typeWidth = max(typeWidth, lipgloss.Width(entry.Type))
		labelWidth = max(labelWidth, lipgloss.Width(entry.Label))
		defaultWidth = max(defaultWidth, lipgloss.Width(marker))
	}
	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if clistyles.NoColorEnabled() {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if col == 0 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateDim)
	}
	tbl := table.New().
		Headers("Auth", "Type", "Label", "Default").
		Rows(rows...).
		Width(idWidth + typeWidth + labelWidth + defaultWidth + 13).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func authPathPayload(auth config.AuthEntry, paths core.AuthPaths) map[string]string {
	payload := map[string]string{
		"id":                  auth.ID,
		"type":                auth.Type,
		"auth_config_path":    paths.AuthConfigPath,
		"profile_config_path": paths.ProfileConfigPath,
	}
	if paths.ClientSecretPath != "" {
		payload["client_secret_path"] = paths.ClientSecretPath
	}
	if paths.TokenPath != "" {
		payload["token_path"] = paths.TokenPath
	}
	if paths.ServiceAccountPath != "" {
		payload["service_account_path"] = paths.ServiceAccountPath
	}
	return payload
}

func authPathLines(auth config.AuthEntry, paths core.AuthPaths) []string {
	switch auth.Type {
	case config.AuthTypeOAuth:
		return nonEmptyStrings(paths.ClientSecretPath, paths.TokenPath)
	case config.AuthTypeServiceAccount:
		return nonEmptyStrings(paths.ServiceAccountPath)
	default:
		return nil
	}
}

func nonEmptyStrings(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func readOAuthClientSecret(cmd *cobra.Command, fromPath string) ([]byte, error) {
	return readJSONFileInput(cmd, fromPath, "client secret")
}

func readJSONFileInput(cmd *cobra.Command, fromPath, label string) ([]byte, error) {
	switch {
	case fromPath != "":
		data, err := os.ReadFile(fromPath)
		if err != nil {
			return nil, fmt.Errorf("read source file: %w", err)
		}
		return data, nil
	case stdinAvailable():
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return data, nil
	default:
		selectedPath, err := pickSecretFile()
		if err != nil {
			return nil, err
		}
		if selectedPath == "" {
			return nil, fmt.Errorf("no %s selected", label)
		}
		data, err := os.ReadFile(selectedPath)
		if err != nil {
			return nil, fmt.Errorf("read selected file: %w", err)
		}
		return data, nil
	}
}

func stdinAvailable() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

type pickerModel struct {
	picker   filepicker.Model
	selected string
	cancel   bool
}

func pickSecretFile() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	picker := filepicker.New()
	picker.CurrentDirectory = currentDir
	picker.AllowedTypes = []string{".json"}
	picker.FileAllowed = true
	picker.DirAllowed = false
	picker.ShowHidden = true
	picker.AutoHeight = true

	finalModel, err := tea.NewProgram(pickerModel{picker: picker}).Run()
	if err != nil {
		return "", err
	}

	model, ok := finalModel.(pickerModel)
	if !ok || model.cancel {
		return "", nil
	}
	return model.selected, nil
}

func (m pickerModel) Init() tea.Cmd {
	return m.picker.Init()
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancel = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)

	if didSelect, path := m.picker.DidSelectFile(msg); didSelect {
		m.selected = path
		return m, tea.Quit
	}

	return m, cmd
}

func (m pickerModel) View() tea.View {
	return tea.NewView(m.picker.View() + "\n\nenter/l to open or select, h/backspace/left/esc to go up, q to cancel")
}
