package firebase

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	corelog "github.com/yumauri/fbrcm/core/log"
)

func gcloudHTTPClient(ctx context.Context) (*http.Client, string, error) {
	logger := corelog.For("firebase")
	logger.Info("load gcloud application default credentials")

	credentials, err := google.FindDefaultCredentials(ctx, cloudPlatformScope)
	if err != nil {
		logger.Error("load gcloud application default credentials failed", "err", err)
		return nil, "", setupAuthenticationError("gcloud", "discover_credentials", fmt.Errorf("loading gcloud application default credentials: %w; run `gcloud auth application-default login`", err))
	}
	tokenSource := authenticationTokenSource{base: credentials.TokenSource, authType: "gcloud", operation: "token_exchange"}
	client := oauth2.NewClient(ctx, tokenSource)

	quotaProjectID, err := credentialQuotaProjectID(credentials.JSON)
	if err != nil {
		logger.Error("load gcloud ADC quota project failed", "err", err)
		return nil, "", err
	}
	logger.Debug("gcloud application default credentials http client ready", "quota_project_id", quotaProjectID != "")
	return wrapAuthHTTPClient(ctx, client), quotaProjectID, nil
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
