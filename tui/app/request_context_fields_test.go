package app

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/parameters"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestContextNamePickersUseSelectedTemplates(t *testing.T) {
	m := viewTestModel(100, 36, panels.Context)
	client := core.Project{ProjectID: "demo"}
	server := core.Project{ProjectID: "server@demo"}
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{client, server}})
	clientConfig := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "client", Expression: `device.os == 'ios' && app.userProperty["environment"].exactlyMatches(['prod']) && app.customSignal['environment'].exactlyMatches(['prod'])`},
			{Name: "unused", Expression: `app.userProperty['unused'].exactlyMatches(['prod'])`},
		},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {ConditionalValues: map[string]firebase.RemoteConfigValue{"client": {Value: "yes"}}},
		},
	}
	serverConfig := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{
			{Name: "server", Expression: `app.customSignal['environment'].exactlyMatches(['prod']) && app.customSignal['server_only'].exactlyMatches(['prod'])`},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"group": {Parameters: map[string]firebase.RemoteConfigParam{
				"flag": {ConditionalValues: map[string]firebase.RemoteConfigValue{"server": {Value: "yes"}}},
			}},
		},
	}
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project: client, Tree: parameters.BuildTree(clientConfig, time.Time{}, ""), Source: "cache",
	})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project: server, Tree: parameters.BuildTree(serverConfig, time.Time{}, ""), Source: "draft", HasDraft: true,
	})
	m.applyContextFact(contextFactMsg{assignment: "device.os=web"})

	next, _ := m.Update(keyPress('a'))
	m = next.(Model)
	layout := m.contextLayout(true)
	property := tea.MouseClickMsg{X: layout.X + 2, Y: layout.Y + 13, Button: tea.MouseLeft}
	next, _ = m.Update(property)
	if got := next.(Model).requestContext.picker.Current().Key; got != contextUserProperty {
		t.Fatalf("click selected %q, want User property", got)
	}
	next, _ = next.Update(property)
	m = next.(Model)
	if m.requestContext.overlay != contextNames || m.requestContext.picker.Len() != 2 {
		t.Fatal("User property should suggest only the referenced environment name and a manual choice")
	}
	enter := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	next, _ = m.Update(enter)
	if got := next.(Model).dialog.InputValue(); got != "app.userProperty['environment']=" {
		t.Fatalf("User property assignment = %q", got)
	}
	next, _ = next.Update(tea.PasteMsg{Content: "qa"})
	next, apply := next.Update(enter)
	next, _ = next.Update(apply())

	next, _ = next.Update(keyPress('a'))
	m = next.(Model)
	layout = m.contextLayout(true)
	signal := tea.MouseClickMsg{X: layout.X + 2, Y: layout.Y + 16, Button: tea.MouseLeft}
	next, _ = m.Update(signal)
	next, _ = next.Update(signal)
	m = next.(Model)
	if m.requestContext.overlay != contextNames || m.requestContext.picker.Len() != 3 ||
		!strings.Contains(ansi.Strip(m.View().Content), "server_only") {
		t.Fatal("Custom signal should merge and deduplicate names from both selected templates")
	}
	esc := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	next, _ = m.Update(esc)
	if got := next.(Model).requestContext.picker.Current().Key; got != contextCustomSignal {
		t.Fatalf("back from names selected %q, want Custom signal", got)
	}
	next, _ = next.Update(enter)
	m = next.(Model)
	layout = m.contextLayout(true)
	click := tea.MouseClickMsg{X: layout.X + 2, Y: layout.Y + 2, Button: tea.MouseLeft}
	next, _ = m.Update(click)
	if next.(Model).dialog.IsOpen() {
		t.Fatal("single click should select without opening the editor")
	}
	click.Y--
	next, _ = next.Update(click)
	if got := next.(Model).dialog.InputValue(); got != "app.customSignal['environment']=" {
		t.Fatalf("double-click opened assignment %q", got)
	}
	next, _ = next.Update(tea.PasteMsg{Content: "prod"})
	next, apply = next.Update(enter)
	next, _ = next.Update(apply())
	facts := next.(Model).requestContext.facts
	if facts["app.userProperty['environment']"].Value != "qa" || facts["app.customSignal['environment']"].Value != "prod" {
		t.Fatalf("same-name inputs did not retain their separate values: %+v", facts)
	}

	m = next.(Model)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{client}})
	m.openContextNames(contextCustomSignal)
	if m.requestContext.picker.Len() != 2 {
		t.Fatal("names from the deselected server template should disappear")
	}
}

func TestContextCanEnterNewNameAndAppBuild(t *testing.T) {
	m := viewTestModel(100, 36, panels.Context)
	m.openContextFields()
	m.requestContext.picker.Move(5, m.contextLayout(true))
	m.selectContextRow()
	if got := m.dialog.InputValue(); got != "app.build=" {
		t.Fatalf("App build assignment = %q", got)
	}
	esc := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	next, _ := m.Update(esc)
	m = next.(Model)
	m.openContextFields()
	m.openContextNames(contextUserProperty)
	if m.requestContext.picker.Current().Key != contextAnotherName {
		t.Fatal("the manual name choice should be available without any loaded templates")
	}
	enter := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	next, _ = m.Update(enter)
	next, _ = next.Update(tea.PasteMsg{Content: "owner's=cohort"})
	next, choose := next.Update(enter)
	next, _ = next.Update(choose())
	next, _ = next.Update(tea.PasteMsg{Content: "paid"})
	next, apply := next.Update(enter)
	next, _ = next.Update(apply())
	if got := next.(Model).requestContext.facts[`app.userProperty['owner\'s=cohort']`].Value; got != "paid" {
		t.Fatalf("manual name did not round-trip, got %q", got)
	}
}
