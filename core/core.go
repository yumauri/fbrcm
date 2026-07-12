package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type Core struct {
	ctx context.Context
	// firebase stores firebase clients by auth id.
	firebase map[string]*firebase.Service
	// firebaseMu protects firebase clients.
	firebaseMu sync.Mutex
	// firebaseInit deduplicates concurrent client creation per auth id.
	firebaseInit singleflight.Group
	// versionHistory deduplicates concurrent version-pair reads per project and selectors.
	versionHistory singleflight.Group
}

func NewService(ctx context.Context) (*Core, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	corelog.For("core").Debug("core service initialized")

	return &Core{ctx: ctx, firebase: make(map[string]*firebase.Service)}, nil
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
		projects, err := s.syncProjects(ctx, "")
		if err != nil {
			return nil, "", err
		}
		return projects, "firebase", nil
	}

	logger.Error("projects cache read failed", "err", loadErr)
	return nil, "", fmt.Errorf("read projects cache: %w", loadErr)
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
