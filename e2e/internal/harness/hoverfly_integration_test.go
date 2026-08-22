package harness

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHoverflyCapturesGuardsAndReplaysOffline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	e2eRoot := filepath.Clean(filepath.Join(workingDirectory, "..", ".."))
	toolsDirectory := t.TempDir()
	hoverflyBinary, err := ResolveHoverfly(ctx, e2eRoot, toolsDirectory)
	if err != nil {
		t.Fatal(err)
	}
	guardPath, err := BuildReadGuard(ctx, e2eRoot, toolsDirectory)
	if err != nil {
		t.Fatal(err)
	}
	certificatePath, keyPath, err := GenerateCA(ctx, hoverflyBinary, filepath.Join(toolsDirectory, "ca"))
	if err != nil {
		t.Fatal(err)
	}

	wantBody := []byte{0, 1, 'f', 'b', 'r', 'c', 'm', '\n'}
	var getCount atomic.Int32
	var mutationCount atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			mutationCount.Add(1)
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		getCount.Add(1)
		writer.Header().Set("Content-Type", "application/octet-stream")
		_, _ = writer.Write(wantBody)
	}))
	t.Cleanup(upstream.Close)
	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	cassettePath := filepath.Join(t.TempDir(), "http.json")
	capture, err := StartHoverfly(ctx, HoverflyOptions{
		Binary:          hoverflyBinary,
		Directory:       filepath.Join(t.TempDir(), "capture"),
		CertificatePath: certificatePath,
		KeyPath:         keyPath,
		GuardPath:       guardPath,
		AllowedRequests: []HTTPExpectation{{Method: http.MethodGet, Host: upstreamURL.Host, Path: "/fixture", Status: http.StatusOK}},
		CassettePath:    cassettePath,
		Capture:         true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = capture.Stop() })
	client := proxyClient(t, capture.ProxyURL())
	requestURL := upstream.URL + "/fixture?format=raw"
	gotBody := getBody(t, client, requestURL)
	if !bytes.Equal(gotBody, wantBody) {
		t.Fatalf("captured body = %q, want %q", gotBody, wantBody)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, strings.NewReader("mutation"))
	if err != nil {
		t.Fatal(err)
	}
	response, requestErr := client.Do(request)
	if response != nil {
		_ = response.Body.Close()
	}
	if requestErr == nil {
		if response == nil {
			t.Fatal("mutation returned neither a response nor an error")
		}
		if response.StatusCode < http.StatusBadRequest {
			t.Fatalf("mutation returned %s", response.Status)
		}
	}
	if mutationCount.Load() != 0 {
		t.Fatalf("upstream received %d mutations", mutationCount.Load())
	}
	simulation, err := capture.ExportSimulation(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := capture.Stop(); err != nil {
		t.Fatal(err)
	}
	upstream.Close()
	if getCount.Load() != 1 {
		t.Fatalf("upstream received %d GET requests", getCount.Load())
	}
	if err := atomicWrite(cassettePath, simulation, 0o644); err != nil {
		t.Fatal(err)
	}

	replay, err := StartHoverfly(ctx, HoverflyOptions{
		Binary:          hoverflyBinary,
		Directory:       filepath.Join(t.TempDir(), "replay"),
		CertificatePath: certificatePath,
		KeyPath:         keyPath,
		GuardPath:       guardPath,
		AllowedRequests: []HTTPExpectation{{Method: http.MethodGet, Host: upstreamURL.Host, Path: "/fixture", Status: http.StatusOK}},
		CassettePath:    cassettePath,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = replay.Stop() }()
	replayedBody := getBody(t, proxyClient(t, replay.ProxyURL()), requestURL)
	if !bytes.Equal(replayedBody, wantBody) {
		t.Fatalf("replayed body = %q, want %q", replayedBody, wantBody)
	}
	journal, err := replay.Journal(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateJournal(journal, []HTTPExpectation{{Method: http.MethodGet, Host: upstreamURL.Host, Path: "/fixture", Status: http.StatusOK}}, false, false); err != nil {
		t.Fatal(err)
	}
}

func proxyClient(t *testing.T, rawProxyURL string) *http.Client {
	t.Helper()
	proxyURL, err := url.Parse(rawProxyURL)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		Timeout:   5 * time.Second,
	}
}

func getBody(t *testing.T, client *http.Client, requestURL string) []byte {
	t.Helper()
	response, err := client.Get(requestURL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(response.Body)
		t.Fatalf("GET returned %s: %s", response.Status, raw)
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
