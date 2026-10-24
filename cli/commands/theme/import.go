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

	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	corelog "github.com/yumauri/fbrcm/core/log"
	"github.com/yumauri/fbrcm/core/strfold"
	"github.com/yumauri/fbrcm/core/themesource"
	"github.com/yumauri/fbrcm/internal/terminal/progress"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/shared"
)

const maximumThemeImportBytes = themesource.MaximumFileBytes

func newImportCommand(client *http.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [source]",
		Short: "Import themes from a file, directory, HTTP URL, or repository URL",
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
			ref, err := cmd.Flags().GetString("ref")
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("ref") && ref == "" {
				return shared.InvalidArgument(fmt.Errorf("--ref cannot be empty"))
			}
			repositorySource, err := themesource.Parse(source, ref)
			if err != nil {
				return err
			}
			if repositorySource != nil {
				return importRepositoryThemes(cmd, client, repositorySource, nameOverride)
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
			return importSingleTheme(cmd, resolvedSource, name, raw, nil)
		},
	}
	cmd.Flags().String("name", "", "Install the imported theme under this name")
	cmd.Flags().String("ref", "", "Disambiguate the revision prefix already present in a repository browser URL")
	return cmd
}

func importSingleTheme(cmd *cobra.Command, source, name string, raw []byte, provenance *themesource.Provenance) error {
	destination, err := coreconfig.ImportTheme(name, raw)
	if err != nil {
		return err
	}
	result := themeImportResult{Theme: name, Status: "imported", Path: destination, Source: source, RepositorySource: provenance}
	if contract.Enabled(cmd) {
		return shared.WriteJSON(cmd, result)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "imported theme: %s\n%s\n", name, destination)
	return err
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
	files := make([]themeBatchSource, 0, len(names))
	for _, filename := range names {
		files = append(files, themeBatchSource{name: filename, read: func() ([]byte, error) {
			child, err := openChild(filename)
			if err != nil {
				return nil, fmt.Errorf("open theme directory file %q: %w", filename, err)
			}
			childInfo, statErr := child.Stat()
			if statErr != nil {
				_ = child.Close()
				return nil, fmt.Errorf("inspect theme directory file %q: %w", filename, statErr)
			}
			if !childInfo.Mode().IsRegular() {
				_ = child.Close()
				return nil, nil
			}
			raw, readErr := readLimitedTheme(child, "theme directory file "+filename)
			closeErr := child.Close()
			if readErr != nil {
				return nil, readErr
			}
			if closeErr != nil {
				return nil, fmt.Errorf("close theme directory file %q: %w", filename, closeErr)
			}
			return raw, nil
		}})
	}
	return importThemeBatch(cmd, source, files, nil)
}

type themeBatchSource struct {
	name string
	read func() ([]byte, error)
}

func importThemeBatch(cmd *cobra.Command, source string, files []themeBatchSource, provenance *themesource.Provenance) error {
	slices.SortFunc(files, func(a, b themeBatchSource) int { return strfold.Compare(a.name, b.name) })
	imports := make([]coreconfig.ThemeImport, 0, len(files))
	skipped := make([]coreconfig.SkippedThemeImport, 0)
	for _, file := range files {
		name, err := themeNameFromFilename(file.name)
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
		raw, err := file.read()
		if err != nil {
			return err
		}
		if raw == nil {
			continue
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
		result := newThemeBatchImportResult(source, imported, skipped)
		result.RepositorySource = provenance
		return shared.WriteJSON(cmd, result)
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
	if firebase.IsOffline() {
		return nil, firebase.ErrOffline
	}
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
