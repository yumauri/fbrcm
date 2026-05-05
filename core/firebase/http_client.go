package firebase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	corelog "fbrcm/core/log"

	charmlog "github.com/charmbracelet/log"
)

const (
	maxConcurrentRequests = 5
	maxRequestAttempts    = 5
	baseRetryDelay        = 500 * time.Millisecond
	maxRetryDelay         = 10 * time.Second
)

func MaxConcurrentRequests() int {
	return maxConcurrentRequests
}

var requestLimiter = make(chan struct{}, maxConcurrentRequests)

type resilientTransport struct {
	base http.RoundTripper
}

func newResilientTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &resilientTransport{base: base}
}

func (t *resilientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := corelog.For("firebase.http")
	if shouldDryRun(req) {
		logDryRun(logger, req)
		return dryRunResponse(req)
	}

	attempts := maxRequestAttempts
	if !requestCanRetry(req) {
		attempts = 1
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		attemptReq, err := cloneRequest(req, attempt)
		if err != nil {
			return nil, err
		}

		if err := acquireRequestSlot(req.Context()); err != nil {
			return nil, err
		}
		resp, err := t.base.RoundTrip(attemptReq)
		releaseRequestSlot()

		if !shouldRetry(resp, err) || attempt == attempts {
			return resp, err
		}

		delay := retryDelay(resp, attempt)
		logRetry(logger, req, resp, err, attempt, delay)
		closeRetryResponse(resp)
		if err := sleepContext(req.Context(), delay); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("request retries exhausted")
}

func shouldDryRun(req *http.Request) bool {
	if req == nil {
		return false
	}
	if !IsDryRun(req.Context()) {
		return false
	}

	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func dryRunResponse(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("dry run request is nil")
	}

	body := []byte("{}")
	if req.GetBody != nil {
		reader, err := req.GetBody()
		if err != nil {
			return nil, fmt.Errorf("clone request body for dry run: %w", err)
		}
		defer func() { _ = reader.Close() }()

		body, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("read request body for dry run: %w", err)
		}
		if len(bytes.TrimSpace(body)) == 0 {
			body = []byte("{}")
		}
	}

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json; charset=utf-8")
	if etag := strings.TrimSpace(req.Header.Get("If-Match")); etag != "" {
		headers.Set("ETag", etag)
	}

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    http.StatusOK,
		Header:        headers,
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
		Request:       req,
	}, nil
}

func acquireRequestSlot(ctx context.Context) error {
	select {
	case requestLimiter <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func releaseRequestSlot() {
	select {
	case <-requestLimiter:
	default:
	}
}

func requestCanRetry(req *http.Request) bool {
	return req == nil || req.Body == nil || req.GetBody != nil
}

func cloneRequest(req *http.Request, attempt int) (*http.Request, error) {
	cloned := req.Clone(req.Context())
	if req.Body == nil {
		return cloned, nil
	}
	if req.GetBody == nil {
		if attempt == 1 {
			return cloned, nil
		}
		return nil, fmt.Errorf("request body is not replayable for %s %s", req.Method, req.URL.String())
	}

	body, err := req.GetBody()
	if err != nil {
		return nil, fmt.Errorf("clone request body: %w", err)
	}
	cloned.Body = body
	return cloned, nil
}

func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp == nil {
		return false
	}

	switch resp.StatusCode {
	case http.StatusTooManyRequests, http.StatusRequestTimeout, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return resp.StatusCode >= 500
	}
}

func retryDelay(resp *http.Response, attempt int) time.Duration {
	if delay, ok := retryAfterDelay(resp); ok {
		return delay
	}

	backoff := min(baseRetryDelay*time.Duration(1<<(attempt-1)), maxRetryDelay)

	jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
	return backoff + jitter
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
	delay := max(time.Until(when), 0)
	return delay, true
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

func logRetry(logger *charmlog.Logger, req *http.Request, resp *http.Response, err error, attempt int, delay time.Duration) {
	if req == nil || req.URL == nil {
		return
	}

	status := ""
	if resp != nil {
		status = resp.Status
	}

	logger.Warn(
		"retry http request",
		"method", req.Method,
		"url", req.URL.String(),
		"attempt", attempt,
		"next_delay", delay.String(),
		"status", status,
		"err", err,
	)
}

func logDryRun(logger *charmlog.Logger, req *http.Request) {
	if req == nil || req.URL == nil {
		return
	}

	logger.Warn(
		"dry run, skip actual request",
		"method", req.Method,
		"url", req.URL.String(),
	)
}
