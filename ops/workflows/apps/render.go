package apps

import (
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/ops/shared"
)

func renderAppsTable(apps []core.FirebaseApp) string {
	return renderAppsTableAtWidth(apps, shared.TerminalWidth())
}

func renderAppsTableAtWidth(apps []core.FirebaseApp, terminalWidth int) string {
	headers := []string{"Name", "Platform", "Namespace", "App ID", "State"}
	rows := make([][]string, 0, len(apps))
	widths := shared.HeaderWidths(headers)
	for _, app := range apps {
		row := []string{emptyDash(app.DisplayName), string(app.Platform), emptyDash(app.Namespace), app.AppID, emptyDash(app.State)}
		shared.UpdateTableWidths(widths, row)
		rows = append(rows, row)
	}
	shared.FitTableColumns(widths, terminalWidth, 0, 2, 3)
	shared.TruncateTableHeaders(headers, widths)
	shared.TruncateTableColumns(rows, widths, 0, 2, 3)
	return shared.StyledTable(headers, rows, widths, nil, nil)
}

func renderAppDetails(app core.FirebaseAppDetails) string {
	var b strings.Builder
	line := func(label, value string) { fmt.Fprintf(&b, "%s: %s\n", label, emptyDash(value)) }
	line("Name", app.DisplayName)
	line("Platform", string(app.Platform))
	line("App ID", app.AppID)
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

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}
