package theme

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	pathpkg "path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yumauri/fbrcm/cli/contract"
	"github.com/yumauri/fbrcm/cli/progress"
	"github.com/yumauri/fbrcm/cli/shared"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
)

const maximumThemeImportBytes = 1024 * 1024

func newImportCommand(client *http.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [source]",
		Short: "Import themes from a file, directory, or HTTP URL",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := ""
			if len(args) == 1 {
				source = args[0]
			}
			nameOverride, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			if nameOverride != "" {
				if err := coreconfig.ValidateThemeName(nameOverride); err != nil {
					return shared.InvalidArgument(err)
				}
			}
			if source != "" {
				if handled, err := importThemeDirectoryPath(cmd, source, nameOverride); handled || err != nil {
					return err
				}
			}
			if source == "" {
				if handled, err := importThemeDirectoryStdin(cmd, nameOverride); handled || err != nil {
					return err
				}
			}
			raw, name, resolvedSource, err := readThemeImport(cmd, client, source, nameOverride)
			if err != nil {
				return err
			}
			path, err := coreconfig.ImportTheme(name, raw)
			if err != nil {
				return err
			}
			result := themeImportResult{Theme: name, Status: "imported", Path: path, Source: resolvedSource}
			if contract.Enabled(cmd) {
				return shared.WriteJSON(cmd, result)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "imported theme: %s\n%s\n", name, path)
			return err
		},
	}
	cmd.Flags().String("name", "", "Install the imported theme under this name")
	return cmd
}

func readThemeImport(cmd *cobra.Command, client *http.Client, source, nameOverride string) ([]byte, string, string, error) {
	if source == "" {
		if shared.StdinAvailable(cmd.InOrStdin()) {
			if nameOverride == "" {
				return nil, "", "", shared.InvalidArgument(fmt.Errorf("--name is required when importing a theme from stdin"))
			}
			raw, err := readLimitedTheme(cmd.InOrStdin(), "theme from stdin")
			return raw, nameOverride, "<stdin>", err
		}
		if shared.MachineMode(cmd) {
			return nil, "", "", shared.InteractionRequiredWithArguments("theme import requires a file path, directory path, HTTP URL, or redirected stdin in JSON mode", "external_input", false, "")
		}
		progress.Stop()
		selected, err := shared.PickFile([]string{".toml"})
		if err != nil {
			return nil, "", "", err
		}
		if selected == "" {
			return nil, "", "", fmt.Errorf("no theme selected")
		}
		source = selected
	}

	parsed, err := url.Parse(source)
	if err != nil {
		return nil, "", "", shared.InvalidArgument(fmt.Errorf("invalid theme source %q: %w", source, err))
	}
	if parsed.Scheme == "http" || parsed.Scheme == "https" {
		filename := pathpkg.Base(parsed.Path)
		name := nameOverride
		if name == "" {
			name, err = themeNameFromFilename(filename)
		} else if filepath.Ext(filename) != ".toml" {
			err = shared.InvalidArgument(fmt.Errorf("theme source filename must end with .toml"))
		}
		if err != nil {
			return nil, "", "", err
		}
		raw, err := downloadTheme(cmd, client, source)
		return raw, name, source, err
	}
	if parsed.Scheme != "" {
		return nil, "", "", shared.InvalidArgument(fmt.Errorf("unsupported theme URL scheme %q; use http or https", parsed.Scheme))
	}
	filename := filepath.Base(source)
	name := nameOverride
	if name == "" {
		name, err = themeNameFromFilename(filename)
	} else if filepath.Ext(filename) != ".toml" {
		err = shared.InvalidArgument(fmt.Errorf("theme source filename must end with .toml"))
	}
	if err != nil {
		return nil, "", "", err
	}
	raw, err := readLimitedThemeFile(source)
	return raw, name, source, err
}

func importThemeDirectoryPath(cmd *cobra.Command, source, nameOverride string) (bool, error) {
	info, err := os.Lstat(source)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return true, fmt.Errorf("inspect theme source directory %s: %w", source, err)
	}
	if !info.IsDir() {
		return false, nil
	}
	dir, err := os.Open(source)
	if err != nil {
		return true, fmt.Errorf("open theme source directory %s: %w", source, err)
	}
	defer func() { _ = dir.Close() }()
	return true, importThemeDirectory(cmd, dir, source, nameOverride, func(filename string) (*os.File, error) {
		return os.Open(filepath.Join(source, filename))
	})
}

func importThemeDirectoryStdin(cmd *cobra.Command, nameOverride string) (bool, error) {
	dir, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return false, nil
	}
	info, err := dir.Stat()
	if err != nil || !info.IsDir() {
		return false, nil
	}
	if shared.MachineMode(cmd) {
		return true, shared.InvalidArgument(fmt.Errorf("directory stdin for theme import is an experimental human-mode feature"))
	}
	return true, importThemeDirectory(cmd, dir, "<stdin-directory>", nameOverride, func(filename string) (*os.File, error) {
		return shared.OpenStdinDirectoryFile(dir, filename)
	})
}

func importThemeDirectory(cmd *cobra.Command, dir *os.File, source, nameOverride string, openChild func(string) (*os.File, error)) error {
	if nameOverride != "" {
		return shared.InvalidArgument(fmt.Errorf("--name cannot be used when importing a theme directory; theme names come from filenames"))
	}
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return fmt.Errorf("read theme directory %s: %w", source, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 || entry.Type().IsDir() || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}
		names = append(names, entry.Name())
	}
	strfold.Sort(names)
	if len(names) == 0 {
		return shared.InvalidArgument(fmt.Errorf("theme directory %s contains no top-level .toml files", source))
	}
	imports := make([]coreconfig.ThemeImport, 0, len(names))
	skipped := make([]coreconfig.SkippedThemeImport, 0)
	for _, filename := range names {
		name, err := themeNameFromFilename(filename)
		if err != nil {
			return err
		}
		destination, exists, err := coreconfig.InspectThemeDestination(name)
		if err != nil {
			return err
		}
		if exists {
			skipped = append(skipped, coreconfig.SkippedThemeImport{Name: name, Path: destination})
			continue
		}
		child, err := openChild(filename)
		if err != nil {
			return fmt.Errorf("open theme directory file %q: %w", filename, err)
		}
		childInfo, statErr := child.Stat()
		if statErr != nil {
			_ = child.Close()
			return fmt.Errorf("inspect theme directory file %q: %w", filename, statErr)
		}
		if !childInfo.Mode().IsRegular() {
			_ = child.Close()
			continue
		}
		raw, readErr := readLimitedTheme(child, "theme directory file "+filename)
		closeErr := child.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return fmt.Errorf("close theme directory file %q: %w", filename, closeErr)
		}
		imports = append(imports, coreconfig.ThemeImport{Name: name, Data: raw})
	}
	if len(imports) == 0 && len(skipped) == 0 {
		return shared.InvalidArgument(fmt.Errorf("theme directory %s contains no top-level regular .toml files", source))
	}
	imported, racedSkips, err := coreconfig.ImportThemesSkippingExisting(imports)
	if err != nil {
		return err
	}
	skipped = append(skipped, racedSkips...)
	for _, item := range skipped {
		corelog.For("theme.import").Warn("theme already exists; skipping", "theme", item.Name, "path", item.Path)
		shared.AddMachineWarning(cmd, shared.MachineWarning{
			Code:    "theme.already_exists",
			Message: "The theme already exists and was skipped during batch import.",
			Target:  item.Name,
			Details: struct {
				Path string `json:"path"`
			}{Path: item.Path},
		})
	}
	if contract.Enabled(cmd) {
		return shared.WriteJSON(cmd, newThemeBatchImportResult(source, imported, skipped))
	}
	for _, item := range imported {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "imported theme: %s\n%s\n", item.Name, item.Path); err != nil {
			return err
		}
	}
	return nil
}

func newThemeBatchImportResult(source string, imported []coreconfig.ImportedTheme, skipped []coreconfig.SkippedThemeImport) themeBatchImportResult {
	items := make([]themeBatchImportItem, 0, len(imported)+len(skipped))
	for _, item := range imported {
		items = append(items, themeBatchImportItem{Theme: item.Name, Status: "imported", Path: item.Path})
	}
	for _, item := range skipped {
		items = append(items, themeBatchImportItem{Theme: item.Name, Status: "skipped", Path: item.Path, Reason: "already_exists"})
	}
	slices.SortFunc(items, func(left, right themeBatchImportItem) int {
		return strfold.Compare(left.Theme, right.Theme)
	})
	return themeBatchImportResult{
		Source: source, Count: len(items), ImportedCount: len(imported), SkippedCount: len(skipped), Items: items,
	}
}

func themeNameFromFilename(filename string) (string, error) {
	if filepath.Ext(filename) != ".toml" {
		return "", shared.InvalidArgument(fmt.Errorf("theme source filename must end with .toml"))
	}
	name := strings.TrimSuffix(filename, ".toml")
	if err := coreconfig.ValidateThemeName(name); err != nil {
		return "", shared.InvalidArgument(err)
	}
	return name, nil
}

func readLimitedThemeFile(path string) ([]byte, error) {
	file, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve theme source path: %w", err)
	}
	handle, err := openThemeSource(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = handle.Close() }()
	return readLimitedTheme(handle, "theme source "+file)
}

func downloadTheme(cmd *cobra.Command, client *http.Client, source string) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(shared.CommandContext(cmd), http.MethodGet, source, nil)
	if err != nil {
		return nil, shared.InvalidArgument(fmt.Errorf("create theme download request: %w", err))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download theme from %s: %w", source, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &shared.ValidationError{Code: "validation.failed", Source: "theme", Stage: "download", Target: source, Err: fmt.Errorf("download theme from %s: server returned %s", source, resp.Status)}
	}
	return readLimitedTheme(resp.Body, "downloaded theme")
}

func readLimitedTheme(reader io.Reader, label string) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(reader, maximumThemeImportBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", label, err)
	}
	if len(raw) > maximumThemeImportBytes {
		return nil, shared.InvalidArgument(fmt.Errorf("%s exceeds the %d-byte limit", label, maximumThemeImportBytes))
	}
	return raw, nil
}
