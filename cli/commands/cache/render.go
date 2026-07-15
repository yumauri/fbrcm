package cache

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	corelog "github.com/yumauri/fbrcm/core/log"
)

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
		if entry.CachedAt != nil && !entry.CachedAt.IsZero() {
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
