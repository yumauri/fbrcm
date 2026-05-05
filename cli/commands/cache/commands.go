package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/cobra"

	clistyles "fbrcm/cli/styles"
	"fbrcm/core/config"
	"fbrcm/core/firebase"
	corelog "fbrcm/core/log"
)

type cacheEntry struct {
	ProjectID string    `json:"project_id"`
	Project   string    `json:"project"`
	Version   string    `json:"version"`
	Size      int64     `json:"size"`
	CachedAt  time.Time `json:"cached_at"`
	Path      string    `json:"path"`
}

func New() *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage cached parameters files",
	}

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print parameters cache directory path",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			path := config.GetParametersCacheDirPath()
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
		Short: "Delete cached parameters files",
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !yes {
				confirm := confirmation.New(
					fmt.Sprintf("Delete cached parameters files in %s?", config.GetParametersCacheDirPath()),
					confirmation.Yes,
				)
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			if err := config.PurgeParametersCache(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "purged: %s\n", config.GetParametersCacheDirPath())
			return nil
		},
	}
	purgeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation dialog")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List cached parameters files",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}

			entries, err := loadCacheEntries()
			if err != nil {
				return err
			}

			if jsonOut {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(entries); err != nil {
					return err
				}
				logCacheTotal(entries)
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), renderCacheTable(entries))
			logCacheTotal(entries)
			return nil
		},
	}
	listCmd.Flags().Bool("json", false, "Print cache entries as JSON")

	cacheCmd.AddCommand(pathCmd, purgeCmd, listCmd)
	return cacheCmd
}

func loadCacheEntries() ([]cacheEntry, error) {
	dir := config.GetParametersCacheDirPath()
	projectNames := loadProjectNames()

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []cacheEntry{}, nil
		}
		return nil, fmt.Errorf("read cache dir: %w", err)
	}

	entries := make([]cacheEntry, 0, len(files))
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		projectID := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		path := filepath.Join(dir, file.Name())
		info, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("stat cache file %s: %w", path, err)
		}

		cache, err := config.LoadParametersCache(projectID)
		if err != nil {
			return nil, err
		}

		version := ""
		if remoteConfig, err := firebase.ParseRemoteConfig(cache.RemoteConfig); err == nil {
			version = remoteConfig.Version.VersionNumber
		}

		project := projectNames[projectID]

		entries = append(entries, cacheEntry{
			ProjectID: projectID,
			Project:   project,
			Version:   version,
			CachedAt:  cache.CachedAt,
			Size:      info.Size(),
			Path:      path,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		left := strings.ToLower(entries[i].ProjectID)
		right := strings.ToLower(entries[j].ProjectID)
		if left == right {
			return entries[i].ProjectID < entries[j].ProjectID
		}
		return left < right
	})

	return entries, nil
}

func loadProjectNames() map[string]string {
	projects, err := config.LoadProjects()
	if err != nil {
		return nil
	}

	names := make(map[string]string, len(projects))
	for _, project := range projects {
		names[project.ProjectID] = project.Name
	}
	return names
}

func renderCacheTable(entries []cacheEntry) string {
	noColor := clistyles.NoColorEnabled()
	rows := make([][]string, 0, len(entries))
	projectIDWidth := lipgloss.Width("Project ID")
	projectWidth := lipgloss.Width("Project")
	versionWidth := lipgloss.Width("Version")
	cachedAtWidth := lipgloss.Width("Cached At")
	sizeWidth := lipgloss.Width("Size")

	for _, entry := range entries {
		cachedAt := ""
		if !entry.CachedAt.IsZero() {
			cachedAt = entry.CachedAt.Local().Format("2006-01-02 15:04:05")
		}
		size := humanSize(entry.Size)

		rows = append(rows, []string{
			entry.ProjectID,
			entry.Project,
			entry.Version,
			size,
			cachedAt,
		})
		projectIDWidth = max(projectIDWidth, lipgloss.Width(entry.ProjectID))
		projectWidth = max(projectWidth, lipgloss.Width(entry.Project))
		versionWidth = max(versionWidth, lipgloss.Width(entry.Version))
		cachedAtWidth = max(cachedAtWidth, lipgloss.Width(cachedAt))
		sizeWidth = max(sizeWidth, lipgloss.Width(size))
	}

	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if col == 2 {
			style = style.AlignHorizontal(lipgloss.Right)
		}
		if col == 3 {
			style = style.AlignHorizontal(lipgloss.Right)
		}
		if noColor {
			return style
		}
		if row == table.HeaderRow {
			return style.Bold(true).Foreground(clistyles.PaletteSlateBright)
		}
		if row >= 0 && row%2 == 1 {
			style = style.Background(clistyles.ColorRowStripe)
		}
		if col == 1 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateDim)
	}

	tbl := table.New().
		Headers("Project ID", "Project", "Version", "Size", "Cached At").
		Rows(rows...).
		Width(projectIDWidth + projectWidth + versionWidth + cachedAtWidth + sizeWidth + 16).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func humanSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B ", size)
	}

	const kb = 1024.0
	const mb = 1024.0 * 1024.0

	if size < 1024*1024 {
		value := float64(size) / kb
		if value < 10 {
			return fmt.Sprintf("%.1f KB", value)
		}
		return fmt.Sprintf("%.0f KB", value)
	}
	value := float64(size) / mb
	if value < 10 {
		return fmt.Sprintf("%.1f MB", value)
	}
	return fmt.Sprintf("%.0f MB", value)
}

func logCacheTotal(entries []cacheEntry) {
	size := totalCacheSize(entries)
	corelog.For("cache").Info("total", "projects", len(entries), "size", size, "hsize", strings.TrimSpace(humanSize(size)))
}

func totalCacheSize(entries []cacheEntry) int64 {
	var total int64
	for _, entry := range entries {
		total += entry.Size
	}
	return total
}
