package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	corelog "fbrcm/core/log"
)

// GetDraftsDirPath gets drafts dir path and returns the resulting value or error.
func GetDraftsDirPath() string {
	return filepath.Join(GetCacheDirPath(), "drafts")
}

// GetDraftPath gets draft path and returns the resulting value or error.
func GetDraftPath(projectID string) string {
	return filepath.Join(GetDraftsDirPath(), projectID+".json")
}

// LoadDraft loads draft and returns the resulting value or error.
func LoadDraft(projectID string) (json.RawMessage, error) {
	path := GetDraftPath(projectID)
	logger := corelog.For("config")
	logger.Debug("read draft", "project_id", projectID, "path", path)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("draft miss", "project_id", projectID, "path", path)
			return nil, fmt.Errorf("read draft: %w", err)
		}
		logger.Error("read draft failed", "project_id", projectID, "path", path, "err", err)
		return nil, fmt.Errorf("read draft: %w", err)
	}

	data = json.RawMessage(strings.TrimSpace(string(data)))
	if !json.Valid(data) {
		logger.Error("draft invalid json", "project_id", projectID, "path", path)
		return nil, fmt.Errorf("draft json is invalid")
	}

	logger.Info("loaded draft", "project_id", projectID, "path", path, "bytes", len(data))
	return data, nil
}

// SaveDraft saves draft and returns the resulting value or error.
func SaveDraft(projectID string, raw json.RawMessage) error {
	logger := corelog.For("config")
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if !json.Valid(raw) {
		return fmt.Errorf("draft json is invalid")
	}
	if err := EnsurePrivateDir(GetDraftsDirPath()); err != nil {
		return fmt.Errorf("create drafts dir: %w", err)
	}

	path := GetDraftPath(projectID)
	logger.Debug("write draft", "project_id", projectID, "path", path)
	data := append(append(json.RawMessage(nil), raw...), '\n')
	if err := os.WriteFile(path, data, PrivateFileMode); err != nil {
		logger.Error("write draft failed", "project_id", projectID, "path", path, "err", err)
		return fmt.Errorf("write draft: %w", err)
	}
	if err := EnsurePrivateFile(path); err != nil {
		return fmt.Errorf("chmod draft: %w", err)
	}

	logger.Info("saved draft", "project_id", projectID, "path", path, "bytes", len(raw))
	return nil
}

// DeleteDraft removes draft and returns the resulting value or error.
func DeleteDraft(projectID string) error {
	path := GetDraftPath(projectID)
	logger := corelog.For("config")
	logger.Debug("remove draft", "project_id", projectID, "path", path)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Error("remove draft failed", "project_id", projectID, "path", path, "err", err)
		return fmt.Errorf("remove draft: %w", err)
	}
	logger.Info("draft removed", "project_id", projectID, "path", path)
	return nil
}

// ListDraftProjectIDs lists draft project ids and returns the resulting value or error.
func ListDraftProjectIDs() ([]string, error) {
	path := GetDraftsDirPath()
	logger := corelog.For("config")
	logger.Debug("list drafts", "path", path)

	entries, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
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
	sort.Strings(ids)
	return ids, nil
}

// PurgeDrafts handles purge drafts and returns the resulting value or error.
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
