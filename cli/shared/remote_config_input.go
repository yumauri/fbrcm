package shared

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/yumauri/fbrcm/core/firebase"
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

func ReadRemoteConfigInput(in io.Reader) (*firebase.RemoteConfig, []byte, error) {
	raw, err := io.ReadAll(in)
	if err != nil {
		return nil, nil, fmt.Errorf("read stdin: %w", err)
	}
	if !json.Valid(raw) {
		return nil, nil, fmt.Errorf("stdin remote config is not valid json")
	}

	remoteConfigRaw, err := ExtractRemoteConfigJSON(raw)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := firebase.ParseRemoteConfig(remoteConfigRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode stdin remote config: %w", err)
	}
	return CloneRemoteConfig(cfg), remoteConfigRaw, nil
}
