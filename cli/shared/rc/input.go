package rc

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/importer"
)

// ExtractRemoteConfigJSON extracts Firebase Remote Config JSON from raw export or fbrcm cache JSON.
func ExtractRemoteConfigJSON(raw []byte) ([]byte, error) {
	remoteRaw, _, err := importer.ExtractRemoteConfigJSON(raw)
	return remoteRaw, err
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

	cloned, err := firebase.ParseCloneRemoteConfig(remoteConfigRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode stdin remote config: %w", err)
	}
	return cloned, remoteConfigRaw, nil
}
