package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	charmlog "charm.land/log/v2"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type AppPlatform string

const (
	AppPlatformAndroid AppPlatform = "android"
	AppPlatformIOS     AppPlatform = "ios"
	AppPlatformWeb     AppPlatform = "web"
)

type FirebaseApp struct {
	ResourceName string      `json:"resource_name"`
	DisplayName  string      `json:"display_name"`
	Platform     AppPlatform `json:"platform" contract:"enum=android|ios|web"`
	AppID        string      `json:"app_id"`
	Namespace    string      `json:"namespace"`
	APIKeyID     string      `json:"api_key_id"`
	State        string      `json:"state" contract:"enum=ACTIVE|DELETED|STATE_UNSPECIFIED"`
	ExpireTime   string      `json:"expire_time,omitempty" contract:"format=date-time"`
}

type FirebaseAppDetails struct {
	FirebaseApp
	ProjectID    string   `json:"project_id"`
	PackageName  *string  `json:"package_name"`
	BundleID     *string  `json:"bundle_id"`
	AppStoreID   *string  `json:"app_store_id"`
	TeamID       *string  `json:"team_id"`
	WebID        *string  `json:"web_id"`
	AppURLs      []string `json:"app_urls"`
	SHA1Hashes   []string `json:"sha1_hashes"`
	SHA256Hashes []string `json:"sha256_hashes"`
	ETag         string   `json:"etag"`
}

type FirebaseAppConfig struct {
	App               FirebaseApp
	SuggestedFilename string
	MediaType         string
	Contents          []byte
}

type ListFirebaseAppsOptions struct {
	ShowDeleted bool
	Update      bool
	CachedOnly  bool
}

type AppCacheSource string

const (
	AppCacheSourceFirebase   AppCacheSource = "firebase"
	AppCacheSourceCache      AppCacheSource = "cache"
	AppCacheSourceCacheStale AppCacheSource = "cache-stale"
)

type FirebaseAppsResult struct {
	Apps         []FirebaseApp
	Source       AppCacheSource
	CachedAt     time.Time
	RefreshError error
	CacheError   error
}

type FirebaseAppDetailsResult struct {
	App          FirebaseAppDetails
	Source       AppCacheSource
	CachedAt     time.Time
	RefreshError error
	CacheError   error
}

type FirebaseAppConfigResult struct {
	Config       FirebaseAppConfig
	Source       AppCacheSource
	CachedAt     time.Time
	RefreshError error
	CacheError   error
}

type AppLookupCandidate struct {
	Name string
	ID   string
}

type AppLookupError struct {
	Kind       string
	Query      string
	Candidates []AppLookupCandidate
}

func (e *AppLookupError) Error() string {
	if e.Kind == "ambiguous" {
		return fmt.Sprintf("several apps match %q", e.Query)
	}
	return fmt.Sprintf("app %q was not found", e.Query)
}

func (s *Core) ListFirebaseApps(ctx context.Context, projectID string, opts ListFirebaseAppsOptions) ([]FirebaseApp, error) {
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	result := make([]FirebaseApp, 0)
	token := ""
	for {
		page, err := fb.ListApps(ctx, projectID, firebase.ListAppsOptions{PageSize: 100, PageToken: token, ShowDeleted: opts.ShowDeleted})
		if err != nil {
			return nil, fmt.Errorf("firebase error: %w", err)
		}
		for _, app := range page.Apps {
			converted, err := firebaseAppFromInfo(app)
			if err != nil {
				return nil, err
			}
			result = append(result, converted)
		}
		if page.NextPageToken == "" {
			break
		}
		token = page.NextPageToken
	}
	sort.SliceStable(result, func(i, j int) bool {
		left := strings.ToLower(string(result[i].Platform) + "\x00" + result[i].DisplayName + "\x00" + result[i].Namespace + "\x00" + result[i].AppID)
		right := strings.ToLower(string(result[j].Platform) + "\x00" + result[j].DisplayName + "\x00" + result[j].Namespace + "\x00" + result[j].AppID)
		return left < right
	})
	return result, nil
}

// ReadFirebaseApps applies the profile-scoped, one-hour application inventory
// cache. The stored index always includes deleted apps; display filtering is
// deliberately local so one cache entry serves both list forms.
func (s *Core) ReadFirebaseApps(ctx context.Context, projectID string, opts ListFirebaseAppsOptions) (FirebaseAppsResult, error) {
	logger := corelog.For("core")
	logger.Debug("read firebase applications requested", "project_id", projectID, "update", opts.Update, "cached_only", opts.CachedOnly, "show_deleted", opts.ShowDeleted)

	policy := ExecutionPolicyFromContext(ctx)
	if !policy.ReadLocalState {
		logger.Info("fetch applications from firebase", "project_id", projectID)
		apps, err := s.ListFirebaseApps(ctx, projectID, ListFirebaseAppsOptions{ShowDeleted: true})
		if err != nil {
			logger.Error("firebase applications fetch failed", "project_id", projectID, "err", err)
			return FirebaseAppsResult{}, err
		}
		return FirebaseAppsResult{Apps: filterDeletedApps(apps, opts.ShowDeleted), Source: AppCacheSourceFirebase}, nil
	}

	cached, cachedErr := loadAppsIndex(projectID)
	if opts.CachedOnly {
		if cachedErr != nil {
			return FirebaseAppsResult{}, appCacheLoadError("index", projectID, "", cachedErr)
		}
		cached.Apps = filterDeletedApps(cached.Apps, opts.ShowDeleted)
		logApplicationsCacheServe(logger, projectID, cached.Source)
		return cached, nil
	}
	if !opts.Update && cachedErr == nil && cached.Source == AppCacheSourceCache {
		cached.Apps = filterDeletedApps(cached.Apps, opts.ShowDeleted)
		logger.Info("serving applications from fresh cache", "project_id", projectID)
		return cached, nil
	}

	switch {
	case opts.Update:
		logger.Info("refresh applications cache from firebase", "project_id", projectID)
	case cachedErr == nil:
		logger.Info("refreshing stale applications cache from firebase", "project_id", projectID)
	default:
		logger.Warn("fetching applications from firebase due to cache miss", "project_id", projectID)
	}
	apps, err := s.ListFirebaseApps(ctx, projectID, ListFirebaseAppsOptions{ShowDeleted: true})
	if err != nil {
		if cachedErr == nil {
			cached.Source = AppCacheSourceCacheStale
			cached.RefreshError = err
			cached.Apps = filterDeletedApps(cached.Apps, opts.ShowDeleted)
			logger.Warn("serving applications from stale cache after refresh failed", "project_id", projectID, "err", err)
			return cached, nil
		}
		logger.Error("firebase applications fetch failed", "project_id", projectID, "err", err)
		return FirebaseAppsResult{}, err
	}
	now := time.Now().UTC()
	result := FirebaseAppsResult{Apps: filterDeletedApps(apps, opts.ShowDeleted), Source: AppCacheSourceFirebase, CachedAt: now}
	if policy.WriteLocalState {
		payload, marshalErr := json.Marshal(apps)
		if marshalErr != nil {
			result.CacheError = marshalErr
		} else if saveErr := coreconfig.SaveAppsIndexCache(projectID, now, payload); saveErr != nil {
			result.CacheError = saveErr
		} else {
			ids := make([]string, 0, len(apps))
			for _, app := range apps {
				ids = append(ids, app.AppID)
			}
			result.CacheError = coreconfig.PruneAppsCacheForProject(projectID, ids)
		}
		if result.CacheError != nil {
			logger.Error("save applications cache failed", "project_id", projectID, "err", result.CacheError)
		} else {
			logger.Info("applications cached from firebase", "project_id", projectID, "app_count", len(apps))
		}
	}
	return result, nil
}

func loadAppsIndex(projectID string) (FirebaseAppsResult, error) {
	logger := corelog.For("core")
	logger.Debug("inspect applications cache", "project_id", projectID)
	record, err := coreconfig.LoadAppsIndexCache(projectID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Warn("applications cache miss", "project_id", projectID)
		} else {
			logger.Error("applications cache load failed", "project_id", projectID, "err", err)
		}
		return FirebaseAppsResult{}, err
	}
	var apps []FirebaseApp
	if err := json.Unmarshal(record.Payload, &apps); err != nil {
		logger.Error("cached applications decode failed", "project_id", projectID, "err", err)
		return FirebaseAppsResult{}, fmt.Errorf("decode apps index payload: %w", err)
	}
	source := AppCacheSourceCache
	if !record.IsFresh(time.Now()) {
		source = AppCacheSourceCacheStale
		logger.Info("applications cache stale", "project_id", projectID, "cached_at", record.CachedAt)
	} else {
		logger.Info("applications cache fresh", "project_id", projectID, "cached_at", record.CachedAt)
	}
	return FirebaseAppsResult{Apps: apps, Source: source, CachedAt: record.CachedAt}, nil
}

func logApplicationsCacheServe(logger *charmlog.Logger, projectID string, source AppCacheSource) {
	if source == AppCacheSourceCache {
		logger.Info("serving applications from fresh cache", "project_id", projectID)
		return
	}
	logger.Info("serving applications from stale cache", "project_id", projectID)
}

func filterDeletedApps(apps []FirebaseApp, showDeleted bool) []FirebaseApp {
	if showDeleted {
		if apps == nil {
			return []FirebaseApp{}
		}
		return apps
	}
	filtered := make([]FirebaseApp, 0, len(apps))
	for _, app := range apps {
		if app.State != "DELETED" {
			filtered = append(filtered, app)
		}
	}
	return filtered
}

func (s *Core) GetFirebaseApp(ctx context.Context, projectID, selector string) (FirebaseAppDetails, error) {
	fb, app, err := s.resolveFirebaseApp(ctx, projectID, selector)
	if err != nil {
		return FirebaseAppDetails{}, err
	}
	return fetchFirebaseAppDetails(ctx, fb, projectID, app)
}

func fetchFirebaseAppDetails(ctx context.Context, fb *firebase.Service, projectID string, app FirebaseApp) (FirebaseAppDetails, error) {
	details := FirebaseAppDetails{FirebaseApp: app, ProjectID: projectID, AppURLs: []string{}, SHA1Hashes: []string{}, SHA256Hashes: []string{}}
	switch app.Platform {
	case AppPlatformAndroid:
		value, err := fb.GetAndroidApp(ctx, projectID, app.ResourceName)
		if err != nil {
			return FirebaseAppDetails{}, fmt.Errorf("firebase error: %w", err)
		}
		details.ProjectID, details.DisplayName, details.AppID, details.APIKeyID, details.State, details.ExpireTime, details.ETag = value.ProjectID, value.DisplayName, value.AppID, value.APIKeyID, value.State, value.ExpireTime, value.ETag
		details.PackageName = stringPointer(value.PackageName)
		details.Namespace = value.PackageName
		details.SHA1Hashes, details.SHA256Hashes = nonNilStrings(value.SHA1Hashes), nonNilStrings(value.SHA256Hashes)
	case AppPlatformIOS:
		value, err := fb.GetIOSApp(ctx, projectID, app.ResourceName)
		if err != nil {
			return FirebaseAppDetails{}, fmt.Errorf("firebase error: %w", err)
		}
		details.ProjectID, details.DisplayName, details.AppID, details.APIKeyID, details.State, details.ExpireTime, details.ETag = value.ProjectID, value.DisplayName, value.AppID, value.APIKeyID, value.State, value.ExpireTime, value.ETag
		details.BundleID, details.AppStoreID, details.TeamID = stringPointer(value.BundleID), stringPointer(value.AppStoreID), stringPointer(value.TeamID)
		details.Namespace = value.BundleID
	case AppPlatformWeb:
		value, err := fb.GetWebApp(ctx, projectID, app.ResourceName)
		if err != nil {
			return FirebaseAppDetails{}, fmt.Errorf("firebase error: %w", err)
		}
		details.ProjectID, details.DisplayName, details.AppID, details.APIKeyID, details.State, details.ExpireTime, details.ETag = value.ProjectID, value.DisplayName, value.AppID, value.APIKeyID, value.State, value.ExpireTime, value.ETag
		details.WebID, details.AppURLs = stringPointer(value.WebID), nonNilStrings(value.AppURLs)
	default:
		return FirebaseAppDetails{}, fmt.Errorf("unsupported Firebase app platform %q", app.Platform)
	}
	return details, nil
}

func (s *Core) ReadFirebaseApp(ctx context.Context, projectID, selector string, opts ListFirebaseAppsOptions) (FirebaseAppDetailsResult, error) {
	logger := corelog.For("core")
	logger.Debug("read firebase application details requested", "project_id", projectID, "selector", selector, "update", opts.Update, "cached_only", opts.CachedOnly)

	index, err := s.ReadFirebaseApps(ctx, projectID, ListFirebaseAppsOptions{ShowDeleted: true, Update: opts.Update, CachedOnly: opts.CachedOnly})
	if err != nil {
		return FirebaseAppDetailsResult{}, err
	}
	app, err := resolveFirebaseAppFromList(selector, index.Apps)
	if err != nil {
		return FirebaseAppDetailsResult{}, err
	}
	if !ExecutionPolicyFromContext(ctx).ReadLocalState {
		logger.Info("fetch application details from firebase", "project_id", projectID, "app_id", app.AppID)
		fb, err := s.firebaseServiceForProject(ctx, projectID)
		if err != nil {
			return FirebaseAppDetailsResult{}, err
		}
		details, err := fetchFirebaseAppDetails(ctx, fb, projectID, app)
		if err != nil {
			logger.Error("firebase application details fetch failed", "project_id", projectID, "app_id", app.AppID, "err", err)
			return FirebaseAppDetailsResult{}, err
		}
		return FirebaseAppDetailsResult{App: details, Source: AppCacheSourceFirebase}, nil
	}
	record, cachedErr := coreconfig.LoadAppDetailsCache(projectID, app.AppID)
	var cached FirebaseAppDetails
	if cachedErr == nil {
		cachedErr = json.Unmarshal(record.Payload, &cached)
		if cachedErr == nil && (cached.AppID != app.AppID || cached.ProjectID != projectID) {
			cachedErr = fmt.Errorf("decode app details cache: cached application identity does not match")
		}
	}
	if cachedErr == nil {
		logAppResourceCacheState(logger, "application details", projectID, app.AppID, record)
	} else if !errors.Is(cachedErr, os.ErrNotExist) {
		logger.Error("application details cache load failed", "project_id", projectID, "app_id", app.AppID, "err", cachedErr)
	}
	if opts.CachedOnly {
		if cachedErr != nil {
			return FirebaseAppDetailsResult{}, appCacheLoadError("details", projectID, app.AppID, cachedErr)
		}
		result := FirebaseAppDetailsResult{App: cached, Source: combinedAppCacheSource(cacheRecordSource(record), index.Source), CachedAt: record.CachedAt, RefreshError: index.RefreshError}
		logAppResourceCacheServe(logger, "application details", projectID, app.AppID, result.Source)
		return result, nil
	}
	if !opts.Update && cachedErr == nil && record.IsFresh(time.Now()) {
		result := FirebaseAppDetailsResult{App: cached, Source: combinedAppCacheSource(AppCacheSourceCache, index.Source), CachedAt: record.CachedAt, RefreshError: index.RefreshError}
		logAppResourceCacheServe(logger, "application details", projectID, app.AppID, result.Source)
		return result, nil
	}
	switch {
	case opts.Update:
		logger.Info("refresh application details cache from firebase", "project_id", projectID, "app_id", app.AppID)
	case cachedErr == nil:
		logger.Info("refreshing stale application details cache from firebase", "project_id", projectID, "app_id", app.AppID)
	default:
		logger.Warn("fetching application details from firebase due to cache miss", "project_id", projectID, "app_id", app.AppID)
	}
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err == nil {
		var details FirebaseAppDetails
		details, err = fetchFirebaseAppDetails(ctx, fb, projectID, app)
		if err == nil {
			now := time.Now().UTC()
			result := FirebaseAppDetailsResult{App: details, Source: combinedAppCacheSource(AppCacheSourceFirebase, index.Source), CachedAt: now, RefreshError: index.RefreshError}
			logger.Info("fetched application details from firebase", "project_id", projectID, "app_id", app.AppID)
			if ExecutionPolicyFromContext(ctx).WriteLocalState {
				payload, marshalErr := json.Marshal(details)
				if marshalErr != nil {
					result.CacheError = marshalErr
				} else {
					result.CacheError = coreconfig.SaveAppDetailsCache(projectID, app.AppID, now, payload)
				}
				if result.CacheError != nil {
					logger.Error("save application details cache failed", "project_id", projectID, "app_id", app.AppID, "err", result.CacheError)
				} else {
					logger.Info("application details cached from firebase", "project_id", projectID, "app_id", app.AppID)
				}
			}
			return result, nil
		}
	}
	if cachedErr == nil {
		logger.Warn("serving application details from stale cache after refresh failed", "project_id", projectID, "app_id", app.AppID, "err", err)
		return FirebaseAppDetailsResult{App: cached, Source: AppCacheSourceCacheStale, CachedAt: record.CachedAt, RefreshError: err}, nil
	}
	logger.Error("firebase application details fetch failed", "project_id", projectID, "app_id", app.AppID, "err", err)
	return FirebaseAppDetailsResult{}, err
}

func (s *Core) GetFirebaseAppConfig(ctx context.Context, projectID, selector string) (FirebaseAppConfig, error) {
	fb, app, err := s.resolveFirebaseApp(ctx, projectID, selector)
	if err != nil {
		return FirebaseAppConfig{}, err
	}
	return fetchFirebaseAppConfig(ctx, fb, projectID, app)
}

func fetchFirebaseAppConfig(ctx context.Context, fb *firebase.Service, projectID string, app FirebaseApp) (FirebaseAppConfig, error) {
	result := FirebaseAppConfig{App: app}
	switch app.Platform {
	case AppPlatformAndroid:
		cfg, err := fb.GetNativeAppConfig(ctx, projectID, app.ResourceName, firebase.AppPlatformAndroid)
		if err != nil {
			return FirebaseAppConfig{}, fmt.Errorf("firebase error: %w", err)
		}
		result.SuggestedFilename, result.MediaType, result.Contents = cfg.ConfigFilename, "application/json", cfg.ConfigFileContents
	case AppPlatformIOS:
		cfg, err := fb.GetNativeAppConfig(ctx, projectID, app.ResourceName, firebase.AppPlatformIOS)
		if err != nil {
			return FirebaseAppConfig{}, fmt.Errorf("firebase error: %w", err)
		}
		result.SuggestedFilename, result.MediaType, result.Contents = cfg.ConfigFilename, "application/x-plist", cfg.ConfigFileContents
	case AppPlatformWeb:
		cfg, err := fb.GetWebAppConfig(ctx, projectID, app.ResourceName)
		if err != nil {
			return FirebaseAppConfig{}, fmt.Errorf("firebase error: %w", err)
		}
		contents, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return FirebaseAppConfig{}, fmt.Errorf("encode Firebase web app configuration: %w", err)
		}
		result.SuggestedFilename, result.MediaType, result.Contents = "firebase-config.json", "application/json", append(contents, '\n')
	default:
		return FirebaseAppConfig{}, fmt.Errorf("unsupported Firebase app platform %q", app.Platform)
	}
	return result, nil
}

func (s *Core) ReadFirebaseAppConfig(ctx context.Context, projectID, selector string, opts ListFirebaseAppsOptions) (FirebaseAppConfigResult, error) {
	logger := corelog.For("core")
	logger.Debug("read firebase application configuration requested", "project_id", projectID, "selector", selector, "update", opts.Update, "cached_only", opts.CachedOnly)

	index, err := s.ReadFirebaseApps(ctx, projectID, ListFirebaseAppsOptions{ShowDeleted: true, Update: opts.Update, CachedOnly: opts.CachedOnly})
	if err != nil {
		return FirebaseAppConfigResult{}, err
	}
	app, err := resolveFirebaseAppFromList(selector, index.Apps)
	if err != nil {
		return FirebaseAppConfigResult{}, err
	}
	if !ExecutionPolicyFromContext(ctx).ReadLocalState {
		logger.Info("fetch application configuration from firebase", "project_id", projectID, "app_id", app.AppID)
		fb, err := s.firebaseServiceForProject(ctx, projectID)
		if err != nil {
			return FirebaseAppConfigResult{}, err
		}
		cfg, err := fetchFirebaseAppConfig(ctx, fb, projectID, app)
		if err != nil {
			logger.Error("firebase application configuration fetch failed", "project_id", projectID, "app_id", app.AppID, "err", err)
			return FirebaseAppConfigResult{}, err
		}
		return FirebaseAppConfigResult{Config: cfg, Source: AppCacheSourceFirebase}, nil
	}
	if !opts.CachedOnly && index.RefreshError != nil {
		return FirebaseAppConfigResult{}, index.RefreshError
	}
	record, contents, cachedErr := coreconfig.LoadAppConfigCache(projectID, app.AppID)
	var cachedApp FirebaseApp
	if cachedErr == nil {
		cachedErr = json.Unmarshal(record.Payload, &cachedApp)
		if cachedErr == nil && cachedApp.AppID != app.AppID {
			cachedErr = fmt.Errorf("decode app config cache: cached application identity does not match")
		}
	}
	if cachedErr == nil {
		logAppResourceCacheState(logger, "application configuration", projectID, app.AppID, record)
	} else if !errors.Is(cachedErr, os.ErrNotExist) {
		logger.Error("application configuration cache load failed", "project_id", projectID, "app_id", app.AppID, "err", cachedErr)
	}
	if opts.CachedOnly {
		if cachedErr != nil {
			return FirebaseAppConfigResult{}, appCacheLoadError("config", projectID, app.AppID, cachedErr)
		}
		result := FirebaseAppConfigResult{Config: FirebaseAppConfig{App: cachedApp, SuggestedFilename: record.SuggestedFilename, MediaType: record.MediaType, Contents: contents}, Source: combinedAppCacheSource(cacheRecordSource(record), index.Source), CachedAt: record.CachedAt, RefreshError: index.RefreshError}
		logAppResourceCacheServe(logger, "application configuration", projectID, app.AppID, result.Source)
		return result, nil
	}
	if !opts.Update && cachedErr == nil && record.IsFresh(time.Now()) {
		result := FirebaseAppConfigResult{Config: FirebaseAppConfig{App: cachedApp, SuggestedFilename: record.SuggestedFilename, MediaType: record.MediaType, Contents: contents}, Source: AppCacheSourceCache, CachedAt: record.CachedAt, RefreshError: index.RefreshError}
		logAppResourceCacheServe(logger, "application configuration", projectID, app.AppID, result.Source)
		return result, nil
	}
	switch {
	case opts.Update:
		logger.Info("refresh application configuration cache from firebase", "project_id", projectID, "app_id", app.AppID)
	case cachedErr == nil:
		logger.Info("refreshing stale application configuration cache from firebase", "project_id", projectID, "app_id", app.AppID)
	default:
		logger.Warn("fetching application configuration from firebase due to cache miss", "project_id", projectID, "app_id", app.AppID)
	}
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return FirebaseAppConfigResult{}, err
	}
	cfg, err := fetchFirebaseAppConfig(ctx, fb, projectID, app)
	if err != nil {
		logger.Error("firebase application configuration fetch failed", "project_id", projectID, "app_id", app.AppID, "err", err)
		return FirebaseAppConfigResult{}, err
	}
	now := time.Now().UTC()
	result := FirebaseAppConfigResult{Config: cfg, Source: AppCacheSourceFirebase, CachedAt: now, RefreshError: index.RefreshError}
	logger.Info("fetched application configuration from firebase", "project_id", projectID, "app_id", app.AppID)
	if ExecutionPolicyFromContext(ctx).WriteLocalState {
		payload, marshalErr := json.Marshal(cfg.App)
		if marshalErr != nil {
			result.CacheError = marshalErr
		} else {
			result.CacheError = coreconfig.SaveAppConfigCache(projectID, app.AppID, now, cfg.SuggestedFilename, cfg.MediaType, payload, cfg.Contents)
		}
		if result.CacheError != nil {
			logger.Error("save application configuration cache failed", "project_id", projectID, "app_id", app.AppID, "err", result.CacheError)
		} else {
			logger.Info("application configuration cached from firebase", "project_id", projectID, "app_id", app.AppID)
		}
	}
	return result, nil
}

func logAppResourceCacheState(logger *charmlog.Logger, resource, projectID, appID string, record *coreconfig.AppsCacheRecord) {
	if record.IsFresh(time.Now()) {
		logger.Info(resource+" cache fresh", "project_id", projectID, "app_id", appID, "cached_at", record.CachedAt)
		return
	}
	logger.Info(resource+" cache stale", "project_id", projectID, "app_id", appID, "cached_at", record.CachedAt)
}

func logAppResourceCacheServe(logger *charmlog.Logger, resource, projectID, appID string, source AppCacheSource) {
	if source == AppCacheSourceCache {
		logger.Info("serving "+resource+" from fresh cache", "project_id", projectID, "app_id", appID)
		return
	}
	logger.Info("serving "+resource+" from stale cache", "project_id", projectID, "app_id", appID)
}

func cacheRecordSource(record *coreconfig.AppsCacheRecord) AppCacheSource {
	if record.IsFresh(time.Now()) {
		return AppCacheSourceCache
	}
	return AppCacheSourceCacheStale
}

func combinedAppCacheSource(primary, inventory AppCacheSource) AppCacheSource {
	if inventory == AppCacheSourceCacheStale {
		return AppCacheSourceCacheStale
	}
	return primary
}

type AppCacheMissError struct {
	Kind      string
	ProjectID string
	AppID     string
	Err       error
}

func (e *AppCacheMissError) Error() string {
	target := e.ProjectID
	if e.AppID != "" {
		target += "/" + e.AppID
	}
	return fmt.Sprintf("cached app %s for %s was not found", e.Kind, target)
}

func (e *AppCacheMissError) Unwrap() error { return e.Err }

func appCacheLoadError(kind, projectID, appID string, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return &AppCacheMissError{Kind: kind, ProjectID: projectID, AppID: appID, Err: err}
	}
	return err
}

func (s *Core) resolveFirebaseApp(ctx context.Context, projectID, selector string) (*firebase.Service, FirebaseApp, error) {
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, FirebaseApp{}, err
	}
	apps, err := s.ListFirebaseApps(ctx, projectID, ListFirebaseAppsOptions{ShowDeleted: true})
	if err != nil {
		return nil, FirebaseApp{}, err
	}
	app, err := resolveFirebaseAppFromList(selector, apps)
	if err != nil {
		return nil, FirebaseApp{}, err
	}
	return fb, app, nil
}

func resolveFirebaseAppFromList(selector string, apps []FirebaseApp) (FirebaseApp, error) {
	matchFields := []func(FirebaseApp) string{
		func(app FirebaseApp) string { return app.AppID },
		func(app FirebaseApp) string { return app.ResourceName },
		func(app FirebaseApp) string { return app.Namespace },
		func(app FirebaseApp) string { return app.DisplayName },
	}
	for _, field := range matchFields {
		var matches []FirebaseApp
		for _, app := range apps {
			if field(app) == selector {
				matches = append(matches, app)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 {
			return FirebaseApp{}, appSelectionError("ambiguous", selector, matches)
		}
	}
	return FirebaseApp{}, appSelectionError("not_found", selector, apps)
}

func appSelectionError(kind, query string, apps []FirebaseApp) error {
	candidates := make([]AppLookupCandidate, 0, len(apps))
	for _, app := range apps {
		candidates = append(candidates, AppLookupCandidate{Name: app.DisplayName, ID: app.AppID})
	}
	return &AppLookupError{Kind: kind, Query: query, Candidates: candidates}
}

func firebaseAppFromInfo(app firebase.AppInfo) (FirebaseApp, error) {
	platform := AppPlatform("")
	switch app.Platform {
	case firebase.AppPlatformAndroid:
		platform = AppPlatformAndroid
	case firebase.AppPlatformIOS:
		platform = AppPlatformIOS
	case firebase.AppPlatformWeb:
		platform = AppPlatformWeb
	default:
		return FirebaseApp{}, fmt.Errorf("unsupported Firebase app platform %q", app.Platform)
	}
	return FirebaseApp{ResourceName: app.Name, DisplayName: app.DisplayName, Platform: platform, AppID: app.AppID, Namespace: app.Namespace, APIKeyID: app.APIKeyID, State: app.State, ExpireTime: app.ExpireTime}, nil
}

func stringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
