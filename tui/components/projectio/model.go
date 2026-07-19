package projectio

import (
	"maps"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/rc/importer"
	"github.com/yumauri/fbrcm/tui/components/inputstyles"
	moveparam "github.com/yumauri/fbrcm/tui/components/moveparam"
)

type Mode int

const (
	ModeNone Mode = iota
	ModeImport
	ModeExport
)

type phase int

const (
	phaseClosed phase = iota
	phaseImportFile
	phaseImportOptions
	phaseImportConflicts
	phaseImportWorking
	phaseExportSource
	phaseExportPath
)

type ImportPlanRequestedMsg struct {
	Project core.Project
	Raw     []byte
	Path    string
	Options core.ProjectImportOptions
}

type ExportRequestedMsg struct {
	Project core.Project
	Path    string
	Draft   bool
}

type optionSelectorKind int

const (
	optionSelectorNone optionSelectorKind = iota
	optionSelectorStrategy
	optionSelectorConditions
)

type Model struct {
	x, y, width, height int
	mode                Mode
	phase               phase
	project             core.Project
	picker              filepicker.Model
	pathInput           textinput.Model
	optionInputs        [4]textinput.Model
	optionCursor        int
	buttonCursor        int
	buttonsFocused      bool
	optionSelector      moveparam.Model
	optionSelectorKind  optionSelectorKind
	strategy            core.ProjectImportStrategy
	conditionPolicy     core.ProjectConditionPolicy
	exportDraft         bool
	sourceRaw           []byte
	sourcePath          string
	summary             importer.Summary
	conflicts           []core.ProjectImportConflict
	resolutions         map[string]core.ProjectImportResolution
	conflictCursor      int
	reviewedConflicts   bool
	workingFrom         phase
	errorText           string
}

func New() Model {
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
	pathInput := inputstyles.NewTextInput()
	pathInput.Placeholder = "destination path"
	inputs := [4]textinput.Model{}
	for i := range inputs {
		inputs[i] = inputstyles.NewTextInput()
	}
	inputs[0].Placeholder = "all groups"
	inputs[1].Placeholder = "no parameter filters"
	inputs[2].Placeholder = "no rich search"
	inputs[3].Placeholder = "no expression"
	return Model{
		picker:          picker,
		pathInput:       pathInput,
		optionInputs:    inputs,
		optionSelector:  moveparam.New(),
		strategy:        core.ProjectImportMerge,
		conditionPolicy: core.ProjectImportKeepConditions,
	}
}

func (m Model) SetBounds(x, y, width, height int) Model {
	m.x, m.y, m.width, m.height = x, y, width, height
	m.picker.SetHeight(min(max(height-14, 5), 18))
	inputWidth := min(max(width-24, 24), 64)
	m.pathInput.SetWidth(inputWidth)
	for i := range m.optionInputs {
		m.optionInputs[i].SetWidth(inputWidth)
	}
	if m.optionSelector.IsOpen() {
		m.openCurrentOptionSelector()
	}
	return m
}

func (m Model) OpenImport(project core.Project) (Model, tea.Cmd) {
	m = m.reset()
	m.mode, m.phase, m.project = ModeImport, phaseImportFile, project
	return m, tea.Batch(m.picker.Init(), tea.ClearScreen)
}

func (m Model) OpenExport(project core.Project, hasDraft bool) (Model, tea.Cmd) {
	m = m.reset()
	m.mode, m.project = ModeExport, project
	m.phase = phaseExportPath
	if hasDraft {
		m.phase = phaseExportSource
	}
	m.pathInput.SetValue(filepath.Join(".", project.ProjectID+"-remote-config.json"))
	m.pathInput.CursorEnd()
	if m.phase == phaseExportPath {
		return m, m.pathInput.Focus()
	}
	return m, nil
}

func (m Model) OpenExportPath(project core.Project, path string, draft bool) (Model, tea.Cmd) {
	m = m.reset()
	m.mode, m.phase, m.project = ModeExport, phaseExportPath, project
	m.exportDraft = draft
	m.pathInput.SetValue(path)
	m.pathInput.CursorEnd()
	return m, m.pathInput.Focus()
}

func (m Model) reset() Model {
	m.mode, m.phase = ModeNone, phaseClosed
	m.project = core.Project{}
	m.sourceRaw, m.sourcePath = nil, ""
	m.summary = importer.Summary{}
	m.conflicts = nil
	m.resolutions = nil
	m.conflictCursor = 0
	m.reviewedConflicts = false
	m.workingFrom = phaseClosed
	m.optionCursor = 0
	m.buttonCursor = 0
	m.buttonsFocused = false
	m.optionSelector = m.optionSelector.Close()
	m.optionSelectorKind = optionSelectorNone
	m.exportDraft = false
	m.strategy = core.ProjectImportMerge
	m.conditionPolicy = core.ProjectImportKeepConditions
	m.errorText = ""
	m.pathInput.Blur()
	for i := range m.optionInputs {
		m.optionInputs[i].SetValue("")
		m.optionInputs[i].Blur()
	}
	return m
}

func (m Model) Close() Model { return m.reset() }
func (m Model) IsOpen() bool { return m.phase != phaseClosed }
func (m Model) Mode() Mode   { return m.mode }

func (m Model) Position() (int, int) {
	view := m.View()
	return max(m.x+(m.width-viewWidth(view))/2, m.x), max(m.y+(m.height-viewHeight(view))/2, m.y)
}

func (m Model) ConflictsOpen() bool { return m.reviewedConflicts }

func (m Model) OpenConflicts(conflicts []core.ProjectImportConflict) Model {
	m.phase = phaseImportConflicts
	m.conflicts = append([]core.ProjectImportConflict(nil), conflicts...)
	m.resolutions = make(map[string]core.ProjectImportResolution, len(conflicts))
	for _, conflict := range conflicts {
		m.resolutions[conflict.ID] = core.ProjectImportKeepCurrent
	}
	m.conflictCursor = 0
	m.reviewedConflicts = false
	return m
}

func (m Model) SetError(err error) Model {
	if err == nil {
		m.errorText = ""
	} else {
		m.errorText = err.Error()
		if m.phase == phaseImportWorking && m.workingFrom != phaseClosed {
			m.phase = m.workingFrom
		}
	}
	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.IsOpen() {
		return m, nil
	}
	if m.optionSelector.IsOpen() {
		return m.updateOptionSelector(msg)
	}
	if next, cmd, handled := m.updateActionButtonsMouse(msg); handled {
		return next, cmd
	}
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
		return m.Close(), nil
	}
	switch m.phase {
	case phaseImportFile:
		return m.updateImportFile(msg)
	case phaseImportOptions:
		return m.updateImportOptions(msg)
	case phaseImportConflicts:
		return m.updateImportConflicts(msg)
	case phaseImportWorking:
		return m, nil
	case phaseExportSource:
		return m.updateExportSource(msg)
	case phaseExportPath:
		return m.updateExportPath(msg)
	default:
		return m, nil
	}
}

func (m Model) updateImportFile(msg tea.Msg) (Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if m.buttonsFocused {
			return m.updateFocusedActionButtons(key.String())
		}
		if key.String() == "tab" {
			m.focusActionButtons()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	if selected, path := m.picker.DidSelectFile(msg); selected {
		raw, err := os.ReadFile(path)
		if err != nil {
			m.errorText = err.Error()
			return m, cmd
		}
		parsed, err := importer.ParseSource(raw)
		if err != nil {
			m.errorText = err.Error()
			return m, cmd
		}
		m.sourceRaw, m.sourcePath = raw, path
		m.summary = importer.Summarize(parsed.Config, parsed.WrappedCache)
		m.phase, m.optionCursor, m.errorText = phaseImportOptions, 0, ""
		m.buttonCursor, m.buttonsFocused = 0, false
		m.focusOptionInput()
		return m, nil
	}
	return m, cmd
}

func (m Model) updateImportOptions(msg tea.Msg) (Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if m.buttonsFocused {
			return m.updateFocusedActionButtons(key.String())
		}
		switch key.String() {
		case "up", "ctrl+k":
			m.optionCursor = (m.optionCursor + optionRowCount - 1) % optionRowCount
			m.focusOptionInput()
			return m, nil
		case "down", "ctrl+j":
			if m.optionCursor == optionRowCount-1 {
				m.focusActionButtons()
				return m, nil
			}
			m.optionCursor++
			m.focusOptionInput()
			return m, nil
		case "tab":
			if m.optionCursor == optionRowCount-1 {
				m.focusActionButtons()
				return m, nil
			}
			m.optionCursor++
			m.focusOptionInput()
			return m, nil
		case "shift+tab":
			if m.optionCursor == 0 {
				m.focusActionButtons()
				m.buttonCursor = 1
				return m, nil
			}
			m.optionCursor--
			m.focusOptionInput()
			return m, nil
		case "space":
			if m.optionCursor == optionStrategy || m.optionCursor == optionConditions {
				m.openCurrentOptionSelector()
				return m, nil
			}
		case "right":
			if m.optionCursor == optionStrategy || m.optionCursor == optionConditions {
				m.openCurrentOptionSelector()
				return m, nil
			}
		case "enter":
			if m.optionCursor == optionStrategy || m.optionCursor == optionConditions {
				m.openCurrentOptionSelector()
				return m, nil
			}
		}
	}
	index, ok := m.currentInputIndex()
	if !ok {
		return m, nil
	}
	var cmd tea.Cmd
	m.optionInputs[index], cmd = m.optionInputs[index].Update(msg)
	return m, cmd
}

func (m Model) updateImportConflicts(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || len(m.conflicts) == 0 {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		m.conflictCursor = (m.conflictCursor + len(m.conflicts) - 1) % len(m.conflicts)
	case "down", "j":
		m.conflictCursor = (m.conflictCursor + 1) % len(m.conflicts)
	case "left", "h":
		m.resolutions[m.conflicts[m.conflictCursor].ID] = core.ProjectImportKeepCurrent
	case "right", "l", "space":
		m.resolutions[m.conflicts[m.conflictCursor].ID] = core.ProjectImportUseImported
	case "C":
		for _, conflict := range m.conflicts {
			m.resolutions[conflict.ID] = core.ProjectImportKeepCurrent
		}
	case "I":
		for _, conflict := range m.conflicts {
			m.resolutions[conflict.ID] = core.ProjectImportUseImported
		}
	case "enter":
		cmd := m.importPlanCmd()
		m.reviewedConflicts = true
		m.workingFrom, m.phase = phaseImportConflicts, phaseImportWorking
		return m, cmd
	}
	return m, nil
}

func (m Model) updateExportSource(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.buttonsFocused {
		return m.updateFocusedActionButtons(key.String())
	}
	switch key.String() {
	case "up", "down", "left", "right", "space":
		m.exportDraft = !m.exportDraft
	case "tab":
		m.focusActionButtons()
	case "enter":
		return m.activateActionButton(0)
	}
	return m, nil
}

func (m Model) updateExportPath(msg tea.Msg) (Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if m.buttonsFocused {
			return m.updateFocusedActionButtons(key.String())
		}
		switch key.String() {
		case "tab":
			m.focusActionButtons()
			return m, nil
		case "enter":
			return m.submitExportPath()
		}
	}
	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	return m, cmd
}

func (m Model) submitExportPath() (Model, tea.Cmd) {
	path := strings.TrimSpace(m.pathInput.Value())
	if path == "" {
		m.errorText = "destination path is empty"
		return m, nil
	}
	return m, func() tea.Msg { return ExportRequestedMsg{Project: m.project, Path: path, Draft: m.exportDraft} }
}

const (
	optionStrategy = iota
	optionGroups
	optionFilters
	optionSearch
	optionExpr
	optionConditions
	optionRowCount
)

func (m *Model) focusOptionInput() {
	for i := range m.optionInputs {
		m.optionInputs[i].Blur()
	}
	if index, ok := m.currentInputIndex(); ok {
		m.optionInputs[index].Focus()
	}
}

func (m Model) currentInputIndex() (int, bool) {
	switch m.optionCursor {
	case optionGroups:
		return 0, true
	case optionFilters:
		return 1, true
	case optionSearch:
		return 2, true
	case optionExpr:
		return 3, true
	default:
		return 0, false
	}
}

func (m Model) importOptions() core.ProjectImportOptions {
	return core.ProjectImportOptions{
		Groups:            splitList(m.optionInputs[0].Value()),
		Filters:           splitList(m.optionInputs[1].Value()),
		Search:            strings.TrimSpace(m.optionInputs[2].Value()),
		Expr:              strings.TrimSpace(m.optionInputs[3].Value()),
		Strategy:          m.strategy,
		ConditionPolicy:   m.conditionPolicy,
		DefaultResolution: core.ProjectImportKeepCurrent,
		Resolutions:       cloneResolutions(m.resolutions),
	}
}

func (m Model) importPlanCmd() tea.Cmd {
	project, raw, path, options := m.project, append([]byte(nil), m.sourceRaw...), m.sourcePath, m.importOptions()
	return func() tea.Msg {
		return ImportPlanRequestedMsg{Project: project, Raw: raw, Path: path, Options: options}
	}
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func cloneResolutions(in map[string]core.ProjectImportResolution) map[string]core.ProjectImportResolution {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]core.ProjectImportResolution, len(in))
	maps.Copy(out, in)
	return out
}
