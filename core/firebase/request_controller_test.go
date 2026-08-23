package firebase

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRequestControllerPacesSharedRequests(t *testing.T) {
	controller := NewRequestController(RequestPolicy{RequestsPerMinute: 1200, RateLimitCooldown: time.Second})
	transport := newResilientTransportWithController(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return requestControllerResponse(req, http.StatusOK), nil
	}), controller)
	req, err := http.NewRequest(http.MethodGet, "https://firebaseremoteconfig.googleapis.com/v1/projects/demo/remoteConfig", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed < 40*time.Millisecond {
		t.Fatalf("second request waited %v, want pacing near 50ms", elapsed)
	}
}

func TestRequestControllerCooldownIsSharedAcrossTransports(t *testing.T) {
	controller := NewRequestController(RequestPolicy{RateLimitCooldown: 50 * time.Millisecond})
	req, err := http.NewRequest(http.MethodGet, "https://firebaseremoteconfig.googleapis.com/v1/projects/demo/remoteConfig", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller.recordRateLimit(requestScheduleKey(req), requestControllerResponse(req, http.StatusTooManyRequests))
	started := time.Now()

	var wg sync.WaitGroup
	elapsed := make(chan time.Duration, 2)
	for range 2 {
		transport := newResilientTransportWithController(roundTripFunc(func(req *http.Request) (*http.Response, error) {
			elapsed <- time.Since(started)
			return requestControllerResponse(req, http.StatusOK), nil
		}), controller)
		wg.Go(func() {
			if _, err := transport.RoundTrip(req); err != nil {
				t.Errorf("RoundTrip = %v", err)
			}
		})
	}
	wg.Wait()
	close(elapsed)
	for delay := range elapsed {
		if delay < 40*time.Millisecond {
			t.Fatalf("shared request started after %v, want cooldown near 50ms", delay)
		}
	}
}

func TestResilientTransportUses429FallbackCooldown(t *testing.T) {
	controller := NewRequestController(RequestPolicy{RateLimitCooldown: 40 * time.Millisecond})
	attempts := 0
	transport := newResilientTransportWithController(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return requestControllerResponse(req, http.StatusTooManyRequests), nil
		}
		return requestControllerResponse(req, http.StatusOK), nil
	}), controller)
	req, err := http.NewRequest(http.MethodGet, "https://firebaseremoteconfig.googleapis.com/v1/projects/demo/remoteConfig", nil)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if elapsed := time.Since(started); elapsed < 30*time.Millisecond {
		t.Fatalf("429 retry waited %v, want fallback cooldown near 40ms", elapsed)
	}
}

func TestRequestControllerCooldownHonorsContextCancellation(t *testing.T) {
	controller := NewRequestController(RequestPolicy{RateLimitCooldown: time.Second})
	req, err := http.NewRequest(http.MethodGet, "https://firebaseremoteconfig.googleapis.com/v1/projects/demo/remoteConfig", nil)
	if err != nil {
		t.Fatal(err)
	}
	controller.recordRateLimit(requestScheduleKey(req), requestControllerResponse(req, http.StatusTooManyRequests))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := controller.wait(ctx, requestScheduleKey(req)); err == nil {
		t.Fatal("canceled cooldown wait succeeded")
	}
}

func TestRequestControllerRateLimitDelayPrefersRetryAfter(t *testing.T) {
	controller := NewRequestController(RequestPolicy{RateLimitCooldown: time.Minute})
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"2"}}}
	if got := controller.recordRateLimit("with-header", resp); got != 2*time.Second {
		t.Fatalf("rate limit delay = %v, want 2s", got)
	}
	if got := controller.recordRateLimit("without-header", &http.Response{}); got != time.Minute {
		t.Fatalf("fallback rate limit delay = %v, want 1m", got)
	}
}

func TestRequestControllerLinearlyIncreasesConsecutive429CooldownsAndResets(t *testing.T) {
	controller := NewRequestController(RequestPolicy{RateLimitCooldown: 30 * time.Millisecond})
	if got := controller.recordRateLimit("quota", &http.Response{}); got != 30*time.Millisecond {
		t.Fatalf("first cooldown = %v", got)
	}
	if got := controller.recordRateLimit("quota", &http.Response{}); got != 60*time.Millisecond {
		t.Fatalf("second cooldown = %v", got)
	}
	if got := controller.recordRateLimit("quota", &http.Response{}); got != 90*time.Millisecond {
		t.Fatalf("third cooldown = %v", got)
	}
	controller.recordNonRateLimit("quota")
	if got := controller.recordRateLimit("quota", &http.Response{}); got != 30*time.Millisecond {
		t.Fatalf("cooldown after reset = %v", got)
	}
}

func TestRequestControllerLimitsConcurrencyAcrossTransports(t *testing.T) {
	controller := NewRequestController(RequestPolicy{MaxConcurrentRequests: 2})
	release := make(chan struct{})
	started := make(chan struct{}, 4)
	transport := newResilientTransportWithController(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		started <- struct{}{}
		<-release
		return requestControllerResponse(req, http.StatusOK), nil
	}), controller)

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			req, err := http.NewRequest(http.MethodGet, "https://example.test", nil)
			if err != nil {
				t.Error(err)
				return
			}
			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Error(err)
				return
			}
			_ = resp.Body.Close()
		})
	}
	for range 2 {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("request did not start")
		}
	}
	select {
	case <-started:
		t.Fatal("more than two requests started concurrently")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	wg.Wait()
}

func TestAccessTokenServicesUseContextRequestController(t *testing.T) {
	controller := NewRequestController(RequestPolicy{RequestsPerMinute: 30, RateLimitCooldown: time.Minute})
	ctx := WithRequestController(context.Background(), controller)
	for range 2 {
		service, err := NewServiceWithAccessToken(ctx, "test-access-token")
		if err != nil {
			t.Fatal(err)
		}
		transport, ok := service.httpClient.Transport.(*resilientTransport)
		if !ok || transport.controller != controller {
			t.Fatalf("transport = %#v, want shared controller", service.httpClient.Transport)
		}
	}
}

func requestControllerResponse(req *http.Request, status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Request:    req,
	}
}
