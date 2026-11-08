package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/rc/evaluate"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/tui/components/contextpanel"
	dialogcmp "github.com/yumauri/fbrcm/tui/components/dialog"
)

const (
	contextUserProperty = "app.userProperty"
	contextCustomSignal = "app.customSignal"
	contextAnotherName  = "another_name"
)

var contextFieldOptions = []contextpanel.Row{
	{Key: "device.os", Label: "Platform", Value: "device.os · web, android, ios"},
	{Key: "device.country", Label: "Country", Value: "device.country"},
	{Key: "device.language", Label: "Language", Value: "device.language"},
	{Key: "app.id", Label: "Firebase app ID", Value: "app.id"},
	{Key: "app.version", Label: "App version", Value: "app.version"},
	{Key: "app.build", Label: "App build", Value: "app.build"},
	{Key: contextUserProperty, Label: "User property", Value: "Choose or enter a name"},
	{Key: contextCustomSignal, Label: "Custom signal", Value: "Choose or enter a name"},
	{Key: contextScopeKey, Label: "Template scope", Value: "Project ID or server@project"},
	{Key: "other", Label: "Other field…", Value: "Enter a supported Firebase input"},
}

type contextNameMsg struct {
	field, name string
}

func (m *Model) openContextFields() {
	c := &m.requestContext
	c.pickerBack, c.overlay = c.overlay, contextFields
	c.picker = contextpanel.Model{}.SetRows(contextFieldOptions, "")
}

func (m *Model) openContextNames(field string) {
	names := make(map[string]struct{})
	for _, expression := range m.parameters.ReferencedConditionExpressions() {
		for _, name := range evaluate.IndexedInputNames(expression, field) {
			names[name] = struct{}{}
		}
	}
	rows := make([]contextpanel.Row, 0, len(names)+1)
	for _, name := range strfold.SortedKeys(names) {
		key := evaluate.IndexedInputKey(field, name)
		rows = append(rows, contextpanel.Row{Key: key, Label: name, Value: key})
	}
	rows = append(rows, contextpanel.Row{Key: contextAnotherName, Label: "Enter another name…", Value: "Type a name"})
	c := &m.requestContext
	c.overlay, c.nameField = contextNames, field
	c.picker = contextpanel.Model{}.SetRows(rows, "")
}

func contextNameTitle(field string) string {
	if field == contextUserProperty {
		return "User Property Name"
	}
	return "Custom Signal Name"
}

func (m *Model) openContextName(field, name, problem string) {
	var body []string
	if problem != "" {
		body = append(body, problem)
	}
	m.dialog = m.dialog.Open(dialogcmp.Config{
		Title: contextNameTitle(field),
		Body:  body,
		Input: &dialogcmp.Input{Label: "Name", Value: name, SubmitOnEnter: true},
		Buttons: []dialogcmp.Button{
			{Label: "Continue", Variant: dialogcmp.ButtonVariantAccent, OnPressWithInput: func(name string) tea.Cmd {
				return func() tea.Msg { return contextNameMsg{field: field, name: name} }
			}},
			{Label: "Cancel"},
		},
	})
}

func (m *Model) applyContextName(msg contextNameMsg) {
	if strings.TrimSpace(msg.name) == "" {
		m.openContextName(msg.field, msg.name, "Enter a name.")
		return
	}
	m.openNamedContextFact(evaluate.IndexedInputKey(msg.field, msg.name))
}

func (m *Model) openNamedContextFact(key string) {
	if fact, ok := m.requestContext.facts[key]; ok {
		m.openContextFact(key, contextFactAssignment(key, contextFactAnswer(fact)), "")
		return
	}
	m.openContextFact("", contextFactAssignment(key, ""), "")
}
