package apps

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yumauri/fbrcm/core"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/ops/shared/fileoutput"
)

type appReader interface {
	ListFirebaseApps(context.Context, string, core.ListFirebaseAppsOptions) ([]core.FirebaseApp, error)
	GetFirebaseApp(context.Context, string, string) (core.FirebaseAppDetails, error)
	GetFirebaseAppConfig(context.Context, string, string) (core.FirebaseAppConfig, error)
}

type appConfigResult struct {
	App               core.FirebaseApp      `json:"app"`
	SuggestedFilename string                `json:"suggested_filename"`
	Artifact          contract.ArtifactData `json:"artifact"`
}

const appProjectFlagHelp = "Project ID, repository alias, exact display name, or mode-prefixed name/ID query (^, /, ~, =); required unless <app> is a complete Firebase App ID"

func NewDefinition(svc *core.Core) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "apps",
		Short: "Inspect Firebase applications",
		Long:  "List Firebase applications, inspect platform-specific application details, and download SDK configuration.",
	}
	cmd.AddCommand(newListDefinition(svc, svc), newShowDefinition(svc, svc), newConfigDefinition(svc, svc))
	invocation.MustRegisterResponsePath(cmd, "list", []core.FirebaseApp{})
	invocation.MustRegisterResponsePath(cmd, "show", core.FirebaseAppDetails{})
	invocation.MustRegisterResponsePath(cmd, "config", appConfigResult{})
	return cmd
}

func newListDefinition(svc *core.Core, reader appReader) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "list <project>",
		Short: "List applications registered in a Firebase project",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			project, ctx, err := resolveProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			platformName, _ := cmd.Flags().GetString("platform")
			var platform core.AppPlatform
			if platformName != "" {
				platform, err = parsePlatform(platformName)
				if err != nil {
					return err
				}
			}
			showDeleted, _ := cmd.Flags().GetBool("show-deleted")
			items, err := reader.ListFirebaseApps(ctx, project.ProjectID, core.ListFirebaseAppsOptions{ShowDeleted: showDeleted})
			if err != nil {
				return classifyAppError(err)
			}
			if platform != "" {
				items = filterByPlatform(items, platform)
			}
			rawFilters, _ := cmd.Flags().GetStringArray("filter")
			items = filterApps(items, rawFilters)
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, items)
			}
			nerdFontGlyphs, err := configuredNerdFontGlyphs()
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%s)\n\n", project.Name, project.ProjectID)
			_, err = fmt.Fprintln(cmd.OutOrStdout(), renderAppsTable(items, nerdFontGlyphs))
			return err
		},
	}
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter applications by mode-prefixed name, namespace, or app ID query (^, /, ~, =); may be repeated")
	cmd.Flags().String("platform", "", "Only list one platform: android, ios, or web")
	cmd.Flags().Bool("show-deleted", false, "Include applications pending permanent deletion")
	cmd.Flags().Bool("json", false, "Print applications as JSON")
	return cmd
}

func newShowDefinition(svc *core.Core, reader appReader) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "show <app>",
		Short: "Show Firebase application details",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			project, ctx, err := resolveAppProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			app, err := reader.GetFirebaseApp(ctx, project.ProjectID, args[0])
			if err != nil {
				return classifyAppError(err)
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, app)
			}
			nerdFontGlyphs, err := configuredNerdFontGlyphs()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), renderAppDetails(app, nerdFontGlyphs))
			return err
		},
	}
	cmd.Flags().StringP("project", "p", "", appProjectFlagHelp)
	cmd.Flags().Bool("json", false, "Print application details as JSON")
	return cmd
}

func newConfigDefinition(svc *core.Core, reader appReader) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "config <app>",
		Short: "Download Firebase application SDK configuration",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			project, ctx, err := resolveAppProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			toPath, _ := cmd.Flags().GetString("to")
			yes, _ := cmd.Flags().GetBool("yes")
			overwrite := false
			if toPath != "" {
				var proceed bool
				overwrite, proceed, err = shared.ConfirmFileOverwrite(cmd, toPath, yes)
				if err != nil || !proceed {
					return err
				}
			}
			cfg, err := reader.GetFirebaseAppConfig(ctx, project.ProjectID, args[0])
			if err != nil {
				return classifyAppError(err)
			}
			if toPath == "" {
				if contract.Enabled(cmd) {
					target := cfg.App.ResourceName
					return shared.WriteJSON(cmd, appConfigResult{App: cfg.App, SuggestedFilename: cfg.SuggestedFilename, Artifact: contract.NewArtifact(&target, cfg.MediaType, cfg.Contents, nil, false)})
				}
				_, err = cmd.OutOrStdout().Write(cfg.Contents)
				return err
			}
			write := fileoutput.Create
			if overwrite {
				write = fileoutput.Write
			}
			if err := write(toPath, cfg.Contents); err != nil {
				return err
			}
			if contract.Enabled(cmd) {
				target, destination := cfg.App.ResourceName, toPath
				return shared.WriteJSON(cmd, appConfigResult{App: cfg.App, SuggestedFilename: cfg.SuggestedFilename, Artifact: contract.NewArtifact(&target, cfg.MediaType, cfg.Contents, &destination, overwrite)})
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "downloaded app configuration: %s\n", toPath)
			return err
		},
	}
	cmd.Flags().StringP("project", "p", "", appProjectFlagHelp)
	cmd.Flags().String("to", "", "Write application configuration to file path")
	cmd.Flags().Bool("json", false, "Print application configuration as a JSON artifact")
	shared.AddYesFlag(cmd, "Overwrite an existing destination without confirmation")
	return cmd
}

func resolveAppProject(cmd invocation.Call, svc *core.Core, appSelector string) (core.Project, context.Context, error) {
	ctx := shared.CommandContext(cmd)
	projectQuery, _ := cmd.Flags().GetString("project")
	var (
		project core.Project
		err     error
	)
	if projectQuery != "" {
		project, err = shared.ResolveProjectScopedResourceForExecution(ctx, cmd, svc, projectQuery, "app")
	} else if appID, ok := core.ParseFirebaseAppID(appSelector); ok {
		project, err = shared.ResolveProjectNumberForExecution(ctx, cmd, svc, appID.ProjectNumber)
	} else {
		return core.Project{}, nil, shared.InvalidArgument(fmt.Errorf("--project is required unless <app> is a complete Firebase App ID"))
	}
	if err != nil {
		return core.Project{}, nil, err
	}
	ctx, err = shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
	if err != nil {
		return core.Project{}, nil, err
	}
	cmd.SetContext(ctx)
	return project, ctx, nil
}

func resolveProject(cmd invocation.Call, svc *core.Core, query string) (core.Project, context.Context, error) {
	ctx := shared.CommandContext(cmd)
	project, err := shared.ResolveProjectScopedResourceForExecution(ctx, cmd, svc, query, "app")
	if err != nil {
		return core.Project{}, nil, err
	}
	ctx, err = shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
	if err != nil {
		return core.Project{}, nil, err
	}
	cmd.SetContext(ctx)
	return project, ctx, nil
}

func configuredNerdFontGlyphs() (bool, error) {
	resolved, err := coreconfig.ResolveAppConfig()
	if err != nil {
		return false, err
	}
	return resolved.Effective.NerdFontGlyphs != nil && *resolved.Effective.NerdFontGlyphs, nil
}

func parsePlatform(raw string) (core.AppPlatform, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "android":
		return core.AppPlatformAndroid, nil
	case "ios":
		return core.AppPlatformIOS, nil
	case "web":
		return core.AppPlatformWeb, nil
	default:
		return "", shared.InvalidArgument(fmt.Errorf("--platform must be android, ios, or web"))
	}
}

func filterByPlatform(apps []core.FirebaseApp, platform core.AppPlatform) []core.FirebaseApp {
	filtered := make([]core.FirebaseApp, 0, len(apps))
	for _, app := range apps {
		if app.Platform == platform {
			filtered = append(filtered, app)
		}
	}
	return filtered
}

func filterApps(apps []core.FirebaseApp, rawFilters []string) []core.FirebaseApp {
	filters := shared.ParseFilters(rawFilters)
	if len(filters) == 0 {
		return apps
	}
	filtered := make([]core.FirebaseApp, 0, len(apps))
	for _, app := range apps {
		for _, value := range []string{app.DisplayName, app.Namespace, app.AppID} {
			if shared.MatchAnyFilter(value, filters) {
				filtered = append(filtered, app)
				break
			}
		}
	}
	return filtered
}

func classifyAppError(err error) error {
	var lookup *core.AppLookupError
	if !errors.As(err, &lookup) {
		return err
	}
	candidates := make([]shared.SelectionCandidate, 0, len(lookup.Candidates))
	for _, candidate := range lookup.Candidates {
		candidates = append(candidates, shared.SelectionCandidate{Name: candidate.Name, ID: candidate.ID})
	}
	return &shared.SelectionError{Resource: "app", Kind: lookup.Kind, Query: lookup.Query, Candidates: candidates, Err: err}
}
