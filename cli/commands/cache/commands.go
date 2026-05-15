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

	"fbrcm/cli/shared"
	clistyles "fbrcm/cli/styles"
	"fbrcm/core/config"
	"fbrcm/core/firebase"
	corelog "fbrcm/core/log"
)

// cacheEntry holds cache entry state used by the cache package.
type cacheEntry struct {
	// ProjectID stores project id for cacheEntry.
	ProjectID string `json:"project_id"`
	// Project stores project for cacheEntry.
	Project string `json:"project"`
	// Version stores version for cacheEntry.
	Version string `json:"version"`
	// Size stores size for cacheEntry.
	Size int64 `json:"size"`
	// CachedAt stores cached at for cacheEntry.
	CachedAt *time.Time `json:"cached_at"`
	// Draft stores draft for cacheEntry.
	Draft bool `json:"draft"`
	// Path stores path for cacheEntry.
	Path string `json:"path"`
}

// New constructs new and returns the resulting value or error.
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

			path := config.GetCacheDirPath()
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

			deleteCaches := true
			if !yes {
				confirm := shared.NewConfirmation(
					fmt.Sprintf("Delete cached parameters files in %s?", config.GetParametersCacheDirPath()),
					confirmation.Yes,
					shared.ConfirmationOptions{Destructive: true},
				)
				ok, err := confirm.RunPrompt()
				if err != nil {
					return err
				}
				deleteCaches = ok
			}
			if deleteCaches {
				if err := config.PurgeParametersCache(); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged caches: %s\n", config.GetParametersCacheDirPath())
			}

			draftIDs, err := config.ListDraftProjectIDs()
			if err != nil {
				return err
			}
			if len(draftIDs) > 0 {
				deleteDrafts := true
				if !yes {
					confirm := shared.NewConfirmation(
						fmt.Sprintf("Delete draft files in %s?", config.GetDraftsDirPath()),
						confirmation.No,
						shared.ConfirmationOptions{Destructive: true},
					)
					ok, err := confirm.RunPrompt()
					if err != nil {
						return err
					}
					deleteDrafts = ok
				}
				if deleteDrafts {
					if err := config.PurgeDrafts(); err != nil {
						return err
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "🧹 purged drafts: %s\n", config.GetDraftsDirPath())
				}
			}

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

// loadCacheEntries loads load cache entries and returns the resulting value or error.
func loadCacheEntries() ([]cacheEntry, error) {
	projectNames := loadProjectNames()
	entries, err := loadParametersCacheEntries(projectNames)
	if err != nil {
		return nil, err
	}
	draftEntries, err := loadDraftEntries(projectNames)
	if err != nil {
		return nil, err
	}
	entries = append(entries, draftEntries...)

	sort.Slice(entries, func(i, j int) bool {
		left := strings.ToLower(entries[i].ProjectID)
		right := strings.ToLower(entries[j].ProjectID)
		if left == right {
			if entries[i].Draft != entries[j].Draft {
				return !entries[i].Draft
			}
			return entries[i].ProjectID < entries[j].ProjectID
		}
		return left < right
	})

	return entries, nil
}

// loadParametersCacheEntries loads load parameters cache entries and returns the resulting value or error.
func loadParametersCacheEntries(projectNames map[string]string) ([]cacheEntry, error) {
	dir := config.GetParametersCacheDirPath()
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
		cachedAt := cache.CachedAt
		entries = append(entries, cacheEntry{
			ProjectID: projectID,
			Project:   projectNames[projectID],
			Version:   version,
			CachedAt:  &cachedAt,
			Size:      info.Size(),
			Path:      path,
		})
	}
	return entries, nil
}

// loadDraftEntries loads load draft entries and returns the resulting value or error.
func loadDraftEntries(projectNames map[string]string) ([]cacheEntry, error) {
	dir := config.GetDraftsDirPath()
	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []cacheEntry{}, nil
		}
		return nil, fmt.Errorf("read drafts dir: %w", err)
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
			return nil, fmt.Errorf("stat draft file %s: %w", path, err)
		}

		raw, err := config.LoadDraft(projectID)
		if err != nil {
			return nil, err
		}

		version := ""
		if remoteConfig, err := firebase.ParseRemoteConfig(raw); err == nil {
			version = remoteConfig.Version.VersionNumber
		}

		entries = append(entries, cacheEntry{
			ProjectID: projectID,
			Project:   projectNames[projectID],
			Version:   version,
			Size:      info.Size(),
			Draft:     true,
			Path:      path,
		})
	}
	return entries, nil
}

// loadProjectNames loads load project names and returns the resulting value or error.
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

// renderCacheTable renders render cache table and returns the resulting value or error.
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
		if entry.Draft {
			cachedAt = "draft"
		} else if entry.CachedAt != nil && !entry.CachedAt.IsZero() {
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
		if row >= 0 && col == 4 && entries[row].Draft {
			return style.Foreground(clistyles.PaletteError)
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

// humanSize handles human size and returns the resulting value or error.
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

// logCacheTotal handles log cache total and returns the resulting value or error.
func logCacheTotal(entries []cacheEntry) {
	size := totalCacheSize(entries)
	corelog.For("cache").Info("total", "projects", len(entries), "size", size, "hsize", strings.TrimSpace(humanSize(size)))
}

// totalCacheSize handles total cache size and returns the resulting value or error.
func totalCacheSize(entries []cacheEntry) int64 {
	var total int64
	for _, entry := range entries {
		total += entry.Size
	}
	return total
}
