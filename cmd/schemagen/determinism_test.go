package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestStageGeneratedContractIsByteDeterministic(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	first := filepath.Join(t.TempDir(), "first")
	second := filepath.Join(t.TempDir(), "second")
	stageGeneratedContract(first)
	stageGeneratedContract(second)

	firstFiles := generatedFiles(t, first)
	secondFiles := generatedFiles(t, second)
	if !slices.Equal(firstFiles, secondFiles) {
		t.Fatalf("generated paths differ:\nfirst:  %v\nsecond: %v", firstFiles, secondFiles)
	}
	for _, name := range firstFiles {
		firstRaw, err := os.ReadFile(filepath.Join(first, name))
		if err != nil {
			t.Fatal(err)
		}
		secondRaw, err := os.ReadFile(filepath.Join(second, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(firstRaw, secondRaw) {
			t.Errorf("generated file %s differs between runs", name)
		}
	}
}

func generatedFiles(t *testing.T, root string) []string {
	t.Helper()
	var result []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		result = append(result, filepath.ToSlash(name))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(result)
	return result
}
