package draft

import (
	"context"
	"errors"
	"testing"
)

func TestMutateSavesDraft(t *testing.T) {
	setupDraftTestEnv(t)
	cache := saveParametersCache(t, "demo", "etag-1", remoteConfigRaw("1", map[string]string{"flag": "old"}))
	deps := (&fakeDeps{cache: cache}).deps()

	result, hasDraft, err := Mutate(context.Background(), deps, "demo", false, MutationSpec{
		Apply: SetStringParameterValue("", "flag", "default", "new"),
	})
	if err != nil {
		t.Fatalf("Mutate returned error: %v", err)
	}
	if !hasDraft || !result.HasDraft {
		t.Fatal("hasDraft = false, want true")
	}
	assertParamValue(t, result.FinalRaw, "flag", "new")

	raw, loaded := loadDraft(t, "demo")
	if !loaded {
		t.Fatal("draft not saved")
	}
	assertParamValue(t, raw, "flag", "new")
}

func TestPreviewUsesDraftWhenPresent(t *testing.T) {
	setupDraftTestEnv(t)
	cache := saveParametersCache(t, "demo", "etag-1", remoteConfigRaw("1", map[string]string{"flag": "cached"}))
	if err := Save("demo", remoteConfigRaw("2", map[string]string{"flag": "draft"})); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	deps := (&fakeDeps{cache: cache}).deps()
	_, finalRaw, err := Preview(deps, "demo", MutationSpec{
		Apply: RenameParameter("", "flag", "renamed"),
	})
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	assertFlagMissing(t, finalRaw)
	assertParamValue(t, finalRaw, "renamed", "draft")
}

func TestPublishExistingDraftFailureKeepsDraft(t *testing.T) {
	setupDraftTestEnv(t)
	cache := saveParametersCache(t, "demo", "etag-1", remoteConfigRaw("1", map[string]string{"flag": "cached"}))
	draftRaw := remoteConfigRaw("2", map[string]string{"flag": "draft"})
	if err := Save("demo", draftRaw); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	fake := &fakeDeps{cache: cache, publishErr: errors.New("publish failed")}
	_, _, err := PublishExistingDraft(context.Background(), fake.deps(), "demo")
	if err == nil {
		t.Fatal("PublishExistingDraft returned nil error")
	}

	raw, hasDraft := loadDraft(t, "demo")
	if !hasDraft {
		t.Fatal("draft was removed after failed publish")
	}
	assertParamValue(t, raw, "flag", "draft")
}

func TestPublishExistingDraftSuccessRemovesDraft(t *testing.T) {
	setupDraftTestEnv(t)
	cache := saveParametersCache(t, "demo", "etag-1", remoteConfigRaw("1", map[string]string{"flag": "cached"}))
	if err := Save("demo", remoteConfigRaw("2", map[string]string{"flag": "draft"})); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	fake := &fakeDeps{cache: cache}
	updatedCache, updatedRaw, err := PublishExistingDraft(context.Background(), fake.deps(), "demo")
	if err != nil {
		t.Fatalf("PublishExistingDraft returned error: %v", err)
	}
	if fake.publishCalls != 1 {
		t.Fatalf("publishCalls = %d, want 1", fake.publishCalls)
	}
	if updatedCache.ETag != "etag-2" {
		t.Fatalf("etag = %q, want etag-2", updatedCache.ETag)
	}
	assertParamValue(t, updatedRaw, "flag", "draft")

	if _, hasDraft := loadDraft(t, "demo"); hasDraft {
		t.Fatal("draft still present after successful publish")
	}
}

func TestPublishExistingDraftRebasesNonConflictingRemoteChanges(t *testing.T) {
	setupDraftTestEnv(t)
	base := saveParametersCache(t, "demo", "etag-1", remoteConfigRaw("1", map[string]string{"local": "old", "remote": "old"}))
	if err := SaveWithBase("demo", base, remoteConfigRaw("1", map[string]string{"local": "draft", "remote": "old"})); err != nil {
		t.Fatalf("SaveWithBase returned error: %v", err)
	}
	latest := saveParametersCache(t, "demo", "etag-2", remoteConfigRaw("2", map[string]string{"local": "old", "remote": "firebase"}))
	fake := &fakeDeps{cache: latest}
	_, publishedRaw, err := PublishExistingDraft(context.Background(), fake.deps(), "demo")
	if err != nil {
		t.Fatalf("PublishExistingDraft returned error: %v", err)
	}
	assertParamValue(t, publishedRaw, "local", "draft")
	assertParamValue(t, publishedRaw, "remote", "firebase")
}

func TestPublishExistingDraftConflictKeepsDraft(t *testing.T) {
	setupDraftTestEnv(t)
	base := saveParametersCache(t, "demo", "etag-1", remoteConfigRaw("1", map[string]string{"flag": "old"}))
	if err := SaveWithBase("demo", base, remoteConfigRaw("1", map[string]string{"flag": "draft"})); err != nil {
		t.Fatalf("SaveWithBase returned error: %v", err)
	}
	latest := saveParametersCache(t, "demo", "etag-2", remoteConfigRaw("2", map[string]string{"flag": "firebase"}))
	fake := &fakeDeps{cache: latest}
	if _, _, err := PublishExistingDraft(context.Background(), fake.deps(), "demo"); err == nil {
		t.Fatal("PublishExistingDraft returned nil conflict error")
	}
	if fake.publishCalls != 0 {
		t.Fatalf("publishCalls = %d, want 0", fake.publishCalls)
	}
	if _, hasDraft := loadDraft(t, "demo"); !hasDraft {
		t.Fatal("draft removed after conflict")
	}
}
