package apps

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yumauri/fbrcm/core"
	coreconfig "github.com/yumauri/fbrcm/core/config"
	"github.com/yumauri/fbrcm/ops/contract"
	"github.com/yumauri/fbrcm/ops/invocation"
	"github.com/yumauri/fbrcm/ops/shared"
	"github.com/yumauri/fbrcm/ops/shared/fileoutput"
)

type appReader interface {
	ReadFirebaseApps(context.Context, string, core.ListFirebaseAppsOptions) (core.FirebaseAppsResult, error)
	ReadFirebaseApp(context.Context, string, string, core.ListFirebaseAppsOptions) (core.FirebaseAppDetailsResult, error)
	ReadFirebaseAppConfig(context.Context, string, string, core.ListFirebaseAppsOptions) (core.FirebaseAppConfigResult, error)
}

type appListItem struct {
	core.FirebaseApp
	ProjectID string              `json:"project_id"`
	Project   string              `json:"project"`
	Source    core.AppCacheSource `json:"source" contract:"enum=firebase|cache|cache-stale"`
	CachedAt  *time.Time          `json:"cached_at,omitempty"`
}

type appListResult struct {
	Count int           `json:"count"`
	Items []appListItem `json:"items"`
}

type appShowResult struct {
	core.FirebaseAppDetails
	Source   core.AppCacheSource `json:"source" contract:"enum=firebase|cache|cache-stale"`
	CachedAt *time.Time          `json:"cached_at,omitempty"`
}

type appConfigResult struct {
	App               core.FirebaseApp      `json:"app"`
	SuggestedFilename string                `json:"suggested_filename"`
	Source            core.AppCacheSource   `json:"source" contract:"enum=firebase|cache|cache-stale"`
	CachedAt          *time.Time            `json:"cached_at,omitempty"`
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
	invocation.MustRegisterResponsePath(cmd, "list", appListResult{})
	invocation.MustRegisterResponsePath(cmd, "show", appShowResult{})
	invocation.MustRegisterResponsePath(cmd, "config", appConfigResult{})
	return cmd
}

func newListDefinition(svc *core.Core, reader appReader) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "list [project]",
		Short: "List applications across Firebase projects",
		Args:  invocation.MaximumNArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			cacheOpts, err := appCacheOptions(cmd)
			if err != nil {
				return err
			}
			projects, ctx, err := shared.ResolveListProjects(cmd, svc, args, "app", cacheOpts.CachedOnly)
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
			cacheOpts.ShowDeleted = showDeleted
			rawFilters, _ := cmd.Flags().GetStringArray("filter")
			items := make([]appListItem, 0)
			var singleRead core.FirebaseAppsResult
			for _, project := range projects {
				projectCtx, err := shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
				if err != nil {
					return err
				}
				read, err := reader.ReadFirebaseApps(projectCtx, project.ProjectID, cacheOpts)
				if err != nil {
					return classifyAppError(err)
				}
				singleRead = read
				addAppCacheWarnings(cmd, project.ProjectID, read.RefreshError, read.CacheError)
				apps := read.Apps
				if platform != "" {
					apps = filterByPlatform(apps, platform)
				}
				for _, app := range filterApps(apps, rawFilters) {
					items = append(items, appListItem{FirebaseApp: app, ProjectID: project.ProjectID, Project: project.Name, Source: read.Source, CachedAt: appCachedAt(read.CachedAt)})
				}
			}
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, appListResult{Count: len(items), Items: items})
			}
			nerdFontGlyphs, err := configuredNerdFontGlyphs()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				project := projects[0]
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project: %s (%s)\nSource: %s%s\n\n", project.Name, project.ProjectID, singleRead.Source, appCachedAtSuffix(singleRead.CachedAt))
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), renderAppListTable(items, len(args) == 0, shared.TerminalWidth(), nerdFontGlyphs))
			return err
		},
	}
	shared.AddProjectFilterFlag(cmd)
	cmd.Flags().StringArrayP("filter", "f", nil, "Filter applications by mode-prefixed name, namespace, or app ID query (^, /, ~, =); may be repeated")
	cmd.Flags().String("platform", "", "Only list one platform: android, ios, or web")
	cmd.Flags().Bool("show-deleted", false, "Include applications pending permanent deletion")
	cmd.Flags().Bool("json", false, "Print applications as JSON")
	addAppCacheFlags(cmd)
	return cmd
}

func newShowDefinition(svc *core.Core, reader appReader) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "show <app>",
		Short: "Show Firebase application details",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			cacheOpts, err := appCacheOptions(cmd)
			if err != nil {
				return err
			}
			project, ctx, err := resolveAppProject(cmd, svc, args[0])
			if err != nil {
				return err
			}
			read, err := reader.ReadFirebaseApp(ctx, project.ProjectID, args[0], cacheOpts)
			if err != nil {
				return classifyAppError(err)
			}
			addAppCacheWarnings(cmd, project.ProjectID, read.RefreshError, read.CacheError)
			jsonOut, _ := cmd.Flags().GetBool("json")
			if jsonOut {
				return shared.WriteJSON(cmd, appShowResult{FirebaseAppDetails: read.App, Source: read.Source, CachedAt: appCachedAt(read.CachedAt)})
			}
			nerdFontGlyphs, err := configuredNerdFontGlyphs()
			if err != nil {
				return err
			}
			details := renderAppDetails(read.App, nerdFontGlyphs) + "\nSource: " + string(read.Source)
			if !read.CachedAt.IsZero() {
				details += "\nCached at: " + read.CachedAt.Local().Format("2006-01-02 15:04:05")
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), details)
			return err
		},
	}
	cmd.Flags().StringP("project", "p", "", appProjectFlagHelp)
	cmd.Flags().Bool("json", false, "Print application details as JSON")
	addAppCacheFlags(cmd)
	return cmd
}

func newConfigDefinition(svc *core.Core, reader appReader) *invocation.Definition {
	cmd := &invocation.Definition{
		Use:   "config <app>",
		Short: "Download Firebase application SDK configuration",
		Args:  invocation.ExactArgs(1),
		RunE: func(cmd invocation.Call, args []string) error {
			cacheOpts, err := appCacheOptions(cmd)
			if err != nil {
				return err
			}
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
			read, err := reader.ReadFirebaseAppConfig(ctx, project.ProjectID, args[0], cacheOpts)
			if err != nil {
				return classifyAppError(err)
			}
			addAppCacheWarnings(cmd, project.ProjectID, read.RefreshError, read.CacheError)
			cfg := read.Config
			if toPath == "" {
				if contract.Enabled(cmd) {
					target := cfg.App.ResourceName
					return shared.WriteJSON(cmd, appConfigResult{App: cfg.App, SuggestedFilename: cfg.SuggestedFilename, Source: read.Source, CachedAt: appCachedAt(read.CachedAt), Artifact: contract.NewArtifact(&target, cfg.MediaType, cfg.Contents, nil, false)})
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
				return shared.WriteJSON(cmd, appConfigResult{App: cfg.App, SuggestedFilename: cfg.SuggestedFilename, Source: read.Source, CachedAt: appCachedAt(read.CachedAt), Artifact: contract.NewArtifact(&target, cfg.MediaType, cfg.Contents, &destination, overwrite)})
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "downloaded app configuration: %s\n", toPath)
			return err
		},
	}
	cmd.Flags().StringP("project", "p", "", appProjectFlagHelp)
	cmd.Flags().String("to", "", "Write application configuration to file path")
	cmd.Flags().Bool("json", false, "Print application configuration as a JSON artifact")
	shared.AddYesFlag(cmd, "Overwrite an existing destination without confirmation")
	addAppCacheFlags(cmd)
	return cmd
}

func addAppCacheFlags(cmd *invocation.Definition) {
	cmd.Flags().Bool("update", false, "Refresh application data from Firebase and update the cache")
	cmd.Flags().Bool("cached", false, "Use cached application data without contacting Firebase, even when stale")
	cmd.MarkFlagsMutuallyExclusive("update", "cached")
}

func appCacheOptions(cmd invocation.Call) (core.ListFirebaseAppsOptions, error) {
	update, _ := cmd.Flags().GetBool("update")
	cached, _ := cmd.Flags().GetBool("cached")
	if !core.ExecutionPolicyFromContext(shared.CommandContext(cmd)).ReadLocalState && (update || cached) {
		return core.ListFirebaseAppsOptions{}, shared.InvalidArgument(fmt.Errorf("--update and --cached cannot be used with --stateless; stateless execution does not use the application cache"))
	}
	return core.ListFirebaseAppsOptions{Update: update, CachedOnly: cached}, nil
}

func appCachedAt(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func appCachedAtSuffix(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return " (cached " + value.Local().Format("2006-01-02 15:04:05") + ")"
}

func addAppCacheWarnings(cmd invocation.Call, projectID string, refreshErr, cacheErr error) {
	if refreshErr != nil {
		shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "cache.stale", Message: "The command used stale cached Firebase application data after refresh failed.", Target: projectID, Details: struct {
			Source string `json:"source"`
		}{Source: "cache-stale"}})
		if !contract.Enabled(cmd) {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: using stale cached application data: %v\n", refreshErr)
		}
	}
	if cacheErr != nil {
		shared.AddMachineWarning(cmd, shared.MachineWarning{Code: "cache.write_failed", Message: "Firebase application data was returned, but the local cache could not be updated.", Target: projectID, Details: struct {
			Error string `json:"error"`
		}{Error: cacheErr.Error()}})
		if !contract.Enabled(cmd) {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: application cache update failed: %v\n", cacheErr)
		}
	}
}

func resolveAppProject(cmd invocation.Call, svc *core.Core, appSelector string) (core.Project, context.Context, error) {
	ctx := shared.CommandContext(cmd)
	projectQuery, _ := cmd.Flags().GetString("project")
	cachedOnly, _ := cmd.Flags().GetBool("cached")
	var (
		project core.Project
		err     error
	)
	if projectQuery != "" {
		if cachedOnly {
			project, err = shared.ResolveCachedProjectScopedResource(cmd, projectQuery, "app")
		} else {
			project, err = shared.ResolveProjectScopedResourceForExecution(ctx, cmd, svc, projectQuery, "app")
		}
	} else if appID, ok := core.ParseFirebaseAppID(appSelector); ok {
		if cachedOnly {
			project, err = shared.ResolveCachedProjectNumber(cmd, appID.ProjectNumber)
		} else {
			project, err = shared.ResolveProjectNumberForExecution(ctx, cmd, svc, appID.ProjectNumber)
		}
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
	if cacheMiss, ok := errors.AsType[*core.AppCacheMissError](err); ok {
		query := cacheMiss.ProjectID
		if cacheMiss.AppID != "" {
			query += "/" + cacheMiss.AppID
		}
		return &shared.SelectionError{Resource: "application cache", Kind: "not_found", Query: query, Err: err}
	}
	lookup, ok := errors.AsType[*core.AppLookupError](err)
	if !ok {
		return err
	}
	candidates := make([]shared.SelectionCandidate, 0, len(lookup.Candidates))
	for _, candidate := range lookup.Candidates {
		candidates = append(candidates, shared.SelectionCandidate{Name: candidate.Name, ID: candidate.ID})
	}
	return &shared.SelectionError{Resource: "app", Kind: lookup.Kind, Query: lookup.Query, Candidates: candidates, Err: err}
}
