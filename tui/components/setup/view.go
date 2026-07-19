package setup

import (
	"fmt"
	"strings"

	bubbleskey "charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/tui/components/viewutil"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/styles"
)

var (
	cardBorderStyle = lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)
	cardTextStyle   = styles.PanelText
	cardMutedStyle  = styles.PanelMuted
	cardErrorStyle  = lipgloss.NewStyle().Foreground(styles.PaletteError)
	cardOKStyle     = lipgloss.NewStyle().Foreground(styles.PaletteYellow)
)

// View renders setup as the only application surface while it is open.
func (m Model) View(width, height int) string {
	if !m.IsOpen() || width <= 0 || height <= 0 {
		return ""
	}
	card := m.PopupViewWithFocus(width, height, true)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, card)
}

// PopupView renders the setup card without a terminal-sized background.
func (m Model) PopupView(width, height int) string {
	return m.PopupViewWithFocus(width, height, true)
}

// PopupViewWithFocus renders the setup card and controls whether its active
// tab receives the selected-title background.
func (m Model) PopupViewWithFocus(width, height int, focused bool) string {
	if !m.IsOpen() || width <= 0 || height <= 0 {
		return ""
	}
	contentWidth := min(max(width-12, 48), 72)
	var title string
	var lines []string
	tabs := false
	switch m.mode {
	case modeChecking:
		title = "Starting fbrcm"
		lines = m.workingLines("Checking profile, authentication, and project cache…")
	case modeAccounts:
		tabs = true
		lines = m.accountsLines(contentWidth)
	case modeProfiles:
		tabs = true
		lines = m.profilesLines(contentWidth)
	case modeMethods:
		title = "Authenticate"
		lines = m.methodsLines(contentWidth)
	case modeIdentity:
		title = "Name authentication"
		lines = m.identityLines(contentWidth)
	case modeFile:
		title = m.methodName()
		lines = m.fileLines(contentWidth)
	case modeAdding:
		title = m.methodName()
		lines = m.workingLines("Validating and importing credentials…")
	case modeAuthenticating:
		title = m.methodName()
		message := "Validating credentials…"
		if m.authMethodForID(m.authID) == methodOAuth {
			message = "Waiting for browser authorization…"
		}
		lines = m.workingLines(message)
		if m.authMethodForID(m.authID) == methodOAuth {
			lines = append(lines,
				"",
				cardMutedStyle.Render("A browser window should open. fbrcm waits for the local callback."),
				cardMutedStyle.Render("If authorization cannot continue, close the page and cancel here."),
			)
		}
		lines = append(lines, "", setupHelp(contentWidth,
			[2]string{"esc", "cancel"},
			[2]string{"q", "quit"},
		))
	case modeDiscovering:
		title = "Discover Firebase projects"
		lines = m.workingLines("Discovering accessible Firebase projects…")
		lines = append(lines, "", setupHelp(contentWidth,
			[2]string{"esc", "cancel"},
			[2]string{"q", "quit"},
		))
	case modeSwitching:
		title = "Switch profile"
		lines = m.workingLines("Switching to profile " + m.profileTo + "…")
	case modePurgingAuth:
		title = "Purge authentication"
		lines = m.workingLines("Purging authentication " + m.authID + "…")
	case modePurgingProfile:
		title = "Purge profile"
		lines = m.workingLines("Purging profile " + m.profileFrom + "…")
	case modeNoProjects:
		title = "No projects found"
		lines = m.noProjectsLines(contentWidth)
	case modeError:
		title = "Setup problem"
		lines = m.errorLines(contentWidth)
	}

	if tabs {
		return renderSetupTabsPanel(m.mode == modeAccounts, focused, lines, contentWidth)
	}
	return renderSetupPanel(title, lines, contentWidth)
}

func renderSetupPanel(title string, body []string, innerWidth int) string {
	titleRendered, titleWidth := styles.PanelHeaderTitle("", title, true, max(innerWidth-1, 0))
	topFillWidth := max(innerWidth-titleWidth-1, 0)
	lines := []string{
		cardBorderStyle.Render("╭─") +
			titleRendered +
			cardBorderStyle.Render(strings.Repeat("─", topFillWidth)+"╮"),
	}

	for line := range strings.SplitSeq(strings.Join(body, "\n"), "\n") {
		line = ansi.Truncate(line, innerWidth, "")
		line += strings.Repeat(" ", max(innerWidth-lipgloss.Width(line), 0))
		lines = append(lines, cardBorderStyle.Render("│")+line+cardBorderStyle.Render("│"))
	}
	lines = append(lines, cardBorderStyle.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(lines, "\n")
}

func renderSetupTabsPanel(accountsSelected, focused bool, body []string, innerWidth int) string {
	accountKey, accountTitle := setupTabTitle(tuiconfig.ActionAccounts, "A", "Accounts")
	profileKey, profileTitle := setupTabTitle(tuiconfig.ActionProfiles, "P", "Profiles")
	accounts, accountsWidth := styles.PanelHeaderTab(accountKey, accountTitle, accountsSelected, focused, max(innerWidth-1, 0))
	profiles, profilesWidth := styles.PanelHeaderTab(profileKey, profileTitle, !accountsSelected, focused, max(innerWidth-accountsWidth-3, 0))
	fill := max(innerWidth-accountsWidth-profilesWidth-3, 0)
	lines := []string{cardBorderStyle.Render("╭─") + accounts + cardBorderStyle.Render("──") + profiles + cardBorderStyle.Render(strings.Repeat("─", fill)+"╮")}
	for line := range strings.SplitSeq(strings.Join(body, "\n"), "\n") {
		line = ansi.Truncate(line, innerWidth, "")
		line += strings.Repeat(" ", max(innerWidth-lipgloss.Width(line), 0))
		lines = append(lines, cardBorderStyle.Render("│")+line+cardBorderStyle.Render("│"))
	}
	lines = append(lines, cardBorderStyle.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(lines, "\n")
}

func setupTabTitle(action tuiconfig.Action, defaultKey, title string) (string, string) {
	keys := tuiconfig.Keys(tuiconfig.BlockGlobal, action)
	if len(keys) == 0 {
		return "", title
	}
	if keys[0] == defaultKey {
		return defaultKey, strings.TrimPrefix(title, defaultKey)
	}
	return tuiconfig.KeyHint(keys[0]), title
}

func (m Model) methodsLines(width int) []string {
	lines := []string{
		"Connect Google credentials to discover your Firebase projects.",
		"",
		cardMutedStyle.Render("Profile: ") + cardTextStyle.Render(m.profileOrDefault()),
		"",
		setupListLine("OAuth desktop login", m.cursor == int(methodOAuth)),
		cardMutedStyle.Render("    Choose a desktop client JSON, then sign in in a browser"),
		"",
		setupListLine("Service account", m.cursor == int(methodServiceAccount)),
		cardMutedStyle.Render("    Import a service account JSON key"),
		"",
		setupListLine("Existing gcloud credentials", m.cursor == int(methodGCloud)),
		cardMutedStyle.Render("    Use Application Default Credentials already on this machine"),
		"",
		setupHelp(width,
			[2]string{"↑/↓", "select"},
			[2]string{"enter", "continue"},
			[2]string{tuiconfig.Label(tuiconfig.BlockGlobal, tuiconfig.ActionProfiles), "profiles"},
			[2]string{"esc", "back"},
			[2]string{"q", "quit"},
		),
	}
	return lines
}

func (m Model) accountsLines(width int) []string {
	profileLabel := m.profileOrDefault()
	if m.profileOverride != "" {
		profileLabel += "  ·  pinned by FBRCM_PROFILE"
	}
	lines := []string{
		cardMutedStyle.Render("Profile: ") + cardTextStyle.Render(profileLabel),
		"",
		"Configured authentication:",
		"",
	}
	for index, entry := range m.auth {
		label := fmt.Sprintf("%s  ·  %s", entry.ID, authTypeLabel(entry.Type))
		if entry.ID == m.defaultID {
			label += "  ·  default"
		}
		if count := m.boundProjects(entry.ID); count == 1 {
			label += "  ·  1 project"
		} else if count > 1 {
			label += fmt.Sprintf("  ·  %d projects", count)
		} else {
			label += "  ·  unused"
		}
		lines = append(lines, setupListLine(label, m.cursor == index))
	}
	lines = append(lines,
		setupListLine("+ add authentication", m.cursor == len(m.auth)),
		"",
		setupHelp(width,
			[2]string{"↑/↓", "select"},
			[2]string{"enter", "validate/sign in"},
			[2]string{"x", "purge"},
			[2]string{"tab/→", "profiles"},
			[2]string{"esc", "workspace"},
			[2]string{"q", "quit"},
		),
	)
	if m.error != nil {
		lines = append(lines[:len(lines)-1], cardErrorStyle.Render(m.error.Error()), "", lines[len(lines)-1])
	}
	return lines
}

func (m Model) profilesLines(width int) []string {
	lines := []string{
		"Profiles keep authentication, projects, caches, and drafts separate.",
		"",
	}
	if m.profileOverride != "" {
		lines = append(lines,
			cardMutedStyle.Render("Profile selection is pinned by FBRCM_PROFILE for this process."),
			cardMutedStyle.Render("Restart fbrcm without it to create or switch profiles."),
			"",
		)
		for _, profile := range m.profiles {
			label := profile
			if profile == m.profile {
				label += "  ·  active  ·  pinned"
			}
			lines = append(lines, "  "+label)
		}
		lines = append(lines, "", setupHelp(width,
			[2]string{"tab/←", "accounts"},
			[2]string{"esc", "workspace"},
			[2]string{"q", "quit"},
		))
		return lines
	}
	for index, profile := range m.profiles {
		label := profile
		if profile == m.profile {
			label += "  ·  active"
		}
		lines = append(lines, setupListLine(label, m.cursor == index))
	}
	input := m.profileIn
	input.SetWidth(max(width-2, 1))
	lines = append(lines, "  "+input.View())
	if m.profileIn.Err != nil {
		lines = append(lines, "  "+cardErrorStyle.Render(m.profileIn.Err.Error()))
	}
	hints := [][2]string{
		{"↑/↓", "select"},
		{"enter", "switch/create"},
		{"r", "rename"},
		{"x", "purge"},
		{"tab/←", "accounts"},
		{"esc", "workspace"},
	}
	if !m.profileInputSelected() {
		hints = append(hints, [2]string{"q", "quit"})
	}
	lines = append(lines, "", setupHelp(width, hints...))
	if m.error != nil {
		lines = append(lines[:len(lines)-1], cardErrorStyle.Render(m.error.Error()), "", lines[len(lines)-1])
	}
	return lines
}

func (m Model) identityLines(width int) []string {
	lines := []string{
		"Give this authentication identity a short name.",
		"Projects can use different identities later.",
		"",
		cardMutedStyle.Render("Method: ") + cardTextStyle.Render(m.methodName()),
		cardMutedStyle.Render("Identity: ") + m.identity.View(),
	}
	if m.identity.Err != nil {
		lines = append(lines, cardErrorStyle.Render(m.identity.Err.Error()))
	}
	lines = append(lines, "", setupHelp(width,
		[2]string{"enter", "continue"},
		[2]string{"esc", "back"},
	))
	return lines
}

func (m Model) fileLines(width int) []string {
	label := "Choose a service account JSON key."
	if m.method == methodOAuth {
		label = "Choose a Desktop OAuth client JSON."
	}
	lines := []string{
		label,
		"",
		cardMutedStyle.Render("Identity: ") + cardTextStyle.Render(m.authID),
		"",
	}
	if m.method == methodOAuth {
		lines = append(lines,
			cardMutedStyle.Render("Need one? Press ")+styles.FilterText.Render("o")+cardMutedStyle.Render(" to open Google Cloud OAuth clients."),
			"",
		)
	}
	hints := [][2]string{
		{"enter/l", "select"},
		{"h/left", "parent"},
	}
	if m.method == methodOAuth {
		hints = append(hints, [2]string{"o", "OAuth clients"})
	}
	hints = append(hints, [2]string{"esc", "back"})
	picker := strings.TrimRight(m.filepicker.View(), "\n")
	lines = append(lines, viewutil.IndentLines(picker, 1), "", setupHelp(width, hints...))
	return lines
}

func (m Model) workingLines(message string) []string {
	return []string{
		m.spinner.View() + " " + cardTextStyle.Render(message),
		"",
		cardMutedStyle.Render("Profile: ") + cardTextStyle.Render(m.profileOrDefault()),
	}
}

func (m Model) noProjectsLines(width int) []string {
	return []string{
		cardOKStyle.Render("✓ Authentication is valid."),
		"",
		"The credentials did not return any accessible Firebase projects.",
		"Check project access and the Cloud Resource Manager API, then retry.",
		"",
		selectedLine("Try again", m.cursor == 0),
		selectedLine("Add another identity", m.cursor == 1),
		selectedLine("Open empty workspace", m.cursor == 2),
		"",
		setupHelp(width,
			[2]string{"↑/↓", "select"},
			[2]string{"enter", "continue"},
			[2]string{"q", "quit"},
		),
	}
}

func (m Model) errorLines(width int) []string {
	detail := "Unknown setup error"
	if m.error != nil {
		detail = m.error.Error()
	}
	detail = ansi.Wrap(detail, max(width-4, 20), "")
	lines := []string{
		cardErrorStyle.Render(detail),
		"",
		setupHelp(width,
			[2]string{"r", "retry"},
			[2]string{"esc", "back"},
			[2]string{"q", "quit"},
		),
	}
	if m.failure == failureInspect {
		lines = append(lines, "", cardMutedStyle.Render("Existing configuration was not changed."))
	}
	return lines
}

func setupHelp(width int, items ...[2]string) string {
	bindings := make([]bubbleskey.Binding, 0, len(items))
	for _, item := range items {
		bindings = append(bindings, viewutil.HelpBinding(item[0], item[1]))
	}
	return viewutil.ShortHelpView(width, bindings...)
}

func setupListLine(value string, selected bool) string {
	return viewutil.SelectorLine(value, selected)
}

func (m Model) profileOrDefault() string {
	if strings.TrimSpace(m.profile) == "" {
		return "default"
	}
	return m.profile
}

func (m Model) authMethodForID(id string) authMethod {
	for _, entry := range m.auth {
		if entry.ID != id {
			continue
		}
		switch entry.Type {
		case "oauth":
			return methodOAuth
		case "service-account":
			return methodServiceAccount
		case "gcloud":
			return methodGCloud
		}
	}
	return m.method
}
