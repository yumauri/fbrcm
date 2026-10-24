package theme

import (
	"fmt"
	"net/http"
	"path"

	"github.com/spf13/cobra"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/core/themesource"
	"github.com/yumauri/fbrcm/ops/shared"
)

func importRepositoryThemes(cmd *cobra.Command, client *http.Client, source *themesource.Source, nameOverride string) error {
	if source.Directory && nameOverride != "" {
		return shared.InvalidArgument(fmt.Errorf("--name cannot be used when importing a theme directory; theme names come from filenames"))
	}
	if firebase.IsOffline() {
		return firebase.ErrOffline
	}
	ctx := shared.CommandContext(cmd)
	repository, err := themesource.NewClient(client, source).Resolve(ctx)
	if err != nil {
		return err
	}
	files, err := repository.Files(ctx)
	if err != nil {
		return err
	}
	if source.Directory {
		batch := make([]themeBatchSource, 0, len(files))
		for _, file := range files {
			batch = append(batch, themeBatchSource{name: file.Name, read: func() ([]byte, error) { return repository.Read(ctx, file) }})
		}
		return importThemeBatch(cmd, source.URL, batch, &repository.Provenance)
	}
	file := files[0]
	name := nameOverride
	if name == "" {
		name, err = themeNameFromFilename(file.Name)
		if err != nil {
			return err
		}
	}
	// Check collisions before downloading a single theme, just as batches do.
	if _, exists, err := coreconfig.InspectThemeDestination(name); err != nil {
		return err
	} else if exists {
		return &coreconfig.ThemeError{Kind: coreconfig.ThemeErrorConflict, Theme: name, Err: fmt.Errorf("theme %q already exists", name)}
	}
	if path.Ext(file.Name) != ".toml" {
		return shared.InvalidArgument(fmt.Errorf("theme source filename must end with .toml"))
	}
	raw, err := repository.Read(ctx, file)
	if err != nil {
		return err
	}
	return importSingleTheme(cmd, source.URL, name, raw, &repository.Provenance)
}
