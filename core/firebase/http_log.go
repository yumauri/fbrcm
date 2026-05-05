package firebase

import (
	"net/http"
	"sort"
	"strings"

	charmlog "github.com/charmbracelet/log"
)

func logHTTPRequest(logger *charmlog.Logger, req *http.Request) {
	if req == nil || req.URL == nil {
		return
	}

	logger.Debug("http request", "method", req.Method, "url", req.URL.String(), "headers", formatHeaders(req.Header))
}

func logHTTPResponse(logger *charmlog.Logger, req *http.Request, resp *http.Response) {
	if req == nil || req.URL == nil || resp == nil {
		return
	}

	logger.Info("http response", "method", req.Method, "url", req.URL.String(), "status", resp.Status)
	logger.Debug("http response headers", "method", req.Method, "url", req.URL.String(), "status", resp.Status, "headers", formatHeaders(resp.Header))
}

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
		parts = append(parts, key+"="+strings.Join(headers.Values(key), ","))
	}

	return strings.Join(parts, " ")
}
