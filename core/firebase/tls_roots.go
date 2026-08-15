package firebase

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"

	coreenv "github.com/yumauri/fbrcm/core/env"
	"golang.org/x/oauth2"
)

// withEnvironmentTLSRoots makes SSL_CERT_FILE effective for authenticated
// clients on every platform. Go already consults it on most Unix systems, but
// macOS normally delegates verification to Keychain and ignores the file.
func withEnvironmentTLSRoots(ctx context.Context) (context.Context, error) {
	path := strings.TrimSpace(os.Getenv(coreenv.TLSCertFile))
	if path == "" {
		return ctx, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read SSL_CERT_FILE: %w", err)
	}
	roots, err := x509.SystemCertPool()
	if err != nil {
		roots = x509.NewCertPool()
	}
	if !roots.AppendCertsFromPEM(raw) {
		return nil, fmt.Errorf("SSL_CERT_FILE contains no PEM certificates")
	}
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("default HTTP transport has type %T", http.DefaultTransport)
	}
	transport := base.Clone()
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{RootCAs: roots}
	} else {
		transport.TLSClientConfig = transport.TLSClientConfig.Clone()
		transport.TLSClientConfig.RootCAs = roots
	}
	client := *http.DefaultClient
	client.Transport = transport
	return context.WithValue(ctx, oauth2.HTTPClient, &client), nil
}
