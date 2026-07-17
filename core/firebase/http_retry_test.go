package firebase

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestRetryAfterDelayHTTPDate(t *testing.T) {
	when := time.Now().Add(5 * time.Second).UTC().Format(http.TimeFormat)
	resp := &http.Response{
		Header: http.Header{"Retry-After": []string{when}},
	}
	delay, ok := retryAfterDelay(resp)
	if !ok {
		t.Fatal("expected retry-after date to be parsed")
	}
	if delay < 4*time.Second || delay > 6*time.Second {
		t.Fatalf("delay = %v, want ~5s", delay)
	}
}

func TestRetryAfterDelayMissing(t *testing.T) {
	if _, ok := retryAfterDelay(nil); ok {
		t.Fatal("expected false for nil response")
	}
	if _, ok := retryAfterDelay(&http.Response{}); ok {
		t.Fatal("expected false for missing header")
	}
}

func TestRetryDelayUsesRetryAfter(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{"Retry-After": []string{"2"}},
	}
	delay := retryDelay(resp, 1)
	if delay != 2*time.Second {
		t.Fatalf("delay = %v, want 2s", delay)
	}
}

func TestSleepContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sleepContext(ctx, time.Second); err == nil {
		t.Fatal("expected context error")
	}
	if err := sleepContext(context.Background(), 0); err != nil {
		t.Fatalf("zero delay: %v", err)
	}
}

func TestResilientTransportDoesNotRetryCanceledRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	transport := newResilientTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		cancel()
		return nil, ctx.Err()
	}))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transport.RoundTrip(req); !errors.Is(err, context.Canceled) {
		t.Fatalf("RoundTrip error = %v, want context canceled", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestSetOfflineModeToggle(t *testing.T) {
	SetOfflineMode(true)
	if !IsOffline() {
		t.Fatal("expected offline")
	}
	SetOfflineMode(false)
	if IsOffline() {
		t.Fatal("expected online")
	}
}
