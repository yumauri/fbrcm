package firebase

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func retryDelay(resp *http.Response, attempt int, policy RetryPolicy) time.Duration {
	if delay, ok := retryAfterDelay(resp); ok {
		return delay
	}
	backoff := exponentialDelay(policy.BaseDelay, policy.MaxDelay, attempt)
	jitterWindow := time.Duration(float64(backoff) * float64(policy.JitterPercent) / 100)
	if jitterWindow <= 0 {
		return backoff
	}
	return backoff + time.Duration(rand.Int63n(int64(jitterWindow)))
}

func exponentialDelay(base, maximum time.Duration, attempt int) time.Duration {
	if attempt <= 1 {
		return base
	}
	delay := base
	for range attempt - 1 {
		if maximum > 0 && delay >= maximum {
			return maximum
		}
		if delay > time.Duration(1<<63-1)/2 {
			return time.Duration(1<<63 - 1)
		}
		delay *= 2
	}
	if maximum > 0 {
		return min(delay, maximum)
	}
	return delay
}

func retryAfterDelay(resp *http.Response) (time.Duration, bool) {
	if resp == nil {
		return 0, false
	}
	raw := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if raw == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		if seconds < 0 {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	when, err := http.ParseTime(raw)
	if err != nil {
		return 0, false
	}
	return max(time.Until(when), 0), true
}

func closeRetryResponse(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
