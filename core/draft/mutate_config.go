package draft

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/yumauri/fbrcm/core/firebase"
	rcmutate "github.com/yumauri/fbrcm/core/rc/mutate"
)

func BuildMutatedRemoteConfig(currentRaw json.RawMessage, unchangedErr string, mutate Mutation) (json.RawMessage, error) {
	currentCfg, err := firebase.ParseRemoteConfig(currentRaw)
	if err != nil {
		return nil, fmt.Errorf("decode current remote config: %w", err)
	}
	finalCfg, err := firebase.CloneRemoteConfig(currentCfg)
	if err != nil {
		return nil, fmt.Errorf("clone current remote config: %w", err)
	}
	if err := mutate(finalCfg); err != nil {
		return nil, err
	}
	rcmutate.DropUnknownConditionReferences(finalCfg)
	rcmutate.NormalizeEmptyParameterMaps(finalCfg)

	if unchangedErr != "" && reflect.DeepEqual(currentCfg, finalCfg) {
		return nil, fmt.Errorf("%s", unchangedErr)
	}

	finalRaw, err := firebase.MarshalRemoteConfig(finalCfg)
	if err != nil {
		return nil, err
	}
	return finalRaw, nil
}

func currentDraftRaw(projectID string, remoteRaw json.RawMessage) (json.RawMessage, bool, error) {
	draftRaw, hasDraft, err := Load(projectID)
	if err != nil {
		return nil, false, err
	}
	if hasDraft {
		return draftRaw, true, nil
	}
	return remoteRaw, false, nil
}
