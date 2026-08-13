package firebase

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestListAndGetExperimentsUseNumericFirebaseProjectResource(t *testing.T) {
	var paths []string
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		paths = append(paths, req.URL.RequestURI())
		switch len(paths) {
		case 1:
			if req.URL.Query().Get("pageSize") != "100" || req.URL.Query().Get("pageToken") != "next" {
				t.Fatalf("list query = %q", req.URL.RawQuery)
			}
			return jsonHTTPResponse(http.StatusOK, `{"experiments":[{"name":"projects/123/namespaces/firebase/experiments/7","state":"RUNNING"}],"nextPageToken":"older"}`, ""), nil
		case 2:
			return jsonHTTPResponse(http.StatusOK, `{"name":"projects/123/namespaces/firebase/experiments/7","definition":{"displayName":"Signup","objectives":{"activationEvent":{"event":"viewed_signup"}},"variants":[{"name":"Baseline","weight":1},{"name":"Variant A","weight":1}]},"state":"RUNNING"}`, ""), nil
		default:
			return nil, io.EOF
		}
	})})

	page, err := svc.ListExperiments(context.Background(), "123", "demo", ListManagedFeaturesOptions{PageSize: 100, PageToken: "next"})
	if err != nil || page.NextPageToken != "older" || len(page.Experiments) != 1 {
		t.Fatalf("ListExperiments = %+v, %v", page, err)
	}
	experiment, err := svc.GetExperiment(context.Background(), "123", "demo", "7")
	if err != nil || experiment.Definition.DisplayName != "Signup" || experiment.Definition.Objectives.ActivationEvent.Event != "viewed_signup" || len(experiment.Definition.Variants) != 2 {
		t.Fatalf("GetExperiment = %+v, %v", experiment, err)
	}
	want := []string{
		"/v1/projects/123/namespaces/firebase/experiments?pageSize=100&pageToken=next",
		"/v1/projects/123/namespaces/firebase/experiments/7",
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("request paths = %#v, want %#v", paths, want)
	}
}

func TestListAndGetRolloutsDecodeDefinition(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v1/projects/949596266151/namespaces/firebase/rollouts":
			return jsonHTTPResponse(http.StatusOK, `{"rollouts":[{"name":"projects/949596266151/namespaces/firebase/rollouts/rollout_1"}]}`, ""), nil
		case "/v1/projects/949596266151/namespaces/firebase/rollouts/rollout_1":
			return jsonHTTPResponse(http.StatusOK, `{
				"name":"projects/949596266151/namespaces/firebase/rollouts/rollout_1",
				"definition":{
					"displayName":"Funding",
					"controlVariant":{"name":"Control","futureNestedField":"kept"},
					"enabledVariant":{"name":"Enabled"}
				},
				"state":"RUNNING",
				"createTime":"2026-07-01T09:10:11Z",
				"futureTopLevelField":{"enabled":true}
			}`, ""), nil
		default:
			return nil, io.EOF
		}
	})})

	page, err := svc.ListRollouts(context.Background(), "949596266151", "northstar-wallet", ListManagedFeaturesOptions{})
	if err != nil || len(page.Rollouts) != 1 {
		t.Fatalf("ListRollouts = %+v, %v", page, err)
	}
	rollout, err := svc.GetRollout(context.Background(), "949596266151", "northstar-wallet", "rollout_1")
	if err != nil || rollout.Definition.EnabledVariant.Name != "Enabled" || rollout.CreateTime != "2026-07-01T09:10:11Z" {
		t.Fatalf("GetRollout = %+v, %v", rollout, err)
	}
	encoded, err := json.Marshal(rollout)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"createTime":"2026-07-01T09:10:11Z"`, `"futureNestedField":"kept"`, `"futureTopLevelField":{"enabled":true}`} {
		if !strings.Contains(string(encoded), want) {
			t.Fatalf("marshaled rollout dropped %s: %s", want, encoded)
		}
	}
}

func TestManagedFeatureResourcesAcceptProjectID(t *testing.T) {
	var path string
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		path = req.URL.Path
		return jsonHTTPResponse(http.StatusOK, `{"experiments":[]}`, ""), nil
	})})
	if _, err := svc.ListExperiments(context.Background(), "demo-project", "demo-project", ListManagedFeaturesOptions{}); err != nil {
		t.Fatal(err)
	}
	if path != "/v1/projects/demo-project/namespaces/firebase/experiments" {
		t.Fatalf("project-ID path = %q", path)
	}
}

func TestDeleteManagedFeaturesUsesFirebaseResource(t *testing.T) {
	var requests []string
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", req.Method)
		}
		requests = append(requests, req.URL.Path)
		if strings.Contains(req.URL.Path, "/experiments/") {
			return jsonHTTPResponse(http.StatusOK, `{}`, ""), nil
		}
		return jsonHTTPResponse(http.StatusNoContent, "", ""), nil
	})})

	if err := svc.DeleteExperiment(context.Background(), "123", "demo", "7"); err != nil {
		t.Fatalf("DeleteExperiment = %v", err)
	}
	if err := svc.DeleteRollout(context.Background(), "123", "demo", "rollout_1"); err != nil {
		t.Fatalf("DeleteRollout = %v", err)
	}
	want := []string{
		"/v1/projects/123/namespaces/firebase/experiments/7",
		"/v1/projects/123/namespaces/firebase/rollouts/rollout_1",
	}
	if !reflect.DeepEqual(requests, want) {
		t.Fatalf("delete paths = %#v, want %#v", requests, want)
	}
}

func TestDeleteManagedFeatureReportsHTTPError(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonHTTPResponse(http.StatusConflict, `{"error":"running"}`, ""), nil
	})})
	err := svc.DeleteExperiment(context.Background(), "123", "demo", "7")
	if err == nil || !strings.Contains(err.Error(), "Conflict") || !strings.Contains(err.Error(), "running") {
		t.Fatalf("DeleteExperiment error = %v", err)
	}
}

func TestManagedFeatureAPIReportsHTTPAndResourceErrors(t *testing.T) {
	svc := NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonHTTPResponse(http.StatusForbidden, `{"error":"denied"}`, ""), nil
	})})

	if _, err := svc.ListExperiments(context.Background(), "", "demo", ListManagedFeaturesOptions{}); err == nil || !strings.Contains(err.Error(), "project identifier") {
		t.Fatalf("empty project identifier error = %v", err)
	}
	if _, err := svc.GetRollout(context.Background(), "123", "demo", "rollout_1"); err == nil || !strings.Contains(err.Error(), "Forbidden") {
		t.Fatalf("HTTP error = %v", err)
	}
}

func TestManagedFeatureID(t *testing.T) {
	if got := ManagedFeatureID("projects/123/namespaces/firebase/rollouts/rollout_1"); got != "rollout_1" {
		t.Fatalf("ManagedFeatureID = %q", got)
	}
}

func TestManagedFeatureResourceRejectsMalformedFullNameWithTypedError(t *testing.T) {
	_, err := managedFeatureResource("123", "experiments", "projects/other/namespaces/firebase/experiments/7")
	var resourceErr *ManagedFeatureResourceError
	if !errors.As(err, &resourceErr) || resourceErr.Collection != "experiments" || resourceErr.ItemID != "projects/other/namespaces/firebase/experiments/7" {
		t.Fatalf("managedFeatureResource error = %#v, want typed resource error", err)
	}
}

func TestManagedFeatureResourcePreservesBareIDWhitespace(t *testing.T) {
	got, err := managedFeatureResource("123", "experiments", " experiment_1 ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "projects/123/namespaces/firebase/experiments/%20experiment_1%20" {
		t.Fatalf("managedFeatureResource = %q", got)
	}
}
