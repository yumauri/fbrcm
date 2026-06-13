package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/oauth2/google"

	corelog "github.com/yumauri/fbrcm/core/log"
)

func gcloudHTTPClient(ctx context.Context) (*http.Client, string, error) {
	logger := corelog.For("firebase")
	logger.Info("load gcloud application default credentials")

	client, err := google.DefaultClient(ctx, cloudPlatformScope)
	if err != nil {
		logger.Error("load gcloud application default credentials failed", "err", err)
		return nil, "", fmt.Errorf("loading gcloud application default credentials: %w; run `gcloud auth application-default login`", err)
	}

	quotaProjectID := adcQuotaProjectID()
	logger.Debug("gcloud application default credentials http client ready", "quota_project_id", quotaProjectID != "")
	return wrapAuthHTTPClient(client), quotaProjectID, nil
}

func adcQuotaProjectID() string {
	path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if strings.TrimSpace(path) == "" {
		path = wellKnownADCFile()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var payload struct {
		QuotaProjectID string `json:"quota_project_id"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.QuotaProjectID)
}

func wellKnownADCFile() string {
	const name = "application_default_credentials.json"
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud", name)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "gcloud", name)
	}
	return filepath.Join(home, ".config", "gcloud", name)
}
