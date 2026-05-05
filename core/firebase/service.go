package firebase

import (
	"context"
	"net/http"

	corelog "fbrcm/core/log"
)

type Service struct {
	httpClient *http.Client
}

func NewService(ctx context.Context) (*Service, error) {
	return NewServiceWithOptions(ctx, true)
}

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
