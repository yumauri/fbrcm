package firebase

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/env"
)

func TestOfflineModeToggle(t *testing.T) {
	SetOfflineMode(true)
	t.Cleanup(func() { SetOfflineMode(false) })
	if !IsOffline() {
		t.Fatal("IsOffline = false, want true")
	}
	SetOfflineMode(false)
	if IsOffline() {
		t.Fatal("IsOffline = true, want false after reset")
	}
}

func TestInitOfflineModeFromEnv(t *testing.T) {
	t.Setenv(env.Offline, "1")
	InitOfflineMode()
	t.Cleanup(func() { SetOfflineMode(false) })
	if !IsOffline() {
		t.Fatal("InitOfflineMode with env should enable offline mode")
	}
}

func TestInitOfflineModeFromConnectivityProbe(t *testing.T) {
	t.Cleanup(func() { SetOfflineMode(false) })

	t.Run("online", func(t *testing.T) {
		initOfflineMode("", false, func() error { return nil })
		if IsOffline() {
			t.Fatal("successful connectivity probe enabled offline mode")
		}
	})

	t.Run("offline", func(t *testing.T) {
		initOfflineMode("", false, func() error { return errors.New("unreachable") })
		if !IsOffline() {
			t.Fatal("failed connectivity probe did not enable offline mode")
		}
	})
}

func TestProbeConnectivityUsesHTTPHead(t *testing.T) {
	var method string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	if err := probeConnectivity(t.Context(), server.Client(), server.URL); err != nil {
		t.Fatalf("probeConnectivity = %v", err)
	}
	if method != http.MethodHead {
		t.Fatalf("probe method = %q, want HEAD", method)
	}
}

func TestServiceAccountHTTPClientMissingKey(t *testing.T) {
	_, err := serviceAccountHTTPClient(t.Context(), filepath.Join(t.TempDir(), "missing.json"))
	if err == nil || !strings.Contains(err.Error(), "reading service account key") {
		t.Fatalf("serviceAccountHTTPClient missing = %v", err)
	}
}

func TestServiceAccountHTTPClientMalformedKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := serviceAccountHTTPClient(t.Context(), path)
	if err == nil || !strings.Contains(err.Error(), "parsing service account key") {
		t.Fatalf("serviceAccountHTTPClient malformed = %v", err)
	}
}
