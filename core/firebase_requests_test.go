package core

import (
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
)

func TestConfigureFirebaseRequestsUsesEffectiveNetworkConfig(t *testing.T) {
	svc := setupCoreTestEnv(t)
	cfg, err := config.LoadGlobalAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	requestsPerMinute := 30
	maxConcurrentRequests := 3
	maxAttempts := 4
	jitterPercent := 25
	cfg.Network = &config.NetworkConfig{
		MaxConcurrentRequests: &maxConcurrentRequests,
		RequestsPerMinute:     &requestsPerMinute,
		RateLimitCooldown:     "90s",
		Retry: &config.RetryConfig{
			MaxAttempts: &maxAttempts, BaseDelay: "250ms", MaxDelay: "4s", JitterPercent: &jitterPercent,
		},
	}
	if err := config.SaveAppConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfigureFirebaseRequests(); err != nil {
		t.Fatal(err)
	}
	policy := svc.firebaseRequests.Policy()
	if policy.MaxConcurrentRequests != 3 || policy.RequestsPerMinute != 30 || policy.RateLimitCooldown != 90*time.Second ||
		policy.Retry.MaxAttempts != 4 || policy.Retry.BaseDelay != 250*time.Millisecond ||
		policy.Retry.MaxDelay != 4*time.Second || policy.Retry.JitterPercent != 25 {
		t.Fatalf("request policy = %+v", policy)
	}

	svc.ResetFirebaseRequestPolicy()
	policy = svc.firebaseRequests.Policy()
	if policy.MaxConcurrentRequests != 5 || policy.RequestsPerMinute != 0 || policy.RateLimitCooldown != 30*time.Second ||
		policy.Retry.MaxAttempts != 5 || policy.Retry.BaseDelay != time.Second ||
		policy.Retry.MaxDelay != 10*time.Second || policy.Retry.JitterPercent != 50 {
		t.Fatalf("default request policy = %+v", policy)
	}
}
