package firebase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	corelog "github.com/yumauri/fbrcm/core/log"

	charmlog "charm.land/log/v2"
)

func MaxConcurrentRequests(ctx context.Context) int {
	return requestControllerFromContext(ctx).Policy().MaxConcurrentRequests
}

type resilientTransport struct {
	base       http.RoundTripper
	controller *RequestController
}

func newResilientTransport(base http.RoundTripper) http.RoundTripper {
	return newResilientTransportWithController(base, defaultRequestController)
}

func newResilientTransportWithController(base http.RoundTripper, controller *RequestController) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if controller == nil {
		controller = defaultRequestController
	}
	return &resilientTransport{base: base, controller: controller}
}

func (t *resilientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := corelog.For("firebase.http")
	if IsOffline() {
		logOffline(logger, req)
		return nil, ErrOffline
	}
	if shouldDryRun(req) {
		logDryRun(logger, req)
		return dryRunResponse(req)
	}
	if requiresQuotaProjectHeader(req) && strings.TrimSpace(req.Header.Get("X-Goog-User-Project")) == "" {
		return nil, &QuotaProjectInvariantError{Method: req.Method, Host: req.URL.Hostname()}
	}

	policy := t.controller.Policy()
	attempts := policy.Retry.MaxAttempts
	if !requestCanRetry(req) {
		attempts = 1
	}
	scheduleKey := requestScheduleKey(req)

	for attempt := 1; attempt <= attempts; attempt++ {
		if err := t.controller.wait(req.Context(), scheduleKey); err != nil {
			return nil, err
		}
		attemptReq, err := cloneRequest(req, attempt)
		if err != nil {
			return nil, err
		}

		if err := t.controller.acquire(req.Context()); err != nil {
			return nil, err
		}
		resp, err := t.base.RoundTrip(attemptReq)
		t.controller.release()
		if contextErr := req.Context().Err(); contextErr != nil {
			closeRetryResponse(resp)
			return nil, contextErr
		}

		retryable := shouldRetry(resp, err)
		rateLimited := resp != nil && resp.StatusCode == http.StatusTooManyRequests
		delay := time.Duration(0)
		if rateLimited {
			delay = t.controller.recordRateLimit(scheduleKey, resp)
		} else {
			t.controller.recordNonRateLimit(scheduleKey)
		}
		if !retryable || attempt == attempts {
			return resp, err
		}

		if !rateLimited {
			delay = retryDelay(resp, attempt, policy.Retry)
		}
		logRetry(logger, req, resp, err, attempt, delay)
		closeRetryResponse(resp)
		if rateLimited {
			continue
		}
		if err := sleepContext(req.Context(), delay); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("request retries exhausted")
}

func requiresQuotaProjectHeader(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	switch strings.ToLower(req.URL.Hostname()) {
	case "cloudresourcemanager.googleapis.com", "firebaseremoteconfig.googleapis.com":
		return true
	default:
		return false
	}
}

func shouldDryRun(req *http.Request) bool {
	if req == nil {
		return false
	}
	if !IsDryRun(req.Context()) {
		return false
	}
	if isRemoteConfigValidationRequest(req) {
		return false
	}

	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func isRemoteConfigValidationRequest(req *http.Request) bool {
	if req == nil || req.URL == nil || req.Method != http.MethodPut {
		return false
	}
	return strings.EqualFold(req.URL.Hostname(), "firebaseremoteconfig.googleapis.com") &&
		strings.HasSuffix(req.URL.Path, "/remoteConfig") &&
		strings.EqualFold(strings.TrimSpace(req.URL.Query().Get("validateOnly")), "true")
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

func requestCanRetry(req *http.Request) bool {
	return req == nil || req.Body == nil || req.GetBody != nil
}

// cloneRequest clones an HTTP request for retry and avoids leaking sensitive query values in errors.
func cloneRequest(req *http.Request, attempt int) (*http.Request, error) {
	cloned := req.Clone(req.Context())
	if req.Body == nil {
		return cloned, nil
	}
	if req.GetBody == nil {
		if attempt == 1 {
			return cloned, nil
		}
		return nil, fmt.Errorf("request body is not replayable for %s %s", req.Method, redactedURLString(req.URL))
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

// logRetry logs retry metadata without exposing sensitive query values.
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
		"url", redactedURLString(req.URL),
		"attempt", attempt,
		"next_delay", delay.String(),
		"status", status,
		"err", err,
	)
}

// logDryRun logs skipped write requests without exposing sensitive query values.
func logDryRun(logger *charmlog.Logger, req *http.Request) {
	if req == nil || req.URL == nil {
		return
	}

	logger.Warn(
		"dry run, skip actual request",
		"method", req.Method,
		"url", redactedURLString(req.URL),
	)
}

// logOffline logs suppressed offline requests without exposing sensitive query values.
func logOffline(logger *charmlog.Logger, req *http.Request) {
	if req == nil || req.URL == nil {
		logger.Warn("offline mode, suppress http request")
		return
	}

	logger.Warn(
		"offline mode, suppress http request",
		"method", req.Method,
		"url", redactedURLString(req.URL),
	)
}
