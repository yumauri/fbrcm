package firebase

import (
	"context"
	"net/http"
	"strings"
)

const (
	cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
	applicationURL     = "https://fbrcm.yumaa.dev"
	developmentVersion = "dev"
)

type applicationVersionContextKey struct{}

// WithApplicationVersion identifies the running fbrcm build to authenticated
// Firebase and Google REST API clients created with ctx. Empty values use
// "dev".
func WithApplicationVersion(ctx context.Context, version string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, applicationVersionContextKey{}, normalizeApplicationVersion(version))
}

func normalizeApplicationVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return developmentVersion
	}
	return version
}

func applicationUserAgent(ctx context.Context) string {
	version := developmentVersion
	if ctx != nil {
		if configured, ok := ctx.Value(applicationVersionContextKey{}).(string); ok {
			version = normalizeApplicationVersion(configured)
		}
	}
	return "fbrcm/" + version + " (" + applicationURL + ")"
}

func wrapAuthHTTPClient(ctx context.Context, client *http.Client) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}
	client.Transport = newResilientTransportWithUserAgent(
		client.Transport,
		requestControllerFromContext(ctx),
		applicationUserAgent(ctx),
	)
	return client
}
