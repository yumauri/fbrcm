package firebase

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestEnvironmentTLSRootsTrustsConfiguredCertificate(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	path := writeServerCertificate(t, server)
	t.Setenv("SSL_CERT_FILE", path)
	ctx, err := withEnvironmentTLSRoots(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "fixture"}))
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %s", response.Status)
	}
}

func TestNewServiceWithAccessTokenUsesEnvironmentTLSRoots(t *testing.T) {
	const accessToken = "fixture-access-token"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if got := req.Header.Get("Authorization"); got != "Bearer "+accessToken {
			t.Errorf("Authorization = %q, want bearer token", got)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	t.Setenv("SSL_CERT_FILE", writeServerCertificate(t, server))
	service, err := NewServiceWithAccessToken(context.Background(), accessToken)
	if err != nil {
		t.Fatal(err)
	}
	response, err := service.httpClient.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %s", response.Status)
	}
}

func TestEnvironmentTLSRootsRejectsInvalidPEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(path, []byte("not a certificate"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SSL_CERT_FILE", path)
	_, err := withEnvironmentTLSRoots(context.Background())
	if err == nil || !strings.Contains(err.Error(), "contains no PEM certificates") {
		t.Fatalf("error = %v", err)
	}
}

func writeServerCertificate(t *testing.T, server *httptest.Server) string {
	t.Helper()
	certificate, err := x509.ParseCertificate(server.TLS.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw}), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
