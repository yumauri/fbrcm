package harness

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type SnapshotChange struct {
	Path    string
	Created bool
	Updated bool
}

type SnapshotReplacement struct {
	Old string
	New string
}

func CanonicalizeSnapshot(raw []byte, replacements ...SnapshotReplacement) []byte {
	ordered := append([]SnapshotReplacement(nil), replacements...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return len(ordered[i].Old) > len(ordered[j].Old)
	})
	canonical := append([]byte(nil), raw...)
	for _, replacement := range ordered {
		if replacement.Old == "" || replacement.Old == replacement.New {
			continue
		}
		canonical = bytes.ReplaceAll(canonical, []byte(replacement.Old), []byte(replacement.New))
	}
	return canonical
}

func CheckSnapshot(path string, actual []byte, update bool) (SnapshotChange, error) {
	expected, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return SnapshotChange{}, fmt.Errorf("read snapshot %s: %w", path, err)
		}
		if !update {
			return SnapshotChange{}, fmt.Errorf("snapshot %s is missing; use -mode=record-missing or -mode=update-output", path)
		}
		if err := atomicWrite(path, actual, 0o644); err != nil {
			return SnapshotChange{}, err
		}
		return SnapshotChange{Path: path, Created: true}, nil
	}
	if bytes.Equal(expected, actual) {
		return SnapshotChange{Path: path}, nil
	}
	if update {
		if err := atomicWrite(path, actual, 0o644); err != nil {
			return SnapshotChange{}, err
		}
		return SnapshotChange{Path: path, Updated: true}, nil
	}
	return SnapshotChange{}, fmt.Errorf("snapshot mismatch for %s\n%s", path, snapshotDiff(expected, actual))
}

func atomicWrite(path string, raw []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create snapshot directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".e2e-snapshot-*")
	if err != nil {
		return fmt.Errorf("create temporary snapshot: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set snapshot permissions: %w", err)
	}
	if _, err := temporary.Write(raw); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary snapshot: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary snapshot: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace snapshot %s: %w", path, err)
	}
	return nil
}

func snapshotDiff(expected, actual []byte) string {
	index := 0
	limit := min(len(expected), len(actual))
	for index < limit && expected[index] == actual[index] {
		index++
	}
	start := max(index-80, 0)
	expectedEnd := min(index+160, len(expected))
	actualEnd := min(index+160, len(actual))
	return fmt.Sprintf(
		"first differing byte: %d\nexpected (%d bytes): %q\nactual   (%d bytes): %q",
		index,
		len(expected), expected[start:expectedEnd],
		len(actual), actual[start:actualEnd],
	)
}
