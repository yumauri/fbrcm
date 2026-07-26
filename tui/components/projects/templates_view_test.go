package projects

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	rctarget "github.com/yumauri/fbrcm/core/rc/target"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/testutil"
)

func TestExpandedTemplatePairRendersAsConnectedProject(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := templatePairViewModel(rctarget.Client)

	got := normalizedProjectContent(m)
	want := `╭PulseForge Fitness
│ pulseforge-fitness-f60f7
│
╰PulseForge Fitness
  server@pulseforge-fitness-f60f7

 Other Project
  other`
	if got != want {
		t.Fatalf("content mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestExpandedTemplatePairConnectorFollowsPrimaryOrder(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := templatePairViewModel(rctarget.Server)

	got := normalizedProjectContent(m)
	want := `╭PulseForge Fitness
│ server@pulseforge-fitness-f60f7
│
╰PulseForge Fitness
  pulseforge-fitness-f60f7

 Other Project
  other`
	if got != want {
		t.Fatalf("content mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFilteredTemplateDoesNotRenderPartialConnector(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := templatePairViewModel(rctarget.Client)
	m, _ = m.Update(keyText("/"))
	m, _ = m.Update(tea.PasteMsg{Content: "server@"})

	got := normalizedProjectContent(m)
	want := ` PulseForge Fitness
  server@pulseforge-fitness-f60f7`
	if got != want {
		t.Fatalf("content mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func templatePairViewModel(primary rctarget.Kind) Model {
	m := New(nil).SetBounds(0, 0, 40, 14).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{
		Projects: []core.Project{
			{
				Name:            "PulseForge Fitness",
				ProjectID:       "pulseforge-fitness-f60f7",
				Templates:       []rctarget.Kind{rctarget.Client, rctarget.Server},
				PrimaryTemplate: primary,
			},
			{Name: "Other Project", ProjectID: "other"},
		},
		Source: "cache",
	})
	return m
}

func normalizedProjectContent(m Model) string {
	return testutil.NormalizeViewSnapshot(strings.Join(m.renderContentLines(), "\n"))
}
