package firebase

import "net/http"

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

func wrapAuthHTTPClient(client *http.Client) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}
	client.Transport = newResilientTransport(client.Transport)
	return client
}
