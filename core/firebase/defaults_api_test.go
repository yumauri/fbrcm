package firebase

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDownloadRemoteConfigDefaults(t *testing.T) {
	const body = "\n<defaults><entry key=\"flag\">on</entry></defaults>\n"
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", req.Method)
		}
		if req.URL.Path != "/v1/projects/demo/remoteConfig:downloadDefaults" {
			t.Fatalf("path = %q", req.URL.Path)
		}
		if got := req.URL.Query().Get("format"); got != "XML" {
			t.Fatalf("format = %q, want XML", got)
		}
		return jsonHTTPResponse(http.StatusOK, body, ""), nil
	})})

	got, err := svc.DownloadRemoteConfigDefaults(context.Background(), "demo", DefaultsFormatXML)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("defaults = %q, want exact response %q", got, body)
	}
}

func TestDownloadRemoteConfigDefaultsRejectsInvalidFormatWithoutRequest(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("unexpected request")
		return nil, io.EOF
	})})
	_, err := svc.DownloadRemoteConfigDefaults(context.Background(), "demo", DefaultsFormat("yaml"))
	if err == nil || !strings.Contains(err.Error(), "allowed: json, xml, plist") {
		t.Fatalf("format error = %v", err)
	}
}

func TestDownloadRemoteConfigDefaultsReportsAPIError(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonHTTPResponse(http.StatusForbidden, `{"error":"denied"}`, ""), nil
	})})
	_, err := svc.DownloadRemoteConfigDefaults(context.Background(), "demo", DefaultsFormatJSON)
	if err == nil || !strings.Contains(err.Error(), "Forbidden") || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("api error = %v", err)
	}
}

func TestParseDefaultsFormatIsCaseInsensitive(t *testing.T) {
	for input, want := range map[string]DefaultsFormat{"json": DefaultsFormatJSON, "XML": DefaultsFormatXML, " PlIsT ": DefaultsFormatPlist} {
		got, err := ParseDefaultsFormat(input)
		if err != nil || got != want {
			t.Errorf("ParseDefaultsFormat(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
}
