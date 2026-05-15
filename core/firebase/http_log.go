package firebase

import (
	"net/http"
	"net/url"
	"sort"
	"strings"

	charmlog "github.com/charmbracelet/log"
)

// logHTTPRequest logs an HTTP request without exposing sensitive headers or query values.
func logHTTPRequest(logger *charmlog.Logger, req *http.Request) {
	if req == nil || req.URL == nil {
		return
	}

	logger.Info("http request", "method", req.Method, "url", redactedURLString(req.URL), "headers", formatHeaders(req.Header))
}

// logHTTPResponse logs an HTTP response without exposing sensitive headers or query values.
func logHTTPResponse(logger *charmlog.Logger, req *http.Request, resp *http.Response) {
	if req == nil || req.URL == nil || resp == nil {
		return
	}

	safeURL := redactedURLString(req.URL)
	logger.Info("http response", "method", req.Method, "url", safeURL, "status", resp.Status)
	logger.Debug("http response headers", "method", req.Method, "url", safeURL, "status", resp.Status, "headers", formatHeaders(resp.Header))
}

// formatHeaders formats headers for logs, replacing sensitive header values with a redaction marker.
func formatHeaders(headers http.Header) string {
	if len(headers) == 0 {
		return ""
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		values := headers.Values(key)
		if isSensitiveHeader(key) {
			values = []string{"[REDACTED]"}
		}
		parts = append(parts, key+"="+strings.Join(values, ","))
	}

	return strings.Join(parts, " ")
}

// redactedURLString returns a URL safe for logs.
func redactedURLString(raw *url.URL) string {
	if raw == nil {
		return ""
	}

	safe := *raw
	query := safe.Query()
	for key := range query {
		if isSensitiveQueryParam(key) {
			query.Set(key, "[REDACTED]")
		}
	}
	safe.RawQuery = query.Encode()
	return safe.String()
}

// redactedURLStringValue parses a raw URL string and returns a log-safe version when parsing succeeds.
func redactedURLStringValue(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return redactedURLString(parsed)
}

// isSensitiveHeader reports whether a header value should be redacted from logs.
func isSensitiveHeader(key string) bool {
	switch strings.ToLower(key) {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-goog-api-key", "x-api-key":
		return true
	default:
		return false
	}
}

// isSensitiveQueryParam reports whether a query parameter value should be redacted from logs.
func isSensitiveQueryParam(key string) bool {
	switch strings.ToLower(key) {
	case "access_token", "authuser", "client_secret", "code", "code_challenge", "code_verifier", "id_token", "password", "refresh_token", "state", "token":
		return true
	default:
		return false
	}
}
