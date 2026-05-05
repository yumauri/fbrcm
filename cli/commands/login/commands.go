package login

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	clistyles "fbrcm/cli/styles"
	"fbrcm/core"
	"fbrcm/core/browser"
	"fbrcm/core/config"
)

const googleAuthClientsURL = "https://console.cloud.google.com/auth/clients"

func New(svc *core.Core) *cobra.Command {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Manage Firebase login files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if svc == nil {
				return fmt.Errorf("login service is not available")
			}
			noOpen, err := cmd.Flags().GetBool("noopen")
			if err != nil {
				return err
			}
			if err := svc.EnsureLogin(context.Background(), noOpen); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "authenticated: %s\n", config.GetTokenFilePath())
			return nil
		},
	}
	loginCmd.Flags().Bool("noopen", false, "Do not open browser automatically")

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print client secret file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := config.GetSecretFilePath()
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

	purgeCmd := &cobra.Command{
		Use:   "purge",
		Short: "Delete login files",
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}

			tokenExists := fileExists(config.GetTokenFilePath())
			deleteToken := yes
			deleteSecret := yes
			if !yes {
				if tokenExists {
					tokenConfirm := newLoginConfirmation(
						fmt.Sprintf("Delete token file %s?", config.GetTokenFilePath()),
						confirmation.Yes,
						confirmationNote{
							text:  "After deleting token file you will have to reauthenticate the app.",
							color: clistyles.ColorNote,
						},
					)
					deleteToken, err = tokenConfirm.RunPrompt()
					if err != nil {
						return err
					}
				}

				secretNotes := []confirmationNote{
					{
						text:  "After deleting client secret file you will have to add a different client secret.",
						color: clistyles.ColorNote,
					},
				}
				if tokenExists && !deleteToken {
					secretNotes = append(secretNotes, confirmationNote{
						text:  "If you delete client secret file, token file will also be deleted because it cannot be used without the secret.",
						color: clistyles.ColorNote,
					})
				}
				secretNotes = append(secretNotes, confirmationNote{
					text:  "There is no way to download it from Google Cloud Console for the same app for the second time.",
					color: clistyles.PaletteError,
				})

				secretConfirm := newLoginConfirmation(
					fmt.Sprintf("Delete client secret file %s?", config.GetSecretFilePath()),
					confirmation.Yes,
					secretNotes...,
				)
				deleteSecret, err = secretConfirm.RunPrompt()
				if err != nil {
					return err
				}
			}

			if deleteSecret {
				deleteToken = true
			}

			if !deleteToken && !deleteSecret {
				return nil
			}

			if deleteToken {
				if err := removeLoginFile(config.GetTokenFilePath()); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "purged: %s\n", config.GetTokenFilePath())
			}
			if deleteSecret {
				if err := removeLoginFile(config.GetSecretFilePath()); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "purged: %s\n", config.GetSecretFilePath())
			}
			return nil
		},
	}
	purgeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation dialog")

	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import client secret file",
		RunE: func(cmd *cobra.Command, args []string) error {
			fromPath, err := cmd.Flags().GetString("from")
			if err != nil {
				return err
			}

			var data []byte
			switch {
			case fromPath != "":
				data, err = os.ReadFile(fromPath)
				if err != nil {
					return fmt.Errorf("read source file: %w", err)
				}
			case stdinAvailable():
				data, err = io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
			default:
				selectedPath, err := pickSecretFile()
				if err != nil {
					return err
				}
				if selectedPath == "" {
					return nil
				}
				data, err = os.ReadFile(selectedPath)
				if err != nil {
					return fmt.Errorf("read selected file: %w", err)
				}
			}

			if _, err := writeSecretFile(data); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "imported: %s\n", config.GetSecretFilePath())
			return nil
		},
	}
	importCmd.Flags().String("from", "", "Import client secret from file path")

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Guide Google OAuth client setup and import",
		RunE: func(cmd *cobra.Command, args []string) error {
			noOpen, err := cmd.Flags().GetBool("noopen")
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "To create Google OAuth client for fbrcm:")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "1. Select or create Google Cloud project.")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "2. Click Create Client.")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "3. Choose Application type: Desktop app.")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "4. Click Create and download JSON file.")
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Open this page: %s\n", googleAuthClientsURL)

			if !noOpen {
				if err := browser.OpenURL(googleAuthClientsURL); err != nil {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Could not open browser automatically: %v\n", err)
				}
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "After downloading JSON file, press enter to choose it.")
			if err := waitForEnter(cmd.InOrStdin()); err != nil {
				return err
			}

			selectedPath, err := pickSecretFile()
			if err != nil {
				return err
			}
			if selectedPath == "" {
				return nil
			}

			data, err := os.ReadFile(selectedPath)
			if err != nil {
				return fmt.Errorf("read selected file: %w", err)
			}
			if _, err := writeSecretFile(data); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "imported: %s\n", config.GetSecretFilePath())
			return nil
		},
	}
	setupCmd.Flags().Bool("noopen", false, "Do not open browser automatically")

	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Print authenticated user information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if svc == nil {
				return fmt.Errorf("login service is not available")
			}

			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			info, err := svc.WhoAmI(context.Background())
			if err != nil {
				return err
			}

			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "secret_path: %s\n", info.SecretPath)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "token_path: %s\n", info.TokenPath)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "token_expiry: %s\n", info.TokenExpiry)
			return nil
		},
	}
	whoamiCmd.Flags().Bool("json", false, "Print user information as JSON")

	loginCmd.AddCommand(pathCmd, purgeCmd, importCmd, setupCmd, whoamiCmd)
	return loginCmd
}

func removeLoginFile(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}

func writeSecretFile(data []byte) (bool, error) {
	path := config.GetSecretFilePath()
	previous, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("read existing client secret: %w", err)
	}
	secretChanged := err == nil && !bytes.Equal(previous, data)

	if err := config.EnsurePrivateDir(filepath.Dir(path)); err != nil {
		return false, fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, data, config.PrivateFileMode); err != nil {
		return false, fmt.Errorf("write client secret: %w", err)
	}
	if err := config.EnsurePrivateFile(path); err != nil {
		return false, fmt.Errorf("chmod client secret: %w", err)
	}

	if secretChanged {
		if err := removeLoginFile(config.GetTokenFilePath()); err != nil {
			return false, fmt.Errorf("remove token for previous client secret: %w", err)
		}
	}

	return secretChanged, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func stdinAvailable() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

func waitForEnter(in io.Reader) error {
	_, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
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

type confirmationNote struct {
	text  string
	color color.Color
}

func newLoginConfirmation(prompt string, defaultValue confirmation.Value, notes ...confirmationNote) *confirmation.Confirmation {
	confirm := confirmation.New(prompt, defaultValue)
	confirm.Template = `
{{- Bold .Prompt -}}
{{- "\n" -}}
{{- Hint -}}
{{- "\n" -}}
{{ if .YesSelected -}}
	{{- print (Bold " ▸Yes ") " No" -}}
{{- else if .NoSelected -}}
	{{- print "  Yes " (Bold "▸No") -}}
{{- else -}}
	{{- "  Yes  No" -}}
{{- end -}}
`
	confirm.ExtendedTemplateFuncs["Hint"] = func() string {
		return renderConfirmationNotes(notes)
	}
	return confirm
}

func renderConfirmationNotes(notes []confirmationNote) string {
	lines := make([]string, 0, len(notes))
	for _, note := range notes {
		if note.text == "" {
			continue
		}
		if clistyles.NoColorEnabled() {
			lines = append(lines, note.text)
			continue
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(note.color).Render(note.text))
	}
	return strings.Join(lines, "\n")
}
