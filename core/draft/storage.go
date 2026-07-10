package draft

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func ListProjectIDs() ([]string, error) {
	return config.ListDraftProjectIDs()
}

func Load(projectID string) (json.RawMessage, bool, error) {
	raw, err := config.LoadDraft(projectID)
	if err != nil {
		var pathErr *os.PathError
		if errors.Is(err, os.ErrNotExist) || (errors.As(err, &pathErr) && errors.Is(pathErr.Err, os.ErrNotExist)) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		return nil, false, fmt.Errorf("decode draft: %w", err)
	}
	return raw, true, nil
}

func Save(projectID string, raw json.RawMessage) error {
	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		return fmt.Errorf("decode draft: %w", err)
	}
	return config.SaveDraft(projectID, raw)
}

func Delete(projectID string) error {
	return config.DeleteDraft(projectID)
}
