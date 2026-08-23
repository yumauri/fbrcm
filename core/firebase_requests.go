package core

import (
	"context"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

// ConfigureFirebaseRequests loads effective application network settings and
// installs one controller shared by every subsequently created Firebase client.
func (s *Core) ConfigureFirebaseRequests() error {
	resolved, err := config.ResolveAppConfig()
	if err != nil {
		return err
	}
	config.SetSessionAppConfigResolution(resolved)
	return s.ConfigureFirebaseRequestsFromConfig(resolved.Effective)
}

// ConfigureFirebaseRequestsFromConfig installs request settings from an
// already resolved effective application configuration.
func (s *Core) ConfigureFirebaseRequestsFromConfig(cfg *config.AppConfig) error {
	if cfg == nil {
		cfg = &config.AppConfig{}
	}
	delay, err := cfg.Network.EffectiveRateLimitCooldown()
	if err != nil {
		return err
	}
	var retry *config.RetryConfig
	if cfg.Network != nil {
		retry = cfg.Network.Retry
	}
	baseDelay, err := retry.EffectiveBaseDelay()
	if err != nil {
		return err
	}
	maxDelay, err := retry.EffectiveMaxDelay()
	if err != nil {
		return err
	}
	s.SetFirebaseRequestPolicy(firebase.RequestPolicy{
		MaxConcurrentRequests: cfg.Network.EffectiveMaxConcurrentRequests(),
		RequestsPerMinute:     cfg.Network.EffectiveRequestsPerMinute(),
		RateLimitCooldown:     delay,
		Retry: firebase.RetryPolicy{
			MaxAttempts:   retry.EffectiveMaxAttempts(),
			BaseDelay:     baseDelay,
			MaxDelay:      maxDelay,
			JitterPercent: retry.EffectiveJitterPercent(),
		},
	})
	return nil
}

// SetFirebaseRequestPolicy installs a new controller. Call before API clients
// are created so all transports share the same policy and schedule.
func (s *Core) SetFirebaseRequestPolicy(policy firebase.RequestPolicy) {
	if s == nil {
		return
	}
	s.firebaseRequestsMu.Lock()
	s.firebaseRequests = firebase.NewRequestController(policy)
	s.firebaseRequestsMu.Unlock()
}

// ResetFirebaseRequestPolicy restores built-in defaults without reading local
// configuration, as required by stateless execution.
func (s *Core) ResetFirebaseRequestPolicy() {
	s.SetFirebaseRequestPolicy(firebase.DefaultRequestPolicy())
}

// WithFirebaseRequestController binds the Core controller to a request context.
func (s *Core) WithFirebaseRequestController(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if s == nil {
		return ctx
	}
	s.firebaseRequestsMu.RLock()
	controller := s.firebaseRequests
	s.firebaseRequestsMu.RUnlock()
	return firebase.WithRequestController(ctx, controller)
}
