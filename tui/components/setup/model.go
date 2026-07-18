package setup

import (
	"context"
	"os"
	"strconv"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/tui/components/inputstyles"
	"github.com/yumauri/fbrcm/tui/styles"
)

type mode int

const (
	modeHidden mode = iota
	modeChecking
	modeAccounts
	modeProfiles
	modeMethods
	modeIdentity
	modeFile
	modeAdding
	modeAuthenticating
	modeDiscovering
	modeSwitching
	modeNoProjects
	modeError
)

type authMethod int

const (
	methodOAuth authMethod = iota
	methodServiceAccount
	methodGCloud
)

type failureStage int

const (
	failureNone failureStage = iota
	failureInspect
	failureOpen
	failureAdd
	failureLogin
	failureSync
	failureSwitch
)

// WorkspaceReadyMsg asks the application to leave setup and populate the
// Projects panel with the supplied local or freshly discovered projects.
type WorkspaceReadyMsg struct {
	Projects   []core.Project
	Source     string
	CachedOnly bool
	Reset      bool
}

// CanceledMsg asks the application to close setup and return to the existing
// workspace. Mandatory first-run setup cannot emit this message.
type CanceledMsg struct{}

// QuitRequestedMsg asks the application to run its normal guarded quit flow.
// Ctrl+C remains the unconditional force-quit path.
type QuitRequestedMsg struct{}

// Model is the interactive authentication and project-discovery setup gate.
type Model struct {
	svc *core.Core

	mode            mode
	initial         bool
	mandatory       bool
	profile         string
	profileOverride string
	profiles        []string
	profileNew      bool
	profileTo       string
	auth            []config.AuthEntry
	defaultID       string
	method          authMethod
	cursor          int
	error           error
	failure         failureStage
	authID          string
	syncAuthID      string
	filePath        string
	loginBack       mode
	loginStop       context.CancelFunc
	loginID         uint64
	syncBack        mode
	syncStop        context.CancelFunc
	syncID          uint64

	filepicker filepicker.Model
	identity   textinput.Model
	profileIn  textinput.Model
	spinner    spinner.Model
}

// New creates startup setup. A nil service keeps setup disabled for isolated
// view/component tests.
func New(svc *core.Core) Model {
	picker := filepicker.New()
	picker.AllowedTypes = []string{".json"}
	picker.FileAllowed = true
	picker.DirAllowed = false
	picker.ShowHidden = true
	picker.ShowPermissions = true
	picker.ShowSize = true
	picker.AutoHeight = false
	if cwd, err := os.Getwd(); err == nil {
		picker.CurrentDirectory = cwd
	}

	identity := textinput.New()
	identity.Prompt = ""
	identity.Placeholder = "identity name"
	identity.CharLimit = 64
	identity.SetStyles(inputstyles.TextInput())
	identity.SetWidth(36)
	identity.Blur()
	profileIn := textinput.New()
	profileIn.Prompt = ""
	profileIn.Placeholder = "New profile"
	profileIn.CharLimit = 64
	profileIn.SetStyles(inputstyles.InlineListTextInput())
	profileIn.SetWidth(36)
	profileIn.Blur()

	spin := spinner.New(spinner.WithSpinner(spinner.Line))
	spin.Style = styles.SecondaryTitleSpinner

	m := Model{
		svc:        svc,
		mode:       modeHidden,
		initial:    true,
		mandatory:  true,
		filepicker: picker,
		identity:   identity,
		profileIn:  profileIn,
		spinner:    spin,
	}
	if svc != nil {
		m.mode = modeChecking
	}
	return m
}

// Init inspects local startup state before the Projects panel starts loading.
func (m Model) Init() tea.Cmd {
	if m.svc == nil {
		return nil
	}
	return tea.Batch(m.inspectCmd(), m.spinner.Tick)
}

// Open reopens authentication management from an existing workspace.
func (m Model) Open() (Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	m.mode = modeChecking
	m.initial = false
	m.mandatory = false
	m.profileNew = false
	m.cursor = 0
	m.error = nil
	m.failure = failureNone
	m.cancelLogin()
	m.cancelSync()
	return m, tea.Batch(m.inspectCmd(), m.spinner.Tick)
}

// Close hides setup.
func (m Model) Close() Model {
	m.cancelLogin()
	m.cancelSync()
	m.mode = modeHidden
	m.identity.Blur()
	m.profileIn.Blur()
	return m
}

// IsOpen reports whether setup currently replaces the normal workspace.
func (m Model) IsOpen() bool { return m.mode != modeHidden }

func (m Model) methodName() string {
	switch m.method {
	case methodOAuth:
		return "OAuth desktop login"
	case methodServiceAccount:
		return "Service account"
	case methodGCloud:
		return "Existing gcloud credentials"
	default:
		return "Authentication"
	}
}

func authTypeLabel(value string) string {
	switch value {
	case config.AuthTypeOAuth:
		return "OAuth"
	case config.AuthTypeServiceAccount:
		return "Service account"
	case config.AuthTypeGCloud:
		return "gcloud ADC"
	default:
		return value
	}
}

func (m Model) suggestedAuthID() string {
	if len(m.auth) == 0 {
		return "default"
	}
	used := make(map[string]struct{}, len(m.auth))
	for _, entry := range m.auth {
		used[entry.ID] = struct{}{}
	}
	for index := 2; ; index++ {
		candidate := "account-" + strconv.Itoa(index)
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func (m *Model) upsertAuth(entry config.AuthEntry) {
	for index := range m.auth {
		if m.auth[index].ID == entry.ID {
			m.auth[index] = entry
			return
		}
	}
	m.auth = append(m.auth, entry)
}

func (m Model) selectedAccountID() string {
	if m.cursor < 0 || m.cursor >= len(m.auth) {
		return ""
	}
	return m.auth[m.cursor].ID
}

func selectedLine(value string, selected bool) string {
	if !selected {
		return value
	}
	if styles.NoColorEnabled() {
		return lipgloss.NewStyle().Bold(true).Reverse(true).Render(value)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(styles.PaletteYellow).Background(styles.PaletteBlueDeep).Render(value)
}
