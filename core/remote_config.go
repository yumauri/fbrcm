package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func (s *Core) ExportRemoteConfig(ctx context.Context, projectID string) (json.RawMessage, string, error) {
	logger := corelog.For("core")
	logger.Info("export remote config requested", "project_id", projectID)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	raw, etag, err := fb.GetRemoteConfig(ctx, projectID)
	if err != nil {
		logger.Error("firebase remote config export failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	return raw, etag, nil
}

func (s *Core) ValidateRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) error {
	logger := corelog.For("core")
	logger.Info("validate remote config with etag requested", "project_id", projectID, "etag", etag)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return err
	}

	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		logger.Error("remote config validation payload decode failed", "project_id", projectID, "err", err)
		return fmt.Errorf("decode remote config: %w", err)
	}

	if err := fb.ValidateRemoteConfig(ctx, projectID, raw, etag); err != nil {
		logger.Error("firebase remote config validation failed", "project_id", projectID, "err", err)
		return fmt.Errorf("firebase error: %w", err)
	}

	return nil
}

func (s *Core) PublishRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error) {
	logger := corelog.For("core")
	logger.Info("publish remote config with etag requested", "project_id", projectID, "etag", etag)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, "", err
	}

	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		logger.Error("remote config publish payload decode failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("decode remote config: %w", err)
	}

	updatedRaw, nextETag, err := fb.UpdateRemoteConfig(ctx, projectID, raw, etag)
	if err != nil {
		logger.Error("firebase remote config publish failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	cache := &config.ParametersCache{
		ETag:         nextETag,
		CachedAt:     time.Now().UTC(),
		RemoteConfig: updatedRaw,
	}
	if firebase.IsDryRun(ctx) {
		logger.Warn("dry run, skip parameters cache update after publish", "project_id", projectID, "etag", nextETag)
		return updatedRaw, nextETag, nil
	}
	if err := config.SaveParametersCache(projectID, cache); err != nil {
		logger.Error("save parameters cache after publish failed", "project_id", projectID, "etag", nextETag, "err", err)
		return nil, "", fmt.Errorf("save parameters cache: %w", err)
	}

	return updatedRaw, nextETag, nil
}
