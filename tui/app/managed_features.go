package app

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/tui/messages"
)

type managedFeatureDetailsLoadedMsg struct {
	kind             messages.ManagedFeatureKind
	projectID        string
	id               string
	detailGeneration uint64
	experiment       firebase.Experiment
	rollout          firebase.Rollout
	err              error
}

type managedFeatureDetailsKey struct {
	kind      messages.ManagedFeatureKind
	projectID string
	id        string
}

type managedFeatureDetailsScope struct {
	kind      messages.ManagedFeatureKind
	projectID string
}

type managedFeatureDetailsEntry struct {
	experiment firebase.Experiment
	rollout    firebase.Rollout
}

func managedFeatureDataID(data *messages.ManagedFeatureViewData) string {
	if data == nil {
		return ""
	}
	switch data.Kind {
	case messages.ManagedFeatureExperiment:
		if data.Experiment != nil {
			return firebase.ManagedFeatureID(data.Experiment.Name)
		}
	case messages.ManagedFeaturePersonalization:
		if data.Personalization != nil {
			return data.Personalization.ID
		}
	case messages.ManagedFeatureRollout:
		if data.Rollout != nil {
			return firebase.ManagedFeatureID(data.Rollout.Name)
		}
	}
	return ""
}

func (m *Model) refreshManagedFeatureDetailsCmd(data *messages.ManagedFeatureViewData) tea.Cmd {
	if m.svc == nil || data == nil {
		return nil
	}
	project := data.Project
	id := managedFeatureDataID(data)
	if id == "" {
		return nil
	}
	switch data.Kind {
	case messages.ManagedFeatureExperiment:
		generation, ok := m.beginManagedFeatureDetailsLoad(data.Kind, project.ProjectID, id)
		if !ok {
			return nil
		}
		return func() tea.Msg {
			experiment, err := m.svc.GetRemoteConfigExperimentMetadata(context.Background(), project, id)
			return managedFeatureDetailsLoadedMsg{
				kind: messages.ManagedFeatureExperiment, projectID: project.ProjectID, id: id,
				detailGeneration: generation, experiment: experiment, err: err,
			}
		}
	case messages.ManagedFeatureRollout:
		generation, ok := m.beginManagedFeatureDetailsLoad(data.Kind, project.ProjectID, id)
		if !ok {
			return nil
		}
		return func() tea.Msg {
			rollout, err := m.svc.GetRemoteConfigRolloutMetadata(context.Background(), project, id)
			return managedFeatureDetailsLoadedMsg{
				kind: messages.ManagedFeatureRollout, projectID: project.ProjectID, id: id,
				detailGeneration: generation, rollout: rollout, err: err,
			}
		}
	default:
		return nil
	}
}

func (m *Model) applyManagedFeatureDetailsLoaded(msg managedFeatureDetailsLoadedMsg) {
	if managedFeatureHasLazyDetails(msg.kind) {
		key := managedFeatureDetailsKey{kind: msg.kind, projectID: msg.projectID, id: msg.id}
		scope := managedFeatureDetailsScope{kind: msg.kind, projectID: msg.projectID}
		m.ensureManagedFeatureDetailsCache()
		if msg.detailGeneration != m.managedFeatureDetailGenerations[scope] {
			return
		}
		if loadingGeneration, ok := m.managedFeatureDetailLoads[key]; ok &&
			loadingGeneration == msg.detailGeneration {
			delete(m.managedFeatureDetailLoads, key)
		}
		if msg.err == nil {
			m.managedFeatureDetails[key] = managedFeatureDetailsEntry{
				experiment: msg.experiment,
				rollout:    msg.rollout,
			}
		}
	}
	if msg.err != nil {
		corelog.For("tui.managed_features").Error(
			"managed feature details refresh failed",
			"kind", msg.kind,
			"project_id", msg.projectID,
			"id", msg.id,
			"err", msg.err,
		)
		return
	}
	current := m.details.ManagedFeatureData()
	if !m.detailsVisible ||
		current == nil ||
		current.Kind != msg.kind ||
		current.Project.ProjectID != msg.projectID ||
		managedFeatureDataID(current) != msg.id {
		return
	}
	next := *current
	switch msg.kind {
	case messages.ManagedFeatureExperiment:
		if current.Experiment == nil {
			return
		}
		experiment := *current.Experiment
		experiment.Experiment = msg.experiment
		next.Experiment = &experiment
	case messages.ManagedFeatureRollout:
		if current.Rollout == nil {
			return
		}
		rollout := *current.Rollout
		rollout.Rollout = msg.rollout
		next.Rollout = &rollout
	default:
		return
	}
	m.details = m.details.RefreshManagedFeatureData(&next)
}

func (m *Model) refreshOpenManagedFeatureFromList(msg messages.ManagedFeaturesLoadedMsg) tea.Cmd {
	if msg.Err != nil || !m.detailsVisible {
		return nil
	}
	current := m.details.ManagedFeatureData()
	if current == nil || current.Kind != msg.Kind || current.Project.ProjectID != msg.Project.ProjectID {
		return nil
	}
	id := managedFeatureDataID(current)
	next := *current
	next.Template = msg.Template
	switch msg.Kind {
	case messages.ManagedFeatureExperiment:
		for index := range msg.Experiments {
			if firebase.ManagedFeatureID(msg.Experiments[index].Name) == id {
				next.Experiment = &msg.Experiments[index]
				data := m.withCachedManagedFeatureDetails(&next)
				m.details = m.details.RefreshManagedFeatureData(data)
				return m.refreshManagedFeatureDetailsCmd(data)
			}
		}
	case messages.ManagedFeaturePersonalization:
		for index := range msg.Personalizations {
			if msg.Personalizations[index].ID == id {
				next.Personalization = &msg.Personalizations[index]
				m.details = m.details.RefreshManagedFeatureData(&next)
				return nil
			}
		}
	case messages.ManagedFeatureRollout:
		for index := range msg.Rollouts {
			if firebase.ManagedFeatureID(msg.Rollouts[index].Name) == id {
				next.Rollout = &msg.Rollouts[index]
				data := m.withCachedManagedFeatureDetails(&next)
				m.details = m.details.RefreshManagedFeatureData(data)
				return m.refreshManagedFeatureDetailsCmd(data)
			}
		}
	}
	return nil
}

func (m *Model) withCachedManagedFeatureDetails(data *messages.ManagedFeatureViewData) *messages.ManagedFeatureViewData {
	if data == nil || !managedFeatureHasLazyDetails(data.Kind) {
		return data
	}
	id := managedFeatureDataID(data)
	key := managedFeatureDetailsKey{kind: data.Kind, projectID: data.Project.ProjectID, id: id}
	entry, ok := m.managedFeatureDetails[key]
	if !ok {
		return data
	}
	next := *data
	switch data.Kind {
	case messages.ManagedFeatureExperiment:
		if data.Experiment == nil {
			return data
		}
		experiment := *data.Experiment
		experiment.Experiment = entry.experiment
		next.Experiment = &experiment
	case messages.ManagedFeatureRollout:
		if data.Rollout == nil {
			return data
		}
		rollout := *data.Rollout
		rollout.Rollout = entry.rollout
		next.Rollout = &rollout
	}
	return &next
}

func (m *Model) beginManagedFeatureDetailsLoad(
	kind messages.ManagedFeatureKind,
	projectID string,
	id string,
) (uint64, bool) {
	m.ensureManagedFeatureDetailsCache()
	key := managedFeatureDetailsKey{kind: kind, projectID: projectID, id: id}
	if _, ok := m.managedFeatureDetails[key]; ok {
		return 0, false
	}
	scope := managedFeatureDetailsScope{kind: kind, projectID: projectID}
	generation := m.managedFeatureDetailGenerations[scope]
	if loadingGeneration, ok := m.managedFeatureDetailLoads[key]; ok && loadingGeneration == generation {
		return 0, false
	}
	m.managedFeatureDetailLoads[key] = generation
	return generation, true
}

func (m *Model) invalidateManagedFeatureDetails(kind messages.ManagedFeatureKind, projectID string) {
	m.ensureManagedFeatureDetailsCache()
	scope := managedFeatureDetailsScope{kind: kind, projectID: projectID}
	m.managedFeatureDetailGenerations[scope]++
	for key := range m.managedFeatureDetails {
		if key.kind == kind && key.projectID == projectID {
			delete(m.managedFeatureDetails, key)
		}
	}
	for key := range m.managedFeatureDetailLoads {
		if key.kind == kind && key.projectID == projectID {
			delete(m.managedFeatureDetailLoads, key)
		}
	}
}

func (m *Model) invalidateAllManagedFeatureDetails(kind messages.ManagedFeatureKind) {
	m.ensureManagedFeatureDetailsCache()
	projects := make(map[string]struct{})
	for key := range m.managedFeatureDetails {
		if key.kind == kind {
			projects[key.projectID] = struct{}{}
		}
	}
	for key := range m.managedFeatureDetailLoads {
		if key.kind == kind {
			projects[key.projectID] = struct{}{}
		}
	}
	for projectID := range projects {
		scope := managedFeatureDetailsScope{kind: kind, projectID: projectID}
		m.managedFeatureDetailGenerations[scope]++
	}
	for key := range m.managedFeatureDetails {
		if key.kind == kind {
			delete(m.managedFeatureDetails, key)
		}
	}
	for key := range m.managedFeatureDetailLoads {
		if key.kind == kind {
			delete(m.managedFeatureDetailLoads, key)
		}
	}
}

func (m *Model) ensureManagedFeatureDetailsCache() {
	if m.managedFeatureDetails == nil {
		m.managedFeatureDetails = make(map[managedFeatureDetailsKey]managedFeatureDetailsEntry)
	}
	if m.managedFeatureDetailLoads == nil {
		m.managedFeatureDetailLoads = make(map[managedFeatureDetailsKey]uint64)
	}
	if m.managedFeatureDetailGenerations == nil {
		m.managedFeatureDetailGenerations = make(map[managedFeatureDetailsScope]uint64)
	}
}

func managedFeatureHasLazyDetails(kind messages.ManagedFeatureKind) bool {
	return kind == messages.ManagedFeatureExperiment || kind == messages.ManagedFeatureRollout
}
