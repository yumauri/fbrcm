package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type Project = config.Project

// AuthPaths describes files owned or used by an auth identity.
type AuthPaths struct {
	AuthConfigPath     string
	ProfileConfigPath  string
	ClientSecretPath   string
	TokenPath          string
	ServiceAccountPath string
}

// Core holds core state used by the core package.
type Core struct {
	// ctx stores ctx for Core.
	ctx context.Context
	// firebase stores firebase clients by auth id.
	firebase map[string]*firebase.Service
	// firebaseMu protects firebase clients.
	firebaseMu sync.Mutex
}

// NewService constructs service and returns the resulting value or error.
func NewService(ctx context.Context) (*Core, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	corelog.For("core").Debug("core service initialized")

	return &Core{ctx: ctx, firebase: make(map[string]*firebase.Service)}, nil
}

// ListProjects lists projects for Core and returns the resulting state or error.
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
		projects, err := s.syncProjects(ctx, "")
		if err != nil {
			return nil, "", err
		}
		return projects, "firebase", nil
	}

	logger.Error("projects cache read failed", "err", loadErr)
	return nil, "", fmt.Errorf("cache error: %v", loadErr)
}

// ListAuth lists configured auth identities.
func (s *Core) ListAuth() ([]config.AuthEntry, string, error) {
	auth, err := config.LoadAuthOrEmpty()
	if err != nil {
		return nil, "", err
	}
	return append([]config.AuthEntry(nil), auth.Auth...), auth.DefaultAuthID, nil
}

// AddOAuthAuth adds or replaces OAuth auth identity.
func (s *Core) AddOAuthAuth(authID, label string, secret []byte) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := config.LoadAuthOrEmpty()
	if err != nil {
		return config.AuthEntry{}, err
	}
	previousAuth, hadPrevious := authFile.FindAuth(authID)
	entry := config.DefaultOAuthAuthEntry(authID, label)
	authFile = config.UpsertAuthEntry(authFile, entry)
	if err := config.SaveAuth(authFile); err != nil {
		return config.AuthEntry{}, err
	}
	secretPath := config.OAuthClientSecretPath(entry)
	tokenPath := config.OAuthTokenPath(entry)
	previousSecret, readErr := os.ReadFile(secretPath)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return config.AuthEntry{}, fmt.Errorf("read existing client secret: %w", readErr)
	}
	secretChanged := readErr == nil && string(previousSecret) != string(secret)
	if err := config.EnsurePrivateDir(filepath.Dir(secretPath)); err != nil {
		return config.AuthEntry{}, fmt.Errorf("create auth dir: %w", err)
	}
	if err := config.EnsurePrivateDir(filepath.Dir(tokenPath)); err != nil {
		return config.AuthEntry{}, fmt.Errorf("create auth cache dir: %w", err)
	}
	if err := writePrivateFile(secretPath, secret); err != nil {
		return config.AuthEntry{}, fmt.Errorf("write client secret: %w", err)
	}
	if secretChanged {
		if err := os.Remove(tokenPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return config.AuthEntry{}, fmt.Errorf("remove token for previous client secret: %w", err)
		}
	}
	if hadPrevious && previousAuth.Type != config.AuthTypeOAuth {
		if err := removeAuthFiles(previousAuth); err != nil {
			return config.AuthEntry{}, err
		}
	}
	s.dropFirebaseService(authID)
	return entry, nil
}

// AddServiceAccountAuth adds or replaces service account auth identity.
func (s *Core) AddServiceAccountAuth(authID, label string, key []byte) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := config.LoadAuthOrEmpty()
	if err != nil {
		return config.AuthEntry{}, err
	}
	previous, hadPrevious := authFile.FindAuth(authID)
	entry := config.DefaultServiceAccountAuthEntry(authID, label)
	authFile = config.UpsertAuthEntry(authFile, entry)
	if err := config.SaveAuth(authFile); err != nil {
		return config.AuthEntry{}, err
	}
	keyPath := config.ServiceAccountKeyPath(entry)
	if err := config.EnsurePrivateDir(filepath.Dir(keyPath)); err != nil {
		return config.AuthEntry{}, fmt.Errorf("create auth dir: %w", err)
	}
	if err := writePrivateFile(keyPath, key); err != nil {
		return config.AuthEntry{}, fmt.Errorf("write service account key: %w", err)
	}
	if hadPrevious && previous.Type != config.AuthTypeServiceAccount {
		if err := removeAuthFiles(previous); err != nil {
			return config.AuthEntry{}, err
		}
	}
	s.dropFirebaseService(authID)
	return entry, nil
}

// AddGCloudAuth adds or replaces gcloud ADC auth identity.
func (s *Core) AddGCloudAuth(authID, label string) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := config.LoadAuthOrEmpty()
	if err != nil {
		return config.AuthEntry{}, err
	}
	previous, hadPrevious := authFile.FindAuth(authID)
	entry := config.DefaultGCloudAuthEntry(authID, label)
	authFile = config.UpsertAuthEntry(authFile, entry)
	if err := config.SaveAuth(authFile); err != nil {
		return config.AuthEntry{}, err
	}
	if hadPrevious {
		if err := removeAuthFiles(previous); err != nil {
			return config.AuthEntry{}, err
		}
	}
	s.dropFirebaseService(authID)
	return entry, nil
}

// AuthPaths gets resolved paths for auth id.
func (s *Core) AuthPaths(authID string) (config.AuthEntry, AuthPaths, error) {
	auth, err := s.authEntry(authID)
	if err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	paths := AuthPaths{
		AuthConfigPath:    config.GetAuthFilePath(),
		ProfileConfigPath: config.GetConfigDirPath(),
	}
	switch auth.Type {
	case config.AuthTypeOAuth:
		paths.ClientSecretPath = config.OAuthClientSecretPath(auth)
		paths.TokenPath = config.OAuthTokenPath(auth)
	case config.AuthTypeServiceAccount:
		paths.ServiceAccountPath = config.ServiceAccountKeyPath(auth)
	}
	return auth, paths, nil
}

// PurgeAuth removes auth identity files and registry entry.
func (s *Core) PurgeAuth(authID string) (config.AuthEntry, AuthPaths, error) {
	authFile, err := config.LoadAuth()
	if err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	auth, ok := authFile.FindAuth(authID)
	if !ok {
		return config.AuthEntry{}, AuthPaths{}, fmt.Errorf("auth %q is not configured", authID)
	}
	_, paths, err := s.AuthPaths(authID)
	if err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	authFile, _ = config.RemoveAuth(authFile, authID)
	if err := config.SaveAuth(authFile); err != nil {
		return config.AuthEntry{}, AuthPaths{}, err
	}
	if err := removeFileIfPresent(paths.ClientSecretPath); err != nil {
		return config.AuthEntry{}, AuthPaths{}, fmt.Errorf("remove client secret: %w", err)
	}
	if err := removeFileIfPresent(paths.TokenPath); err != nil {
		return config.AuthEntry{}, AuthPaths{}, fmt.Errorf("remove token: %w", err)
	}
	if err := removeFileIfPresent(paths.ServiceAccountPath); err != nil {
		return config.AuthEntry{}, AuthPaths{}, fmt.Errorf("remove service account key: %w", err)
	}
	s.dropFirebaseService(authID)
	return auth, paths, nil
}

// BindProjectsAuth binds matched projects to auth id.
func (s *Core) BindProjectsAuth(filters []string, authID string) ([]Project, error) {
	if _, err := s.authEntry(authID); err != nil {
		return nil, err
	}
	projects, err := config.LoadProjects()
	if err != nil {
		return nil, err
	}
	matched := make([]Project, 0)
	for i := range projects {
		if matchProjectFilter(projects[i], filters) {
			projects[i].AuthID = authID
			projects[i].DiscoveredBy = appendUnique(projects[i].DiscoveredBy, authID)
			matched = append(matched, projects[i])
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("no projects matched")
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		return nil, err
	}
	return matched, nil
}

// ProjectByID gets project by id.
func (s *Core) ProjectByID(projectID string) (Project, error) {
	projects, err := config.LoadProjects()
	if err != nil {
		return Project{}, err
	}
	for _, project := range projects {
		if project.ProjectID == projectID {
			return project, nil
		}
	}
	return Project{}, fmt.Errorf("project %q is not in projects config", projectID)
}

// SyncProjects handles sync projects for Core and returns the resulting state or error.
func (s *Core) SyncProjects(ctx context.Context) ([]Project, string, error) {
	corelog.For("core").Info("projects sync requested")
	projects, err := s.syncProjects(ctx, "")
	if err != nil {
		return nil, "", err
	}
	return projects, "firebase", nil
}

// SyncProjectsForAuth syncs projects for one auth id.
func (s *Core) SyncProjectsForAuth(ctx context.Context, authID string) ([]Project, string, error) {
	corelog.For("core").Info("projects sync requested", "auth_id", authID)
	projects, err := s.syncProjects(ctx, authID)
	if err != nil {
		return nil, "", err
	}
	return projects, "firebase", nil
}

// EnsureAuthLogin handles ensure login for Core and returns the resulting state or error.
func (s *Core) EnsureAuthLogin(ctx context.Context, authID string, noOpen bool) error {
	logger := corelog.For("core")
	logger.Info("login requested", "auth_id", authID, "no_open", noOpen)
	auth, err := s.authEntry(authID)
	if err != nil {
		return err
	}
	fb, err := firebase.NewServiceForAuth(s.ctx, auth, !noOpen)
	if err != nil {
		logger.Error("login failed", "err", err)
		return err
	}
	s.firebaseMu.Lock()
	s.firebase[authID] = fb
	s.firebaseMu.Unlock()
	logger.Info("login ready")
	return nil
}

// ExportRemoteConfig handles export remote config for Core and returns the resulting state or error.
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

// ValidateRemoteConfig handles validate remote config for Core and returns the resulting state or error.
func (s *Core) ValidateRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage) error {
	logger := corelog.For("core")
	logger.Info("validate remote config requested", "project_id", projectID)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
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

// ValidateRemoteConfigWithETag handles validate remote config with etag for Core and returns the resulting state or error.
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

// PublishRemoteConfig handles publish remote config for Core and returns the resulting state or error.
func (s *Core) PublishRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage) (json.RawMessage, string, error) {
	logger := corelog.For("core")
	logger.Info("publish remote config requested", "project_id", projectID)

	fb, err := s.firebaseServiceForProject(ctx, projectID)
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

// PublishRemoteConfigWithETag handles publish remote config with etag for Core and returns the resulting state or error.
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

// ImportRemoteConfig handles import remote config for Core and returns the resulting state or error.
func (s *Core) ImportRemoteConfig(ctx context.Context, projectID string, raw json.RawMessage) (json.RawMessage, string, error) {
	if err := s.ValidateRemoteConfig(ctx, projectID, raw); err != nil {
		return nil, "", err
	}
	return s.PublishRemoteConfig(ctx, projectID, raw)
}

// PurgeProjects handles purge projects for Core and returns the resulting state or error.
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

// syncProjects handles sync projects for Core and returns the resulting state or error.
func (s *Core) syncProjects(ctx context.Context, onlyAuthID string) ([]Project, error) {
	logger := corelog.For("core")
	logger.Info("syncing projects from firebase")

	authFile, err := config.LoadAuth()
	if err != nil {
		return nil, err
	}

	authEntries := authFile.Auth
	if onlyAuthID != "" {
		auth, ok := authFile.FindAuth(onlyAuthID)
		if !ok {
			return nil, fmt.Errorf("auth %q is not configured", onlyAuthID)
		}
		authEntries = []config.AuthEntry{auth}
	}

	existing, loadErr := config.LoadProjects()
	if loadErr != nil && !errors.Is(loadErr, os.ErrNotExist) && !errors.Is(loadErr, config.ErrEmptyProjectsFile) {
		logger.Error("load existing projects cache failed before merge", "err", loadErr)
		return nil, fmt.Errorf("load existing projects cache: %w", loadErr)
	}

	incomingByID := map[string]config.Project{}
	discoveredByID := map[string][]string{}
	for _, auth := range authEntries {
		fb, err := s.firebaseServiceForAuth(ctx, auth.ID)
		if err != nil {
			return nil, err
		}
		projects, err := fb.ListProjects(ctx)
		if err != nil {
			logger.Error("firebase projects sync failed", "auth_id", auth.ID, "err", err)
			return nil, fmt.Errorf("firebase error: %w", err)
		}
		for _, project := range toConfigProjects(projects) {
			incomingByID[project.ProjectID] = project
			discoveredByID[project.ProjectID] = appendUnique(discoveredByID[project.ProjectID], auth.ID)
		}
	}

	incoming := make([]config.Project, 0, len(incomingByID))
	for projectID, project := range incomingByID {
		project.DiscoveredBy = discoveredByID[projectID]
		incoming = append(incoming, project)
	}

	updatedAt := time.Now().UTC()
	cfgProjects := mergeProjects(existing, incoming, authFile.DefaultAuthID, authOrder(authFile.Auth), onlyAuthID, updatedAt)
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

// firebaseServiceForProject handles firebase service lookup for project.
func (s *Core) firebaseServiceForProject(ctx context.Context, projectID string) (*firebase.Service, error) {
	project, err := s.ProjectByID(projectID)
	if err != nil {
		return nil, err
	}
	return s.firebaseServiceForAuth(ctx, project.AuthID)
}

// firebaseServiceForAuth handles firebase service lookup for auth id.
func (s *Core) firebaseServiceForAuth(ctx context.Context, authID string) (*firebase.Service, error) {
	auth, err := s.authEntry(authID)
	if err != nil {
		return nil, err
	}
	s.firebaseMu.Lock()
	if fb, ok := s.firebase[authID]; ok {
		s.firebaseMu.Unlock()
		return fb, nil
	}
	s.firebaseMu.Unlock()

	serviceCtx := s.ctx
	if ctx != nil {
		serviceCtx = ctx
	}
	fb, err := firebase.NewServiceForAuth(serviceCtx, auth, true)
	if err != nil {
		return nil, err
	}

	s.firebaseMu.Lock()
	s.firebase[authID] = fb
	s.firebaseMu.Unlock()
	return fb, nil
}

func (s *Core) authEntry(authID string) (config.AuthEntry, error) {
	if err := config.ValidateAuthID(authID); err != nil {
		return config.AuthEntry{}, err
	}
	authFile, err := config.LoadAuth()
	if err != nil {
		return config.AuthEntry{}, err
	}
	auth, ok := authFile.FindAuth(authID)
	if !ok {
		return config.AuthEntry{}, fmt.Errorf("auth %q is not configured", authID)
	}
	return auth, nil
}

func (s *Core) dropFirebaseService(authID string) {
	s.firebaseMu.Lock()
	delete(s.firebase, authID)
	s.firebaseMu.Unlock()
}

func writePrivateFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, config.PrivateFileMode); err != nil {
		return err
	}
	if err := config.EnsurePrivateFile(path); err != nil {
		return err
	}
	return nil
}

func removeAuthFiles(auth config.AuthEntry) error {
	switch auth.Type {
	case config.AuthTypeOAuth:
		if err := removeFileIfPresent(config.OAuthClientSecretPath(auth)); err != nil {
			return fmt.Errorf("remove client secret: %w", err)
		}
		if err := removeFileIfPresent(config.OAuthTokenPath(auth)); err != nil {
			return fmt.Errorf("remove token: %w", err)
		}
	case config.AuthTypeServiceAccount:
		if err := removeFileIfPresent(config.ServiceAccountKeyPath(auth)); err != nil {
			return fmt.Errorf("remove service account key: %w", err)
		}
	}
	return nil
}

func removeFileIfPresent(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// toConfigProjects handles to config projects and returns the resulting value or error.
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

// mergeProjects handles merge projects and returns the resulting value or error.
func mergeProjects(existing, incoming []config.Project, defaultAuthID string, authIDs []string, onlyAuthID string, now time.Time) []config.Project {
	byID := make(map[string]config.Project, len(existing))
	for _, project := range existing {
		byID[project.ProjectID] = project
	}

	updatedAt := now.Format(time.RFC3339)
	mergedByID := make(map[string]config.Project, len(existing)+len(incoming))
	if onlyAuthID != "" {
		for _, project := range existing {
			mergedByID[project.ProjectID] = project
		}
	}
	for _, project := range incoming {
		if previous, ok := byID[project.ProjectID]; ok {
			project.AuthID = chooseProjectAuth(previous.AuthID, project.DiscoveredBy, defaultAuthID, authIDs)
			if sameProject(previous, project) {
				project.SyncedAt = previous.SyncedAt
			} else {
				project.SyncedAt = updatedAt
			}
		} else {
			project.AuthID = chooseProjectAuth("", project.DiscoveredBy, defaultAuthID, authIDs)
			project.SyncedAt = updatedAt
		}
		mergedByID[project.ProjectID] = project
	}

	merged := make([]config.Project, 0, len(mergedByID))
	for _, project := range mergedByID {
		merged = append(merged, project)
	}
	return merged
}

// sameProject handles same project and returns the resulting value or error.
func sameProject(left, right config.Project) bool {
	authSame := left.AuthID == right.AuthID &&
		strings.Join(left.DiscoveredBy, "\x00") == strings.Join(right.DiscoveredBy, "\x00")
	if left.ETag != "" && right.ETag != "" {
		return left.ETag == right.ETag && authSame
	}
	if left.UpdatedAt != "" && right.UpdatedAt != "" {
		return left.UpdatedAt == right.UpdatedAt && authSame
	}

	return left.Name == right.Name &&
		left.ProjectNumber == right.ProjectNumber &&
		left.State == right.State &&
		authSame
}

func chooseProjectAuth(previous string, discovered []string, defaultAuthID string, authIDs []string) string {
	if contains(discovered, previous) {
		return previous
	}
	if contains(discovered, defaultAuthID) {
		return defaultAuthID
	}
	for _, authID := range authIDs {
		if contains(discovered, authID) {
			return authID
		}
	}
	if len(discovered) > 0 {
		return discovered[0]
	}
	return previous
}

func authOrder(auth []config.AuthEntry) []string {
	out := make([]string, 0, len(auth))
	for _, entry := range auth {
		out = append(out, entry.ID)
	}
	return out
}

func appendUnique(values []string, value string) []string {
	if contains(values, value) {
		return values
	}
	return append(values, value)
}

func contains(values []string, value string) bool {
	return slices.Contains(values, value)
}

func matchProjectFilter(project Project, rawFilters []string) bool {
	if len(rawFilters) == 0 {
		return true
	}
	for _, raw := range rawFilters {
		mode := filter.ModeFuzzy
		query := strings.TrimSpace(raw)
		if query == "" {
			continue
		}
		if parsedMode, ok := filter.ModeFromLabel(string([]rune(query)[0])); ok {
			mode = parsedMode
			query = string([]rune(query)[1:])
		}
		nameMatch, _ := filter.Match(project.Name, query, mode)
		idMatch, _ := filter.Match(project.ProjectID, query, mode)
		if nameMatch || idMatch {
			return true
		}
	}
	return false
}
