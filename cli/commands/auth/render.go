package auth

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	clistyles "github.com/yumauri/fbrcm/cli/styles"
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

func renderAuthTable(entries []config.AuthEntry, defaultAuthID string) string {
	rows := make([][]string, 0, len(entries))
	idWidth := lipgloss.Width("Auth")
	typeWidth := lipgloss.Width("Type")
	labelWidth := lipgloss.Width("Label")
	defaultWidth := lipgloss.Width("Default")
	for _, entry := range entries {
		marker := ""
		if entry.ID == defaultAuthID {
			marker = "✓"
		}
		rows = append(rows, []string{entry.ID, entry.Type, entry.Label, marker})
		idWidth = max(idWidth, lipgloss.Width(entry.ID))
		typeWidth = max(typeWidth, lipgloss.Width(entry.Type))
		labelWidth = max(labelWidth, lipgloss.Width(entry.Label))
		defaultWidth = max(defaultWidth, lipgloss.Width(marker))
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
		Headers("Auth", "Type", "Label", "Default").
		Rows(rows...).
		Width(idWidth + typeWidth + labelWidth + defaultWidth + 13).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderRow(false).
		StyleFunc(styleFunc)
	if !clistyles.NoColorEnabled() {
		tbl = tbl.BorderStyle(clistyles.BorderStyle(false))
	}
	return tbl.String()
}

type authPathResult struct {
	ID                 string `json:"id"`
	Type               string `json:"type" contract:"enum=oauth|service-account|gcloud"`
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
