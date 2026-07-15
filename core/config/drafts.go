package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
)

const DraftFormatVersion = 1

// Draft is the complete, self-contained on-disk draft representation.
// BaseRemoteConfig is immutable for the lifetime of a draft unless the draft is
// explicitly rebased onto a newer Remote Config.
type Draft struct {
	FormatVersion    int             `json:"format_version"`
	ProjectID        string          `json:"project_id"`
	BaseVersion      string          `json:"base_version"`
	BaseETag         string          `json:"base_etag"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	BaseRemoteConfig json.RawMessage `json:"base_remote_config"`
	RemoteConfig     json.RawMessage `json:"remote_config"`
}

func GetDraftsDirPath() string {
	return filepath.Join(GetCacheDirPath(), "drafts")
}

func GetDraftPath(projectID string) string {
	return filepath.Join(GetDraftsDirPath(), projectID+".json")
}

func LoadDraft(projectID string) (*Draft, error) {
	path := GetDraftPath(projectID)
	logger := corelog.For("config")
	logger.Debug("read draft", "project_id", projectID, "path", path)

	var stored Draft
	if err := readJSONFile(path, &stored); err != nil {
		if isNotExist(err) {
			logger.Debug("draft miss", "project_id", projectID, "path", path)
		} else {
			logger.Error("read draft failed", "project_id", projectID, "path", path, "err", err)
		}
		return nil, fmt.Errorf("read draft: %w", err)
	}
	if err := validateDraft(projectID, &stored); err != nil {
		logger.Error("invalid draft", "project_id", projectID, "path", path, "err", err)
		return nil, err
	}

	logger.Info("loaded draft", "project_id", projectID, "path", path, "bytes", len(stored.RemoteConfig))
	return &stored, nil
}

func SaveDraft(stored *Draft) error {
	if stored == nil {
		return fmt.Errorf("draft is nil")
	}
	stored.ProjectID = strings.TrimSpace(stored.ProjectID)
	if stored.FormatVersion == 0 {
		stored.FormatVersion = DraftFormatVersion
	}
	if err := validateDraft(stored.ProjectID, stored); err != nil {
		return err
	}
	if err := EnsurePrivateDir(GetDraftsDirPath()); err != nil {
		return fmt.Errorf("create drafts dir: %w", err)
	}

	path := GetDraftPath(stored.ProjectID)
	logger := corelog.For("config")
	logger.Debug("write draft", "project_id", stored.ProjectID, "path", path)
	if err := writeJSONFile(path, stored); err != nil {
		logger.Error("write draft failed", "project_id", stored.ProjectID, "path", path, "err", err)
		return fmt.Errorf("write draft: %w", err)
	}

	logger.Info("saved draft", "project_id", stored.ProjectID, "path", path, "bytes", len(stored.RemoteConfig))
	return nil
}

func validateDraft(projectID string, stored *Draft) error {
	if stored == nil {
		return fmt.Errorf("draft is nil")
	}
	if strings.TrimSpace(projectID) == "" {
		return fmt.Errorf("draft project id is empty")
	}
	if stored.FormatVersion != DraftFormatVersion {
		return fmt.Errorf("unsupported draft format version %d", stored.FormatVersion)
	}
	if stored.ProjectID != projectID {
		return fmt.Errorf("draft project id %q does not match %q", stored.ProjectID, projectID)
	}
	if !json.Valid(stored.BaseRemoteConfig) {
		return fmt.Errorf("draft base remote config is invalid json")
	}
	if !json.Valid(stored.RemoteConfig) {
		return fmt.Errorf("draft remote config is invalid json")
	}
	if stored.CreatedAt.IsZero() || stored.UpdatedAt.IsZero() {
		return fmt.Errorf("draft timestamps are missing")
	}
	return nil
}

func DeleteDraft(projectID string) error {
	path := GetDraftPath(projectID)
	logger := corelog.For("config")
	logger.Debug("remove draft", "project_id", projectID, "path", path)
	if err := os.Remove(path); err != nil && !isNotExist(err) {
		logger.Error("remove draft failed", "project_id", projectID, "path", path, "err", err)
		return fmt.Errorf("remove draft: %w", err)
	}
	logger.Info("draft removed", "project_id", projectID, "path", path)
	return nil
}

func ListDraftProjectIDs() ([]string, error) {
	path := GetDraftsDirPath()
	logger := corelog.For("config")
	logger.Debug("list drafts", "path", path)

	entries, err := os.ReadDir(path)
	if err != nil {
		if isNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read drafts dir: %w", err)
	}

	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(name, ".json"))
	}
	strfold.Sort(ids)
	return ids, nil
}

func PurgeDrafts() error {
	path := GetDraftsDirPath()
	logger := corelog.For("config")
	logger.Debug("remove drafts dir", "path", path)
	if err := os.RemoveAll(path); err != nil {
		logger.Error("remove drafts dir failed", "path", path, "err", err)
		return fmt.Errorf("remove drafts dir: %w", err)
	}

	logger.Info("drafts dir removed", "path", path)
	return nil
}
