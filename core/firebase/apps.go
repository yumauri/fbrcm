package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	corelog "github.com/yumauri/fbrcm/core/log"
)

const firebaseManagementBaseURL = "https://firebase.googleapis.com/v1beta1/"

type AppPlatform string

const (
	AppPlatformAndroid AppPlatform = "ANDROID"
	AppPlatformIOS     AppPlatform = "IOS"
	AppPlatformWeb     AppPlatform = "WEB"
)

type AppInfo struct {
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName,omitempty"`
	Platform    AppPlatform `json:"platform"`
	AppID       string      `json:"appId"`
	Namespace   string      `json:"namespace,omitempty"`
	APIKeyID    string      `json:"apiKeyId,omitempty"`
	State       string      `json:"state,omitempty"`
	ExpireTime  string      `json:"expireTime,omitempty" contract:"format=date-time"`
}

type AppsPage struct {
	Apps          []AppInfo `json:"apps"`
	NextPageToken string    `json:"nextPageToken,omitempty"`
}

type ListAppsOptions struct {
	PageSize    int
	PageToken   string
	ShowDeleted bool
}

type AndroidApp struct {
	Name         string   `json:"name"`
	AppID        string   `json:"appId"`
	DisplayName  string   `json:"displayName,omitempty"`
	ProjectID    string   `json:"projectId"`
	PackageName  string   `json:"packageName"`
	APIKeyID     string   `json:"apiKeyId,omitempty"`
	State        string   `json:"state,omitempty"`
	SHA1Hashes   []string `json:"sha1Hashes,omitempty"`
	SHA256Hashes []string `json:"sha256Hashes,omitempty"`
	ExpireTime   string   `json:"expireTime,omitempty" contract:"format=date-time"`
	ETag         string   `json:"etag,omitempty"`
}

type IOSApp struct {
	Name        string `json:"name"`
	AppID       string `json:"appId"`
	DisplayName string `json:"displayName,omitempty"`
	ProjectID   string `json:"projectId"`
	BundleID    string `json:"bundleId"`
	AppStoreID  string `json:"appStoreId,omitempty"`
	TeamID      string `json:"teamId,omitempty"`
	APIKeyID    string `json:"apiKeyId,omitempty"`
	State       string `json:"state,omitempty"`
	ExpireTime  string `json:"expireTime,omitempty" contract:"format=date-time"`
	ETag        string `json:"etag,omitempty"`
}

type WebApp struct {
	Name        string   `json:"name"`
	AppID       string   `json:"appId"`
	DisplayName string   `json:"displayName,omitempty"`
	ProjectID   string   `json:"projectId"`
	AppURLs     []string `json:"appUrls,omitempty"`
	WebID       string   `json:"webId,omitempty"`
	APIKeyID    string   `json:"apiKeyId,omitempty"`
	State       string   `json:"state,omitempty"`
	ExpireTime  string   `json:"expireTime,omitempty" contract:"format=date-time"`
	ETag        string   `json:"etag,omitempty"`
}

type NativeAppConfig struct {
	ConfigFilename     string `json:"configFilename"`
	ConfigFileContents []byte `json:"configFileContents"`
}

type WebAppConfig struct {
	ProjectID         string `json:"projectId,omitempty"`
	AppID             string `json:"appId,omitempty"`
	DatabaseURL       string `json:"databaseURL,omitempty"`
	StorageBucket     string `json:"storageBucket,omitempty"`
	LocationID        string `json:"locationId,omitempty"`
	APIKey            string `json:"apiKey,omitempty"`
	AuthDomain        string `json:"authDomain,omitempty"`
	MessagingSenderID string `json:"messagingSenderId,omitempty"`
	MeasurementID     string `json:"measurementId,omitempty"`
	ProjectNumber     string `json:"projectNumber,omitempty"`
}

type AppResourceError struct {
	Resource string
	Platform AppPlatform
	Err      error
}

func (e *AppResourceError) Error() string { return e.Err.Error() }
func (e *AppResourceError) Unwrap() error { return e.Err }

func (s *Service) ListApps(ctx context.Context, projectID string, opts ListAppsOptions) (AppsPage, error) {
	values := url.Values{}
	if opts.PageSize > 0 {
		values.Set("pageSize", fmt.Sprint(opts.PageSize))
	}
	if strings.TrimSpace(opts.PageToken) != "" {
		values.Set("pageToken", opts.PageToken)
	}
	if opts.ShowDeleted {
		values.Set("showDeleted", "true")
	}
	endpoint := firebaseManagementBaseURL + "projects/" + url.PathEscape(projectID) + ":searchApps"
	if encoded := values.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	var page AppsPage
	if err := s.getManagementJSON(ctx, projectID, "list_apps", endpoint, &page); err != nil {
		return AppsPage{}, err
	}
	return page, nil
}

func (s *Service) GetAndroidApp(ctx context.Context, targetProjectID, resource string) (AndroidApp, error) {
	var app AndroidApp
	if err := s.getAppResourceJSON(ctx, targetProjectID, resource, AppPlatformAndroid, "get_android_app", &app); err != nil {
		return AndroidApp{}, err
	}
	return app, nil
}

func (s *Service) GetIOSApp(ctx context.Context, targetProjectID, resource string) (IOSApp, error) {
	var app IOSApp
	if err := s.getAppResourceJSON(ctx, targetProjectID, resource, AppPlatformIOS, "get_ios_app", &app); err != nil {
		return IOSApp{}, err
	}
	return app, nil
}

func (s *Service) GetWebApp(ctx context.Context, targetProjectID, resource string) (WebApp, error) {
	var app WebApp
	if err := s.getAppResourceJSON(ctx, targetProjectID, resource, AppPlatformWeb, "get_web_app", &app); err != nil {
		return WebApp{}, err
	}
	return app, nil
}

func (s *Service) GetNativeAppConfig(ctx context.Context, targetProjectID, resource string, platform AppPlatform) (NativeAppConfig, error) {
	if platform != AppPlatformAndroid && platform != AppPlatformIOS {
		return NativeAppConfig{}, &AppResourceError{Resource: resource, Platform: platform, Err: fmt.Errorf("platform %q does not have a native app configuration", platform)}
	}
	resource, err := validateAppResource(targetProjectID, resource, platform)
	if err != nil {
		return NativeAppConfig{}, err
	}
	var cfg NativeAppConfig
	if err := s.getManagementJSON(ctx, targetProjectID, "get_"+strings.ToLower(string(platform))+"_app_config", firebaseManagementBaseURL+resource+"/config", &cfg); err != nil {
		return NativeAppConfig{}, err
	}
	return cfg, nil
}

func (s *Service) GetWebAppConfig(ctx context.Context, targetProjectID, resource string) (WebAppConfig, error) {
	resource, err := validateAppResource(targetProjectID, resource, AppPlatformWeb)
	if err != nil {
		return WebAppConfig{}, err
	}
	var cfg WebAppConfig
	if err := s.getManagementJSON(ctx, targetProjectID, "get_web_app_config", firebaseManagementBaseURL+resource+"/config", &cfg); err != nil {
		return WebAppConfig{}, err
	}
	return cfg, nil
}

func (s *Service) getAppResourceJSON(ctx context.Context, targetProjectID, resource string, platform AppPlatform, operation string, destination any) error {
	resource, err := validateAppResource(targetProjectID, resource, platform)
	if err != nil {
		return err
	}
	return s.getManagementJSON(ctx, targetProjectID, operation, firebaseManagementBaseURL+resource, destination)
}

func (s *Service) getManagementJSON(ctx context.Context, targetProjectID, operation, endpoint string, destination any) error {
	logger := corelog.For("firebase")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create Firebase Management request: %w", err)
	}
	if _, err := s.setQuotaProject(req, targetProjectID); err != nil {
		return fmt.Errorf("select Firebase Management quota project: %w", err)
	}
	logHTTPRequest(logger.With("project_id", targetProjectID, "operation", operation), req)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("firebase management request: %w", err)
	}
	logHTTPResponse(logger.With("project_id", targetProjectID, "operation", operation), req, resp)
	body, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		return fmt.Errorf("read Firebase Management response: %w", readErr)
	}
	if resp.StatusCode != http.StatusOK {
		return newAPIError("firebase_management", operation, resp, body)
	}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(destination); err != nil {
		return fmt.Errorf("decode Firebase Management response: %w", err)
	}
	return nil
}

func validateAppResource(projectID, resource string, platform AppPlatform) (string, error) {
	resource = strings.TrimPrefix(strings.TrimSpace(resource), "/")
	collection := ""
	switch platform {
	case AppPlatformAndroid:
		collection = "androidApps"
	case AppPlatformIOS:
		collection = "iosApps"
	case AppPlatformWeb:
		collection = "webApps"
	default:
		return "", &AppResourceError{Resource: resource, Platform: platform, Err: fmt.Errorf("unsupported Firebase app platform %q", platform)}
	}
	parts := strings.Split(resource, "/")
	// The API may canonicalize the project segment to either a project ID or a
	// project number. Callers resolve this resource only from searchApps results
	// for targetProjectID, so validate the shape and platform collection here.
	if len(parts) != 4 || parts[0] != "projects" || strings.TrimSpace(parts[1]) == "" || parts[2] != collection || strings.TrimSpace(parts[3]) == "" {
		return "", &AppResourceError{Resource: resource, Platform: platform, Err: fmt.Errorf("invalid %s Firebase app resource %q for project %s", strings.ToLower(string(platform)), resource, projectID)}
	}
	return strings.Join(parts, "/"), nil
}
