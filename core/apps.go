package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/yumauri/fbrcm/core/firebase"
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
	var result []FirebaseApp
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

func (s *Core) GetFirebaseApp(ctx context.Context, projectID, selector string) (FirebaseAppDetails, error) {
	fb, app, err := s.resolveFirebaseApp(ctx, projectID, selector)
	if err != nil {
		return FirebaseAppDetails{}, err
	}
	details := FirebaseAppDetails{FirebaseApp: app, ProjectID: projectID, AppURLs: []string{}, SHA1Hashes: []string{}, SHA256Hashes: []string{}}
	switch app.Platform {
	case AppPlatformAndroid:
		value, err := fb.GetAndroidApp(ctx, projectID, app.ResourceName)
		if err != nil {
			return FirebaseAppDetails{}, fmt.Errorf("firebase error: %w", err)
		}
		details.DisplayName, details.AppID, details.APIKeyID, details.State, details.ExpireTime, details.ETag = value.DisplayName, value.AppID, value.APIKeyID, value.State, value.ExpireTime, value.ETag
		details.PackageName = stringPointer(value.PackageName)
		details.Namespace = value.PackageName
		details.SHA1Hashes, details.SHA256Hashes = nonNilStrings(value.SHA1Hashes), nonNilStrings(value.SHA256Hashes)
	case AppPlatformIOS:
		value, err := fb.GetIOSApp(ctx, projectID, app.ResourceName)
		if err != nil {
			return FirebaseAppDetails{}, fmt.Errorf("firebase error: %w", err)
		}
		details.DisplayName, details.AppID, details.APIKeyID, details.State, details.ExpireTime, details.ETag = value.DisplayName, value.AppID, value.APIKeyID, value.State, value.ExpireTime, value.ETag
		details.BundleID, details.AppStoreID, details.TeamID = stringPointer(value.BundleID), stringPointer(value.AppStoreID), stringPointer(value.TeamID)
		details.Namespace = value.BundleID
	case AppPlatformWeb:
		value, err := fb.GetWebApp(ctx, projectID, app.ResourceName)
		if err != nil {
			return FirebaseAppDetails{}, fmt.Errorf("firebase error: %w", err)
		}
		details.DisplayName, details.AppID, details.APIKeyID, details.State, details.ExpireTime, details.ETag = value.DisplayName, value.AppID, value.APIKeyID, value.State, value.ExpireTime, value.ETag
		details.WebID, details.AppURLs = stringPointer(value.WebID), nonNilStrings(value.AppURLs)
	}
	return details, nil
}

func (s *Core) GetFirebaseAppConfig(ctx context.Context, projectID, selector string) (FirebaseAppConfig, error) {
	fb, app, err := s.resolveFirebaseApp(ctx, projectID, selector)
	if err != nil {
		return FirebaseAppConfig{}, err
	}
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

func (s *Core) resolveFirebaseApp(ctx context.Context, projectID, selector string) (*firebase.Service, FirebaseApp, error) {
	fb, err := s.firebaseServiceForProject(ctx, projectID)
	if err != nil {
		return nil, FirebaseApp{}, err
	}
	apps, err := s.ListFirebaseApps(ctx, projectID, ListFirebaseAppsOptions{ShowDeleted: true})
	if err != nil {
		return nil, FirebaseApp{}, err
	}
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
			return fb, matches[0], nil
		}
		if len(matches) > 1 {
			return nil, FirebaseApp{}, appSelectionError("ambiguous", selector, matches)
		}
	}
	return nil, FirebaseApp{}, appSelectionError("not_found", selector, apps)
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
