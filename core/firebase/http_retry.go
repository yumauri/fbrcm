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

func retryDelay(resp *http.Response, attempt int) time.Duration {
	if delay, ok := retryAfterDelay(resp); ok {
		return delay
	}
	backoff := min(baseRetryDelay*time.Duration(1<<(attempt-1)), maxRetryDelay)
	return backoff + time.Duration(rand.Int63n(int64(backoff/2)))
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
