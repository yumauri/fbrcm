package shared

import (
	"context"
	"fmt"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

// ProjectConfig is a revalidated Remote Config for one project.
type ProjectConfig struct {
	Project core.Project
	Cache   *core.ParametersCache
	Config  *firebase.RemoteConfig
}

// NewProjectConfig parses and clones a project's cached Remote Config.
func NewProjectConfig(project core.Project, cache *core.ParametersCache) (*ProjectConfig, error) {
	cfg, err := firebase.ParseRemoteConfig(cache.RemoteConfig)
	if err != nil {
		return nil, fmt.Errorf("decode remote config for %s: %w", project.ProjectID, err)
	}
	return &ProjectConfig{
		Project: project,
		Cache:   cache,
		Config:  CloneRemoteConfig(cfg),
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
