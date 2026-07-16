package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestConditionsTabRemainsSelectedAcrossProjectsFocus(t *testing.T) {
	m := New(nil)
	m.setActive(panels.Conditions)
	m.setActive(panels.Projects)

	if got := m.nextTabPanel(); got != panels.Conditions {
		t.Fatalf("next tab from Projects = %v, want Conditions", got)
	}
	next, _, _ := m.updateAppMessage(messages.SetActivePanelMsg{Panel: panels.Parameters})
	if next.active != panels.Conditions {
		t.Fatalf("project selection active = %v; want Conditions", next.active)
	}
}

func TestGlobalFocusKeysWorkFromConditionDetails(t *testing.T) {
	m := New(nil)
	m.setActive(panels.Conditions)
	m.applyConditionSelection(messages.ConditionSelectionChangedMsg{
		Data: &messages.ConditionViewData{
			Project:   core.Project{ProjectID: "demo"},
			Condition: core.ConditionEntry{Name: "beta"},
		},
		Activate: true,
	})

	next, _, handled := m.updateAppMessage(tea.KeyPressMsg(tea.Key{Code: '2'}))
	if !handled || next.active != panels.Parameters {
		t.Fatalf("2 from condition Details: active=%v handled=%v; want Parameters, true", next.active, handled)
	}

	next.setActive(panels.Details)
	next, _, handled = next.updateAppMessage(tea.KeyPressMsg(tea.Key{Code: '3'}))
	if !handled || next.active != panels.Conditions {
		t.Fatalf("3 from condition Details: active=%v handled=%v; want Conditions, true", next.active, handled)
	}
}

func TestGlobalFocusKeyWorksFromParameterDetails(t *testing.T) {
	m := New(nil)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project:   core.Project{ProjectID: "demo"},
		Parameter: core.ParametersEntry{Key: "feature"},
	})
	m.detailsVisible = true
	m.setActive(panels.Details)

	next, _, handled := m.updateAppMessage(tea.KeyPressMsg(tea.Key{Code: '3'}))
	if !handled || next.active != panels.Conditions {
		t.Fatalf("3 from parameter Details: active=%v handled=%v; want Conditions, true", next.active, handled)
	}
}

func TestGlobalFocusKeyDoesNotStealDetailsFieldInput(t *testing.T) {
	m := New(nil)
	m.details = m.details.SetData(&messages.ParameterViewData{
		Project:   core.Project{ProjectID: "demo"},
		Parameter: core.ParametersEntry{Key: "feature"},
	})
	m.detailsVisible = true
	m.setActive(panels.Details)
	m.details, _ = m.details.ActivateName()

	next, _, handled := m.updateAppMessage(tea.KeyPressMsg(tea.Key{Code: '2'}))
	if !handled || next.active != panels.Details {
		t.Fatalf("2 while editing Details: active=%v handled=%v; want Details, true", next.active, handled)
	}
}

func TestConditionSelectionOpensReadOnlyDetailsAndReturnsToConditions(t *testing.T) {
	m := New(nil)
	m.setActive(panels.Conditions)
	data := &messages.ConditionViewData{
		Project: core.Project{ProjectID: "demo", Name: "Demo"},
		Condition: core.ConditionEntry{
			Priority:   1,
			Name:       "beta",
			Expression: "app.version > '2'",
		},
	}

	m.applyConditionSelection(messages.ConditionSelectionChangedMsg{Data: data, Activate: true})
	if !m.detailsVisible || m.active != panels.Details || !m.details.IsCondition() {
		t.Fatalf("condition details state = visible:%v active:%v condition:%v", m.detailsVisible, m.active, m.details.IsCondition())
	}
	m.closeDetailsPanel()
	if m.detailsVisible || m.active != panels.Conditions {
		t.Fatalf("closed condition details state = visible:%v active:%v", m.detailsVisible, m.active)
	}
}

func TestConditionMouseClickDoesNotReachHiddenParametersPanel(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := viewTestModel(100, 30, panels.Conditions)
	m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
		Project: project,
		Tree: &core.ParametersTree{Groups: []core.ParametersGroup{{
			Key:   "__default__",
			Label: "(root)",
			Parameters: []core.ParametersEntry{{
				Key: "hidden_parameter",
				Values: []core.ParametersValue{{
					Label: "default", Value: "hidden", RawValue: "hidden", ValueType: "STRING", Plain: true,
				}},
			}},
		}}},
	})
	m.conditions, _ = m.conditions.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.conditions, _ = m.conditions.Update(messages.ConditionsLoadedMsg{
		Project: project,
		Tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{
			{Priority: 1, Name: "first", Expression: "true"},
			{Priority: 2, Name: "second", Expression: "true"},
		}},
	})
	m.details = m.details.SetConditionData(&messages.ConditionViewData{
		Project: project, Condition: core.ConditionEntry{Name: "first"},
	})
	m.detailsVisible = true
	m.setActive(panels.Conditions)
	m.applyLayout()

	layout := newPanelLayout(m.width, m.height, m.projects.PreferredWidth(), m.logsHeight, m.projectsMode)
	next, cmd, handled := m.updatePanelMouseMessage(tea.MouseClickMsg{
		X: layout.leftWidth + 2, Y: 3, Button: tea.MouseLeft,
	})
	if !handled || cmd == nil {
		t.Fatalf("condition click handled=%v cmd=%v; want true and selection command", handled, cmd)
	}
	selection, ok := cmd().(messages.ConditionSelectionChangedMsg)
	if !ok || selection.Data == nil || selection.Data.Condition.Name != "second" {
		t.Fatalf("condition click emitted %#v; want second condition selection", selection)
	}
	next.applyConditionSelection(selection)
	if !next.details.IsCondition() || next.details.ConditionData().Condition.Name != "second" {
		t.Fatalf("Details after condition click = %#v; want second condition", next.details.ConditionData())
	}
	if _, ok := next.parameters.CurrentParameterViewData(); ok {
		t.Fatal("hidden Parameters panel selected a parameter from the condition click coordinates")
	}
}

func TestConditionsUsesWorkspaceMaximizeBinding(t *testing.T) {
	m := New(nil)
	m.setActive(panels.Conditions)

	next, _, handled := m.updateGlobalKeyMessage("z")
	if !handled || next.projectsMode != projectsPanelModeCollapsed || next.logsMode != logsPanelModeCollapsed {
		t.Fatalf("z from Conditions: handled=%v projects=%v logs=%v; want maximized workspace", handled, next.projectsMode, next.logsMode)
	}
	next, _, handled = next.updateGlobalKeyMessage("z")
	if !handled || next.projectsMode != projectsPanelModeExpanded || next.logsMode != logsPanelModeExpanded {
		t.Fatalf("second z from Conditions: handled=%v projects=%v logs=%v; want restored workspace", handled, next.projectsMode, next.logsMode)
	}
}

func TestConditionsUsesParameterReloadBindings(t *testing.T) {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	newModel := func() Model {
		m := New(nil)
		m.parameters, _ = m.parameters.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
		m.parameters, _ = m.parameters.Update(messages.ParametersLoadedMsg{
			Project: project,
			Tree:    &core.ParametersTree{Version: "1"},
			Source:  "cache",
		})
		m.conditions, _ = m.conditions.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
		m.conditions, _ = m.conditions.Update(messages.ConditionsLoadedMsg{
			Project: project,
			Tree:    &core.ConditionsTree{Version: "1"},
			Source:  "cache",
		})
		m.setActive(panels.Conditions)
		return m
	}

	for _, key := range []string{"u", "U"} {
		m := newModel()
		next, cmd, handled := m.updateGlobalKeyMessage(key)
		if !handled || cmd == nil {
			t.Fatalf("%s from Conditions: handled=%v cmd=%v; want refresh command", key, handled, cmd)
		}
		if got := next.conditions.ViewWithBorder(true, true); strings.Contains(got, "staled") {
			t.Fatalf("%s did not mark Conditions project as reloading:\n%s", key, got)
		}
	}
}

func TestConditionManagementKeysOpenSharedEditors(t *testing.T) {
	tests := []struct {
		key   string
		check func(Model) bool
	}{
		{key: "a", check: func(m Model) bool {
			_, anchorX, anchorY, ok := m.conditions.CurrentProjectAnchor()
			x, y := m.renameInput.Position()
			return m.renameInput.IsOpen() && m.conditionEdit.mode == conditionAddName && ok &&
				x == anchorX+4 && y == anchorY
		}},
		{key: "r", check: func(m Model) bool {
			anchor, ok := m.conditions.CurrentEditAnchor()
			x, y := m.renameInput.Position()
			wantX, wantY := anchor.NameOverlayPosition()
			return m.renameInput.IsOpen() && m.conditionEdit.mode == conditionRename && ok &&
				x == wantX && y == wantY
		}},
		{key: "e", check: func(m Model) bool {
			return m.stringInput.IsOpen() && m.stringInput.IsExpanded() && m.conditionEdit.mode == conditionExpression
		}},
		{key: "c", check: func(m Model) bool {
			anchor, ok := m.conditions.CurrentEditAnchor()
			x, y := m.moveParam.Position()
			wantX, wantY := anchor.NameOverlayPosition()
			list := ansi.Strip(m.moveParam.ListView())
			return m.moveParam.IsOpen() && m.conditionEdit.mode == conditionColor && ok &&
				x == wantX && y == wantY && strings.Contains(list, "● GREEN") && strings.Contains(list, "● DEEP ORANGE")
		}},
		{key: "m", check: func(m Model) bool {
			return m.conditions.MoveActive() && !m.renameInput.IsOpen() && m.conditionEdit.mode == conditionMove
		}},
	}
	for _, test := range tests {
		m := conditionManagementTestModel()
		next, _, handled := m.updateGlobalKeyMessage(test.key)
		if !handled || next.conditionEdit == nil || !test.check(next) {
			t.Fatalf("condition key %q: handled=%v session=%#v", test.key, handled, next.conditionEdit)
		}
	}
}

func TestConditionDetailsDefinitionKeysUseInlineFields(t *testing.T) {
	newDetailsModel := func() Model {
		m := conditionManagementTestModel()
		data, ok := m.conditions.CurrentCondition()
		if !ok {
			t.Fatal("condition not selected")
		}
		m.details = m.details.SetConditionData(data)
		m.detailsVisible = true
		m.setActive(panels.Details)
		return m
	}

	t.Run("name", func(t *testing.T) {
		m, _, handled := newDetailsModel().updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'r', Text: "r"}))
		if !handled || !m.details.TextInputActive() || m.renameInput.IsOpen() {
			t.Fatalf("name edit = handled:%v text:%v row overlay:%v", handled, m.details.TextInputActive(), m.renameInput.IsOpen())
		}
	})

	t.Run("priority", func(t *testing.T) {
		m, _, handled := newDetailsModel().updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'm', Text: "m"}))
		if !handled || !m.details.TextInputActive() || m.conditions.MoveActive() {
			t.Fatalf("priority edit = handled:%v text:%v row move:%v", handled, m.details.TextInputActive(), m.conditions.MoveActive())
		}
	})

	t.Run("color", func(t *testing.T) {
		m, _, handled := newDetailsModel().updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'c', Text: "c"}))
		if !handled || !m.details.DropdownOpen() || m.moveParam.IsOpen() {
			t.Fatalf("color edit = handled:%v details picker:%v row picker:%v", handled, m.details.DropdownOpen(), m.moveParam.IsOpen())
		}
	})

	t.Run("right opens focused color", func(t *testing.T) {
		m := newDetailsModel()
		m.details = m.details.ActivateConditionColor()
		m.details = m.details.DeactivateField()
		m.details = m.details.ActivateConditionColor()
		m.details, _ = m.details.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
		if m.details.DropdownOpen() || !m.details.FieldActive() {
			t.Fatalf("color setup = field:%v picker:%v; want focused, closed", m.details.FieldActive(), m.details.DropdownOpen())
		}

		m, _, handled := m.updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: tea.KeyRight}))
		if !handled || !m.details.DropdownOpen() || m.stringInput.IsOpen() {
			t.Fatalf("right on color = handled:%v picker:%v expression:%v", handled, m.details.DropdownOpen(), m.stringInput.IsOpen())
		}
	})

	t.Run("expression", func(t *testing.T) {
		m, _, handled := newDetailsModel().updateKeyMessage(tea.KeyPressMsg(tea.Key{Code: 'e', Text: "e"}))
		if !handled || !m.stringInput.IsOpen() || m.conditionEdit == nil || m.conditionEdit.mode != conditionDetailsExpression {
			t.Fatalf("expression edit = handled:%v open:%v session:%#v", handled, m.stringInput.IsOpen(), m.conditionEdit)
		}
	})
}

func TestConditionMoveKeysReorderCancelAndSubmit(t *testing.T) {
	m := conditionManagementTestModel()
	m, _, handled := m.updateGlobalKeyMessage("m")
	if !handled || !m.conditions.MoveActive() {
		t.Fatalf("move start handled=%v active=%v; want true, true", handled, m.conditions.MoveActive())
	}

	m, _, handled = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	if !handled || strings.Join(conditionOrder(m), ",") != "second,first" {
		t.Fatalf("down move handled=%v order=%v", handled, conditionOrder(m))
	}
	m, _, _ = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: 'k', Text: "k"}))
	if got := strings.Join(conditionOrder(m), ","); got != "first,second" {
		t.Fatalf("k move order = %s, want first,second", got)
	}
	m, _, _ = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
	if got := strings.Join(conditionOrder(m), ","); got != "second,first" {
		t.Fatalf("j move order = %s, want second,first", got)
	}
	m, _, handled = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	if !handled || m.conditions.MoveActive() || strings.Join(conditionOrder(m), ",") != "first,second" {
		t.Fatalf("cancel handled=%v active=%v order=%v", handled, m.conditions.MoveActive(), conditionOrder(m))
	}

	m, _, _ = m.updateGlobalKeyMessage("m")
	m, _, _ = m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	m, cmd, handled := m.updateOpenModal(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !handled || cmd == nil || m.conditions.MoveActive() || m.conditionEdit != nil {
		t.Fatalf("submit handled=%v cmd=%v active=%v session=%#v", handled, cmd, m.conditions.MoveActive(), m.conditionEdit)
	}
}

func conditionManagementTestModel() Model {
	project := core.Project{ProjectID: "demo", Name: "Demo"}
	m := viewTestModel(100, 30, panels.Conditions)
	m.conditions, _ = m.conditions.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.conditions, _ = m.conditions.Update(messages.ConditionsLoadedMsg{
		Project: project,
		Source:  "draft",
		Tree: &core.ConditionsTree{Conditions: []core.ConditionEntry{
			{Priority: 1, Name: "first", Expression: "true", TagColor: "GREEN"},
			{Priority: 2, Name: "second", Expression: "false"},
		}},
	})
	m.setActive(panels.Conditions)
	m.conditions, _ = m.conditions.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	return m
}

func conditionOrder(m Model) []string {
	entries, _ := m.conditions.CurrentConditions()
	names := make([]string, len(entries))
	for index := range entries {
		names[index] = entries[index].Name
	}
	return names
}
