package shared

import (
	"encoding/json"
	"fmt"
)

// ExtractRemoteConfigJSON extracts Firebase Remote Config JSON from raw export or fbrcm cache JSON.
func ExtractRemoteConfigJSON(raw []byte) ([]byte, error) {
	var payload struct {
		RemoteConfig json.RawMessage `json:"remote_config"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("decode input: %w", err)
	}
	if len(payload.RemoteConfig) == 0 {
		return raw, nil
	}
	if !json.Valid(payload.RemoteConfig) {
		return nil, fmt.Errorf("remote_config is not valid json")
	}
	return payload.RemoteConfig, nil
}
