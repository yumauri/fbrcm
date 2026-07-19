package vhs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rc/importer"
)

func TestRemoteConfigFixturesAreImportable(t *testing.T) {
	paths, err := filepath.Glob("*.json")
	if err != nil {
		t.Fatalf("list fixtures: %v", err)
	}
	if len(paths) != 5 {
		t.Fatalf("fixture count = %d, want 5", len(paths))
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			source, err := importer.ParseSource(raw)
			if err != nil {
				t.Fatalf("ParseSource = %v", err)
			}
			if len(source.Config.Conditions) != 5 {
				t.Fatalf("conditions = %d, want 5", len(source.Config.Conditions))
			}
			for _, condition := range source.Config.Conditions {
				if _, err := firebase.NormalizeConditionTagColor(condition.TagColor); err != nil {
					t.Fatalf("condition %q: %v", condition.Name, err)
				}
			}
			portable, err := firebase.CloneRemoteConfig(source.Config)
			if err != nil {
				t.Fatalf("CloneRemoteConfig = %v", err)
			}
			if err := importer.Transform("fixture", "Fixture", portable, importer.Options{ConditionPolicy: importer.ConditionPolicyKeepPortableOnly}); err != nil {
				t.Fatalf("portable condition transform = %v", err)
			}
			if len(portable.Conditions) != len(source.Config.Conditions) {
				t.Fatalf("non-portable conditions removed: got %d, want %d", len(portable.Conditions), len(source.Config.Conditions))
			}
			if _, err := firebase.MarshalRemoteConfigForUpdate(source.Config); err != nil {
				t.Fatalf("MarshalRemoteConfigForUpdate = %v", err)
			}
		})
	}
}
