package firebase

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestFormatHeadersRedactsSensitiveValues(t *testing.T) {
	headers := http.Header{
		"Authorization": []string{"Bearer secret"},
		"Accept":        []string{"application/json"},
	}
	got := formatHeaders(headers)
	if strings.Contains(got, "secret") {
		t.Fatalf("formatHeaders leaked secret: %q", got)
	}
	if !strings.Contains(got, "Authorization=[REDACTED]") {
		t.Fatalf("formatHeaders = %q, want redacted Authorization", got)
	}
	if !strings.Contains(got, "Accept=application/json") {
		t.Fatalf("formatHeaders = %q, want Accept header", got)
	}
}

func TestRedactedURLString(t *testing.T) {
	raw, err := url.Parse("https://example.com/oauth?access_token=secret&page=1")
	if err != nil {
		t.Fatal(err)
	}
	got := redactedURLString(raw)
	if strings.Contains(got, "secret") {
		t.Fatalf("redacted URL leaked token: %q", got)
	}
	if !strings.Contains(got, "access_token=%5BREDACTED%5D") && !strings.Contains(got, "access_token=[REDACTED]") {
		t.Fatalf("redacted URL = %q, want redacted access_token", got)
	}
}

func TestLogHTTPRequestNilSafe(t *testing.T) {
	logHTTPRequest(nil, nil)
	logHTTPResponse(nil, nil, nil)
	if formatHeaders(nil) != "" {
		t.Fatal("formatHeaders(nil) should be empty")
	}
}
