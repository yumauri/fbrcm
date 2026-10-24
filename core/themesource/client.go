package themesource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core/env"
)

// Client is scoped to one import. Its budgets include revision discovery.
type Client struct {
	http          http.Client
	token         string
	source        *Source
	requests      int
	metadataBytes int
	downloadBytes int
	cache         map[string][]entry
}

func NewClient(client *http.Client, source *Source) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	scoped := *client
	if scoped.Timeout == 0 || scoped.Timeout > 20*time.Second {
		scoped.Timeout = 20 * time.Second
	}
	scoped.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	token := ""
	if source.Provider == "github" {
		token, _ = env.LookupNonEmpty(env.GitHubToken)
		if token == "" {
			token, _ = env.LookupNonEmpty(env.GitHubTokenFallback)
		}
	} else {
		token, _ = env.LookupNonEmpty(env.GitLabToken)
	}
	return &Client{http: scoped, source: source, token: token, cache: make(map[string][]entry)}
}

func (c *Client) base() string {
	if c.source.Provider == "github" {
		return "https://api.github.com/repos/" + escapedPath(c.source.Repository)
	}
	return "https://gitlab.com/api/v4/projects/" + url.PathEscape(c.source.Repository)
}

func (c *Client) get(ctx context.Context, endpoint string, raw bool) ([]byte, http.Header, error) {
	if c.requests >= maximumRequests {
		return nil, nil, failure("limit", "repository import exceeds the 1024-request limit")
	}
	c.requests++
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, failure("invalid", "invalid repository API request")
	}
	if c.source.Provider == "github" {
		req.Header.Set("Accept", "application/vnd.github+json")
		if raw {
			req.Header.Set("Accept", "application/vnd.github.raw+json")
		}
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
	} else if c.token != "" {
		req.Header.Set("PRIVATE-TOKEN", c.token)
	}
	req.Header.Set("User-Agent", "fbrcm-theme-import")
	resp, err := c.http.Do(req)
	if err != nil {
		kind := "unavailable"
		var network net.Error
		if errors.Is(err, context.Canceled) {
			return nil, nil, context.Canceled
		}
		if errors.As(err, &network) && network.Timeout() {
			kind = "timeout"
		}
		return nil, nil, &Error{Kind: kind, Provider: c.source.Provider, Operation: "download", Err: fmt.Errorf("%s repository request failed (%s)", c.source.Provider, kind)}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		kind, message := "request_failed", "repository API request failed"
		switch {
		case resp.StatusCode == 401:
			kind, message = "auth", "repository authentication failed; check the provider token"
		case resp.StatusCode == 429 || resp.StatusCode == 403 && (resp.Header.Get("X-RateLimit-Remaining") == "0" || resp.Header.Get("Retry-After") != ""):
			kind, message = "rate_limited", "repository API rate limit reached; retry later or provide a token"
		case resp.StatusCode == 403:
			kind, message = "permission", "repository access denied; check the token permissions"
		case resp.StatusCode == 404:
			kind, message = "not_found", "repository, revision, or path was not found or is inaccessible; check the URL and token access"
		case resp.StatusCode >= 300 && resp.StatusCode < 400:
			kind, message = "request_failed", "repository API redirected; use the current repository browser URL"
		case resp.StatusCode == 408 || resp.StatusCode == 504:
			kind, message = "timeout", "repository API request timed out"
		case resp.StatusCode >= 500:
			kind, message = "unavailable", "repository API is unavailable"
		}
		return nil, nil, &Error{Kind: kind, Provider: c.source.Provider, Operation: "download", HTTPStatus: resp.StatusCode, RetryAfterMS: retryAfter(resp.Header), Err: fmt.Errorf("%s: %s (HTTP %d)", c.source.Provider, message, resp.StatusCode)}
	}
	limit := maximumMetadataBytes
	if raw {
		limit = MaximumFileBytes
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, int64(limit)+1))
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, nil, context.Canceled
		}
		kind := "unavailable"
		var network net.Error
		if errors.As(err, &network) && network.Timeout() {
			kind = "timeout"
		}
		return nil, nil, failure(kind, "could not read repository API response")
	}
	if len(data) > limit {
		return nil, nil, failure("limit", fmt.Sprintf("repository response exceeds the %d-byte limit", limit))
	}
	if raw {
		c.downloadBytes += len(data)
		if c.downloadBytes > MaximumBatchBytes {
			return nil, nil, failure("limit", "repository import exceeds the 16 MiB download limit")
		}
	} else {
		c.metadataBytes += len(data)
		if c.metadataBytes > maximumTotalMetadataBytes {
			return nil, nil, failure("limit", "repository import exceeds the 64 MiB metadata limit")
		}
	}
	return data, resp.Header, nil
}

func (c *Client) json(ctx context.Context, endpoint string, target any) (http.Header, error) {
	data, headers, err := c.get(ctx, endpoint, false)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return nil, failure("request_failed", "repository API returned invalid JSON")
	}
	return headers, nil
}

func retryAfter(headers http.Header) int64 {
	value := headers.Get("Retry-After")
	if seconds, err := strconv.ParseInt(value, 10, 32); err == nil && seconds > 0 {
		return seconds * 1000
	}
	if when, err := http.ParseTime(value); err == nil {
		return max(0, time.Until(when).Milliseconds())
	}
	if seconds, err := strconv.ParseInt(headers.Get("X-RateLimit-Reset"), 10, 64); err == nil {
		return max(0, time.Until(time.Unix(seconds, 0)).Milliseconds())
	}
	return 0
}

func escapedPath(path string) string {
	parts := strings.Split(path, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}
