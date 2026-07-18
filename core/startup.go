package core

import (
	"errors"
	"os"

	"github.com/yumauri/fbrcm/core/config"
)

// StartupState describes the local state needed to choose between opening the
// workspace and showing interactive setup. It deliberately performs no
// network requests and never starts an authentication flow.
type StartupState struct {
	Profile       string
	Profiles      []string
	Auth          []config.AuthEntry
	DefaultAuthID string
	Projects      []Project
}

// InspectStartupState reads the active profile's auth registry and projects
// cache without attempting project discovery.
func (s *Core) InspectStartupState() (StartupState, error) {
	profiles, err := config.ListProfiles()
	if err != nil {
		return StartupState{}, err
	}
	auth, defaultAuthID, err := s.ListAuth()
	if err != nil {
		return StartupState{}, err
	}

	projects, err := config.LoadProjects()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, config.ErrEmptyProjectsFile) {
			return StartupState{}, err
		}
		projects = nil
	}

	return StartupState{
		Profile:       config.GetActiveProfileName(),
		Profiles:      profiles,
		Auth:          auth,
		DefaultAuthID: defaultAuthID,
		Projects:      append([]Project(nil), projects...),
	}, nil
}

// SwitchProfile selects or creates a profile for this Core service. Firebase
// clients are cached with the active profile in their key, so identities with
// the same name in different profiles remain isolated.
func (s *Core) SwitchProfile(name string) error {
	return config.SwitchProfile(name)
}
