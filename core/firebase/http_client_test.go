package firebase

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNormalizeJSONEscapes(t *testing.T) {
	in := []byte(`{"a":"\u003cscript\u003e","b":"\u0026amp;"}`)
	got := NormalizeJSONEscapes(in)
	want := `{"a":"<script>","b":"&amp;"}`
	if string(got) != want {
		t.Fatalf("NormalizeJSONEscapes = %q, want %q", got, want)
	}
}

func TestNormalizeJSONEscapesUppercaseUnicode(t *testing.T) {
	in := []byte(`{"x":"\u003Ctag\u003E"}`)
	got := NormalizeJSONEscapes(in)
	if !strings.Contains(string(got), "<tag>") {
		t.Fatalf("expected uppercase unicode escapes normalized, got %q", got)
	}
}

func TestIsDryRun(t *testing.T) {
	if IsDryRun(context.Background()) {
		t.Fatal("background context should not be dry run")
	}
	ctx := WithDryRun(context.Background())
	if !IsDryRun(ctx) {
		t.Fatal("expected dry run context")
	}
}

func TestShouldDryRun(t *testing.T) {
	ctx := WithDryRun(context.Background())
	getReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
	if shouldDryRun(getReq) {
		t.Fatal("GET should not dry-run")
	}
	putReq, _ := http.NewRequestWithContext(ctx, http.MethodPut, "https://example.com", strings.NewReader(`{}`))
	if !shouldDryRun(putReq) {
		t.Fatal("PUT in dry-run context should dry-run")
	}
	plainPut, _ := http.NewRequest(http.MethodPut, "https://example.com", strings.NewReader(`{}`))
	if shouldDryRun(plainPut) {
		t.Fatal("PUT without dry-run context should not dry-run")
	}
}

func TestDryRunResponseEchoesBodyAndETag(t *testing.T) {
	body := strings.NewReader(`{"version":{"versionNumber":"3"}}`)
	req, err := http.NewRequest(http.MethodPut, "https://example.com/v1/config", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("If-Match", `"etag-1"`)
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(`{"version":{"versionNumber":"3"}}`)), nil
	}

	resp, err := dryRunResponse(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if resp.Header.Get("ETag") != `"etag-1"` {
		t.Fatalf("ETag = %q", resp.Header.Get("ETag"))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"version":{"versionNumber":"3"}}` {
		t.Fatalf("body = %q", data)
	}
}

func TestShouldRetry(t *testing.T) {
	if !shouldRetry(nil, io.EOF) {
		t.Fatal("errors should retry")
	}
	resp503 := &http.Response{StatusCode: http.StatusServiceUnavailable}
	if !shouldRetry(resp503, nil) {
		t.Fatal("503 should retry")
	}
	resp404 := &http.Response{StatusCode: http.StatusNotFound}
	if shouldRetry(resp404, nil) {
		t.Fatal("404 should not retry")
	}
}

func TestRetryAfterDelay(t *testing.T) {
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("Retry-After", "2")
	delay, ok := retryAfterDelay(resp)
	if !ok || delay != 2*time.Second {
		t.Fatalf("Retry-After seconds = %v, ok=%v", delay, ok)
	}
}

func TestResilientTransportRetriesOn503(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	transport := newResilientTransport(http.DefaultTransport)
	body := []byte(`{"x":1}`)
	req, err := http.NewRequest(http.MethodPut, server.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestResilientTransportOffline(t *testing.T) {
	SetOfflineMode(true)
	t.Cleanup(func() { SetOfflineMode(false) })

	transport := newResilientTransport(http.DefaultTransport)
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = transport.RoundTrip(req)
	if err != ErrOffline {
		t.Fatalf("err = %v, want ErrOffline", err)
	}
}

func TestResilientTransportDryRunSkipsNetwork(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := newResilientTransport(http.DefaultTransport)
	ctx := WithDryRun(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, server.URL, strings.NewReader(`{"dry":true}`))
	if err != nil {
		t.Fatal(err)
	}
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(`{"dry":true}`)), nil
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if called {
		t.Fatal("dry run should not hit server")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}
