package versions

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
	"github.com/yumauri/fbrcm/core"
)

func renderVersionsTable(entries []core.RemoteConfigVersionEntry, cachedOnly bool) string {
	noColor := clistyles.NoColorEnabled()
	headers := []string{"Version", "State", "Published", "Updated By", "Origin", "Type", "Cached", "Description"}
	if cachedOnly {
		headers = []string{"Version", "State", "Cached At", "Size", "Description"}
	}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = lipgloss.Width(header)
	}
	rows := make([][]string, 0, len(entries))
	for _, entry := range entries {
		state := ""
		if entry.Current {
			state = "current"
		}
		var row []string
		if cachedOnly {
			row = []string{entry.VersionNumber, state, formatVersionTime(entry.CachedAt), humanVersionSize(entry.Size), entry.Description}
		} else {
			cached := "no"
			if entry.Cached {
				cached = "yes"
			}
			user := entry.UpdateUser.Email
			if user == "" {
				user = entry.UpdateUser.Name
			}
			row = []string{entry.VersionNumber, state, formatFirebaseVersionTime(entry.UpdateTime), user, friendlyVersionEnum(entry.UpdateOrigin), friendlyVersionEnum(entry.UpdateType), cached, entry.Description}
		}
		for i, cell := range row {
			widths[i] = max(widths[i], lipgloss.Width(cell))
		}
		rows = append(rows, row)
	}

	styleFunc := func(row, col int) lipgloss.Style {
		style := lipgloss.NewStyle().Padding(0, 1)
		if col == 0 || (cachedOnly && col == 3) {
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
		if row >= 0 && entries[row].Current && (col == 0 || col == 1) {
			return style.Bold(true).Foreground(clistyles.PaletteBlueBright)
		}
		cachedColumn := 6
		if cachedOnly {
			cachedColumn = 1
		}
		if row >= 0 && col == cachedColumn && ((cachedOnly && entries[row].Current) || (!cachedOnly && entries[row].Cached)) {
			return style.Foreground(clistyles.PaletteBlueBright)
		}
		if col == 0 || col == len(headers)-1 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		return style.Foreground(clistyles.PaletteSlateDim)
	}

	width := 3*len(headers) + 1
	for _, cellWidth := range widths {
		width += cellWidth
	}
	tbl := table.New().Headers(headers...).Rows(rows...).Width(width).Border(lipgloss.NormalBorder()).BorderHeader(true).BorderRow(false).StyleFunc(styleFunc)
	if !noColor {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

func formatVersionTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format("2006-01-02 15:04:05")
}

func formatFirebaseVersionTime(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return formatVersionTime(parsed)
}

func humanVersionSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
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

func friendlyVersionEnum(value string) string {
	switch value {
	case "REST_API":
		return "REST API"
	case "ADMIN_SDK_NODE":
		return "Admin SDK"
	case "INCREMENTAL_UPDATE":
		return "Update"
	case "FORCED_UPDATE":
		return "Forced"
	case "ROLLBACK":
		return "Rollback"
	case "CONSOLE":
		return "Console"
	default:
		return value
	}
}
