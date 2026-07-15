package draft

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

type Record = config.Draft

func ListProjectIDs() ([]string, error) {
	return config.ListDraftProjectIDs()
}

func LoadRecord(projectID string) (*Record, bool, error) {
	stored, err := config.LoadDraft(projectID)
	if err != nil {
		var pathErr *os.PathError
		if errors.Is(err, os.ErrNotExist) || (errors.As(err, &pathErr) && errors.Is(pathErr.Err, os.ErrNotExist)) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if _, err := firebase.ParseRemoteConfig(stored.BaseRemoteConfig); err != nil {
		return nil, false, fmt.Errorf("decode draft base: %w", err)
	}
	if _, err := firebase.ParseRemoteConfig(stored.RemoteConfig); err != nil {
		return nil, false, fmt.Errorf("decode draft: %w", err)
	}
	return stored, true, nil
}

func Load(projectID string) (json.RawMessage, bool, error) {
	stored, ok, err := LoadRecord(projectID)
	if err != nil || !ok {
		return nil, ok, err
	}
	return append(json.RawMessage(nil), stored.RemoteConfig...), true, nil
}

// Save stores raw as a draft using the current parameters cache as its base.
// Existing drafts retain their original base.
func Save(projectID string, raw json.RawMessage) error {
	if existing, ok, err := LoadRecord(projectID); err != nil {
		return err
	} else if ok {
		return saveRecord(existing, raw, existing.BaseRemoteConfig, existing.BaseETag, existing.BaseVersion, existing.CreatedAt)
	}
	cache, err := config.LoadParametersCache(projectID)
	if err != nil {
		return fmt.Errorf("load draft base: %w", err)
	}
	return SaveWithBase(projectID, cache, raw)
}

func SaveWithBase(projectID string, base *config.ParametersCache, raw json.RawMessage) error {
	if base == nil {
		return fmt.Errorf("draft base is nil")
	}
	createdAt := time.Now().UTC()
	if existing, ok, err := LoadRecord(projectID); err != nil {
		return err
	} else if ok {
		return saveRecord(existing, raw, existing.BaseRemoteConfig, existing.BaseETag, existing.BaseVersion, existing.CreatedAt)
	}
	baseCfg, err := firebase.ParseRemoteConfig(base.RemoteConfig)
	if err != nil {
		return fmt.Errorf("decode draft base: %w", err)
	}
	return saveRecord(&Record{ProjectID: projectID}, raw, base.RemoteConfig, base.ETag, baseCfg.Version.VersionNumber, createdAt)
}

func SaveRebased(projectID string, base *config.ParametersCache, raw json.RawMessage, createdAt time.Time) error {
	if base == nil {
		return fmt.Errorf("draft base is nil")
	}
	baseCfg, err := firebase.ParseRemoteConfig(base.RemoteConfig)
	if err != nil {
		return fmt.Errorf("decode draft base: %w", err)
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	return saveRecord(&Record{ProjectID: projectID}, raw, base.RemoteConfig, base.ETag, baseCfg.Version.VersionNumber, createdAt)
}

func saveRecord(stored *Record, raw, baseRaw json.RawMessage, baseETag, baseVersion string, createdAt time.Time) error {
	if _, err := firebase.ParseRemoteConfig(baseRaw); err != nil {
		return fmt.Errorf("decode draft base: %w", err)
	}
	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		return fmt.Errorf("decode draft: %w", err)
	}
	now := time.Now().UTC()
	stored.FormatVersion = config.DraftFormatVersion
	stored.BaseVersion = baseVersion
	stored.BaseETag = baseETag
	stored.CreatedAt = createdAt
	stored.UpdatedAt = now
	stored.BaseRemoteConfig = append(json.RawMessage(nil), baseRaw...)
	stored.RemoteConfig = append(json.RawMessage(nil), raw...)
	return config.SaveDraft(stored)
}

func Delete(projectID string) error {
	return config.DeleteDraft(projectID)
}
