package get

import (
	"context"
	"sync"

	"github.com/yumauri/fbrcm/cli/shared"
	sharedrc "github.com/yumauri/fbrcm/cli/shared/rc"
	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
)

func loadProjectParameters(ctx context.Context, svc *core.Core, projectID string, update bool) (*core.ParametersCache, string, error) {
	if update {
		return svc.RevalidateParameters(ctx, projectID)
	}
	return svc.GetParameters(ctx, projectID, false)
}

type loadedProjectParameters struct {
	project core.Project
	cache   *core.ParametersCache
	cfg     *firebase.RemoteConfig
	source  string
	status  string
}

func loadProjectsParameters(ctx context.Context, svc *core.Core, projects []core.Project, update bool) ([]loadedProjectParameters, error) {
	if len(projects) == 0 {
		return nil, nil
	}

	type job struct {
		index   int
		project core.Project
	}
	type result struct {
		index  int
		loaded loadedProjectParameters
		err    error
	}

	jobs := make(chan job)
	results := make(chan result, len(projects))

	workerCount := min(firebase.MaxConcurrentRequests(), len(projects))

	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for work := range jobs {
				loaded, err := loadProjectParametersWithFallback(ctx, svc, work.project, update)
				select {
				case results <- result{index: work.index, loaded: loaded, err: err}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for i, project := range projects {
			select {
			case jobs <- job{index: i, project: project}:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	loaded := make([]loadedProjectParameters, len(projects))
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		loaded[res.index] = res.loaded
	}

	return loaded, nil
}

func loadProjectParametersWithFallback(ctx context.Context, svc *core.Core, project core.Project, update bool) (loadedProjectParameters, error) {
	executionCtx, err := shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
	if err != nil {
		return loadedProjectParameters{}, err
	}
	cache, source, err := loadProjectParameters(executionCtx, svc, project.ProjectID, update)
	if err == nil {
		cfg, parseErr := sharedrc.ParseProjectRemoteConfig(project.ProjectID, cache.RemoteConfig)
		if parseErr != nil {
			return loadedProjectParameters{}, parseErr
		}
		return loadedProjectParameters{
			project: project,
			cache:   cache,
			cfg:     cfg,
			source:  source,
			status:  core.ParametersStatusLabel(source, cache.CachedAt, true, nil),
		}, nil
	}
	if !core.ExecutionPolicyFromContext(executionCtx).ReadLocalState {
		return loadedProjectParameters{}, err
	}

	cache, state, inspectErr := svc.InspectParametersCache(project.ProjectID)
	if inspectErr != nil {
		return loadedProjectParameters{}, err
	}
	if state != core.ParametersCacheMissing && cache != nil {
		cfg, parseErr := sharedrc.ParseCachedProjectRemoteConfig(project.ProjectID, cache.RemoteConfig)
		if parseErr != nil {
			return loadedProjectParameters{}, parseErr
		}
		return loadedProjectParameters{
			project: project,
			cache:   cache,
			cfg:     cfg,
			source:  "cache-stale",
			status:  "stale",
		}, nil
	}

	return loadedProjectParameters{
		project: project,
		status:  "missing",
	}, nil
}
