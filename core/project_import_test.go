package core

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
)

func TestPrepareAndExecuteProjectImportDraft(t *testing.T) {
	svc := setupCoreTestEnv(t)
	base := remoteConfigRaw("1", map[string]string{"flag": "current"})
	cache := &config.ParametersCache{ETag: "etag-1", CachedAt: time.Now().UTC(), RemoteConfig: base}
	if err := config.SaveParametersCache("demo", cache); err != nil {
		t.Fatalf("SaveParametersCache = %v", err)
	}

	source := []byte(`{"parameters":{"flag":{"defaultValue":{"value":"imported"}},"new":{"defaultValue":{"value":"added"}}}}`)
	plan, err := svc.PrepareProjectImport(context.Background(), Project{Name: "Demo", ProjectID: "demo"}, source, ProjectImportOptions{
		Strategy:          ProjectImportMerge,
		DefaultResolution: ProjectImportUseImported,
	})
	if err != nil {
		t.Fatalf("PrepareProjectImport = %v", err)
	}
	if !plan.HasChanges || plan.HasDraft || len(plan.Conflicts) != 1 {
		t.Fatalf("plan changes=%v draft=%v conflicts=%d", plan.HasChanges, plan.HasDraft, len(plan.Conflicts))
	}
	result, err := svc.ExecuteProjectImport(context.Background(), plan, false)
	if err != nil {
		t.Fatalf("ExecuteProjectImport draft = %v", err)
	}
	if !result.Drafted || result.Published || result.Tree == nil {
		t.Fatalf("result = %+v", result)
	}
	draftRaw, ok, err := svc.LoadDraft("demo")
	if err != nil || !ok {
		t.Fatalf("LoadDraft = ok:%v err:%v", ok, err)
	}
	cfg, err := firebase.ParseRemoteConfig(draftRaw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig draft = %v", err)
	}
	if got := cfg.Parameters["flag"].DefaultValue.Value; got != "imported" {
		t.Fatalf("flag = %q", got)
	}
	if got := cfg.Parameters["new"].DefaultValue.Value; got != "added" {
		t.Fatalf("new = %q", got)
	}
	if cfg.Version.VersionNumber != "1" {
		t.Fatalf("draft version = %q, want 1", cfg.Version.VersionNumber)
	}
}

func TestExecuteProjectImportPublishesValidatedCandidate(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "demo")
	base := remoteConfigRaw("1", map[string]string{"flag": "current"})
	cache := &config.ParametersCache{ETag: `"etag-1"`, CachedAt: time.Now().UTC(), RemoteConfig: base}
	if err := config.SaveParametersCache("demo", cache); err != nil {
		t.Fatalf("SaveParametersCache = %v", err)
	}
	validated, published := false, false
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions"):
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"1"}]}`, ""), nil
		case req.Method == http.MethodPut && strings.Contains(req.URL.RawQuery, "validateOnly=true"):
			validated = true
			return jsonResponse(http.StatusOK, `{"parameters":{"flag":{"defaultValue":{"value":"imported"}}}}`, `"etag-1"`), nil
		case req.Method == http.MethodPut:
			published = true
			var payload firebase.RemoteConfig
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatalf("decode publish payload = %v", err)
			}
			if got := payload.Parameters["flag"].DefaultValue.Value; got != "imported" {
				t.Fatalf("publish flag = %q", got)
			}
			return jsonResponse(http.StatusOK, string(remoteConfigRaw("2", map[string]string{"flag": "imported"})), `"etag-2"`), nil
		default:
			return nil, io.EOF
		}
	})})
	injectFirebaseService(t, svc, "main", client)

	plan, err := svc.PrepareProjectImport(context.Background(), Project{Name: "Demo", ProjectID: "demo"}, []byte(`{"parameters":{"flag":{"defaultValue":{"value":"imported"}}}}`), ProjectImportOptions{Strategy: ProjectImportReplace})
	if err != nil {
		t.Fatalf("PrepareProjectImport = %v", err)
	}
	result, err := svc.ExecuteProjectImport(context.Background(), plan, true)
	if err != nil {
		t.Fatalf("ExecuteProjectImport publish = %v", err)
	}
	if !validated || !published || !result.Published || result.Drafted || result.Tree.Version != "2" {
		t.Fatalf("validated=%v published=%v result=%+v", validated, published, result)
	}
}

func TestProjectImportUpdatesExistingDraftAndRejectsDirectPublish(t *testing.T) {
	svc := setupCoreTestEnv(t)
	base := remoteConfigRaw("1", map[string]string{"flag": "current"})
	cache := &config.ParametersCache{ETag: "etag-1", CachedAt: time.Now().UTC(), RemoteConfig: base}
	if err := config.SaveParametersCache("demo", cache); err != nil {
		t.Fatalf("SaveParametersCache = %v", err)
	}
	if err := svc.SaveDraft("demo", remoteConfigRaw("1", map[string]string{"flag": "draft"})); err != nil {
		t.Fatalf("SaveDraft = %v", err)
	}

	plan, err := svc.PrepareProjectImport(context.Background(), Project{Name: "Demo", ProjectID: "demo"}, []byte(`{"parameters":{"other":{"defaultValue":{"value":"x"}}}}`), ProjectImportOptions{Strategy: ProjectImportMerge})
	if err != nil {
		t.Fatalf("PrepareProjectImport = %v", err)
	}
	if !plan.HasDraft {
		t.Fatal("plan did not use existing draft")
	}
	if _, err := svc.ExecuteProjectImport(context.Background(), plan, true); err == nil {
		t.Fatal("direct publish with existing draft succeeded")
	}
	if _, err := svc.ExecuteProjectImport(context.Background(), plan, false); err != nil {
		t.Fatalf("update draft = %v", err)
	}
	draftRaw, _, _ := svc.LoadDraft("demo")
	cfg, _ := firebase.ParseRemoteConfig(draftRaw)
	if got := cfg.Parameters["flag"].DefaultValue.Value; got != "draft" {
		t.Fatalf("existing draft flag = %q", got)
	}
	if got := cfg.Parameters["other"].DefaultValue.Value; got != "x" {
		t.Fatalf("imported draft value = %q", got)
	}
}
