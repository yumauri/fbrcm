package firebase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
)

const (
	offlineProbeURL     = "https://firebaseremoteconfig.googleapis.com/"
	offlineProbeTimeout = 2 * time.Second
)

// ErrOffline is returned when network access is suppressed by offline mode.
var ErrOffline = errors.New("offline mode: request suppressed")

var offlineEnabled atomic.Bool

// InitOfflineMode initializes offline mode from env or a connectivity probe.
func InitOfflineMode() {
	raw, envSet := os.LookupEnv(env.Offline)
	initOfflineMode(raw, envSet, defaultConnectivityProbe)
}

func initOfflineMode(raw string, envSet bool, probe func() error) {
	logger := corelog.For("firebase.offline")
	if envSet {
		offlineEnabled.Store(true)
		logger.Warn("offline mode enabled by environment", "env", env.Offline, "value", raw)
		return
	}

	if err := probe(); err != nil {
		offlineEnabled.Store(true)
		logger.Warn("offline mode enabled after connectivity check failed", "url", offlineProbeURL, "timeout", offlineProbeTimeout.String(), "err", err)
		return
	}
	offlineEnabled.Store(false)
	logger.Debug("connectivity check passed", "url", offlineProbeURL)
}

func defaultConnectivityProbe() error {
	ctx, cancel := context.WithTimeout(context.Background(), offlineProbeTimeout)
	defer cancel()

	client := &http.Client{Transport: http.DefaultTransport}
	return probeConnectivity(ctx, client, offlineProbeURL)
}

func probeConnectivity(ctx context.Context, client *http.Client, endpoint string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create connectivity request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send connectivity request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// IsOffline returns true when network requests must be suppressed.
func IsOffline() bool {
	return offlineEnabled.Load()
}

// SetOfflineMode sets offline mode directly, mostly for tests.
func SetOfflineMode(enabled bool) {
	offlineEnabled.Store(enabled)
}
