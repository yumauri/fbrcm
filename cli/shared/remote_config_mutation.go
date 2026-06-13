package shared

import (
	"context"
	"fmt"
	"io"

	"github.com/yumauri/fbrcm/core/firebase"
)

// RemoteConfigMutation applies a command-specific change to a cloned config.
type RemoteConfigMutation func(current *firebase.RemoteConfig) (changedCount int, finalCfg *firebase.RemoteConfig, err error)

// PublishProjectConfigMutation applies mutation and publishes the result, returning whether callers should retry.
func PublishProjectConfigMutation(ctx context.Context, publisher RemoteConfigPublisher, projectCfg *ProjectConfig, operation string, errOut io.Writer, mutate RemoteConfigMutation) (int, bool, error) {
	if projectCfg == nil || projectCfg.Cache == nil {
		return 0, false, fmt.Errorf("project config is incomplete")
	}

	changedCount, finalCfg, err := mutate(projectCfg.Config)
	if err != nil {
		return 0, false, err
	}
	if changedCount == 0 {
		return 0, false, nil
	}

	finalRaw, err := MarshalRemoteConfig(finalCfg)
	if err != nil {
		return 0, false, err
	}
	retry, err := ValidateAndPublishRemoteConfig(ctx, publisher, projectCfg.Project.ProjectID, finalRaw, projectCfg.Cache.ETag, operation, errOut)
	if err != nil {
		return 0, false, err
	}
	return changedCount, retry, nil
}
