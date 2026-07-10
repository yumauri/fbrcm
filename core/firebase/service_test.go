package firebase

import (
	"net/http"
	"testing"
)

func TestNewServiceWithHTTPClient(t *testing.T) {
	client := &http.Client{}
	svc := NewServiceWithHTTPClient(client)
	if svc == nil || svc.httpClient != client {
		t.Fatal("NewServiceWithHTTPClient did not preserve client")
	}

	defaultSvc := NewServiceWithHTTPClient(nil)
	if defaultSvc == nil || defaultSvc.httpClient != http.DefaultClient {
		t.Fatal("NewServiceWithHTTPClient(nil) should use http.DefaultClient")
	}
}
