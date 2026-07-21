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
	rcdisplay "github.com/yumauri/fbrcm/core/rc/display"
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

// ProjectAuthBindingFailure describes a selected project that the requested
// auth identity did not discover.
type ProjectAuthBindingFailure struct {
	Project Project
	Reason  string
}

// ProjectAuthBindingResult summarizes a batch auth binding operation.
type ProjectAuthBindingResult struct {
	Bound   []Project
	Skipped []ProjectAuthBindingFailure
}

// BindProjectsAuth binds matched projects that were discovered by auth id.
func (s *Core) BindProjectsAuth(filters []string, authID string) (ProjectAuthBindingResult, error) {
	return s.bindProjectsAuth(authID, func(project Project) bool {
		return matchProjectFilter(project, filters)
	})
}

// BindProjectIDsAuth binds exact cached project IDs to an auth identity.
func (s *Core) BindProjectIDsAuth(projectIDs []string, authID string) ([]Project, error) {
	ids := make(map[string]struct{}, len(projectIDs))
	for _, projectID := range projectIDs {
		ids[projectID] = struct{}{}
	}
	result, err := s.bindProjectsAuth(authID, func(project Project) bool {
		_, ok := ids[project.ProjectID]
		return ok
	})
	if err != nil {
		return nil, err
	}
	if len(result.Skipped) > 0 {
		return result.Bound, fmt.Errorf("auth %q did not discover %s", authID, rcdisplay.FormatCount(len(result.Skipped), "selected project", "selected projects"))
	}
	return result.Bound, nil
}

func (s *Core) bindProjectsAuth(authID string, match func(Project) bool) (ProjectAuthBindingResult, error) {
	if _, err := s.authEntry(authID); err != nil {
		return ProjectAuthBindingResult{}, err
	}
	projects, err := config.LoadProjects()
	if err != nil {
		return ProjectAuthBindingResult{}, err
	}
	result := ProjectAuthBindingResult{
		Bound:   make([]Project, 0),
		Skipped: make([]ProjectAuthBindingFailure, 0),
	}
	matched := 0
	for i := range projects {
		if !match(projects[i]) {
			continue
		}
		matched++
		if !contains(projects[i].DiscoveredBy, authID) {
			result.Skipped = append(result.Skipped, ProjectAuthBindingFailure{
				Project: projects[i],
				Reason:  fmt.Sprintf("project was not discovered by auth %q", authID),
			})
			continue
		}
		projects[i].AuthID = authID
		projects[i].Disabled = false
		result.Bound = append(result.Bound, projects[i])
	}
	if matched == 0 {
		return ProjectAuthBindingResult{}, fmt.Errorf("no projects matched")
	}
	if len(result.Bound) > 0 {
		if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
			return ProjectAuthBindingResult{}, err
		}
	}
	return result, nil
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
	serviceCtx := s.ctx
	if ctx != nil {
		serviceCtx = ctx
	}
	fb, err := firebase.NewServiceForAuth(serviceCtx, auth, !noOpen)
	if err != nil {
		logger.Error("login failed", "err", err)
		return err
	}
	s.firebaseMu.Lock()
	s.firebase[firebaseClientKey(authID)] = fb
	s.firebaseMu.Unlock()
	logger.Info("login ready")
	return nil
}

func (s *Core) ResetProjects() error {
	logger := corelog.For("core")
	logger.Info("reset projects registry requested")
	if err := config.ResetProjects(); err != nil {
		logger.Error("reset projects registry failed", "err", err)
		return fmt.Errorf("reset projects registry: %w", err)
	}

	logger.Info("reset projects registry")
	return nil
}

// DeleteProjectIDs removes projects and all of their local Remote Config
// caches, version snapshots, and drafts. It never creates or calls a Firebase
// client.
func (s *Core) DeleteProjectIDs(projectIDs []string) ([]Project, error) {
	ids := make(map[string]struct{}, len(projectIDs))
	for _, projectID := range projectIDs {
		if projectID != "" {
			ids[projectID] = struct{}{}
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no projects matched")
	}

	projects, err := config.LoadProjects()
	if err != nil {
		return nil, err
	}
	remaining := make([]Project, 0, len(projects))
	deleted := make([]Project, 0, len(ids))
	for _, project := range projects {
		if _, ok := ids[project.ProjectID]; ok {
			deleted = append(deleted, project)
			continue
		}
		remaining = append(remaining, project)
	}
	if len(deleted) == 0 {
		return nil, fmt.Errorf("no projects matched")
	}
	if len(deleted) != len(ids) {
		return nil, fmt.Errorf("projects config does not contain %s", rcdisplay.FormatCount(len(ids)-len(deleted), "selected project", "selected projects"))
	}

	logger := corelog.For("core")
	logger.Info("delete local projects requested", "count", len(deleted))
	for _, project := range deleted {
		if err := config.DeleteParametersCacheForProject(project.ProjectID); err != nil {
			return nil, fmt.Errorf("delete caches for project %s: %w", project.ProjectID, err)
		}
		if err := config.DeleteDraft(project.ProjectID); err != nil {
			return nil, fmt.Errorf("delete draft for project %s: %w", project.ProjectID, err)
		}
	}
	if err := config.SaveProjects(remaining, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("save projects after delete: %w", err)
	}

	logger.Info("deleted local projects", "count", len(deleted))
	return deleted, nil
}

func (s *Core) syncProjects(ctx context.Context, onlyAuthID string) ([]Project, error) {
	logger := corelog.For("core")
	logger.Info("syncing projects from firebase")

	var authFile *config.AuthFile
	var err error
	if onlyAuthID == "" {
		authFile, err = loadRequiredAuth()
	} else {
		authFile, err = loadAuthWithSetupHint()
	}
	if err != nil {
		return nil, err
	}

	authEntries := authFile.Auth
	if onlyAuthID != "" {
		auth, ok := authFile.FindAuth(onlyAuthID)
		if !ok {
			return nil, authNotConfiguredError(authFile, onlyAuthID)
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
