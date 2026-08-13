package groups

import (
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
)

func TestGroupMetadataLifecyclePreservesEmptyGroups(t *testing.T) {
	cfg := &firebase.RemoteConfig{}
	if err := EditDetails(cfg, DetailsEdit{Create: true, NextName: " empty ", NextDescription: " Metadata only "}); err != nil {
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

func TestResolveNameRequiresExactCaseAndWhitespace(t *testing.T) {
	cfg := &firebase.RemoteConfig{ParameterGroups: map[string]firebase.RemoteConfigGroup{"Canonical": {}}}
	if got, ok := ResolveName(cfg, "Canonical"); !ok || got != "Canonical" {
		t.Fatalf("exact ResolveName = %q, %t", got, ok)
	}
	for _, query := range []string{"canonical", " Canonical"} {
		if got, ok := ResolveName(cfg, query); ok || got != "" {
			t.Fatalf("ResolveName(%q) = %q, %t; want not found", query, got, ok)
		}
	}
}
