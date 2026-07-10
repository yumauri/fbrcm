package app

import (
	"encoding/json"

	"github.com/yumauri/fbrcm/core/firebase"
)

func parseRemoteConfigPair(currentRaw, finalRaw json.RawMessage) (*firebase.RemoteConfig, *firebase.RemoteConfig, error) {
	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, nil, err
	}
	finalCfg, err := firebase.ParseRemoteConfig(finalRaw)
	if err != nil {
		return nil, nil, err
	}
	return currentCfg, finalCfg, nil
}
