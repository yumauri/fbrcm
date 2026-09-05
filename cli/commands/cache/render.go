package cache

import (
	"fmt"
	"strings"

	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/ops/shared"
)

func renderCacheTable(entries []cacheEntry) string {
	return renderCacheTableAtWidth(entries, shared.TerminalWidth())
}

func renderCacheTableAtWidth(entries []cacheEntry, terminalWidth int) string {
	hasResource := false
	for _, entry := range entries {
		if entry.Resource != "" {
			hasResource = true
			break
		}
	}
	headers := []string{"Kind", "Project ID", "Project"}
	if hasResource {
		headers = append(headers, "Resource")
	}
	headers = append(headers, "Version", "Size", "Cached At")
	rows := make([][]string, 0, len(entries))
	widths := shared.HeaderWidths(headers)
	for _, entry := range entries {
		cachedAt := ""
		if entry.CachedAt != nil && !entry.CachedAt.IsZero() {
			cachedAt = entry.CachedAt.Local().Format("2006-01-02 15:04:05")
		}
		row := []string{entry.Kind, entry.ProjectID, entry.Project}
		if hasResource {
			resource := entry.Resource
			if resource == "" {
				resource = "—"
			}
			row = append(row, resource)
		}
		version := entry.Version
		if version == "" {
			version = "—"
		}
		row = append(row, version, humanSize(entry.Size), cachedAt)
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
	}
	flexible := []int{2, 1}
	if hasResource {
		flexible = append(flexible, 3)
	}
	flexible = append(flexible, len(headers)-1, 0)
	shared.FitTableColumns(widths, terminalWidth, flexible...)
	shared.TruncateTableHeaders(headers, widths)
	shared.TruncateTableColumns(rows, widths, flexible...)
	versionColumn := 3
	if hasResource {
		versionColumn++
	}
	alignments := map[int]bool{versionColumn: true, versionColumn + 1: true}
	return shared.StyledTable(headers, rows, widths, alignments, nil)
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
	corelog.For("cache").Info("total", "entries", len(entries), "size", size, "hsize", strings.TrimSpace(humanSize(size)))
}

func totalCacheSize(entries []cacheEntry) int64 {
	var total int64
	for _, entry := range entries {
		total += entry.Size
	}
	return total
}
