package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirebaseRCDiscoveryAndCommentedJSON(t *testing.T) {
	setupTestDirs(t)
	root := t.TempDir()
	nested := filepath.Join(root, "packages", "app")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, FirebaseConfigFileName), "{}\n", 0o644)
	writeFile(t, filepath.Join(root, FirebaseRCFileName), `{
  // Shared Firebase aliases.
  "projects": {
    "default": "acme-development-42",
    "prod": "acme-production-42",
  },
  "targets": {"acme-production-42": {}}
}
`, 0o644)
	withWorkingDirectory(t, nested)

	path, exists, err := GetFirebaseRCFilePath()
	if err != nil {
		t.Fatal(err)
	}
	wantPath, err := filepath.EvalSymlinks(filepath.Join(root, FirebaseRCFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !exists || path != wantPath {
		t.Fatalf("Firebase RC path = %q, %t", path, exists)
	}
	aliases, err := LoadFirebaseProjectAliasesFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if aliases["default"] != "acme-development-42" || aliases["prod"] != "acme-production-42" || len(aliases) != 2 {
		t.Fatalf("Firebase aliases = %#v", aliases)
	}
}

func TestFirebaseRCFallsBackToCurrentDirectoryWithoutFirebaseConfig(t *testing.T) {
	setupTestDirs(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(cwd, FirebaseRCFileName), `{"projects":{"prod":"acme-production-42"}}`, 0o644)

	path, exists, err := GetFirebaseRCFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if !exists || path != filepath.Join(cwd, FirebaseRCFileName) {
		t.Fatalf("Firebase RC fallback = %q, %t", path, exists)
	}
}

func TestLoadProjectAliasRegistryCombinesSources(t *testing.T) {
	setupTestDirs(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(cwd, LocalConfigFileName), `[projects.aliases]
native = "acme-native-42"
prod = "acme-production-42"
`, 0o644)
	writeFile(t, filepath.Join(cwd, FirebaseConfigFileName), "{}\n", 0o644)
	writeFile(t, filepath.Join(cwd, FirebaseRCFileName), `{"projects":{"default":"acme-development-42","prod":"acme-production-42"}}`, 0o644)

	registry, err := LoadProjectAliasRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if len(registry.Aliases) != 3 || registry.Entries["native"].Source != ProjectAliasSourceFBRCM || registry.Entries["default"].Source != ProjectAliasSourceFirebase || registry.Entries["prod"].Source != ProjectAliasSourceBoth {
		t.Fatalf("alias registry = %#v", registry)
	}
	entries := registry.SortedEntries()
	if len(entries) != 3 || entries[0].Alias != "default" || entries[1].Alias != "native" || entries[2].Alias != "prod" {
		t.Fatalf("sorted entries = %#v", entries)
	}
}

func TestLoadProjectAliasRegistryRejectsConflicts(t *testing.T) {
	setupTestDirs(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	localPath := filepath.Join(cwd, LocalConfigFileName)
	firebasePath := filepath.Join(cwd, FirebaseRCFileName)
	writeFile(t, localPath, "[projects.aliases]\nprod = \"native-production-42\"\n", 0o644)
	writeFile(t, filepath.Join(cwd, FirebaseConfigFileName), "{}\n", 0o644)
	writeFile(t, firebasePath, `{"projects":{"prod":"firebase-production-42"}}`, 0o644)

	_, err = LoadProjectAliasRegistry()
	if err == nil || !strings.Contains(err.Error(), "conflicts") || !strings.Contains(err.Error(), localPath) || !strings.Contains(err.Error(), firebasePath) {
		t.Fatalf("registry conflict = %v", err)
	}
}

func TestLoadFirebaseProjectAliasesFileValidatesSchemaAndAliases(t *testing.T) {
	for name, contents := range map[string]string{
		"wrong projects type":   `{"projects":[]}`,
		"wrong project id type": `{"projects":{"prod":42}}`,
		"unsupported alias":     `{"projects":{"Prod":"acme-production-42"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), FirebaseRCFileName)
			writeFile(t, path, contents, 0o644)
			if _, err := LoadFirebaseProjectAliasesFile(path); err == nil || !strings.Contains(err.Error(), path) {
				t.Fatalf("LoadFirebaseProjectAliasesFile error = %v", err)
			}
		})
	}
}
