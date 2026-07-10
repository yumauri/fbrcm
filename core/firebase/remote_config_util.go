package firebase

import (
	"encoding/json"
	"fmt"
)

// ParseCloneRemoteConfig parses Remote Config JSON and returns a deep copy safe for mutation.
func ParseCloneRemoteConfig(raw json.RawMessage) (*RemoteConfig, error) {
	cfg, err := ParseRemoteConfig(raw)
	if err != nil {
		return nil, err
	}
	return CloneRemoteConfig(cfg)
}

// CloneRemoteConfig deep-copies a Remote Config value.
func CloneRemoteConfig(cfg *RemoteConfig) (*RemoteConfig, error) {
	if cfg == nil {
		return &RemoteConfig{}, nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode remote config for clone: %w", err)
	}
	var out RemoteConfig
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode remote config for clone: %w", err)
	}
	return &out, nil
}

// MarshalRemoteConfig encodes Remote Config JSON for upload or transform output.
func MarshalRemoteConfig(cfg *RemoteConfig) ([]byte, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("encode remote config: %w", err)
	}
	return data, nil
}
