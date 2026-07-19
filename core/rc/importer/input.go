package importer

import (
	"encoding/json"
	"fmt"

	"github.com/yumauri/fbrcm/core/firebase"
)

func ParseSource(raw []byte) (*ParsedSource, error) {
	if !json.Valid(raw) {
		return nil, fmt.Errorf("remote config input is not valid json")
	}
	remoteRaw, wrapped, err := ExtractRemoteConfigJSON(raw)
	if err != nil {
		return nil, err
	}
	cfg, err := firebase.ParseCloneRemoteConfig(remoteRaw)
	if err != nil {
		return nil, fmt.Errorf("decode remote config: %w", err)
	}
	if err := firebase.NormalizeRemoteConfigForUpdate(cfg); err != nil {
		return nil, fmt.Errorf("normalize remote config input: %w", err)
	}
	return &ParsedSource{Config: cfg, Raw: remoteRaw, WrappedCache: wrapped}, nil
}

func ExtractRemoteConfigJSON(raw []byte) ([]byte, bool, error) {
	var payload struct {
		RemoteConfig json.RawMessage `json:"remote_config"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, false, fmt.Errorf("decode input: %w", err)
	}
	if len(payload.RemoteConfig) == 0 {
		return raw, false, nil
	}
	if !json.Valid(payload.RemoteConfig) {
		return nil, true, fmt.Errorf("remote_config is not valid json")
	}
	return payload.RemoteConfig, true, nil
}

func Summarize(cfg *firebase.RemoteConfig, wrapped bool) Summary {
	if cfg == nil {
		return Summary{WrappedCache: wrapped}
	}
	summary := Summary{
		RootParameters: len(cfg.Parameters),
		Groups:         len(cfg.ParameterGroups),
		Conditions:     len(cfg.Conditions),
		WrappedCache:   wrapped,
	}
	for _, condition := range cfg.Conditions {
		if isNonPortableCondition(condition.Expression) {
			summary.NonPortableConditions++
		}
	}
	for _, group := range cfg.ParameterGroups {
		summary.GroupParameters += len(group.Parameters)
	}
	return summary
}
