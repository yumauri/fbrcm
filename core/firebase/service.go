package firebase

import (
	"context"
	"net/http"

	"github.com/yumauri/fbrcm/core/config"
	corelog "github.com/yumauri/fbrcm/core/log"
)

// Service holds service state used by the firebase package.
type Service struct {
	// httpClient stores http client for Service.
	httpClient *http.Client
	// quotaProjectID stores optional quota project for client-based Google APIs.
	quotaProjectID string
	// useTargetProjectQuota stores whether project requests may use their target project for quota.
	useTargetProjectQuota bool
}

// NewService constructs service and returns the resulting value or error.
func NewService(ctx context.Context) (*Service, error) {
	return nil, errAuthRequired()
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
