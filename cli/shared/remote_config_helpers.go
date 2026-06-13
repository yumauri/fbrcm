package shared

import (
	"encoding/json"
	"fmt"

	"github.com/yumauri/fbrcm/core/firebase"
)

// CloneRemoteConfig deep-copies a Remote Config value.
func CloneRemoteConfig(cfg *firebase.RemoteConfig) *firebase.RemoteConfig {
	if cfg == nil {
		return &firebase.RemoteConfig{}
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return &firebase.RemoteConfig{}
	}
	var out firebase.RemoteConfig
	if err := json.Unmarshal(data, &out); err != nil {
		return &firebase.RemoteConfig{}
	}
	return &out
}

// MarshalRemoteConfig encodes Remote Config JSON for upload or transform output.
func MarshalRemoteConfig(cfg *firebase.RemoteConfig) ([]byte, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode remote config: %w", err)
	}
	return data, nil
}
