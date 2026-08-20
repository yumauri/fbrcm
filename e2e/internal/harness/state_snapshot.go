package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func checkExpectedStateFileSnapshots(
	scenario Scenario,
	environment Environment,
	mode Mode,
	runtimeReplacements []SnapshotReplacement,
) ([]SnapshotChange, error) {
	variables := environmentMap(environment.Variables)
	roots := map[string]string{
		"config": variables["FBRCM_CONFIG_DIR"],
		"cache":  variables["FBRCM_CACHE_DIR"],
	}
	changes := make([]SnapshotChange, 0, len(scenario.ExpectedStateFiles))
	for _, expectation := range scenario.ExpectedStateFiles {
		actualPath := filepath.Join(roots[expectation.Root], expectation.Path)
		info, err := os.Lstat(actualPath)
		if err != nil {
			return nil, fmt.Errorf("inspect expected %s state file %s: %w", expectation.Root, expectation.Path, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("expected %s state file %s is not a regular file", expectation.Root, expectation.Path)
		}
		raw, err := os.ReadFile(actualPath)
		if err != nil {
			return nil, fmt.Errorf("read expected %s state file %s: %w", expectation.Root, expectation.Path, err)
		}
		jsonReplacements, err := jsonPointerSnapshotReplacements(raw, expectation.JSONReplacements)
		if err != nil {
			return nil, fmt.Errorf("canonicalize expected %s state file %s: %w", expectation.Root, expectation.Path, err)
		}
		replacements := append(append([]SnapshotReplacement(nil), runtimeReplacements...), jsonReplacements...)
		canonical := CanonicalizeSnapshot(raw, replacements...)
		snapshotPath := filepath.Join(scenario.Directory, "state", expectation.Root, expectation.Path+".golden")
		_, snapshotErr := os.Stat(snapshotPath)
		exists := snapshotErr == nil
		if snapshotErr != nil && !os.IsNotExist(snapshotErr) {
			return nil, fmt.Errorf("stat state snapshot %s: %w", snapshotPath, snapshotErr)
		}
		change, err := CheckSnapshot(snapshotPath, canonical, mode.UpdateOutput(exists))
		if err != nil {
			return nil, err
		}
		if change.Created || change.Updated {
			changes = append(changes, change)
		}
	}
	return changes, nil
}

func checkExpectedAbsentStatePaths(scenario Scenario, environment Environment) error {
	variables := environmentMap(environment.Variables)
	roots := map[string]string{
		"config": variables["FBRCM_CONFIG_DIR"],
		"cache":  variables["FBRCM_CACHE_DIR"],
	}
	for _, expectation := range scenario.ExpectedAbsentStatePaths {
		path := filepath.Join(roots[expectation.Root], expectation.Path)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect absent %s state path %s: %w", expectation.Root, expectation.Path, err)
		}
		kind := "file"
		if info.IsDir() {
			kind = "directory"
		}
		return fmt.Errorf("expected %s state path %s to be absent, but a %s still exists", expectation.Root, expectation.Path, kind)
	}
	return nil
}

func expectedStateFileSnapshotReplacements(scenario Scenario, environment Environment) ([]SnapshotReplacement, error) {
	variables := environmentMap(environment.Variables)
	roots := map[string]string{
		"config": variables["FBRCM_CONFIG_DIR"],
		"cache":  variables["FBRCM_CACHE_DIR"],
	}
	var replacements []SnapshotReplacement
	for _, expectation := range scenario.ExpectedStateFiles {
		if len(expectation.JSONReplacements) == 0 {
			continue
		}
		path := filepath.Join(roots[expectation.Root], expectation.Path)
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read expected %s state file %s for canonicalization: %w", expectation.Root, expectation.Path, err)
		}
		fileReplacements, err := jsonPointerSnapshotReplacements(raw, expectation.JSONReplacements)
		if err != nil {
			return nil, fmt.Errorf("canonicalize expected %s state file %s: %w", expectation.Root, expectation.Path, err)
		}
		replacements = append(replacements, fileReplacements...)
	}
	return replacements, nil
}

func jsonPointerSnapshotReplacements(raw []byte, pointers map[string]string) ([]SnapshotReplacement, error) {
	if len(pointers) == 0 {
		return nil, nil
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, fmt.Errorf("decode JSON for declared replacements: %w", err)
	}
	ordered := make([]string, 0, len(pointers))
	for pointer := range pointers {
		ordered = append(ordered, pointer)
	}
	sort.Strings(ordered)
	replacements := make([]SnapshotReplacement, 0, len(ordered))
	for _, pointer := range ordered {
		value, err := resolveJSONPointer(document, pointer)
		if err != nil {
			return nil, err
		}
		if stringValue, ok := value.(string); ok {
			replacements = append(replacements, SnapshotReplacement{Old: stringValue, New: pointers[pointer]})
			continue
		}
		encodedValue, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("encode JSON pointer %q value: %w", pointer, err)
		}
		encodedPlaceholder := strconv.Quote(pointers[pointer])
		replacements = append(replacements, SnapshotReplacement{Old: string(encodedValue), New: encodedPlaceholder})
	}
	return replacements, nil
}

func resolveJSONPointer(document any, pointer string) (any, error) {
	if pointer == "" {
		return document, nil
	}
	current := document
	for encodedToken := range strings.SplitSeq(strings.TrimPrefix(pointer, "/"), "/") {
		token := strings.ReplaceAll(strings.ReplaceAll(encodedToken, "~1", "/"), "~0", "~")
		switch value := current.(type) {
		case map[string]any:
			next, ok := value[token]
			if !ok {
				return nil, fmt.Errorf("JSON pointer %q does not exist", pointer)
			}
			current = next
		case []any:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(value) {
				return nil, fmt.Errorf("JSON pointer %q has invalid array index %q", pointer, token)
			}
			current = value[index]
		default:
			return nil, fmt.Errorf("JSON pointer %q traverses a scalar at %q", pointer, token)
		}
	}
	return current, nil
}
