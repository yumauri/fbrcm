package firebase

import (
	"context"
	"net/http"
)

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

func wrapAuthHTTPClient(ctx context.Context, client *http.Client) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}
	client.Transport = newResilientTransportWithController(client.Transport, requestControllerFromContext(ctx))
	return client
}
