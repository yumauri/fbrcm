package managedfeatures

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/components/viewutil"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/styles"
)

func TestManagedFeaturePanelKeepsOnlyClientTemplateProjects(t *testing.T) {
	m := New(nil, messages.ManagedFeatureExperiment).SetBounds(0, 0, 80, 12)
	m, _ = m.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{
		{Name: "Demo server", ProjectID: "server@demo", ProjectNumber: "123"},
		{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"},
	}})

	if len(m.projects) != 1 || m.projects[0].project.ProjectID != "demo" {
		t.Fatalf("managed-feature projects = %#v, want only client demo", m.projects)
	}

	serverOnly := New(nil, messages.ManagedFeatureExperiment).SetBounds(0, 0, 80, 12)
	serverOnly, _ = serverOnly.Update(messages.ProjectsSelectionChangedMsg{
		Projects: []core.Project{{Name: "Demo", ProjectID: "server@demo", ProjectNumber: "123"}},
	})
	if len(serverOnly.projects) != 0 {
		t.Fatalf("server-only selection created managed-feature projects: %#v", serverOnly.projects)
	}
	if got := serverOnly.ViewWithBorder(true, true); !strings.Contains(got, "not available for server templates") {
		t.Fatalf("server-only empty state lacks namespace explanation:\n%s", got)
	}
}

func TestManagedFeaturePanelLoadsOnlyWhileActive(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"}

	experiments := New(nil, messages.ManagedFeatureExperiment)
	experiments, cmd := experiments.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	if cmd != nil || experiments.projects[0].loading {
		t.Fatalf("inactive A/B Tests selection started loading: cmd=%v state=%#v", cmd, experiments.projects[0])
	}
	experiments, cmd = experiments.Activate()
	if cmd == nil || !experiments.projects[0].loading {
		t.Fatalf("activating A/B Tests did not start loading: cmd=%v state=%#v", cmd, experiments.projects[0])
	}
	experiments, _ = experiments.Update(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureExperiment, Project: project,
	})
	experiments = experiments.SetActive(false)
	experiments, cmd = experiments.Activate()
	if cmd != nil || experiments.projects[0].loading {
		t.Fatalf("re-activating loaded A/B Tests reloaded content: cmd=%v state=%#v", cmd, experiments.projects[0])
	}

	personalizations := New(nil, messages.ManagedFeaturePersonalization)
	personalizations, cmd = personalizations.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	if cmd != nil || personalizations.projects[0].loading {
		t.Fatalf("inactive Personalizations selection started loading: cmd=%v state=%#v", cmd, personalizations.projects[0])
	}
	personalizations, cmd = personalizations.Update(messages.ParametersLoadedMsg{Project: project})
	if cmd != nil || !personalizations.projects[0].templateReady || personalizations.projects[0].loading {
		t.Fatalf("inactive template readiness started managed-feature loading: cmd=%v state=%#v", cmd, personalizations.projects[0])
	}
	personalizations, cmd = personalizations.Activate()
	if cmd == nil || !personalizations.projects[0].loading {
		t.Fatalf("activating Personalizations did not start loading: cmd=%v state=%#v", cmd, personalizations.projects[0])
	}
	personalizations, _ = personalizations.Update(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeaturePersonalization, Project: project,
	})
	personalizations = personalizations.SetActive(false)
	personalizations, cmd = personalizations.Activate()
	if cmd != nil || personalizations.projects[0].loading {
		t.Fatalf("re-activating loaded Personalizations reloaded content: cmd=%v state=%#v", cmd, personalizations.projects[0])
	}
}

func TestManagedFeatureProjectSelectionKeepsLoadedProjectsIdle(t *testing.T) {
	alpha := core.Project{Name: "Alpha", ProjectID: "alpha", ProjectNumber: "1"}
	beta := core.Project{Name: "Beta", ProjectID: "beta", ProjectNumber: "2"}
	m := New(nil, messages.ManagedFeatureExperiment).SetActive(true)
	m.setProjects([]core.Project{alpha})
	m.projects[0].templateReady = true
	m.projects[0].loaded = true
	m.projects[0].experiments = []core.ExperimentEntry{{
		Name: "projects/1/namespaces/firebase/experiments/exp-1",
	}}
	m.syncVisible()

	m, cmd := m.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{alpha, beta}})
	if m.projects[0].loading || !m.projects[0].loaded || len(m.projects[0].experiments) != 1 {
		t.Fatalf("existing project was reset after selection change: %#v", m.projects[0])
	}
	if cmd == nil || !m.projects[1].loading || !m.projects[1].waitingTemplate {
		t.Fatalf("new project did not enter template wait: cmd=%v state=%#v", cmd, m.projects[1])
	}
}

func TestManagedFeatureLoadErrorIsActionableAndRetried(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	project := core.Project{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"}
	m := New(nil, messages.ManagedFeatureExperiment).SetBounds(0, 0, 100, 8).SetActive(true)
	m.setProjects([]core.Project{project})
	m.projects[0].templateReady = true
	m, _ = m.startProjectLoads([]int{0}, false)
	m, _ = m.Update(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureExperiment, Project: project, Err: errors.New("permission denied"),
	})
	if m.projects[0].loaded || m.projects[0].loading {
		t.Fatalf("failed load became sticky: %#v", m.projects[0])
	}
	if view := m.ViewWithBorder(true, true); !strings.Contains(view, "permission denied") {
		t.Fatalf("project error is not actionable:\n%s", view)
	}
	m, cmd := m.Activate()
	if cmd == nil || !m.projects[0].loading {
		t.Fatalf("failed project was not retried: cmd=%v state=%#v", cmd, m.projects[0])
	}
}

func TestManagedFeatureTemplateLoadErrorIsRetried(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"}
	m := New(nil, messages.ManagedFeaturePersonalization).SetActive(true)
	m.setProjects([]core.Project{project})
	m, _ = m.Activate()
	if !m.projects[0].waitingTemplate {
		t.Fatalf("project is not waiting for its template: %#v", m.projects[0])
	}
	m, _ = m.Update(messages.ParametersLoadedMsg{Project: project, Err: errors.New("cache unavailable")})
	if m.projects[0].loaded || m.projects[0].loading || m.projects[0].waitingTemplate {
		t.Fatalf("template error became a loaded state: %#v", m.projects[0])
	}
	m, cmd := m.Activate()
	if cmd == nil || !m.projects[0].loading || m.projects[0].waitingTemplate {
		t.Fatalf("template failure was not retried directly: cmd=%v state=%#v", cmd, m.projects[0])
	}
}

func TestManagedFeatureRefreshKeysReloadCurrentOrAllProjects(t *testing.T) {
	alpha := core.Project{Name: "Alpha", ProjectID: "alpha", ProjectNumber: "1"}
	beta := core.Project{Name: "Beta", ProjectID: "beta", ProjectNumber: "2"}
	m := New(nil, messages.ManagedFeatureExperiment).SetActive(true)
	m.setProjects([]core.Project{alpha, beta})
	for index := range m.projects {
		m.projects[index].loaded = true
	}
	m.syncVisible()

	m, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'u'}))
	if cmd == nil || !m.projects[0].loading || m.projects[1].loading {
		t.Fatalf("u loading states = alpha %t, beta %t; want current only", m.projects[0].loading, m.projects[1].loading)
	}
	m.projects[0].loading = false
	m, cmd = m.Update(tea.KeyPressMsg(tea.Key{Code: 'U'}))
	if cmd == nil || !m.projects[0].loading || !m.projects[1].loading {
		t.Fatalf("U loading states = alpha %t, beta %t; want all", m.projects[0].loading, m.projects[1].loading)
	}
}

func TestSelectedProjectRowUsesSharedProjectSelectionStyle(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	m := New(nil, messages.ManagedFeatureExperiment)
	m.projects = []projectState{{project: core.Project{Name: "Demo", ProjectID: "demo"}}}
	m.projectIndex = map[string]int{"demo": 0}

	row := m.renderNode(visibleNode{kind: nodeProject, projectID: "demo", index: -1}, true, 48)
	prefix := managedFeatureStylePrefix(styles.TreeProjectSelectionStyle())
	if prefix == "" || !strings.HasPrefix(row, prefix) {
		t.Fatalf("selected project row does not use shared project style: %q", row)
	}
}

func TestManagedFeaturePanelsRenderEntitiesPerProject(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	tests := []struct {
		name string
		kind messages.ManagedFeatureKind
		load func(project core.Project) messages.ManagedFeaturesLoadedMsg
		want []string
	}{
		{
			name: "A/B tests",
			kind: messages.ManagedFeatureExperiment,
			load: func(project core.Project) messages.ManagedFeaturesLoadedMsg {
				return messages.ManagedFeaturesLoadedMsg{
					Kind: messages.ManagedFeatureExperiment, Project: project,
					Template: core.ManagedFeatureTemplate{Version: "12", Source: "remote"},
					Experiments: []core.ExperimentEntry{{
						Name: "projects/123/namespaces/firebase/experiments/exp-1",
						Definition: firebase.ExperimentDefinition{
							DisplayName: "Passkey signup",
							Description: "Offer passkeys during signup",
							Variants:    []firebase.ExperimentVariant{{Name: "Baseline"}, {Name: "Variant A"}},
						},
						State:          "RUNNING",
						StartTime:      "2026-07-27T09:10:11Z",
						LastUpdateTime: "2026-07-28T10:11:12Z",
						References:     []core.ManagedValueReference{{Parameter: "signup_message"}},
					}},
				}
			},
			want: []string{
				"Alpha", "Beta", "RC v12", "● Passkey signup", "RUNNING",
				"#exp-1", "updated", "1 binding",
			},
		},
		{
			name: "personalizations",
			kind: messages.ManagedFeaturePersonalization,
			load: func(project core.Project) messages.ManagedFeaturesLoadedMsg {
				return messages.ManagedFeaturesLoadedMsg{
					Kind: messages.ManagedFeaturePersonalization, Project: project,
					Template: core.ManagedFeatureTemplate{Version: "12", Source: "remote"},
					Personalizations: []core.PersonalizationEntry{{
						ID: "kyc-provider",
						References: []core.ManagedValueReference{{
							Group: "kyc", Parameter: "kyc_provider", Default: true, ValueType: "STRING",
						}},
					}},
				}
			},
			want: []string{"Alpha", "Beta", "RC v12", "◆ kyc-provider", "kyc / kyc_provider", "Default value", "STRING"},
		},
		{
			name: "rollouts",
			kind: messages.ManagedFeatureRollout,
			load: func(project core.Project) messages.ManagedFeaturesLoadedMsg {
				value := "100"
				percentage := 25.0
				return messages.ManagedFeaturesLoadedMsg{
					Kind: messages.ManagedFeatureRollout, Project: project,
					Template: core.ManagedFeatureTemplate{Version: "12", Source: "remote"},
					Rollouts: []core.RolloutEntry{{
						Name:       "projects/123/namespaces/firebase/rollouts/rollout-1",
						Definition: firebase.RolloutDefinition{DisplayName: "Funding rollout"},
						State:      "RUNNING",
						References: []core.ManagedValueReference{{
							Group: "funding", Parameter: "funding_minimum_amount",
							Value: &value, Percentage: &percentage,
						}},
					}},
				}
			},
			want: []string{
				"Alpha", "Beta", "RC v12", "● Funding rollout", "RUNNING",
				"25% → funding / funding_minimum_amount = 100", "#rollout-1",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alpha := core.Project{Name: "Alpha", ProjectID: "alpha", ProjectNumber: "1"}
			beta := core.Project{Name: "Beta", ProjectID: "beta", ProjectNumber: "2"}
			m := New(nil, test.kind).SetBounds(0, 0, 180, 14).SetActive(true)
			m, _ = m.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{beta, alpha}})
			m, _ = m.Update(test.load(alpha))
			m, _ = m.Update(test.load(beta))
			view := m.ViewWithBorder(true, true)
			plain := ansi.Strip(view)
			for _, want := range test.want {
				if !strings.Contains(plain, want) {
					t.Fatalf("%s panel missing %q:\n%s", test.name, want, view)
				}
			}
			if test.kind == messages.ManagedFeatureExperiment && strings.Contains(plain, "variant") {
				t.Fatalf("A/B Tests panel renders get-only variant data:\n%s", view)
			}
			if test.kind == messages.ManagedFeatureExperiment && strings.Contains(plain, "Offer passkeys during signup") {
				t.Fatalf("A/B Tests summary renders the detail-only description:\n%s", view)
			}
			if strings.Index(plain, "Alpha") > strings.Index(plain, "Beta") {
				t.Fatalf("projects are not sorted and separated:\n%s", view)
			}
		})
	}
}

func TestABTestsFilterMatchesDisplayNameOnly(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	project := core.Project{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"}
	m := New(nil, messages.ManagedFeatureExperiment).SetBounds(0, 0, 120, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m, _ = m.Update(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureExperiment, Project: project,
		Experiments: []core.ExperimentEntry{
			{
				Name: "projects/123/namespaces/firebase/experiments/passkey",
				Definition: firebase.ExperimentDefinition{
					DisplayName: "Passkey signup",
					Description: "Funding description",
				},
			},
			{
				Name: "projects/123/namespaces/firebase/experiments/funding",
				Definition: firebase.ExperimentDefinition{
					DisplayName: "Funding amount",
					Description: "Passkey words in a description do not count",
				},
			},
		},
	})

	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: '/', Text: "/"}))
	m, _ = m.Update(tea.PasteMsg{Content: "passkey"})
	view := m.ViewWithBorder(true, true)
	if !strings.Contains(view, "Passkey signup") || strings.Contains(view, "Funding amount") {
		t.Fatalf("display-name filter rendered unexpected experiments:\n%s", view)
	}
	if !strings.Contains(ansi.Strip(view), "/ passkey") {
		t.Fatalf("A/B Tests filter footer is missing its query:\n%s", view)
	}

	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: '=', Text: "="}))
	m, _ = m.Update(tea.PasteMsg{Content: "Passkey words in a description do not count"})
	if got := m.visibleEntityCount(); got != 0 {
		t.Fatalf("description-only filter matched %d experiments, want 0", got)
	}
}

func TestABTestsListOmitsUnavailableMetadataLabels(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	project := core.Project{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"}
	m := New(nil, messages.ManagedFeatureExperiment).SetBounds(0, 0, 100, 8)
	m.projects = []projectState{{
		project: project,
		experiments: []core.ExperimentEntry{{
			Name:       "projects/123/namespaces/firebase/experiments/signup",
			Definition: firebase.ExperimentDefinition{DisplayName: "Signup"},
		}},
	}}
	m.projectIndex = map[string]int{"demo": 0}
	m.syncVisible()

	view := m.ViewWithBorder(true, true)
	if strings.Contains(view, "started ") || strings.Contains(view, "updated ") {
		t.Fatalf("A/B Tests list labels unavailable timestamps:\n%s", view)
	}
	if !strings.Contains(ansi.Strip(view), "bindings not exposed") {
		t.Fatalf("A/B Tests list presents absent associations as a definitive zero:\n%s", view)
	}
}

func TestManagedFeatureSummaryCompactsMultipleParameterValues(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	percentage := 20.0
	references := []core.ManagedValueReference{
		{Group: "checkout", Parameter: "enabled", Percentage: &percentage},
		{Group: "checkout", Parameter: "layout", Percentage: &percentage},
		{Group: "checkout", Parameter: "headline", Percentage: &percentage},
	}

	personalization := personalizationReferencesMeta(references)
	for _, want := range []string{"3 parameter values", "checkout / enabled", "checkout / layout", "+1 more"} {
		if !strings.Contains(personalization, want) {
			t.Fatalf("personalization summary missing %q: %q", want, personalization)
		}
	}
	rollout := rolloutReferencesMeta("rollout-1", references)
	for _, want := range []string{"20%", "3 parameter values", "checkout / enabled", "+1 more", "#rollout-1"} {
		if !strings.Contains(rollout, want) {
			t.Fatalf("rollout summary missing %q: %q", want, rollout)
		}
	}
}

func TestManagedFeatureSummaryUsesSharedColorsAndKeepsStatusBesideName(t *testing.T) {
	t.Setenv("NO_COLOR", "")

	title := renderEntityTitle("●", "Passkey signup", "RUNNING", 100)
	if got := strings.TrimSpace(ansi.Strip(title)); got != "● Passkey signup · RUNNING" {
		t.Fatalf("experiment title = %q, want status beside name", got)
	}
	runningPrefix := managedFeatureStylePrefix(
		lipgloss.NewStyle().Foreground(styles.PaletteSuccess),
	)
	if strings.Count(title, runningPrefix) < 2 {
		t.Fatalf("running marker and status do not use the shared success color: %q", title)
	}

	reference := viewutil.RenderManagedReferenceIdentity(core.ManagedValueReference{
		Group: "checkout", Parameter: "enabled",
		Condition: "android_beta", ConditionColor: "GREEN",
	})
	if !strings.Contains(reference, styles.ParameterGroup.Render("checkout")) {
		t.Fatalf("reference group does not use the shared group style: %q", reference)
	}
	if !strings.Contains(reference, styles.ParameterName.Render("enabled")) {
		t.Fatalf("reference parameter does not use the shared parameter style: %q", reference)
	}
	if !strings.Contains(reference, styles.DetailsConditionValueStyle("GREEN").Render("android_beta")) {
		t.Fatalf("reference condition does not use its configured color: %q", reference)
	}

	statuses := []struct {
		state string
		style lipgloss.Style
	}{
		{state: "RUNNING", style: lipgloss.NewStyle().Foreground(styles.PaletteSuccess)},
		{state: "PENDING", style: lipgloss.NewStyle().Foreground(styles.PaletteYellow)},
		{state: "DONE", style: lipgloss.NewStyle().Foreground(styles.PaletteBlueBright)},
		{state: "EXPIRED", style: lipgloss.NewStyle().Foreground(styles.PaletteSlateDim)},
	}
	for _, status := range statuses {
		t.Run(status.state, func(t *testing.T) {
			got := managedFeatureStylePrefix(styles.ManagedFeatureStatusStyle(status.state))
			want := managedFeatureStylePrefix(status.style)
			if got == "" || got != want {
				t.Fatalf("%s status prefix = %q, want %q", status.state, got, want)
			}
		})
	}
}

func TestManagedFeatureSummaryColorsRespectNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	title := renderEntityTitle("●", "Passkey signup", "RUNNING", 100)
	reference := viewutil.RenderManagedReferenceIdentity(core.ManagedValueReference{
		Group: "checkout", Parameter: "enabled",
		Condition: "android_beta", ConditionColor: "GREEN",
	})
	for name, rendered := range map[string]string{"title": title, "reference": reference} {
		if strings.Contains(rendered, "\x1b[") {
			t.Fatalf("%s contains color escapes with NO_COLOR: %q", name, rendered)
		}
	}
}

func TestManagedFeatureEntityOpensWithEnterAndDoubleClick(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"}
	m := New(nil, messages.ManagedFeatureExperiment).SetBounds(10, 3, 80, 12).SetActive(true)
	m, _ = m.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m, _ = m.Update(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureExperiment, Project: project,
		Experiments: []core.ExperimentEntry{{
			Name:       "projects/123/namespaces/firebase/experiments/exp-1",
			Definition: firebase.ExperimentDefinition{DisplayName: "Signup"},
		}},
	})

	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	m, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	assertActivatedSelection(t, cmd, "exp-1")

	titleClick := tea.MouseClickMsg{X: 12, Y: 5, Button: tea.MouseLeft}
	m, first := m.Update(titleClick)
	if msg := commandMessage(first); isActivatedSelection(msg) {
		t.Fatalf("single click activated Details: %#v", msg)
	}
	summaryClick := tea.MouseClickMsg{X: 12, Y: 6, Button: tea.MouseLeft}
	_, second := m.Update(summaryClick)
	assertActivatedSelection(t, second, "exp-1")
}

func TestManagedFeatureSelectionStylesBothSummaryLines(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m := New(nil, messages.ManagedFeatureRollout)
	m.projects = []projectState{{
		project: project,
		rollouts: []core.RolloutEntry{{
			Name:       "projects/demo/namespaces/firebase/rollouts/rollout-1",
			Definition: firebase.RolloutDefinition{DisplayName: "Checkout"},
			State:      "RUNNING"}},
	}}
	m.projectIndex = map[string]int{"demo": 0}
	m.syncVisible()

	lines := m.renderNodeLines(m.visible[1], true, 48)
	if len(lines) != 2 {
		t.Fatalf("selected rollout rendered %d lines, want 2", len(lines))
	}
	prefix := managedFeatureStylePrefix(styles.TreeItemSelectionStyle())
	for index, line := range lines {
		if prefix == "" || !strings.HasPrefix(line, prefix) {
			t.Fatalf("selected rollout line %d does not use shared item style: %q", index, line)
		}
	}
}

func TestManagedFeatureSummaryRowsRespectNarrowPanelWidth(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	value := "A very long rollout value that cannot fit"
	percentage := 12.5
	project := core.Project{Name: "A very long project name", ProjectID: "long-project-id"}
	m := New(nil, messages.ManagedFeatureRollout).SetBounds(0, 0, 34, 8)
	m.projects = []projectState{{
		project:  project,
		template: core.ManagedFeatureTemplate{Version: "123456789"},
		rollouts: []core.RolloutEntry{{
			Name:       "projects/demo/namespaces/firebase/rollouts/long-rollout-id",
			Definition: firebase.RolloutDefinition{DisplayName: "A very long rollout display name"},
			State:      "RUNNING",
			References: []core.ManagedValueReference{{
				Group: "long group", Parameter: "long_parameter_name", Value: &value, Percentage: &percentage,
			}},
		}},
	}}
	m.projectIndex = map[string]int{project.ProjectID: 0}
	m.syncVisible()

	view := m.ViewWithBorder(true, true)
	for index, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(line); got != 34 {
			t.Fatalf("narrow panel line %d width = %d, want 34:\n%s", index, got, view)
		}
	}
	if !strings.Contains(view, "…") {
		t.Fatalf("narrow panel did not apply explicit ellipsis overflow:\n%s", view)
	}
}

func TestManagedFeaturePageNavigationUsesRenderedRowHeights(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m := New(nil, messages.ManagedFeatureRollout).SetBounds(0, 0, 60, 5)
	m.projects = []projectState{{
		project: project,
		rollouts: []core.RolloutEntry{
			{Name: "projects/demo/namespaces/firebase/rollouts/one"},
			{Name: "projects/demo/namespaces/firebase/rollouts/two"},
			{Name: "projects/demo/namespaces/firebase/rollouts/three"},
		},
	}}
	m.projectIndex = map[string]int{"demo": 0}
	m.syncVisible()

	m.movePage(m.contentHeight())
	if got := m.visible[m.cursor].entityID; got != "two" {
		t.Fatalf("page down selected %q, want rollout two", got)
	}
	if m.offset != 2 {
		t.Fatalf("page down scroll offset = %d, want physical row 2", m.offset)
	}
	m.movePage(-m.contentHeight())
	if m.visible[m.cursor].kind != nodeProject || m.offset != 0 {
		t.Fatalf("page up did not return to project: cursor=%#v offset=%d", m.visible[m.cursor], m.offset)
	}
}

func assertActivatedSelection(t *testing.T, cmd tea.Cmd, wantID string) {
	t.Helper()
	msg := commandMessage(cmd)
	selection, ok := msg.(messages.ManagedFeatureSelectionChangedMsg)
	if !ok || !selection.Activate || selection.Data == nil || firebase.ManagedFeatureID(selection.Data.Experiment.Name) != wantID {
		t.Fatalf("selection message = %#v, want activated %s", msg, wantID)
	}
}

func commandMessage(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

func isActivatedSelection(msg tea.Msg) bool {
	selection, ok := msg.(messages.ManagedFeatureSelectionChangedMsg)
	return ok && selection.Activate
}

func managedFeatureStylePrefix(style lipgloss.Style) string {
	rendered := style.Render("x")
	prefix, _, _ := strings.Cut(rendered, "x")
	return prefix
}
