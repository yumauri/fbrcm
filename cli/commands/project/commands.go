package project

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	clistyles "fbrcm/cli/styles"
	"fbrcm/core"
	"fbrcm/core/firebase"
)

func New(svc *core.Core) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Manage project remote config",
	}

	exportCmd := &cobra.Command{
		Use:   "export <project>",
		Short: "Export project Remote Config JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := resolveProjectArg(context.Background(), cmd, svc, args[0])
			if err != nil {
				return err
			}

			raw, _, err := svc.ExportRemoteConfig(context.Background(), project.ProjectID)
			if err != nil {
				return err
			}

			toPath, err := cmd.Flags().GetString("to")
			if err != nil {
				return err
			}
			if toPath == "" {
				body := trimTrailingLineBreaks(normalizeExportJSON(raw))
				_, err = cmd.OutOrStdout().Write(body)
				return err
			}

			if err := writeRemoteConfigFile(toPath, raw); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "exported: %s\n", toPath)
			return nil
		},
	}
	exportCmd.Flags().String("to", "", "Write Remote Config JSON to file path")

	importCmd := &cobra.Command{
		Use:   "import <project>",
		Short: "Import project Remote Config JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			if dryRun {
				ctx = firebase.WithDryRun(ctx)
			}

			project, err := resolveProjectArg(ctx, cmd, svc, args[0])
			if err != nil {
				return err
			}
			return runImportCommand(cmd, svc, project)
		},
	}
	importCmd.Flags().String("from", "", "Read Remote Config JSON from file path")
	importCmd.Flags().StringArray("group", nil, "Import only specified parameter group; may be repeated")
	importCmd.Flags().StringP("filter", "f", "", "Filter parameters by mode-prefixed query (^, /, ~, =)")
	importCmd.Flags().String("expr", "", "Filter imported config by expr-lang expression")
	importCmd.Flags().Bool("dry-run", false, "Log Firebase write requests without sending them")
	importCmd.Flags().Bool("remove-all-conditions", false, "Remove all conditions and conditional values from imported config")
	importCmd.Flags().Bool("remove-project-specific-conditions", false, "Remove project specific conditions and their usages from imported config")
	importCmd.Flags().Bool("merge", false, "Merge imported config into current project config")
	importCmd.Flags().Bool("override", false, "Replace current project config with imported config")
	importCmd.Flags().String("merge-resolve", "", "Conflict resolution for merge: current or import")
	importCmd.MarkFlagsMutuallyExclusive("remove-all-conditions", "remove-project-specific-conditions")
	importCmd.MarkFlagsMutuallyExclusive("merge", "override")

	projectCmd.AddCommand(exportCmd, importCmd)
	return projectCmd
}

func resolveProjectArg(ctx context.Context, cmd *cobra.Command, svc *core.Core, query string) (core.Project, error) {
	projects, _, err := svc.ListProjects(ctx)
	if err != nil {
		return core.Project{}, err
	}

	for _, project := range projects {
		if strings.EqualFold(project.ProjectID, query) {
			return project, nil
		}
	}

	matches := make([]core.Project, 0, 1)
	for _, project := range projects {
		if strings.EqualFold(project.Name, query) {
			matches = append(matches, project)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		if len(projects) > 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderAmbiguousProjectsTable(projects))
		}
		return core.Project{}, fmt.Errorf("no project matches %q", query)
	default:
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderAmbiguousProjectsTable(matches))
		return core.Project{}, fmt.Errorf("several projects match %q", query)
	}
}

func renderAmbiguousProjectsTable(projects []core.Project) string {
	rows := make([][]string, 0, len(projects))
	projectWidth := lipgloss.Width("Project")
	idWidth := lipgloss.Width("Project ID")
	for _, project := range projects {
		rows = append(rows, []string{project.Name, project.ProjectID})
		projectWidth = max(projectWidth, lipgloss.Width(project.Name))
		idWidth = max(idWidth, lipgloss.Width(project.ProjectID))
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
		Headers("Project", "Project ID").
		Rows(rows...).
		Width(projectWidth + idWidth + 7).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func readImportRemoteConfig(cmd *cobra.Command) ([]byte, error) {
	fromPath, err := cmd.Flags().GetString("from")
	if err != nil {
		return nil, err
	}

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
		selectedPath, err := pickJSONFile()
		if err != nil {
			return nil, err
		}
		if selectedPath == "" {
			return nil, nil
		}
		data, err := os.ReadFile(selectedPath)
		if err != nil {
			return nil, fmt.Errorf("read selected file: %w", err)
		}
		return data, nil
	}
}

func writeRemoteConfigFile(path string, raw []byte) error {
	raw = trimTrailingLineBreaks(normalizeExportJSON(raw))
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create destination dir: %w", err)
		}
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write destination file: %w", err)
	}
	return nil
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

func pickJSONFile() (string, error) {
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
