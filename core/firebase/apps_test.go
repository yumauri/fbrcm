package firebase

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestListAppsEncodesPaginationAndDeletedOption(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet || req.URL.Path != "/v1beta1/projects/demo-project:searchApps" {
			t.Fatalf("request = %s %s", req.Method, req.URL.Path)
		}
		if got := req.URL.Query().Get("pageSize"); got != "100" {
			t.Fatalf("pageSize = %q", got)
		}
		if got := req.URL.Query().Get("pageToken"); got != "next token" {
			t.Fatalf("pageToken = %q", got)
		}
		if got := req.URL.Query().Get("showDeleted"); got != "true" {
			t.Fatalf("showDeleted = %q", got)
		}
		if got := req.Header.Get("X-Goog-User-Project"); got != "demo-project" {
			t.Fatalf("quota project = %q", got)
		}
		return jsonHTTPResponse(http.StatusOK, `{"apps":[{"name":"projects/demo-project/androidApps/abc","displayName":"Android","platform":"ANDROID","appId":"1:2:android:3","namespace":"com.example"}],"nextPageToken":"older"}`, ""), nil
	})})

	page, err := svc.ListApps(context.Background(), "demo-project", ListAppsOptions{PageSize: 100, PageToken: "next token", ShowDeleted: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Apps) != 1 || page.Apps[0].Platform != AppPlatformAndroid || page.NextPageToken != "older" {
		t.Fatalf("page = %#v", page)
	}
}

func TestGetPlatformAppsUsesValidatedResource(t *testing.T) {
	paths := []string{}
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		paths = append(paths, req.URL.Path)
		switch req.URL.Path {
		case "/v1beta1/projects/123456/androidApps/a":
			return jsonHTTPResponse(http.StatusOK, `{"name":"projects/123456/androidApps/a","appId":"android-id","projectId":"demo","packageName":"com.example"}`, ""), nil
		case "/v1beta1/projects/demo/iosApps/i":
			return jsonHTTPResponse(http.StatusOK, `{"name":"projects/demo/iosApps/i","appId":"ios-id","projectId":"demo","bundleId":"com.example.ios"}`, ""), nil
		case "/v1beta1/projects/demo/webApps/w":
			return jsonHTTPResponse(http.StatusOK, `{"name":"projects/demo/webApps/w","appId":"web-id","projectId":"demo","appUrls":["https://example.com"]}`, ""), nil
		default:
			t.Fatalf("unexpected path %q", req.URL.Path)
			return nil, errors.New("unexpected path")
		}
	})})
	android, err := svc.GetAndroidApp(context.Background(), "demo", "projects/123456/androidApps/a")
	if err != nil || android.PackageName != "com.example" {
		t.Fatalf("android = %#v, %v", android, err)
	}
	ios, err := svc.GetIOSApp(context.Background(), "demo", "projects/demo/iosApps/i")
	if err != nil || ios.BundleID != "com.example.ios" {
		t.Fatalf("ios = %#v, %v", ios, err)
	}
	web, err := svc.GetWebApp(context.Background(), "demo", "projects/demo/webApps/w")
	if err != nil || len(web.AppURLs) != 1 {
		t.Fatalf("web = %#v, %v", web, err)
	}
	if len(paths) != 3 {
		t.Fatalf("paths = %v", paths)
	}
}

func TestGetAppConfigDecodesNativeBytesAndReadsWebJSON(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v1beta1/projects/demo/androidApps/a/config":
			return jsonHTTPResponse(http.StatusOK, `{"configFilename":"google-services.json","configFileContents":"eyJvayI6dHJ1ZX0K"}`, ""), nil
		case "/v1beta1/projects/demo/iosApps/i/config":
			return jsonHTTPResponse(http.StatusOK, `{"configFilename":"GoogleService-Info.plist","configFileContents":"PHBsaXN0Lz4K"}`, ""), nil
		case "/v1beta1/projects/demo/webApps/w/config":
			return jsonHTTPResponse(http.StatusOK, `{"projectId":"demo","appId":"web-id","apiKey":"key"}`, ""), nil
		default:
			t.Fatalf("unexpected path %q", req.URL.Path)
			return nil, io.EOF
		}
	})})
	native, err := svc.GetNativeAppConfig(context.Background(), "demo", "projects/demo/androidApps/a", AppPlatformAndroid)
	if err != nil || native.ConfigFilename != "google-services.json" || string(native.ConfigFileContents) != "{\"ok\":true}\n" {
		t.Fatalf("native config = %#v, %v", native, err)
	}
	ios, err := svc.GetNativeAppConfig(context.Background(), "demo", "projects/demo/iosApps/i", AppPlatformIOS)
	if err != nil || ios.ConfigFilename != "GoogleService-Info.plist" || string(ios.ConfigFileContents) != "<plist/>\n" {
		t.Fatalf("iOS config = %#v, %v", ios, err)
	}
	web, err := svc.GetWebAppConfig(context.Background(), "demo", "projects/demo/webApps/w")
	if err != nil || web.APIKey != "key" {
		t.Fatalf("web config = %#v, %v", web, err)
	}
}

func TestGetAppRejectsMalformedOrWrongPlatformResourceWithoutRequest(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("unexpected request")
		return nil, io.EOF
	})})
	for _, resource := range []string{"projects/demo/webApps/a", "not-a-resource"} {
		_, err := svc.GetAndroidApp(context.Background(), "demo", resource)
		var resourceErr *AppResourceError
		if err == nil || !errors.As(err, &resourceErr) {
			t.Errorf("GetAndroidApp(%q) error = %v", resource, err)
		}
	}
}

func TestListAppsReportsTypedAPIError(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonHTTPResponse(http.StatusForbidden, `{"error":{"message":"denied"}}`, ""), nil
	})})
	_, err := svc.ListApps(context.Background(), "demo", ListAppsOptions{})
	var apiErr *APIError
	if err == nil || !errors.As(err, &apiErr) || apiErr.Service != "firebase_management" || apiErr.Operation != "list_apps" || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("error = %#v", err)
	}
}
