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
	want := `╭Example Fitness App
│ example-fitness-app-a1b2
│
╰Example Fitness App
  server@example-fitness-app-a1b2

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
	want := `╭Example Fitness App
│ server@example-fitness-app-a1b2
│
╰Example Fitness App
  example-fitness-app-a1b2

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
	want := ` Example Fitness App
  server@example-fitness-app-a1b2`
	if got != want {
		t.Fatalf("content mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestProjectAliasesRenderAndFilter(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	m := templatePairViewModel(rctarget.Client)
	m.aliasesByID = map[string][]string{"example-fitness-app-a1b2": {"prod", "production"}}
	m.applyFilter()
	m.syncViewport()

	got := normalizedProjectContent(m)
	if !strings.Contains(got, "Example Fitness App [prod, production]") {
		t.Fatalf("alias content = %q", got)
	}
	wantMeta := "[prod, production]"
	if len(m.lineMeta) == 0 || len(m.lineMeta[0]) != len([]rune(wantMeta)) {
		t.Fatalf("alias metadata = %#v, want bracketed suffix %q", m.lineMeta, wantMeta)
	}
	m, _ = m.Update(keyText("="))
	m, _ = m.Update(tea.PasteMsg{Content: "prod"})
	if len(m.projects) != 2 || m.projects[0].ProjectID != "example-fitness-app-a1b2" || m.projects[1].ProjectID != "server@example-fitness-app-a1b2" {
		t.Fatalf("alias-filtered projects = %#v", m.projects)
	}
	m.contentLines()
	if len(m.lineHighlights) == 0 || len(m.lineHighlights[0]) != len("prod") {
		t.Fatalf("alias highlights = %#v", m.lineHighlights)
	}
}

func templatePairViewModel(primary rctarget.Kind) Model {
	m := New(nil).SetBounds(0, 0, 40, 14).SetActive(true)
	m, _ = m.Update(messages.ProjectsLoadedMsg{
		Projects: []core.Project{
			{
				Name:            "Example Fitness App",
				ProjectID:       "example-fitness-app-a1b2",
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
