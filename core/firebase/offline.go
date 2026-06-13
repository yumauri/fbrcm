package firebase

import (
	"errors"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/yumauri/fbrcm/core/env"
	corelog "github.com/yumauri/fbrcm/core/log"
)

const (
	offlineProbeAddress = "firebaseremoteconfig.googleapis.com:443"
	offlineProbeTimeout = 2 * time.Second
)

// ErrOffline is returned when network access is suppressed by offline mode.
var ErrOffline = errors.New("offline mode: request suppressed")

var offlineEnabled atomic.Bool

// InitOfflineMode initializes offline mode from env or a startup connectivity probe.
func InitOfflineMode() {
	logger := corelog.For("firebase.offline")
	if raw, ok := os.LookupEnv(env.Offline); ok {
		offlineEnabled.Store(true)
		logger.Warn("offline mode enabled by environment", "env", env.Offline, "value", raw)
		return
	}

	conn, err := net.DialTimeout("tcp", offlineProbeAddress, offlineProbeTimeout)
	if err != nil {
		offlineEnabled.Store(true)
		logger.Warn("offline mode enabled after connectivity check failed", "address", offlineProbeAddress, "timeout", offlineProbeTimeout.String(), "err", err)
		return
	}
	_ = conn.Close()
	offlineEnabled.Store(false)
	logger.Debug("connectivity check passed", "address", offlineProbeAddress)
}

// IsOffline returns true when network requests must be suppressed.
func IsOffline() bool {
	return offlineEnabled.Load()
}

// SetOfflineMode sets offline mode directly, mostly for tests.
func SetOfflineMode(enabled bool) {
	offlineEnabled.Store(enabled)
}
