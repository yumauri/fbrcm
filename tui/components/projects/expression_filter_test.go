package projects

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestProjectsExpressionFilterUsesCachedConfigAndKeepsOperatorsTypable(t *testing.T) {
	m := loadedProjectsModel()
	m.expressionConfigsReady = true
	m.expressionConfigs = map[string]*firebase.RemoteConfig{
		"alpha": {
			Parameters: map[string]firebase.RemoteConfigParam{
				"enabled": {ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "true"}},
			},
		},
		"beta": {
			Parameters: map[string]firebase.RemoteConfigParam{
				"enabled": {ValueType: "BOOLEAN", DefaultValue: &firebase.RemoteConfigValue{Value: "false"}},
			},
		},
	}

	m, _ = m.Update(keyText(":"))
	m, _ = m.Update(tea.PasteMsg{Content: `parameters["enabled"].default == true`})
	if len(m.projects) != 1 || m.projects[0].ProjectID != "alpha" {
		t.Fatalf("expression-filtered projects = %+v, want alpha", m.projects)
	}

	m, _ = m.Update(keyText("="))
	if !m.filter.ExpressionMode() || m.filter.Value() != `parameters["enabled"].default == true=` {
		t.Fatalf("equals key changed mode or was not entered: mode=%v value=%q", m.filter.ExpressionMode(), m.filter.Value())
	}
	if len(m.projects) != 1 || m.projects[0].ProjectID != "alpha" {
		t.Fatalf("temporary invalid expression changed last valid results: %+v", m.projects)
	}
}

func TestProjectsExpressionErrorRendersAboveUnchangedBottomBorder(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := loadedProjectsModel()
	m.expressionConfigsReady = true
	m, _ = m.Update(keyText(":"))
	validHeight := len(strings.Split(ansi.Strip(m.View(true)), "\n"))
	m, _ = m.Update(tea.PasteMsg{Content: `name ==`})

	lines := strings.Split(ansi.Strip(m.View(true)), "\n")
	if len(lines) != validHeight || m.filter.Height() != 2 {
		t.Fatalf("invalid expression changed layout height: panel=%d/%d filter=%d", len(lines), validHeight, m.filter.Height())
	}
	bottom := lines[len(lines)-1]
	if !strings.Contains(bottom, "Expression error:") || !strings.HasPrefix(bottom, "──") || !strings.HasSuffix(bottom, "╯") {
		t.Fatalf("expression error was not overlaid on the bottom border: %q", bottom)
	}
}
