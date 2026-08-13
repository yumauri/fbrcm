package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestListRemoteConfigExperimentsUsesPagedListResourcesWithoutHydration(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	saveParametersCacheRaw(t, "demo", "etag-9", json.RawMessage(`{
		"version":{"versionNumber":"9"},
		"parameters":{
			"signup_message":{
				"defaultValue":{"experimentValue":{
					"experimentId":"exp-1",
					"variantValue":[
						{"variantId":"control","value":""},
						{"variantId":"enabled","noChange":true}
					],
					"exposurePercent":0
				}},
				"valueType":"STRING"
			}
		}
	}`))
	var requests []string
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.RequestURI())
		switch req.URL.Query().Get("pageToken") {
		case "":
			return jsonResponse(http.StatusOK, `{
				"experiments":[{
					"name":"projects/123/namespaces/firebase/experiments/exp-1",
					"definition":{"displayName":"Signup","variants":[{"name":"Baseline","weight":1}]},
					"state":"RUNNING"
				}],
				"nextPageToken":"next"
			}`, ""), nil
		case "next":
			return jsonResponse(http.StatusOK, `{
				"experiments":[{
					"name":"projects/123/namespaces/firebase/experiments/exp-2",
					"definition":{"displayName":"Funding","variants":[{"name":"Enabled","weight":2}]},
					"state":"ENDED"
				}]
			}`, ""), nil
		default:
			return nil, errors.New("unexpected experiments request: " + req.URL.String())
		}
	})})
	injectFirebaseService(t, svc, "main", client)

	result, err := svc.ListRemoteConfigExperiments(
		context.Background(),
		Project{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"},
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Experiments) != 2 ||
		result.Experiments[0].Definition.DisplayName != "Signup" ||
		result.Experiments[1].Definition.Variants[0].Name != "Enabled" ||
		len(result.Experiments[0].References) != 1 ||
		result.Experiments[0].References[0].Percentage == nil ||
		*result.Experiments[0].References[0].Percentage != 0 ||
		result.Experiments[0].References[0].Variants[0].Value == nil ||
		*result.Experiments[0].References[0].Variants[0].Value != "" ||
		result.Experiments[0].References[0].Variants[1].NoChange == nil ||
		!*result.Experiments[0].References[0].Variants[1].NoChange {
		t.Fatalf("experiments = %#v", result.Experiments)
	}
	wantRequests := []string{
		"/v1/projects/123/namespaces/firebase/experiments?pageSize=100",
		"/v1/projects/123/namespaces/firebase/experiments?pageSize=100&pageToken=next",
	}
	if !reflect.DeepEqual(requests, wantRequests) {
		t.Fatalf("requests = %#v, want only list pages %#v", requests, wantRequests)
	}
}

func TestListRemoteConfigRolloutsUsesListResourcesWithoutHydration(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	saveParametersCacheRaw(t, "demo", "etag-9", json.RawMessage(`{
		"version":{"versionNumber":"9"},
		"parameters":{
			"funding":{
				"defaultValue":{"rolloutValue":{"rolloutId":"rollout-1","value":"20","percent":10}},
				"valueType":"NUMBER"
			}
		}
	}`))
	var requests []string
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.RequestURI())
		if req.URL.Path != "/v1/projects/123/namespaces/firebase/rollouts" {
			return nil, errors.New("unexpected rollout request: " + req.URL.String())
		}
		return jsonResponse(http.StatusOK, `{
			"rollouts":[{
				"name":"projects/123/namespaces/firebase/rollouts/rollout-1",
				"definition":{
					"displayName":"Funding",
					"controlVariant":{"name":"Control","weight":90},
					"enabledVariant":{"name":"Enabled","weight":10}
				},
				"state":"RUNNING"
			}]
		}`, ""), nil
	})})
	injectFirebaseService(t, svc, "main", client)

	result, err := svc.ListRemoteConfigRollouts(
		context.Background(),
		Project{Name: "Demo", ProjectID: "demo", ProjectNumber: "123"},
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rollouts) != 1 ||
		result.Rollouts[0].Definition.EnabledVariant.Name != "Enabled" ||
		len(result.Rollouts[0].References) != 1 ||
		result.Rollouts[0].References[0].Parameter != "funding" {
		t.Fatalf("rollouts = %#v", result.Rollouts)
	}
	wantRequests := []string{"/v1/projects/123/namespaces/firebase/rollouts?pageSize=100"}
	if !reflect.DeepEqual(requests, wantRequests) {
		t.Fatalf("requests = %#v, want only the list page %#v", requests, wantRequests)
	}
}

func TestCollectManagedFeatureReferencesIncludesGroupsDefaultsAndConditionOrder(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Conditions: []firebase.RemoteConfigCondition{{Name: "beta", TagColor: "GREEN"}, {Name: "staff", TagColor: "PURPLE"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"root_flag": {
				ValueType: "BOOLEAN",
				DefaultValue: &firebase.RemoteConfigValue{
					RolloutValue: json.RawMessage(`{"rolloutId":"rollout_1","value":"true","percent":12.5}`),
				},
			},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"onboarding": {Parameters: map[string]firebase.RemoteConfigParam{
				"provider": {
					ValueType: "STRING",
					ConditionalValues: map[string]firebase.RemoteConfigValue{
						"staff": {PersonalizationValue: json.RawMessage(`{"personalizationId":"personalization_1"}`)},
						"beta":  {RolloutValue: json.RawMessage(`{"rolloutId":"rollout_1","value":"kyc2","percent":10}`)},
					},
				},
			}},
		},
	}

	rollouts, err := collectRolloutReferences(cfg)
	if err != nil {
		t.Fatal(err)
	}
	wantRollouts := []ManagedValueReference{
		{
			Group: "(root)", Parameter: "root_flag", Default: true, ValueType: "BOOLEAN",
			Value: new("true"), Percentage: new(12.5),
		},
		{
			Group: "onboarding", Parameter: "provider", Condition: "beta", ConditionColor: "GREEN",
			ValueType: "STRING", Value: new("kyc2"), Percentage: new(10.0),
		},
	}
	if !reflect.DeepEqual(rollouts["rollout_1"], wantRollouts) {
		t.Fatalf("rollout references = %#v, want %#v", rollouts["rollout_1"], wantRollouts)
	}

	personalizations, err := collectPersonalizationReferences(cfg)
	if err != nil {
		t.Fatal(err)
	}
	wantPersonalizations := []ManagedValueReference{{
		Group: "onboarding", Parameter: "provider", Condition: "staff", ConditionColor: "PURPLE", ValueType: "STRING",
	}}
	if !reflect.DeepEqual(personalizations["personalization_1"], wantPersonalizations) {
		t.Fatalf("personalization references = %#v, want %#v", personalizations["personalization_1"], wantPersonalizations)
	}

	encoded, err := json.Marshal(personalizations["personalization_1"][0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "color") {
		t.Fatalf("display-only condition color leaked into JSON: %s", encoded)
	}
}

func TestGetRemoteConfigPersonalizationRequiresExactCase(t *testing.T) {
	svc := setupCoreTestEnv(t)
	saveParametersCacheRaw(t, "demo", "etag-1", json.RawMessage(`{
		"version":{"versionNumber":"1"},
		"parameters":{"flag":{"defaultValue":{"personalizationValue":{"personalizationId":"Personal_1"}}}}
	}`))
	project := Project{Name: "Demo", ProjectID: "demo"}
	if got, _, err := svc.GetRemoteConfigPersonalization(context.Background(), project, "Personal_1", false); err != nil || got.ID != "Personal_1" {
		t.Fatalf("exact personalization = %#v, %v", got, err)
	}
	_, _, err := svc.GetRemoteConfigPersonalization(context.Background(), project, "personal_1", false)
	var lookup *ManagedFeatureLookupError
	if !errors.As(err, &lookup) || lookup.ID != "personal_1" {
		t.Fatalf("case-mismatched personalization error = %#v", err)
	}
}

func TestCollectManagedFeatureReferencesRejectsUnexpectedWireShape(t *testing.T) {
	cfg := &firebase.RemoteConfig{Parameters: map[string]firebase.RemoteConfigParam{
		"flag": {DefaultValue: &firebase.RemoteConfigValue{RolloutValue: json.RawMessage(`{"percent":"ten"}`)}},
	}}
	_, err := collectRolloutReferences(cfg)
	if err == nil || !strings.Contains(err.Error(), "flag default value") {
		t.Fatalf("collectRolloutReferences error = %v", err)
	}
}
