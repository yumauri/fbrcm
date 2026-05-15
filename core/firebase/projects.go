package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	corelog "fbrcm/core/log"
)

// Project holds project state used by the firebase package.
type Project struct {
	// Name stores name for Project.
	Name string
	// ProjectID stores project id for Project.
	ProjectID string
	// ProjectNumber stores project number for Project.
	ProjectNumber string
	// State stores state for Project.
	State string
	// ETag stores etag for Project.
	ETag string
	// UpdateTime stores update time for Project.
	UpdateTime string
}

// listProjectsResponse holds list projects response state used by the firebase package.
type listProjectsResponse struct {
	// Projects stores projects for listProjectsResponse.
	Projects []struct {
		ProjectID      string `json:"projectId"`
		ProjectNumber  string `json:"projectNumber"`
		Name           string `json:"name"`
		LifecycleState string `json:"lifecycleState"`
	} `json:"projects"`
	// NextPageToken stores next page token for listProjectsResponse.
	NextPageToken string `json:"nextPageToken"`
}

// getProjectV3Response holds get project v3 response state used by the firebase package.
type getProjectV3Response struct {
	// ProjectID stores project id for getProjectV3Response.
	ProjectID string `json:"projectId"`
	// DisplayName stores display name for getProjectV3Response.
	DisplayName string `json:"displayName"`
	// State stores state for getProjectV3Response.
	State string `json:"state"`
	// Etag stores etag for getProjectV3Response.
	Etag string `json:"etag"`
	// UpdateTime stores update time for getProjectV3Response.
	UpdateTime string `json:"updateTime"`
}

// Fetch all Firebase projects accessible to the authenticated user
func (s *Service) ListProjects(ctx context.Context) ([]Project, error) {
	logger := corelog.For("firebase")
	logger.Info("firebase list projects start")

	pageToken := ""
	projects := make([]Project, 0)
	page := 0

	for {
		page++
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"https://cloudresourcemanager.googleapis.com/v1/projects",
			nil,
		)
		if err != nil {
			logger.Error("create projects request failed", "page", page, "err", err)
			return nil, fmt.Errorf("create project list request: %w", err)
		}

		q := req.URL.Query()
		q.Set("pageSize", "1000")
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		req.URL.RawQuery = q.Encode()
		logHTTPRequest(logger.With("page", page), req)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			logger.Error("projects request failed", "page", page, "err", err)
			return nil, fmt.Errorf("list projects: %w", err)
		}
		logHTTPResponse(logger.With("page", page), req, resp)

		body, err := io.ReadAll(resp.Body)
		defer func() { _ = resp.Body.Close() }()
		if err != nil {
			logger.Error("read projects response failed", "page", page, "err", err)
			return nil, fmt.Errorf("read project list response: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			logger.Error("projects api returned non-200", "page", page, "status", resp.Status)
			return nil, fmt.Errorf("project list api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}

		var payload listProjectsResponse
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&payload); err != nil {
			logger.Error("decode projects response failed", "page", page, "err", err)
			return nil, fmt.Errorf("decode project list response: %w", err)
		}

		before := len(projects)
		for _, project := range payload.Projects {
			if project.ProjectID == "" || project.LifecycleState == "DELETE_REQUESTED" {
				continue
			}
			projects = append(projects, Project{
				Name:          strings.TrimSpace(project.Name),
				ProjectID:     project.ProjectID,
				ProjectNumber: strings.TrimSpace(project.ProjectNumber),
				State:         strings.TrimSpace(project.LifecycleState),
			})
		}
		logger.Info("projects page loaded", "page", page, "accepted", len(projects)-before, "next_page", payload.NextPageToken != "")

		if payload.NextPageToken == "" {
			break
		}
		pageToken = payload.NextPageToken
	}

	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Name == projects[j].Name {
			return projects[i].ProjectID < projects[j].ProjectID
		}
		return projects[i].Name < projects[j].Name
	})

	enriched, err := s.enrichProjects(ctx, projects)
	if err != nil {
		logger.Warn("project details enrichment interrupted; using partial details", "err", err)
	} else {
		projects = enriched
	}

	logger.Info("firebase list projects complete", "count", len(projects), "pages", page)
	return projects, nil
}

// enrichProjects handles enrich projects for Service and returns the resulting state or error.
func (s *Service) enrichProjects(ctx context.Context, projects []Project) ([]Project, error) {
	logger := corelog.For("firebase")
	if len(projects) == 0 {
		return nil, nil
	}

	// job holds job state used by the firebase package.
	type job struct {
		// index stores index for job.
		index int
		// project stores project for job.
		project Project
	}
	// result holds result state used by the firebase package.
	type result struct {
		// index stores index for result.
		index int
		// project stores project for result.
		project Project
	}

	jobs := make(chan job)
	results := make(chan result, len(projects))
	errCh := make(chan error, 1)

	workerCount := min(maxConcurrentRequests, len(projects))

	for range workerCount {
		go func() {
			for work := range jobs {
				details, err := s.GetProject(ctx, work.project.ProjectID)
				if err != nil {
					logger.Warn("project details lookup failed; using list response only", "project_id", work.project.ProjectID, "err", err)
					results <- result(work)
					continue
				}
				results <- result{index: work.index, project: details.mergeInto(work.project)}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for i, project := range projects {
			select {
			case jobs <- job{index: i, project: project}:
			case <-ctx.Done():
				select {
				case errCh <- ctx.Err():
				default:
				}
				return
			}
		}
	}()

	enriched := make([]Project, len(projects))
	for range projects {
		select {
		case res := <-results:
			enriched[res.index] = res.project
		case err := <-errCh:
			return enriched, err
		case <-ctx.Done():
			return enriched, ctx.Err()
		}
	}

	return enriched, nil
}

// GetProject gets project for Service and returns the resulting state or error.
func (s *Service) GetProject(ctx context.Context, projectID string) (Project, error) {
	logger := corelog.For("firebase")
	logger.Info("get project details", "project_id", projectID)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://cloudresourcemanager.googleapis.com/v3/projects/%s", projectID),
		nil,
	)
	if err != nil {
		logger.Error("create project details request failed", "project_id", projectID, "err", err)
		return Project{}, fmt.Errorf("create project details request: %w", err)
	}

	logHTTPRequest(logger.With("project_id", projectID), req)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("project details request failed", "project_id", projectID, "err", err)
		return Project{}, fmt.Errorf("get project details: %w", err)
	}
	logHTTPResponse(logger.With("project_id", projectID), req, resp)

	body, err := io.ReadAll(resp.Body)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		logger.Error("read project details response failed", "project_id", projectID, "err", err)
		return Project{}, fmt.Errorf("read project details response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("project details api returned non-200", "project_id", projectID, "status", resp.Status)
		return Project{}, fmt.Errorf("project details api returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload getProjectV3Response
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&payload); err != nil {
		logger.Error("decode project details response failed", "project_id", projectID, "err", err)
		return Project{}, fmt.Errorf("decode project details response: %w", err)
	}

	return Project{
		Name:       strings.TrimSpace(payload.DisplayName),
		ProjectID:  strings.TrimSpace(payload.ProjectID),
		State:      strings.TrimSpace(payload.State),
		ETag:       strings.TrimSpace(payload.Etag),
		UpdateTime: strings.TrimSpace(payload.UpdateTime),
	}, nil
}

// mergeInto handles merge into for Project and returns the resulting state or error.
func (p Project) mergeInto(base Project) Project {
	if p.Name != "" {
		base.Name = p.Name
	}
	if p.ProjectID != "" {
		base.ProjectID = p.ProjectID
	}
	if p.ProjectNumber != "" {
		base.ProjectNumber = p.ProjectNumber
	}
	if p.State != "" {
		base.State = p.State
	}
	if p.ETag != "" {
		base.ETag = p.ETag
	}
	if p.UpdateTime != "" {
		base.UpdateTime = p.UpdateTime
	}
	return base
}
