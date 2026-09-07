package core

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

func TestGetRemoteConfigParameterHistorySkipsUnchangedPublications(t *testing.T) {
	svc := parameterHistoryTestService(t, map[string]string{
		"5": parameterHistoryConfig("5", "", "new"),
		"4": parameterHistoryConfig("4", "", "new"),
		"3": parameterHistoryConfig("3", "", "old"),
		"2": parameterHistoryConfig("2", "", "old"),
		"1": `{"version":{"versionNumber":"1"}}`,
	})

	result, err := svc.GetRemoteConfigParameterHistory(context.Background(), "demo", "flag", ParameterHistoryOptions{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if result.AtVersion != "5" || result.ScannedVersionCount != 3 || result.HistoryExhausted {
		t.Fatalf("history coverage = %#v", result)
	}
	if len(result.Changes) != 1 {
		t.Fatalf("changes = %#v, want one", result.Changes)
	}
	change := result.Changes[0]
	if change.PreviousVersion != "3" || change.Version.VersionNumber != "4" || change.Version.UpdateUser.Email != "user4@example.com" || change.Change.Kind != rcdiff.ChangeChanged {
		t.Fatalf("change = %#v", change)
	}
}

func TestGetRemoteConfigParameterHistoryScansAllAndReportsAbsentBoundary(t *testing.T) {
	svc := parameterHistoryTestService(t, map[string]string{
		"5": parameterHistoryConfig("5", "", "new"),
		"4": parameterHistoryConfig("4", "", "new"),
		"3": parameterHistoryConfig("3", "", "old"),
		"2": parameterHistoryConfig("2", "", "old"),
		"1": `{"version":{"versionNumber":"1"}}`,
	})

	result, err := svc.GetRemoteConfigParameterHistory(context.Background(), "demo", "flag", ParameterHistoryOptions{Limit: 1, All: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.HistoryExhausted || result.ScannedVersionCount != 5 || result.Boundary == nil || result.Boundary.Present || result.Boundary.Version != "1" {
		t.Fatalf("history boundary = %#v", result)
	}
	if len(result.Changes) != 2 || result.Changes[0].Version.VersionNumber != "4" || result.Changes[1].Version.VersionNumber != "2" || result.Changes[1].Change.Kind != rcdiff.ChangeAdded {
		t.Fatalf("changes = %#v", result.Changes)
	}
}

func TestGetRemoteConfigParameterHistoryReportsPresentBoundaryWithoutFalseAddition(t *testing.T) {
	svc := parameterHistoryTestService(t, map[string]string{
		"2": parameterHistoryConfig("2", "settings", "same"),
		"1": parameterHistoryConfig("1", "settings", "same"),
	})

	result, err := svc.GetRemoteConfigParameterHistory(context.Background(), "demo", "flag", ParameterHistoryOptions{Limit: 1, All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes) != 0 || result.Boundary == nil || !result.Boundary.Present || result.Boundary.Group != "settings" {
		t.Fatalf("history = %#v", result)
	}
	if result.AtGroup == nil || *result.AtGroup != "settings" {
		t.Fatalf("at group = %#v", result.AtGroup)
	}
}

func TestGetRemoteConfigParameterHistoryReturnsTypedNotFound(t *testing.T) {
	svc := parameterHistoryTestService(t, map[string]string{
		"2": `{"version":{"versionNumber":"2"}}`,
		"1": `{"version":{"versionNumber":"1"}}`,
	})

	_, err := svc.GetRemoteConfigParameterHistory(context.Background(), "demo", "missing", ParameterHistoryOptions{Limit: 1})
	var lookup *RemoteConfigParameterLookupError
	if !errors.As(err, &lookup) || lookup.ProjectID != "demo" || lookup.Parameter != "missing" {
		t.Fatalf("error = %#v, lookup = %#v", err, lookup)
	}
}

func TestGetRemoteConfigParameterHistoryStopsBeforeLoadingNextMetadataPage(t *testing.T) {
	listCalls, configCalls := 0, 0
	svc := parameterHistoryServiceWithTransport(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "listVersions") {
			listCalls++
			if token := req.URL.Query().Get("pageToken"); token != "" {
				t.Fatalf("unexpected metadata page token %q", token)
			}
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"5"},{"versionNumber":"4"}],"nextPageToken":"older"}`, ""), nil
		}
		configCalls++
		version := req.URL.Query().Get("versionNumber")
		values := map[string]string{"5": "new", "4": "old"}
		return jsonResponse(http.StatusOK, parameterHistoryConfig(version, "", values[version]), `"etag-`+version+`"`), nil
	}))

	result, err := svc.GetRemoteConfigParameterHistory(context.Background(), "demo", "flag", ParameterHistoryOptions{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if listCalls != 1 || configCalls != 2 {
		t.Fatalf("requests = %d metadata, %d configs; want 1 metadata, 2 configs", listCalls, configCalls)
	}
	if len(result.Changes) != 1 || result.Changes[0].PreviousVersion != "4" || result.Changes[0].Version.VersionNumber != "5" || result.HistoryExhausted {
		t.Fatalf("result = %#v", result)
	}
}

func TestGetRemoteConfigParameterHistoryContinuesAcrossMetadataPages(t *testing.T) {
	listCalls := 0
	svc := parameterHistoryServiceWithTransport(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "listVersions") {
			listCalls++
			if req.URL.Query().Get("pageToken") == "" {
				return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"4"},{"versionNumber":"3"}],"nextPageToken":"older"}`, ""), nil
			}
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"2"},{"versionNumber":"1"}]}`, ""), nil
		}
		version := req.URL.Query().Get("versionNumber")
		values := map[string]string{"4": "new", "3": "new", "2": "old", "1": "old"}
		return jsonResponse(http.StatusOK, parameterHistoryConfig(version, "", values[version]), `"etag-`+version+`"`), nil
	}))

	result, err := svc.GetRemoteConfigParameterHistory(context.Background(), "demo", "flag", ParameterHistoryOptions{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if listCalls != 2 || len(result.Changes) != 1 || result.Changes[0].PreviousVersion != "2" || result.Changes[0].Version.VersionNumber != "3" {
		t.Fatalf("list calls = %d, result = %#v", listCalls, result)
	}
}

func TestGetRemoteConfigParameterHistoryPaginatesToNumericAnchor(t *testing.T) {
	listCalls := 0
	svc := parameterHistoryServiceWithTransport(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "listVersions") {
			listCalls++
			if req.URL.Query().Get("pageToken") == "" {
				return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"5"},{"versionNumber":"4"}],"nextPageToken":"older"}`, ""), nil
			}
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"2"},{"versionNumber":"1"}]}`, ""), nil
		}
		version := req.URL.Query().Get("versionNumber")
		values := map[string]string{"2": "new", "1": "old"}
		return jsonResponse(http.StatusOK, parameterHistoryConfig(version, "", values[version]), `"etag-`+version+`"`), nil
	}))

	result, err := svc.GetRemoteConfigParameterHistory(context.Background(), "demo", "flag", ParameterHistoryOptions{At: "+0002", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if listCalls != 2 || result.AtVersion != "2" || len(result.Changes) != 1 || result.Changes[0].PreviousVersion != "1" {
		t.Fatalf("list calls = %d, result = %#v", listCalls, result)
	}
}

func parameterHistoryTestService(t *testing.T, configs map[string]string) *Core {
	t.Helper()
	return parameterHistoryServiceWithTransport(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions") {
			versions := make([]string, 0, len(configs))
			for _, version := range []string{"5", "4", "3", "2", "1"} {
				if _, ok := configs[version]; ok {
					versions = append(versions, `{"versionNumber":"`+version+`","updateTime":"2026-09-0`+version+`T10:00:00Z","updateUser":{"email":"user`+version+`@example.com"}}`)
				}
			}
			return jsonResponse(http.StatusOK, `{"versions":[`+strings.Join(versions, ",")+`]}`, ""), nil
		}
		if req.Method == http.MethodGet && strings.Contains(req.URL.Path, "/remoteConfig") {
			version := req.URL.Query().Get("versionNumber")
			config, ok := configs[version]
			if !ok {
				return nil, errors.New("unexpected version request: " + version)
			}
			return jsonResponse(http.StatusOK, config, `"etag-`+version+`"`), nil
		}
		return nil, errors.New("unexpected request: " + req.Method + " " + req.URL.String())
	}))
}

func parameterHistoryServiceWithTransport(t *testing.T, transport http.RoundTripper) *Core {
	t.Helper()
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: transport})
	injectFirebaseService(t, svc, "main", client)
	return svc
}

func parameterHistoryConfig(version, group, value string) string {
	parameter := `{"defaultValue":{"value":"` + value + `"},"valueType":"STRING"}`
	if group == "" {
		return `{"version":{"versionNumber":"` + version + `"},"parameters":{"flag":` + parameter + `}}`
	}
	return `{"version":{"versionNumber":"` + version + `"},"parameterGroups":{"` + group + `":{"parameters":{"flag":` + parameter + `}}}}`
}
