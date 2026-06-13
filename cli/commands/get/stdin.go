package get

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/shared"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/filter"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
)

func writeRowsJSON(cmd *cobra.Command, rows []parameterRow) error {
	out := make([]parameterRowJSON, 0, len(rows))
	for _, row := range rows {
		out = append(out, parameterRowJSON{
			Project:      row.Project,
			ProjectID:    row.ProjectID,
			Group:        row.Group,
			Key:          row.Key,
			Description:  row.Description,
			DefaultValue: row.DefaultValue,
			Conditional:  row.Conditional,
			Conditions:   row.Conditions,
			Type:         row.Type,
			Version:      stringPtrOrNil(row.Version),
			CachedAt:     timePtrOrNil(row.CachedAt),
			Status:       stringPtrOrNil(row.Status),
		})
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(out); err != nil {
		return fmt.Errorf("encode parameters json: %w", err)
	}
	return nil
}

func loadStdinParameterRows(cmd *cobra.Command, compiledExpr *filter.Expression, search shared.ParameterSearch) (*firebase.RemoteConfig, []parameterRow, error) {
	raw, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return nil, nil, fmt.Errorf("read stdin: %w", err)
	}
	corelog.For("get").Info("loaded remote config from stdin", "bytes", len(raw))
	if !json.Valid(raw) {
		return nil, nil, fmt.Errorf("stdin remote config is not valid json")
	}

	remoteConfigRaw, err := shared.ExtractRemoteConfigJSON(raw)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := firebase.ParseRemoteConfig(remoteConfigRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("decode stdin remote config: %w", err)
	}

	version := stdinVersion(remoteConfigRaw)
	project := core.Project{
		Name:      "<stdin>",
		ProjectID: "<stdin>",
	}
	rows := flattenParameters(project, cfg, time.Time{}, "", version, compiledExpr, search)
	corelog.For("get").Info("parsed stdin remote config", "parameters", len(rows), "version", version)
	return cfg, rows, nil
}

// loadStdinDirectoryParameterRows loads JSON remote configs from a directory passed as stdin.
func loadStdinDirectoryParameterRows(cmd *cobra.Command, projectExpr string, search shared.ParameterSearch) (bool, []parameterRow, error) {
	dir, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return false, nil, nil
	}
	info, err := dir.Stat()
	if err != nil {
		return false, nil, nil
	}
	if !info.IsDir() {
		return false, nil, nil
	}

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return true, nil, fmt.Errorf("read stdin directory: %w", err)
	}
	names = filterJSONFileNames(names)
	sort.Strings(names)

	compiledExpr, ok := shared.CompileExpr(projectExpr, "<stdin-dir>")
	if !ok {
		return true, nil, nil
	}

	rows := make([]parameterRow, 0)
	for _, name := range names {
		raw, err := readStdinDirectoryFile(dir, name)
		if err != nil {
			return true, nil, fmt.Errorf("read stdin directory file %q: %w", name, err)
		}
		if !json.Valid(raw) {
			return true, nil, fmt.Errorf("stdin directory file %q is not valid json", name)
		}

		remoteConfigRaw, err := shared.ExtractRemoteConfigJSON(raw)
		if err != nil {
			return true, nil, fmt.Errorf("extract remote config from stdin directory file %q: %w", name, err)
		}

		cfg, err := firebase.ParseRemoteConfig(remoteConfigRaw)
		if err != nil {
			return true, nil, fmt.Errorf("decode stdin directory file %q: %w", name, err)
		}

		projectID := strings.TrimSuffix(name, filepath.Ext(name))
		project := core.Project{
			Name:      stdinDirectoryProjectName(projectID),
			ProjectID: projectID,
		}
		version := stdinVersion(remoteConfigRaw)
		rows = append(rows, flattenParameters(project, cfg, time.Time{}, "", version, compiledExpr, search)...)
	}
	corelog.For("get").Info("parsed remote configs from stdin directory", "files", len(names), "parameters", len(rows))
	return true, rows, nil
}

// filterJSONFileNames returns only top-level .json file names.
func filterJSONFileNames(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if strings.EqualFold(filepath.Ext(name), ".json") {
			out = append(out, name)
		}
	}
	return out
}

// stdinDirectoryProjectName builds a display name from a file stem.
func stdinDirectoryProjectName(stem string) string {
	parts := strings.FieldsFunc(stem, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, part := range parts {
		parts[i] = capitalizeProjectNamePart(part)
	}
	return strings.Join(parts, " ")
}

// capitalizeProjectNamePart capitalizes a project name segment.
func capitalizeProjectNamePart(part string) string {
	if part == "" {
		return ""
	}
	runes := []rune(strings.ToLower(part))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func stdinVersion(raw []byte) string {
	var payload struct {
		Version *struct {
			VersionNumber string `json:"versionNumber"`
		} `json:"version"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if payload.Version == nil {
		return ""
	}
	return strings.TrimSpace(payload.Version.VersionNumber)
}
