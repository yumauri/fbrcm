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
	rcdiff "github.com/yumauri/fbrcm/core/rc/diff"
	rcpromote "github.com/yumauri/fbrcm/core/rc/promote"
)

func TestProjectPromotionPreviewReportsDependenciesAndSavesDraft(t *testing.T) {
	svc := setupCoreTestEnv(t)
	source := Project{Name: "Development", ProjectID: "dev"}
	target := Project{Name: "Production", ProjectID: "prod"}
	sourceCfg := &firebase.RemoteConfig{
		Version:    firebase.RemoteConfigVersion{VersionNumber: "7"},
		Conditions: []firebase.RemoteConfigCondition{{Name: "beta", Expression: "true"}},
		Parameters: map[string]firebase.RemoteConfigParam{
			"flag": {
				ValueType:    "STRING",
				DefaultValue: &firebase.RemoteConfigValue{Value: "off"},
				ConditionalValues: map[string]firebase.RemoteConfigValue{
					"beta": {Value: "on"},
				},
			},
		},
	}
	targetCfg := &firebase.RemoteConfig{Version: firebase.RemoteConfigVersion{VersionNumber: "3"}, Parameters: map[string]firebase.RemoteConfigParam{}}
	savePromotionCache(t, source.ProjectID, "etag-dev", sourceCfg)
	savePromotionCache(t, target.ProjectID, "etag-prod", targetCfg)

	plan, err := svc.PrepareProjectPromotion(context.Background(), source, target, ProjectPromotionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	requested := map[rcpromote.ItemID]bool{{Kind: rcdiff.ItemParameter, Name: "flag"}: true}
	preview, err := svc.PreviewProjectPromotion(plan, requested, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Required) != 1 || preview.Required[0].ID.Kind != rcdiff.ItemCondition || preview.Required[0].ID.Name != "beta" {
		t.Fatalf("required = %#v, want condition beta", preview.Required)
	}
	result, err := svc.SaveProjectPromotionDraft(preview)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Drafted || !result.HasDraft || result.Tree == nil {
		t.Fatalf("result = %#v, want saved draft tree", result)
	}
	record, ok, err := svc.LoadDraftRecord(target.ProjectID)
	if err != nil || !ok {
		t.Fatalf("LoadDraftRecord = ok:%v err:%v", ok, err)
	}
	if record.BaseVersion != "3" || record.BaseETag != "etag-prod" {
		t.Fatalf("draft base = version:%q etag:%q", record.BaseVersion, record.BaseETag)
	}
	saved, err := firebase.ParseRemoteConfig(record.RemoteConfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := saved.Parameters["flag"]; !ok || len(saved.Conditions) != 1 || saved.Conditions[0].Name != "beta" {
		t.Fatalf("saved promotion = %#v", saved)
	}
}

func TestProjectPromotionComposesOntoExistingTargetDraft(t *testing.T) {
	svc := setupCoreTestEnv(t)
	source := Project{Name: "Development", ProjectID: "dev"}
	target := Project{Name: "Production", ProjectID: "prod"}
	sourceCfg := promotionConfig("8", map[string]string{"promoted": "yes"})
	targetCfg := promotionConfig("4", map[string]string{"published": "yes"})
	draftCfg := promotionConfig("4", map[string]string{"published": "yes", "local": "draft"})
	savePromotionCache(t, source.ProjectID, "etag-dev", sourceCfg)
	savePromotionCache(t, target.ProjectID, "etag-prod", targetCfg)
	draftRaw := marshalPromotionConfig(t, draftCfg)
	if err := svc.SaveDraft(target.ProjectID, draftRaw); err != nil {
		t.Fatal(err)
	}
	before, _, err := svc.LoadDraftRecord(target.ProjectID)
	if err != nil {
		t.Fatal(err)
	}

	plan, err := svc.PrepareProjectPromotion(context.Background(), source, target, ProjectPromotionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Target.HasDraft || plan.Target.Source != "draft" {
		t.Fatalf("target snapshot = %#v, want draft", plan.Target)
	}
	preview, err := svc.PreviewProjectPromotion(plan, map[rcpromote.ItemID]bool{{Kind: rcdiff.ItemParameter, Name: "promoted"}: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveProjectPromotionDraft(preview); err != nil {
		t.Fatal(err)
	}
	after, ok, err := svc.LoadDraftRecord(target.ProjectID)
	if err != nil || !ok {
		t.Fatalf("LoadDraftRecord = ok:%v err:%v", ok, err)
	}
	if after.BaseVersion != before.BaseVersion || after.BaseETag != before.BaseETag || !after.CreatedAt.Equal(before.CreatedAt) {
		t.Fatalf("draft base changed: before=%#v after=%#v", before, after)
	}
	saved, err := firebase.ParseRemoteConfig(after.RemoteConfig)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"published", "local", "promoted"} {
		if _, ok := saved.Parameters[key]; !ok {
			t.Fatalf("saved draft missing %q: %#v", key, saved.Parameters)
		}
	}
}

func TestProjectPromotionRequiresExplicitPruneForTargetOnlyItem(t *testing.T) {
	svc := setupCoreTestEnv(t)
	source := Project{ProjectID: "dev"}
	target := Project{ProjectID: "prod"}
	savePromotionCache(t, source.ProjectID, "etag-dev", promotionConfig("2", nil))
	savePromotionCache(t, target.ProjectID, "etag-prod", promotionConfig("5", map[string]string{"obsolete": "yes"}))
	plan, err := svc.PrepareProjectPromotion(context.Background(), source, target, ProjectPromotionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	id := rcpromote.ItemID{Kind: rcdiff.ItemParameter, Name: "obsolete"}
	withoutPrune, err := svc.PreviewProjectPromotion(plan, map[rcpromote.ItemID]bool{id: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	if withoutPrune.HasChanges || len(withoutPrune.Requested) != 0 {
		t.Fatalf("without prune = %#v, want no requested changes", withoutPrune)
	}
	withPrune, err := svc.PreviewProjectPromotion(plan, map[rcpromote.ItemID]bool{id: true}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !withPrune.HasChanges || !withPrune.Requested[id] {
		t.Fatalf("with prune = %#v, want selected removal", withPrune)
	}
}

func TestPublishProjectPromotionStagesValidatesAndPublishes(t *testing.T) {
	svc := setupCoreTestEnv(t)
	seedAuthAndProject(t, svc, "main", "prod")
	source := Project{Name: "Development", ProjectID: "dev"}
	target := Project{Name: "Production", ProjectID: "prod", AuthID: "main", DiscoveredBy: []string{"main"}}
	savePromotionCache(t, source.ProjectID, `"etag-dev"`, promotionConfig("7", map[string]string{"flag": "promoted"}))
	savePromotionCache(t, target.ProjectID, `"etag-1"`, promotionConfig("1", map[string]string{"flag": "current"}))

	validated, published := false, false
	client := firebase.NewServiceWithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch {
		case req.Method == http.MethodGet && strings.Contains(req.URL.Path, "listVersions"):
			return jsonResponse(http.StatusOK, `{"versions":[{"versionNumber":"1"}]}`, ""), nil
		case req.Method == http.MethodPut && strings.Contains(req.URL.RawQuery, "validateOnly=true"):
			validated = true
			return jsonResponse(http.StatusOK, string(remoteConfigRaw("1", map[string]string{"flag": "promoted"})), `"etag-1"`), nil
		case req.Method == http.MethodPut:
			published = true
			var payload firebase.RemoteConfig
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if got := payload.Parameters["flag"].DefaultValue.Value; got != "promoted" {
				t.Fatalf("published flag = %q", got)
			}
			return jsonResponse(http.StatusOK, string(remoteConfigRaw("2", map[string]string{"flag": "promoted"})), `"etag-2"`), nil
		default:
			return nil, io.EOF
		}
	})})
	injectFirebaseService(t, svc, "main", client)

	plan, err := svc.PrepareProjectPromotion(context.Background(), source, target, ProjectPromotionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := svc.PreviewProjectPromotion(plan, map[rcpromote.ItemID]bool{{Kind: rcdiff.ItemParameter, Name: "flag"}: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.PublishProjectPromotion(context.Background(), preview)
	if err != nil {
		t.Fatal(err)
	}
	if !validated || !published || !result.Published || result.HasDraft || result.Tree == nil || result.Tree.Version != "2" {
		t.Fatalf("validated=%v published=%v result=%#v", validated, published, result)
	}
	if _, ok, err := svc.LoadDraft(target.ProjectID); err != nil || ok {
		t.Fatalf("draft after publish = ok:%v err:%v", ok, err)
	}
}

func savePromotionCache(t *testing.T, projectID, etag string, cfg *firebase.RemoteConfig) {
	t.Helper()
	if err := config.SaveParametersCache(projectID, &config.ParametersCache{ETag: etag, CachedAt: time.Now().UTC(), RemoteConfig: marshalPromotionConfig(t, cfg)}); err != nil {
		t.Fatal(err)
	}
}

func marshalPromotionConfig(t *testing.T, cfg *firebase.RemoteConfig) json.RawMessage {
	t.Helper()
	raw, err := firebase.MarshalRemoteConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func promotionConfig(version string, values map[string]string) *firebase.RemoteConfig {
	cfg := &firebase.RemoteConfig{Version: firebase.RemoteConfigVersion{VersionNumber: version}, Parameters: map[string]firebase.RemoteConfigParam{}}
	for key, value := range values {
		cfg.Parameters[key] = firebase.RemoteConfigParam{ValueType: "STRING", DefaultValue: &firebase.RemoteConfigValue{Value: value}}
	}
	return cfg
}
