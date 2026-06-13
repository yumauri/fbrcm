package firebase

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"

	corelog "github.com/yumauri/fbrcm/core/log"
)

func serviceAccountHTTPClient(ctx context.Context, keyPath string) (*http.Client, error) {
	logger := corelog.For("firebase")
	logger.Info("load service account key", "path", keyPath)

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		logger.Error("read service account key failed", "path", keyPath, "err", err)
		return nil, fmt.Errorf("reading service account key: %w", err)
	}

	cfg, err := google.JWTConfigFromJSON(keyData, cloudPlatformScope)
	if err != nil {
		logger.Error("parse service account key failed", "path", keyPath, "err", err)
		return nil, fmt.Errorf("parsing service account key: %w", err)
	}

	logger.Debug("service account http client ready")
	return wrapAuthHTTPClient(cfg.Client(ctx)), nil
}
