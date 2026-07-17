package groups

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestGroupMetadataLifecyclePreservesEmptyGroups(t *testing.T) {
	cfg := &firebase.RemoteConfig{}
	if err := Add(cfg, Definition{Name: " empty ", Description: " Metadata only "}); err != nil {
		t.Fatal(err)
	}
	group, ok := cfg.ParameterGroups["empty"]
	if !ok || group.Description != "Metadata only" || group.Parameters != nil {
		t.Fatalf("added group = %#v", cfg.ParameterGroups)
	}
	if err := EditDetails(cfg, DetailsEdit{Name: "empty", NextName: "renamed", NextDescription: "Updated"}); err != nil {
		t.Fatal(err)
	}
	group, ok = cfg.ParameterGroups["renamed"]
	if !ok || group.Description != "Updated" || group.Parameters != nil {
		t.Fatalf("edited group = %#v", cfg.ParameterGroups)
	}
	if err := Delete(cfg, "renamed"); err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.ParameterGroups["renamed"]; ok {
		t.Fatal("deleted group remains")
	}
}

func TestEditMetadataCanIntentionallyClearDescription(t *testing.T) {
	cfg := &firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{"empty": {Description: "old"}}}
	next := ""
	if err := EditMetadata(cfg, "empty", Edit{Description: &next}); err != nil {
		t.Fatal(err)
	}
	if group := cfg.ParameterGroups["empty"]; group.Description != "" || group.Parameters != nil {
		t.Fatalf("cleared group = %#v", group)
	}
}
