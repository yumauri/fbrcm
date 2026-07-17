package parameters

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/rootgroup"
)

func TestBuildTreeNilRemoteConfig(t *testing.T) {
	cachedAt := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	tree := BuildTree(nil, cachedAt, "etag-1")

	if tree.Version != "" {
		t.Fatalf("Version = %q, want empty", tree.Version)
	}
	if !tree.CachedAt.Equal(cachedAt) {
		t.Fatalf("CachedAt = %v, want %v", tree.CachedAt, cachedAt)
	}
	if tree.ETag != "etag-1" {
		t.Fatalf("ETag = %q, want etag-1", tree.ETag)
	}
	if len(tree.Groups) != 0 {
		t.Fatalf("Groups = %d, want 0", len(tree.Groups))
	}
}

func TestBuildTreeFromFixtures(t *testing.T) {
	cachedAt := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		fixture        string
		wantVersion    string
		wantGroupKeys  []string
		wantRootParams []string
		wantGrouped    map[string][]string
	}{
		{
			name:           "root only params",
			fixture:        "root_params.json",
			wantVersion:    "1",
			wantGroupKeys:  []string{rootgroup.TreeKey},
			wantRootParams: []string{"feature_login", "max_retries"},
		},
		{
			name:          "mixed root and grouped",
			fixture:       "grouped_params.json",
			wantVersion:   "2",
			wantGroupKeys: []string{rootgroup.TreeKey, "checkout"},
			wantRootParams: []string{
				"app_theme",
			},
			wantGrouped: map[string][]string{
				"checkout": {"enable_coupons", "tax_rate"},
			},
		},
		{
			name:           "conditions preserve order and colors",
			fixture:        "with_conditions.json",
			wantVersion:    "3",
			wantGroupKeys:  []string{rootgroup.TreeKey},
			wantRootParams: []string{"feature_login"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadFixtureRemoteConfig(t, tt.fixture)
			tree := BuildTree(cfg, cachedAt, "fixture-etag")

			if tree.Version != tt.wantVersion {
				t.Fatalf("Version = %q, want %q", tree.Version, tt.wantVersion)
			}
			if tree.ETag != "fixture-etag" {
				t.Fatalf("ETag = %q, want fixture-etag", tree.ETag)
			}
			if !tree.CachedAt.Equal(cachedAt) {
				t.Fatalf("CachedAt = %v, want %v", tree.CachedAt, cachedAt)
			}

			gotKeys := groupKeys(tree.Groups)
			if len(gotKeys) != len(tt.wantGroupKeys) {
				t.Fatalf("group keys = %v, want %v", gotKeys, tt.wantGroupKeys)
			}
			for i, want := range tt.wantGroupKeys {
				if gotKeys[i] != want {
					t.Fatalf("group[%d] key = %q, want %q", i, gotKeys[i], want)
				}
			}

			root := tree.Groups[0]
			if root.Key != rootgroup.TreeKey {
				t.Fatalf("first group key = %q, want %q", root.Key, rootgroup.TreeKey)
			}
			if root.Label != rootgroup.Label {
				t.Fatalf("first group label = %q, want %q", root.Label, rootgroup.Label)
			}
			assertParamKeys(t, root.Parameters, tt.wantRootParams)

			for groupKey, wantParams := range tt.wantGrouped {
				group, ok := findGroup(tree.Groups, groupKey)
				if !ok {
					t.Fatalf("group %q not found in %v", groupKey, gotKeys)
				}
				if group.Label != groupKey {
					t.Fatalf("group %q label = %q, want same as key", groupKey, group.Label)
				}
				assertParamKeys(t, group.Parameters, wantParams)
			}

			if tt.fixture == "with_conditions.json" {
				if len(tree.Conditions) != 2 || tree.Conditions[0].Name != "android" || tree.Conditions[0].Color != "BLUE" || tree.Conditions[1].Name != "ios" {
					t.Fatalf("conditions = %+v, want android/BLUE then ios", tree.Conditions)
				}
				entry := root.Parameters[0]
				if entry.Key != "feature_login" {
					t.Fatalf("param key = %q, want feature_login", entry.Key)
				}
				if len(entry.Values) != 3 {
					t.Fatalf("conditional values = %d, want 3 (android, ios, default)", len(entry.Values))
				}
				if entry.Values[0].Label != "android" || entry.Values[0].Color != "BLUE" {
					t.Fatalf("first conditional = %+v, want android/BLUE", entry.Values[0])
				}
				if entry.Values[1].Label != "ios" || entry.Values[1].Color != "INDIGO" {
					t.Fatalf("second conditional = %+v, want ios/INDIGO", entry.Values[1])
				}
				if entry.Values[2].Label != "default" || entry.Values[2].Value != "off" {
					t.Fatalf("default value = %+v, want default/off", entry.Values[2])
				}
			}
		})
	}
}

func TestBuildTreeGroupedOnlyNoRootBucket(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Version: firebase.RemoteConfigVersion{VersionNumber: "9"},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"alpha": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"flag_a": {DefaultValue: &firebase.RemoteConfigValue{Value: "1"}},
				},
			},
		},
	}
	tree := BuildTree(cfg, time.Now(), "etag")

	if len(tree.Groups) != 1 {
		t.Fatalf("groups = %d, want 1 grouped-only bucket", len(tree.Groups))
	}
	if tree.Groups[0].Key != "alpha" {
		t.Fatalf("group key = %q, want alpha", tree.Groups[0].Key)
	}
}

func TestBuildTreeIncludesEmptyAndDescriptionOnlyGroups(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"empty": {},
			"ROKU":  {Description: "FLAGS FOR ROKU"},
		},
	}

	tree := BuildTree(cfg, time.Now(), "etag")
	if got, want := groupKeys(tree.Groups), []string{"empty", "ROKU"}; !slices.Equal(got, want) {
		t.Fatalf("group keys = %v, want %v", got, want)
	}
	for _, group := range tree.Groups {
		if len(group.Parameters) != 0 {
			t.Fatalf("group %q parameters = %#v, want empty", group.Key, group.Parameters)
		}
	}
	if group, ok := findGroup(tree.Groups, "ROKU"); !ok || group.Description != "FLAGS FOR ROKU" {
		t.Fatalf("ROKU group metadata = %#v, want description", group)
	}
}

func TestBuildTreeSkipsDuplicateRootParamsInGroups(t *testing.T) {
	cfg := &firebase.RemoteConfig{
		Version: firebase.RemoteConfigVersion{VersionNumber: "10"},
		Parameters: map[string]firebase.RemoteConfigParam{
			"shared_flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "root"}},
			"root_only":   {DefaultValue: &firebase.RemoteConfigValue{Value: "yes"}},
		},
		ParameterGroups: map[string]firebase.RemoteConfigGroup{
			"experiments": {
				Parameters: map[string]firebase.RemoteConfigParam{
					"shared_flag": {DefaultValue: &firebase.RemoteConfigValue{Value: "group"}},
				},
			},
		},
	}
	tree := BuildTree(cfg, time.Now(), "etag")

	root, ok := findGroup(tree.Groups, rootgroup.TreeKey)
	if !ok {
		t.Fatal("root group not found")
	}
	assertParamKeys(t, root.Parameters, []string{"root_only"})

	experiments, ok := findGroup(tree.Groups, "experiments")
	if !ok {
		t.Fatal("experiments group not found")
	}
	assertParamKeys(t, experiments.Parameters, []string{"shared_flag"})
}

func loadFixtureRemoteConfig(t *testing.T, name string) *firebase.RemoteConfig {
	t.Helper()
	dir := remoteConfigFixtureDir(t)
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	cfg, err := firebase.ParseRemoteConfig(raw)
	if err != nil {
		t.Fatalf("ParseRemoteConfig(%s): %v", name, err)
	}
	return cfg
}

func remoteConfigFixtureDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "remoteconfig")
}

func groupKeys(groups []Group) []string {
	keys := make([]string, len(groups))
	for i, group := range groups {
		keys[i] = group.Key
	}
	return keys
}

func findGroup(groups []Group, key string) (Group, bool) {
	for _, group := range groups {
		if group.Key == key {
			return group, true
		}
	}
	return Group{}, false
}

func assertParamKeys(t *testing.T, entries []Entry, want []string) {
	t.Helper()
	if len(entries) != len(want) {
		t.Fatalf("param count = %d, want %d; got keys %v", len(entries), len(want), entryKeys(entries))
	}
	for i, key := range want {
		if entries[i].Key != key {
			t.Fatalf("param[%d] = %q, want %q; all keys %v", i, entries[i].Key, key, entryKeys(entries))
		}
	}
}

func entryKeys(entries []Entry) []string {
	keys := make([]string, len(entries))
	for i, entry := range entries {
		keys[i] = entry.Key
	}
	return keys
}
