package core

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/config"
)

func TestDeleteProjectIDsRemovesOnlySelectedLocalData(t *testing.T) {
	svc := setupCoreTestEnv(t)
	projects := []config.Project{
		{Name: "Alpha", ProjectID: "alpha", AuthID: "main"},
		{Name: "Beta", ProjectID: "beta", AuthID: "main"},
	}
	if err := config.SaveProjects(projects, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	for _, version := range []string{"1", "2"} {
		cache := &config.ParametersCache{
			ETag:         "etag-" + version,
			CachedAt:     now,
			RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"` + version + `"}}`),
		}
		var err error
		if version == "1" {
			err = config.SaveParametersCache("alpha", cache)
		} else {
			err = config.SaveParametersCacheSnapshot("alpha", cache)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := config.SaveParametersCache("beta", &config.ParametersCache{
		ETag:         "etag-9",
		CachedAt:     now,
		RemoteConfig: json.RawMessage(`{"version":{"versionNumber":"9"}}`),
	}); err != nil {
		t.Fatal(err)
	}
	if err := config.SaveDraft(&config.Draft{
		FormatVersion:    config.DraftFormatVersion,
		ProjectID:        "alpha",
		BaseVersion:      "1",
		BaseETag:         "etag-1",
		CreatedAt:        now,
		UpdatedAt:        now,
		BaseRemoteConfig: json.RawMessage(`{}`),
		RemoteConfig:     json.RawMessage(`{}`),
	}); err != nil {
		t.Fatal(err)
	}

	deleted, err := svc.DeleteProjectIDs([]string{"alpha"})
	if err != nil {
		t.Fatalf("DeleteProjectIDs = %v", err)
	}
	if len(deleted) != 1 || deleted[0].ProjectID != "alpha" {
		t.Fatalf("deleted = %+v, want alpha", deleted)
	}
	remaining, err := config.LoadProjects()
	if err != nil || len(remaining) != 1 || remaining[0].ProjectID != "beta" {
		t.Fatalf("remaining = %+v, err=%v; want beta", remaining, err)
	}
	for _, path := range []string{
		config.GetParametersCachePath("alpha"),
		config.GetParametersCacheVersionPath("alpha", "1"),
		config.GetParametersCacheVersionPath("alpha", "2"),
		config.GetDraftPath("alpha"),
	} {
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("deleted path %s still exists or returned err=%v", path, err)
		}
	}
	if _, err := os.Lstat(config.GetParametersCachePath("beta")); err != nil {
		t.Fatalf("beta cache was removed: %v", err)
	}
}

func TestDeleteProjectIDsRejectsUnknownProjects(t *testing.T) {
	svc := setupCoreTestEnv(t)
	if err := config.SaveProjects([]config.Project{{Name: "Alpha", ProjectID: "alpha", AuthID: "main"}}, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DeleteProjectIDs([]string{"missing"}); err == nil {
		t.Fatal("DeleteProjectIDs unknown project = nil, want error")
	}
}
