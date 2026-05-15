package firebase

import (
	"context"
	"net/http"

	corelog "github.com/yumauri/fbrcm/core/log"
)

// Service holds service state used by the firebase package.
type Service struct {
	// httpClient stores http client for Service.
	httpClient *http.Client
}

// NewService constructs service and returns the resulting value or error.
func NewService(ctx context.Context) (*Service, error) {
	return NewServiceWithOptions(ctx, true)
}

// NewServiceWithOptions constructs service with options and returns the resulting value or error.
func NewServiceWithOptions(ctx context.Context, autoOpen bool) (*Service, error) {
	logger := corelog.For("firebase")
	logger.Debug("create firebase service")

	client, err := oauthHTTPClient(ctx, autoOpen)
	if err != nil {
		logger.Error("create firebase http client failed", "err", err)
		return nil, err
	}

	logger.Debug("firebase service ready")
	return &Service{
		httpClient: client,
	}, nil
}
