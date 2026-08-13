package firebase

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewAPIErrorPreservesRemoteFieldsAndRetryHint(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusTooManyRequests, Status: "429 Too Many Requests", Header: http.Header{"Retry-After": []string{"3"}}}
	err := newAPIError("remote-config", "update", response, []byte(`{"error":{"code":429,"message":"quota reached","status":"RESOURCE_EXHAUSTED"}}`))
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T", err)
	}
	if apiErr.Service != "remote-config" || apiErr.Operation != "update" || apiErr.RemoteCode != "429" || apiErr.RemoteStatus != "RESOURCE_EXHAUSTED" || apiErr.Message != "quota reached" || apiErr.RetryAfter != 3*time.Second || !apiErr.Retryable() {
		t.Fatalf("API error = %#v", apiErr)
	}
}

func TestNewAPIErrorIgnoresNegativeRetryHint(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusTooManyRequests, Status: "429 Too Many Requests", Header: http.Header{"Retry-After": []string{"-1"}}}
	err := newAPIError("remote-config", "update", response, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.RetryAfter != 0 {
		t.Fatalf("API error = %#v", apiErr)
	}
}
