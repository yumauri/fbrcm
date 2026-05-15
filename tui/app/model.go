package app

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	boolpicker "github.com/yumauri/fbrcm/tui/components/boolpicker"
	"github.com/yumauri/fbrcm/tui/components/details"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	jsoninput "github.com/yumauri/fbrcm/tui/components/jsoninput"
	"github.com/yumauri/fbrcm/tui/components/logs"
	moveparam "github.com/yumauri/fbrcm/tui/components/moveparam"
	numberinput "github.com/yumauri/fbrcm/tui/components/numberinput"
	"github.com/yumauri/fbrcm/tui/components/parameters"
	"github.com/yumauri/fbrcm/tui/components/projects"
	renameinput "github.com/yumauri/fbrcm/tui/components/renameinput"
	stringinput "github.com/yumauri/fbrcm/tui/components/stringinput"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

// Model holds model state used by the app package.
type Model struct {
	// svc stores svc for Model.
	svc *core.Core

	// projects stores projects for Model.
	projects projects.Model
	// parameters stores parameters for Model.
	parameters parameters.Model
	// details stores details for Model.
	details details.Model
	// logs stores logs for Model.
	logs logs.Model
	// projectsMode stores projects mode for Model.
	projectsMode projectsPanelMode
	// logsHeight stores logs height for Model.
	logsHeight int
	// logsSized stores logs sized for Model.
	logsSized bool
	// logsMode stores logs mode for Model.
	logsMode logsPanelMode
	// logsSaved stores logs saved for Model.
	logsSaved int
	// help stores help for Model.
	help help.Model
	// active stores active for Model.
	active panels.ID
	// prevTop stores prev top for Model.
	prevTop panels.ID
	// capture stores capture for Model.
	capture panels.ID
	// detailsVisible stores details visible for Model.
	detailsVisible bool
	// dialog stores dialog for Model.
	dialog dialogcmp.Model
	// jsonInput stores json input for Model.
	jsonInput jsoninput.Model
	// boolPicker stores bool picker for Model.
	boolPicker boolpicker.Model
	// numberInput stores number input for Model.
	numberInput numberinput.Model
	// stringInput stores string input for Model.
	stringInput stringinput.Model
	// moveParam stores move param for Model.
	moveParam moveparam.Model
	// renameInput stores rename input for Model.
	renameInput renameinput.Model
	// dialogQueue stores dialog queue for Model.
	dialogQueue []pendingDialog
	// duplicate stores duplicate for Model.
	duplicate *duplicateSession
	// newParameter stores new parameter for Model.
	newParameter *newParameterSession
	// pendingDetails stores pending details for Model.
	pendingDetails *pendingDetailsSelection
	// valueEditSource stores value edit source for Model.
	valueEditSource panels.ID

	// width stores width for Model.
	width int
	// height stores height for Model.
	height int
}

// pendingDetailsSelection holds pending details selection state used by the app package.
type pendingDetailsSelection struct {
	// data stores data for pendingDetailsSelection.
	data *messages.ParameterViewData
	// activate stores activate for pendingDetailsSelection.
	activate bool
}

// duplicateSession holds duplicate session state used by the app package.
type duplicateSession struct {
	// project stores project for duplicateSession.
	project core.Project
	// groupKey stores group key for duplicateSession.
	groupKey string
	// sourceParamKey stores source param key for duplicateSession.
	sourceParamKey string
	// visibleName stores visible name for duplicateSession.
	visibleName string
}

// newParameterSession holds new parameter session state used by the app package.
type newParameterSession struct {
	// projectID stores project id for newParameterSession.
	projectID string
	// groupKey stores group key for newParameterSession.
	groupKey string
}

// New constructs new and returns the resulting value or error.
func New(svc *core.Core) Model {
	m := Model{
		svc:         svc,
		projects:    projects.New(svc),
		parameters:  parameters.New(svc),
		dialog:      dialogcmp.New(),
		jsonInput:   jsoninput.New(),
		boolPicker:  boolpicker.New(),
		numberInput: numberinput.New(),
		stringInput: stringinput.New(),
		moveParam:   moveparam.New(),
		renameInput: renameinput.New(),
		details:     details.New(),
		logs:        logs.New(svc),
		logsHeight:  defaultLogsPanelHeight,
		help:        newHelpModel(),
		active:      panels.Projects,
		prevTop:     panels.Projects,
	}

	m.projects = m.projects.SetActive(true)
	return m
}

// Init initializes init for Model and returns the resulting state or error.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.projects.Init(),
		m.parameters.Init(),
		m.details.Init(),
		m.logs.Init(),
	)
}
