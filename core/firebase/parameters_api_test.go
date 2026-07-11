package firebase

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGetRemoteConfigAndLatestVersion(t *testing.T) {
	const rcBody = `{"version":{"versionNumber":"12"},"parameters":{"flag":{"defaultValue":{"value":"on"}}}}`
	svc := NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/remoteConfig") && !strings.Contains(req.URL.Path, "listVersions"):
				return jsonHTTPResponse(http.StatusOK, rcBody, `"etag-rc"`), nil
			case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions"):
				return jsonHTTPResponse(http.StatusOK, `{"versions":[{"versionNumber":"12"}]}`, ""), nil
			default:
				return nil, io.EOF
			}
		}),
	})

	raw, etag, err := svc.GetRemoteConfig(context.Background(), "demo")
	if err != nil {
		t.Fatalf("GetRemoteConfig = %v", err)
	}
	if etag != `"etag-rc"` {
		t.Fatalf("etag = %q", etag)
	}
	cfg, err := ParseRemoteConfig(raw)
	if err != nil || cfg.Version.VersionNumber != "12" {
		t.Fatalf("ParseRemoteConfig = %v version=%q", err, cfg.Version.VersionNumber)
	}

	version, err := svc.GetLatestRemoteConfigVersion(context.Background(), "demo")
	if err != nil || version.VersionNumber != "12" {
		t.Fatalf("GetLatestRemoteConfigVersion = %+v err=%v", version, err)
	}
}

func TestListRemoteConfigVersionsAndRollback(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.Method {
		case http.MethodGet:
			if req.URL.Query().Get("pageSize") != "25" || req.URL.Query().Get("pageToken") != "next" || req.URL.Query().Get("endVersionNumber") != "40" {
				t.Fatalf("query = %q", req.URL.RawQuery)
			}
			return jsonHTTPResponse(http.StatusOK, `{"versions":[{"versionNumber":"39","updateUser":{"email":"a@example.com"},"updateOrigin":"REST_API","updateType":"ROLLBACK","rollbackSource":"37"}],"nextPageToken":"older"}`, ""), nil
		case http.MethodPost:
			var payload map[string]string
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload["versionNumber"] != "37" {
				t.Fatalf("rollback payload = %#v", payload)
			}
			return jsonHTTPResponse(http.StatusOK, `{"version":{"versionNumber":"41","rollbackSource":"37"}}`, `"etag-41"`), nil
		default:
			return nil, io.EOF
		}
	})})
	page, err := svc.ListRemoteConfigVersions(context.Background(), "demo", ListVersionsOptions{PageSize: 25, PageToken: "next", EndVersionNumber: "40"})
	if err != nil || page.NextPageToken != "older" || len(page.Versions) != 1 || page.Versions[0].UpdateUser.Email != "a@example.com" {
		t.Fatalf("ListRemoteConfigVersions = %+v err=%v", page, err)
	}
	raw, etag, err := svc.RollbackRemoteConfig(context.Background(), "demo", "37")
	if err != nil || etag != `"etag-41"` {
		t.Fatalf("RollbackRemoteConfig etag=%q err=%v", etag, err)
	}
	cfg, _ := ParseRemoteConfig(raw)
	if cfg.Version.VersionNumber != "41" {
		t.Fatalf("rollback version = %s", cfg.Version.VersionNumber)
	}
}

func TestGetRemoteConfigVersion(t *testing.T) {
	const rcBody = `{"version":{"versionNumber":"7"}}`
	svc := NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if got := req.URL.Query().Get("versionNumber"); got != "7" {
				t.Fatalf("versionNumber query = %q, want 7", got)
			}
			return jsonHTTPResponse(http.StatusOK, rcBody, `"etag-7"`), nil
		}),
	})

	raw, etag, err := svc.GetRemoteConfig(context.Background(), "demo", "7")
	if err != nil {
		t.Fatalf("GetRemoteConfig version = %v", err)
	}
	if etag != `"etag-7"` {
		t.Fatalf("etag = %q, want etag-7", etag)
	}
	cfg, err := ParseRemoteConfig(raw)
	if err != nil || cfg.Version.VersionNumber != "7" {
		t.Fatalf("ParseRemoteConfig = %v version=%q", err, cfg.Version.VersionNumber)
	}
}

func TestUpdateRemoteConfigValidateOnly(t *testing.T) {
	payload := []byte(`{"version":{"versionNumber":"3"},"parameters":{"flag":{"defaultValue":{"value":"x"}}}}`)
	svc := NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPut {
				t.Fatalf("method = %s, want PUT", req.Method)
			}
			if !strings.Contains(req.URL.RawQuery, "validateOnly=true") {
				t.Fatalf("query = %q, want validateOnly", req.URL.RawQuery)
			}
			if got := req.Header.Get("If-Match"); got != "etag-1" {
				t.Fatalf("If-Match = %q, want etag-1", got)
			}
			return jsonHTTPResponse(http.StatusOK, string(payload), `"etag-1"`), nil
		}),
	})

	if err := svc.ValidateRemoteConfig(context.Background(), "demo", payload, "etag-1"); err != nil {
		t.Fatalf("ValidateRemoteConfig = %v", err)
	}
}

func TestGetRemoteConfigNon200(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return jsonHTTPResponse(http.StatusForbidden, `{"error":"denied"}`, ""), nil
		}),
	})

	_, _, err := svc.GetRemoteConfig(context.Background(), "demo")
	if err == nil || !strings.Contains(err.Error(), "Forbidden") {
		t.Fatalf("GetRemoteConfig = %v, want Forbidden error", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func jsonHTTPResponse(status int, body, etag string) *http.Response {
	resp := &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	if etag != "" {
		resp.Header.Set("ETag", etag)
	}
	return resp
}
