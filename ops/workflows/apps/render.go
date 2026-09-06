package apps

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	clistyles "github.com/yumauri/fbrcm/internal/terminal/styles"
	"github.com/yumauri/fbrcm/ops/shared"
)

const (
	nerdFontAndroidGlyph = "\uf17b" // nf-fa-android
	nerdFontIOSGlyph     = "\uf179" // nf-fa-apple
	nerdFontWebGlyph     = "\uf0ac" // nf-fa-globe
)

func renderAppsTableAtWidth(apps []core.FirebaseApp, terminalWidth int, nerdFontGlyphs bool) string {
	items := make([]appListItem, len(apps))
	for i, app := range apps {
		items[i] = appListItem{FirebaseApp: app}
	}
	return renderAppListTable(items, false, terminalWidth, nerdFontGlyphs)
}

func renderAppListTable(apps []appListItem, showProject bool, terminalWidth int, nerdFontGlyphs bool) string {
	headers := []string{"Name", "Platform", "Namespace", "App ID", "State"}
	offset := 0
	if showProject {
		headers = append([]string{"Project"}, headers...)
		offset = 1
	}
	rows := make([][]string, 0, len(apps))
	platforms := make([]core.AppPlatform, 0, len(apps))
	widths := shared.HeaderWidths(headers)
	for _, app := range apps {
		row := []string{emptyDash(app.DisplayName), platformLabel(app.Platform, nerdFontGlyphs), emptyDash(app.Namespace), app.AppID, emptyDash(app.State)}
		if showProject {
			row = append([]string{app.Project}, row...)
		}
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
		platforms = append(platforms, app.Platform)
	}
	flexible := []int{offset, 2 + offset, 3 + offset}
	if showProject {
		flexible = append(flexible, 0)
	}
	shared.FitTableColumns(widths, terminalWidth, flexible...)
	shared.TruncateTableHeaders(headers, widths)
	for row := range rows {
		rows[row][3+offset] = renderAppID(rows[row][3+offset], appTableCellStyle(row, clistyles.PanelMuted))
	}
	shared.TruncateTableColumns(rows, widths, flexible...)
	return shared.StyledTable(headers, rows, widths, nil, func(row, col int, style lipgloss.Style) lipgloss.Style {
		if showProject && col == 0 {
			return style.Foreground(clistyles.PaletteSlateBright)
		}
		if row >= 0 && row < len(platforms) && col == 1+offset {
			return style.Foreground(clistyles.FirebaseAppPlatformColor(platforms[row]))
		}
		return style
	})
}

func renderAppDetails(app core.FirebaseAppDetails, nerdFontGlyphs bool) string {
	var b strings.Builder
	line := func(label, value string) { fmt.Fprintf(&b, "%s: %s\n", label, emptyDash(value)) }
	line("Name", app.DisplayName)
	line("Platform", renderPlatformLabel(platformLabel(app.Platform, nerdFontGlyphs), app.Platform))
	line("App ID", renderAppID(app.AppID, lipgloss.NewStyle()))
	line("Namespace", app.Namespace)
	line("Resource", app.ResourceName)
	line("Project ID", app.ProjectID)
	line("State", app.State)
	line("API key ID", app.APIKeyID)
	line("Expires", app.ExpireTime)
	line("ETag", app.ETag)
	if app.PackageName != nil {
		line("Package name", *app.PackageName)
	}
	if app.BundleID != nil {
		line("Bundle ID", *app.BundleID)
	}
	if app.AppStoreID != nil {
		line("App Store ID", *app.AppStoreID)
	}
	if app.TeamID != nil {
		line("Team ID", *app.TeamID)
	}
	if app.WebID != nil {
		line("Web ID", *app.WebID)
	}
	if len(app.AppURLs) > 0 {
		line("App URLs", strings.Join(app.AppURLs, ", "))
	}
	if len(app.SHA1Hashes) > 0 {
		line("SHA-1 hashes", strings.Join(app.SHA1Hashes, ", "))
	}
	if len(app.SHA256Hashes) > 0 {
		line("SHA-256 hashes", strings.Join(app.SHA256Hashes, ", "))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func renderPlatformLabel(label string, platform core.AppPlatform) string {
	if clistyles.NoColorEnabled() {
		return label
	}
	return lipgloss.NewStyle().Foreground(clistyles.FirebaseAppPlatformColor(platform)).Render(label)
}

func renderAppID(appID string, base lipgloss.Style) string {
	return clistyles.RenderFirebaseAppPlatforms(appID, base)
}

func appTableCellStyle(row int, style lipgloss.Style) lipgloss.Style {
	if !clistyles.NoColorEnabled() && row%2 == 1 {
		return style.Background(clistyles.ColorRowStripe)
	}
	return style
}

func platformLabel(platform core.AppPlatform, nerdFontGlyphs bool) string {
	name := string(platform)
	if !nerdFontGlyphs {
		return name
	}
	var glyph string
	switch platform {
	case core.AppPlatformAndroid:
		glyph = nerdFontAndroidGlyph
	case core.AppPlatformIOS:
		glyph = nerdFontIOSGlyph
	case core.AppPlatformWeb:
		glyph = nerdFontWebGlyph
	default:
		return name
	}
	return glyph + " " + name
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}
