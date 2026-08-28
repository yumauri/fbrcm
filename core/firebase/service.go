package firebase

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/yumauri/fbrcm/core/config"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type Service struct {
	httpClient         *http.Client
	quotaProjectPolicy quotaProjectPolicy
	requestController  *RequestController
}

type authHTTPClientResult struct {
	client                   *http.Client
	credentialQuotaProjectID string
	useTargetProjectQuota    bool
}

// NewServiceForAuth constructs service for auth entry with options.
func NewServiceForAuth(ctx context.Context, auth config.AuthEntry, autoOpen bool) (*Service, error) {
	logger := corelog.For("firebase")
	logger.Debug("create firebase service", "auth_id", auth.ID, "auth_type", auth.Type)
	ctx, err := withEnvironmentTLSRoots(ctx)
	if err != nil {
		return nil, err
	}

	environmentQuotaProjectID, err := environmentQuotaProjectID()
	if err != nil {
		return nil, err
	}
	result, err := authHTTPClient(ctx, auth, autoOpen)
	if err != nil {
		logger.Error("create firebase http client failed", "err", err)
		return nil, err
	}

	logger.Debug("firebase service ready")
	service := serviceFromAuthHTTPClientResult(result, environmentQuotaProjectID, auth.QuotaProjectID)
	service.requestController = requestControllerFromContext(ctx)
	return service, nil
}

// NewServiceWithAccessToken constructs a service backed by one static Google
// OAuth 2.0 access token. The token is used only in memory and is neither
// refreshed nor persisted.
func NewServiceWithAccessToken(ctx context.Context, accessToken string) (*Service, error) {
	ctx, err := withEnvironmentTLSRoots(ctx)
	if err != nil {
		return nil, err
	}
	environmentQuotaProjectID, err := environmentQuotaProjectID()
	if err != nil {
		return nil, err
	}
	client, err := accessTokenHTTPClient(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	service := serviceFromAuthHTTPClientResult(authHTTPClientResult{client: client, useTargetProjectQuota: true}, environmentQuotaProjectID, "")
	service.requestController = requestControllerFromContext(ctx)
	return service, nil
}

// NewDiagnosticServiceForAuth constructs a service without starting an
// interactive OAuth authorization flow or persisting refreshed credentials.
func NewDiagnosticServiceForAuth(ctx context.Context, auth config.AuthEntry) (*Service, error) {
	ctx, err := withEnvironmentTLSRoots(ctx)
	if err != nil {
		return nil, err
	}
	environmentQuotaProjectID, err := environmentQuotaProjectID()
	if err != nil {
		return nil, err
	}
	result, err := diagnosticAuthHTTPClient(ctx, auth)
	if err != nil {
		return nil, err
	}
	service := serviceFromAuthHTTPClientResult(result, environmentQuotaProjectID, auth.QuotaProjectID)
	service.requestController = requestControllerFromContext(ctx)
	return service, nil
}

func serviceFromAuthHTTPClientResult(result authHTTPClientResult, environmentQuotaProjectID, authQuotaProjectID string) *Service {
	return &Service{
		httpClient:        result.client,
		requestController: defaultRequestController,
		quotaProjectPolicy: quotaProjectPolicy{
			environmentQuotaProjectID: environmentQuotaProjectID,
			authQuotaProjectID:        strings.TrimSpace(authQuotaProjectID),
			credentialQuotaProjectID:  result.credentialQuotaProjectID,
			useTargetProjectQuota:     result.useTargetProjectQuota,
		},
	}
}

// ResolveQuotaProjectForAuth evaluates the same quota policy used by live
// services without exchanging an OAuth token or sending a Firebase or Cloud
// Resource Manager request. Resolving gcloud ADC may probe the metadata server.
func ResolveQuotaProjectForAuth(ctx context.Context, auth config.AuthEntry, projectQuotaProjectID, targetProjectID string) (QuotaProjectSelection, error) {
	environmentQuotaProjectID, err := environmentQuotaProjectID()
	if err != nil {
		return QuotaProjectSelection{Source: QuotaProjectSourceUnresolved}, err
	}
	policy := quotaProjectPolicy{
		environmentQuotaProjectID: environmentQuotaProjectID,
		projectQuotaProjectID:     strings.TrimSpace(projectQuotaProjectID),
		authQuotaProjectID:        strings.TrimSpace(auth.QuotaProjectID),
		useTargetProjectQuota:     true,
	}
	if policy.environmentQuotaProjectID != "" || policy.projectQuotaProjectID != "" || policy.authQuotaProjectID != "" {
		return policy.selectProject(targetProjectID)
	}
	credentialQuotaProjectID := ""
	if auth.Type == config.AuthTypeGCloud {
		_, credentialQuotaProjectID, err = gcloudHTTPClient(ctx)
		if err != nil {
			return QuotaProjectSelection{Source: QuotaProjectSourceUnresolved}, err
		}
	}
	policy.credentialQuotaProjectID = credentialQuotaProjectID
	return policy.selectProject(targetProjectID)
}

func diagnosticAuthHTTPClient(ctx context.Context, auth config.AuthEntry) (authHTTPClientResult, error) {
	switch auth.Type {
	case config.AuthTypeOAuth:
		client, err := diagnosticOAuthHTTPClient(ctx, config.OAuthClientSecretPath(auth), config.OAuthTokenPath(auth))
		return authHTTPClientResult{client: client, useTargetProjectQuota: true}, err
	case config.AuthTypeServiceAccount:
		client, err := serviceAccountHTTPClient(ctx, config.ServiceAccountKeyPath(auth))
		return authHTTPClientResult{client: client, useTargetProjectQuota: true}, err
	case config.AuthTypeGCloud:
		client, quotaProjectID, err := gcloudHTTPClient(ctx)
		return authHTTPClientResult{client: client, credentialQuotaProjectID: quotaProjectID, useTargetProjectQuota: true}, err
	default:
		return authHTTPClientResult{}, errAuthRequired()
	}
}

// NewServiceWithHTTPClient constructs a Service that sends API requests with client.
// It exists for tests that stub Firebase HTTP responses.
func NewServiceWithHTTPClient(client *http.Client) *Service {
	if client == nil {
		client = http.DefaultClient
	}
	return &Service{
		httpClient:        client,
		requestController: defaultRequestController,
		quotaProjectPolicy: quotaProjectPolicy{
			useTargetProjectQuota: true,
		},
	}
}

func authHTTPClient(ctx context.Context, auth config.AuthEntry, autoOpen bool) (authHTTPClientResult, error) {
	switch auth.Type {
	case config.AuthTypeOAuth:
		client, err := oauthHTTPClient(ctx, config.OAuthClientSecretPath(auth), config.OAuthTokenPath(auth), autoOpen)
		return authHTTPClientResult{client: client, useTargetProjectQuota: true}, err
	case config.AuthTypeServiceAccount:
		client, err := serviceAccountHTTPClient(ctx, config.ServiceAccountKeyPath(auth))
		return authHTTPClientResult{client: client, useTargetProjectQuota: true}, err
	case config.AuthTypeGCloud:
		client, quotaProjectID, err := gcloudHTTPClient(ctx)
		return authHTTPClientResult{client: client, credentialQuotaProjectID: quotaProjectID, useTargetProjectQuota: true}, err
	default:
		return authHTTPClientResult{}, errAuthRequired()
	}
}

// WithQuotaProjectOverride returns an immutable project-scoped service view.
func (s *Service) WithQuotaProjectOverride(quotaProjectID string) *Service {
	if s == nil {
		return nil
	}
	clone := *s
	clone.quotaProjectPolicy.projectQuotaProjectID = strings.TrimSpace(quotaProjectID)
	return &clone
}

// QuotaProject resolves the quota project without sending a request.
func (s *Service) QuotaProject(targetProjectID string) (QuotaProjectSelection, error) {
	if s == nil {
		return QuotaProjectSelection{Source: QuotaProjectSourceUnresolved}, &QuotaProjectRequiredError{TargetProjectID: strings.TrimSpace(targetProjectID)}
	}
	return s.quotaProjectPolicy.selectProject(targetProjectID)
}

func (s *Service) setQuotaProject(req *http.Request, targetProjectID string) (QuotaProjectSelection, error) {
	if req == nil {
		return QuotaProjectSelection{Source: QuotaProjectSourceUnresolved}, fmt.Errorf("quota project request is nil")
	}
	selection, err := s.QuotaProject(targetProjectID)
	if err != nil {
		return selection, err
	}
	req.Header.Set("X-Goog-User-Project", selection.ProjectID)
	return selection, nil
}

func errAuthRequired() error {
	return &authRequiredError{}
}

type authRequiredError struct{}

func (e *authRequiredError) Error() string {
	return "auth identity is required"
}
