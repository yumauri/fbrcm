package app

import (
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/rc/evaluate"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/tui/components/contextpanel"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
	tuiconfig "github.com/yumauri/fbrcm/tui/config"
	"github.com/yumauri/fbrcm/tui/contextview"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

type contextOverlay int

const (
	contextInline contextOverlay = iota
	contextOverview
	contextFields
	contextNames
	contextRequests
)

func (o contextOverlay) isPicker() bool {
	return o == contextFields || o == contextNames || o == contextRequests
}

const (
	contextScopeKey     = "template_scope"
	manualContextSource = "manual context"
)

type requestContextState struct {
	facts          map[string]evaluate.Fact
	capture        *evaluate.Exchange
	captureSource  string
	scope          rctarget.Target
	scenario       *contextview.Scenario
	list           contextpanel.Model
	picker         contextpanel.Model
	nameField      string
	exchanges      []evaluate.Exchange
	exchangeSource string
	overlay        contextOverlay
	pickerBack     contextOverlay
	reading        *contextCaptureLoad
	pasting        bool
	paused         bool
	showExcluded   bool
}

type contextFactMsg struct {
	original, assignment string
	remove               bool
}

func (m *Model) refreshContext(selected string) {
	c := &m.requestContext
	if len(c.facts) == 0 && c.scope.ProjectID == "" {
		c.capture, c.captureSource, c.paused = nil, "", false
	}
	c.scenario = nil
	if !c.paused {
		c.scenario = contextview.New(c.facts, c.capture, c.scope, c.showExcluded)
	}
	rows := make([]contextpanel.Row, 0, len(c.facts)+1)
	if c.scope.ProjectID != "" {
		rows = append(rows, contextpanel.Row{Key: contextScopeKey, Label: "Template scope", Value: c.scope.String()})
	}
	for _, key := range slices.Sorted(maps.Keys(c.facts)) {
		rows = append(rows, contextpanel.Row{Key: key, Label: contextFieldLabel(key), Value: contextFactValue(c.facts[key])})
	}
	c.list = c.list.SetRows(rows, selected)
	m.parameters = m.parameters.SetScenario(c.scenario)
	m.conditions = m.conditions.SetScenario(c.scenario)
	m.details = m.details.SetScenario(c.scenario)
	m.applyLayout()
}

func contextFieldLabel(key string) string {
	for _, prefix := range []struct{ field, label string }{
		{"app.customSignal", "Signal"},
		{"app.userProperty", "Property"},
		{"app.audiences", "Audience"},
	} {
		if indexed, ok := strings.CutPrefix(key, prefix.field+"["); ok {
			return prefix.label + " · " + strings.Trim(strings.TrimSuffix(indexed, "]"), "'")
		}
	}
	return key
}

func contextFactValue(fact evaluate.Fact) string {
	switch {
	case !fact.Known:
		return "(unknown)"
	case fact.Absent:
		return "(unset)"
	case fact.Value == "":
		return "(empty string)"
	default:
		return fact.Value
	}
}

func contextFactAnswer(fact evaluate.Fact) string {
	switch {
	case !fact.Known:
		return "/unknown"
	case fact.Absent:
		return "/unset"
	default:
		quoted, _ := json.Marshal(fact.Value)
		return string(quoted)
	}
}

func (m Model) updateContextPopup(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateWindowSize(msg)
		return m, nil, true
	case tea.KeyMsg:
		key := msg.String()
		if tuiconfig.Matches(tuiconfig.BlockGlobal, tuiconfig.ActionOpenContext, key) && m.requestContext.overlay == contextOverview {
			m.closeContextPopup()
			return m, nil, true
		}
		m, cmd, _ := m.updateContextKey(key, true)
		return m, cmd, true
	case tea.MouseMsg:
		if click, ok := msg.(tea.MouseClickMsg); ok && click.Mouse().Button == tea.MouseLeft &&
			!m.contextLayout(true).Contains(click.Mouse().X, click.Mouse().Y) {
			m.backContextPopup()
			return m, nil, true
		}
		m.updateContextMouse(msg, true)
		return m, nil, true
	}
	return m, nil, false
}

func (m *Model) contextList() *contextpanel.Model {
	if m.requestContext.overlay.isPicker() {
		return &m.requestContext.picker
	}
	return &m.requestContext.list
}

func (m Model) updateContextKey(key string, popup bool) (Model, tea.Cmd, bool) {
	matches := func(action tuiconfig.Action) bool {
		return tuiconfig.Matches(tuiconfig.BlockContext, action, key)
	}
	switch {
	case matches(tuiconfig.ActionCancel):
		if popup {
			m.backContextPopup()
		} else {
			m.setActive(m.selectedParametersTab())
		}
	case matches(tuiconfig.ActionSubmit):
		m.selectContextRow()
	case matches(tuiconfig.ActionUp):
		m.contextList().Move(-1, m.contextLayout(popup))
	case matches(tuiconfig.ActionDown):
		m.contextList().Move(1, m.contextLayout(popup))
	case matches(tuiconfig.ActionHome), matches(tuiconfig.ActionEnd):
		m.contextList().Jump(matches(tuiconfig.ActionEnd), m.contextLayout(popup))
	case m.requestContext.overlay.isPicker():
		return m, nil, false
	case matches(tuiconfig.ActionNew):
		m.openContextFields()
	case matches(tuiconfig.ActionImport):
		m.openContextImportDialog()
	case matches(tuiconfig.ActionDelete):
		m.applyContextFact(contextFactMsg{original: m.requestContext.list.Current().Key, remove: true})
	case matches(tuiconfig.ActionClear):
		m.requestContext.facts = nil
		m.requestContext.scope = rctarget.Target{}
		m.refreshContext("")
	case matches(tuiconfig.ActionToggleContext):
		if m.requestContext.list.Len() > 0 {
			m.requestContext.paused = !m.requestContext.paused
			m.refreshContext("")
		}
	case matches(tuiconfig.ActionShowHidden):
		m.requestContext.showExcluded = !m.requestContext.showExcluded
		m.refreshContext("")
	case !popup && matches(tuiconfig.ActionToggleMode):
		m.toggleProjectsMode()
	default:
		return m, nil, false
	}
	return m, nil, true
}

func (m *Model) updateContextMouse(msg tea.MouseMsg, popup bool) {
	if !popup && m.projectsMode == projectsPanelModeCollapsed {
		return
	}
	layout := m.contextLayout(popup)
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		mouse := msg.Mouse()
		if mouse.Button == tea.MouseLeft && m.contextList().Click(layout, mouse.X, mouse.Y, time.Now()) {
			m.selectContextRow()
		}
	case tea.MouseWheelMsg:
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			m.contextList().Move(-1, layout)
		case tea.MouseWheelDown:
			m.contextList().Move(1, layout)
		}
	}
}

func (m *Model) closeContextPopup() {
	c := &m.requestContext
	c.overlay = c.pickerBack
	c.pickerBack = contextInline
	c.nameField = ""
	c.exchanges = nil
	c.exchangeSource = ""
}

func (m *Model) backContextPopup() {
	c := &m.requestContext
	if c.overlay == contextNames {
		c.overlay = contextFields
		c.picker = contextpanel.Model{}.SetRows(contextFieldOptions, c.nameField)
		c.nameField = ""
		return
	}
	m.closeContextPopup()
}

func (m *Model) selectContextRow() {
	c := &m.requestContext
	switch c.overlay {
	case contextFields:
		field := c.picker.Current().Key
		if field == contextUserProperty || field == contextCustomSignal {
			m.openContextNames(field)
			return
		}
		m.closeContextPopup()
		if field == contextScopeKey {
			m.openContextFact(contextScopeKey, c.scope.String(), "")
			return
		}
		assignment := field + "="
		if field == "other" {
			assignment = ""
		}
		m.openContextFact("", assignment, "")
	case contextNames:
		field, key := c.nameField, c.picker.Current().Key
		m.closeContextPopup()
		if key == contextAnotherName {
			m.openContextName(field, "", "")
			return
		}
		m.openNamedContextFact(key)
	case contextRequests:
		index, err := strconv.Atoi(c.picker.Current().Key)
		if err == nil && index >= 0 && index < len(c.exchanges) {
			exchange, source := c.exchanges[index], c.exchangeSource
			m.closeContextPopup()
			m.useContextCapture(exchange, source)
		}
	default:
		switch key := c.list.Current().Key; key {
		case contextScopeKey:
			m.openContextFact(key, c.scope.String(), "")
		case "":
			m.openContextFields()
		default:
			m.openContextFact(key, contextFactAssignment(key, contextFactAnswer(c.facts[key])), "")
		}
	}
}

func contextFactAssignment(key, value string) string {
	// Escape any equals sign within a quoted name so the assignment separator
	// is the first literal equals sign.
	return strings.ReplaceAll(key, "=", `\u003d`) + "=" + value
}

func (m *Model) openContextFact(original, assignment, problem string) {
	title := "Add Context Input"
	if original != "" {
		title = "Edit Context Input"
	}
	body := []string{"Set a request input such as device.os=web.", "Use /unknown or /unset for missing inputs."}
	label, placeholder := "Field=value", "device.os=web"
	canRemove := original != ""
	if original == contextScopeKey {
		title = "Context Template Scope"
		body = []string{
			"Enter a project ID or number for its client template.",
			"Use server@project for its server template.",
		}
		label, placeholder = "Template target", "server@demo"
		canRemove = m.requestContext.scope.ProjectID != ""
	}
	if problem != "" {
		body = append([]string{problem, ""}, body...)
	}
	buttons := []dialogcmp.Button{{
		Label: "Apply", Variant: dialogcmp.ButtonVariantAccent,
		OnPressWithInput: func(value string) tea.Cmd {
			return func() tea.Msg { return contextFactMsg{original: original, assignment: value} }
		},
	}}
	if canRemove {
		buttons = append(buttons, dialogcmp.Button{Label: "Remove", OnPress: func() tea.Msg {
			return contextFactMsg{original: original, remove: true}
		}})
	}
	buttons = append(buttons, dialogcmp.Button{Label: "Cancel"})
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: title, Body: body, Buttons: buttons,
		Input: &dialogcmp.Input{Label: label, Value: assignment, Placeholder: placeholder, SubmitOnEnter: true},
	})
}

func (m *Model) applyContextFact(msg contextFactMsg) {
	c := &m.requestContext
	if msg.original == contextScopeKey {
		var scope rctarget.Target
		if !msg.remove {
			var err error
			scope, err = rctarget.Parse(msg.assignment)
			if err != nil {
				m.openContextFact(contextScopeKey, msg.assignment, err.Error())
				return
			}
		}
		c.scope = scope
		m.refreshContext(contextScopeKey)
		return
	}
	if msg.remove {
		if msg.original == "" {
			return
		}
		c.facts = maps.Clone(c.facts)
		delete(c.facts, msg.original)
		m.refreshContext("")
		return
	}
	need, value, ok := evaluate.SplitInput(msg.assignment)
	if !ok {
		m.openContextFact(msg.original, msg.assignment, "Enter a supported Firebase field=value assignment.")
		return
	}
	fact, err := evaluate.Answer(need, value, manualContextSource)
	if err != nil {
		m.openContextFact(msg.original, msg.assignment, err.Error())
		return
	}
	facts := maps.Clone(c.facts)
	if facts == nil {
		facts = make(map[string]evaluate.Fact)
	}
	delete(facts, msg.original)
	facts[need.Key] = fact
	c.facts = facts
	m.refreshContext(need.Key)
}

func (m Model) contextExplanationData() *messages.ParameterViewData {
	var data *messages.ParameterViewData
	switch m.active {
	case panels.Parameters:
		data, _ = m.parameters.CurrentParameterViewData()
	case panels.Details:
		data = m.details.Data()
	}
	if data == nil || data.Parameter.Key == "" || !m.requestContext.scenario.AppliesTo(data.Project) {
		return nil
	}
	return data
}
