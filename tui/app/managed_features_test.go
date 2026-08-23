package app

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/env"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/tui/messages"
	"github.com/yumauri/fbrcm/tui/panels"
)

func TestManagedFeatureSelectionOpensReadOnlyDetailsAndReturnsToItsTab(t *testing.T) {
	m := newModel(nil, nil)
	m.setActive(panels.ABTests)
	m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{
		Activate: true,
		Data: &messages.ManagedFeatureViewData{
			Kind:    messages.ManagedFeatureExperiment,
			Project: core.Project{Name: "Demo", ProjectID: "demo"},
			Experiment: &core.ExperimentEntry{
				Name:       "projects/123/namespaces/firebase/experiments/exp-1",
				Definition: firebase.ExperimentDefinition{DisplayName: "Signup"},
			},
		},
	})

	if !m.detailsVisible || m.active != panels.Details || m.parametersTab != panels.ABTests || !m.details.IsManagedFeature() {
		t.Fatalf(
			"managed-feature Details state = visible %t, active %v, tab %v, managed %t",
			m.detailsVisible,
			m.active,
			m.parametersTab,
			m.details.IsManagedFeature(),
		)
	}

	next, cmd, handled := m.updateDetailsKeyMessage(
		tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}),
		"esc",
	)
	if !handled || cmd != nil || next.detailsVisible || next.active != panels.ABTests {
		t.Fatalf(
			"closing managed-feature Details = handled %t, cmd %v, visible %t, active %v",
			handled,
			cmd,
			next.detailsVisible,
			next.active,
		)
	}
}

func TestManagedFeatureDetailsOpenFromListThenRefreshExactResource(t *testing.T) {
	tests := []struct {
		name       string
		kind       messages.ManagedFeatureKind
		collection string
		data       func(project core.Project) *messages.ManagedFeatureViewData
		response   string
		assert     func(t *testing.T, data *messages.ManagedFeatureViewData)
	}{
		{
			name:       "experiment",
			kind:       messages.ManagedFeatureExperiment,
			collection: "experiments",
			data: func(project core.Project) *messages.ManagedFeatureViewData {
				return &messages.ManagedFeatureViewData{
					Kind: messages.ManagedFeatureExperiment, Project: project,
					Experiment: &core.ExperimentEntry{
						Name:       "projects/123/namespaces/firebase/experiments/exp-1",
						Definition: firebase.ExperimentDefinition{DisplayName: "Cached experiment"},
					},
				}
			},
			response: `{
				"name":"projects/123/namespaces/firebase/experiments/exp-1",
				"definition":{"displayName":"Fresh experiment","variants":[{"name":"Variant A","weight":50}]},
				"state":"RUNNING"
			}`,
			assert: func(t *testing.T, data *messages.ManagedFeatureViewData) {
				t.Helper()
				if data.Experiment == nil ||
					data.Experiment.Definition.DisplayName != "Fresh experiment" ||
					len(data.Experiment.Definition.Variants) != 1 {
					t.Fatalf("fresh experiment details = %#v", data.Experiment)
				}
			},
		},
		{
			name:       "rollout",
			kind:       messages.ManagedFeatureRollout,
			collection: "rollouts",
			data: func(project core.Project) *messages.ManagedFeatureViewData {
				return &messages.ManagedFeatureViewData{
					Kind: messages.ManagedFeatureRollout, Project: project,
					Rollout: &core.RolloutEntry{
						Name:       "projects/123/namespaces/firebase/rollouts/rollout-1",
						Definition: firebase.RolloutDefinition{DisplayName: "Cached rollout"},
						References: []core.ManagedValueReference{{Parameter: "funding"}},
					},
				}
			},
			response: `{
				"name":"projects/123/namespaces/firebase/rollouts/rollout-1",
				"definition":{"displayName":"Fresh rollout","enabledVariant":{"name":"Enabled","weight":10}},
				"state":"RUNNING"
			}`,
			assert: func(t *testing.T, data *messages.ManagedFeatureViewData) {
				t.Helper()
				if data.Rollout == nil ||
					data.Rollout.Definition.DisplayName != "Fresh rollout" ||
					data.Rollout.Definition.EnabledVariant.Name != "Enabled" ||
					len(data.Rollout.References) != 1 ||
					data.Rollout.References[0].Parameter != "funding" {
					t.Fatalf("fresh rollout details = %#v", data.Rollout)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project := core.Project{
				Name: "Demo", ProjectID: "demo", ProjectNumber: "123", AuthID: "main",
				DiscoveredBy: []string{"main"},
			}
			var requests []string
			svc := managedFeatureAppTestService(t, project, func(req *http.Request) (*http.Response, error) {
				requests = append(requests, req.URL.RequestURI())
				return managedFeatureJSONResponse(test.response), nil
			})
			m := newModel(svc, nil)
			m.setActive(panels.ABTests)
			listData := test.data(project)

			cmd := m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{
				Data: listData, Activate: true,
			})
			if cmd == nil || !m.detailsVisible || managedFeatureDataID(m.details.ManagedFeatureData()) != managedFeatureDataID(listData) {
				t.Fatalf("cached Details did not open before refresh: cmd=%v data=%#v", cmd, m.details.ManagedFeatureData())
			}
			if len(requests) != 0 {
				t.Fatalf("opening cached Details synchronously made requests: %#v", requests)
			}

			loaded, ok := cmd().(managedFeatureDetailsLoadedMsg)
			if !ok || loaded.err != nil {
				t.Fatalf("lazy refresh message = %#v", loaded)
			}
			wantPath := "/v1/projects/123/namespaces/firebase/" + test.collection + "/" + managedFeatureDataID(listData)
			if len(requests) != 1 || requests[0] != wantPath {
				t.Fatalf("lazy refresh requests = %#v, want [%q]", requests, wantPath)
			}
			m.applyManagedFeatureDetailsLoaded(loaded)
			test.assert(t, m.details.ManagedFeatureData())
		})
	}
}

func TestManagedFeatureDetailsIgnoreLateRefreshForAnotherSelection(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m := newModel(nil, nil)
	m.detailsVisible = true
	m.details = m.details.SetManagedFeatureData(&messages.ManagedFeatureViewData{
		Kind: messages.ManagedFeatureExperiment, Project: project,
		Experiment: &core.ExperimentEntry{
			Name:       "projects/123/namespaces/firebase/experiments/exp-2",
			Definition: firebase.ExperimentDefinition{DisplayName: "Current"},
		},
	})

	m.applyManagedFeatureDetailsLoaded(managedFeatureDetailsLoadedMsg{
		kind: messages.ManagedFeatureExperiment, projectID: "demo", id: "exp-1",
		experiment: firebase.Experiment{
			Name:       "projects/123/namespaces/firebase/experiments/exp-1",
			Definition: firebase.ExperimentDefinition{DisplayName: "Late"},
		},
	})
	if got := m.details.ManagedFeatureData().Experiment.Definition.DisplayName; got != "Current" {
		t.Fatalf("late refresh replaced current Details with %q", got)
	}
}

func TestExperimentDetailsLazyLoadEachSelectionOnceAndReuseMemoryCache(t *testing.T) {
	project := core.Project{
		Name: "Demo", ProjectID: "demo", ProjectNumber: "123", AuthID: "main",
		DiscoveredBy: []string{"main"},
	}
	var requests []string
	svc := managedFeatureAppTestService(t, project, func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.RequestURI())
		id := "exp-1"
		name := "First full"
		if strings.HasSuffix(req.URL.Path, "/exp-2") {
			id = "exp-2"
			name = "Second full"
		}
		return managedFeatureJSONResponse(`{
			"name":"projects/123/namespaces/firebase/experiments/` + id + `",
			"definition":{"displayName":"` + name + `","variants":[{"name":"Baseline","weight":1}]},
			"state":"RUNNING"
		}`), nil
	})
	m := newModel(svc, nil)
	m.setActive(panels.ABTests)

	first := experimentListViewData(project, "exp-1", "First list")
	cmd := m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: first, Activate: true})
	if cmd == nil {
		t.Fatal("opening the first experiment did not start its lazy load")
	}
	m.applyManagedFeatureDetailsLoaded(cmd().(managedFeatureDetailsLoadedMsg))
	if got := m.details.ManagedFeatureData().Experiment.Definition.Variants; len(got) != 1 {
		t.Fatalf("first experiment variants = %#v, want one lazy-loaded variant", got)
	}

	m.setActive(panels.ABTests)
	second := experimentListViewData(project, "exp-2", "Second list")
	cmd = m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: second})
	if cmd == nil {
		t.Fatal("selecting an uncached experiment while Details is visible did not start its lazy load")
	}
	if got := m.details.ManagedFeatureData().Experiment.Definition.Variants; len(got) != 0 {
		t.Fatalf("second experiment showed variants before lazy load: %#v", got)
	}
	m.applyManagedFeatureDetailsLoaded(cmd().(managedFeatureDetailsLoadedMsg))
	if got := m.details.ManagedFeatureData().Experiment.Definition.Variants; len(got) != 1 {
		t.Fatalf("second experiment variants = %#v, want one lazy-loaded variant", got)
	}

	cmd = m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: first})
	if cmd != nil {
		t.Fatal("returning to a memory-cached experiment started another request")
	}
	data := m.details.ManagedFeatureData()
	if data.Experiment.Definition.DisplayName != "First full" ||
		len(data.Experiment.Definition.Variants) != 1 {
		t.Fatalf("cached first experiment details = %#v", data.Experiment)
	}
	if len(requests) != 2 {
		t.Fatalf("experiment detail requests = %#v, want one request per experiment", requests)
	}
}

func TestRolloutDetailsLazyLoadEachSelectionOnceAndReuseMemoryCache(t *testing.T) {
	project := core.Project{
		Name: "Demo", ProjectID: "demo", ProjectNumber: "123", AuthID: "main",
		DiscoveredBy: []string{"main"},
	}
	var requests []string
	svc := managedFeatureAppTestService(t, project, func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.RequestURI())
		id := "rollout-1"
		name := "First full"
		if strings.HasSuffix(req.URL.Path, "/rollout-2") {
			id = "rollout-2"
			name = "Second full"
		}
		return managedFeatureJSONResponse(`{
			"name":"projects/123/namespaces/firebase/rollouts/` + id + `",
			"definition":{
				"displayName":"` + name + `",
				"controlVariant":{"name":"Control","weight":90},
				"enabledVariant":{"name":"Enabled","weight":10}
			},
			"state":"RUNNING"
		}`), nil
	})
	m := newModel(svc, nil)
	m.setActive(panels.Rollouts)

	first := rolloutListViewData(project, "rollout-1", "First list")
	cmd := m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: first, Activate: true})
	if cmd == nil {
		t.Fatal("opening the first rollout did not start its lazy load")
	}
	m.applyManagedFeatureDetailsLoaded(cmd().(managedFeatureDetailsLoadedMsg))
	if got := m.details.ManagedFeatureData().Rollout.Definition.EnabledVariant.Name; got != "Enabled" {
		t.Fatalf("first rollout enabled variant = %q, want lazy-loaded value", got)
	}

	m.setActive(panels.Rollouts)
	second := rolloutListViewData(project, "rollout-2", "Second list")
	cmd = m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: second})
	if cmd == nil {
		t.Fatal("selecting an uncached rollout while Details is visible did not start its lazy load")
	}
	if got := m.details.ManagedFeatureData().Rollout.Definition.EnabledVariant.Name; got != "" {
		t.Fatalf("second rollout showed enabled variant before lazy load: %q", got)
	}
	m.applyManagedFeatureDetailsLoaded(cmd().(managedFeatureDetailsLoadedMsg))
	if got := m.details.ManagedFeatureData().Rollout.Definition.EnabledVariant.Name; got != "Enabled" {
		t.Fatalf("second rollout enabled variant = %q, want lazy-loaded value", got)
	}

	cmd = m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: first})
	if cmd != nil {
		t.Fatal("returning to a memory-cached rollout started another request")
	}
	data := m.details.ManagedFeatureData()
	if data.Rollout.Definition.DisplayName != "First full" ||
		data.Rollout.Definition.EnabledVariant.Name != "Enabled" {
		t.Fatalf("cached first rollout details = %#v", data.Rollout)
	}
	if len(requests) != 2 {
		t.Fatalf("rollout detail requests = %#v, want one request per rollout", requests)
	}
}

func TestExperimentListReloadInvalidatesMemoryCacheAndLazyLoadsDetailsAgain(t *testing.T) {
	project := core.Project{
		Name: "Demo", ProjectID: "demo", ProjectNumber: "123", AuthID: "main",
		DiscoveredBy: []string{"main"},
	}
	requests := 0
	svc := managedFeatureAppTestService(t, project, func(_ *http.Request) (*http.Response, error) {
		requests++
		return managedFeatureJSONResponse(`{
			"name":"projects/123/namespaces/firebase/experiments/exp-1",
			"definition":{"displayName":"Full experiment","variants":[{"name":"Variant A","weight":1}]},
			"state":"RUNNING"
		}`), nil
	})
	m := newModel(svc, nil)
	first := experimentListViewData(project, "exp-1", "List experiment")
	cmd := m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: first, Activate: true})
	m.applyManagedFeatureDetailsLoaded(cmd().(managedFeatureDetailsLoadedMsg))

	m.abTests, _ = m.abTests.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.abTests, _ = m.abTests.Update(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureExperiment, Project: project, Experiments: []core.ExperimentEntry{*first.Experiment},
	})
	m.setActive(panels.ABTests)
	next, _, handled := m.updateGlobalKeyMessage("u")
	m = next
	if handled {
		t.Fatal("A/B Tests reload was consumed before reaching its panel")
	}
	if len(m.managedFeatureDetails) != 0 {
		t.Fatalf("u retained experiment detail cache: %#v", m.managedFeatureDetails)
	}

	cmd = m.refreshOpenManagedFeatureFromList(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureExperiment, Project: project, Experiments: []core.ExperimentEntry{*first.Experiment},
	})
	if cmd == nil {
		t.Fatal("reloaded experiment list did not restart the visible Details lazy load")
	}
	if got := m.details.ManagedFeatureData().Experiment.Definition.Variants; len(got) != 0 {
		t.Fatalf("reloaded list retained stale variants before lazy load: %#v", got)
	}
	m.applyManagedFeatureDetailsLoaded(cmd().(managedFeatureDetailsLoadedMsg))
	if got := m.details.ManagedFeatureData().Experiment.Definition.Variants; len(got) != 1 {
		t.Fatalf("variants after cache-invalidating reload = %#v", got)
	}
	if requests != 2 {
		t.Fatalf("detail requests after reload = %d, want 2", requests)
	}

	m.managedFeatureDetails[managedFeatureDetailsKey{
		kind: messages.ManagedFeatureExperiment, projectID: "other", id: "exp-2",
	}] = managedFeatureDetailsEntry{experiment: firebase.Experiment{}}
	next, _, handled = m.updateGlobalKeyMessage("U")
	m = next
	if handled || len(m.managedFeatureDetails) != 0 {
		t.Fatalf("U cache invalidation = handled %t, cache %#v", handled, m.managedFeatureDetails)
	}
}

func TestRolloutListReloadInvalidatesOnlyRolloutMemoryCacheAndLazyLoadsDetailsAgain(t *testing.T) {
	project := core.Project{
		Name: "Demo", ProjectID: "demo", ProjectNumber: "123", AuthID: "main",
		DiscoveredBy: []string{"main"},
	}
	requests := 0
	svc := managedFeatureAppTestService(t, project, func(_ *http.Request) (*http.Response, error) {
		requests++
		return managedFeatureJSONResponse(`{
			"name":"projects/123/namespaces/firebase/rollouts/rollout-1",
			"definition":{
				"displayName":"Full rollout",
				"controlVariant":{"name":"Control","weight":90},
				"enabledVariant":{"name":"Enabled","weight":10}
			},
			"state":"RUNNING"
		}`), nil
	})
	m := newModel(svc, nil)
	first := rolloutListViewData(project, "rollout-1", "List rollout")
	cmd := m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: first, Activate: true})
	m.applyManagedFeatureDetailsLoaded(cmd().(managedFeatureDetailsLoadedMsg))
	m.ensureManagedFeatureDetailsCache()
	otherRolloutKey := managedFeatureDetailsKey{
		kind: messages.ManagedFeatureRollout, projectID: "other", id: "rollout-2",
	}
	experimentKey := managedFeatureDetailsKey{
		kind: messages.ManagedFeatureExperiment, projectID: "demo", id: "exp-1",
	}
	m.managedFeatureDetails[otherRolloutKey] = managedFeatureDetailsEntry{rollout: firebase.Rollout{}}
	m.managedFeatureDetails[experimentKey] = managedFeatureDetailsEntry{experiment: firebase.Experiment{}}

	m.rollouts, _ = m.rollouts.Update(messages.ProjectsSelectionChangedMsg{Projects: []core.Project{project}})
	m.rollouts, _ = m.rollouts.Update(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureRollout, Project: project, Rollouts: []core.RolloutEntry{*first.Rollout},
	})
	m.setActive(panels.Rollouts)
	next, _, handled := m.updateGlobalKeyMessage("u")
	m = next
	if handled {
		t.Fatal("Rollouts reload was consumed before reaching its panel")
	}
	demoRolloutKey := managedFeatureDetailsKey{
		kind: messages.ManagedFeatureRollout, projectID: "demo", id: "rollout-1",
	}
	if _, ok := m.managedFeatureDetails[demoRolloutKey]; ok {
		t.Fatalf("u retained selected-project rollout detail cache: %#v", m.managedFeatureDetails)
	}
	if _, ok := m.managedFeatureDetails[otherRolloutKey]; !ok {
		t.Fatalf("u cleared another project's rollout cache: %#v", m.managedFeatureDetails)
	}
	if _, ok := m.managedFeatureDetails[experimentKey]; !ok {
		t.Fatalf("rollout u cleared experiment cache: %#v", m.managedFeatureDetails)
	}

	cmd = m.refreshOpenManagedFeatureFromList(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureRollout, Project: project, Rollouts: []core.RolloutEntry{*first.Rollout},
	})
	if cmd == nil {
		t.Fatal("reloaded rollout list did not restart the visible Details lazy load")
	}
	if got := m.details.ManagedFeatureData().Rollout.Definition.EnabledVariant.Name; got != "" {
		t.Fatalf("reloaded list retained stale enabled variant before lazy load: %q", got)
	}
	m.applyManagedFeatureDetailsLoaded(cmd().(managedFeatureDetailsLoadedMsg))
	if got := m.details.ManagedFeatureData().Rollout.Definition.EnabledVariant.Name; got != "Enabled" {
		t.Fatalf("enabled variant after cache-invalidating reload = %q", got)
	}
	if requests != 2 {
		t.Fatalf("rollout detail requests after reload = %d, want 2", requests)
	}

	next, _, handled = m.updateGlobalKeyMessage("U")
	m = next
	if handled {
		t.Fatal("Rollouts reload-all was consumed before reaching its panel")
	}
	for key := range m.managedFeatureDetails {
		if key.kind == messages.ManagedFeatureRollout {
			t.Fatalf("U retained rollout detail cache: %#v", m.managedFeatureDetails)
		}
	}
	if _, ok := m.managedFeatureDetails[experimentKey]; !ok {
		t.Fatalf("rollout U cleared experiment cache: %#v", m.managedFeatureDetails)
	}
}

func TestExperimentDetailsIgnoreResponseStartedBeforeListReload(t *testing.T) {
	m := newModel(nil, nil)
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	data := experimentListViewData(project, "exp-1", "List experiment")
	m.detailsVisible = true
	m.details = m.details.SetManagedFeatureData(data)
	m.ensureManagedFeatureDetailsCache()
	m.managedFeatureDetailLoads[managedFeatureDetailsKey{
		kind: messages.ManagedFeatureExperiment, projectID: "demo", id: "exp-1",
	}] = 0
	m.invalidateManagedFeatureDetails(messages.ManagedFeatureExperiment, "demo")

	m.applyManagedFeatureDetailsLoaded(managedFeatureDetailsLoadedMsg{
		kind: messages.ManagedFeatureExperiment, projectID: "demo", id: "exp-1",
		detailGeneration: 0,
		experiment: firebase.Experiment{
			Name: "projects/123/namespaces/firebase/experiments/exp-1",
			Definition: firebase.ExperimentDefinition{
				DisplayName: "Stale full experiment",
				Variants:    []firebase.ExperimentVariant{{Name: "Stale variant"}},
			},
		},
	})
	if len(m.managedFeatureDetails) != 0 {
		t.Fatalf("stale response repopulated cache: %#v", m.managedFeatureDetails)
	}
	if got := m.details.ManagedFeatureData().Experiment.Definition.DisplayName; got != "List experiment" {
		t.Fatalf("stale response replaced visible details with %q", got)
	}
}

func TestManagedFeatureSelectionDoesNotDiscardDirtyDetails(t *testing.T) {
	m := conditionalValueDetailsTestModel()
	m.details, _ = m.details.ActivateName()
	m.details, _ = m.details.Update(tea.KeyPressMsg(tea.Key{Code: 'x', Text: "x"}))
	m.details = m.details.DeactivateField()
	before := m.details.Data().Parameter.Key
	managed := &messages.ManagedFeatureViewData{
		Kind:    messages.ManagedFeatureExperiment,
		Project: core.Project{Name: "Demo", ProjectID: "demo"},
		Experiment: &core.ExperimentEntry{
			Name: "projects/demo/namespaces/firebase/experiments/exp-1"},
	}

	cmd := m.applyManagedFeatureSelection(messages.ManagedFeatureSelectionChangedMsg{Data: managed, Activate: true})
	if cmd != nil || !m.dialog.IsOpen() || m.details.Data() == nil || m.details.Data().Parameter.Key != before {
		t.Fatalf(
			"dirty Details selection = cmd:%v dialog:%v data:%#v",
			cmd != nil,
			m.dialog.IsOpen(),
			m.details.Data(),
		)
	}
	if m.pendingDetails == nil || m.pendingDetails.managedData == nil {
		t.Fatal("managed selection was not queued behind the unsaved Details decision")
	}

	m.dialog = m.dialog.Close()
	next, refresh, _ := m.updateDetailsEditCanceled(messages.DetailsEditCanceledMsg{})
	if refresh != nil || next.details.ManagedFeatureData() == nil || managedFeatureDataID(next.details.ManagedFeatureData()) != "exp-1" {
		t.Fatalf("discard did not apply queued managed selection: refresh=%v data=%#v", refresh != nil, next.details.ManagedFeatureData())
	}
}

func experimentListViewData(project core.Project, id, displayName string) *messages.ManagedFeatureViewData {
	return &messages.ManagedFeatureViewData{
		Kind: messages.ManagedFeatureExperiment, Project: project,
		Experiment: &core.ExperimentEntry{
			Name:       "projects/123/namespaces/firebase/experiments/" + id,
			Definition: firebase.ExperimentDefinition{DisplayName: displayName}},
	}
}

func rolloutListViewData(project core.Project, id, displayName string) *messages.ManagedFeatureViewData {
	return &messages.ManagedFeatureViewData{
		Kind: messages.ManagedFeatureRollout, Project: project,
		Rollout: &core.RolloutEntry{
			Name:       "projects/123/namespaces/firebase/rollouts/" + id,
			Definition: firebase.RolloutDefinition{DisplayName: displayName}},
	}
}

func TestOpenManagedFeatureDetailsRefreshWithReloadedList(t *testing.T) {
	project := core.Project{Name: "Demo", ProjectID: "demo"}
	m := newModel(nil, nil)
	m.detailsVisible = true
	m.details = m.details.SetManagedFeatureData(&messages.ManagedFeatureViewData{
		Kind: messages.ManagedFeatureExperiment, Project: project,
		Experiment: &core.ExperimentEntry{
			Name:       "projects/demo/namespaces/firebase/experiments/exp-1",
			Definition: firebase.ExperimentDefinition{DisplayName: "Old"}},
	})
	m.refreshOpenManagedFeatureFromList(messages.ManagedFeaturesLoadedMsg{
		Kind: messages.ManagedFeatureExperiment, Project: project,
		Template: core.ManagedFeatureTemplate{Version: "12", Source: "remote"},
		Experiments: []core.ExperimentEntry{{
			Name:       "projects/demo/namespaces/firebase/experiments/exp-1",
			Definition: firebase.ExperimentDefinition{DisplayName: "Fresh"},
			References: []core.ManagedValueReference{{Parameter: "signup_message"}},
		}},
	})
	data := m.details.ManagedFeatureData()
	if data.Experiment.Definition.DisplayName != "Fresh" ||
		len(data.Experiment.References) != 1 ||
		data.Template.Version != "12" {
		t.Fatalf("reloaded managed Details = %#v", data)
	}
}

type managedFeatureAppRoundTripFunc func(*http.Request) (*http.Response, error)

func (f managedFeatureAppRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func managedFeatureAppTestService(
	t *testing.T,
	project core.Project,
	roundTrip managedFeatureAppRoundTripFunc,
) *core.Core {
	t.Helper()
	root := t.TempDir()
	t.Setenv(env.ConfigDir, filepath.Join(root, "config"))
	t.Setenv(env.CacheDir, filepath.Join(root, "cache"))
	if err := coreconfig.SwitchProfile(coreconfig.DefaultProfileName); err != nil {
		t.Fatal(err)
	}
	svc, err := core.NewService(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddGCloudAuth("main", "main"); err != nil {
		t.Fatal(err)
	}
	if err := coreconfig.SaveProjects([]coreconfig.Project{project}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	svc.InjectFirebaseService("main", firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTrip}))
	return svc
}

func managedFeatureJSONResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
