package get

import (
	"context"
	"sync"

	"github.com/yumauri/fbrcm/core"
	"github.com/yumauri/fbrcm/core/firebase"
	"github.com/yumauri/fbrcm/ops/shared"
	sharedrc "github.com/yumauri/fbrcm/ops/shared/rc"
)

func loadProjectParameters(ctx context.Context, svc *core.Core, projectID string, opts core.ParametersReadOptions) (*core.ParametersCache, string, error) {
	return shared.ReadParametersWithFallback(ctx, svc, projectID, opts)
}

type loadedProjectParameters struct {
	project core.Project
	cache   *core.ParametersCache
	cfg     *firebase.RemoteConfig
	source  string
	status  string
}

func loadProjectsParameters(ctx context.Context, svc *core.Core, projects []core.Project, opts core.ParametersReadOptions) ([]loadedProjectParameters, error) {
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

	workerCount := min(firebase.MaxConcurrentRequests(ctx), len(projects))

	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for work := range jobs {
				loaded, err := loadProjectParametersWithFallback(ctx, svc, work.project, opts)
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

func loadProjectParametersWithFallback(ctx context.Context, svc *core.Core, project core.Project, opts core.ParametersReadOptions) (loadedProjectParameters, error) {
	executionCtx, err := shared.FirebaseServiceContextForExecution(ctx, project.ProjectID)
	if err != nil {
		return loadedProjectParameters{}, err
	}
	cache, source, err := loadProjectParameters(executionCtx, svc, project.ProjectID, opts)
	if err == nil {
		parse := sharedrc.ParseProjectRemoteConfig
		status := core.ParametersStatusLabel(source, cache.CachedAt, true, nil)
		if source == "cache-stale" {
			parse = sharedrc.ParseCachedProjectRemoteConfig
			status = "stale"
		}
		cfg, parseErr := parse(project.ProjectID, cache.RemoteConfig)
		if parseErr != nil {
			return loadedProjectParameters{}, parseErr
		}
		return loadedProjectParameters{
			project: project,
			cache:   cache,
			cfg:     cfg,
			source:  source,
			status:  status,
		}, nil
	}
	if !core.ExecutionPolicyFromContext(executionCtx).ReadLocalState {
		return loadedProjectParameters{}, err
	}

	_, _, inspectErr := svc.InspectParametersCache(project.ProjectID)
	if inspectErr != nil {
		return loadedProjectParameters{}, err
	}

	return loadedProjectParameters{
		project: project,
		status:  "missing",
	}, nil
}
