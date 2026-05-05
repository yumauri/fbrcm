package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"fbrcm/core/config"
	"fbrcm/core/firebase"
	corelog "fbrcm/core/log"
)

type Project = config.Project
type WhoAmI = firebase.WhoAmI

type Core struct {
	ctx          context.Context
	firebase     *firebase.Service
	firebaseErr  error
	firebaseOnce sync.Once
}

func NewService(ctx context.Context) (*Core, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	corelog.For("core").Debug("core service initialized")

	return &Core{ctx: ctx}, nil
}

func (s *Core) ListProjects(ctx context.Context) ([]Project, string, error) {
	logger := corelog.For("core")
	logger.Debug("list projects requested")

	cached, loadErr := config.LoadProjects()
	if loadErr == nil {
		logger.Info("loaded projects from cache", "count", len(cached))
		return cached, "cache", nil
	}

	if errors.Is(loadErr, os.ErrNotExist) || errors.Is(loadErr, config.ErrEmptyProjectsFile) {
		logger.Warn("projects cache miss; syncing from firebase", "reason", loadErr)
		projects, err := s.syncProjects(ctx)
		if err != nil {
			return nil, "", err
		}
		return projects, "firebase", nil
	}

	logger.Error("projects cache read failed", "err", loadErr)
	return nil, "", fmt.Errorf("cache error: %v", loadErr)
}

func (s *Core) SyncProjects(ctx context.Context) ([]Project, string, error) {
	corelog.For("core").Info("projects sync requested")
	projects, err := s.syncProjects(ctx)
	if err != nil {
		return nil, "", err
	}
	return projects, "firebase", nil
}

func (s *Core) EnsureLogin(ctx context.Context, noOpen bool) error {
	logger := corelog.For("core")
	logger.Info("login requested", "no_open", noOpen)
	fb, err := firebase.NewServiceWithOptions(s.ctx, !noOpen)
	if err != nil {
		logger.Error("login failed", "err", err)
		return err
	}
	s.firebase = fb
	s.firebaseErr = nil
	logger.Info("login ready")
	return nil
}

func (s *Core) WhoAmI(ctx context.Context) (*WhoAmI, error) {
	logger := corelog.For("core")
	logger.Info("whoami requested")
	return firebase.ReadWhoAmI(ctx)
}

func (s *Core) ExportRemoteConfig(ctx context.Context, projectID string) (json.RawMessage, string, error) {
	logger := corelog.For("core")
	logger.Info("export remote config requested", "project_id", projectID)

	fb, err := s.firebaseService(ctx)
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

func (s *Core) ValidateRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage) error {
	logger := corelog.For("core")
	logger.Info("validate remote config requested", "project_id", projectID)

	fb, err := s.firebaseService(ctx)
	if err != nil {
		return err
	}

	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		logger.Error("remote config validation payload decode failed", "project_id", projectID, "err", err)
		return fmt.Errorf("decode remote config: %w", err)
	}

	_, etag, err := fb.GetRemoteConfig(ctx, projectID)
	if err != nil {
		logger.Error("firebase remote config validation preflight fetch failed", "project_id", projectID, "err", err)
		return fmt.Errorf("firebase error: %w", err)
	}

	return s.ValidateRemoteConfigWithETag(ctx, projectID, raw, etag)
}

func (s *Core) ValidateRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) error {
	logger := corelog.For("core")
	logger.Info("validate remote config with etag requested", "project_id", projectID, "etag", etag)

	fb, err := s.firebaseService(ctx)
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

func (s *Core) PublishRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage) (json.RawMessage, string, error) {
	logger := corelog.For("core")
	logger.Info("publish remote config requested", "project_id", projectID)

	fb, err := s.firebaseService(ctx)
	if err != nil {
		return nil, "", err
	}

	if _, err := firebase.ParseRemoteConfig(raw); err != nil {
		logger.Error("remote config publish payload decode failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("decode remote config: %w", err)
	}

	_, etag, err := fb.GetRemoteConfig(ctx, projectID)
	if err != nil {
		logger.Error("firebase remote config publish preflight fetch failed", "project_id", projectID, "err", err)
		return nil, "", fmt.Errorf("firebase error: %w", err)
	}

	return s.PublishRemoteConfigWithETag(ctx, projectID, raw, etag)
}

func (s *Core) PublishRemoteConfigWithETag(ctx context.Context, projectID string, raw json.RawMessage, etag string) (json.RawMessage, string, error) {
	logger := corelog.For("core")
	logger.Info("publish remote config with etag requested", "project_id", projectID, "etag", etag)

	fb, err := s.firebaseService(ctx)
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

func (s *Core) ImportRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage) (json.RawMessage, string, error) {
	if err := s.ValidateRemoteConfig(ctx, projectID, raw); err != nil {
		return nil, "", err
	}
	return s.PublishRemoteConfig(ctx, projectID, raw)
}

func (s *Core) PurgeProjects() error {
	logger := corelog.For("core")
	logger.Info("purge projects cache requested")
	if err := config.PurgeProjects(); err != nil {
		logger.Error("purge projects cache failed", "err", err)
		return fmt.Errorf("purge projects cache: %w", err)
	}

	logger.Info("purged projects cache")
	return nil
}

func (s *Core) syncProjects(ctx context.Context) ([]Project, error) {
	logger := corelog.For("core")
	logger.Info("syncing projects from firebase")

	fb, err := s.firebaseService(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := fb.ListProjects(ctx)
	if err != nil {
		logger.Error("firebase projects sync failed", "err", err)
		return nil, fmt.Errorf("firebase error: %w", err)
	}

	existing, loadErr := config.LoadProjects()
	if loadErr != nil && !errors.Is(loadErr, os.ErrNotExist) && !errors.Is(loadErr, config.ErrEmptyProjectsFile) {
		logger.Error("load existing projects cache failed before merge", "err", loadErr)
		return nil, fmt.Errorf("load existing projects cache: %w", loadErr)
	}

	updatedAt := time.Now().UTC()
	cfgProjects := mergeProjects(existing, toConfigProjects(projects), updatedAt)
	if firebase.IsDryRun(ctx) {
		logger.Warn("dry run, skip projects cache save after sync", "count", len(cfgProjects))
		return cfgProjects, nil
	}
	if saveErr := config.SaveProjects(cfgProjects, updatedAt); saveErr != nil {
		logger.Error("save projects cache failed", "count", len(cfgProjects), "err", saveErr)
		return nil, fmt.Errorf("save projects cache: %w", saveErr)
	}

	logger.Info("projects synced from firebase", "count", len(cfgProjects), "updated_at", updatedAt.Format(time.RFC3339))
	return cfgProjects, nil
}

func (s *Core) firebaseService(ctx context.Context) (*firebase.Service, error) {
	s.firebaseOnce.Do(func() {
		logger := corelog.For("core")
		logger.Debug("initializing firebase service")

		serviceCtx := s.ctx
		if ctx != nil {
			serviceCtx = ctx
		}
		s.firebase, s.firebaseErr = firebase.NewService(serviceCtx)
		if s.firebaseErr != nil {
			logger.Error("firebase service initialization failed", "err", s.firebaseErr)
			return
		}

		logger.Debug("firebase service initialized")
	})

	return s.firebase, s.firebaseErr
}

func toConfigProjects(projects []firebase.Project) []config.Project {
	out := make([]config.Project, len(projects))
	for i, p := range projects {
		out[i] = config.Project{
			Name:          p.Name,
			ProjectID:     p.ProjectID,
			ProjectNumber: p.ProjectNumber,
			State:         p.State,
			ETag:          p.ETag,
			UpdatedAt:     p.UpdateTime,
		}
	}
	return out
}

func mergeProjects(existing, incoming []config.Project, now time.Time) []config.Project {
	byID := make(map[string]config.Project, len(existing))
	for _, project := range existing {
		byID[project.ProjectID] = project
	}

	updatedAt := now.Format(time.RFC3339)
	merged := make([]config.Project, 0, len(incoming))
	for _, project := range incoming {
		if previous, ok := byID[project.ProjectID]; ok {
			if sameProject(previous, project) {
				project.SyncedAt = previous.SyncedAt
			} else {
				project.SyncedAt = updatedAt
			}
		} else {
			project.SyncedAt = updatedAt
		}
		merged = append(merged, project)
	}

	return merged
}

func sameProject(left, right config.Project) bool {
	if left.ETag != "" && right.ETag != "" {
		return left.ETag == right.ETag
	}
	if left.UpdatedAt != "" && right.UpdatedAt != "" {
		return left.UpdatedAt == right.UpdatedAt
	}

	return left.Name == right.Name &&
		left.ProjectNumber == right.ProjectNumber &&
		left.State == right.State
}
