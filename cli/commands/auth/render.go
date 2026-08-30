package auth

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
)

type authListItem struct {
	config.AuthEntry
	Default bool `json:"default"`
}

func newAuthListItems(entries []config.AuthEntry, defaultAuthID string) []authListItem {
	items := make([]authListItem, len(entries))
	for i, entry := range entries {
		items[i] = authListItem{AuthEntry: entry, Default: entry.ID == defaultAuthID}
	}
	return items
}

func renderAuthTable(entries []config.AuthEntry, defaultAuthID string, terminalWidth int) string {
	rows := make([][]string, 0, len(entries))
	idWidth := lipgloss.Width("Auth")
	typeWidth := lipgloss.Width("Type")
	labelWidth := lipgloss.Width("Label")
	quotaWidth := lipgloss.Width("Quota Project")
	defaultWidth := lipgloss.Width("Default")
	for _, entry := range entries {
		marker := ""
		if entry.ID == defaultAuthID {
			marker = "✓"
		}
		quotaProjectID := entry.QuotaProjectID
		if quotaProjectID == "" {
			quotaProjectID = "—"
		}
		rows = append(rows, []string{entry.ID, entry.Type, entry.Label, quotaProjectID, marker})
		idWidth = max(idWidth, lipgloss.Width(entry.ID))
		typeWidth = max(typeWidth, lipgloss.Width(entry.Type))
		labelWidth = max(labelWidth, lipgloss.Width(entry.Label))
		quotaWidth = max(quotaWidth, lipgloss.Width(quotaProjectID))
		defaultWidth = max(defaultWidth, lipgloss.Width(marker))
	}
	const decorationWidth = 16
	overflow := idWidth + typeWidth + labelWidth + quotaWidth + defaultWidth + decorationWidth - terminalWidth
	if terminalWidth > 0 && overflow > 0 {
		reduction := min(overflow, max(labelWidth-1, 0))
		labelWidth -= reduction
		overflow -= reduction
		if overflow > 0 {
			reduction = min(overflow, max(quotaWidth-1, 0))
			quotaWidth -= reduction
		}
		for i := range rows {
			rows[i][2] = ansi.Truncate(rows[i][2], labelWidth, "…")
			rows[i][3] = ansi.Truncate(rows[i][3], quotaWidth, "…")
		}
	}
	return shared.StyledTable(
		[]string{"Auth", "Type", ansi.Truncate("Label", labelWidth, "…"), ansi.Truncate("Quota Project", quotaWidth, "…"), "Default"},
		rows,
		[]int{idWidth, typeWidth, labelWidth, quotaWidth, defaultWidth},
		nil,
		nil,
	)
}

type authPathResult struct {
	ID                 string `json:"id"`
	Type               string `json:"type" contract:"enum=google|oauth|service-account|gcloud"`
	AuthConfigPath     string `json:"auth_config_path"`
	ProfileConfigPath  string `json:"profile_config_path"`
	ClientSecretPath   string `json:"client_secret_path,omitempty"`
	TokenPath          string `json:"token_path,omitempty"`
	ServiceAccountPath string `json:"service_account_path,omitempty"`
}

func authPathPayload(auth config.AuthEntry, paths core.AuthPaths) authPathResult {
	return authPathResult{
		ID: auth.ID, Type: auth.Type,
		AuthConfigPath: paths.AuthConfigPath, ProfileConfigPath: paths.ProfileConfigPath,
		ClientSecretPath: paths.ClientSecretPath, TokenPath: paths.TokenPath, ServiceAccountPath: paths.ServiceAccountPath,
	}
}

func authPathLines(auth config.AuthEntry, paths core.AuthPaths) []string {
	switch auth.Type {
	case config.AuthTypeOAuth:
		return nonEmptyStrings(paths.ClientSecretPath, paths.TokenPath)
	case config.AuthTypeGoogle:
		return nonEmptyStrings(paths.TokenPath)
	case config.AuthTypeServiceAccount:
		return nonEmptyStrings(paths.ServiceAccountPath)
	default:
		return nil
	}
}

func nonEmptyStrings(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
