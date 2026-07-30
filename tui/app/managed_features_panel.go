package app

import (
	"github.com/yumauri/fbrcm/tui/components/managedfeatures"
	"github.com/yumauri/fbrcm/tui/panels"
)

func (m *Model) activeManagedFeaturesPanel() *managedfeatures.Model {
	switch m.active {
	case panels.ABTests:
		return &m.abTests
	case panels.Personalizations:
		return &m.personalizations
	case panels.Rollouts:
		return &m.rollouts
	default:
		return nil
	}
}
