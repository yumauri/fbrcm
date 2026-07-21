package app

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/components/authpicker"
	boolpicker "github.com/yumauri/fbrcm/tui/components/boolpicker"
	"github.com/yumauri/fbrcm/tui/components/conditions"
	"github.com/yumauri/fbrcm/tui/components/details"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	jsoninput "github.com/yumauri/fbrcm/tui/components/jsoninput"
	"github.com/yumauri/fbrcm/tui/components/logs"
	moveparam "github.com/yumauri/fbrcm/tui/components/moveparam"
	numberinput "github.com/yumauri/fbrcm/tui/components/numberinput"
	"github.com/yumauri/fbrcm/tui/components/parameters"
	"github.com/yumauri/fbrcm/tui/components/projectio"
	"github.com/yumauri/fbrcm/tui/components/projects"
	renameinput "github.com/yumauri/fbrcm/tui/components/renameinput"
	"github.com/yumauri/fbrcm/tui/components/setup"
	stringinput "github.com/yumauri/fbrcm/tui/components/stringinput"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

type Model struct {
	svc *core.Core

	projects        projects.Model
	parameters      parameters.Model
	conditions      conditions.Model
	details         details.Model
	logs            logs.Model
	projectsMode    projectsPanelMode
	logsHeight      int
	logsSized       bool
	logsMode        logsPanelMode
	logsSaved       int
	help            help.Model
	helpPalette     helpPaletteModel
	setup           setup.Model
	active          panels.ID
	parametersTab   panels.ID
	prevTop         panels.ID
	capture         panels.ID
	detailsVisible  bool
	dialog          dialogcmp.Model
	jsonInput       jsoninput.Model
	boolPicker      boolpicker.Model
	numberInput     numberinput.Model
	stringInput     stringinput.Model
	moveParam       moveparam.Model
	authPicker      authpicker.Model
	renameInput     renameinput.Model
	projectIO       projectio.Model
	dialogQueue     []pendingDialog
	duplicate       *duplicateSession
	newParameter    *newParameterSession
	pendingDetails  *pendingDetailsSelection
	historyRollback *historyRollbackSession
	conditionEdit   *conditionEditSession
	conditionalAdd  *conditionalValueAddSession
	authBind        *authBindingSession
	profileRename   *profileRenameSession
	projectImport   *core.ProjectImportPlan
	projectExport   *projectExportSession
	projectDefaults *projectDefaultsSession
	valueEditSource panels.ID
	authCount       int

	width  int
	height int
}

type pendingDetailsSelection struct {
	data          *messages.ParameterViewData
	groupData     *messages.GroupViewData
	conditionData *messages.ConditionViewData
	activate      bool
}

type duplicateSession struct {
	project        core.Project
	groupKey       string
	sourceParamKey string
	visibleName    string
}

type newParameterSession struct {
	projectID string
	groupKey  string
}

type conditionEditMode int

const (
	conditionAddName conditionEditMode = iota + 1
	conditionRename
	conditionExpression
	conditionColor
	conditionMove
	conditionDetailsExpression
)

type conditionEditSession struct {
	mode         conditionEditMode
	project      core.Project
	originalName string
	name         string
	creating     bool
	currentColor string
}

type conditionalValueAddSession struct {
	condition string
}

type profileRenameSession struct {
	profile string
}

func New(svc *core.Core) Model {
	authCount := 0
	if svc != nil {
		if entries, _, err := svc.ListAuth(); err == nil {
			authCount = len(entries)
		}
	}
	m := Model{
		svc:           svc,
		projects:      projects.New(svc),
		parameters:    parameters.New(svc),
		conditions:    conditions.New(svc),
		dialog:        dialogcmp.New(),
		jsonInput:     jsoninput.New(),
		boolPicker:    boolpicker.New(),
		numberInput:   numberinput.New(),
		stringInput:   stringinput.New(),
		moveParam:     moveparam.New(),
		authPicker:    authpicker.New(),
		renameInput:   renameinput.New(),
		projectIO:     projectio.New(),
		details:       details.New(),
		logs:          logs.New(svc),
		logsHeight:    defaultLogsPanelHeight,
		help:          newHelpModel(),
		helpPalette:   newHelpPaletteModel(),
		setup:         setup.New(svc),
		authCount:     authCount,
		active:        panels.Projects,
		parametersTab: panels.Parameters,
		prevTop:       panels.Projects,
	}

	m.projects = m.projects.SetActive(true)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.setup.Init(),
		m.parameters.Init(),
		m.conditions.Init(),
		m.details.Init(),
		m.logs.Init(),
	)
}
