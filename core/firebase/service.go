package firebase

import (
	"context"
	"net/http"

	"github.com/yumauri/fbrcm/core/config"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type Service struct {
	httpClient         *http.Client
	quotaProjectPolicy quotaProjectPolicy
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
	return serviceFromAuthHTTPClientResult(result, environmentQuotaProjectID), nil
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
	return serviceFromAuthHTTPClientResult(result, environmentQuotaProjectID), nil
}

func serviceFromAuthHTTPClientResult(result authHTTPClientResult, environmentQuotaProjectID string) *Service {
	return &Service{
		httpClient: result.client,
		quotaProjectPolicy: quotaProjectPolicy{
			environmentQuotaProjectID: environmentQuotaProjectID,
			credentialQuotaProjectID:  result.credentialQuotaProjectID,
			useTargetProjectQuota:     result.useTargetProjectQuota,
		},
	}
}

func diagnosticAuthHTTPClient(ctx context.Context, auth config.AuthEntry) (authHTTPClientResult, error) {
	switch auth.Type {
	case config.AuthTypeOAuth:
		client, err := diagnosticOAuthHTTPClient(ctx, config.OAuthClientSecretPath(auth), config.OAuthTokenPath(auth))
		return authHTTPClientResult{client: client}, err
	case config.AuthTypeServiceAccount:
		client, err := serviceAccountHTTPClient(ctx, config.ServiceAccountKeyPath(auth))
		return authHTTPClientResult{client: client}, err
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
	return &Service{httpClient: client}
}

func authHTTPClient(ctx context.Context, auth config.AuthEntry, autoOpen bool) (authHTTPClientResult, error) {
	switch auth.Type {
	case config.AuthTypeOAuth:
		client, err := oauthHTTPClient(ctx, config.OAuthClientSecretPath(auth), config.OAuthTokenPath(auth), autoOpen)
		return authHTTPClientResult{client: client}, err
	case config.AuthTypeServiceAccount:
		client, err := serviceAccountHTTPClient(ctx, config.ServiceAccountKeyPath(auth))
		return authHTTPClientResult{client: client}, err
	case config.AuthTypeGCloud:
		client, quotaProjectID, err := gcloudHTTPClient(ctx)
		return authHTTPClientResult{client: client, credentialQuotaProjectID: quotaProjectID, useTargetProjectQuota: true}, err
	default:
		return authHTTPClientResult{}, errAuthRequired()
	}
}

func (s *Service) setQuotaProject(req *http.Request, targetProjectID string) {
	if req == nil {
		return
	}
	quotaProjectID := s.quotaProjectPolicy.projectID(targetProjectID)
	if quotaProjectID == "" {
		return
	}
	req.Header.Set("X-Goog-User-Project", quotaProjectID)
}

func errAuthRequired() error {
	return &authRequiredError{}
}

type authRequiredError struct{}

func (e *authRequiredError) Error() string {
	return "auth identity is required"
}
