package rc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

// ParseProjectRemoteConfig parses Remote Config JSON with a project-scoped decode error.
func ParseProjectRemoteConfig(projectID string, raw json.RawMessage) (*firebase.RemoteConfig, error) {
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return nil, fmt.Errorf("decode remote config for %s: %w", projectID, err)
	}
	return cfg, nil
}

// ParseCachedProjectRemoteConfig parses cached Remote Config JSON with a project-scoped decode error.
func ParseCachedProjectRemoteConfig(projectID string, raw json.RawMessage) (*firebase.RemoteConfig, error) {
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		return nil, fmt.Errorf("decode cached remote config for %s: %w", projectID, err)
	}
	return cfg, nil
}

// ProjectConfig is a revalidated Remote Config for one project.
type ProjectConfig struct {
	Project core.Project
	Cache   *core.ParametersCache
	Config  *firebase.RemoteConfig
}

// NewProjectConfig parses and clones a project's cached Remote Config.
func NewProjectConfig(project core.Project, cache *core.ParametersCache) (*ProjectConfig, error) {
	cloned, err := firebase.ParseCloneRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, fmt.Errorf("decode remote config for %s: %w", project.ProjectID, err)
	}
	return &ProjectConfig{
		Project: project,
		Cache:   cache,
		Config:  cloned,
	}, nil
}

// RevalidateProjectConfig reloads a project's Remote Config and parses it for command use.
func RevalidateProjectConfig(ctx context.Context, svc *core.Core, project core.Project) (*ProjectConfig, error) {
	cache, _, err := svc.RevalidateParameters(ctx, project.ProjectID)
	if err != nil {
		return nil, err
	}
	return NewProjectConfig(project, cache)
}
