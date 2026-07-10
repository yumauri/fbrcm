package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/spf13/cobra"

	importpkg "github.com/yumauri/fbrcm/cli/commands/project/import"
	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/cli/shared/rc"
	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

// New constructs the project command.
func New(svc *core.Core) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Manage project remote config",
	}
	projectCmd.AddCommand(newExportCommand(svc), newImportCommand(svc))
	return projectCmd
}

func newExportCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
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
				body := rc.TrimTrailingLineBreaks(rc.NormalizeExportJSON(raw))
				_, err = cmd.OutOrStdout().Write(body)
				return err
			}

			if err := writeRemoteConfigFile(toPath, raw); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "📤 exported: %s\n", toPath)
			return nil
		},
	}
	cmd.Flags().String("to", "", "Write Remote Config JSON to file path")
	return cmd
}

func newImportCommand(svc *core.Core) *cobra.Command {
	cmd := &cobra.Command{
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
			return importpkg.Run(cmd, svc, project)
		},
	}
	cmd.Flags().String("from", "", "Read Remote Config JSON from file path")
	cmd.Flags().StringArray("group", nil, "Import only specified parameter group; may be repeated")
	shared.AddParameterFilterFlags(cmd)
	cmd.Flags().String("expr", "", "Filter imported config by expr-lang expression")
	shared.AddDryRunFlag(cmd)
	cmd.Flags().Bool("remove-all-conditions", false, "Remove all conditions and conditional values from imported config")
	cmd.Flags().Bool("remove-project-specific-conditions", false, "Remove project specific conditions and their usages from imported config")
	cmd.Flags().Bool("merge", false, "Merge imported config into current project config")
	cmd.Flags().Bool("override", false, "Replace current project config with imported config")
	cmd.Flags().String("merge-resolve", "", "Conflict resolution for merge: current or import")
	cmd.MarkFlagsMutuallyExclusive("remove-all-conditions", "remove-project-specific-conditions")
	cmd.MarkFlagsMutuallyExclusive("merge", "override")
	return cmd
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

func writeRemoteConfigFile(path string, raw []byte) error {
	raw = rc.TrimTrailingLineBreaks(rc.NormalizeExportJSON(raw))
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
