package app

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	boolpicker "github.com/yumauri/fbrcm/tui/components/boolpicker"
	"github.com/yumauri/fbrcm/tui/components/conditions"
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
	renameInput     renameinput.Model
	dialogQueue     []pendingDialog
	duplicate       *duplicateSession
	newParameter    *newParameterSession
	pendingDetails  *pendingDetailsSelection
	historyRollback *historyRollbackSession
	conditionEdit   *conditionEditSession
	conditionalAdd  *conditionalValueAddSession
	valueEditSource panels.ID

	width  int
	height int
}

type pendingDetailsSelection struct {
	data          *messages.ParameterViewData
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

func New(svc *core.Core) Model {
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
		renameInput:   renameinput.New(),
		details:       details.New(),
		logs:          logs.New(svc),
		logsHeight:    defaultLogsPanelHeight,
		help:          newHelpModel(),
		active:        panels.Projects,
		parametersTab: panels.Parameters,
		prevTop:       panels.Projects,
	}

	m.projects = m.projects.SetActive(true)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.projects.Init(),
		m.parameters.Init(),
		m.conditions.Init(),
		m.details.Init(),
		m.logs.Init(),
	)
}
