package firebase

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestWrapAuthHTTPClientSetsApplicationUserAgent(t *testing.T) {
	tests := []struct {
		name    string
		context context.Context
		want    string
	}{
		{
			name:    "release",
			context: WithApplicationVersion(context.Background(), "0.22.0"),
			want:    "fbrcm/0.22.0 (https://fbrcm.yumaa.dev)",
		},
		{
			name:    "empty version",
			context: WithApplicationVersion(context.Background(), "  "),
			want:    "fbrcm/dev (https://fbrcm.yumaa.dev)",
		},
		{
			name:    "unconfigured version",
			context: context.Background(),
			want:    "fbrcm/dev (https://fbrcm.yumaa.dev)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got string
			base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				got = req.Header.Get("User-Agent")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{}`)),
					Request:    req,
				}, nil
			})
			client := wrapAuthHTTPClient(test.context, &http.Client{Transport: base})
			req, err := http.NewRequest(http.MethodGet, "https://firebase.googleapis.com/v1beta1/projects", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Goog-User-Project", "quota-project")
			req.Header.Set("User-Agent", "Go-http-client/1.1")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			_ = resp.Body.Close()
			if got != test.want {
				t.Fatalf("User-Agent = %q, want %q", got, test.want)
			}
			if original := req.Header.Get("User-Agent"); original != "Go-http-client/1.1" {
				t.Fatalf("original request User-Agent = %q, want unchanged", original)
			}
		})
	}
}
