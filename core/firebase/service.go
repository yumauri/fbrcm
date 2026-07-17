package firebase

import (
	"context"
	"net/http"

	"github.com/yumauri/fbrcm/core/config"
	corelog "github.com/yumauri/fbrcm/core/log"
)

type Service struct {
	httpClient *http.Client
	// quotaProjectID stores optional quota project for client-based Google APIs.
	quotaProjectID        string
	useTargetProjectQuota bool
}

// NewServiceForAuth constructs service for auth entry with options.
func NewServiceForAuth(ctx context.Context, auth config.AuthEntry, autoOpen bool) (*Service, error) {
	logger := corelog.For("firebase")
	logger.Debug("create firebase service", "auth_id", auth.ID, "auth_type", auth.Type)

	client, quotaProjectID, useTargetProjectQuota, err := authHTTPClient(ctx, auth, autoOpen)
	if err != nil {
		logger.Error("create firebase http client failed", "err", err)
		return nil, err
	}

	logger.Debug("firebase service ready")
	return &Service{
		httpClient:            client,
		quotaProjectID:        quotaProjectID,
		useTargetProjectQuota: useTargetProjectQuota,
	}, nil
}

// NewDiagnosticServiceForAuth constructs a service without starting an
// interactive OAuth authorization flow or persisting refreshed credentials.
func NewDiagnosticServiceForAuth(ctx context.Context, auth config.AuthEntry) (*Service, error) {
	client, quotaProjectID, useTargetProjectQuota, err := diagnosticAuthHTTPClient(ctx, auth)
	if err != nil {
		return nil, err
	}
	return &Service{
		httpClient:            client,
		quotaProjectID:        quotaProjectID,
		useTargetProjectQuota: useTargetProjectQuota,
	}, nil
}

func diagnosticAuthHTTPClient(ctx context.Context, auth config.AuthEntry) (*http.Client, string, bool, error) {
	switch auth.Type {
	case config.AuthTypeOAuth:
		client, err := diagnosticOAuthHTTPClient(ctx, config.OAuthClientSecretPath(auth), config.OAuthTokenPath(auth))
		return client, "", false, err
	case config.AuthTypeServiceAccount:
		client, err := serviceAccountHTTPClient(ctx, config.ServiceAccountKeyPath(auth))
		return client, "", false, err
	case config.AuthTypeGCloud:
		client, quotaProjectID, err := gcloudHTTPClient(ctx)
		return client, quotaProjectID, true, err
	default:
		return nil, "", false, errAuthRequired()
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

func authHTTPClient(ctx context.Context, auth config.AuthEntry, autoOpen bool) (*http.Client, string, bool, error) {
	switch auth.Type {
	case config.AuthTypeOAuth:
		client, err := oauthHTTPClient(ctx, config.OAuthClientSecretPath(auth), config.OAuthTokenPath(auth), autoOpen)
		return client, "", false, err
	case config.AuthTypeServiceAccount:
		client, err := serviceAccountHTTPClient(ctx, config.ServiceAccountKeyPath(auth))
		return client, "", false, err
	case config.AuthTypeGCloud:
		client, quotaProjectID, err := gcloudHTTPClient(ctx)
		return client, quotaProjectID, true, err
	default:
		return nil, "", false, errAuthRequired()
	}
}

func (s *Service) setQuotaProject(req *http.Request, targetProjectID string) {
	if req == nil {
		return
	}
	quotaProjectID := s.quotaProjectID
	if quotaProjectID == "" && s.useTargetProjectQuota {
		quotaProjectID = targetProjectID
	}
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
