package setup

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/config"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case inspectedMsg:
		return m.updateInspected(msg)
	case authAddedMsg:
		return m.updateAuthAdded(msg)
	case authReadyMsg:
		return m.updateAuthReady(msg)
	case projectsSyncedMsg:
		return m.updateProjectsSynced(msg)
	case profileSwitchedMsg:
		return m.updateProfileSwitched(msg)
	case externalOpenedMsg:
		if msg.err != nil {
			m.setFailure(failureOpen, fmt.Errorf("open OAuth client page: %w", msg.err))
		}
		return m, nil
	case spinner.TickMsg:
		if !m.working() {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.filepicker.SetHeight(min(max(msg.Height-12, 5), 18))
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(msg)
	default:
		if m.mode == modeIdentity {
			var cmd tea.Cmd
			m.identity, cmd = m.identity.Update(msg)
			return m, cmd
		}
		if m.mode == modeProfiles && m.profileInputSelected() {
			var cmd tea.Cmd
			m.profileIn, cmd = m.profileIn.Update(msg)
			return m, cmd
		}
		if m.mode == modeFile {
			return m.updateFilepicker(msg)
		}
	}
	return m, nil
}

func (m Model) updateInspected(msg inspectedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.mandatory = m.initial
		m.setFailure(failureInspect, fmt.Errorf("read local setup state: %w", msg.err))
		return m, nil
	}
	m.profile = msg.state.Profile
	m.profiles = append([]string(nil), msg.state.Profiles...)
	m.auth = append([]config.AuthEntry(nil), msg.state.Auth...)
	m.defaultID = msg.state.DefaultAuthID
	m.error = nil
	m.failure = failureNone
	m.cursor = 0

	if !m.initial {
		m.mandatory = false
		if len(m.auth) == 0 {
			m.mode = modeMethods
		} else {
			m.mode = modeAccounts
		}
		return m, nil
	}

	if len(msg.state.Projects) > 0 {
		cachedOnly := len(m.auth) == 0
		reset := m.profileNew
		return m, func() tea.Msg {
			return WorkspaceReadyMsg{Projects: msg.state.Projects, Source: "cache", CachedOnly: cachedOnly, Reset: reset}
		}
	}
	if len(m.auth) == 0 {
		m.mandatory = true
		m.mode = modeMethods
		return m, nil
	}

	m.mandatory = true
	m.mode = modeDiscovering
	m.syncAuthID = ""
	syncCmd := m.startSync(modeAccounts)
	return m, tea.Batch(syncCmd, m.spinner.Tick)
}

func (m Model) updateAuthAdded(msg authAddedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.setFailure(failureAdd, fmt.Errorf("add authentication: %w", msg.err))
		return m, nil
	}
	m.upsertAuth(msg.entry)
	if m.defaultID == "" {
		m.defaultID = msg.entry.ID
	}
	m.authID = msg.entry.ID
	m.mode = modeAuthenticating
	m.error = nil
	m.failure = failureNone
	back := modeMethods
	if m.method == methodOAuth || m.method == methodServiceAccount {
		back = modeFile
	}
	loginCmd := m.startLogin(back)
	return m, tea.Batch(loginCmd, m.spinner.Tick, clearScreenCmd())
}

func (m Model) updateAuthReady(msg authReadyMsg) (Model, tea.Cmd) {
	if msg.loginID != m.loginID {
		return m, nil
	}
	m.loginStop = nil
	if msg.err != nil {
		m.setFailure(failureLogin, fmt.Errorf("authenticate %s: %w", m.authID, msg.err))
		return m, nil
	}
	m.mode = modeDiscovering
	m.error = nil
	m.failure = failureNone
	m.syncAuthID = m.authID
	syncCmd := m.startSync(modeAccounts)
	return m, tea.Batch(syncCmd, m.spinner.Tick)
}

func (m Model) updateProjectsSynced(msg projectsSyncedMsg) (Model, tea.Cmd) {
	if msg.syncID != m.syncID {
		return m, nil
	}
	m.syncStop = nil
	if msg.err != nil {
		m.setFailure(failureSync, fmt.Errorf("discover Firebase projects: %w", msg.err))
		return m, nil
	}
	m.error = nil
	m.failure = failureNone
	if len(msg.projects) == 0 {
		m.mode = modeNoProjects
		m.cursor = 0
		return m, nil
	}
	projects := append([]config.Project(nil), msg.projects...)
	reset := m.profileNew
	return m, func() tea.Msg {
		return WorkspaceReadyMsg{Projects: projects, Source: msg.source, Reset: reset}
	}
}

func (m Model) updateProfileSwitched(msg profileSwitchedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.setFailure(failureSwitch, fmt.Errorf("switch profile: %w", msg.err))
		return m, nil
	}
	m.initial = true
	m.mandatory = true
	m.profileNew = true
	m.mode = modeChecking
	m.error = nil
	m.failure = failureNone
	return m, tea.Batch(m.inspectCmd(), m.spinner.Tick)
}

func (m Model) updateKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	k := msg.String()
	if k == "ctrl+c" {
		return m, tea.Quit
	}
	switch m.mode {
	case modeAccounts:
		return m.updateAccountsKey(k)
	case modeProfiles:
		return m.updateProfilesKey(msg, k)
	case modeMethods:
		return m.updateMethodsKey(k)
	case modeIdentity:
		return m.updateIdentityKey(msg, k)
	case modeFile:
		return m.updateFileKey(msg, k)
	case modeAuthenticating:
		return m.updateAuthenticatingKey(k)
	case modeDiscovering:
		return m.updateDiscoveringKey(k)
	case modeNoProjects:
		return m.updateNoProjectsKey(k)
	case modeError:
		return m.updateErrorKey(k)
	case modeChecking, modeAdding, modeSwitching:
		if k == "q" && m.mandatory {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) updateAccountsKey(k string) (Model, tea.Cmd) {
	switch k {
	case "p":
		return m, m.openProfiles()
	case "up", "k":
		m.moveCursor(-1, len(m.auth)+1)
	case "down", "j":
		m.moveCursor(1, len(m.auth)+1)
	case "enter":
		if m.cursor == len(m.auth) {
			m.mode = modeMethods
			m.cursor = 0
			return m, nil
		}
		m.authID = m.selectedAccountID()
		if m.authID == "" {
			return m, nil
		}
		m.mode = modeAuthenticating
		loginCmd := m.startLogin(modeAccounts)
		return m, tea.Batch(loginCmd, m.spinner.Tick)
	case "esc":
		if !m.mandatory {
			return m, func() tea.Msg { return CanceledMsg{} }
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateMethodsKey(k string) (Model, tea.Cmd) {
	switch k {
	case "p":
		return m, m.openProfiles()
	case "up", "k":
		m.moveCursor(-1, 3)
	case "down", "j":
		m.moveCursor(1, 3)
	case "enter":
		m.method = authMethod(m.cursor)
		m.authID = m.suggestedAuthID()
		if len(m.auth) > 0 {
			m.mode = modeIdentity
			m.identity.SetValue(m.authID)
			m.identity.CursorEnd()
			return m, m.identity.Focus()
		}
		return m.continueSelectedMethod()
	case "esc":
		if len(m.auth) > 0 {
			m.mode = modeAccounts
			m.cursor = 0
			return m, nil
		}
		if !m.mandatory {
			return m, func() tea.Msg { return CanceledMsg{} }
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateProfilesKey(msg tea.KeyMsg, k string) (Model, tea.Cmd) {
	switch k {
	case "up", "k":
		return m, m.moveProfileCursor(-1)
	case "down", "j":
		return m, m.moveProfileCursor(1)
	case "enter":
		if m.profileInputSelected() {
			return m.submitNewProfile()
		}
		if m.cursor < 0 || m.cursor >= len(m.profiles) {
			return m, nil
		}
		selected := m.profiles[m.cursor]
		if selected == m.profile {
			return m.returnFromProfiles()
		}
		m.profileTo = selected
		m.mode = modeSwitching
		return m, tea.Batch(m.switchProfileCmd(), m.spinner.Tick)
	case "esc":
		m.profileIn.Blur()
		return m.returnFromProfiles()
	case "q":
		if !m.profileInputSelected() {
			return m, tea.Quit
		}
	}
	if m.profileInputSelected() {
		m.profileIn.Err = nil
		var cmd tea.Cmd
		m.profileIn, cmd = m.profileIn.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) submitNewProfile() (Model, tea.Cmd) {
	value := strings.TrimSpace(m.profileIn.Value())
	if value == "" {
		return m, nil
	}
	if err := config.ValidateProfileName(value); err != nil {
		m.profileIn.Err = err
		return m, nil
	}
	if slices.Contains(m.profiles, value) {
		m.profileIn.Err = fmt.Errorf("profile %q already exists", value)
		return m, nil
	}
	m.profileIn.Err = nil
	m.profileIn.Blur()
	m.profileTo = value
	m.mode = modeSwitching
	return m, tea.Batch(m.switchProfileCmd(), m.spinner.Tick)
}

func (m Model) returnFromProfiles() (Model, tea.Cmd) {
	if len(m.auth) > 0 {
		m.mode = modeAccounts
		m.cursor = 0
	} else {
		m.mode = modeMethods
		m.cursor = 0
	}
	return m, nil
}

func (m *Model) openProfiles() tea.Cmd {
	m.mode = modeProfiles
	m.cursor = m.currentProfileIndex()
	m.profileIn.SetValue("")
	m.profileIn.Err = nil
	m.profileIn.Blur()
	if m.profileInputSelected() {
		return m.profileIn.Focus()
	}
	return nil
}

func (m Model) profileInputSelected() bool {
	return m.cursor == len(m.profiles)
}

func (m *Model) moveProfileCursor(delta int) tea.Cmd {
	wasInput := m.profileInputSelected()
	m.moveCursor(delta, len(m.profiles)+1)
	if m.profileInputSelected() && !wasInput {
		return m.profileIn.Focus()
	}
	if wasInput && !m.profileInputSelected() {
		m.profileIn.Blur()
	}
	return nil
}

func (m Model) updateIdentityKey(msg tea.KeyMsg, k string) (Model, tea.Cmd) {
	switch k {
	case "esc":
		m.identity.Blur()
		m.mode = modeMethods
		m.cursor = int(m.method)
		return m, nil
	case "enter":
		value := strings.TrimSpace(m.identity.Value())
		if err := config.ValidateAuthID(value); err != nil {
			m.identity.Err = err
			return m, nil
		}
		if slices.ContainsFunc(m.auth, func(entry config.AuthEntry) bool { return entry.ID == value }) {
			m.identity.Err = fmt.Errorf("auth identity %q already exists", value)
			return m, nil
		}
		m.identity.Err = nil
		m.identity.Blur()
		m.authID = value
		return m.continueSelectedMethod()
	}
	var cmd tea.Cmd
	m.identity, cmd = m.identity.Update(msg)
	return m, cmd
}

func (m Model) continueSelectedMethod() (Model, tea.Cmd) {
	m.error = nil
	m.failure = failureNone
	if m.method == methodGCloud {
		m.mode = modeAdding
		return m, tea.Batch(m.addAuthCmd(), m.spinner.Tick)
	}
	m.mode = modeFile
	return m, m.filepicker.Init()
}

func (m Model) updateFileKey(msg tea.KeyMsg, k string) (Model, tea.Cmd) {
	if k == "esc" || k == "q" {
		if len(m.auth) > 0 {
			m.mode = modeIdentity
			m.identity.SetValue(m.authID)
			return m, m.identity.Focus()
		}
		m.mode = modeMethods
		m.cursor = int(m.method)
		return m, nil
	}
	if m.method == methodOAuth && k == "o" {
		return m, openOAuthClientsCmd()
	}
	return m.updateFilepicker(msg)
}

func (m Model) updateFilepicker(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)
	if selected, path := m.filepicker.DidSelectFile(msg); selected {
		m.filePath = path
		m.mode = modeAdding
		return m, tea.Batch(m.addAuthCmd(), m.spinner.Tick, clearScreenCmd())
	}
	return m, cmd
}

func (m Model) updateAuthenticatingKey(k string) (Model, tea.Cmd) {
	switch k {
	case "esc":
		m.cancelLogin()
		m.mode = m.loginBack
		m.error = nil
		m.failure = failureNone
		return m, clearScreenCmd()
	case "q":
		m.cancelLogin()
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateDiscoveringKey(k string) (Model, tea.Cmd) {
	switch k {
	case "esc":
		m.cancelSync()
		m.mode = m.syncBack
		m.error = nil
		m.failure = failureNone
		return m, clearScreenCmd()
	case "q":
		m.cancelSync()
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateNoProjectsKey(k string) (Model, tea.Cmd) {
	switch k {
	case "up", "k":
		m.moveCursor(-1, 3)
	case "down", "j":
		m.moveCursor(1, 3)
	case "enter":
		switch m.cursor {
		case 0:
			m.mode = modeDiscovering
			syncCmd := m.startSync(modeNoProjects)
			return m, tea.Batch(syncCmd, m.spinner.Tick)
		case 1:
			m.mode = modeMethods
			m.cursor = 0
		case 2:
			reset := m.profileNew
			return m, func() tea.Msg { return WorkspaceReadyMsg{Source: "firebase", Reset: reset} }
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) updateErrorKey(k string) (Model, tea.Cmd) {
	switch k {
	case "r":
		m.error = nil
		switch m.failure {
		case failureInspect:
			m.mode = modeChecking
			return m, tea.Batch(m.inspectCmd(), m.spinner.Tick)
		case failureOpen:
			m.mode = modeFile
			return m, openOAuthClientsCmd()
		case failureAdd:
			m.mode = modeAdding
			return m, tea.Batch(m.addAuthCmd(), m.spinner.Tick)
		case failureLogin:
			m.mode = modeAuthenticating
			loginCmd := m.startLogin(m.loginBack)
			return m, tea.Batch(loginCmd, m.spinner.Tick)
		case failureSync:
			m.mode = modeDiscovering
			syncCmd := m.startSync(m.syncBack)
			return m, tea.Batch(syncCmd, m.spinner.Tick)
		case failureSwitch:
			m.mode = modeSwitching
			return m, tea.Batch(m.switchProfileCmd(), m.spinner.Tick)
		}
	case "esc":
		switch m.failure {
		case failureInspect:
			if !m.mandatory {
				return m, func() tea.Msg { return CanceledMsg{} }
			}
		case failureOpen:
			m.mode = modeFile
		case failureAdd:
			if m.method == methodGCloud {
				m.mode = modeMethods
				m.cursor = int(m.method)
			} else {
				m.mode = modeFile
			}
		case failureLogin:
			m.mode = m.loginBack
			m.error = nil
			m.failure = failureNone
			return m, clearScreenCmd()
		case failureSync:
			if len(m.auth) > 0 {
				m.mode = modeAccounts
				m.cursor = 0
			} else {
				m.mode = modeMethods
			}
		case failureSwitch:
			m.mode = modeProfiles
			m.cursor = m.currentProfileIndex()
		}
		m.error = nil
		m.failure = failureNone
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) moveCursor(delta, count int) {
	if count <= 0 {
		m.cursor = 0
		return
	}
	m.cursor = (m.cursor + delta + count) % count
}

func (m *Model) setFailure(stage failureStage, err error) {
	m.failure = stage
	m.error = err
	m.mode = modeError
}

func (m Model) working() bool {
	switch m.mode {
	case modeChecking, modeAdding, modeAuthenticating, modeDiscovering, modeSwitching:
		return true
	default:
		return false
	}
}

func (m Model) currentProfileIndex() int {
	for index, profile := range m.profiles {
		if profile == m.profile {
			return index
		}
	}
	return 0
}
