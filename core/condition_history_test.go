package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
)

func conditionHistoryConfig(t *testing.T, version string, conditions ...firebase.RemoteConfigCondition) string {
	t.Helper()
	raw, err := json.Marshal(firebase.RemoteConfig{
		Version: firebase.RemoteConfigVersion{VersionNumber: version}, Conditions: conditions,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestConditionHistoryTracksDefinitionAndPriorityChanges(t *testing.T) {
	base := firebase.RemoteConfigCondition{Name: "Beta Users", Expression: "true", TagColor: "BLUE"}
	expression := base
	expression.Expression = "false"
	color := expression
	color.TagColor = "GREEN"
	other := firebase.RemoteConfigCondition{Name: "Other", Expression: "true"}
	svc := parameterHistoryTestService(t, map[string]string{
		"5": conditionHistoryConfig(t, "5", other, color),
		"4": conditionHistoryConfig(t, "4", color, other),
		"3": conditionHistoryConfig(t, "3", expression, other),
		"2": conditionHistoryConfig(t, "2", base, other),
		"1": conditionHistoryConfig(t, "1", other),
	})
	result, err := svc.GetRemoteConfigConditionHistory(context.Background(), "demo", base.Name, ConditionHistoryOptions{Limit: 1, All: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Condition != base.Name || result.AtVersion != "5" || result.ScannedVersionCount != 5 || !result.HistoryExhausted || len(result.Changes) != 4 {
		t.Fatalf("history = %#v", result)
	}
	if result.Boundary == nil || result.Boundary.Present || result.Boundary.Version != "1" {
		t.Fatalf("boundary = %#v", result.Boundary)
	}
	for i, change := range result.Changes {
		if change.Version.VersionNumber != []string{"5", "4", "3", "2"}[i] || change.Version.UpdateUser.Email != "user"+change.Version.VersionNumber+"@example.com" {
			t.Fatalf("change %d attribution = %#v", i, change)
		}
	}
	if result.Changes[0].Change.PreviousPosition != 1 || result.Changes[0].Change.FinalPosition != 2 || result.Changes[0].Change.Kind != rcdiff.ChangeChanged {
		t.Fatalf("priority change = %#v", result.Changes[0])
	}
	if result.Changes[1].Change.Current.TagColor != "BLUE" || result.Changes[1].Change.Final.TagColor != "GREEN" || result.Changes[2].Change.Current.Expression != "true" || result.Changes[2].Change.Final.Expression != "false" {
		t.Fatalf("definition changes = %#v", result.Changes)
	}
	if result.Changes[3].Change.Kind != rcdiff.ChangeAdded {
		t.Fatalf("creation = %#v", result.Changes[3])
	}
}

func TestConditionHistorySkipsParameterUsageChangesAndHonorsLimit(t *testing.T) {
	condition := firebase.RemoteConfigCondition{Name: "beta", Expression: "true"}
	newest := conditionHistoryConfig(t, "5", condition)
	newest = strings.TrimSuffix(newest, "}") + `,"parameters":{"flag":{"conditionalValues":{"beta":{"value":"true"}}}}}`
	svc := parameterHistoryTestService(t, map[string]string{
		"5": newest,
		"4": conditionHistoryConfig(t, "4", condition),
		"3": conditionHistoryConfig(t, "3"),
		"2": conditionHistoryConfig(t, "2"),
		"1": conditionHistoryConfig(t, "1"),
	})
	result, err := svc.GetRemoteConfigConditionHistory(context.Background(), "demo", "beta", ConditionHistoryOptions{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if result.ScannedVersionCount != 3 || result.HistoryExhausted || result.Boundary != nil || len(result.Changes) != 1 || result.Changes[0].PreviousVersion != "3" || result.Changes[0].Version.VersionNumber != "4" {
		t.Fatalf("history = %#v", result)
	}
}

func TestConditionHistoryReportsRetentionBoundaryWithoutInventingCreation(t *testing.T) {
	condition := firebase.RemoteConfigCondition{Name: "beta", Expression: "true"}
	for _, versions := range [][]string{{"1"}, {"2", "1"}} {
		t.Run(strings.Join(versions, ","), func(t *testing.T) {
			configs := make(map[string]string)
			for _, version := range versions {
				configs[version] = conditionHistoryConfig(t, version, condition)
			}
			svc := parameterHistoryTestService(t, configs)
			result, err := svc.GetRemoteConfigConditionHistory(context.Background(), "demo", "beta", ConditionHistoryOptions{Limit: 1, All: true})
			if err != nil {
				t.Fatal(err)
			}
			if !result.HistoryExhausted || len(result.Changes) != 0 || result.Boundary == nil || !result.Boundary.Present || result.Boundary.Version != "1" {
				t.Fatalf("history = %#v", result)
			}
		})
	}
}

func TestConditionHistoryTreatsRenamesAsSeparateNames(t *testing.T) {
	svc := parameterHistoryTestService(t, map[string]string{
		"3": conditionHistoryConfig(t, "3", firebase.RemoteConfigCondition{Name: "new", Expression: "true"}),
		"2": conditionHistoryConfig(t, "2", firebase.RemoteConfigCondition{Name: "old", Expression: "true"}),
		"1": conditionHistoryConfig(t, "1"),
	})
	for _, test := range []struct {
		name  string
		kinds []rcdiff.ChangeKind
	}{
		{"new", []rcdiff.ChangeKind{rcdiff.ChangeAdded}},
		{"old", []rcdiff.ChangeKind{rcdiff.ChangeRemoved, rcdiff.ChangeAdded}},
	} {
		result, err := svc.GetRemoteConfigConditionHistory(context.Background(), "demo", test.name, ConditionHistoryOptions{Limit: 1, All: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Changes) != len(test.kinds) {
			t.Fatalf("%s changes = %#v", test.name, result.Changes)
		}
		for i, kind := range test.kinds {
			if result.Changes[i].Change.Kind != kind {
				t.Fatalf("%s change %d = %#v", test.name, i, result.Changes[i])
			}
		}
	}
}

func TestConditionHistoryRequiresExactNameAndReturnsTypedNotFound(t *testing.T) {
	svc := parameterHistoryTestService(t, map[string]string{
		"1": conditionHistoryConfig(t, "1", firebase.RemoteConfigCondition{Name: "Beta Users", Expression: "true"}),
	})
	for _, name := range []string{"missing", "beta users", " Beta Users", "Beta Users "} {
		_, err := svc.GetRemoteConfigConditionHistory(context.Background(), "demo", name, ConditionHistoryOptions{Limit: 1})
		var lookup *RemoteConfigConditionLookupError
		if !errors.As(err, &lookup) || lookup.ProjectID != "demo" || lookup.Condition != name {
			t.Fatalf("%q error = %#v, lookup = %#v", name, err, lookup)
		}
	}
}

func TestConditionHistoryPaginatesToAnchorAndStopsAtLimit(t *testing.T) {
	for _, at := range []string{"previous", "current~1", "+0003"} {
		t.Run(at, func(t *testing.T) {
			listCalls, getCalls := 0, 0
			svc := parameterHistoryServiceWithTransport(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "listVersions") {
					listCalls++
					if req.URL.Query().Get("pageToken") == "" {
						return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"5"}],"nextPageToken":"older"}`, ""), nil
					}
					return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"3"},{"versionNumber":"1"}],"nextPageToken":"unneeded"}`, ""), nil
				}
				getCalls++
				version := req.URL.Query().Get("versionNumber")
				if version == "3" {
					return jsonResponse(http.StatusOK, conditionHistoryConfig(t, version, firebase.RemoteConfigCondition{Name: "beta", Expression: "true"}), ""), nil
				}
				if version != "1" {
					t.Fatalf("unexpected fetched version %s", version)
				}
				return jsonResponse(http.StatusOK, conditionHistoryConfig(t, version), ""), nil
			}))
			result, err := svc.GetRemoteConfigConditionHistory(context.Background(), "demo", "beta", ConditionHistoryOptions{At: at, Limit: 1})
			if err != nil {
				t.Fatal(err)
			}
			if result.AtVersion != "3" || result.HistoryExhausted || len(result.Changes) != 1 || result.Changes[0].PreviousVersion != "1" || listCalls != 2 || getCalls != 2 {
				t.Fatalf("history = %#v, calls = %d metadata / %d templates", result, listCalls, getCalls)
			}
		})
	}
}
